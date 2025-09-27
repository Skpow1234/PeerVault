package translation

import (
	"encoding/json"
	"fmt"
	"net"
)

// WebSocketToCoAPTranslator translates WebSocket messages to CoAP format
type WebSocketToCoAPTranslator struct {
	coapAddr string
	conn     *net.UDPConn
}

// NewWebSocketToCoAPTranslator creates a new WebSocket to CoAP translator
func NewWebSocketToCoAPTranslator(coapAddr string) *WebSocketToCoAPTranslator {
	return &WebSocketToCoAPTranslator{
		coapAddr: coapAddr,
	}
}

// Translate translates a WebSocket message to CoAP format
func (t *WebSocketToCoAPTranslator) Translate(message *Message) (*Message, error) {
	// Convert WebSocket message to CoAP format
	coapMessage := &Message{
		ID:        message.ID,
		Protocol:  "coap",
		Type:      t.mapMessageType(message.Type),
		Topic:     t.mapTopic(message.Topic),
		Payload:   message.Payload,
		Headers:   t.mapHeaders(message.Headers),
		Metadata:  t.mapMetadata(message.Metadata),
		Timestamp: message.Timestamp,
	}

	// Add CoAP-specific metadata
	if coapMessage.Metadata == nil {
		coapMessage.Metadata = make(map[string]interface{})
	}
	coapMessage.Metadata["coap_method"] = t.mapMethod(message.Type)
	coapMessage.Metadata["coap_code"] = t.mapCode(message.Type)
	coapMessage.Metadata["coap_type"] = t.mapCoAPType(message.Type)
	coapMessage.Metadata["coap_token"] = t.generateToken(message.ID)

	return coapMessage, nil
}

// CanTranslate checks if this translator can handle the given protocols
func (t *WebSocketToCoAPTranslator) CanTranslate(from, to string) bool {
	return from == "websocket" && to == "coap"
}

// GetSupportedProtocols returns the protocols this translator supports
func (t *WebSocketToCoAPTranslator) GetSupportedProtocols() []string {
	return []string{"websocket", "coap"}
}

// mapMessageType maps WebSocket message types to CoAP equivalents
func (t *WebSocketToCoAPTranslator) mapMessageType(wsType string) string {
	switch wsType {
	case "text", "message":
		return "request"
	case "binary":
		return "request"
	case "ping":
		return "ping"
	case "pong":
		return "pong"
	case "close":
		return "reset"
	case "error":
		return "reset"
	case "subscribe":
		return "observe"
	case "unsubscribe":
		return "unobserve"
	default:
		return "request"
	}
}

// mapTopic maps WebSocket topics to CoAP URIs
func (t *WebSocketToCoAPTranslator) mapTopic(wsTopic string) string {
	if wsTopic == "" {
		return "/ws/messages"
	}

	// Convert WebSocket topic format to CoAP URI format
	// WebSocket topics might use / separators, CoAP uses / as well
	coapURI := fmt.Sprintf("/ws%s", wsTopic)

	// Ensure URI starts with /
	if coapURI[0] != '/' {
		coapURI = "/" + coapURI
	}

	return coapURI
}

// mapMethod maps WebSocket message types to CoAP methods
func (t *WebSocketToCoAPTranslator) mapMethod(wsType string) string {
	switch wsType {
	case "text", "message":
		return "POST" // POST for sending messages
	case "binary":
		return "POST" // POST for binary data
	case "subscribe":
		return "GET" // GET for observing/subscribing
	case "unsubscribe":
		return "GET" // GET for unobserving/unsubscribing
	case "ping":
		return "GET" // GET for ping
	case "pong":
		return "GET" // GET for pong
	default:
		return "POST"
	}
}

// mapCode maps WebSocket message types to CoAP response codes
func (t *WebSocketToCoAPTranslator) mapCode(wsType string) string {
	switch wsType {
	case "text", "message":
		return "2.05" // Content
	case "binary":
		return "2.05" // Content
	case "subscribe":
		return "2.05" // Content
	case "unsubscribe":
		return "2.05" // Content
	case "ping":
		return "2.05" // Content
	case "pong":
		return "2.05" // Content
	case "close":
		return "4.00" // Bad Request
	case "error":
		return "5.00" // Internal Server Error
	default:
		return "2.05" // Content
	}
}

