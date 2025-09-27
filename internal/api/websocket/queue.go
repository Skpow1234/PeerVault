package websocket

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// MessageQueue represents a message queue for WebSocket connections
type MessageQueue struct {
	messages chan QueuedMessage
	clients  map[string]*QueuedClient
	mu       sync.RWMutex
	logger   *slog.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

// QueuedMessage represents a message in the queue
type QueuedMessage struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Topic     string        `json:"topic,omitempty"`
	Data      interface{}   `json:"data"`
	Timestamp time.Time     `json:"timestamp"`
	Priority  int           `json:"priority"` // Higher number = higher priority
	TTL       time.Duration `json:"ttl,omitempty"`
}

// QueuedClient represents a client with queuing capabilities
type QueuedClient struct {
	ID           string
	Queue        chan QueuedMessage
	MaxQueueSize int
	LastSeen     time.Time
	IsActive     bool
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(logger *slog.Logger) *MessageQueue {
	ctx, cancel := context.WithCancel(context.Background())

	queue := &MessageQueue{
		messages: make(chan QueuedMessage, 1000), // Buffer for 1000 messages
		clients:  make(map[string]*QueuedClient),
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start the queue processor
	go queue.processMessages()

	return queue
}

// AddClient adds a client to the queue system
func (mq *MessageQueue) AddClient(clientID string, maxQueueSize int) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.clients[clientID] = &QueuedClient{
		ID:           clientID,
		Queue:        make(chan QueuedMessage, maxQueueSize),
		MaxQueueSize: maxQueueSize,
		LastSeen:     time.Now(),
		IsActive:     true,
	}

	mq.logger.Info("Client added to message queue", "clientId", clientID, "maxQueueSize", maxQueueSize)
}

// RemoveClient removes a client from the queue system
func (mq *MessageQueue) RemoveClient(clientID string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if client, exists := mq.clients[clientID]; exists {
		close(client.Queue)
		delete(mq.clients, clientID)
		mq.logger.Info("Client removed from message queue", "clientId", clientID)
	}
}

// EnqueueMessage adds a message to the queue
func (mq *MessageQueue) EnqueueMessage(message QueuedMessage) {
	select {
	case mq.messages <- message:
		mq.logger.Debug("Message enqueued", "messageId", message.ID, "type", message.Type)
	default:
		mq.logger.Warn("Message queue full, dropping message", "messageId", message.ID)
	}
}

// EnqueueMessageToClient adds a message to a specific client's queue
func (mq *MessageQueue) EnqueueMessageToClient(clientID string, message QueuedMessage) {
	mq.mu.RLock()
	client, exists := mq.clients[clientID]
	mq.mu.RUnlock()

	if !exists {
		mq.logger.Warn("Client not found for message", "clientId", clientID, "messageId", message.ID)
		return
	}

	select {
	case client.Queue <- message:
		client.LastSeen = time.Now()
		mq.logger.Debug("Message enqueued to client", "clientId", clientID, "messageId", message.ID)
	default:
		mq.logger.Warn("Client queue full, dropping message", "clientId", clientID, "messageId", message.ID)
	}
}

// GetClientQueue returns the message queue for a specific client
func (mq *MessageQueue) GetClientQueue(clientID string) <-chan QueuedMessage {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	if client, exists := mq.clients[clientID]; exists {
		return client.Queue
	}
	return nil
}

// GetQueueStats returns statistics about the message queue
func (mq *MessageQueue) GetQueueStats() map[string]interface{} {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients":    len(mq.clients),
		"active_clients":   0,
		"total_queue_size": 0,
		"clients":          make([]map[string]interface{}, 0, len(mq.clients)),
	}

	for clientID, client := range mq.clients {
		if client.IsActive {
			stats["active_clients"] = stats["active_clients"].(int) + 1
		}

		queueSize := len(client.Queue)
		stats["total_queue_size"] = stats["total_queue_size"].(int) + queueSize

		clientStats := map[string]interface{}{
			"id":             clientID,
			"queue_size":     queueSize,
			"max_queue_size": client.MaxQueueSize,
			"last_seen":      client.LastSeen,
			"is_active":      client.IsActive,
		}
		stats["clients"] = append(stats["clients"].([]map[string]interface{}), clientStats)
	}

	return stats
}

// processMessages processes messages from the main queue
func (mq *MessageQueue) processMessages() {
	for {
		select {
		case message := <-mq.messages:
			mq.distributeMessage(message)
		case <-mq.ctx.Done():
			mq.logger.Info("Message queue processor shutting down")
			return
		}
	}
}

// distributeMessage distributes a message to appropriate clients
func (mq *MessageQueue) distributeMessage(message QueuedMessage) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	// Check if message has expired
	if message.TTL > 0 && time.Since(message.Timestamp) > message.TTL {
		mq.logger.Debug("Message expired, dropping", "messageId", message.ID)
		return
	}

	// If message has a topic, send to clients subscribed to that topic
	if message.Topic != "" {
		// TODO: Implement topic-based distribution
		// For now, send to all active clients
		for clientID, client := range mq.clients {
			if client.IsActive {
				mq.EnqueueMessageToClient(clientID, message)
			}
		}
	} else {
		// Broadcast to all active clients
		for clientID, client := range mq.clients {
			if client.IsActive {
				mq.EnqueueMessageToClient(clientID, message)
			}
		}
	}
}

// CleanupInactiveClients removes clients that haven't been seen for a while
func (mq *MessageQueue) CleanupInactiveClients(timeout time.Duration) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	now := time.Now()
	for clientID, client := range mq.clients {
		if now.Sub(client.LastSeen) > timeout {
			client.IsActive = false
			mq.logger.Info("Client marked as inactive", "clientId", clientID, "lastSeen", client.LastSeen)
		}
	}
}

// Shutdown gracefully shuts down the message queue
func (mq *MessageQueue) Shutdown() {
	mq.cancel()

	mq.mu.Lock()
	defer mq.mu.Unlock()

	// Close all client queues
	for clientID, client := range mq.clients {
		close(client.Queue)
		mq.logger.Info("Client queue closed", "clientId", clientID)
	}

	mq.logger.Info("Message queue shutdown complete")
}
