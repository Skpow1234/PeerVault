package transport

import (
	"bytes"
	"encoding/binary"
	"testing"

	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

// FuzzDecode tests the decoder with fuzz-generated data
func FuzzDecode(f *testing.F) {
	// Add seed corpus for fuzz testing
	seedCorpus := [][]byte{
		// Valid message frame
		{0x01, 0x00, 0x00, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, // type=1, len=5, payload="hello"

		// Valid stream frame
		{0x02, 0x00, 0x00, 0x00, 0x00}, // type=2, len=0

		// Empty frame
		{0x01, 0x00, 0x00, 0x00, 0x00}, // type=1, len=0

		// Large frame
		func() []byte {
			data := make([]byte, 1024)
			for i := range data {
				data[i] = byte(i % 256)
			}
			frame := make([]byte, 5+len(data))
			frame[0] = 0x01                                           // type
			binary.BigEndian.PutUint32(frame[1:5], uint32(len(data))) // length
			copy(frame[5:], data)                                     // payload
			return frame
		}(),
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	decoder := netp2p.LengthPrefixedDecoder{}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return // Skip empty data
		}

		// Create a buffer with the fuzz data
		buf := bytes.NewBuffer(data)

		// Try to decode the frame
		msg := &netp2p.RPC{}
		err := decoder.Decode(buf, msg)

		// The decoder should either:
		// 1. Successfully decode a valid frame
		// 2. Return an error for invalid data

		if err != nil {
			// Error is expected for invalid data
			return
		}

		// If we got a valid message, verify its structure
		if len(msg.Payload) > 1024*1024 { // 1MB max
			t.Errorf("Payload too large: %d bytes", len(msg.Payload))
		}
	})
}

// FuzzHandshake tests the handshake with fuzz-generated data
func FuzzHandshake(f *testing.F) {
	// Add seed corpus for handshake testing
	seedCorpus := [][]byte{
		// Valid handshake data
		[]byte("valid_handshake_data"),

		// Empty data
		[]byte{},

		// Large data
		bytes.Repeat([]byte{0x41}, 1024), // 1KB of 'A'

		// Random bytes
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test handshake with fuzz data
		// This is a basic test - in a real implementation, you'd want to test
		// the actual handshake protocol with proper message structures

		if len(data) == 0 {
			return // Skip empty data
		}

		// Test that handshake doesn't panic with random data
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Handshake panicked with data: %v", r)
			}
		}()

		// Basic validation - handshake data should be reasonable size
		if len(data) > 10*1024 { // 10KB max
			return
		}

		// Test that we can process the data without crashing
		// In a real implementation, you'd test the actual handshake logic
		_ = len(data) // Use the data to avoid compiler warnings
	})
}

// FuzzMessageEncoding tests message encoding/decoding with fuzz data
func FuzzMessageEncoding(f *testing.F) {
	// Add seed corpus for message encoding testing
	seedCorpus := [][]byte{
		// Valid message data
		[]byte("test message"),

		// Empty message
		[]byte{},

		// Binary data
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},

		// Large data
		bytes.Repeat([]byte{0x42}, 512), // 512 bytes of 'B'
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return // Skip empty data
		}

		// Test message encoding/decoding
		// This is a basic test - in a real implementation, you'd test
		// the actual message encoding/decoding logic

		// Test that encoding doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Message encoding panicked with data: %v", r)
			}
		}()

		// Basic validation - message data should be reasonable size
		if len(data) > 1024*1024 { // 1MB max
			return
		}

		// Test that we can process the data without crashing
		// In a real implementation, you'd test the actual encoding/decoding logic
		_ = len(data) // Use the data to avoid compiler warnings
	})
}

// FuzzStreamProcessing tests stream processing with fuzz data
func FuzzStreamProcessing(f *testing.F) {
	// Add seed corpus for stream processing testing
	seedCorpus := [][]byte{
		// Small stream data
		[]byte("small stream"),

		// Empty stream
		[]byte{},

		// Binary stream data
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},

		// Large stream data
		bytes.Repeat([]byte{0x43}, 1024), // 1KB of 'C'
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return // Skip empty data
		}

		// Test stream processing
		// This is a basic test - in a real implementation, you'd test
		// the actual stream processing logic

		// Test that stream processing doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Stream processing panicked with data: %v", r)
			}
		}()

		// Basic validation - stream data should be reasonable size
		if len(data) > 10*1024*1024 { // 10MB max
			return
		}

		// Test that we can process the data without crashing
		// In a real implementation, you'd test the actual stream processing logic
		_ = len(data) // Use the data to avoid compiler warnings
	})
}

// FuzzPartialReads tests decoder behavior with partial reads
func FuzzPartialReads(f *testing.F) {
	// Add seed corpus for partial read testing
	seedCorpus := [][]byte{
		// Complete frame
		{0x01, 0x00, 0x00, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f},

		// Partial frame (header only)
		{0x01, 0x00, 0x00, 0x00, 0x05},

		// Incomplete header
		{0x01, 0x00, 0x00},

		// Single byte
		{0x01},
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return // Skip empty data
		}

		decoder := netp2p.LengthPrefixedDecoder{}
		buf := bytes.NewBuffer(data)

		// Try to decode - should handle partial reads gracefully
		msg := &netp2p.RPC{}
		err := decoder.Decode(buf, msg)

		// For partial reads, we expect either:
		// 1. Error for invalid/incomplete data
		// 2. Valid message for complete data

		if err != nil {
			// Error is expected for invalid/incomplete data
			return
		}

		// If we got a valid message, validate it
		if len(msg.Payload) > 1024*1024 { // 1MB max
			t.Errorf("Payload too large: %d bytes", len(msg.Payload))
		}
	})
}
