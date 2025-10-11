package realtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/websocket"
)

// Manager manages real-time updates for the CLI
type Manager struct {
	wsClient    *websocket.Client
	connected   bool
	mu          sync.RWMutex
	subscribers map[string][]Subscriber
	stopChan    chan struct{}
}

// Subscriber represents a subscriber to real-time updates
type Subscriber struct {
	ID      string
	Handler func(data map[string]interface{})
}

// New creates a new real-time manager
func New(serverURL string) *Manager {
	return &Manager{
		wsClient:    websocket.New(serverURL),
		subscribers: make(map[string][]Subscriber),
		stopChan:    make(chan struct{}),
	}
}

// Connect connects to the real-time service
func (m *Manager) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.wsClient.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to real-time service: %w", err)
	}

	m.connected = true

	// Set up event handlers
	m.setupEventHandlers()

	// Start keep-alive
	go m.keepAlive()

	return nil
}

// Disconnect disconnects from the real-time service
func (m *Manager) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = false
	close(m.stopChan)

	return m.wsClient.Disconnect()
}

// IsConnected returns whether the manager is connected
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// Subscribe subscribes to real-time updates
func (m *Manager) Subscribe(eventType string, subscriber Subscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscribers[eventType] = append(m.subscribers[eventType], subscriber)
}

// Unsubscribe unsubscribes from real-time updates
func (m *Manager) Unsubscribe(eventType string, subscriberID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscribers := m.subscribers[eventType]
	for i, sub := range subscribers {
		if sub.ID == subscriberID {
			m.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
}

// setupEventHandlers sets up WebSocket event handlers
func (m *Manager) setupEventHandlers() {
	// File events
	m.wsClient.Subscribe("file_uploaded", m.handleFileEvent)
	m.wsClient.Subscribe("file_deleted", m.handleFileEvent)
	m.wsClient.Subscribe("file_updated", m.handleFileEvent)

	// Peer events
	m.wsClient.Subscribe("peer_connected", m.handlePeerEvent)
	m.wsClient.Subscribe("peer_disconnected", m.handlePeerEvent)
	m.wsClient.Subscribe("peer_updated", m.handlePeerEvent)

	// System events
	m.wsClient.Subscribe("health_changed", m.handleSystemEvent)
	m.wsClient.Subscribe("metrics_updated", m.handleSystemEvent)
	m.wsClient.Subscribe("status_changed", m.handleSystemEvent)

	// Backup events
	m.wsClient.Subscribe("backup_started", m.handleBackupEvent)
	m.wsClient.Subscribe("backup_completed", m.handleBackupEvent)
	m.wsClient.Subscribe("backup_failed", m.handleBackupEvent)

	// IoT events
	m.wsClient.Subscribe("device_connected", m.handleIoTEvent)
	m.wsClient.Subscribe("device_disconnected", m.handleIoTEvent)
	m.wsClient.Subscribe("sensor_data", m.handleIoTEvent)
}

// handleFileEvent handles file-related events
func (m *Manager) handleFileEvent(event websocket.Event) {
	m.mu.RLock()
	subscribers := m.subscribers["file"]
	m.mu.RUnlock()

	for _, subscriber := range subscribers {
		go subscriber.Handler(event.Data)
	}
}

// handlePeerEvent handles peer-related events
func (m *Manager) handlePeerEvent(event websocket.Event) {
	m.mu.RLock()
	subscribers := m.subscribers["peer"]
	m.mu.RUnlock()

	for _, subscriber := range subscribers {
		go subscriber.Handler(event.Data)
	}
}

// handleSystemEvent handles system-related events
func (m *Manager) handleSystemEvent(event websocket.Event) {
	m.mu.RLock()
	subscribers := m.subscribers["system"]
	m.mu.RUnlock()

	for _, subscriber := range subscribers {
		go subscriber.Handler(event.Data)
	}
}

// handleBackupEvent handles backup-related events
func (m *Manager) handleBackupEvent(event websocket.Event) {
	m.mu.RLock()
	subscribers := m.subscribers["backup"]
	m.mu.RUnlock()

	for _, subscriber := range subscribers {
		go subscriber.Handler(event.Data)
	}
}

// handleIoTEvent handles IoT-related events
func (m *Manager) handleIoTEvent(event websocket.Event) {
	m.mu.RLock()
	subscribers := m.subscribers["iot"]
	m.mu.RUnlock()

	for _, subscriber := range subscribers {
		go subscriber.Handler(event.Data)
	}
}

// keepAlive sends periodic ping messages
func (m *Manager) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if m.IsConnected() {
				err := m.wsClient.Ping()
				if err != nil {
					fmt.Printf("Keep-alive ping failed: %v\n", err)
				}
			}
		case <-m.stopChan:
			return
		}
	}
}

// SendEvent sends an event to the server
func (m *Manager) SendEvent(eventType string, data map[string]interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return fmt.Errorf("not connected to real-time service")
	}

	return m.wsClient.Send(eventType, data)
}

// GetConnectionStatus returns the connection status
func (m *Manager) GetConnectionStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"connected":   m.connected,
		"subscribers": len(m.subscribers),
		"event_types": len(m.subscribers),
	}
}

// Reconnect attempts to reconnect to the real-time service
func (m *Manager) Reconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		_ = m.wsClient.Disconnect()
	}

	m.connected = false
	m.stopChan = make(chan struct{})

	return m.Connect(ctx)
}
