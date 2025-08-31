package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOpts{ListenAddr: ":0", HandshakeFunc: NOPHandshakeFunc}
	tr := NewTCPTransport(opts)
	assert.NotNil(t, tr)
	assert.Nil(t, tr.ListenAndAccept())
	assert.Nil(t, tr.Close())
}
