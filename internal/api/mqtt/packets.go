package mqtt

import (
	"encoding/binary"
	"fmt"
)

// MessageType represents the MQTT message type
type MessageType byte

const (
	RESERVED    MessageType = 0
	CONNECT     MessageType = 1
	CONNACK     MessageType = 2
	PUBLISH     MessageType = 3
	PUBACK      MessageType = 4
	PUBREC      MessageType = 5
	PUBREL      MessageType = 6
	PUBCOMP     MessageType = 7
	SUBSCRIBE   MessageType = 8
	SUBACK      MessageType = 9
	UNSUBSCRIBE MessageType = 10
	UNSUBACK    MessageType = 11
	PINGREQ     MessageType = 12
	PINGRESP    MessageType = 13
	DISCONNECT  MessageType = 14
)

// QoS represents the Quality of Service level
type QoS byte

const (
	QoS0 QoS = 0 // At most once
	QoS1 QoS = 1 // At least once
	QoS2 QoS = 2 // Exactly once
)

// ConnackReturnCode represents the CONNACK return code
type ConnackReturnCode byte

const (
	CONNACK_ACCEPTED              ConnackReturnCode = 0
	CONNACK_UNACCEPTABLE_PROTOCOL ConnackReturnCode = 1
	CONNACK_IDENTIFIER_REJECTED   ConnackReturnCode = 2
	CONNACK_SERVER_UNAVAILABLE    ConnackReturnCode = 3
	CONNACK_BAD_USERNAME_PASSWORD ConnackReturnCode = 4
	CONNACK_NOT_AUTHORIZED        ConnackReturnCode = 5
)

// SubackReturnCode represents the SUBACK return code
type SubackReturnCode byte

const (
	SUBACK_SUCCESS_QOS0 SubackReturnCode = 0
	SUBACK_SUCCESS_QOS1 SubackReturnCode = 1
	SUBACK_SUCCESS_QOS2 SubackReturnCode = 2
	SUBACK_FAILURE      SubackReturnCode = 128
)

// Packet represents an MQTT packet
type Packet struct {
	Header *FixedHeader
	Data   []byte
}

// FixedHeader represents the MQTT fixed header
type FixedHeader struct {
	MessageType MessageType
	Flags       byte
}

// ConnectPacket represents a CONNECT packet
type ConnectPacket struct {
	ProtocolName  string
	ProtocolLevel byte
	ConnectFlags  byte
	KeepAlive     uint16
	ClientID      string
	WillTopic     string
	WillMessage   []byte
	Username      string
	Password      []byte
	WillFlag      bool
	WillQoS       byte
	WillRetain    bool
	PasswordFlag  bool
	UsernameFlag  bool
	CleanSession  bool
}

// ConnackPacket represents a CONNACK packet
type ConnackPacket struct {
	SessionPresent bool
	ReturnCode     ConnackReturnCode
}

// PublishPacket represents a PUBLISH packet
type PublishPacket struct {
	Topic    string
	PacketID uint16
	Payload  []byte
	QoS      QoS
	Retain   bool
	Dup      bool
}

// PubackPacket represents a PUBACK packet
type PubackPacket struct {
	PacketID uint16
}

// PubrecPacket represents a PUBREC packet
type PubrecPacket struct {
	PacketID uint16
}

// PubrelPacket represents a PUBREL packet
type PubrelPacket struct {
	PacketID uint16
}

// PubcompPacket represents a PUBCOMP packet
type PubcompPacket struct {
	PacketID uint16
}

// SubscribePacket represents a SUBSCRIBE packet
type SubscribePacket struct {
	PacketID      uint16
	Subscriptions []Subscription
}

// Subscription represents a topic subscription
type Subscription struct {
	Topic string
	QoS   QoS
}

// SubackPacket represents a SUBACK packet
type SubackPacket struct {
	PacketID    uint16
	ReturnCodes []byte
}

// UnsubscribePacket represents an UNSUBSCRIBE packet
type UnsubscribePacket struct {
	PacketID uint16
	Topics   []string
}

// UnsubackPacket represents an UNSUBACK packet
type UnsubackPacket struct {
	PacketID uint16
}

// Message represents an MQTT message
type Message struct {
	Topic   string
	Payload []byte
	QoS     QoS
	Retain  bool
}

