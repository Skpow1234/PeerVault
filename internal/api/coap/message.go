package coap

import (
	"encoding/binary"
	"fmt"
)

// MessageType represents the CoAP message type
type MessageType byte

const (
	Confirmable     MessageType = 0
	NonConfirmable  MessageType = 1
	Acknowledgement MessageType = 2
	Reset           MessageType = 3
)

// MethodCode represents the CoAP method code
type MethodCode byte

const (
	GET    MethodCode = 1
	POST   MethodCode = 2
	PUT    MethodCode = 3
	DELETE MethodCode = 4
)

// ResponseCode represents the CoAP response code
type ResponseCode byte

const (
	// Success responses
	Created ResponseCode = 65 // 2.01
	Deleted ResponseCode = 66 // 2.02
	Valid   ResponseCode = 67 // 2.03
	Changed ResponseCode = 68 // 2.04
	Content ResponseCode = 69 // 2.05

	// Client error responses
	BadRequest               ResponseCode = 128 // 4.00
	Unauthorized             ResponseCode = 129 // 4.01
	BadOption                ResponseCode = 130 // 4.02
	Forbidden                ResponseCode = 131 // 4.03
	NotFound                 ResponseCode = 132 // 4.04
	MethodNotAllowed         ResponseCode = 133 // 4.05
	NotAcceptable            ResponseCode = 134 // 4.06
	PreconditionFailed       ResponseCode = 140 // 4.12
	RequestEntityTooLarge    ResponseCode = 141 // 4.13
	UnsupportedContentFormat ResponseCode = 143 // 4.15

	// Server error responses
	InternalServerError  ResponseCode = 160 // 5.00
	NotImplemented       ResponseCode = 161 // 5.01
	BadGateway           ResponseCode = 162 // 5.02
	ServiceUnavailable   ResponseCode = 163 // 5.03
	GatewayTimeout       ResponseCode = 164 // 5.04
	ProxyingNotSupported ResponseCode = 165 // 5.05
)

// OptionNumber represents CoAP option numbers
type OptionNumber uint16

const (
	IfMatch       OptionNumber = 1
	UriHost       OptionNumber = 3
	ETag          OptionNumber = 4
	IfNoneMatch   OptionNumber = 5
	UriPort       OptionNumber = 7
	LocationPath  OptionNumber = 8
	UriPath       OptionNumber = 11
	ContentFormat OptionNumber = 12
	MaxAge        OptionNumber = 14
	UriQuery      OptionNumber = 15
	Accept        OptionNumber = 17
	LocationQuery OptionNumber = 20
	ProxyUri      OptionNumber = 35
	ProxyScheme   OptionNumber = 39
	Size1         OptionNumber = 60
	Observe       OptionNumber = 6
	Block1        OptionNumber = 27
	Block2        OptionNumber = 23
)

// CoAPContentFormat represents CoAP content formats
type CoAPContentFormat uint16

const (
	ContentFormatTextPlain              CoAPContentFormat = 0
	ContentFormatApplicationLinkFormat  CoAPContentFormat = 40
	ContentFormatApplicationXML         CoAPContentFormat = 41
	ContentFormatApplicationOctetStream CoAPContentFormat = 42
	ContentFormatApplicationEXI         CoAPContentFormat = 47
	ContentFormatApplicationJSON        CoAPContentFormat = 50
	ContentFormatApplicationCBOR        CoAPContentFormat = 60
)

// Message represents a CoAP message
type Message struct {
	Type      MessageType
	Code      byte
	MessageID uint16
	Token     []byte
	Options   []Option
	Payload   []byte
}

// Option represents a CoAP option
type Option struct {
	Number OptionNumber
	Value  []byte
}

// AddOption adds an option to the message
func (m *Message) AddOption(number OptionNumber, value interface{}) {
	var optionValue []byte

	switch v := value.(type) {
	case uint32:
		optionValue = encodeUint32(v)
	case uint16:
		optionValue = encodeUint16(v)
	case uint8:
		optionValue = []byte{v}
	case string:
		optionValue = []byte(v)
	case []byte:
		optionValue = v
	default:
		optionValue = []byte(fmt.Sprintf("%v", v))
	}

	option := Option{
		Number: number,
		Value:  optionValue,
	}

	m.Options = append(m.Options, option)
}

