package client

import (
    "bytes"
    "context"
    "fmt"
    "github.com/gobwas/ws"
    "github.com/gobwas/ws/wsutil"
    "math/rand"
    "strconv"
    "strings"
    "sync"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixMilli())
}

func Run(port, num int) {
    wg := sync.WaitGroup{}
    wg.Add(num)
    for c := 0; c < num; c++ {
        c := c
        go func() {
            defer wg.Done()
            time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
            conn, _, _, err := ws.Dialer{Timeout: 5 * time.Second}.Dial(context.Background(), "Ws://127.0.0.1:"+strconv.Itoa(port)+"/test/Ws?test=1&test1=2&client="+strconv.Itoa(c))
            if err != nil {
                fmt.Println(c, "创建连接失败", err)
                return
            }
            defer conn.Close()
            go func() {
                for {
                    message, err := wsutil.ReadServerMessage(conn, nil)
                    if err != nil {
                        if !strings.Contains(err.Error(), "use of closed network connection") {
                            fmt.Println(c, "read error: ", err)
                        }
                        return
                    } else {
                        for _, m := range message {
                            if m.OpCode == ws.OpPing {
                                fmt.Println(c, "recv", "ping")
                                wsutil.WriteClientMessage(conn, ws.OpPong, nil)
                            } else {
                                fmt.Println(c, "recv", m.OpCode, string(m.Payload))
                            }
                        }
                    }
                }
            }()
            for i := 0; i < 10; i++ {
                buffer := &bytes.Buffer{}
                if i%2 == 0 {
                    fmt.Println(c, "send", i, "text")
                    wsutil.WriteClientMessage(buffer, ws.OpText, []byte(strconv.Itoa(c)+":hello server,text! "+strconv.Itoa(i)))
                } else {
                    fmt.Println(c, "send", i, "binary")
                    wsutil.WriteClientMessage(buffer, ws.OpBinary, []byte(strconv.Itoa(c)+":hello server,binary! "+strconv.Itoa(i)))
                }
                _, err = conn.Write(buffer.Bytes())
                if err != nil {
                    fmt.Println(c, "error:", err)
                    return
                }
                time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
            }
        }()
    }
    wg.Wait()
}
