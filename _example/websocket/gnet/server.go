package gnet

import (
    "client"
    "context"
    "fmt"
    "github.com/mzzsfy/go-async-adapter/base"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "github.com/panjf2000/gnet/v2"
    "github.com/panjf2000/gnet/v2/pkg/logging"
    "strconv"
    "strings"
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
    return w.conn.AsyncWrite(bytes, func(c gnet.Conn, err error) error {
        if err != nil {
            fmt.Printf("server write error: %v\n", err)
        }
        return err
    })
}

func (w *wsc_) AsyncClose() error {
    return w.conn.CloseWithCallback(nil)
}

type server struct {
    gnet.BuiltinEventEngine
}

func (s *server) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
    handler := &wsc_{conn: c}
    ws, err := websocket.NewServerWs(handler, handler)
    handler.Wsc = client.NewWsc(ws)
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

type log__ struct {
}

func (l log__) Debugf(format string, args ...interface{}) {
}

func (l log__) Infof(format string, args ...interface{}) {
}

func (l log__) Warnf(format string, args ...interface{}) {
    if strings.HasPrefix(format, "error occurs in user-defined function") {
        return
    }
}

func (l log__) Errorf(format string, args ...interface{}) {
}

func (l log__) Fatalf(format string, args ...interface{}) {
}

func Run(port int) func() {
    logging.SetDefaultLoggerAndFlusher(log__{}, func() error { return nil })
    go gnet.Run(&server{}, ":"+strconv.Itoa(port))
    return func() {
        gnet.Stop(context.Background(), ":"+strconv.Itoa(port))
    }
}
