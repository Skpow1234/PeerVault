package p2p

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOpts{ListenAddr: ":0", HandshakeFunc: NOPHandshakeFunc}
	tr := NewTCPTransport(opts)
	assert.NotNil(t, tr)
	assert.Nil(t, tr.ListenAndAccept())

	// Give the accept loop time to start
	time.Sleep(10 * time.Millisecond)

	// Close the transport
	assert.Nil(t, tr.Close())

	// Give the accept loop time to stop gracefully
	time.Sleep(10 * time.Millisecond)
}