// mapCoAPType maps WebSocket message types to CoAP message types
func (t *WebSocketToCoAPTranslator) mapCoAPType(wsType string) string {
	switch wsType {
	case "text", "message", "binary":
		return "CON" // Confirmable for important messages
	case "subscribe", "unsubscribe":
		return "CON" // Confirmable for subscriptions
	case "ping":
		return "CON" // Confirmable for ping
	case "pong":
		return "ACK" // Acknowledgment for pong
	case "close", "error":
		return "NON" // Non-confirmable for close/error
	default:
		return "CON" // Default to confirmable
	}
}

// mapHeaders maps WebSocket headers to CoAP options
func (t *WebSocketToCoAPTranslator) mapHeaders(wsHeaders map[string]string) map[string]string {
	if wsHeaders == nil {
		return nil
	}

	coapHeaders := make(map[string]string)
	for key, value := range wsHeaders {
		// Map WebSocket-specific headers to CoAP options
		switch key {
		case "Sec-WebSocket-Protocol":
			// Skip WebSocket-specific headers
			continue
		case "Sec-WebSocket-Extensions":
			// Skip WebSocket-specific headers
			continue
		case "Upgrade":
			// Skip WebSocket-specific headers
			continue
		case "Connection":
			// Skip WebSocket-specific headers
			continue
		case "Content-Type":
			coapHeaders["content_format"] = t.mapContentType(value)
		case "Content-Encoding":
			coapHeaders["content_encoding"] = value
		case "User-Agent":
			coapHeaders["user_agent"] = value
		case "Authorization":
			coapHeaders["authorization"] = value
		default:
			// Map other headers as custom options
			coapHeaders["custom_"+key] = value
		}
	}

	return coapHeaders
}

// mapContentType maps HTTP content types to CoAP content format numbers
func (t *WebSocketToCoAPTranslator) mapContentType(contentType string) string {
	switch contentType {
	case "application/json":
		return "50" // JSON
	case "text/plain":
		return "0" // Text/plain
	case "application/octet-stream":
		return "42" // Octet-stream
	case "application/xml":
		return "41" // XML
	case "text/html":
		return "0" // Text/plain (closest match)
	default:
		return "0" // Default to text/plain
	}
}

// mapMetadata maps WebSocket metadata to CoAP metadata
func (t *WebSocketToCoAPTranslator) mapMetadata(wsMetadata map[string]interface{}) map[string]interface{} {
	if wsMetadata == nil {
		return nil
	}

	coapMetadata := make(map[string]interface{})
	for key, value := range wsMetadata {
		// Map WebSocket-specific metadata to CoAP equivalents
		switch key {
		case "websocket_opcode":
			// Convert WebSocket opcode to CoAP message type
			if opcode, ok := value.(int); ok {
				coapMetadata["coap_message_type"] = t.mapOpcodeToCoAPType(opcode)
			}
		case "websocket_fin":
			// CoAP doesn't have frame fragmentation
			continue
		case "websocket_mask":
			// CoAP doesn't use masking
			continue
		case "websocket_length":
			// Map to CoAP payload length
			coapMetadata["coap_payload_length"] = value
		case "qos":
			// Map QoS to CoAP reliability
			if qos, ok := value.(int); ok {
				coapMetadata["coap_reliability"] = t.mapQoSToReliability(qos)
			}
		default:
			coapMetadata[key] = value
		}
	}

	return coapMetadata
}

// mapOpcodeToCoAPType maps WebSocket opcodes to CoAP message types
func (t *WebSocketToCoAPTranslator) mapOpcodeToCoAPType(opcode int) string {
	switch opcode {
	case 0x0: // Continuation frame
		return "NON" // Non-confirmable
	case 0x1: // Text frame
		return "CON" // Confirmable
	case 0x2: // Binary frame
		return "CON" // Confirmable
	case 0x8: // Connection close
		return "RST" // Reset
	case 0x9: // Ping
		return "CON" // Confirmable
	case 0xA: // Pong
		return "ACK" // Acknowledgment
	default:
		return "CON" // Default to confirmable
	}
}

// mapQoSToReliability maps MQTT QoS levels to CoAP reliability
func (t *WebSocketToCoAPTranslator) mapQoSToReliability(qos int) string {
	switch qos {
	case 0:
		return "NON" // Non-confirmable
	case 1:
		return "CON" // Confirmable
	case 2:
		return "CON" // Confirmable (CoAP doesn't have exactly-once)
	default:
		return "CON" // Default to confirmable
	}
}

