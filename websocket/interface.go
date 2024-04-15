package websocket

type OpCode byte

var (
    symbolAnd      = []byte("&")
    symbolEq       = []byte("=")
    symbolQuestion = []byte("?")
    symbolSlash    = []byte("/")
    symbolHost     = []byte("Host")
)

const (
    OpContinuation OpCode = 0x0
    OpText         OpCode = 0x1
    OpBinary       OpCode = 0x2
    OpClose        OpCode = 0x8
    OpPing         OpCode = 0x9
    OpPong         OpCode = 0xa
)

type UpgradeType int8

const (
    UpgradeTypeHeader UpgradeType = iota + 1
    UpgradeTypePath
    UpgradeTypeQuery
    UpgradeTypeProtocol  //todo:支持 Sec-WebSocket-Protocol
    UpgradeTypeExtension //todo:支持 Sec-WebSocket-Extensions
)

type Message struct {
    OpCode  OpCode
    Payload []byte
}

type DataWrapper interface {
    // Name 获取包装名称,如: flate 或 flate,xxx,不同协议不要重复,否则使用缓存会导致数据错乱
    Name() string
    // ReadData 解码数据
    ReadData([]byte) []byte
    // WritData 编码数据
    WritData([]byte) []byte
}

type UpgradeHandler interface {
    // On 处理升级参数,data为完整数据,如:
    //UpgradeTypeHeader: []byte("Host: host")
    //UpgradeTypeQuery: []byte("a=1&b=2&c&d")
    //UpgradeTypeProtocol: []byte("aaa,bbb;ccc")
    On(UpgradeType UpgradeType, data []byte) error
    // CheckUpgrade 参数回调完成,最后的检验
    CheckUpgrade() error
    // ResponseHeader 添加响应头,会被多次调用,直到返回name为nil
    ResponseHeader() (name, value []byte)
    // DataWrapper 提供数据传输压缩等功能支持
    DataWrapper() DataWrapper
}

type DoNothingUpgrade struct{}

func (d DoNothingUpgrade) On(UpgradeType, []byte) error     { return nil }
func (d DoNothingUpgrade) CheckUpgrade() error              { return nil }
func (d DoNothingUpgrade) ResponseHeader() ([]byte, []byte) { return nil, nil }
func (d DoNothingUpgrade) DataWrapper() DataWrapper         { return nil }

type MessageHandler interface {
    OnMessage(Message) error
    OnUpgrade() UpgradeHandler
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

func (d DoNothingHandler) OnMessage(Message) error   { return nil }
func (d DoNothingHandler) OnUpgrade() UpgradeHandler { return nil }
func (d DoNothingHandler) OnClose(uint16, string)    {}
func (d DoNothingHandler) OnError(error)             {}

type SendMessage struct {
    // IsBinary 是否为二进制数据
    IsBinary bool
    // EnableCache 是否启用缓存,服务端给大量用户发送相同消息时可以有效减少内存分配
    EnableCache bool
    Data        []byte
    cache       []struct {
        name string
        data []byte
    }
}

type AsyncWebsocket interface {
    Send(*SendMessage) error
    Ping() error
    // Close code含义:https://datatracker.ietf.org/doc/html/rfc6455#section-7.4.1
    Close(closeCode uint16, data string) error
}
