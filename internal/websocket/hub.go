package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Subscription channels for different event types
	subscriptions map[string]map[*Client]bool

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger
	logger *slog.Logger
}

// Client represents a websocket client
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Client ID
	id string

	// Subscribed topics
	subscriptions map[string]bool

	// Hub reference
	hub *Hub

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// Message represents a websocket message
type Message struct {
	Type      string      `json:"type"`
	Topic     string      `json:"topic,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	ClientID  string      `json:"clientId,omitempty"`
}

// NewHub creates a new websocket hub
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscriptions: make(map[string]map[*Client]bool),
		logger:        logger,
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("Client registered", "clientId", client.id, "totalClients", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove client from all subscriptions
				for topic := range client.subscriptions {
					if clients, exists := h.subscriptions[topic]; exists {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.subscriptions, topic)
						}
					}
				}
			}
			h.mu.Unlock()
			h.logger.Info("Client unregistered", "clientId", client.id, "totalClients", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()

		case <-ctx.Done():
			h.logger.Info("Hub shutting down")
			return
		}
	}
}

// Subscribe subscribes a client to a topic
func (h *Hub) Subscribe(client *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subscriptions[topic] == nil {
		h.subscriptions[topic] = make(map[*Client]bool)
	}
	h.subscriptions[topic][client] = true
	client.subscriptions[topic] = true

	h.logger.Info("Client subscribed to topic", "clientId", client.id, "topic", topic)
}

// Unsubscribe unsubscribes a client from a topic
func (h *Hub) Unsubscribe(client *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, exists := h.subscriptions[topic]; exists {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.subscriptions, topic)
		}
	}
	delete(client.subscriptions, topic)

	h.logger.Info("Client unsubscribed from topic", "clientId", client.id, "topic", topic)
}

// BroadcastToTopic broadcasts a message to all clients subscribed to a specific topic
func (h *Hub) BroadcastToTopic(topic string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message := Message{
		Type:      "broadcast",
		Topic:     topic,
		Data:      data,
		Timestamp: time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", "error", err)
		return
	}

	if clients, exists := h.subscriptions[topic]; exists {
		for client := range clients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetSubscriptionCount returns the number of subscriptions for a topic
func (h *Hub) GetSubscriptionCount(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, exists := h.subscriptions[topic]; exists {
		return len(clients)
	}
	return 0
}

// GetTopics returns all active topics
func (h *Hub) GetTopics() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	topics := make([]string, 0, len(h.subscriptions))
	for topic := range h.subscriptions {
		topics = append(topics, topic)
	}
	return topics
}

// Client methods

// NewClient creates a new websocket client
func NewClient(conn *websocket.Conn, hub *Hub, id string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		conn:          conn,
		send:          make(chan []byte, 256),
		id:            id,
		subscriptions: make(map[string]bool),
		hub:           hub,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, messageBytes, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.hub.logger.Error("WebSocket error", "error", err)
				}
				return
			}

			var message Message
			if err := json.Unmarshal(messageBytes, &message); err != nil {
				c.hub.logger.Error("Failed to unmarshal message", "error", err)
				continue
			}

			message.ClientID = c.id
			c.handleMessage(&message)
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// handleMessage handles incoming messages from the client
func (c *Client) handleMessage(message *Message) {
	switch message.Type {
	case "subscribe":
		if topic, ok := message.Data.(string); ok {
			c.hub.Subscribe(c, topic)
		}
	case "unsubscribe":
		if topic, ok := message.Data.(string); ok {
			c.hub.Unsubscribe(c, topic)
		}
	case "ping":
		// Respond with pong
		pongMessage := Message{
			Type:      "pong",
			Timestamp: time.Now(),
			ClientID:  c.id,
		}
		if pongBytes, err := json.Marshal(pongMessage); err == nil {
			select {
			case c.send <- pongBytes:
			default:
				close(c.send)
			}
		}
	default:
		c.hub.logger.Warn("Unknown message type", "type", message.Type, "clientId", c.id)
	}
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	c.conn.Close()
}
