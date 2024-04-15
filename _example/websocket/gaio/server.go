package gaio

import (
    "client"
    "fmt"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "github.com/xtaci/gaio"
    "net"
    "strconv"
)

type wsc_ struct {
    client.Wsc
    handler base.AsyncConnHandler
    w       *gaio.Watcher
    conn    net.Conn
}

func (w *wsc_) SetAsyncCallback(handler base.AsyncConnHandler) error {
    w.handler = handler
    return nil
}

func (w *wsc_) AsyncWrite(bytes []byte) error {
    return w.w.Write(w, w.conn, bytes)
}

func (w *wsc_) AsyncClose() error {
    return w.conn.Close()
}
func echoServer(w *gaio.Watcher) {
    for {
        results, err := w.WaitIO()
        if err != nil {
            fmt.Println(err)
            return
        }
        for _, res := range results {
            switch res.Operation {
            case gaio.OpRead:
                wsc := res.Context.(*wsc_)
                if res.Error == nil {
                    err := wsc.handler.OnData(res.Buffer[:res.Size])
                    if err != nil {
                        wsc.handler.OnClose(res.Error)
                        return
                    }
                    //触发读事件后,可能没有读取完成,声明再次读取
                    err = w.Read(wsc, res.Conn, nil)
                    if err != nil {
                        wsc.handler.OnClose(err)
                        return
                    }
                } else {
                    wsc.handler.OnClose(res.Error)
                }
            case gaio.OpWrite:
                if res.Error != nil {
                    res.Context.(*wsc_).handler.OnClose(res.Error)
                    return
                } else {
                    wsc := res.Context.(*wsc_)
                    //触发读事件后,可能没有读取完成,声明再次读取
                    err = w.Read(wsc, res.Conn, nil)
                    if err != nil {
                        wsc.handler.OnClose(err)
                        return
                    }
                }
            }
        }
    }
}
func Run(port int) func() {
    ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
    if err != nil {
        panic(err)
    }
    w, err := gaio.NewWatcher()
    if err != nil {
        panic(err)
    }
    go echoServer(w)
    go func() {
        for {
            conn, err := ln.Accept()
            if err != nil {
                return
            }
            handler := &wsc_{conn: conn, w: w}
            ws, err := websocket.NewServerWs(handler, handler)
            handler.Wsc = client.NewWsc(ws)
            err = w.Read(handler, conn, nil)
            if err != nil {
                return
            }
        }
    }()
    return func() { w.Close(); ln.Close() }
}
