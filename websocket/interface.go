package websocket

import "bytes"

type UpgradeInfo interface {
    Path() []byte
    Param([]byte) []byte
    Params() []byte
    Header([]byte) []byte
    Headers() [][]byte //[]{"name","value","name1","value1"}
}

type upgradeInfo struct {
    path    []byte
    params  []byte
    headers [][]byte
}

func (u upgradeInfo) Path() []byte {
    u.init()
    return u.path
}

var (
    symbolAnd      = []byte("&")
    symbolEq       = []byte("=")
    symbolHost     = []byte("Host")
    symbolQuestion = []byte("?")
    symbolSlash    = []byte("/")
)

func (u upgradeInfo) Param(name []byte) []byte {
    u.init()
    for _, b := range bytes.Split(u.params, symbolAnd) {
        n := bytes.SplitN(b, symbolEq, 2)
        if bytes.Equal(n[0], name) {
            return n[1]
        }
    }
    return nil
}

func (u upgradeInfo) Params() []byte {
    u.init()
    return u.params
}

func (u upgradeInfo) Header(name []byte) []byte {
    u.init()
    for i := 0; i < len(u.headers); i += 2 {
        if bytes.Equal(u.headers[i], name) {
            return u.headers[i+1]
        }
    }
    return nil
}

func (u upgradeInfo) Headers() [][]byte {
    u.init()
    if len(u.headers) == 0 {
        return nil
    }
    return u.headers
}

func (u upgradeInfo) AddHeader(k, v []byte) {
    u.headers = append(u.headers, k, v)
}

func (u upgradeInfo) init() {
}

type OpCode byte

const (
    OpContinuation OpCode = 0x0
    OpText         OpCode = 0x1
    OpBinary       OpCode = 0x2
    OpClose        OpCode = 0x8
    OpPing         OpCode = 0x9
    OpPong         OpCode = 0xa
)

type Message struct {
    OpCode  OpCode
    Payload []byte
}

type MessageHandler interface {
    OnMessage(Message) error
    OnUpgrade(UpgradeInfo) error
    // OnClose code含义:https://datatracker.ietf.org/doc/html/rfc6455#section-7.4.1
    OnClose(closeCode uint16, data string)
    // OnError 非websocket错误,如连接中断,数据异常等,触发这个回调后还会触发 OnClose(1006,error.Error())
    OnError(error)
}
type ControlMessageHandler interface {
    MessageHandler
    // OnControlMessage 自定义处理控制消息,返回true表示已自己处理完成,否则将执行默认逻辑
    OnControlMessage(Message) (bool, error)
}

type DoNothingHandler struct{}

func (d DoNothingHandler) OnMessage(message Message) error       { return nil }
func (d DoNothingHandler) OnUpgrade(info UpgradeInfo) error      { return nil }
func (d DoNothingHandler) OnClose(closeCode uint16, data string) {}
func (d DoNothingHandler) OnError(error)                         {}

type SendMessage struct {
    // IsBinary 是否为二进制数据
    IsBinary bool
    // EnableCache 是否启用缓存,服务端给大量用户发送相同消息时可以有效减少内存分配
    EnableCache bool
    Data        []byte
    cache       []byte
}

type AsyncWebsocket interface {
    Send(*SendMessage) error
    Ping() error
    // Close code含义:https://datatracker.ietf.org/doc/html/rfc6455#section-7.4.1
    Close(closeCode uint16, data string) error
}
