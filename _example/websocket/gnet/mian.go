package main

import (
    "client"
    "fmt"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "github.com/panjf2000/gnet/v2"
    "math/rand"
    "os"
    "strconv"
    "time"
)

type wsc_ struct {
    client.Wsc
    conn gnet.Conn
}

func (w *wsc_) SetAsyncCallback(handler base.AsyncConnHandler) error {
    w.conn.SetContext(handler)
    return nil
}

func (w *wsc_) AsyncWrite(bytes []byte) error {
    return w.conn.AsyncWrite(bytes, nil)
}

func (w *wsc_) AsyncClose() error {
    return w.conn.CloseWithCallback(nil)
}

type server struct {
    gnet.BuiltinEventEngine
}

func (s *server) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
    handler := &wsc_{conn: c}
    w, err := websocket.NewServerWs(handler, handler)
    handler.Ws = w
    if err != nil {
        return nil, gnet.Close
    }
    return nil, gnet.None
}

func (s *server) OnTraffic(c gnet.Conn) (action gnet.Action) {
    buf, err := c.Next(-1)
    if err != nil {
        panic(err)
    }
    if w, ok := c.Context().(base.AsyncConnHandler); ok {
        err = w.OnData(buf)
        if err != nil {
            fmt.Sprintln(err)
            return gnet.Close
        }
    } else {
        panic("no handler")
        return gnet.Close
    }
    return gnet.None
}

func (s *server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
    if w, ok := c.Context().(base.AsyncConnHandler); ok {
        w.OnClose(err)
    }
    return gnet.None
}

func main() {
    port := 20000 + rand.Intn(9999)
    go func() {
        time.Sleep(10 * time.Second)
        os.Exit(0)
    }()
    client.Run(port, 5)
    fmt.Println("start on port", port)
    gnet.Run(&server{}, ":"+strconv.Itoa(port))
}
