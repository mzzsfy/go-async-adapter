package nbio

import (
    "client"
    "fmt"
    "github.com/lesismal/nbio"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "strconv"
)

type wsc_ struct {
    client.Wsc
    conn    *nbio.Conn
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
    engine := nbio.NewEngine(nbio.Config{
        Network: "tcp",
        Addrs:   []string{":" + strconv.Itoa(port)},
    })
    engine.OnOpen(func(c *nbio.Conn) {
        handler := &wsc_{conn: c}
        ws, err := websocket.NewServerWs(handler, handler)
        handler.Wsc = client.NewWsc(ws)
        if err != nil {
            fmt.Println("websocket.NewServerWs failed:", err)
            return
        }
        c.SetSession(handler)
    })
    engine.OnClose(func(c *nbio.Conn, err error) {
        c.Session().(*wsc_).handler.OnClose(err)
    })
    engine.OnData(func(c *nbio.Conn, data []byte) {
        c.Session().(*wsc_).handler.OnData(data)
    })
    err := engine.Start()
    if err != nil {
        panic(err)
    }
    return engine.Stop
}
