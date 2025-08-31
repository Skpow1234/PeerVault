package main

import (
	"bytes"
	"testing"

	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

func TestLengthPrefixedDecoder(t *testing.T) {
	// Test message decoding
	payload := []byte("test message")

	// Create a buffer with a properly framed message
	buf := new(bytes.Buffer)
	frameWriter := netp2p.NewFrameWriter(buf)
	if err := frameWriter.WriteMessage(payload); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Decode the message
	decoder := netp2p.LengthPrefixedDecoder{}
	var rpc netp2p.RPC
	if err := decoder.Decode(buf, &rpc); err != nil {
		t.Fatalf("failed to decode message: %v", err)
	}

	// Verify the decoded message
	if rpc.Stream {
		t.Error("expected Stream to be false for message")
	}
	if !bytes.Equal(rpc.Payload, payload) {
		t.Errorf("payload mismatch: expected %v, got %v", payload, rpc.Payload)
	}
}

func TestStreamHeaderDecoding(t *testing.T) {
	// Test stream header decoding
	buf := new(bytes.Buffer)
	frameWriter := netp2p.NewFrameWriter(buf)
	if err := frameWriter.WriteStreamHeader(); err != nil {
		t.Fatalf("failed to write stream header: %v", err)
	}

	// Decode the stream header
	decoder := netp2p.LengthPrefixedDecoder{}
	var rpc netp2p.RPC
	if err := decoder.Decode(buf, &rpc); err != nil {
		t.Fatalf("failed to decode stream header: %v", err)
	}

	// Verify the decoded stream header
	if !rpc.Stream {
		t.Error("expected Stream to be true for stream header")
	}
	if len(rpc.Payload) != 0 {
		t.Errorf("expected empty payload for stream header, got %v", rpc.Payload)
	}
}

func TestFrameWriter(t *testing.T) {
	// Test frame writer functionality
	buf := new(bytes.Buffer)
	frameWriter := netp2p.NewFrameWriter(buf)

	// Test writing a message
	messagePayload := []byte("hello world")
	if err := frameWriter.WriteMessage(messagePayload); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Test writing a stream header
	if err := frameWriter.WriteStreamHeader(); err != nil {
		t.Fatalf("failed to write stream header: %v", err)
	}

	// Verify the written data
	data := buf.Bytes()

	// Expected: [type:1][len:4][payload:11] + [type:2][len:4]
	expectedLen := 1 + 4 + len(messagePayload) + 1 + 4
	if len(data) != expectedLen {
		t.Errorf("expected %d bytes, got %d", expectedLen, len(data))
	}

	// Verify message frame
	if data[0] != netp2p.IncomingMessage {
		t.Errorf("expected message type %d, got %d", netp2p.IncomingMessage, data[0])
	}

	// Verify stream frame
	streamFrameStart := 1 + 4 + len(messagePayload)
	if data[streamFrameStart] != netp2p.IncomingStream {
		t.Errorf("expected stream type %d, got %d", netp2p.IncomingStream, data[streamFrameStart])
	}
}

func TestDecoderWithPartialReads(t *testing.T) {
	// Test decoder with partial reads (simulating network conditions)
	payload := []byte("large message for testing")

	// Create a buffer with a properly framed message
	buf := new(bytes.Buffer)
	frameWriter := netp2p.NewFrameWriter(buf)
	if err := frameWriter.WriteMessage(payload); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Create a reader that reads one byte at a time
	slowReader := &slowReader{buf: buf}

	// Decode the message
	decoder := netp2p.LengthPrefixedDecoder{}
	var rpc netp2p.RPC
	if err := decoder.Decode(slowReader, &rpc); err != nil {
		t.Fatalf("failed to decode message with partial reads: %v", err)
	}

	// Verify the decoded message
	if !bytes.Equal(rpc.Payload, payload) {
		t.Errorf("payload mismatch: expected %v, got %v", payload, rpc.Payload)
	}
}

// slowReader reads one byte at a time to simulate network conditions
type slowReader struct {
	buf *bytes.Buffer
}

func (sr *slowReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Read only one byte at a time
	return sr.buf.Read(p[:1])
}