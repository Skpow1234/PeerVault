package p2p

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// Frame header structure: [type:u8][len:u32]
const (
	FrameHeaderSize = 5           // 1 byte type + 4 bytes length
	MaxFrameSize    = 1024 * 1024 // 1MB max frame size
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

// LengthPrefixedDecoder implements proper message framing
type LengthPrefixedDecoder struct{}

// Decode reads a length-prefixed frame from the reader
func (dec LengthPrefixedDecoder) Decode(r io.Reader, msg *RPC) error {
	if r == nil {
		return fmt.Errorf("reader is nil")
	}
	
	// Read frame header: [type:u8][len:u32]
	header := make([]byte, FrameHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return fmt.Errorf("failed to read frame header: %w", err)
	}

	// Parse message type
	msgType := header[0]

	// Parse payload length
	payloadLen := binary.BigEndian.Uint32(header[1:])

	// Validate payload length
	if payloadLen > MaxFrameSize {
		return fmt.Errorf("frame too large: %d bytes (max: %d)", payloadLen, MaxFrameSize)
	}

	// Handle stream type
	if msgType == IncomingStream {
		msg.Stream = true
		return nil
	}

	// Handle message type
	if msgType == IncomingMessage {
		msg.Stream = false

		// Read payload
		if payloadLen > 0 {
			msg.Payload = make([]byte, payloadLen)
			if _, err := io.ReadFull(r, msg.Payload); err != nil {
				return fmt.Errorf("failed to read payload: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("unknown message type: %d", msgType)
}

// DefaultDecoder is kept for backward compatibility but deprecated
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	if r == nil {
		return fmt.Errorf("reader is nil")
	}

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

// FrameWriter provides methods to write properly framed messages
type FrameWriter struct {
	writer io.Writer
}

// NewFrameWriter creates a new frame writer
func NewFrameWriter(writer io.Writer) *FrameWriter {
	return &FrameWriter{writer: writer}
}

// WriteMessage writes a message with proper framing
func (fw *FrameWriter) WriteMessage(payload []byte) error {
	return fw.writeFrame(IncomingMessage, payload)
}

// WriteStreamHeader writes a stream header
func (fw *FrameWriter) WriteStreamHeader() error {
	return fw.writeFrame(IncomingStream, nil)
}

// writeFrame writes a frame with the given type and payload
func (fw *FrameWriter) writeFrame(msgType byte, payload []byte) error {
	// Create frame header: [type:u8][len:u32]
	header := make([]byte, FrameHeaderSize)
	header[0] = msgType
	binary.BigEndian.PutUint32(header[1:], uint32(len(payload)))

	// Write header
	if _, err := fw.writer.Write(header); err != nil {
		return fmt.Errorf("failed to write frame header: %w", err)
	}

	// Write payload if present
	if len(payload) > 0 {
		if _, err := fw.writer.Write(payload); err != nil {
			return fmt.Errorf("failed to write payload: %w", err)
		}
	}

	return nil
}