package main

import (
	"bytes"
	"net"
	"os"
	"testing"
	"time"

	netp2p "github.com/anthdm/foreverstore/internal/transport/p2p"
)

func TestAuthenticatedHandshake(t *testing.T) {
	// Set up test auth token
	os.Setenv("PEERVAULT_AUTH_TOKEN", "test-auth-token-123")
	defer os.Unsetenv("PEERVAULT_AUTH_TOKEN")

	// Create a test listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	// Get the listener address
	listenAddr := listener.Addr().String()

	// Create handshake functions for both sides
	nodeID1 := "test-node-1"
	nodeID2 := "test-node-2"
	handshake1 := netp2p.AuthenticatedHandshakeFunc(nodeID1)
	handshake2 := netp2p.AuthenticatedHandshakeFunc(nodeID2)

	// Start server goroutine
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		// Create peer and perform handshake
		peer := netp2p.NewTCPPeer(conn, false)
		if err := handshake2(peer); err != nil {
			t.Errorf("server handshake failed: %v", err)
		}
	}()

	// Connect client
	conn, err := net.Dial("tcp", listenAddr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Create peer and perform handshake
	peer := netp2p.NewTCPPeer(conn, true)
	if err := handshake1(peer); err != nil {
		t.Fatalf("client handshake failed: %v", err)
	}

	// Give some time for the server to complete
	time.Sleep(100 * time.Millisecond)
}

func TestHandshakeMessageSerialization(t *testing.T) {
	// Test message serialization/deserialization
	msg := netp2p.HandshakeMessage{
		NodeID:    "test-node",
		Timestamp: 1234567890,
		Signature: []byte{1, 2, 3, 4, 5},
	}

	// Serialize
	data := netp2p.SerializeHandshakeMessage(msg)

	// Deserialize
	deserialized, err := netp2p.DeserializeHandshakeMessage(data)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	// Verify
	if deserialized.NodeID != msg.NodeID {
		t.Errorf("NodeID mismatch: expected %s, got %s", msg.NodeID, deserialized.NodeID)
	}
	if deserialized.Timestamp != msg.Timestamp {
		t.Errorf("Timestamp mismatch: expected %d, got %d", msg.Timestamp, deserialized.Timestamp)
	}
	if !bytes.Equal(deserialized.Signature, msg.Signature) {
		t.Errorf("Signature mismatch")
	}
}

func TestHandshakeSignature(t *testing.T) {
	authToken := "test-token"
	msg := netp2p.HandshakeMessage{
		NodeID:    "test-node",
		Timestamp: time.Now().Unix(),
	}

	// Sign the message
	msg.Signature = netp2p.SignHandshakeMessage(msg, authToken)

	// Verify the signature
	if !netp2p.VerifyHandshakeMessage(msg, authToken) {
		t.Error("signature verification failed")
	}

	// Test with wrong token
	if netp2p.VerifyHandshakeMessage(msg, "wrong-token") {
		t.Error("signature verification should fail with wrong token")
	}
}
