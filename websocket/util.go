package websocket

import (
    "encoding/binary"
    "github.com/gobwas/ws"
    "github.com/mzzsfy/go-async-adapter/base"
    "io"
)

const (
    len7  = int64(125)
    len16 = int64(^(uint16(0)))
    len64 = int64(^(uint64(0)) >> 1)
)

func wsHeadLength(header *ws.Header) int {
    n := 0
    switch {
    case header.Length <= len7:
        n = 2
    case header.Length <= len16:
        n = 4
    case header.Length <= len64:
        n = 10
    }
    if header.Masked {
        n += 4
    }
    return n
}

func writeWsHeader(bts []byte, h *ws.Header) {
    if h.Fin {
        bts[0] |= 0x80
    }
    bts[0] |= h.Rsv << 4
    bts[0] |= byte(h.OpCode)

    var n int
    switch {
    case h.Length <= len7:
        bts[1] = byte(h.Length)
        n = 2
    case h.Length <= len16:
        bts[1] = 126
        binary.BigEndian.PutUint16(bts[2:4], uint16(h.Length))
        n = 4

    case h.Length <= len64:
        bts[1] = 127
        binary.BigEndian.PutUint64(bts[2:10], uint64(h.Length))
        n = 10

    default:
        panic(ws.ErrHeaderLengthUnexpected)
    }

    if h.Masked {
        bts[1] |= 0x80
        n += copy(bts[n:], h.Mask[:])
    }
}

type rw_ struct {
    conn   base.AsyncConnect
    bufOld []byte
    bufNew []byte
    ri     int
}

func (r *rw_) Read(p []byte) (n int, err error) {
    if r.ri < len(r.bufOld) {
        n = copy(p, r.bufOld[r.ri:])
        r.ri += n
        if r.ri <= len(r.bufOld) {
            return
        }
    }
    if r.ri < len(r.bufOld)+len(r.bufNew) {
        i := copy(p[n:], r.bufNew[r.ri-len(r.bufOld):])
        n += i
        r.ri += i
    } else {
        return n, io.EOF
    }
    return
}

func (r *rw_) Write(p []byte) (n int, err error) {
    return len(p), r.conn.AsyncWrite(p)
}
