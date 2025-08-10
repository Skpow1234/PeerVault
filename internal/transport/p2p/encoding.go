package p2p

import (
    "encoding/gob"
    "io"
)

const (
    IncomingMessage = 0x1
    IncomingStream  = 0x2
)

// RPC holds any arbitrary data that is being sent over the
// each transport between two nodes in the network.
type RPC struct {
    From    string
    Payload []byte
    Stream  bool
}

type Decoder interface {
    Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
    return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
    peekBuf := make([]byte, 1)
    if _, err := r.Read(peekBuf); err != nil {
        return nil
    }

    // In case of a stream we are not decoding what is being sent over the network.
    // We are just setting Stream true so we can handle that in our logic.
    stream := peekBuf[0] == IncomingStream
    if stream {
        msg.Stream = true
        return nil
    }

    buf := make([]byte, 1028)
    n, err := r.Read(buf)
    if err != nil {
        return err
    }

    msg.Payload = buf[:n]
    return nil
}


