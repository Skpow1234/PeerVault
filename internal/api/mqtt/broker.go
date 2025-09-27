package mqtt

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/app/fileserver"
)

// Broker represents the MQTT broker
type Broker struct {
	fileserver *fileserver.Server
	config     *BrokerConfig
	logger     *slog.Logger

	// Client management
	clients   map[string]*Client
	clientsMu sync.RWMutex

	// Topic management
	topics   map[string]*Topic
	topicsMu sync.RWMutex

	// Message store for persistence
	messageStore *MessageStore

	// Statistics
	stats   *BrokerStats
	statsMu sync.RWMutex

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// BrokerConfig holds the configuration for the MQTT broker
type BrokerConfig struct {
	Port            int
	Host            string
	EnableWebSocket bool
	WebSocketPort   int
	KeepAlive       time.Duration
	MaxConnections  int
	MaxMessageSize  int
	RetainEnabled   bool
	WillEnabled     bool
	CleanSession    bool
}

// BrokerStats holds broker statistics
type BrokerStats struct {
	StartTime         time.Time
	TotalConnections  int
	ActiveConnections int
	TotalMessages     int
	TotalTopics       int
	BytesReceived     int64
	BytesSent         int64
}

// NewBroker creates a new MQTT broker
func NewBroker(fileserver *fileserver.Server, config *BrokerConfig, logger *slog.Logger) *Broker {
	ctx, cancel := context.WithCancel(context.Background())

	broker := &Broker{
		fileserver:   fileserver,
		config:       config,
		logger:       logger,
		clients:      make(map[string]*Client),
		topics:       make(map[string]*Topic),
		messageStore: NewMessageStore(),
		stats: &BrokerStats{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Start background tasks
	go broker.startBackgroundTasks()

	return broker
}

// ServeTCP starts the TCP MQTT server
func (b *Broker) ServeTCP(ctx context.Context, listener net.Listener) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Set read timeout
		if err := listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
			b.logger.Error("Failed to set deadline", "error", err)
			continue
		}

		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return err
		}

		// Handle connection in goroutine
		go b.handleConnection(conn)
	}
}

// ServeWebSocket starts the WebSocket MQTT server
func (b *Broker) ServeWebSocket(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mqtt", b.handleWebSocketUpgrade)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			b.logger.Error("Failed to shutdown WebSocket server", "error", err)
		}
	}()

	return server.ListenAndServe()
}

// handleConnection handles a new TCP connection
func (b *Broker) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			b.logger.Error("Failed to close connection", "error", err)
		}
	}()

	// Create client
	client := NewClient(conn, b, b.logger)

	// Handle client session
	if err := client.Handle(); err != nil {
		b.logger.Error("Client session error", "error", err, "clientId", client.ID)
	}

	// Remove client from broker
	b.removeClient(client.ID)
}

// handleWebSocketUpgrade handles WebSocket upgrade for MQTT over WebSocket
func (b *Broker) handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket upgrade for MQTT over WebSocket
	// This would involve upgrading the HTTP connection to WebSocket
	// and then handling MQTT packets over the WebSocket connection
	b.logger.Info("WebSocket MQTT connection attempt", "remoteAddr", r.RemoteAddr)
	http.Error(w, "MQTT over WebSocket not yet implemented", http.StatusNotImplemented)
}

// addClient adds a client to the broker
func (b *Broker) addClient(client *Client) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	b.clients[client.ID] = client
	b.updateStats(func(stats *BrokerStats) {
		stats.TotalConnections++
		stats.ActiveConnections++
	})

	b.logger.Info("Client connected",
		"clientId", client.ID,
		"remoteAddr", client.RemoteAddr(),
		"activeConnections", b.getActiveConnections(),
	)
}

// removeClient removes a client from the broker
func (b *Broker) removeClient(clientID string) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	if client, exists := b.clients[clientID]; exists {
		// Unsubscribe from all topics
		b.unsubscribeClientFromAllTopics(client)

		// Send will message if configured
		if client.WillMessage != nil {
			if err := b.publishMessage(client.WillMessage); err != nil {
				b.logger.Error("Failed to publish will message", "error", err, "clientId", clientID)
			}
		}

		delete(b.clients, clientID)
		b.updateStats(func(stats *BrokerStats) {
			stats.ActiveConnections--
		})

		b.logger.Info("Client disconnected",
			"clientId", clientID,
			"activeConnections", b.getActiveConnections(),
		)
	}
}

// getActiveConnections returns the number of active connections
func (b *Broker) getActiveConnections() int {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()
	return len(b.clients)
}

