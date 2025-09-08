package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Upgrader specifies parameters for upgrading an HTTP connection to a WebSocket
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin
		// In production, you should implement proper origin checking
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Handler handles WebSocket connections for GraphQL subscriptions
type Handler struct {
	hub    *Hub
	logger *slog.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, logger *slog.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// ServeHTTP handles WebSocket upgrade requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", "error", err)
		return
	}

	// Generate a unique client ID
	clientID := generateClientID()

	client := NewClient(conn, h.hub, clientID)
	h.hub.register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()

	h.logger.Info("WebSocket connection established", "clientId", clientID, "remoteAddr", r.RemoteAddr)
}

// generateClientID generates a unique client ID
func generateClientID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes) // Ignore error for client ID generation
	return hex.EncodeToString(bytes)
}

// GraphQLSubscriptionHandler handles GraphQL subscription connections
type GraphQLSubscriptionHandler struct {
	hub    *Hub
	logger *slog.Logger
}

// NewGraphQLSubscriptionHandler creates a new GraphQL subscription handler
func NewGraphQLSubscriptionHandler(hub *Hub, logger *slog.Logger) *GraphQLSubscriptionHandler {
	return &GraphQLSubscriptionHandler{
		hub:    hub,
		logger: logger,
	}
}

// ServeHTTP handles GraphQL subscription WebSocket connections
func (h *GraphQLSubscriptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade GraphQL subscription connection", "error", err)
		return
	}

	clientID := generateClientID()
	client := NewClient(conn, h.hub, clientID)
	h.hub.register <- client

	// Send initial connection acknowledgment
	initMessage := Message{
		Type:      "connection_init",
		Timestamp: time.Now(),
		ClientID:  clientID,
	}
	if initBytes, err := json.Marshal(initMessage); err == nil {
		select {
		case client.send <- initBytes:
		default:
			close(client.send)
		}
	}

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()

	h.logger.Info("GraphQL subscription connection established", "clientId", clientID, "remoteAddr", r.RemoteAddr)
}