// Encode encodes a packet to bytes
func (p *Packet) Encode() ([]byte, error) {
	// Encode variable header and payload
	var data []byte
	var err error

	switch p.Header.MessageType {
	case CONNACK:
		data, err = p.encodeConnack()
	case PUBACK:
		data, err = p.encodePuback()
	case PUBREC:
		data, err = p.encodePubrec()
	case PUBREL:
		data, err = p.encodePubrel()
	case PUBCOMP:
		data, err = p.encodePubcomp()
	case SUBACK:
		data, err = p.encodeSuback()
	case UNSUBACK:
		data, err = p.encodeUnsuback()
	case PINGRESP:
		data, err = p.encodePingresp()
	default:
		return nil, fmt.Errorf("unsupported message type for encoding: %d", p.Header.MessageType)
	}

	if err != nil {
		return nil, err
	}

	// Encode fixed header
	header := p.encodeFixedHeader(len(data))

	// Combine header and data
	result := make([]byte, len(header)+len(data))
	copy(result, header)
	copy(result[len(header):], data)

	return result, nil
}

// encodeFixedHeader encodes the fixed header
func (p *Packet) encodeFixedHeader(remainingLength int) []byte {
	// Message type and flags
	header := byte(p.Header.MessageType)<<4 | p.Header.Flags

	// Remaining length
	lengthBytes := encodeRemainingLength(remainingLength)

	// Combine
	result := make([]byte, 1+len(lengthBytes))
	result[0] = header
	copy(result[1:], lengthBytes)

	return result
}

// encodeRemainingLength encodes the remaining length
func encodeRemainingLength(length int) []byte {
	var result []byte

	for {
		encodedByte := byte(length % 128)
		length /= 128

		if length > 0 {
			encodedByte |= 128
		}

		result = append(result, encodedByte)

		if length == 0 {
			break
		}
	}

	return result
}

// encodeConnack encodes a CONNACK packet
func (p *Packet) encodeConnack() ([]byte, error) {
	// CONNACK is always 2 bytes
	result := make([]byte, 2)

	// Session present flag
	if p.Data[0] != 0 {
		result[0] = 1
	}

	// Return code
	result[1] = p.Data[1]

	return result, nil
}

// encodePuback encodes a PUBACK packet
func (p *Packet) encodePuback() ([]byte, error) {
	// PUBACK is always 2 bytes (packet ID)
	return p.Data, nil
}

// encodePubrec encodes a PUBREC packet
func (p *Packet) encodePubrec() ([]byte, error) {
	// PUBREC is always 2 bytes (packet ID)
	return p.Data, nil
}

// encodePubrel encodes a PUBREL packet
func (p *Packet) encodePubrel() ([]byte, error) {
	// PUBREL is always 2 bytes (packet ID)
	return p.Data, nil
}

// encodePubcomp encodes a PUBCOMP packet
func (p *Packet) encodePubcomp() ([]byte, error) {
	// PUBCOMP is always 2 bytes (packet ID)
	return p.Data, nil
}

// encodeSuback encodes a SUBACK packet
func (p *Packet) encodeSuback() ([]byte, error) {
	// SUBACK: packet ID (2 bytes) + return codes
	return p.Data, nil
}

// encodeUnsuback encodes an UNSUBACK packet
func (p *Packet) encodeUnsuback() ([]byte, error) {
	// UNSUBACK is always 2 bytes (packet ID)
	return p.Data, nil
}

// encodePingresp encodes a PINGRESP packet
func (p *Packet) encodePingresp() ([]byte, error) {
	// PINGRESP has no payload
	return []byte{}, nil
}

// Helper functions for reading MQTT data

// readString reads an MQTT string (2-byte length + string)
func readString(data []byte, offset int) (string, int, error) {
	if offset+2 > len(data) {
		return "", offset, fmt.Errorf("insufficient data for string length")
	}

	length := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	if offset+length > len(data) {
		return "", offset, fmt.Errorf("insufficient data for string content")
	}

	str := string(data[offset : offset+length])
	offset += length

	return str, offset, nil
}

// writeString writes an MQTT string (2-byte length + string)
func writeString(data []byte, offset int, str string) int {
	length := len(str)
	binary.BigEndian.PutUint16(data[offset:offset+2], uint16(length))
	offset += 2
	copy(data[offset:offset+length], str)
	offset += length
	return offset
}

// readUint16 reads a 16-bit unsigned integer
func readUint16(data []byte, offset int) (uint16, int, error) {
	if offset+2 > len(data) {
		return 0, offset, fmt.Errorf("insufficient data for uint16")
	}

	value := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	return value, offset, nil
}

// writeUint16 writes a 16-bit unsigned integer
func writeUint16(data []byte, offset int, value uint16) int {
	binary.BigEndian.PutUint16(data[offset:offset+2], value)
	return offset + 2
}

// readByte reads a single byte
func readByte(data []byte, offset int) (byte, int, error) {
	if offset >= len(data) {
		return 0, offset, fmt.Errorf("insufficient data for byte")
	}

	value := data[offset]
	offset++

	return value, offset, nil
}

// writeByte writes a single byte
func writeByte(data []byte, offset int, value byte) int {
	data[offset] = value
	return offset + 1
}
