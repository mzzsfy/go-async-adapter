package base

type AsyncConnHandler interface {
    OnData([]byte) error
    OnClose(error)
}

type AsyncConnect interface {
    SetAsyncCallback(AsyncConnHandler) error
    AsyncWrite([]byte) error
    AsyncClose() error
}
