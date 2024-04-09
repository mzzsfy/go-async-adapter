package main

import (
    "client"
    "fmt"
    "github.com/lesismal/nbio"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "log"
    "math/rand"
    "os"
    "strconv"
    "time"
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

func main() {
    port := 20000 + rand.Intn(9999)
    engine := nbio.NewEngine(nbio.Config{
        Network: "tcp",
        Addrs:   []string{":" + strconv.Itoa(port)},
    })
    engine.OnOpen(func(c *nbio.Conn) {
        handler := &wsc_{conn: c}
        w, err := websocket.NewServerWs(handler, handler)
        handler.Ws = w
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
    go func() {
        time.Sleep(10 * time.Second)
        os.Exit(0)
    }()
    client.Run(port, 5)
    err := engine.Start()
    if err != nil {
        log.Fatalf("nbio.Start failed: %v\n", err)
        return
    }
    defer engine.Stop()
    select {}
}