// subscribeClient subscribes a client to a topic
func (b *Broker) subscribeClient(client *Client, topicName string, qos QoS) error {
	b.topicsMu.Lock()
	defer b.topicsMu.Unlock()

	// Get or create topic
	topic, exists := b.topics[topicName]
	if !exists {
		topic = NewTopic(topicName, b.logger)
		b.topics[topicName] = topic
		b.updateStats(func(stats *BrokerStats) {
			stats.TotalTopics++
		})
	}

	// Subscribe client to topic
	topic.Subscribe(client, qos)

	b.logger.Info("Client subscribed to topic",
		"clientId", client.ID,
		"topic", topicName,
		"qos", qos,
	)

	return nil
}

// unsubscribeClient unsubscribes a client from a topic
func (b *Broker) unsubscribeClient(client *Client, topicName string) error {
	b.topicsMu.Lock()
	defer b.topicsMu.Unlock()

	if topic, exists := b.topics[topicName]; exists {
		topic.Unsubscribe(client)
		b.logger.Info("Client unsubscribed from topic",
			"clientId", client.ID,
			"topic", topicName,
		)
	}

	return nil
}

// unsubscribeClientFromAllTopics unsubscribes a client from all topics
func (b *Broker) unsubscribeClientFromAllTopics(client *Client) {
	b.topicsMu.Lock()
	defer b.topicsMu.Unlock()

	for _, topic := range b.topics {
		topic.Unsubscribe(client)
	}
}

// publishMessage publishes a message to a topic
func (b *Broker) publishMessage(message *Message) error {
	b.topicsMu.RLock()
	defer b.topicsMu.RUnlock()

	// Find matching topics (support wildcards)
	matchingTopics := b.findMatchingTopics(message.Topic)

	if len(matchingTopics) == 0 {
		b.logger.Debug("No subscribers for topic", "topic", message.Topic)
		return nil
	}

	// Publish to all matching topics
	for _, topic := range matchingTopics {
		topic.Publish(message)
	}

	// Store message if retain is enabled
	if message.Retain && b.config.RetainEnabled {
		b.messageStore.StoreRetainedMessage(message.Topic, message)
	}

	// Update statistics
	b.updateStats(func(stats *BrokerStats) {
		stats.TotalMessages++
		stats.BytesSent += int64(len(message.Payload))
	})

	b.logger.Debug("Message published",
		"topic", message.Topic,
		"qos", message.QoS,
		"retain", message.Retain,
		"subscribers", len(matchingTopics),
	)

	return nil
}

// findMatchingTopics finds topics that match the given topic pattern
func (b *Broker) findMatchingTopics(topicPattern string) []*Topic {
	var matching []*Topic

	for topicName, topic := range b.topics {
		if b.topicMatches(topicPattern, topicName) {
			matching = append(matching, topic)
		}
	}

	return matching
}

// topicMatches checks if a topic pattern matches a topic name
func (b *Broker) topicMatches(pattern, topic string) bool {
	// Simple exact match for now
	// TODO: Implement wildcard matching (+ and #)
	return pattern == topic
}

// updateStats updates broker statistics
func (b *Broker) updateStats(updater func(*BrokerStats)) {
	b.statsMu.Lock()
	defer b.statsMu.Unlock()
	updater(b.stats)
}

// GetStats returns broker statistics
func (b *Broker) GetStats() *BrokerStats {
	b.statsMu.RLock()
	defer b.statsMu.RUnlock()

	// Return a copy
	stats := *b.stats
	return &stats
}

// startBackgroundTasks starts background maintenance tasks
func (b *Broker) startBackgroundTasks() {
	// Cleanup task
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.cleanup()
		case <-b.ctx.Done():
			return
		}
	}
}

// cleanup performs periodic cleanup tasks
func (b *Broker) cleanup() {
	// Remove empty topics
	b.topicsMu.Lock()
	for topicName, topic := range b.topics {
		if topic.SubscriberCount() == 0 {
			delete(b.topics, topicName)
			b.logger.Debug("Removed empty topic", "topic", topicName)
		}
	}
	b.topicsMu.Unlock()

	// Clean up expired retained messages
	b.messageStore.Cleanup()
}

// Shutdown gracefully shuts down the broker
func (b *Broker) Shutdown() {
	b.logger.Info("Shutting down MQTT broker...")
	b.cancel()

	// Close all client connections
	b.clientsMu.Lock()
	for _, client := range b.clients {
		client.Close()
	}
	b.clientsMu.Unlock()

	b.logger.Info("MQTT broker shutdown complete")
}
