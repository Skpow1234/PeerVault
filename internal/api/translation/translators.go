package translation

import (
	"fmt"
)

// SSEToWebSocketTranslator translates SSE messages to WebSocket format
type SSEToWebSocketTranslator struct {
	websocketAddr string
}

// NewSSEToWebSocketTranslator creates a new SSE to WebSocket translator
func NewSSEToWebSocketTranslator(websocketAddr string) *SSEToWebSocketTranslator {
	return &SSEToWebSocketTranslator{
		websocketAddr: websocketAddr,
	}
}

func (t *SSEToWebSocketTranslator) Translate(message *Message) (*Message, error) {
	wsMessage := &Message{
		ID:        message.ID,
		Protocol:  "websocket",
		Type:      t.mapMessageType(message.Type),
		Topic:     message.Topic,
		Payload:   message.Payload,
		Headers:   t.mapHeaders(message.Headers),
		Metadata:  t.mapMetadata(message.Metadata),
		Timestamp: message.Timestamp,
	}

	// Add WebSocket-specific metadata
	if wsMessage.Metadata == nil {
		wsMessage.Metadata = make(map[string]interface{})
	}
	wsMessage.Metadata["websocket_opcode"] = t.mapOpcode(message.Type)
	wsMessage.Metadata["websocket_fin"] = true

	return wsMessage, nil
}

func (t *SSEToWebSocketTranslator) CanTranslate(from, to string) bool {
	return from == "sse" && to == "websocket"
}

func (t *SSEToWebSocketTranslator) GetSupportedProtocols() []string {
	return []string{"sse", "websocket"}
}

func (t *SSEToWebSocketTranslator) mapMessageType(sseType string) string {
	switch sseType {
	case "data":
		return "text"
	case "ping":
		return "ping"
	case "pong":
		return "pong"
	case "close":
		return "close"
	case "error":
		return "error"
	default:
		return "text"
	}
}

func (t *SSEToWebSocketTranslator) mapOpcode(sseType string) int {
	switch sseType {
	case "data":
		return 0x1 // Text frame
	case "ping":
		return 0x9 // Ping
	case "pong":
		return 0xA // Pong
	case "close":
		return 0x8 // Close
	case "error":
		return 0x8 // Close
	default:
		return 0x1 // Text frame
	}
}

func (t *SSEToWebSocketTranslator) mapHeaders(sseHeaders map[string]string) map[string]string {
	if sseHeaders == nil {
		return nil
	}

	wsHeaders := make(map[string]string)
	for key, value := range sseHeaders {
		switch key {
		case "Content-Type":
			wsHeaders["Content-Type"] = value
		case "Cache-Control":
			// Skip SSE-specific headers
			continue
		case "Connection":
			wsHeaders["Connection"] = "Upgrade"
		default:
			wsHeaders[key] = value
		}
	}

	// Add WebSocket-specific headers
	wsHeaders["Upgrade"] = "websocket"
	wsHeaders["Sec-WebSocket-Version"] = "13"

	return wsHeaders
}

func (t *SSEToWebSocketTranslator) mapMetadata(sseMetadata map[string]interface{}) map[string]interface{} {
	if sseMetadata == nil {
		return nil
	}

	wsMetadata := make(map[string]interface{})
	for key, value := range sseMetadata {
		switch key {
		case "sse_event":
			// Map SSE event to WebSocket message type
			if event, ok := value.(string); ok {
				wsMetadata["websocket_message_type"] = t.mapEventToMessageType(event)
			}
		case "sse_id":
			wsMetadata["websocket_message_id"] = value
		case "sse_retry":
			// Skip SSE-specific metadata
			continue
		default:
			wsMetadata[key] = value
		}
	}

	return wsMetadata
}

func (t *SSEToWebSocketTranslator) mapEventToMessageType(event string) string {
	switch event {
	case "message":
		return "text"
	case "binary":
		return "binary"
	case "ping":
		return "ping"
	case "pong":
		return "pong"
	case "close":
		return "close"
	case "error":
		return "error"
	default:
		return "text"
	}
}

