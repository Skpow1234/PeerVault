package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/google/uuid"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	URL           string                 `json:"url"`
	Method        string                 `json:"method"` // GET, POST, PUT, DELETE
	Headers       map[string]string      `json:"headers"`
	Events        []string               `json:"events"` // file_uploaded, file_deleted, etc.
	Secret        string                 `json:"secret"`
	RetryCount    int                    `json:"retry_count"`
	Timeout       time.Duration          `json:"timeout"`
	IsActive      bool                   `json:"is_active"`
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
	SuccessCount  int                    `json:"success_count"`
	FailureCount  int                    `json:"failure_count"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CreatedBy     string                 `json:"created_by"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID        string                 `json:"id"`
	WebhookID string                 `json:"webhook_id"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	Status    string                 `json:"status"` // pending, sent, failed, retrying
	Attempts  int                    `json:"attempts"`
	Response  *WebhookResponse       `json:"response,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	SentAt    *time.Time             `json:"sent_at,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// WebhookResponse represents a webhook response
type WebhookResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   time.Duration     `json:"duration"`
	Error      string            `json:"error,omitempty"`
}

// WebhookManager manages webhooks
type WebhookManager struct {
	mu        sync.RWMutex
	client    *client.Client
	configDir string
	webhooks  map[string]*Webhook
	events    map[string]*WebhookEvent
	stats     *WebhookStats
}

// WebhookStats represents webhook statistics
type WebhookStats struct {
	TotalWebhooks    int       `json:"total_webhooks"`
	ActiveWebhooks   int       `json:"active_webhooks"`
	TotalEvents      int       `json:"total_events"`
	SuccessfulEvents int       `json:"successful_events"`
	FailedEvents     int       `json:"failed_events"`
	PendingEvents    int       `json:"pending_events"`
	LastUpdated      time.Time `json:"last_updated"`
}

// NewWebhookManager creates a new webhook manager
func NewWebhookManager(client *client.Client, configDir string) *WebhookManager {
	wm := &WebhookManager{
		client:    client,
		configDir: configDir,
		webhooks:  make(map[string]*Webhook),
		events:    make(map[string]*WebhookEvent),
		stats:     &WebhookStats{},
	}
	_ = wm.loadWebhooks() // Ignore error for initialization
	_ = wm.loadEvents()   // Ignore error for initialization
	_ = wm.loadStats()    // Ignore error for initialization
	return wm
}

// CreateWebhook creates a new webhook
func (wm *WebhookManager) CreateWebhook(ctx context.Context, name, description, url, method, createdBy string, events []string, headers map[string]string, retryCount int, timeout time.Duration) (*Webhook, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	webhook := &Webhook{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		URL:          url,
		Method:       method,
		Headers:      headers,
		Events:       events,
		Secret:       fmt.Sprintf("wh_%s", uuid.New().String()[:16]),
		RetryCount:   retryCount,
		Timeout:      timeout,
		IsActive:     true,
		SuccessCount: 0,
		FailureCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    createdBy,
		Metadata:     make(map[string]interface{}),
	}

	wm.webhooks[webhook.ID] = webhook

	// Simulate API call - store webhook data as JSON
	webhookData, err := json.Marshal(webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webhook: %v", err)
	}

	tempFilePath := filepath.Join(wm.configDir, fmt.Sprintf("webhooks/%s.json", webhook.ID))
	if err := os.WriteFile(tempFilePath, webhookData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write webhook data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = wm.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store webhook: %v", err)
	}

	wm.stats.TotalWebhooks++
	wm.stats.ActiveWebhooks++
	_ = wm.saveStats()
	_ = wm.saveWebhooks()
	return webhook, nil
}

// ListWebhooks returns all webhooks
func (wm *WebhookManager) ListWebhooks(ctx context.Context) ([]*Webhook, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	webhooks := make([]*Webhook, 0, len(wm.webhooks))
	for _, webhook := range wm.webhooks {
		webhooks = append(webhooks, webhook)
	}
	return webhooks, nil
}

// GetWebhook returns a webhook by ID
func (wm *WebhookManager) GetWebhook(ctx context.Context, webhookID string) (*Webhook, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return nil, fmt.Errorf("webhook not found: %s", webhookID)
	}
	return webhook, nil
}

