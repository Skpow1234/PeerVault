package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

// TestHelper functions
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{}))
}

func createTestHub() *Hub {
	return NewHub(createTestLogger())
}

// Test WebSocket Handler
func TestNewHandler(t *testing.T) {
	hub := createTestHub()
	logger := createTestLogger()
	handler := NewHandler(hub, logger)

	if handler.hub != hub {
		t.Error("Handler hub not set correctly")
	}
	if handler.logger != logger {
		t.Error("Handler logger not set correctly")
	}
}

func TestGenerateClientID(t *testing.T) {
	id1 := generateClientID()
	id2 := generateClientID()

	if id1 == "" {
		t.Error("Client ID should not be empty")
	}
	if id2 == "" {
		t.Error("Client ID should not be empty")
	}
	if id1 == id2 {
		t.Error("Client IDs should be unique")
	}
	if len(id1) != 32 {
		t.Errorf("Expected client ID length of 32, got %d", len(id1))
	}
}

// Test GraphQL Subscription Handler
func TestNewGraphQLSubscriptionHandler(t *testing.T) {
	hub := createTestHub()
	logger := createTestLogger()
	handler := NewGraphQLSubscriptionHandler(hub, logger)

	if handler.hub != hub {
		t.Error("GraphQL handler hub not set correctly")
	}
	if handler.logger != logger {
		t.Error("GraphQL handler logger not set correctly")
	}
}

func TestGraphQLSubscriptionMessage(t *testing.T) {
	message := GraphQLSubscriptionMessage{
		ID:   "test-id",
		Type: "start",
		Payload: map[string]interface{}{
			"query": "subscription { test }",
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal GraphQL subscription message: %v", err)
	}

	var decoded GraphQLSubscriptionMessage
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal GraphQL subscription message: %v", err)
	}

	if decoded.ID != message.ID {
		t.Errorf("Expected ID %s, got %s", message.ID, decoded.ID)
	}
	if decoded.Type != message.Type {
		t.Errorf("Expected Type %s, got %s", message.Type, decoded.Type)
	}
}

// Test Hub functionality
func TestNewHub(t *testing.T) {
	logger := createTestLogger()
	hub := NewHub(logger)

	if hub.clients == nil {
		t.Error("Hub clients map not initialized")
	}
	if hub.broadcast == nil {
		t.Error("Hub broadcast channel not initialized")
	}
	if hub.register == nil {
		t.Error("Hub register channel not initialized")
	}
	if hub.unregister == nil {
		t.Error("Hub unregister channel not initialized")
	}
	if hub.subscriptions == nil {
		t.Error("Hub subscriptions map not initialized")
	}
	if hub.logger != logger {
		t.Error("Hub logger not set correctly")
	}
}

func TestHubSubscribeUnsubscribe(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-client")

	// Test subscribe
	hub.Subscribe(client, "test-topic")

	if len(hub.subscriptions) != 1 {
		t.Errorf("Expected 1 subscription topic, got %d", len(hub.subscriptions))
	}
	if len(hub.subscriptions["test-topic"]) != 1 {
		t.Errorf("Expected 1 client subscribed to topic, got %d", len(hub.subscriptions["test-topic"]))
	}
	if !client.subscriptions["test-topic"] {
		t.Error("Client should be subscribed to topic")
	}

	// Test unsubscribe
	hub.Unsubscribe(client, "test-topic")

	if len(hub.subscriptions) != 0 {
		t.Errorf("Expected 0 subscription topics after unsubscribe, got %d", len(hub.subscriptions))
	}
	if client.subscriptions["test-topic"] {
		t.Error("Client should not be subscribed to topic after unsubscribe")
	}
}

func TestHubGetClientCount(t *testing.T) {
	hub := createTestHub()

	if count := hub.GetClientCount(); count != 0 {
		t.Errorf("Expected client count 0, got %d", count)
	}

	// Simulate adding a client
	hub.mu.Lock()
	hub.clients[&Client{}] = true
	hub.mu.Unlock()

	if count := hub.GetClientCount(); count != 1 {
		t.Errorf("Expected client count 1, got %d", count)
	}
}

