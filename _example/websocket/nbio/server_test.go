package nbio

import (
    "client"
    "math/rand"
    "testing"
)

func TestRun(t *testing.T) {
    port := 20000 + rand.Intn(9999)
    f := Run(port)
    client.Run(port, 5)
    f()
}
