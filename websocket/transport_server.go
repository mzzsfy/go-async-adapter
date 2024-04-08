package websocket

import (
    "bytes"
    "github.com/gobwas/ws"
    "github.com/gobwas/ws/wsutil"
    "github.com/mzzsfy/go-async-adapter/base"
    "io"
    "sync"
)

var (
    wsuPool = &sync.Pool{
        New: func() any {
            return &ws.Upgrader{}
        },
    }
)

//服务端接受到的websocket请求
type serverAsyncWs struct {
    conn     base.AsyncConnect
    handler  MessageHandler
    upgraded bool

    readBuf   bytes.Buffer
    curHeader *ws.Header //当前头数据
    lastIndex int        //上次读取字节的位置
}

func (s *serverAsyncWs) OnData(bs []byte) (err error) {
    if !s.upgraded {
        bs, err = s.upgrade(bs)
        if err != nil {
            return
        }
        if len(bs) == 0 {
            return
        }
    }
    err = s.data(bs)
    if err != nil {
        s.handler.OnError(err)
        if s.upgraded {
            s.handler.OnClose(1006, err.Error())
        }
    }
    return err
}

func (s *serverAsyncWs) upgrade(buf []byte) ([]byte, error) {
    //检测是否是一个完整的请求头,buf必须是\r\n结尾
    if len(buf) <= 1 || buf[len(buf)-1] != '\n' {
        s.readBuf.Write(buf)
        return nil, nil
    } else if len(buf) <= 2 || buf[len(buf)-2] != '\r' {
        s.readBuf.Write(buf)
        return nil, nil
    }
    if len(buf) <= 4 {
        s.readBuf.Write(buf)
        buf = nil
        bs := s.readBuf.Bytes()
        if bs[len(buf)-3] != '\n' || bs[len(buf)-4] != '\r' {
            return nil, nil
        }
    } else {
        if buf[len(buf)-3] != '\n' || buf[len(buf)-4] != '\r' {
            s.readBuf.Write(buf)
            return nil, nil
        }
    }
    wsUp := wsuPool.Get().(*ws.Upgrader)
    info := upgradeInfo{}
    wsUp.OnHeader = func(key, value []byte) error {
        info.AddHeader(key, value)
        return nil
    }
    wsUp.OnHost = func(host []byte) error {
        info.AddHeader(symbolHost, host)
        return nil
    }
    wsUp.OnRequest = func(uri []byte) error {
        n := bytes.SplitN(uri, symbolQuestion, 2)
        if len(n) != 0 {
            info.path = n[0]
        } else {
            info.path = symbolSlash
        }
        if len(n) > 1 {
            info.params = n[1]
        }
        return nil
    }
    wsUp.OnBeforeUpgrade = func() (header ws.HandshakeHeader, err error) { return nil, s.handler.OnUpgrade(info) }
    _, err := wsUp.Upgrade(&rw_{
        conn:   s.conn,
        bufNew: buf,
        bufOld: s.readBuf.Bytes(),
    })
    wsUp.OnHeader = nil
    wsUp.OnHost = nil
    wsUp.OnRequest = nil
    wsUp.OnBeforeUpgrade = nil
    wsuPool.Put(wsUp)
    if err != nil {
        return nil, err
    }
    s.upgraded = true
    return nil, nil
}

func (s *serverAsyncWs) data(bs []byte) (err error) {
    rw := &rw_{
        conn:   s.conn,
        bufNew: bs,
        bufOld: s.readBuf.Bytes(),
    }
    defer func() {
        if rw.ri < len(rw.bufNew)+len(rw.bufOld) {
            //有待处理数据,写入并等待数据完成
            if rw.ri >= len(rw.bufOld) {
                s.readBuf.Write(bs[rw.ri-len(rw.bufOld):])
            } else {
                s.readBuf.Write(bs)
            }
        } else {
            s.readBuf.Reset()
        }
    }()
    for {
        if s.curHeader == nil {
            if len(rw.bufNew)+len(rw.bufOld)-rw.ri < ws.MinHeaderSize { //头长度至少是2
                return
            }
            var head ws.Header
            head, err = ws.ReadHeader(bytes.NewReader(getHeadBytes(rw.bufOld, rw.ri, rw.bufOld)))
            if err == io.EOF { //数据不完整,合并到下次处理
                return nil
            }
            if err != nil {
                return err
            }
            s.curHeader = &head
        } else if s.lastIndex > 0 {
            if len(rw.bufNew)+len(rw.bufOld)-rw.ri <= s.lastIndex {
                return
            }
            var head ws.Header
            head, err = ws.ReadHeader(bytes.NewReader(getHeadBytes(rw.bufOld, rw.ri, rw.bufOld)))
            if err == io.EOF { //数据不完整,合并到下次处理
                return nil
            }
            if err != nil {
                return err
            }
            s.curHeader = &head
        }
        dataLen := int(s.curHeader.Length)
        if dataLen > 0 {
            if len(rw.bufNew)+len(rw.bufOld)-rw.ri < dataLen {
                return
            }
        }
        //当前 header 已经是一个完整消息
        if s.curHeader.Fin {
            var messages []wsutil.Message
            messages, err = wsutil.ReadClientMessage(rw, messages)
            s.lastIndex = 0
            s.curHeader = nil
            if err != nil {
                return err
            }
            for _, message := range messages {
                err = s.handler.OnMessage(Message{
                    OpCode:  OpCode(message.OpCode),
                    Payload: message.Payload,
                })
            }
        } else {
            //如果不是完整消息,改变reader位置
            s.lastIndex = s.lastIndex + wsHeadLength(s.curHeader) + dataLen
        }
    }
}

func getHeadBytes(newBuf []byte, start int, oldBuf []byte) []byte {
    if len(oldBuf) > ws.MaxHeaderSize {
        oldBuf = oldBuf[:ws.MaxHeaderSize]
    } else {
        end := start + ws.MaxHeaderSize - len(oldBuf) + start
        if end >= len(newBuf) {
            end = len(newBuf)
        }
        oldBuf = append(oldBuf, newBuf[start:end]...)
    }
    return oldBuf
}

func (s *serverAsyncWs) OnClose(err error) {
    s.handler.OnError(err)
}

func (s *serverAsyncWs) Close(closeCode uint16, data string) error {
    return s.conn.AsyncWrite(ws.NewCloseFrameBody(ws.StatusCode(closeCode), data))
}

func (s *serverAsyncWs) Ping() error {
    return s.conn.AsyncWrite(ws.CompiledPing)
}

func (s *serverAsyncWs) Send(m *SendMessage) error {
    if m.EnableCache {
        if len(m.cache) > 0 {
            return s.conn.AsyncWrite(m.cache)
        }
    }
    var header *ws.Header
    header = &ws.Header{
        Fin:    true,
        OpCode: ws.OpText,
        Length: int64(len(m.Data)),
    }
    if m.IsBinary {
        header.OpCode = ws.OpBinary
    }
    headLen := wsHeadLength(header)
    bs := make([]byte, int(header.Length)+headLen)
    writeWsHeader(bs, header)
    //ws.Cipher(message.Payload, header.Mask, 0)
    copy(bs[headLen:], m.Data)
    if m.EnableCache {
        m.cache = bs
    }
    return s.conn.AsyncWrite(bs)
}

func NewServerWs(conn base.AsyncConnect, handler MessageHandler) (AsyncWebsocket, error) {
    if handler == nil {
        panic("消息回调不能为nil")
    }
    s := &serverAsyncWs{handler: handler, conn: conn}
    return s, conn.SetAsyncCallback(s)
}
