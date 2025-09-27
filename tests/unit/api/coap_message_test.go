package api_test

import (
	"testing"

	"github.com/Skpow1234/Peervault/internal/api/coap"
	"github.com/stretchr/testify/assert"
)

func TestMessageTypeConstants(t *testing.T) {
	// Test CoAP message type constants
	assert.Equal(t, coap.MessageType(0), coap.Confirmable)
	assert.Equal(t, coap.MessageType(1), coap.NonConfirmable)
	assert.Equal(t, coap.MessageType(2), coap.Acknowledgement)
	assert.Equal(t, coap.MessageType(3), coap.Reset)
}

func TestMethodCodeConstants(t *testing.T) {
	// Test CoAP method code constants
	assert.Equal(t, coap.MethodCode(1), coap.GET)
	assert.Equal(t, coap.MethodCode(2), coap.POST)
	assert.Equal(t, coap.MethodCode(3), coap.PUT)
	assert.Equal(t, coap.MethodCode(4), coap.DELETE)
}

func TestResponseCodeConstants(t *testing.T) {
	// Test CoAP response code constants
	// Success responses
	assert.Equal(t, coap.ResponseCode(65), coap.Created)
	assert.Equal(t, coap.ResponseCode(66), coap.Deleted)
	assert.Equal(t, coap.ResponseCode(67), coap.Valid)
	assert.Equal(t, coap.ResponseCode(68), coap.Changed)
	assert.Equal(t, coap.ResponseCode(69), coap.Content)

	// Client error responses
	assert.Equal(t, coap.ResponseCode(128), coap.BadRequest)
	assert.Equal(t, coap.ResponseCode(129), coap.Unauthorized)
	assert.Equal(t, coap.ResponseCode(130), coap.BadOption)
	assert.Equal(t, coap.ResponseCode(131), coap.Forbidden)
	assert.Equal(t, coap.ResponseCode(132), coap.NotFound)
	assert.Equal(t, coap.ResponseCode(133), coap.MethodNotAllowed)
	assert.Equal(t, coap.ResponseCode(134), coap.NotAcceptable)
	assert.Equal(t, coap.ResponseCode(140), coap.PreconditionFailed)
	assert.Equal(t, coap.ResponseCode(141), coap.RequestEntityTooLarge)
	assert.Equal(t, coap.ResponseCode(143), coap.UnsupportedContentFormat)
}

func TestMessageTypeValues(t *testing.T) {
	// Test MessageType values
	assert.Equal(t, byte(0), byte(coap.Confirmable))
	assert.Equal(t, byte(1), byte(coap.NonConfirmable))
	assert.Equal(t, byte(2), byte(coap.Acknowledgement))
	assert.Equal(t, byte(3), byte(coap.Reset))
}

func TestMethodCodeValues(t *testing.T) {
	// Test MethodCode values
	assert.Equal(t, byte(1), byte(coap.GET))
	assert.Equal(t, byte(2), byte(coap.POST))
	assert.Equal(t, byte(3), byte(coap.PUT))
	assert.Equal(t, byte(4), byte(coap.DELETE))
}

func TestResponseCodeValues(t *testing.T) {
	// Test ResponseCode values
	assert.Equal(t, byte(65), byte(coap.Created))
	assert.Equal(t, byte(66), byte(coap.Deleted))
	assert.Equal(t, byte(67), byte(coap.Valid))
	assert.Equal(t, byte(68), byte(coap.Changed))
	assert.Equal(t, byte(69), byte(coap.Content))
	assert.Equal(t, byte(128), byte(coap.BadRequest))
	assert.Equal(t, byte(129), byte(coap.Unauthorized))
	assert.Equal(t, byte(130), byte(coap.BadOption))
	assert.Equal(t, byte(131), byte(coap.Forbidden))
	assert.Equal(t, byte(132), byte(coap.NotFound))
	assert.Equal(t, byte(133), byte(coap.MethodNotAllowed))
	assert.Equal(t, byte(134), byte(coap.NotAcceptable))
	assert.Equal(t, byte(140), byte(coap.PreconditionFailed))
	assert.Equal(t, byte(141), byte(coap.RequestEntityTooLarge))
	assert.Equal(t, byte(143), byte(coap.UnsupportedContentFormat))
}
