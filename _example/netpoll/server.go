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

type wsc struct {
    client.Wsc
    conn    netpoll.Connection
    handler base.AsyncConnHandler
}

func (w *wsc) SetAsyncCallback(handler base.AsyncConnHandler) error {
    w.handler = handler
    return nil
}

func (w *wsc) AsyncWrite(bytes []byte) error {
    _, err := w.conn.Writer().WriteBinary(bytes)
    return err
}

func (w *wsc) AsyncClose() error {
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
        handler := ctx.Value("handler").(*wsc).handler
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
        handler := &wsc{conn: connection}
        ws, _ := websocket.NewServerWs(handler, handler)
        connection.AddCloseCallback(func(connection netpoll.Connection) error {
            handler.handler.OnClose(errors.New("close"))
            return nil
        })
        handler.Ws = ws
        return context.WithValue(context.Background(), "handler", handler)
    }))
    if err != nil {
        panic("create netpoll event loop failed")
    }
    client.Run(port, 5)
    loop.Serve(listener)
}
