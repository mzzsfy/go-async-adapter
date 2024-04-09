package main

import (
    "bytes"
    "client"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "github.com/tidwall/evio"
    "math/rand"
    "os"
    "strconv"
    "time"
)

type wsc_ struct {
    client.Wsc
    conn    evio.Conn
    handler base.AsyncConnHandler
    writer  bytes.Buffer
    close   bool
}

func (w *wsc_) SetAsyncCallback(handler base.AsyncConnHandler) error {
    w.handler = handler
    return nil
}

func (w *wsc_) AsyncWrite(bytes []byte) error {
    w.writer.Write(bytes)
    w.conn.Wake()
    return nil
}

func (w *wsc_) AsyncClose() error {
    w.close = true
    w.conn.Wake()
    return nil
}

func main() {
    var events evio.Events
    port := 20000 + rand.Intn(9999)
    events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
        var h *wsc_
        var ok bool
        if h, ok = c.Context().(*wsc_); ok {
            h.conn = c
        } else {
            h = &wsc_{conn: c}
            ws, err := websocket.NewServerWs(h, h)
            h.Ws = ws
            if err != nil {
                return nil, evio.Close
            }
            c.SetContext(h)
        }
        err := h.handler.OnData(in)
        if err != nil {
            return nil, evio.Close
        }
        if h.writer.Len() > 0 {
            out = h.writer.Bytes()
            h.writer = bytes.Buffer{}
        }
        if h.close {
            action = evio.Close
        }
        return
    }
    go func() {
        time.Sleep(10 * time.Second)
        os.Exit(0)
    }()
    client.Run(port, 5)
    if err := evio.Serve(events, "tcp://:"+strconv.Itoa(port)); err != nil {
        panic(err.Error())
    }
}
