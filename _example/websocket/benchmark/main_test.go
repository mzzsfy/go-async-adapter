package benchmark

import (
    "bufio"
    "bytes"
    "client"
    "evio"
    "github.com/gobwas/ws"
    "github.com/gobwas/ws/wsutil"
    "github.com/mzzsfy/go-async-adapter/websocket"
    "gnet"
    "io"
    "math/rand"
    "nbio"
    "net"
    "net/url"
    "netpoll"
    "strconv"
    "strings"
    "sync"
    "testing"
    "time"
)

func Run(t interface{ Error(...any) }, port, clientNum, sendEventNum, startWait, sendWait int) {
    wg := sync.WaitGroup{}
    if sendEventNum < 100 {
        clientNum = 1
    } else {
        sendEventNum1 := sendEventNum / clientNum
        if sendEventNum1 <= 10 {
            sendEventNum = 10
        } else {
            sendEventNum = sendEventNum1
        }
    }
    wg.Add(clientNum)
    for c := 0; c < clientNum; c++ {
        c := c
        go func() {
            defer wg.Done()
            dialer := ws.Dialer{Timeout: 10 * time.Second}
            conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(port))
            if err != nil {
                t.Error(c, "拨号失败", err)
                return
            }
            {
                url1, _ := url.Parse("ws://localhost:" + strconv.Itoa(port) + "/test/Ws?test=1&test1=2&client=" + strconv.Itoa(c))
                r := &bytes.Buffer{}
                r1 := io.TeeReader(conn, r)
                rw := bufio.NewReadWriter(bufio.NewReader(r1), bufio.NewWriterSize(conn, 1))
                _, _, err := dialer.Upgrade(rw, url1)
                if err != nil {
                    t.Error(c, "创建连接失败", err)
                    return
                }
            }
            go func() {
                for {
                    _, err := wsutil.ReadServerMessage(conn, nil)
                    if err != nil {
                        if !strings.Contains(err.Error(), "use of closed network connection") {
                            t.Error(c, "read error: ", err)
                        }
                        return
                    }
                }
            }()
            for i := 0; i < sendEventNum; i++ {
                buffer := &bytes.Buffer{}
                if i%2 == 0 {
                    wsutil.WriteClientMessage(buffer, ws.OpText, []byte(strconv.Itoa(c)+":hello server,text! "+strconv.Itoa(i)))
                } else {
                    wsutil.WriteClientMessage(buffer, ws.OpBinary, []byte(strconv.Itoa(c)+":hello server,binary! "+strconv.Itoa(i)))
                }
                _, err = conn.Write(buffer.Bytes())
                if err != nil {
                    t.Error(c, "error:", err)
                    return
                }
                time.Sleep(time.Duration(sendWait) * time.Millisecond)
            }
            conn.Close()
        }()
        time.Sleep(time.Duration(startWait) * time.Millisecond)
    }
    wg.Wait()
}

var nowRun string

func BenchmarkEcho(b *testing.B) {
    client.NewWsc = func(ws websocket.AsyncWebsocket) client.Wsc {
        c := &client.WscEcho{Ws: ws}
        if nowRun == "netpoll" {
            c.NoCopy = true
        }
        return c
    }
    port := 30000 + rand.Intn(9999)
    clientNum := 2
    startWait := 0
    sendWait := 0
    port++
    run(b, "evio", evio.Run, port, clientNum, startWait, sendWait)
    //port++
    //run(b, "gaio", gaio.Run, port, clientNum, startWait, sendWait)
    port++
    run(b, "gnet", gnet.Run, port, clientNum, startWait, sendWait)
    port++
    run(b, "nbio", nbio.Run, port, clientNum, startWait, sendWait)
    port++
    run(b, "netpoll", netpoll.Run, port, clientNum, startWait, sendWait)
}

type log struct {
    name string
    l    interface {
        Helper()
        Error(...any)
    }
}

func (l log) Error(v ...any) {
    l.l.Helper()
    l.l.Error(append([]any{time.Now().Format("05.000"), l.name}, v...)...)
}

func run(b *testing.B, name string, run func(port int) func(), port int, clientNum int, startWait int, sendWait int) {
    nowRun = name
    f := run(port)
    defer f()
    time.Sleep(100 * time.Millisecond)
    b.Run(name, func(b *testing.B) {
        b.ResetTimer()
        Run(log{
            name: name,
            l:    b,
        }, port, clientNum, b.N, startWait, sendWait)
    })
}
