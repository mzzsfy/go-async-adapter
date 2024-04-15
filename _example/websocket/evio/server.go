package evio

import (
    "bytes"
    "client"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "github.com/tidwall/evio"
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

func Run(port int) func() {
    var events evio.Events
    events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
        var h *wsc_
        var ok bool
        if h, ok = c.Context().(*wsc_); ok {
            h.conn = c
        } else {
            h = &wsc_{conn: c}
            ws, err := websocket.NewServerWs(h, h)
            h.Wsc = client.NewWsc(ws)
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
    closed := false
    events.Tick = func() (delay time.Duration, action evio.Action) {
        if closed {
            return 0, evio.Close
        }
        return 100 * time.Millisecond, evio.None
    }
    go func() {
        if err := evio.Serve(events, "tcp://:"+strconv.Itoa(port)); err != nil {
            panic(err.Error())
        }
    }()
    return func() { closed = true }
}