// GetOption gets an option value by number
func (m *Message) GetOption(number OptionNumber) []byte {
	for _, option := range m.Options {
		if option.Number == number {
			return option.Value
		}
	}
	return nil
}

// HasOption checks if the message has an option
func (m *Message) HasOption(number OptionNumber) bool {
	return m.GetOption(number) != nil
}

// GetPath gets the URI path from options
func (m *Message) GetPath() string {
	var path string
	for _, option := range m.Options {
		if option.Number == UriPath {
			if path != "" {
				path += "/"
			}
			path += string(option.Value)
		}
	}
	return "/" + path
}

// Encode encodes a CoAP message to bytes
func (m *Message) Encode() ([]byte, error) {
	// Calculate header size
	headerSize := 4 // Fixed header size

	// Calculate token length
	tokenLength := len(m.Token)
	if tokenLength > 8 {
		return nil, fmt.Errorf("token too long: %d bytes", tokenLength)
	}

	// Calculate options size
	optionsSize := 0
	for _, option := range m.Options {
		optionsSize += m.encodeOptionSize(option)
	}

	// Calculate total size
	totalSize := headerSize + tokenLength + optionsSize + len(m.Payload)

	// Create buffer
	buffer := make([]byte, totalSize)
	offset := 0

	// Encode header
	offset = m.encodeHeader(buffer, offset, tokenLength)

	// Encode token
	if tokenLength > 0 {
		copy(buffer[offset:offset+tokenLength], m.Token)
		offset += tokenLength
	}

	// Encode options
	for _, option := range m.Options {
		offset = m.encodeOption(buffer, offset, option)
	}

	// Encode payload
	if len(m.Payload) > 0 {
		// Add payload marker (0xFF)
		buffer[offset] = 0xFF
		offset++
		copy(buffer[offset:], m.Payload)
	}

	return buffer, nil
}

// encodeHeader encodes the CoAP header
func (m *Message) encodeHeader(buffer []byte, offset int, tokenLength int) int {
	// Version (2 bits), Type (2 bits), Token Length (4 bits)
	header := byte(1)<<6 | byte(m.Type)<<4 | byte(tokenLength)

	// Code (8 bits)
	code := byte(m.Code)

	// Message ID (16 bits)
	messageID := m.MessageID

	// Write header
	buffer[offset] = header
	offset++
	buffer[offset] = code
	offset++
	binary.BigEndian.PutUint16(buffer[offset:], messageID)
	offset += 2

	return offset
}

// encodeOptionSize calculates the size needed to encode an option
func (m *Message) encodeOptionSize(option Option) int {
	// Option delta and length are encoded in the first byte
	// Additional bytes for delta and length if needed
	size := 1

	// Delta encoding
	delta := uint16(option.Number)
	if delta >= 13 {
		if delta < 269 {
			size++ // 1 additional byte
		} else {
			size += 2 // 2 additional bytes
		}
	}

	// Length encoding
	length := uint16(len(option.Value))
	if length >= 13 {
		if length < 269 {
			size++ // 1 additional byte
		} else {
			size += 2 // 2 additional bytes
		}
	}

	// Value
	size += len(option.Value)

	return size
}

// encodeOption encodes an option
func (m *Message) encodeOption(buffer []byte, offset int, option Option) int {
	delta := uint16(option.Number)
	length := uint16(len(option.Value))

	// Encode delta and length in first byte
	firstByte := byte(0)

	// Delta encoding
	if delta < 13 {
		firstByte |= byte(delta) << 4
	} else if delta < 269 {
		firstByte |= 13 << 4
		delta -= 13
	} else {
		firstByte |= 14 << 4
		delta -= 269
	}

	// Length encoding
	if length < 13 {
		firstByte |= byte(length)
	} else if length < 269 {
		firstByte |= 13
		length -= 13
	} else {
		firstByte |= 14
		length -= 269
	}

	buffer[offset] = firstByte
	offset++

	// Encode additional delta bytes
	if delta >= 13 {
		if delta < 269 {
			buffer[offset] = byte(delta)
			offset++
		} else {
			binary.BigEndian.PutUint16(buffer[offset:], delta)
			offset += 2
		}
	}

	// Encode additional length bytes
	if length >= 13 {
		if length < 269 {
			buffer[offset] = byte(length)
			offset++
		} else {
			binary.BigEndian.PutUint16(buffer[offset:], length)
			offset += 2
		}
	}

	// Encode value
	if len(option.Value) > 0 {
		copy(buffer[offset:], option.Value)
		offset += len(option.Value)
	}

	return offset
}

