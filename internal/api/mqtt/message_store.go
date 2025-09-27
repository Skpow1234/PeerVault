package mqtt

import (
	"log/slog"
	"sync"
	"time"
)

// MessageStore manages message persistence and retention
type MessageStore struct {
	// Retained messages by topic
	retainedMessages map[string]*Message
	retainedMu       sync.RWMutex

	// Message history for QoS 2
	qos2Messages map[uint16]*Message
	qos2Mu       sync.RWMutex

	// Logger
	logger *slog.Logger

	// Configuration
	maxRetainedMessages int
	retentionDuration   time.Duration
}

// NewMessageStore creates a new message store
func NewMessageStore() *MessageStore {
	return &MessageStore{
		retainedMessages:    make(map[string]*Message),
		qos2Messages:        make(map[uint16]*Message),
		logger:              slog.Default(),
		maxRetainedMessages: 1000,
		retentionDuration:   24 * time.Hour,
	}
}

// StoreRetainedMessage stores a retained message for a topic
func (ms *MessageStore) StoreRetainedMessage(topic string, message *Message) {
	ms.retainedMu.Lock()
	defer ms.retainedMu.Unlock()

	// Check if we need to clean up old messages
	if len(ms.retainedMessages) >= ms.maxRetainedMessages {
		ms.cleanupOldRetainedMessages()
	}

	// Store the message
	ms.retainedMessages[topic] = message

	ms.logger.Debug("Retained message stored",
		"topic", topic,
		"qos", message.QoS,
		"retain", message.Retain,
		"totalRetained", len(ms.retainedMessages),
	)
}

// GetRetainedMessage retrieves a retained message for a topic
func (ms *MessageStore) GetRetainedMessage(topic string) *Message {
	ms.retainedMu.RLock()
	defer ms.retainedMu.RUnlock()

	if message, exists := ms.retainedMessages[topic]; exists {
		return message
	}

	return nil
}

// GetRetainedMessagesForTopic returns all retained messages that match a topic pattern
func (ms *MessageStore) GetRetainedMessagesForTopic(topicPattern string) []*Message {
	ms.retainedMu.RLock()
	defer ms.retainedMu.RUnlock()

	var messages []*Message
	for topic, message := range ms.retainedMessages {
		if ms.topicMatches(topicPattern, topic) {
			messages = append(messages, message)
		}
	}

	return messages
}

// ClearRetainedMessage clears a retained message for a topic
func (ms *MessageStore) ClearRetainedMessage(topic string) {
	ms.retainedMu.Lock()
	defer ms.retainedMu.Unlock()

	delete(ms.retainedMessages, topic)

	ms.logger.Debug("Retained message cleared",
		"topic", topic,
		"totalRetained", len(ms.retainedMessages),
	)
}

// StoreQoS2Message stores a QoS 2 message for later processing
func (ms *MessageStore) StoreQoS2Message(packetID uint16, message *Message) {
	ms.qos2Mu.Lock()
	defer ms.qos2Mu.Unlock()

	ms.qos2Messages[packetID] = message

	ms.logger.Debug("QoS 2 message stored",
		"packetId", packetID,
		"topic", message.Topic,
		"totalQoS2", len(ms.qos2Messages),
	)
}

// GetQoS2Message retrieves a QoS 2 message
func (ms *MessageStore) GetQoS2Message(packetID uint16) *Message {
	ms.qos2Mu.RLock()
	defer ms.qos2Mu.RUnlock()

	if message, exists := ms.qos2Messages[packetID]; exists {
		return message
	}

	return nil
}

// RemoveQoS2Message removes a QoS 2 message
func (ms *MessageStore) RemoveQoS2Message(packetID uint16) {
	ms.qos2Mu.Lock()
	defer ms.qos2Mu.Unlock()

	delete(ms.qos2Messages, packetID)

	ms.logger.Debug("QoS 2 message removed",
		"packetId", packetID,
		"totalQoS2", len(ms.qos2Messages),
	)
}

// Cleanup performs periodic cleanup of expired messages
func (ms *MessageStore) Cleanup() {
	ms.cleanupOldRetainedMessages()
	ms.cleanupOldQoS2Messages()
}

// cleanupOldRetainedMessages removes old retained messages
func (ms *MessageStore) cleanupOldRetainedMessages() {
	// Simple cleanup - remove oldest messages if we exceed the limit
	// In a real implementation, you might want to track timestamps
	if len(ms.retainedMessages) > ms.maxRetainedMessages {
		// Remove some old messages (simple implementation)
		count := 0
		for topic := range ms.retainedMessages {
			if count >= 100 { // Remove 100 messages at a time
				break
			}
			delete(ms.retainedMessages, topic)
			count++
		}

		ms.logger.Info("Cleaned up old retained messages",
			"removed", count,
			"remaining", len(ms.retainedMessages),
		)
	}
}

// cleanupOldQoS2Messages removes old QoS 2 messages
func (ms *MessageStore) cleanupOldQoS2Messages() {
	// Simple cleanup - remove all QoS 2 messages older than retention duration
	// In a real implementation, you would track message timestamps
	if len(ms.qos2Messages) > 1000 {
		// Clear all QoS 2 messages if we have too many
		ms.qos2Messages = make(map[uint16]*Message)

		ms.logger.Info("Cleaned up old QoS 2 messages")
	}
}

// topicMatches checks if a topic pattern matches a topic name
func (ms *MessageStore) topicMatches(pattern, topic string) bool {
	// Simple exact match for now
	// TODO: Implement wildcard matching (+ and #)
	return pattern == topic
}

// GetStats returns message store statistics
func (ms *MessageStore) GetStats() map[string]interface{} {
	ms.retainedMu.RLock()
	retainedCount := len(ms.retainedMessages)
	ms.retainedMu.RUnlock()

	ms.qos2Mu.RLock()
	qos2Count := len(ms.qos2Messages)
	ms.qos2Mu.RUnlock()

	return map[string]interface{}{
		"retained_messages":  retainedCount,
		"qos2_messages":      qos2Count,
		"max_retained":       ms.maxRetainedMessages,
		"retention_duration": ms.retentionDuration.String(),
	}
}

// SetMaxRetainedMessages sets the maximum number of retained messages
func (ms *MessageStore) SetMaxRetainedMessages(max int) {
	ms.maxRetainedMessages = max
}

// SetRetentionDuration sets the retention duration for messages
func (ms *MessageStore) SetRetentionDuration(duration time.Duration) {
	ms.retentionDuration = duration
}
