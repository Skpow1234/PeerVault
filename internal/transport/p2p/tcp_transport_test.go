package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOpts{ListenAddr: ":0", HandshakeFunc: NOPHandshakeFunc, Decoder: DefaultDecoder{}}
	tr := NewTCPTransport(opts)
	assert.NotNil(t, tr)
	assert.Nil(t, tr.ListenAndAccept())
	assert.Nil(t, tr.Close())
}
