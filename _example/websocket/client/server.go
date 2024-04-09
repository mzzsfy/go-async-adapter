package client

import (
    "fmt"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "strconv"
    "time"
)

type Wsc struct {
    websocket.DoNothingHandler
    Ws websocket.AsyncWebsocket
}

func (w *Wsc) OnControlMessage(message websocket.Message) (bool, error) {
    fmt.Println("OnControlMessage", message.OpCode, string(message.Payload))
    return false, nil
}

func (w *Wsc) OnMessage(message websocket.Message) error {
    fmt.Println("OnMessage", message.OpCode, string(message.Payload))
    return nil
}

func (w *Wsc) OnUpgrade(info websocket.UpgradeInfo) error {
    p := string(info.Params())
    go func() {
        time.Sleep(100 * time.Millisecond)
        for i := 0; i < 6; i++ {
            time.Sleep(time.Duration(i*120) * time.Millisecond)
            w.Ws.Ping()
        }
    }()
    go func() {
        time.Sleep(100 * time.Millisecond)
        for i := 0; i < 7; i++ {
            time.Sleep(time.Duration(i*88) * time.Millisecond)
            w.Ws.Send(&websocket.SendMessage{Data: []byte(strconv.Itoa(i) + ": hello,client! " + p)})
        }
    }()
    fmt.Println("OnUpgrade", string(info.Path()), p, info.Headers())
    return nil
}
