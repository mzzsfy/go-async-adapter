package netpoll

import (
    "client"
    "context"
    "errors"
    "github.com/cloudwego/netpoll"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "net"
    "strconv"
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

func Run(port int) func() {
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
        wsc.Wsc = client.NewWsc(ws)
        return context.WithValue(context.Background(), "handler", wsc.handler)
    }))
    if err != nil {
        panic("create netpoll event loop failed")
    }
    go loop.Serve(listener)
    return func() {
        loop.Shutdown(context.Background())
    }
}
