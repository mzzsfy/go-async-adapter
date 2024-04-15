package evio

import (
    "client"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "math/rand"
    "testing"
)

func TestRun(t *testing.T) {
    client.NewWsc = func(ws websocket.AsyncWebsocket) client.Wsc {
        c := &client.WscEcho{Ws: ws}
        return c
    }
    port := 20000 + rand.Intn(9999)
    f := Run(port)
    client.Run(port, 5)
    f()
}