// SSEToMQTTTranslator translates SSE messages to MQTT format
type SSEToMQTTTranslator struct {
	mqttAddr string
}

func NewSSEToMQTTTranslator(mqttAddr string) *SSEToMQTTTranslator {
	return &SSEToMQTTTranslator{
		mqttAddr: mqttAddr,
	}
}

func (t *SSEToMQTTTranslator) Translate(message *Message) (*Message, error) {
	mqttMessage := &Message{
		ID:        message.ID,
		Protocol:  "mqtt",
		Type:      "publish",
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
	mqttMessage.Metadata["mqtt_qos"] = 0
	mqttMessage.Metadata["mqtt_retain"] = false
	mqttMessage.Metadata["mqtt_dup"] = false

	return mqttMessage, nil
}

func (t *SSEToMQTTTranslator) CanTranslate(from, to string) bool {
	return from == "sse" && to == "mqtt"
}

func (t *SSEToMQTTTranslator) GetSupportedProtocols() []string {
	return []string{"sse", "mqtt"}
}

func (t *SSEToMQTTTranslator) mapTopic(sseTopic string) string {
	if sseTopic == "" {
		return "sse/messages"
	}
	return fmt.Sprintf("sse/%s", sseTopic)
}

func (t *SSEToMQTTTranslator) mapHeaders(sseHeaders map[string]string) map[string]string {
	if sseHeaders == nil {
		return nil
	}

	mqttHeaders := make(map[string]string)
	for key, value := range sseHeaders {
		switch key {
		case "Content-Type":
			mqttHeaders["content_type"] = value
		case "Cache-Control":
			// Skip SSE-specific headers
			continue
		case "Connection":
			// Skip SSE-specific headers
			continue
		default:
			mqttHeaders[key] = value
		}
	}

	return mqttHeaders
}

func (t *SSEToMQTTTranslator) mapMetadata(sseMetadata map[string]interface{}) map[string]interface{} {
	if sseMetadata == nil {
		return nil
	}

	mqttMetadata := make(map[string]interface{})
	for key, value := range sseMetadata {
		switch key {
		case "sse_event":
			mqttMetadata["mqtt_message_type"] = value
		case "sse_id":
			mqttMetadata["mqtt_message_id"] = value
		case "sse_retry":
			// Skip SSE-specific metadata
			continue
		default:
			mqttMetadata[key] = value
		}
	}

	return mqttMetadata
}

// SSEToCoAPTranslator translates SSE messages to CoAP format
type SSEToCoAPTranslator struct {
	coapAddr string
}

func NewSSEToCoAPTranslator(coapAddr string) *SSEToCoAPTranslator {
	return &SSEToCoAPTranslator{
		coapAddr: coapAddr,
	}
}

func (t *SSEToCoAPTranslator) Translate(message *Message) (*Message, error) {
	coapMessage := &Message{
		ID:        message.ID,
		Protocol:  "coap",
		Type:      "request",
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
	coapMessage.Metadata["coap_method"] = "POST"
	coapMessage.Metadata["coap_code"] = "2.05"
	coapMessage.Metadata["coap_type"] = "CON"
	coapMessage.Metadata["coap_token"] = t.generateToken(message.ID)

	return coapMessage, nil
}

func (t *SSEToCoAPTranslator) CanTranslate(from, to string) bool {
	return from == "sse" && to == "coap"
}

func (t *SSEToCoAPTranslator) GetSupportedProtocols() []string {
	return []string{"sse", "coap"}
}

func (t *SSEToCoAPTranslator) mapTopic(sseTopic string) string {
	if sseTopic == "" {
		return "/sse/messages"
	}
	return fmt.Sprintf("/sse%s", sseTopic)
}

func (t *SSEToCoAPTranslator) mapHeaders(sseHeaders map[string]string) map[string]string {
	if sseHeaders == nil {
		return nil
	}

	coapHeaders := make(map[string]string)
	for key, value := range sseHeaders {
		switch key {
		case "Content-Type":
			coapHeaders["content_format"] = t.mapContentType(value)
		case "Cache-Control":
			// Skip SSE-specific headers
			continue
		case "Connection":
			// Skip SSE-specific headers
			continue
		default:
			coapHeaders[key] = value
		}
	}

	return coapHeaders
}

func (t *SSEToCoAPTranslator) mapContentType(contentType string) string {
	switch contentType {
	case "application/json":
		return "50" // JSON
	case "text/plain":
		return "0" // Text/plain
	case "application/octet-stream":
		return "42" // Octet-stream
	default:
		return "0" // Default to text/plain
	}
}

func (t *SSEToCoAPTranslator) mapMetadata(sseMetadata map[string]interface{}) map[string]interface{} {
	if sseMetadata == nil {
		return nil
	}

	coapMetadata := make(map[string]interface{})
	for key, value := range sseMetadata {
		switch key {
		case "sse_event":
			coapMetadata["coap_message_type"] = value
		case "sse_id":
			coapMetadata["coap_message_id"] = value
		case "sse_retry":
			// Skip SSE-specific metadata
			continue
		default:
			coapMetadata[key] = value
		}
	}

	return coapMetadata
}

func (t *SSEToCoAPTranslator) generateToken(messageID string) string {
	token := fmt.Sprintf("sse_%s", messageID)
	if len(token) > 8 {
		token = token[:8]
	}
	return token
}

// MQTTToWebSocketTranslator translates MQTT messages to WebSocket format
type MQTTToWebSocketTranslator struct {
	websocketAddr string
}

func NewMQTTToWebSocketTranslator(websocketAddr string) *MQTTToWebSocketTranslator {
	return &MQTTToWebSocketTranslator{
		websocketAddr: websocketAddr,
	}
}

func (t *MQTTToWebSocketTranslator) Translate(message *Message) (*Message, error) {
	wsMessage := &Message{
		ID:        message.ID,
		Protocol:  "websocket",
		Type:      t.mapMessageType(message.Type),
		Topic:     t.mapTopic(message.Topic),
		Payload:   message.Payload,
		Headers:   t.mapHeaders(message.Headers),
		Metadata:  t.mapMetadata(message.Metadata),
		Timestamp: message.Timestamp,
	}

	// Add WebSocket-specific metadata
	if wsMessage.Metadata == nil {
		wsMessage.Metadata = make(map[string]interface{})
	}
	wsMessage.Metadata["websocket_opcode"] = t.mapOpcode(message.Type)
	wsMessage.Metadata["websocket_fin"] = true

	return wsMessage, nil
}

func (t *MQTTToWebSocketTranslator) CanTranslate(from, to string) bool {
	return from == "mqtt" && to == "websocket"
}

func (t *MQTTToWebSocketTranslator) GetSupportedProtocols() []string {
	return []string{"mqtt", "websocket"}
}

func (t *MQTTToWebSocketTranslator) mapMessageType(mqttType string) string {
	switch mqttType {
	case "publish":
		return "text"
	case "subscribe":
		return "subscribe"
	case "unsubscribe":
		return "unsubscribe"
	case "pingreq":
		return "ping"
	case "pingresp":
		return "pong"
	case "disconnect":
		return "close"
	default:
		return "text"
	}
}

func (t *MQTTToWebSocketTranslator) mapTopic(mqttTopic string) string {
	// Remove MQTT topic prefix if present
	if len(mqttTopic) > 3 && mqttTopic[:3] == "ws/" {
		return mqttTopic[3:]
	}
	return mqttTopic
}

func (t *MQTTToWebSocketTranslator) mapOpcode(mqttType string) int {
	switch mqttType {
	case "publish":
		return 0x1 // Text frame
	case "subscribe":
		return 0x1 // Text frame
	case "unsubscribe":
		return 0x1 // Text frame
	case "pingreq":
		return 0x9 // Ping
	case "pingresp":
		return 0xA // Pong
	case "disconnect":
		return 0x8 // Close
	default:
		return 0x1 // Text frame
	}
}

func (t *MQTTToWebSocketTranslator) mapHeaders(mqttHeaders map[string]string) map[string]string {
	if mqttHeaders == nil {
		return nil
	}

	wsHeaders := make(map[string]string)
	for key, value := range mqttHeaders {
		switch key {
		case "content_type":
			wsHeaders["Content-Type"] = value
		case "content_encoding":
			wsHeaders["Content-Encoding"] = value
		case "user_property":
			wsHeaders["User-Agent"] = value
		default:
			if len(key) > 14 && key[:14] == "user_property_" {
				wsHeaders[key[14:]] = value
			}
		}
	}

	// Add WebSocket-specific headers
	wsHeaders["Upgrade"] = "websocket"
	wsHeaders["Connection"] = "Upgrade"
	wsHeaders["Sec-WebSocket-Version"] = "13"

	return wsHeaders
}

func (t *MQTTToWebSocketTranslator) mapMetadata(mqttMetadata map[string]interface{}) map[string]interface{} {
	if mqttMetadata == nil {
		return nil
	}

	wsMetadata := make(map[string]interface{})
	for key, value := range mqttMetadata {
		switch key {
		case "mqtt_qos":
			wsMetadata["qos"] = value
		case "mqtt_retain":
			wsMetadata["retain"] = value
		case "mqtt_dup":
			wsMetadata["duplicate"] = value
		case "mqtt_message_type":
			wsMetadata["websocket_message_type"] = value
		case "mqtt_message_id":
			wsMetadata["websocket_message_id"] = value
		default:
			wsMetadata[key] = value
		}
	}

	return wsMetadata
}

// MQTTToSSETranslator translates MQTT messages to SSE format
type MQTTToSSETranslator struct {
	sseAddr string
}

func NewMQTTToSSETranslator(sseAddr string) *MQTTToSSETranslator {
	return &MQTTToSSETranslator{
		sseAddr: sseAddr,
	}
}

func (t *MQTTToSSETranslator) Translate(message *Message) (*Message, error) {
	sseMessage := &Message{
		ID:        message.ID,
		Protocol:  "sse",
		Type:      "data",
		Topic:     t.mapTopic(message.Topic),
		Payload:   message.Payload,
		Headers:   t.mapHeaders(message.Headers),
		Metadata:  t.mapMetadata(message.Metadata),
		Timestamp: message.Timestamp,
	}

	// Add SSE-specific metadata
	if sseMessage.Metadata == nil {
		sseMessage.Metadata = make(map[string]interface{})
	}
	sseMessage.Metadata["sse_event"] = "message"
	sseMessage.Metadata["sse_id"] = message.ID
	sseMessage.Metadata["sse_retry"] = 3000

	return sseMessage, nil
}

func (t *MQTTToSSETranslator) CanTranslate(from, to string) bool {
	return from == "mqtt" && to == "sse"
}

func (t *MQTTToSSETranslator) GetSupportedProtocols() []string {
	return []string{"mqtt", "sse"}
}

func (t *MQTTToSSETranslator) mapTopic(mqttTopic string) string {
	// Remove MQTT topic prefix if present
	if len(mqttTopic) > 4 && mqttTopic[:4] == "sse/" {
		return mqttTopic[4:]
	}
	return mqttTopic
}

func (t *MQTTToSSETranslator) mapHeaders(mqttHeaders map[string]string) map[string]string {
	if mqttHeaders == nil {
		return nil
	}

	sseHeaders := make(map[string]string)
	for key, value := range mqttHeaders {
		switch key {
		case "content_type":
			sseHeaders["Content-Type"] = value
		case "content_encoding":
			sseHeaders["Content-Encoding"] = value
		default:
			sseHeaders[key] = value
		}
	}

	// Add SSE-specific headers
	sseHeaders["Content-Type"] = "text/event-stream"
	sseHeaders["Cache-Control"] = "no-cache"
	sseHeaders["Connection"] = "keep-alive"

	return sseHeaders
}

func (t *MQTTToSSETranslator) mapMetadata(mqttMetadata map[string]interface{}) map[string]interface{} {
	if mqttMetadata == nil {
		return nil
	}

	sseMetadata := make(map[string]interface{})
	for key, value := range mqttMetadata {
		switch key {
		case "mqtt_qos":
			sseMetadata["qos"] = value
		case "mqtt_retain":
			sseMetadata["retain"] = value
		case "mqtt_dup":
			sseMetadata["duplicate"] = value
		case "mqtt_message_type":
			sseMetadata["sse_event"] = value
		case "mqtt_message_id":
			sseMetadata["sse_id"] = value
		default:
			sseMetadata[key] = value
		}
	}

	return sseMetadata
}

// MQTTToCoAPTranslator translates MQTT messages to CoAP format
type MQTTToCoAPTranslator struct {
	coapAddr string
}

func NewMQTTToCoAPTranslator(coapAddr string) *MQTTToCoAPTranslator {
	return &MQTTToCoAPTranslator{
		coapAddr: coapAddr,
	}
}

func (t *MQTTToCoAPTranslator) Translate(message *Message) (*Message, error) {
	coapMessage := &Message{
		ID:        message.ID,
		Protocol:  "coap",
		Type:      "request",
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
	coapMessage.Metadata["coap_method"] = "POST"
	coapMessage.Metadata["coap_code"] = "2.05"
	coapMessage.Metadata["coap_type"] = "CON"
	coapMessage.Metadata["coap_token"] = t.generateToken(message.ID)

	return coapMessage, nil
}

func (t *MQTTToCoAPTranslator) CanTranslate(from, to string) bool {
	return from == "mqtt" && to == "coap"
}

func (t *MQTTToCoAPTranslator) GetSupportedProtocols() []string {
	return []string{"mqtt", "coap"}
}

func (t *MQTTToCoAPTranslator) mapTopic(mqttTopic string) string {
	// Remove MQTT topic prefix if present
	if len(mqttTopic) > 5 && mqttTopic[:5] == "coap/" {
		return mqttTopic[5:]
	}
	return fmt.Sprintf("/mqtt%s", mqttTopic)
}

func (t *MQTTToCoAPTranslator) mapHeaders(mqttHeaders map[string]string) map[string]string {
	if mqttHeaders == nil {
		return nil
	}

	coapHeaders := make(map[string]string)
	for key, value := range mqttHeaders {
		switch key {
		case "content_type":
			coapHeaders["content_format"] = t.mapContentType(value)
		case "content_encoding":
			coapHeaders["content_encoding"] = value
		default:
			coapHeaders[key] = value
		}
	}

	return coapHeaders
}

func (t *MQTTToCoAPTranslator) mapContentType(contentType string) string {
	switch contentType {
	case "application/json":
		return "50" // JSON
	case "text/plain":
		return "0" // Text/plain
	case "application/octet-stream":
		return "42" // Octet-stream
	default:
		return "0" // Default to text/plain
	}
}

func (t *MQTTToCoAPTranslator) mapMetadata(mqttMetadata map[string]interface{}) map[string]interface{} {
	if mqttMetadata == nil {
		return nil
	}

	coapMetadata := make(map[string]interface{})
	for key, value := range mqttMetadata {
		switch key {
		case "mqtt_qos":
			coapMetadata["coap_reliability"] = t.mapQoSToReliability(value)
		case "mqtt_retain":
			coapMetadata["coap_persistent"] = value
		case "mqtt_dup":
			coapMetadata["coap_duplicate"] = value
		case "mqtt_message_type":
			coapMetadata["coap_message_type"] = value
		case "mqtt_message_id":
			coapMetadata["coap_message_id"] = value
		default:
			coapMetadata[key] = value
		}
	}

	return coapMetadata
}

func (t *MQTTToCoAPTranslator) mapQoSToReliability(qos interface{}) string {
	if qosInt, ok := qos.(int); ok {
		switch qosInt {
		case 0:
			return "NON" // Non-confirmable
		case 1:
			return "CON" // Confirmable
		case 2:
			return "CON" // Confirmable
		default:
			return "CON" // Default to confirmable
		}
	}
	return "CON"
}

func (t *MQTTToCoAPTranslator) generateToken(messageID string) string {
	token := fmt.Sprintf("mqtt_%s", messageID)
	if len(token) > 8 {
		token = token[:8]
	}
	return token
}

// CoAPToWebSocketTranslator translates CoAP messages to WebSocket format
type CoAPToWebSocketTranslator struct {
	websocketAddr string
}

func NewCoAPToWebSocketTranslator(websocketAddr string) *CoAPToWebSocketTranslator {
	return &CoAPToWebSocketTranslator{
		websocketAddr: websocketAddr,
	}
}

func (t *CoAPToWebSocketTranslator) Translate(message *Message) (*Message, error) {
	wsMessage := &Message{
		ID:        message.ID,
		Protocol:  "websocket",
		Type:      t.mapMessageType(message.Type),
		Topic:     t.mapTopic(message.Topic),
		Payload:   message.Payload,
		Headers:   t.mapHeaders(message.Headers),
		Metadata:  t.mapMetadata(message.Metadata),
		Timestamp: message.Timestamp,
	}

	// Add WebSocket-specific metadata
	if wsMessage.Metadata == nil {
		wsMessage.Metadata = make(map[string]interface{})
	}
	wsMessage.Metadata["websocket_opcode"] = t.mapOpcode(message.Type)
	wsMessage.Metadata["websocket_fin"] = true

	return wsMessage, nil
}

func (t *CoAPToWebSocketTranslator) CanTranslate(from, to string) bool {
	return from == "coap" && to == "websocket"
}

func (t *CoAPToWebSocketTranslator) GetSupportedProtocols() []string {
	return []string{"coap", "websocket"}
}

func (t *CoAPToWebSocketTranslator) mapMessageType(coapType string) string {
	switch coapType {
	case "request":
		return "text"
	case "response":
		return "text"
	case "ping":
		return "ping"
	case "pong":
		return "pong"
	case "reset":
		return "close"
	default:
		return "text"
	}
}

func (t *CoAPToWebSocketTranslator) mapTopic(coapTopic string) string {
	// Remove CoAP URI prefix if present
	if len(coapTopic) > 3 && coapTopic[:3] == "/ws" {
		return coapTopic[3:]
	}
	return coapTopic
}

func (t *CoAPToWebSocketTranslator) mapOpcode(coapType string) int {
	switch coapType {
	case "request":
		return 0x1 // Text frame
	case "response":
		return 0x1 // Text frame
	case "ping":
		return 0x9 // Ping
	case "pong":
		return 0xA // Pong
	case "reset":
		return 0x8 // Close
	default:
		return 0x1 // Text frame
	}
}

func (t *CoAPToWebSocketTranslator) mapHeaders(coapHeaders map[string]string) map[string]string {
	if coapHeaders == nil {
		return nil
	}

	wsHeaders := make(map[string]string)
	for key, value := range coapHeaders {
		switch key {
		case "content_format":
			wsHeaders["Content-Type"] = t.mapContentFormat(value)
		case "content_encoding":
			wsHeaders["Content-Encoding"] = value
		case "authorization":
			wsHeaders["Authorization"] = value
		default:
			if len(key) > 7 && key[:7] == "custom_" {
				wsHeaders[key[7:]] = value
			}
		}
	}

	// Add WebSocket-specific headers
	wsHeaders["Upgrade"] = "websocket"
	wsHeaders["Connection"] = "Upgrade"
	wsHeaders["Sec-WebSocket-Version"] = "13"

	return wsHeaders
}

func (t *CoAPToWebSocketTranslator) mapContentFormat(contentFormat string) string {
	switch contentFormat {
	case "50":
		return "application/json"
	case "0":
		return "text/plain"
	case "42":
		return "application/octet-stream"
	case "41":
		return "application/xml"
	default:
		return "text/plain"
	}
}

func (t *CoAPToWebSocketTranslator) mapMetadata(coapMetadata map[string]interface{}) map[string]interface{} {
	if coapMetadata == nil {
		return nil
	}

	wsMetadata := make(map[string]interface{})
	for key, value := range coapMetadata {
		switch key {
		case "coap_method":
			wsMetadata["http_method"] = value
		case "coap_code":
			wsMetadata["http_status"] = value
		case "coap_type":
			wsMetadata["coap_message_type"] = value
		case "coap_token":
			wsMetadata["coap_token"] = value
		case "coap_reliability":
			wsMetadata["reliability"] = value
		case "coap_persistent":
			wsMetadata["persistent"] = value
		case "coap_duplicate":
			wsMetadata["duplicate"] = value
		default:
			wsMetadata[key] = value
		}
	}

	return wsMetadata
}

// CoAPToSSETranslator translates CoAP messages to SSE format
type CoAPToSSETranslator struct {
	sseAddr string
}

func NewCoAPToSSETranslator(sseAddr string) *CoAPToSSETranslator {
	return &CoAPToSSETranslator{
		sseAddr: sseAddr,
	}
}

func (t *CoAPToSSETranslator) Translate(message *Message) (*Message, error) {
	sseMessage := &Message{
		ID:        message.ID,
		Protocol:  "sse",
		Type:      "data",
		Topic:     t.mapTopic(message.Topic),
		Payload:   message.Payload,
		Headers:   t.mapHeaders(message.Headers),
		Metadata:  t.mapMetadata(message.Metadata),
		Timestamp: message.Timestamp,
	}

	// Add SSE-specific metadata
	if sseMessage.Metadata == nil {
		sseMessage.Metadata = make(map[string]interface{})
	}
	sseMessage.Metadata["sse_event"] = "message"
	sseMessage.Metadata["sse_id"] = message.ID
	sseMessage.Metadata["sse_retry"] = 3000

	return sseMessage, nil
}

func (t *CoAPToSSETranslator) CanTranslate(from, to string) bool {
	return from == "coap" && to == "sse"
}

func (t *CoAPToSSETranslator) GetSupportedProtocols() []string {
	return []string{"coap", "sse"}
}

func (t *CoAPToSSETranslator) mapTopic(coapTopic string) string {
	// Remove CoAP URI prefix if present
	if len(coapTopic) > 4 && coapTopic[:4] == "/sse" {
		return coapTopic[4:]
	}
	return coapTopic
}

func (t *CoAPToSSETranslator) mapHeaders(coapHeaders map[string]string) map[string]string {
	if coapHeaders == nil {
		return nil
	}

	sseHeaders := make(map[string]string)
	for key, value := range coapHeaders {
		switch key {
		case "content_format":
			sseHeaders["Content-Type"] = t.mapContentFormat(value)
		case "content_encoding":
			sseHeaders["Content-Encoding"] = value
		case "authorization":
			sseHeaders["Authorization"] = value
		default:
			sseHeaders[key] = value
		}
	}

	// Add SSE-specific headers
	sseHeaders["Content-Type"] = "text/event-stream"
	sseHeaders["Cache-Control"] = "no-cache"
	sseHeaders["Connection"] = "keep-alive"

	return sseHeaders
}

func (t *CoAPToSSETranslator) mapContentFormat(contentFormat string) string {
	switch contentFormat {
	case "50":
		return "application/json"
	case "0":
		return "text/plain"
	case "42":
		return "application/octet-stream"
	case "41":
		return "application/xml"
	default:
		return "text/plain"
	}
}

func (t *CoAPToSSETranslator) mapMetadata(coapMetadata map[string]interface{}) map[string]interface{} {
	if coapMetadata == nil {
		return nil
	}

	sseMetadata := make(map[string]interface{})
	for key, value := range coapMetadata {
		switch key {
		case "coap_method":
			sseMetadata["http_method"] = value
		case "coap_code":
			sseMetadata["http_status"] = value
		case "coap_type":
			sseMetadata["coap_message_type"] = value
		case "coap_token":
			sseMetadata["coap_token"] = value
		case "coap_reliability":
			sseMetadata["reliability"] = value
		case "coap_persistent":
			sseMetadata["persistent"] = value
		case "coap_duplicate":
			sseMetadata["duplicate"] = value
		default:
			sseMetadata[key] = value
		}
	}

	return sseMetadata
}

// CoAPToMQTTTranslator translates CoAP messages to MQTT format
type CoAPToMQTTTranslator struct {
	mqttAddr string
}

func NewCoAPToMQTTTranslator(mqttAddr string) *CoAPToMQTTTranslator {
	return &CoAPToMQTTTranslator{
		mqttAddr: mqttAddr,
	}
}

func (t *CoAPToMQTTTranslator) Translate(message *Message) (*Message, error) {
	mqttMessage := &Message{
		ID:        message.ID,
		Protocol:  "mqtt",
		Type:      "publish",
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

func (t *CoAPToMQTTTranslator) CanTranslate(from, to string) bool {
	return from == "coap" && to == "mqtt"
}

func (t *CoAPToMQTTTranslator) GetSupportedProtocols() []string {
	return []string{"coap", "mqtt"}
}

func (t *CoAPToMQTTTranslator) mapTopic(coapTopic string) string {
	// Remove CoAP URI prefix if present
	if len(coapTopic) > 5 && coapTopic[:5] == "/mqtt" {
		return coapTopic[5:]
	}
	return fmt.Sprintf("coap%s", coapTopic)
}

func (t *CoAPToMQTTTranslator) mapQoS(coapMetadata map[string]interface{}) int {
	if coapMetadata == nil {
		return 0
	}

	if reliability, exists := coapMetadata["coap_reliability"]; exists {
		if reliabilityStr, ok := reliability.(string); ok {
			switch reliabilityStr {
			case "CON":
				return 1 // QoS 1 for confirmable
			case "NON":
				return 0 // QoS 0 for non-confirmable
			}
		}
	}

	return 0 // Default to QoS 0
}

func (t *CoAPToMQTTTranslator) mapRetain(coapMetadata map[string]interface{}) bool {
	if coapMetadata == nil {
		return false
	}

	if persistent, exists := coapMetadata["coap_persistent"]; exists {
		if persistentBool, ok := persistent.(bool); ok {
			return persistentBool
		}
	}

	return false
}

func (t *CoAPToMQTTTranslator) mapHeaders(coapHeaders map[string]string) map[string]string {
	if coapHeaders == nil {
		return nil
	}

	mqttHeaders := make(map[string]string)
	for key, value := range coapHeaders {
		switch key {
		case "content_format":
			mqttHeaders["content_type"] = t.mapContentFormat(value)
		case "content_encoding":
			mqttHeaders["content_encoding"] = value
		case "authorization":
			mqttHeaders["authorization"] = value
		default:
			mqttHeaders[key] = value
		}
	}

	return mqttHeaders
}

func (t *CoAPToMQTTTranslator) mapContentFormat(contentFormat string) string {
	switch contentFormat {
	case "50":
		return "application/json"
	case "0":
		return "text/plain"
	case "42":
		return "application/octet-stream"
	case "41":
		return "application/xml"
	default:
		return "text/plain"
	}
}

func (t *CoAPToMQTTTranslator) mapMetadata(coapMetadata map[string]interface{}) map[string]interface{} {
	if coapMetadata == nil {
		return nil
	}

	mqttMetadata := make(map[string]interface{})
	for key, value := range coapMetadata {
		switch key {
		case "coap_method":
			mqttMetadata["http_method"] = value
		case "coap_code":
			mqttMetadata["http_status"] = value
		case "coap_type":
			mqttMetadata["coap_message_type"] = value
		case "coap_token":
			mqttMetadata["coap_token"] = value
		case "coap_reliability":
			mqttMetadata["mqtt_qos"] = t.mapReliabilityToQoS(value)
		case "coap_persistent":
			mqttMetadata["mqtt_retain"] = value
		case "coap_duplicate":
			mqttMetadata["mqtt_dup"] = value
		default:
			mqttMetadata[key] = value
		}
	}

	return mqttMetadata
}

func (t *CoAPToMQTTTranslator) mapReliabilityToQoS(reliability interface{}) int {
	if reliabilityStr, ok := reliability.(string); ok {
		switch reliabilityStr {
		case "CON":
			return 1 // QoS 1 for confirmable
		case "NON":
			return 0 // QoS 0 for non-confirmable
		}
	}
	return 0 // Default to QoS 0
}