// TriggerWebhook triggers a webhook for an event
func (wm *WebhookManager) TriggerWebhook(ctx context.Context, webhookID, eventType string, payload map[string]interface{}) (*WebhookEvent, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return nil, fmt.Errorf("webhook not found: %s", webhookID)
	}

	if !webhook.IsActive {
		return nil, fmt.Errorf("webhook is not active: %s", webhookID)
	}

	// Check if webhook subscribes to this event
	eventAllowed := false
	for _, event := range webhook.Events {
		if event == eventType || event == "*" {
			eventAllowed = true
			break
		}
	}

	if !eventAllowed {
		return nil, fmt.Errorf("webhook does not subscribe to event: %s", eventType)
	}

	event := &WebhookEvent{
		ID:        uuid.New().String(),
		WebhookID: webhookID,
		EventType: eventType,
		Payload:   payload,
		Status:    "pending",
		Attempts:  0,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	wm.events[event.ID] = event

	// Simulate webhook call
	go wm.sendWebhook(webhook, event)

	wm.stats.TotalEvents++
	wm.stats.PendingEvents++
	_ = wm.saveStats()
	_ = wm.saveEvents()
	return event, nil
}

// sendWebhook simulates sending a webhook
func (wm *WebhookManager) sendWebhook(webhook *Webhook, event *WebhookEvent) {
	// Simulate HTTP request
	time.Sleep(100 * time.Millisecond) // Simulate network delay

	// Simulate response
	response := &WebhookResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:     `{"status": "success"}`,
		Duration: 100 * time.Millisecond,
	}

	// Update event
	wm.mu.Lock()
	event.Status = "sent"
	event.Response = response
	now := time.Now()
	event.SentAt = &now
	webhook.LastTriggered = &now
	webhook.SuccessCount++
	wm.stats.PendingEvents--
	wm.stats.SuccessfulEvents++
	wm.mu.Unlock()

	_ = wm.saveEvents()
	_ = wm.saveWebhooks()
	_ = wm.saveStats()
}

// UpdateWebhookStatus updates webhook status
func (wm *WebhookManager) UpdateWebhookStatus(ctx context.Context, webhookID string, isActive bool) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	webhook.IsActive = isActive
	webhook.UpdatedAt = time.Now()

	if isActive {
		wm.stats.ActiveWebhooks++
	} else {
		wm.stats.ActiveWebhooks--
	}

	_ = wm.saveWebhooks()
	_ = wm.saveStats()
	return nil
}

// ListWebhookEvents returns all webhook events
func (wm *WebhookManager) ListWebhookEvents(ctx context.Context) ([]*WebhookEvent, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	events := make([]*WebhookEvent, 0, len(wm.events))
	for _, event := range wm.events {
		events = append(events, event)
	}
	return events, nil
}

// GetWebhookStats returns webhook statistics
func (wm *WebhookManager) GetWebhookStats(ctx context.Context) (*WebhookStats, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// Update stats
	wm.stats.LastUpdated = time.Now()
	return wm.stats, nil
}

// File operations
func (wm *WebhookManager) loadWebhooks() error {
	webhooksFile := filepath.Join(wm.configDir, "webhooks.json")
	if _, err := os.Stat(webhooksFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(webhooksFile)
	if err != nil {
		return fmt.Errorf("failed to read webhooks file: %w", err)
	}

	var webhooks []*Webhook
	if err := json.Unmarshal(data, &webhooks); err != nil {
		return fmt.Errorf("failed to unmarshal webhooks: %w", err)
	}

	for _, webhook := range webhooks {
		wm.webhooks[webhook.ID] = webhook
		if webhook.IsActive {
			wm.stats.ActiveWebhooks++
		}
	}
	return nil
}

func (wm *WebhookManager) saveWebhooks() error {
	webhooksFile := filepath.Join(wm.configDir, "webhooks.json")

	var webhooks []*Webhook
	for _, webhook := range wm.webhooks {
		webhooks = append(webhooks, webhook)
	}

	data, err := json.MarshalIndent(webhooks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal webhooks: %w", err)
	}

	return os.WriteFile(webhooksFile, data, 0644)
}

func (wm *WebhookManager) loadEvents() error {
	eventsFile := filepath.Join(wm.configDir, "webhook_events.json")
	if _, err := os.Stat(eventsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(eventsFile)
	if err != nil {
		return fmt.Errorf("failed to read events file: %w", err)
	}

	var events []*WebhookEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("failed to unmarshal events: %w", err)
	}

	for _, event := range events {
		wm.events[event.ID] = event
		switch event.Status {
		case "pending":
			wm.stats.PendingEvents++
		case "sent":
			wm.stats.SuccessfulEvents++
		case "failed":
			wm.stats.FailedEvents++
		}
	}
	return nil
}

func (wm *WebhookManager) saveEvents() error {
	eventsFile := filepath.Join(wm.configDir, "webhook_events.json")

	var events []*WebhookEvent
	for _, event := range wm.events {
		events = append(events, event)
	}

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	return os.WriteFile(eventsFile, data, 0644)
}

func (wm *WebhookManager) loadStats() error {
	statsFile := filepath.Join(wm.configDir, "webhook_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats WebhookStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	wm.stats = &stats
	return nil
}

func (wm *WebhookManager) saveStats() error {
	statsFile := filepath.Join(wm.configDir, "webhook_stats.json")

	data, err := json.MarshalIndent(wm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
