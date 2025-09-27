package sse

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Client represents an SSE client connection
type Client struct {
	// The HTTP response writer
	w http.ResponseWriter

	// The HTTP request
	r *http.Request

	// Buffered channel of outbound events
	send chan *Event

	// Client ID
	ID string

	// Subscribed topics
	subscriptions map[string]bool

	// Logger
	logger *slog.Logger

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Connection metadata
	remoteAddr  string
	userAgent   string
	connectedAt time.Time
}

// Event represents an SSE event
type Event struct {
	Type      string      `json:"type"`
	Topic     string      `json:"topic,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	ID        string      `json:"id,omitempty"`
}

// NewClient creates a new SSE client
func NewClient(w http.ResponseWriter, r *http.Request, logger *slog.Logger) *Client {
	ctx, cancel := context.WithCancel(r.Context())

	client := &Client{
		w:             w,
		r:             r,
		send:          make(chan *Event, 256), // Buffer for 256 events
		ID:            generateClientID(),
		subscriptions: make(map[string]bool),
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		remoteAddr:    r.RemoteAddr,
		userAgent:     r.UserAgent(),
		connectedAt:   time.Now(),
	}

	return client
}

// generateClientID generates a unique client ID
func generateClientID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes) // Ignore error for client ID generation
	return hex.EncodeToString(bytes)
}

// Handle handles the SSE client connection
func (c *Client) Handle() {
	// Start goroutine to send events
	go c.writePump()

	// Wait for context cancellation or connection close
	<-c.ctx.Done()

	c.logger.Info("SSE client connection closed",
		"clientId", c.ID,
		"remoteAddr", c.remoteAddr,
		"duration", time.Since(c.connectedAt),
	)
}

// writePump pumps events from the send channel to the client
func (c *Client) writePump() {
	defer c.Close()

	for {
		select {
		case event := <-c.send:
			if err := c.writeEvent(event); err != nil {
				c.logger.Error("Failed to write SSE event", "error", err, "clientId", c.ID)
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// writeEvent writes an SSE event to the client
func (c *Client) writeEvent(event *Event) error {
	// Write event ID
	if event.ID != "" {
		if _, err := fmt.Fprintf(c.w, "id: %s\n", event.ID); err != nil {
			return err
		}
	}

	// Write event type
	if event.Type != "" {
		if _, err := fmt.Fprintf(c.w, "event: %s\n", event.Type); err != nil {
			return err
		}
	}

	// Write event data
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Write data lines (SSE format requires splitting on newlines)
	lines := splitLines(string(data))
	for _, line := range lines {
		if _, err := fmt.Fprintf(c.w, "data: %s\n", line); err != nil {
			return err
		}
	}

	// Write empty line to end the event
	if _, err := fmt.Fprintf(c.w, "\n"); err != nil {
		return err
	}

	// Flush the response writer
	if flusher, ok := c.w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// splitLines splits a string into lines for SSE format
func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}

	var lines []string
	start := 0

	for i, char := range s {
		if char == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}

	// Add the last line if it doesn't end with newline
	if start < len(s) {
		lines = append(lines, s[start:])
	} else if len(lines) == 0 {
		lines = append(lines, "")
	}

	return lines
}

// SendEvent sends an event to the client
func (c *Client) SendEvent(eventType string, data interface{}) {
	event := &Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		ID:        generateEventID(),
	}

	select {
	case c.send <- event:
	default:
		c.logger.Warn("SSE client send channel full", "clientId", c.ID)
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes) // Ignore error for event ID generation
	return hex.EncodeToString(bytes)
}

// SubscribeToTopic subscribes the client to a topic
func (c *Client) SubscribeToTopic(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.subscriptions[topic] = true
	c.logger.Info("SSE client subscribed to topic", "clientId", c.ID, "topic", topic)
}

// UnsubscribeFromTopic unsubscribes the client from a topic
func (c *Client) UnsubscribeFromTopic(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.subscriptions, topic)
	c.logger.Info("SSE client unsubscribed from topic", "clientId", c.ID, "topic", topic)
}

// IsSubscribedToTopic checks if the client is subscribed to a topic
func (c *Client) IsSubscribedToTopic(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// If no subscriptions, client receives all events
	if len(c.subscriptions) == 0 {
		return true
	}

	return c.subscriptions[topic]
}

// GetSubscriptions returns the client's subscriptions
func (c *Client) GetSubscriptions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	topics := make([]string, 0, len(c.subscriptions))
	for topic := range c.subscriptions {
		topics = append(topics, topic)
	}

	return topics
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	close(c.send)
}

// GetConnectionInfo returns connection information
func (c *Client) GetConnectionInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"id":            c.ID,
		"remoteAddr":    c.remoteAddr,
		"userAgent":     c.userAgent,
		"connectedAt":   c.connectedAt,
		"duration":      time.Since(c.connectedAt).String(),
		"subscriptions": c.GetSubscriptions(),
	}
}
