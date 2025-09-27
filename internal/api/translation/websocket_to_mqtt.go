package translation

import (
	"encoding/json"
	"fmt"
	"net"
)

// WebSocketToMQTTTranslator translates WebSocket messages to MQTT format
type WebSocketToMQTTTranslator struct {
	mqttAddr string
	conn     net.Conn
}

// NewWebSocketToMQTTTranslator creates a new WebSocket to MQTT translator
func NewWebSocketToMQTTTranslator(mqttAddr string) *WebSocketToMQTTTranslator {
	return &WebSocketToMQTTTranslator{
		mqttAddr: mqttAddr,
	}
}

// Translate translates a WebSocket message to MQTT format
func (t *WebSocketToMQTTTranslator) Translate(message *Message) (*Message, error) {
	// Convert WebSocket message to MQTT format
	mqttMessage := &Message{
		ID:        message.ID,
		Protocol:  "mqtt",
		Type:      t.mapMessageType(message.Type),
		Topic:     t.mapTopic(message.Topic),
		Payload:   message.Payload,
		Headers:   t.mapHeaders(message.Headers),
		Metadata:  t.mapMetadata(message.Metadata),
		Timestamp: message.Timestamp,
	}

	// Add MQTT-specific metadata
	if mqttMessage.Metadata == nil {
		mqttMessage.Metadata = make(map[string]interface{})
	}
	mqttMessage.Metadata["mqtt_qos"] = t.mapQoS(message.Metadata)
	mqttMessage.Metadata["mqtt_retain"] = t.mapRetain(message.Metadata)
	mqttMessage.Metadata["mqtt_dup"] = false

	return mqttMessage, nil
}

// CanTranslate checks if this translator can handle the given protocols
func (t *WebSocketToMQTTTranslator) CanTranslate(from, to string) bool {
	return from == "websocket" && to == "mqtt"
}

// GetSupportedProtocols returns the protocols this translator supports
func (t *WebSocketToMQTTTranslator) GetSupportedProtocols() []string {
	return []string{"websocket", "mqtt"}
}

// mapMessageType maps WebSocket message types to MQTT equivalents
func (t *WebSocketToMQTTTranslator) mapMessageType(wsType string) string {
	switch wsType {
	case "text", "message":
		return "publish"
	case "binary":
		return "publish"
	case "ping":
		return "pingreq"
	case "pong":
		return "pingresp"
	case "close":
		return "disconnect"
	case "error":
		return "disconnect"
	case "subscribe":
		return "subscribe"
	case "unsubscribe":
		return "unsubscribe"
	default:
		return "publish"
	}
}

// mapTopic maps WebSocket topics to MQTT topics
func (t *WebSocketToMQTTTranslator) mapTopic(wsTopic string) string {
	if wsTopic == "" {
		return "websocket/messages"
	}

	// Convert WebSocket topic format to MQTT format
	// WebSocket topics might use / separators, MQTT uses / as well
	// But we need to ensure it's MQTT-compliant
	mqttTopic := fmt.Sprintf("ws/%s", wsTopic)

	// Ensure topic doesn't start with $ (reserved for MQTT system topics)
	if mqttTopic[0] == '$' {
		mqttTopic = "ws" + mqttTopic
	}

	return mqttTopic
}

// mapQoS maps WebSocket metadata to MQTT QoS levels
func (t *WebSocketToMQTTTranslator) mapQoS(wsMetadata map[string]interface{}) int {
	if wsMetadata == nil {
		return 0 // Default to QoS 0
	}

	// Check for explicit QoS setting
	if qos, exists := wsMetadata["qos"]; exists {
		if qosInt, ok := qos.(int); ok && qosInt >= 0 && qosInt <= 2 {
			return qosInt
		}
	}

	// Check for priority or reliability hints
	if priority, exists := wsMetadata["priority"]; exists {
		if priorityStr, ok := priority.(string); ok {
			switch priorityStr {
			case "high", "critical":
				return 2 // QoS 2 for high priority
			case "medium", "normal":
				return 1 // QoS 1 for medium priority
			case "low":
				return 0 // QoS 0 for low priority
			}
		}
	}

	// Check for reliability requirements
	if reliable, exists := wsMetadata["reliable"]; exists {
		if reliableBool, ok := reliable.(bool); ok && reliableBool {
			return 1 // QoS 1 for reliable delivery
		}
	}

	return 0 // Default to QoS 0
}

// mapRetain maps WebSocket metadata to MQTT retain flag
func (t *WebSocketToMQTTTranslator) mapRetain(wsMetadata map[string]interface{}) bool {
	if wsMetadata == nil {
		return false
	}

	// Check for explicit retain setting
	if retain, exists := wsMetadata["retain"]; exists {
		if retainBool, ok := retain.(bool); ok {
			return retainBool
		}
	}

	// Check for persistence hints
	if persistent, exists := wsMetadata["persistent"]; exists {
		if persistentBool, ok := persistent.(bool); ok {
			return persistentBool
		}
	}

	// Check for "last will" type messages
	if messageType, exists := wsMetadata["type"]; exists {
		if typeStr, ok := messageType.(string); ok {
			switch typeStr {
			case "status", "state", "configuration":
				return true // Retain status/state messages
			}
		}
	}

	return false
}