func TestHubGetSubscriptionCount(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-client")

	// Test with no subscriptions
	if count := hub.GetSubscriptionCount("nonexistent-topic"); count != 0 {
		t.Errorf("Expected subscription count 0 for nonexistent topic, got %d", count)
	}

	// Test with subscription
	hub.Subscribe(client, "test-topic")

	if count := hub.GetSubscriptionCount("test-topic"); count != 1 {
		t.Errorf("Expected subscription count 1, got %d", count)
	}
}

func TestHubGetTopics(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-client")

	// Test with no topics
	topics := hub.GetTopics()
	if len(topics) != 0 {
		t.Errorf("Expected 0 topics, got %d", len(topics))
	}

	// Test with topics
	hub.Subscribe(client, "topic1")
	hub.Subscribe(client, "topic2")

	topics = hub.GetTopics()
	if len(topics) != 2 {
		t.Errorf("Expected 2 topics, got %d", len(topics))
	}

	// Check if both topics are present
	topicMap := make(map[string]bool)
	for _, topic := range topics {
		topicMap[topic] = true
	}

	if !topicMap["topic1"] || !topicMap["topic2"] {
		t.Errorf("Expected topics topic1 and topic2, got %v", topics)
	}
}

// Test Client functionality
func TestNewClient(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-id")

	if client.id != "test-id" {
		t.Errorf("Expected client ID 'test-id', got '%s'", client.id)
	}
	if client.hub != hub {
		t.Error("Client hub not set correctly")
	}
	if client.send == nil {
		t.Error("Client send channel not initialized")
	}
	if client.subscriptions == nil {
		t.Error("Client subscriptions map not initialized")
	}
	if client.ctx == nil {
		t.Error("Client context not set")
	}
	if client.cancel == nil {
		t.Error("Client cancel function not set")
	}
}

func TestClientHandleMessageSubscribe(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-client")

	message := &Message{
		Type: "subscribe",
		Data: "test-topic",
	}

	client.handleMessage(message)

	if !client.subscriptions["test-topic"] {
		t.Error("Client should be subscribed to test-topic")
	}
}

func TestClientHandleMessageUnsubscribe(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-client")

	// First subscribe
	hub.Subscribe(client, "test-topic")

	// Then unsubscribe via message
	message := &Message{
		Type: "unsubscribe",
		Data: "test-topic",
	}

	client.handleMessage(message)

	if client.subscriptions["test-topic"] {
		t.Error("Client should not be subscribed to test-topic after unsubscribe")
	}
}

func TestClientHandleMessagePing(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-client")

	// Mock send channel to capture pong response
	pongReceived := make(chan bool, 1)
	client.send = make(chan []byte, 1)

	go func() {
		defer func() {
			// Ensure we always send a result to prevent test hanging
			select {
			case pongReceived <- false:
			default:
			}
		}()

		select {
		case msgBytes := <-client.send:
			var msg Message
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				pongReceived <- false
			} else if msg.Type == "pong" {
				pongReceived <- true
			} else {
				pongReceived <- false
			}
		case <-time.After(5 * time.Second): // Increased timeout for CI
			pongReceived <- false
		}
	}()

	message := &Message{
		Type: "ping",
	}

	client.handleMessage(message)

	select {
	case received := <-pongReceived:
		if !received {
			t.Error("Expected pong response to ping message")
		}
	case <-time.After(10 * time.Second): // Increased timeout for CI
		t.Error("Timeout waiting for pong response")
	}
}

func TestClientClose(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-client")

	// Test that context is cancelled
	select {
	case <-client.ctx.Done():
		t.Error("Client context should not be cancelled before Close()")
	default:
		// Expected - context not cancelled yet
	}

	// Test that close works with nil connection (should not panic)
	client.Close()

	// Test that context is cancelled after Close()
	select {
	case <-client.ctx.Done():
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("Client context should be cancelled after Close()")
	}
}

