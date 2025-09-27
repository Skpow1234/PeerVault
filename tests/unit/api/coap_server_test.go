package api_test

import (
	"testing"

	"github.com/Skpow1234/Peervault/internal/api/coap"
	"github.com/stretchr/testify/assert"
)

func TestServerConfig(t *testing.T) {
	// Test server configuration
	config := &coap.ServerConfig{
		Port:           5683,
		Host:           "localhost",
		MaxConnections: 100,
		MaxMessageSize: 1024,
		BlockSize:      512,
	}

	assert.Equal(t, 5683, config.Port)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 100, config.MaxConnections)
	assert.Equal(t, 1024, config.MaxMessageSize)
	assert.Equal(t, 512, config.BlockSize)
}

func TestServerStruct(t *testing.T) {
	// Test server struct creation
	config := &coap.ServerConfig{
		Port: 5683,
		Host: "localhost",
	}

	server := &coap.Server{
		Config:    config,
		Resources: make(map[string]*coap.Resource),
		Observers: make(map[string][]*coap.Observer),
		Clients:   make(map[string]*coap.Client),
		Stats:     &coap.ServerStats{},
	}

	assert.NotNil(t, server)
	assert.Equal(t, config, server.Config)
	assert.NotNil(t, server.Resources)
	assert.NotNil(t, server.Observers)
	assert.NotNil(t, server.Clients)
	assert.NotNil(t, server.Stats)
}
