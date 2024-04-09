package main

import (
    "client"
    "context"
    "errors"
    "github.com/cloudwego/netpoll"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "math/rand"
    "net"
    "os"
    "strconv"
    "time"
)

type wsc_ struct {
    client.Wsc
    conn    netpoll.Connection
    handler base.AsyncConnHandler
}

func (w *wsc_) SetAsyncCallback(handler base.AsyncConnHandler) error {
    w.handler = handler
    return nil
}

func (w *wsc_) AsyncWrite(bytes []byte) error {
    _, err := w.conn.Write(bytes)
    return err
}

func (w *wsc_) AsyncClose() error {
    return w.conn.Close()
}

func main() {
    port := 20000 + rand.Intn(9999)
    go func() {
        time.Sleep(10 * time.Second)
        os.Exit(0)
    }()
    client.Run(port, 5)
    listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
    if err != nil {
        panic("create netpoll listener failed")
    }
    loop, err := netpoll.NewEventLoop(func(ctx context.Context, connection netpoll.Connection) error {
        handler := ctx.Value("handler").(base.AsyncConnHandler)
        reader := connection.Reader()
        defer reader.Release()
        p, err := reader.Next(reader.Len())
        if err != nil {
            handler.OnClose(err)
            return nil
        }
        err = handler.OnData(p)
        if err != nil {
            handler.OnClose(err)
        }
        return err
    }, netpoll.WithOnPrepare(func(connection netpoll.Connection) context.Context {
        wsc := &wsc_{conn: connection}
        ws, _ := websocket.NewServerWs(wsc, wsc)
        connection.AddCloseCallback(func(connection netpoll.Connection) error {
            wsc.handler.OnClose(errors.New("close"))
            return nil
        })
        wsc.Ws = ws
        return context.WithValue(context.Background(), "handler", wsc.handler)
    }))
    if err != nil {
        panic("create netpoll event loop failed")
    }
    client.Run(port, 5)
    loop.Serve(listener)
}