// Test Subscription Manager
func TestNewSubscriptionManager(t *testing.T) {
	hub := createTestHub()
	logger := createTestLogger()
	sm := NewSubscriptionManager(hub, logger)

	if sm.hub != hub {
		t.Error("Subscription manager hub not set correctly")
	}
	if sm.logger != logger {
		t.Error("Subscription manager logger not set correctly")
	}
	if sm.subscriptions == nil {
		t.Error("Subscription manager subscriptions map not initialized")
	}
}

func TestSubscriptionManagerCreateSubscription(t *testing.T) {
	hub := createTestHub()
	logger := createTestLogger()
	sm := NewSubscriptionManager(hub, logger)
	client := NewClient(nil, hub, "test-client")

	subscription, err := sm.CreateSubscription(client, "subscription { fileUploaded }", nil)
	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	if subscription.ID == "" {
		t.Error("Subscription ID should not be empty")
	}
	if subscription.Query != "subscription { fileUploaded }" {
		t.Errorf("Expected query 'subscription { fileUploaded }', got '%s'", subscription.Query)
	}
	if subscription.Client != client {
		t.Error("Subscription client not set correctly")
	}
	if len(subscription.Topics) == 0 {
		t.Error("Subscription should have extracted topics")
	}
}

func TestSubscriptionManagerRemoveSubscription(t *testing.T) {
	hub := createTestHub()
	logger := createTestLogger()
	sm := NewSubscriptionManager(hub, logger)
	client := NewClient(nil, hub, "test-client")

	subscription, _ := sm.CreateSubscription(client, "subscription { fileUploaded }", nil)

	// Verify subscription exists
	if _, exists := sm.GetSubscription(subscription.ID); !exists {
		t.Error("Subscription should exist before removal")
	}

	sm.RemoveSubscription(subscription.ID)

	// Verify subscription is removed
	if _, exists := sm.GetSubscription(subscription.ID); exists {
		t.Error("Subscription should not exist after removal")
	}
}

func TestSubscriptionManagerGetSubscriptionsByTopic(t *testing.T) {
	hub := createTestHub()
	logger := createTestLogger()
	sm := NewSubscriptionManager(hub, logger)
	client := NewClient(nil, hub, "test-client")

	// Create subscription with fileUploaded topic
	subscription1, _ := sm.CreateSubscription(client, "subscription { fileUploaded }", nil)

	// Create another subscription with different topic
	subscription2, _ := sm.CreateSubscription(client, "subscription { peerConnected }", nil)

	// Get subscriptions for file.uploaded topic
	subscriptions := sm.GetSubscriptionsByTopic("file.uploaded")

	if len(subscriptions) != 1 {
		t.Errorf("Expected 1 subscription for file.uploaded topic, got %d", len(subscriptions))
	}
	if subscriptions[0].ID != subscription1.ID {
		t.Errorf("Expected subscription ID %s, got %s", subscription1.ID, subscriptions[0].ID)
	}

	// Get subscriptions for peer.connected topic
	subscriptions = sm.GetSubscriptionsByTopic("peer.connected")

	if len(subscriptions) != 1 {
		t.Errorf("Expected 1 subscription for peer.connected topic, got %d", len(subscriptions))
	}
	if subscriptions[0].ID != subscription2.ID {
		t.Errorf("Expected subscription ID %s, got %s", subscription2.ID, subscriptions[0].ID)
	}
}

// Test utility functions
func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "goodbye", false},
		{"test", "test", true},
		{"", "", true},
		{"hello", "", true},
		{"", "test", false},
	}

	for _, test := range tests {
		result := contains(test.s, test.substr)
		if result != test.expected {
			t.Errorf("contains(%q, %q) = %v, expected %v", test.s, test.substr, result, test.expected)
		}
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected int
	}{
		{"hello world", "world", 6},
		{"hello world", "goodbye", -1},
		{"test", "test", 0},
		{"", "", 0},
		{"hello", "", 0},
		{"", "test", -1},
		{"aaa", "a", 0},
	}

	for _, test := range tests {
		result := indexOf(test.s, test.substr)
		if result != test.expected {
			t.Errorf("indexOf(%q, %q) = %d, expected %d", test.s, test.substr, result, test.expected)
		}
	}
}

