package fastws

import (
    "bytes"
    "errors"
    "github.com/gobwas/ws"
    "github.com/gobwas/ws/wsutil"
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
    conn     AsyncConn
    handler  MessageHandler
    upgraded bool

    readBuf   bytes.Buffer
    curHeader *ws.Header //当前头数据
    lastIndex int        //上次读取字节的位置
}

func (s *serverAsyncWs) OnData(bs []byte) (err error) {
    if s.upgraded {
        bs, err = s.upgrade(bs)
        if err != nil {
            return
        }
        if len(bs) == 0 {
            return
        }
    }
    return s.data(bs)
}

func (s *serverAsyncWs) upgrade(buf []byte) ([]byte, error) {
    wsUp := wsuPool.Get().(*ws.Upgrader)
    buffer := bytes.NewBuffer(buf)
    //fixme:实现rw
    _, err := wsUp.Upgrade(buffer)
    wsuPool.Put(wsUp)
    //wsePool.Put(wse)
    if err != nil {
        return nil, err
    }
    s.upgraded = true
    s.handler.OnUpgrade(nil)
    //fixme
    return nil, err
}

func (s *serverAsyncWs) data(bs []byte) (err error) {
    remaining := len(bs) + s.readBuf.Len()
    defer func() {
        if remaining > 0 {
            //有待处理数据,写入并等待数据完成
            s.readBuf.Write(bs[len(bs)-remaining:])
        } else {
            s.readBuf.Reset()
        }
    }()
    for {
        if s.curHeader == nil {
            if remaining < ws.MinHeaderSize { //头长度至少是2
                return
            }
            var head ws.Header
            head, err = ws.ReadHeader(bytes.NewReader(getHeadBytes(bs, 0, s.readBuf.Bytes())))
            if err == io.EOF || err == io.ErrUnexpectedEOF { //数据不完整,合并到下次处理
                return nil
            }
            if err != nil {
                return
            }
            s.curHeader = &head
        } else if s.lastIndex > 0 {
            if remaining <= s.lastIndex {
                return
            }
            var head ws.Header
            if s.readBuf.Len() > s.lastIndex {
                bs1 := s.readBuf.Bytes()[s.lastIndex:]
                head, err = ws.ReadHeader(bytes.NewReader(getHeadBytes(bs, 0, bs1)))
            } else {
                head, err = ws.ReadHeader(bytes.NewReader(getHeadBytes(bs, s.lastIndex-s.readBuf.Len(), nil)))
            }
            if err == io.EOF || err == io.ErrUnexpectedEOF { //数据不完整,合并到下次处理
                return nil
            }
            if err != nil {
                return
            }
            s.curHeader = &head
        }
        dataLen := int(s.curHeader.Length)
        if dataLen > 0 {
            if remaining < dataLen {
                return
            }
        }
        //当前 header 已经是一个完整消息
        if s.curHeader.Fin {
            //fixme
            messages, err = wsutil.ReadClientMessage(reader, messages)
            s.lastIndex = 0
            s.curHeader = nil
            if err != nil {
                return err
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
        oldBuf = append(oldBuf, newBuf[start:start+ws.MaxHeaderSize-len(oldBuf)+start]...)
    }
    return oldBuf
}
func (s *serverAsyncWs) OnClose(err error) {
    s.handler.OnError(err)
}

func (s *serverAsyncWs) Close(closeCode uint16, data string) error {
    return s.conn.AsyncWrite(ws.NewCloseFrameBody(ws.StatusCode(closeCode), data))
}

func (s *serverAsyncWs) Start(conn AsyncConn) error {
    if conn == nil {
        return errors.New("连接不能为空")
    }
    if s.conn != nil {
        return errors.New("不允许重复启动")
    }
    err := conn.SetCallback(s)
    return err
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

func NewServerWs(handler MessageHandler) AsyncWs {
    if handler == nil {
        panic("消息回调不能为nil")
    }
    return &serverAsyncWs{handler: handler}
}