// mapHeaders maps WebSocket headers to MQTT properties
func (t *WebSocketToMQTTTranslator) mapHeaders(wsHeaders map[string]string) map[string]string {
	if wsHeaders == nil {
		return nil
	}

	mqttHeaders := make(map[string]string)
	for key, value := range wsHeaders {
		// Map WebSocket-specific headers to MQTT properties
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
			mqttHeaders["content_type"] = value
		case "Content-Encoding":
			mqttHeaders["content_encoding"] = value
		case "User-Agent":
			mqttHeaders["user_property"] = value
		default:
			// Map other headers as user properties
			mqttHeaders["user_property_"+key] = value
		}
	}

	return mqttHeaders
}

// mapMetadata maps WebSocket metadata to MQTT metadata
func (t *WebSocketToMQTTTranslator) mapMetadata(wsMetadata map[string]interface{}) map[string]interface{} {
	if wsMetadata == nil {
		return nil
	}

	mqttMetadata := make(map[string]interface{})
	for key, value := range wsMetadata {
		// Map WebSocket-specific metadata to MQTT equivalents
		switch key {
		case "websocket_opcode":
			// Convert WebSocket opcode to MQTT message type
			if opcode, ok := value.(int); ok {
				mqttMetadata["mqtt_message_type"] = t.mapOpcodeToMessageType(opcode)
			}
		case "websocket_fin":
			// MQTT doesn't have frame fragmentation
			continue
		case "websocket_mask":
			// MQTT doesn't use masking
			continue
		case "websocket_length":
			// Map to MQTT payload length
			mqttMetadata["mqtt_payload_length"] = value
		default:
			mqttMetadata[key] = value
		}
	}

	return mqttMetadata
}

// mapOpcodeToMessageType maps WebSocket opcodes to MQTT message types
func (t *WebSocketToMQTTTranslator) mapOpcodeToMessageType(opcode int) string {
	switch opcode {
	case 0x0: // Continuation frame
		return "continue"
	case 0x1: // Text frame
		return "publish"
	case 0x2: // Binary frame
		return "publish"
	case 0x8: // Connection close
		return "disconnect"
	case 0x9: // Ping
		return "pingreq"
	case 0xA: // Pong
		return "pingresp"
	default:
		return "publish"
	}
}

// SendToMQTT sends the translated message to the MQTT broker
func (t *WebSocketToMQTTTranslator) SendToMQTT(message *Message) error {
	// Connect to MQTT broker if not connected
	if t.conn == nil {
		conn, err := net.Dial("tcp", t.mqttAddr)
		if err != nil {
			return fmt.Errorf("failed to connect to MQTT broker: %w", err)
		}
		t.conn = conn
	}

	// Create MQTT PUBLISH packet
	publishPacket := t.createPublishPacket(message)

	// Send packet to MQTT broker
	_, err := t.conn.Write(publishPacket)
	if err != nil {
		return fmt.Errorf("failed to send MQTT message: %w", err)
	}

	return nil
}

// createPublishPacket creates an MQTT PUBLISH packet
func (t *WebSocketToMQTTTranslator) createPublishPacket(message *Message) []byte {
	// This is a simplified MQTT packet creation
	// In a real implementation, you would use a proper MQTT library

	topic := message.Topic
	if topic == "" {
		topic = "websocket/messages"
	}

	// Get QoS and retain from metadata
	qos := 0
	retain := false
	if message.Metadata != nil {
		if qosVal, exists := message.Metadata["mqtt_qos"]; exists {
			if qosInt, ok := qosVal.(int); ok {
				qos = qosInt
			}
		}
		if retainVal, exists := message.Metadata["mqtt_retain"]; exists {
			if retainBool, ok := retainVal.(bool); ok {
				retain = retainBool
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

	// Create MQTT PUBLISH packet (simplified)
	// Fixed header: Message Type (3) + DUP (0) + QoS (0-2) + RETAIN (0-1)
	fixedHeader := byte(0x30) // PUBLISH message type
	if retain {
		fixedHeader |= 0x01
	}
	fixedHeader |= byte(qos << 1)

	// Variable header: Topic Length + Topic + Packet Identifier (if QoS > 0)
	topicBytes := []byte(topic)
	topicLength := len(topicBytes)

	var variableHeader []byte
	variableHeader = append(variableHeader, byte(topicLength>>8), byte(topicLength&0xFF))
	variableHeader = append(variableHeader, topicBytes...)

	if qos > 0 {
		// Add packet identifier (simplified - using message ID hash)
		packetID := uint16(len(message.ID) % 65536)
		variableHeader = append(variableHeader, byte(packetID>>8), byte(packetID&0xFF))
	}

	// Payload
	payloadLength := len(payload)

	// Remaining length
	remainingLength := len(variableHeader) + payloadLength
	remainingLengthBytes := t.encodeRemainingLength(remainingLength)

	// Build complete packet
	var packet []byte
	packet = append(packet, fixedHeader)
	packet = append(packet, remainingLengthBytes...)
	packet = append(packet, variableHeader...)
	packet = append(packet, payload...)

	return packet
}

// encodeRemainingLength encodes the remaining length field for MQTT
func (t *WebSocketToMQTTTranslator) encodeRemainingLength(length int) []byte {
	var bytes []byte
	for {
		byteVal := length % 128
		length /= 128
		if length > 0 {
			byteVal |= 0x80
		}
		bytes = append(bytes, byte(byteVal))
		if length == 0 {
			break
		}
	}
	return bytes
}

// Close closes the MQTT connection
func (t *WebSocketToMQTTTranslator) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