func TestExtractTopicsFromQuery(t *testing.T) {
	hub := createTestHub()
	logger := createTestLogger()
	sm := NewSubscriptionManager(hub, logger)

	tests := []struct {
		query    string
		expected []string
	}{
		{"subscription { fileUploaded }", []string{"file.uploaded"}},
		{"subscription { fileDeleted peerConnected }", []string{"file.deleted", "peer.connected"}},
		{"subscription { systemMetricsUpdated }", []string{"system.metrics_updated"}},
		{"subscription { performanceAlert }", []string{"system.performance_alert"}},
		{"query { test }", []string{}},
	}

	for _, test := range tests {
		topics := sm.extractTopicsFromQuery(test.query)
		if len(topics) != len(test.expected) {
			t.Errorf("Expected %d topics for query %q, got %d", len(test.expected), test.query, len(topics))
			continue
		}

		for i, expectedTopic := range test.expected {
			if i >= len(topics) || topics[i] != expectedTopic {
				t.Errorf("Expected topic %q at index %d for query %q, got %v", expectedTopic, i, test.query, topics)
			}
		}
	}
}

// Test Message structure
func TestMessageJSON(t *testing.T) {
	message := Message{
		Type:      "test",
		Topic:     "test-topic",
		Data:      map[string]string{"key": "value"},
		Timestamp: time.Now(),
		ClientID:  "test-client",
	}

	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var decoded Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if decoded.Type != message.Type {
		t.Errorf("Expected Type %s, got %s", message.Type, decoded.Type)
	}
	if decoded.Topic != message.Topic {
		t.Errorf("Expected Topic %s, got %s", message.Topic, decoded.Topic)
	}
	if decoded.ClientID != message.ClientID {
		t.Errorf("Expected ClientID %s, got %s", message.ClientID, decoded.ClientID)
	}
}

// Integration test for hub operations
func TestHubIntegration(t *testing.T) {
	hub := createTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start hub in background
	go hub.Run(ctx)

	// Create test clients
	client1 := NewClient(nil, hub, "client1")
	client2 := NewClient(nil, hub, "client2")

	// Register clients
	hub.register <- client1
	hub.register <- client2

	// Wait for registration with exponential backoff
	for i := 0; i < 10; i++ {
		if hub.GetClientCount() == 2 {
			break
		}
		time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
	}

	if count := hub.GetClientCount(); count != 2 {
		t.Errorf("Expected 2 clients, got %d", count)
	}

	// Test topic subscription
	hub.Subscribe(client1, "test-topic")
	hub.Subscribe(client2, "test-topic")

	if count := hub.GetSubscriptionCount("test-topic"); count != 2 {
		t.Errorf("Expected 2 subscriptions to test-topic, got %d", count)
	}

	// Test unregistration
	hub.unregister <- client1

	// Wait for unregistration with exponential backoff
	for i := 0; i < 10; i++ {
		if hub.GetClientCount() == 1 {
			break
		}
		time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
	}

	if count := hub.GetClientCount(); count != 1 {
		t.Errorf("Expected 1 client after unregistration, got %d", count)
	}

	if count := hub.GetSubscriptionCount("test-topic"); count != 1 {
		t.Errorf("Expected 1 subscription to test-topic after unregistration, got %d", count)
	}
}

