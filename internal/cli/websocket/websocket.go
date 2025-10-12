package websocket

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client for real-time updates
type Client struct {
	conn              *websocket.Conn
	url               string
	connected         bool
	mu                sync.RWMutex
	handlers          map[string][]EventHandler
	stopChan          chan struct{}
	reconnect         bool
	reconnectInterval time.Duration
}

// EventHandler handles incoming WebSocket events
type EventHandler func(event Event)

// Event represents a WebSocket event
type Event struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// New creates a new WebSocket client
func New(serverURL string) *Client {
	return &Client{
		url:               serverURL,
		handlers:          make(map[string][]EventHandler),
		stopChan:          make(chan struct{}),
		reconnect:         true,
		reconnectInterval: 5 * time.Second,
	}
}

// Connect connects to the WebSocket server
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Parse URL
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Convert HTTP to WebSocket scheme
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.connected = true

	// Start message handling
	go c.handleMessages()

	return nil
}

// Disconnect disconnects from the WebSocket server
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	close(c.stopChan)

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Subscribe subscribes to events of a specific type
func (c *Client) Subscribe(eventType string, handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handlers[eventType] = append(c.handlers[eventType], handler)
}

// Unsubscribe unsubscribes from events of a specific type
func (c *Client) Unsubscribe(eventType string, handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	handlers := c.handlers[eventType]
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			c.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Send sends a message to the server
func (c *Client) Send(eventType string, data map[string]interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.conn == nil {
		return fmt.Errorf("not connected")
	}

	event := Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	return c.conn.WriteJSON(event)
}

// handleMessages handles incoming WebSocket messages
func (c *Client) handleMessages() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			var event Event
			err := c.conn.ReadJSON(&event)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("WebSocket error: %v\n", err)
				}
				return
			}

			// Handle the event
			c.handleEvent(event)
		}
	}
}

// handleEvent handles a single event
func (c *Client) handleEvent(event Event) {
	c.mu.RLock()
	handlers := c.handlers[event.Type]
	c.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

// SetReconnect sets whether to automatically reconnect
func (c *Client) SetReconnect(reconnect bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reconnect = reconnect
}

// SetReconnectInterval sets the reconnect interval
func (c *Client) SetReconnectInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reconnectInterval = interval
}

// Reconnect attempts to reconnect to the server
func (c *Client) Reconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		_ = c.conn.Close() // Ignore error for demo purposes
	}

	c.connected = false
	c.stopChan = make(chan struct{})

	return c.Connect(ctx)
}

// Ping sends a ping message to the server
func (c *Client) Ping() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.WriteMessage(websocket.PingMessage, nil)
}

// SetPingHandler sets the ping handler
func (c *Client) SetPingHandler(handler func(string) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.SetPingHandler(handler)
	}
}

// SetPongHandler sets the pong handler
func (c *Client) SetPongHandler(handler func(string) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.SetPongHandler(handler)
	}
}

// SetCloseHandler sets the close handler
func (c *Client) SetCloseHandler(handler func(int, string) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.SetCloseHandler(handler)
	}
}
