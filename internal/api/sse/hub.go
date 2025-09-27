package sse

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Hub manages SSE connections and event broadcasting
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger
	logger *slog.Logger

	// Configuration
	maxConnections int

	// Statistics
	totalConnections int
}

// NewHub creates a new SSE hub
func NewHub(logger *slog.Logger, maxConnections int) *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		logger:           logger,
		maxConnections:   maxConnections,
		totalConnections: 0,
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Send keep-alive every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.sendKeepAlive()
		case <-ctx.Done():
			h.logger.Info("SSE hub shutting down")
			return
		}
	}
}

// Register registers a new SSE client
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check connection limit
	if len(h.clients) >= h.maxConnections {
		h.logger.Warn("Maximum SSE connections reached", "maxConnections", h.maxConnections)
		client.Close()
		return
	}

	h.clients[client] = true
	h.totalConnections++

	h.logger.Info("SSE client registered",
		"clientId", client.ID,
		"totalClients", len(h.clients),
		"totalConnections", h.totalConnections,
	)
}

// Unregister unregisters an SSE client
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		client.Close()

		h.logger.Info("SSE client unregistered",
			"clientId", client.ID,
			"totalClients", len(h.clients),
		)
	}
}

// BroadcastEvent broadcasts an event to all connected clients
func (h *Hub) BroadcastEvent(eventType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	event := &Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	for client := range h.clients {
		select {
		case client.send <- event:
		default:
			// Client channel is full, remove client
			h.logger.Warn("SSE client channel full, removing client", "clientId", client.ID)
			go h.Unregister(client)
		}
	}
}

// BroadcastEventToTopic broadcasts an event to clients subscribed to a specific topic
func (h *Hub) BroadcastEventToTopic(topic, eventType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	event := &Event{
		Type:      eventType,
		Topic:     topic,
		Data:      data,
		Timestamp: time.Now(),
	}

	for client := range h.clients {
		// Check if client is subscribed to this topic
		if client.IsSubscribedToTopic(topic) {
			select {
			case client.send <- event:
			default:
				// Client channel is full, remove client
				h.logger.Warn("SSE client channel full, removing client", "clientId", client.ID)
				go h.Unregister(client)
			}
		}
	}
}

// GetActiveConnections returns the number of active connections
func (h *Hub) GetActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetTotalConnections returns the total number of connections since startup
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.totalConnections
}

// GetClients returns all connected clients
func (h *Hub) GetClients() map[*Client]bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy to avoid race conditions
	clients := make(map[*Client]bool)
	for client := range h.clients {
		clients[client] = true
	}
	return clients
}

// sendKeepAlive sends keep-alive events to all clients
func (h *Hub) sendKeepAlive() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	keepAliveEvent := &Event{
		Type:      "keepalive",
		Data:      map[string]interface{}{"timestamp": time.Now()},
		Timestamp: time.Now(),
	}

	for client := range h.clients {
		select {
		case client.send <- keepAliveEvent:
		default:
			// Client channel is full, remove client
			h.logger.Warn("SSE client channel full during keep-alive, removing client", "clientId", client.ID)
			go h.Unregister(client)
		}
	}
}