// Test GraphQL subscription handling
func TestHandleGraphQLSubscription(t *testing.T) {
	hub := createTestHub()
	client := NewClient(nil, hub, "test-client")

	// Test connection_init - should create acknowledgment message
	message := &GraphQLSubscriptionMessage{
		Type: "connection_init",
	}

	// Mock send channel to capture acknowledgment
	sendChan := make(chan []byte, 1)
	client.send = sendChan

	client.HandleGraphQLSubscription(message)

	// Wait for acknowledgment message with increased timeout for CI
	select {
	case msgBytes := <-sendChan:
		var msg GraphQLSubscriptionMessage
		err := json.Unmarshal(msgBytes, &msg)
		if err != nil {
			t.Fatalf("Failed to unmarshal acknowledgment message: %v", err)
		}
		if msg.Type != "connection_ack" {
			t.Errorf("Expected message type 'connection_ack', got '%s'", msg.Type)
		}
	case <-time.After(5 * time.Second): // Increased timeout for CI
		t.Error("Timeout waiting for connection acknowledgment")
	}

	// Test connection_terminate
	message = &GraphQLSubscriptionMessage{
		Type: "connection_terminate",
	}

	// Test that context gets cancelled
	select {
	case <-client.ctx.Done():
		t.Error("Context should not be cancelled before connection_terminate")
	default:
		// Expected
	}

	client.HandleGraphQLSubscription(message)

	// Verify context is cancelled after connection_terminate
	select {
	case <-client.ctx.Done():
		// Expected
	case <-time.After(5 * time.Second): // Increased timeout for CI
		t.Error("Context should be cancelled after connection_terminate")
	}
}

func TestBroadcastGraphQLData(t *testing.T) {
	hub := createTestHub()

	// Create mock client with send channel
	sendChan := make(chan []byte, 1)
	client := NewClient(nil, hub, "test-client")
	client.send = sendChan

	// Add client to hub manually
	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	// Broadcast GraphQL data
	testData := map[string]interface{}{"test": "data"}
	hub.BroadcastGraphQLData("test-subscription", testData)

	// Wait for message with increased timeout for CI
	select {
	case msgBytes := <-sendChan:
		var msg GraphQLSubscriptionMessage
		err := json.Unmarshal(msgBytes, &msg)
		if err != nil {
			t.Fatalf("Failed to unmarshal broadcast message: %v", err)
		}

		if msg.Type != "data" {
			t.Errorf("Expected message type 'data', got '%s'", msg.Type)
		}
		if msg.ID != "test-subscription" {
			t.Errorf("Expected subscription ID 'test-subscription', got '%s'", msg.ID)
		}
	case <-time.After(5 * time.Second): // Increased timeout for CI
		t.Error("Timeout waiting for broadcast message")
	}
}

// Test Hub Run method (basic smoke test)
func TestHubRun(t *testing.T) {
	hub := createTestHub()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond) // Increased timeout for CI
	defer cancel()                                                                 // Ensure cancel is always called to prevent context leak

	// Start hub
	done := make(chan bool, 1) // Buffered to prevent goroutine leak
	go func() {
		defer func() {
			// Ensure done is always signaled
			select {
			case done <- true:
			default:
			}
		}()
		hub.Run(ctx)
	}()

	// Wait for context timeout or hub to finish
	select {
	case <-done:
		// Hub finished normally
	case <-ctx.Done():
		// Context timed out, wait for hub to finish with additional timeout
		select {
		case <-done:
			// Hub finished after context cancellation
		case <-time.After(1 * time.Second):
			t.Error("Hub did not finish gracefully after context cancellation")
		}
	}
}

// Test helper function for creating HTTP requests
func createTestHTTPRequest(method, url string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	return req
}

// Test Upgrader configuration
func TestUpgraderConfiguration(t *testing.T) {
	if Upgrader.ReadBufferSize != 1024 {
		t.Errorf("Expected ReadBufferSize 1024, got %d", Upgrader.ReadBufferSize)
	}
	if Upgrader.WriteBufferSize != 1024 {
		t.Errorf("Expected WriteBufferSize 1024, got %d", Upgrader.WriteBufferSize)
	}
	if Upgrader.CheckOrigin == nil {
		t.Error("CheckOrigin function should be set")
	}

	// Test CheckOrigin function
	req := createTestHTTPRequest("GET", "http://example.com")
	if !Upgrader.CheckOrigin(req) {
		t.Error("CheckOrigin should return true for any origin")
	}
}
