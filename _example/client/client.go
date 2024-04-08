package client

import (
    "bytes"
    "context"
    "fmt"
    "github.com/gobwas/ws"
    "github.com/gobwas/ws/wsutil"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "math/rand"
    "strconv"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixMilli())
}

func Run(port, num int) {
    for c := 0; c < num; c++ {
        c := c
        go func() {
            time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
            conn, _, _, err := ws.Dialer{Timeout: 5 * time.Second}.Dial(context.Background(), "Ws://127.0.0.1:"+strconv.Itoa(port)+"/test/Ws?test=1&test1=2&client="+strconv.Itoa(c))
            if err != nil {
                fmt.Println(c, "创建连接失败", err)
                return
            }
            message, err := wsutil.ReadServerMessage(conn, nil)
            if err != nil {
                fmt.Println(c, "read error: ", err)
            } else {
                fmt.Println(c, "recv", message[0].OpCode, string(message[0].Payload))
            }
            for i := 0; i < 10; i++ {
                buffer := &bytes.Buffer{}
                if i%2 == 0 {
                    fmt.Println(c, "send", i, "text")
                    wsutil.WriteClientMessage(buffer, ws.OpText, []byte(strconv.Itoa(c)+":hello,text! "+strconv.Itoa(i)))
                } else {
                    fmt.Println(c, "send", i, "binary")
                    wsutil.WriteClientMessage(buffer, ws.OpBinary, []byte(strconv.Itoa(c)+":hello,text! "+strconv.Itoa(i)))
                }
                _, err = conn.Write(buffer.Bytes())
                if err != nil {
                    fmt.Println(c, "error:", err)
                    return
                }
                time.Sleep(time.Millisecond * 300)
            }
            fmt.Println(c, "over")
            time.Sleep(time.Second * 5)
        }()
    }
}

type Wsc struct {
    websocket.DoNothingHandler
    Ws websocket.AsyncWebsocket
}

func (w *Wsc) OnMessage(message websocket.Message) error {
    fmt.Println("OnMessage", message.OpCode, string(message.Payload))
    return nil
}

func (w *Wsc) OnUpgrade(info websocket.UpgradeInfo) error {
    p := string(info.Params())
    go func() {
        time.Sleep(100 * time.Millisecond)
        w.Ws.Send(&websocket.SendMessage{Data: []byte("hello,client! " + p)})
    }()
    fmt.Println("OnUpgrade", string(info.Path()), p, info.Headers())
    return nil
}