// GraphQLSubscriptionMessage represents a GraphQL subscription message
type GraphQLSubscriptionMessage struct {
	ID      string                 `json:"id,omitempty"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// HandleGraphQLSubscription handles GraphQL subscription protocol messages
func (c *Client) HandleGraphQLSubscription(message *GraphQLSubscriptionMessage) {
	switch message.Type {
	case "start":
		// Handle subscription start
		if query, ok := message.Payload["query"].(string); ok {
			// TODO: Parse query and subscribe to appropriate topics
			_ = query // Use query variable to avoid unused variable warning
		}
	case "stop":
		// Handle subscription stop
		// TODO: Unsubscribe from topics
		_ = message.ID // Use message.ID to avoid unused variable warning
	case "connection_init":
		// Handle connection initialization
		ackMessage := GraphQLSubscriptionMessage{
			Type: "connection_ack",
		}
		if ackBytes, err := json.Marshal(ackMessage); err == nil {
			select {
			case c.send <- ackBytes:
			default:
				close(c.send)
			}
		}
	case "connection_terminate":
		// Handle connection termination
		c.Close()
	default:
		// Unknown message type - ignore
		_ = message.Type // Use message.Type to avoid unused variable warning
	}
}

// BroadcastGraphQLData broadcasts data to GraphQL subscribers
func (h *Hub) BroadcastGraphQLData(subscriptionID string, data interface{}) {
	message := GraphQLSubscriptionMessage{
		ID:   subscriptionID,
		Type: "data",
		Payload: map[string]interface{}{
			"data": data,
		},
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		// Log error if logger is available, otherwise ignore
		_ = err // Use err to avoid unused variable warning
		return
	}

	// Broadcast to all clients subscribed to this subscription
	h.mu.RLock()
	for client := range h.clients {
		select {
		case client.send <- messageBytes:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
	h.mu.RUnlock()
}

// SubscriptionManager manages GraphQL subscriptions
type SubscriptionManager struct {
	hub           *Hub
	subscriptions map[string]*GraphQLSubscription
	mu            sync.RWMutex
	logger        *slog.Logger
}

// GraphQLSubscription represents a GraphQL subscription
type GraphQLSubscription struct {
	ID        string
	Query     string
	Variables map[string]interface{}
	Client    *Client
	Topics    []string
	CreatedAt time.Time
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager(hub *Hub, logger *slog.Logger) *SubscriptionManager {
	return &SubscriptionManager{
		hub:           hub,
		subscriptions: make(map[string]*GraphQLSubscription),
		logger:        logger,
	}
}

// CreateSubscription creates a new GraphQL subscription
func (sm *SubscriptionManager) CreateSubscription(client *Client, query string, variables map[string]interface{}) (*GraphQLSubscription, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscriptionID := generateClientID()
	subscription := &GraphQLSubscription{
		ID:        subscriptionID,
		Query:     query,
		Variables: variables,
		Client:    client,
		Topics:    sm.extractTopicsFromQuery(query),
		CreatedAt: time.Now(),
	}

	sm.subscriptions[subscriptionID] = subscription

	// Subscribe client to relevant topics
	for _, topic := range subscription.Topics {
		sm.hub.Subscribe(client, topic)
	}

	sm.logger.Info("Created GraphQL subscription", "subscriptionId", subscriptionID, "topics", subscription.Topics)
	return subscription, nil
}

// RemoveSubscription removes a GraphQL subscription
func (sm *SubscriptionManager) RemoveSubscription(subscriptionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if subscription, exists := sm.subscriptions[subscriptionID]; exists {
		// Unsubscribe client from topics
		for _, topic := range subscription.Topics {
			sm.hub.Unsubscribe(subscription.Client, topic)
		}
		delete(sm.subscriptions, subscriptionID)
		sm.logger.Info("Removed GraphQL subscription", "subscriptionId", subscriptionID)
	}
}

// GetSubscription returns a subscription by ID
func (sm *SubscriptionManager) GetSubscription(subscriptionID string) (*GraphQLSubscription, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	subscription, exists := sm.subscriptions[subscriptionID]
	return subscription, exists
}

// GetSubscriptionsByTopic returns all subscriptions for a specific topic
func (sm *SubscriptionManager) GetSubscriptionsByTopic(topic string) []*GraphQLSubscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var subscriptions []*GraphQLSubscription
	for _, subscription := range sm.subscriptions {
		for _, subscriptionTopic := range subscription.Topics {
			if subscriptionTopic == topic {
				subscriptions = append(subscriptions, subscription)
				break
			}
		}
	}
	return subscriptions
}

// extractTopicsFromQuery extracts subscription topics from a GraphQL query
func (sm *SubscriptionManager) extractTopicsFromQuery(query string) []string {
	// This is a simplified implementation
	// In a real implementation, you would parse the GraphQL query and extract subscription fields
	topics := []string{}

	// Simple keyword-based topic extraction
	if contains(query, "fileUploaded") {
		topics = append(topics, "file.uploaded")
	}
	if contains(query, "fileDeleted") {
		topics = append(topics, "file.deleted")
	}
	if contains(query, "fileUpdated") {
		topics = append(topics, "file.updated")
	}
	if contains(query, "peerConnected") {
		topics = append(topics, "peer.connected")
	}
	if contains(query, "peerDisconnected") {
		topics = append(topics, "peer.disconnected")
	}
	if contains(query, "peerHealthChanged") {
		topics = append(topics, "peer.health_changed")
	}
	if contains(query, "systemMetricsUpdated") {
		topics = append(topics, "system.metrics_updated")
	}
	if contains(query, "performanceAlert") {
		topics = append(topics, "system.performance_alert")
	}

	return topics
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

// indexOf returns the index of the first occurrence of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
