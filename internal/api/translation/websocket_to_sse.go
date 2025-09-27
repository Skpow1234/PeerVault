package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebSocketToSSETranslator translates WebSocket messages to SSE format
type WebSocketToSSETranslator struct {
	sseAddr string
	client  *http.Client
}

// NewWebSocketToSSETranslator creates a new WebSocket to SSE translator
func NewWebSocketToSSETranslator(sseAddr string) *WebSocketToSSETranslator {
	return &WebSocketToSSETranslator{
		sseAddr: sseAddr,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Translate translates a WebSocket message to SSE format
func (t *WebSocketToSSETranslator) Translate(message *Message) (*Message, error) {
	// Convert WebSocket message to SSE format
	sseMessage := &Message{
		ID:        message.ID,
		Protocol:  "sse",
		Type:      t.mapMessageType(message.Type),
		Topic:     message.Topic,
		Payload:   message.Payload,
		Headers:   t.mapHeaders(message.Headers),
		Metadata:  t.mapMetadata(message.Metadata),
		Timestamp: message.Timestamp,
	}

	// Add SSE-specific headers
	if sseMessage.Headers == nil {
		sseMessage.Headers = make(map[string]string)
	}
	sseMessage.Headers["Content-Type"] = "text/event-stream"
	sseMessage.Headers["Cache-Control"] = "no-cache"
	sseMessage.Headers["Connection"] = "keep-alive"

	// Add SSE-specific metadata
	if sseMessage.Metadata == nil {
		sseMessage.Metadata = make(map[string]interface{})
	}
	sseMessage.Metadata["sse_event"] = t.mapEventType(message.Type)
	sseMessage.Metadata["sse_id"] = message.ID
	sseMessage.Metadata["sse_retry"] = 3000 // 3 seconds

	return sseMessage, nil
}

// CanTranslate checks if this translator can handle the given protocols
func (t *WebSocketToSSETranslator) CanTranslate(from, to string) bool {
	return from == "websocket" && to == "sse"
}

// GetSupportedProtocols returns the protocols this translator supports
func (t *WebSocketToSSETranslator) GetSupportedProtocols() []string {
	return []string{"websocket", "sse"}
}

// mapMessageType maps WebSocket message types to SSE equivalents
func (t *WebSocketToSSETranslator) mapMessageType(wsType string) string {
	switch wsType {
	case "text", "message":
		return "data"
	case "binary":
		return "data" // SSE doesn't support binary, convert to base64
	case "ping":
		return "ping"
	case "pong":
		return "pong"
	case "close":
		return "close"
	case "error":
		return "error"
	default:
		return "data"
	}
}

// mapEventType maps WebSocket message types to SSE event types
func (t *WebSocketToSSETranslator) mapEventType(wsType string) string {
	switch wsType {
	case "text", "message":
		return "message"
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
	case "subscribe":
		return "subscription"
	case "unsubscribe":
		return "unsubscription"
	default:
		return "message"
	}
}

// mapHeaders maps WebSocket headers to SSE headers
func (t *WebSocketToSSETranslator) mapHeaders(wsHeaders map[string]string) map[string]string {
	if wsHeaders == nil {
		return nil
	}

	sseHeaders := make(map[string]string)
	for key, value := range wsHeaders {
		// Map WebSocket-specific headers to SSE equivalents
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
			sseHeaders["Connection"] = "keep-alive"
		default:
			sseHeaders[key] = value
		}
	}

	return sseHeaders
}

// mapMetadata maps WebSocket metadata to SSE metadata
func (t *WebSocketToSSETranslator) mapMetadata(wsMetadata map[string]interface{}) map[string]interface{} {
	if wsMetadata == nil {
		return nil
	}

	sseMetadata := make(map[string]interface{})
	for key, value := range wsMetadata {
		// Map WebSocket-specific metadata to SSE equivalents
		switch key {
		case "websocket_opcode":
			// Convert WebSocket opcode to SSE event type
			if opcode, ok := value.(int); ok {
				sseMetadata["sse_event"] = t.mapOpcodeToEvent(opcode)
			}
		case "websocket_fin":
			// SSE doesn't have frame fragmentation
			continue
		case "websocket_mask":
			// SSE doesn't use masking
			continue
		default:
			sseMetadata[key] = value
		}
	}

	return sseMetadata
}

// mapOpcodeToEvent maps WebSocket opcodes to SSE event types
func (t *WebSocketToSSETranslator) mapOpcodeToEvent(opcode int) string {
	switch opcode {
	case 0x0: // Continuation frame
		return "continuation"
	case 0x1: // Text frame
		return "message"
	case 0x2: // Binary frame
		return "binary"
	case 0x8: // Connection close
		return "close"
	case 0x9: // Ping
		return "ping"
	case 0xA: // Pong
		return "pong"
	default:
		return "message"
	}
}

// SendToSSE sends the translated message to the SSE server
func (t *WebSocketToSSETranslator) SendToSSE(message *Message) error {
	// Create SSE event format
	eventData := t.formatSSEEvent(message)

	// Send to SSE server
	url := fmt.Sprintf("http://%s/sse", t.sseAddr)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(eventData))
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Add custom headers from message
	for key, value := range message.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SSE message: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't fail the operation
			// as the main operation has already completed
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE server returned status %d", resp.StatusCode)
	}

	return nil
}

// formatSSEEvent formats a message as an SSE event
func (t *WebSocketToSSETranslator) formatSSEEvent(message *Message) string {
	var event bytes.Buffer

	// Add event type
	if eventType, exists := message.Metadata["sse_event"]; exists {
		event.WriteString(fmt.Sprintf("event: %v\n", eventType))
	}

	// Add event ID
	if eventID, exists := message.Metadata["sse_id"]; exists {
		event.WriteString(fmt.Sprintf("id: %v\n", eventID))
	}

	// Add retry interval
	if retry, exists := message.Metadata["sse_retry"]; exists {
		event.WriteString(fmt.Sprintf("retry: %v\n", retry))
	}

	// Add data
	event.WriteString("data: ")
	if payloadStr, ok := message.Payload.(string); ok {
		event.WriteString(payloadStr)
	} else {
		// Convert to JSON if not string
		payloadJSON, err := json.Marshal(message.Payload)
		if err != nil {
			event.WriteString(fmt.Sprintf("Error marshaling payload: %v", err))
		} else {
			event.Write(payloadJSON)
		}
	}
	event.WriteString("\n\n")

	return event.String()
}