// ParseMessage parses a CoAP message from bytes
func ParseMessage(data []byte) (*Message, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("message too short: %d bytes", len(data))
	}

	message := &Message{}

	// Parse header
	header := data[0]
	version := (header >> 6) & 0x03
	if version != 1 {
		return nil, fmt.Errorf("unsupported CoAP version: %d", version)
	}

	message.Type = MessageType((header >> 4) & 0x03)
	tokenLength := header & 0x0F

	message.Code = data[1]
	message.MessageID = binary.BigEndian.Uint16(data[2:4])

	offset := 4

	// Parse token
	if tokenLength > 0 {
		if offset+int(tokenLength) > len(data) {
			return nil, fmt.Errorf("token extends beyond message")
		}
		message.Token = make([]byte, tokenLength)
		copy(message.Token, data[offset:offset+int(tokenLength)])
		offset += int(tokenLength)
	}

	// Parse options
	prevOptionNumber := uint16(0)
	for offset < len(data) {
		if data[offset] == 0xFF {
			// Payload marker
			offset++
			break
		}

		option, newOffset, err := parseOption(data, offset, prevOptionNumber)
		if err != nil {
			return nil, err
		}

		message.Options = append(message.Options, option)
		prevOptionNumber = uint16(option.Number)
		offset = newOffset
	}

	// Parse payload
	if offset < len(data) {
		message.Payload = data[offset:]
	}

	return message, nil
}

// parseOption parses a CoAP option
func parseOption(data []byte, offset int, prevOptionNumber uint16) (Option, int, error) {
	if offset >= len(data) {
		return Option{}, offset, fmt.Errorf("unexpected end of message")
	}

	firstByte := data[offset]
	offset++

	// Parse delta
	delta := uint16((firstByte >> 4) & 0x0F)
	switch delta {
	case 13:
		if offset >= len(data) {
			return Option{}, offset, fmt.Errorf("unexpected end of message")
		}
		delta = uint16(data[offset]) + 13
		offset++
	case 14:
		if offset+1 >= len(data) {
			return Option{}, offset, fmt.Errorf("unexpected end of message")
		}
		delta = binary.BigEndian.Uint16(data[offset:]) + 269
		offset += 2
	}

	// Parse length
	length := uint16(firstByte & 0x0F)
	switch length {
	case 13:
		if offset >= len(data) {
			return Option{}, offset, fmt.Errorf("unexpected end of message")
		}
		length = uint16(data[offset]) + 13
		offset++
	case 14:
		if offset+1 >= len(data) {
			return Option{}, offset, fmt.Errorf("unexpected end of message")
		}
		length = binary.BigEndian.Uint16(data[offset:]) + 269
		offset += 2
	}

	// Parse value
	if offset+int(length) > len(data) {
		return Option{}, offset, fmt.Errorf("option value extends beyond message")
	}

	value := make([]byte, length)
	copy(value, data[offset:offset+int(length)])
	offset += int(length)

	optionNumber := prevOptionNumber + delta

	return Option{
		Number: OptionNumber(optionNumber),
		Value:  value,
	}, offset, nil
}

// Helper functions for encoding values

// encodeUint32 encodes a uint32 to bytes
func encodeUint32(value uint32) []byte {
	if value == 0 {
		return []byte{}
	}

	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, value)

	// Remove leading zeros
	for i, b := range bytes {
		if b != 0 {
			return bytes[i:]
		}
	}

	return []byte{0}
}

// encodeUint16 encodes a uint16 to bytes
func encodeUint16(value uint16) []byte {
	if value == 0 {
		return []byte{}
	}

	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, value)

	// Remove leading zeros
	if bytes[0] != 0 {
		return bytes
	}

	return bytes[1:]
}
