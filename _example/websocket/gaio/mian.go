package main

import (
    "client"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "github.com/xtaci/gaio"
    "log"
    "math/rand"
    "net"
    "os"
    "strconv"
    "time"
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
    return w.w.Write(nil, w.conn, bytes)
}

func (w *wsc_) AsyncClose() error {
    return w.conn.Close()
}
func echoServer(w *gaio.Watcher) {
    for {
        results, err := w.WaitIO()
        if err != nil {
            log.Println(err)
            return
        }
        for _, res := range results {
            switch res.Operation {
            case gaio.OpRead:
                if res.Error == nil {
                    err := res.Context.(*wsc_).handler.OnData(res.Buffer[:res.Size])
                    if err != nil {
                        res.Context.(*wsc_).handler.OnClose(res.Error)
                    }
                } else {
                    res.Context.(*wsc_).handler.OnClose(res.Error)
                }
            case gaio.OpWrite:
                if res.Error != nil {
                    res.Context.(*wsc_).handler.OnClose(res.Error)
                }
            }
        }
    }
}

func main() {
    port := 20000 + rand.Intn(9999)
    go func() {
        time.Sleep(10 * time.Second)
        os.Exit(0)
    }()
    ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
    if err != nil {
        log.Fatal(err)
    }
    w, err := gaio.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer w.Close()
    go echoServer(w)
    client.Run(port, 5)
    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            return
        }
        handler := &wsc_{conn: conn, w: w}
        ws, err := websocket.NewServerWs(handler, handler)
        handler.Ws = ws
        err = w.Read(handler, conn, make([]byte, 128))
        if err != nil {
            log.Println(err)
            return
        }
    }
}
