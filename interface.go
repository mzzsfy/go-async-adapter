package fastws

type AsyncConnHandler interface {
    OnData([]byte) error
    OnClose(error)
}

type AsyncConn interface {
    SetCallback(AsyncConnHandler) error
    AsyncWrite([]byte) error
    Close() error
}

type UpgradeInfo interface {
    Path() []byte
    Param([]byte) []byte
    Params() []byte
    Header([]byte) []byte
    Headers() [][]byte //[]{[]{"host: xxxx"}}
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

type AsyncWs interface {
    Send(*SendMessage) error
    // Close code含义:https://datatracker.ietf.org/doc/html/rfc6455#section-7.4.1
    Close(closeCode uint16, data string) error
    Start(AsyncConn) error
}