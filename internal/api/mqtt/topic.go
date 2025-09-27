package mqtt

import (
	"log/slog"
	"sync"
)

// Topic represents an MQTT topic
type Topic struct {
	name        string
	subscribers map[*Client]QoS
	mu          sync.RWMutex
	logger      *slog.Logger
}

// NewTopic creates a new topic
func NewTopic(name string, logger *slog.Logger) *Topic {
	return &Topic{
		name:        name,
		subscribers: make(map[*Client]QoS),
		logger:      logger,
	}
}

// Subscribe subscribes a client to the topic
func (t *Topic) Subscribe(client *Client, qos QoS) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers[client] = qos
	t.logger.Debug("Client subscribed to topic",
		"topic", t.name,
		"clientId", client.ID,
		"qos", qos,
		"subscriberCount", len(t.subscribers),
	)
}

// Unsubscribe unsubscribes a client from the topic
func (t *Topic) Unsubscribe(client *Client) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.subscribers[client]; exists {
		delete(t.subscribers, client)
		t.logger.Debug("Client unsubscribed from topic",
			"topic", t.name,
			"clientId", client.ID,
			"subscriberCount", len(t.subscribers),
		)
	}
}

// Publish publishes a message to all subscribers
func (t *Topic) Publish(message *Message) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for client, qos := range t.subscribers {
		// Create message copy with appropriate QoS
		clientMessage := &Message{
			Topic:   message.Topic,
			Payload: message.Payload,
			QoS:     qos,
			Retain:  message.Retain,
		}

		// Send message to client
		select {
		case client.outgoingMessages <- clientMessage:
		default:
			// Client channel is full, skip this message
			t.logger.Warn("Client message channel full, skipping message",
				"topic", t.name,
				"clientId", client.ID,
			)
		}
	}

	t.logger.Debug("Message published to topic",
		"topic", t.name,
		"subscriberCount", len(t.subscribers),
		"qos", message.QoS,
		"retain", message.Retain,
	)
}

// SubscriberCount returns the number of subscribers
func (t *Topic) SubscriberCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.subscribers)
}

// GetSubscribers returns a copy of the subscribers
func (t *Topic) GetSubscribers() map[*Client]QoS {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to avoid race conditions
	subscribers := make(map[*Client]QoS)
	for client, qos := range t.subscribers {
		subscribers[client] = qos
	}
	return subscribers
}

// Name returns the topic name
func (t *Topic) Name() string {
	return t.name
}

// IsEmpty returns true if the topic has no subscribers
func (t *Topic) IsEmpty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.subscribers) == 0
}
