package client

import (
    "fmt"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "strconv"
    "time"
)

type Wsc websocket.ControlMessageHandler

var NewWsc = func(ws websocket.AsyncWebsocket) Wsc {
    return &WscPrint{Ws: ws}
}

type WscPrint struct {
    websocket.DoNothingHandler
    Ws websocket.AsyncWebsocket
}

func (w *WscPrint) OnControlMessage(message websocket.Message) (bool, error) {
    fmt.Println("OnControlMessage", message.OpCode, string(message.Payload))
    return false, nil
}

func (w *WscPrint) OnMessage(message websocket.Message) error {
    fmt.Println("OnMessage", message.OpCode, string(message.Payload))
    return nil
}

func (w *WscPrint) OnUpgrade() websocket.UpgradeHandler {
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
            w.Ws.Send(&websocket.SendMessage{Data: []byte(strconv.Itoa(i) + ": hello,client! ")})
        }
    }()
    return nil
}

type WscEcho struct {
    NoCopy bool
    websocket.DoNothingHandler
    Ws websocket.AsyncWebsocket
}

func (w *WscEcho) OnControlMessage(_ websocket.Message) (bool, error) {
    return false, nil
}

func (w *WscEcho) OnMessage(message websocket.Message) error {
    if w.NoCopy {
        return w.Ws.Send(&websocket.SendMessage{Data: message.Payload})
    }
    bs := make([]byte, len(message.Payload))
    copy(bs, message.Payload)
    return w.Ws.Send(&websocket.SendMessage{Data: bs})
}

func (w *WscEcho) OnUpgrade() websocket.UpgradeHandler {
    return nil
}