// generateToken generates a CoAP token from message ID
func (t *WebSocketToCoAPTranslator) generateToken(messageID string) string {
	// Generate a simple token from message ID
	// In a real implementation, you might want a more sophisticated token generation
	token := fmt.Sprintf("ws_%s", messageID)
	if len(token) > 8 {
		token = token[:8] // CoAP tokens are typically 1-8 bytes
	}
	return token
}

// SendToCoAP sends the translated message to the CoAP server
func (t *WebSocketToCoAPTranslator) SendToCoAP(message *Message) error {
	// Connect to CoAP server if not connected
	if t.conn == nil {
		addr, err := net.ResolveUDPAddr("udp", t.coapAddr)
		if err != nil {
			return fmt.Errorf("failed to resolve CoAP address: %w", err)
		}

		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			return fmt.Errorf("failed to connect to CoAP server: %w", err)
		}
		t.conn = conn
	}

	// Create CoAP message
	coapMessage := t.createCoAPMessage(message)

	// Send message to CoAP server
	_, err := t.conn.Write(coapMessage)
	if err != nil {
		return fmt.Errorf("failed to send CoAP message: %w", err)
	}

	return nil
}

// createCoAPMessage creates a CoAP message
func (t *WebSocketToCoAPTranslator) createCoAPMessage(message *Message) []byte {
	// This is a simplified CoAP message creation
	// In a real implementation, you would use a proper CoAP library

	// Get method and code from metadata
	_ = "POST" // method not used in simplified implementation
	code := "2.05"
	msgType := "CON"
	if message.Metadata != nil {
		if methodVal, exists := message.Metadata["coap_method"]; exists {
			if _, ok := methodVal.(string); ok {
				// method = methodStr // not used in simplified implementation
			}
		}
		if codeVal, exists := message.Metadata["coap_code"]; exists {
			if codeStr, ok := codeVal.(string); ok {
				code = codeStr
			}
		}
		if typeVal, exists := message.Metadata["coap_type"]; exists {
			if typeStr, ok := typeVal.(string); ok {
				msgType = typeStr
			}
		}
	}

	// Convert payload to bytes
	var payload []byte
	if payloadStr, ok := message.Payload.(string); ok {
		payload = []byte(payloadStr)
	} else {
		payloadJSON, err := json.Marshal(message.Payload)
		if err != nil {
			payload = []byte(fmt.Sprintf("Error marshaling payload: %v", err))
		} else {
			payload = payloadJSON
		}
	}

	// Create CoAP message (simplified)
	// Version (2 bits) + Type (2 bits) + Token Length (4 bits)
	version := 0x01 // CoAP version 1
	var typeBits byte
	switch msgType {
	case "CON":
		typeBits = 0x00
	case "NON":
		typeBits = 0x01
	case "ACK":
		typeBits = 0x02
	case "RST":
		typeBits = 0x03
	}

	// Token length (simplified - using 4 bytes)
	tokenLength := 0x04

	// First byte
	firstByte := byte(version<<6) | byte(typeBits<<4) | byte(tokenLength)

	// Code (8 bits)
	var codeByte byte
	switch code {
	case "2.05":
		codeByte = 0x45 // 2.05 Content
	case "4.00":
		codeByte = 0x80 // 4.00 Bad Request
	case "5.00":
		codeByte = 0xA0 // 5.00 Internal Server Error
	default:
		codeByte = 0x45 // Default to 2.05 Content
	}

	// Message ID (16 bits) - simplified
	messageID := uint16(len(message.ID) % 65536)

	// Token (4 bytes) - simplified
	token := []byte("ws01")

	// Options (simplified - just URI-Path)
	uriPath := message.Topic
	if uriPath == "" {
		uriPath = "/ws/messages"
	}

	// URI-Path option
	uriPathBytes := []byte(uriPath)
	uriPathOption := []byte{0x00, byte(len(uriPathBytes))} // Option 0 (URI-Path)
	uriPathOption = append(uriPathOption, uriPathBytes...)

	// Payload marker (0xFF)
	payloadMarker := []byte{0xFF}

	// Build complete message
	var messageBytes []byte
	messageBytes = append(messageBytes, firstByte)
	messageBytes = append(messageBytes, codeByte)
	messageBytes = append(messageBytes, byte(messageID>>8), byte(messageID&0xFF))
	messageBytes = append(messageBytes, token...)
	messageBytes = append(messageBytes, uriPathOption...)
	messageBytes = append(messageBytes, payloadMarker...)
	messageBytes = append(messageBytes, payload...)

	return messageBytes
}

// Close closes the CoAP connection
func (t *WebSocketToCoAPTranslator) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
