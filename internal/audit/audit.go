package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditLevel represents the level of an audit event
type AuditLevel string

const (
	AuditLevelInfo     AuditLevel = "info"
	AuditLevelWarning  AuditLevel = "warning"
	AuditLevelError    AuditLevel = "error"
	AuditLevelCritical AuditLevel = "critical"
)

// AuditEventType represents the type of an audit event
type AuditEventType string

const (
	AuditEventTypeAuth       AuditEventType = "authentication"
	AuditEventTypeAccess     AuditEventType = "access"
	AuditEventTypeData       AuditEventType = "data"
	AuditEventTypeSystem     AuditEventType = "system"
	AuditEventTypeSecurity   AuditEventType = "security"
	AuditEventTypeCompliance AuditEventType = "compliance"
	AuditEventTypeAdmin      AuditEventType = "administrative"
)

// AuditEvent represents an audit event
type AuditEvent struct {
	ID        string                 `json:"id"`
	Type      AuditEventType         `json:"type"`
	Level     AuditLevel             `json:"level"`
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Result    string                 `json:"result,omitempty"` // success, failure, denied
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source,omitempty"`
	Category  string                 `json:"category,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
}

// AuditLogger manages audit logging
type AuditLogger struct {
	events        []AuditEvent
	mu            sync.RWMutex
	logFile       *os.File
	logPath       string
	buffer        []AuditEvent
	bufferMu      sync.Mutex
	batchSize     int
	flushInterval time.Duration
	stopChan      chan struct{}
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logPath string) (*AuditLogger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := &AuditLogger{
		events:        make([]AuditEvent, 0),
		logFile:       logFile,
		logPath:       logPath,
		buffer:        make([]AuditEvent, 0),
		batchSize:     100,
		flushInterval: 30 * time.Second,
		stopChan:      make(chan struct{}),
	}

	// Start background flush routine
	go logger.flushRoutine()

	return logger, nil
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) error {
	// Set default values
	if event.ID == "" {
		event.ID = fmt.Sprintf("audit_%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Level == "" {
		event.Level = AuditLevelInfo
	}

	// Add to buffer
	al.bufferMu.Lock()
	al.buffer = append(al.buffer, *event)
	al.bufferMu.Unlock()

	// Add to in-memory events
	al.mu.Lock()
	al.events = append(al.events, *event)
	al.mu.Unlock()

	// Flush if buffer is full
	if len(al.buffer) >= al.batchSize {
		go func() {
			_ = al.flush() // Ignore error in goroutine
		}()
	}

	return nil
}

// LogAuthEvent logs an authentication event
func (al *AuditLogger) LogAuthEvent(ctx context.Context, userID, action, result, ipAddress, userAgent string, details map[string]interface{}) error {
	event := &AuditEvent{
		Type:      AuditEventTypeAuth,
		Level:     AuditLevelInfo,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Action:    action,
		Result:    result,
		Message:   fmt.Sprintf("Authentication %s: %s", action, result),
		Details:   details,
		Category:  "authentication",
		Tags:      []string{"auth", action, result},
	}

	return al.LogEvent(ctx, event)
}

// LogAccessEvent logs an access event
func (al *AuditLogger) LogAccessEvent(ctx context.Context, userID, resource, action, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		Type:     AuditEventTypeAccess,
		Level:    AuditLevelInfo,
		UserID:   userID,
		Resource: resource,
		Action:   action,
		Result:   result,
		Message:  fmt.Sprintf("Access %s to %s: %s", action, resource, result),
		Details:  details,
		Category: "access_control",
		Tags:     []string{"access", action, result},
	}

	return al.LogEvent(ctx, event)
}

// LogDataEvent logs a data access event
func (al *AuditLogger) LogDataEvent(ctx context.Context, userID, resource, action, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		Type:     AuditEventTypeData,
		Level:    AuditLevelInfo,
		UserID:   userID,
		Resource: resource,
		Action:   action,
		Result:   result,
		Message:  fmt.Sprintf("Data %s on %s: %s", action, resource, result),
		Details:  details,
		Category: "data_access",
		Tags:     []string{"data", action, result},
	}

	return al.LogEvent(ctx, event)
}

// LogSecurityEvent logs a security event
func (al *AuditLogger) LogSecurityEvent(ctx context.Context, level AuditLevel, message string, details map[string]interface{}) error {
	event := &AuditEvent{
		Type:     AuditEventTypeSecurity,
		Level:    level,
		Message:  message,
		Details:  details,
		Category: "security",
		Tags:     []string{"security", string(level)},
	}

	return al.LogEvent(ctx, event)
}

// LogSystemEvent logs a system event
func (al *AuditLogger) LogSystemEvent(ctx context.Context, level AuditLevel, message string, details map[string]interface{}) error {
	event := &AuditEvent{
		Type:     AuditEventTypeSystem,
		Level:    level,
		Message:  message,
		Details:  details,
		Category: "system",
		Tags:     []string{"system", string(level)},
	}

	return al.LogEvent(ctx, event)
}

// LogComplianceEvent logs a compliance event
func (al *AuditLogger) LogComplianceEvent(ctx context.Context, message string, details map[string]interface{}) error {
	event := &AuditEvent{
		Type:     AuditEventTypeCompliance,
		Level:    AuditLevelInfo,
		Message:  message,
		Details:  details,
		Category: "compliance",
		Tags:     []string{"compliance"},
	}

	return al.LogEvent(ctx, event)
}

// LogAdminEvent logs an administrative event
func (al *AuditLogger) LogAdminEvent(ctx context.Context, userID, action, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		Type:     AuditEventTypeAdmin,
		Level:    AuditLevelInfo,
		UserID:   userID,
		Action:   action,
		Result:   result,
		Message:  fmt.Sprintf("Admin %s: %s", action, result),
		Details:  details,
		Category: "administrative",
		Tags:     []string{"admin", action, result},
	}

	return al.LogEvent(ctx, event)
}

// GetEvents retrieves audit events with optional filtering
func (al *AuditLogger) GetEvents(filter *AuditFilter) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	if filter == nil {
		return al.events
	}

	var filtered []AuditEvent
	for _, event := range al.events {
		if al.matchesFilter(event, filter) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// AuditFilter represents a filter for audit events
type AuditFilter struct {
	UserID    string         `json:"user_id,omitempty"`
	Type      AuditEventType `json:"type,omitempty"`
	Level     AuditLevel     `json:"level,omitempty"`
	Resource  string         `json:"resource,omitempty"`
	Action    string         `json:"action,omitempty"`
	Result    string         `json:"result,omitempty"`
	StartTime time.Time      `json:"start_time,omitempty"`
	EndTime   time.Time      `json:"end_time,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
	Limit     int            `json:"limit,omitempty"`
}

// matchesFilter checks if an event matches the filter
func (al *AuditLogger) matchesFilter(event AuditEvent, filter *AuditFilter) bool {
	if filter.UserID != "" && event.UserID != filter.UserID {
		return false
	}
	if filter.Type != "" && event.Type != filter.Type {
		return false
	}
	if filter.Level != "" && event.Level != filter.Level {
		return false
	}
	if filter.Resource != "" && event.Resource != filter.Resource {
		return false
	}
	if filter.Action != "" && event.Action != filter.Action {
		return false
	}
	if filter.Result != "" && event.Result != filter.Result {
		return false
	}
	if !filter.StartTime.IsZero() && event.Timestamp.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && event.Timestamp.After(filter.EndTime) {
		return false
	}
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, tag := range filter.Tags {
			for _, eventTag := range event.Tags {
				if tag == eventTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

// flush writes buffered events to the log file
func (al *AuditLogger) flush() error {
	al.bufferMu.Lock()
	defer al.bufferMu.Unlock()

	if len(al.buffer) == 0 {
		return nil
	}

	// Write events to file
	for _, event := range al.buffer {
		data, err := json.Marshal(event)
		if err != nil {
			continue // Skip malformed events
		}

		_, err = al.logFile.Write(append(data, '\n'))
		if err != nil {
			return err
		}
	}

	// Clear buffer
	al.buffer = al.buffer[:0]

	// Sync file
	return al.logFile.Sync()
}

// flushRoutine periodically flushes the buffer
func (al *AuditLogger) flushRoutine() {
	ticker := time.NewTicker(al.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = al.flush() // Ignore error in periodic flush
		case <-al.stopChan:
			_ = al.flush() // Final flush - ignore error
			return
		}
	}
}

// Close closes the audit logger
func (al *AuditLogger) Close() error {
	close(al.stopChan)
	return al.logFile.Close()
}

// GetAuditSummary returns a summary of audit events
func (al *AuditLogger) GetAuditSummary(startTime, endTime time.Time) *AuditSummary {
	al.mu.RLock()
	defer al.mu.RUnlock()

	summary := &AuditSummary{
		StartTime: startTime,
		EndTime:   endTime,
	}

	for _, event := range al.events {
		if !startTime.IsZero() && event.Timestamp.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && event.Timestamp.After(endTime) {
			continue
		}

		summary.TotalEvents++

		// Count by level
		switch event.Level {
		case AuditLevelInfo:
			summary.InfoCount++
		case AuditLevelWarning:
			summary.WarningCount++
		case AuditLevelError:
			summary.ErrorCount++
		case AuditLevelCritical:
			summary.CriticalCount++
		}

		// Count by type
		switch event.Type {
		case AuditEventTypeAuth:
			summary.AuthEvents++
		case AuditEventTypeAccess:
			summary.AccessEvents++
		case AuditEventTypeData:
			summary.DataEvents++
		case AuditEventTypeSystem:
			summary.SystemEvents++
		case AuditEventTypeSecurity:
			summary.SecurityEvents++
		case AuditEventTypeCompliance:
			summary.ComplianceEvents++
		case AuditEventTypeAdmin:
			summary.AdminEvents++
		}

		// Count by result
		switch event.Result {
		case "success":
			summary.SuccessCount++
		case "failure":
			summary.FailureCount++
		case "denied":
			summary.DeniedCount++
		}
	}

	return summary
}

// AuditSummary represents a summary of audit events
type AuditSummary struct {
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	TotalEvents      int       `json:"total_events"`
	InfoCount        int       `json:"info_count"`
	WarningCount     int       `json:"warning_count"`
	ErrorCount       int       `json:"error_count"`
	CriticalCount    int       `json:"critical_count"`
	AuthEvents       int       `json:"auth_events"`
	AccessEvents     int       `json:"access_events"`
	DataEvents       int       `json:"data_events"`
	SystemEvents     int       `json:"system_events"`
	SecurityEvents   int       `json:"security_events"`
	ComplianceEvents int       `json:"compliance_events"`
	AdminEvents      int       `json:"admin_events"`
	SuccessCount     int       `json:"success_count"`
	FailureCount     int       `json:"failure_count"`
	DeniedCount      int       `json:"denied_count"`
}

// Global audit logger instance
var GlobalAuditLogger *AuditLogger

// InitializeAuditLogger initializes the global audit logger
func InitializeAuditLogger(logPath string) error {
	var err error
	GlobalAuditLogger, err = NewAuditLogger(logPath)
	return err
}

// Convenience functions
func LogEvent(ctx context.Context, event *AuditEvent) error {
	if GlobalAuditLogger == nil {
		return fmt.Errorf("audit logger not initialized")
	}
	return GlobalAuditLogger.LogEvent(ctx, event)
}

func LogAuthEvent(ctx context.Context, userID, action, result, ipAddress, userAgent string, details map[string]interface{}) error {
	if GlobalAuditLogger == nil {
		return fmt.Errorf("audit logger not initialized")
	}
	return GlobalAuditLogger.LogAuthEvent(ctx, userID, action, result, ipAddress, userAgent, details)
}

func LogAccessEvent(ctx context.Context, userID, resource, action, result string, details map[string]interface{}) error {
	if GlobalAuditLogger == nil {
		return fmt.Errorf("audit logger not initialized")
	}
	return GlobalAuditLogger.LogAccessEvent(ctx, userID, resource, action, result, details)
}

func LogDataEvent(ctx context.Context, userID, resource, action, result string, details map[string]interface{}) error {
	if GlobalAuditLogger == nil {
		return fmt.Errorf("audit logger not initialized")
	}
	return GlobalAuditLogger.LogDataEvent(ctx, userID, resource, action, result, details)
}

func LogSecurityEvent(ctx context.Context, level AuditLevel, message string, details map[string]interface{}) error {
	if GlobalAuditLogger == nil {
		return fmt.Errorf("audit logger not initialized")
	}
	return GlobalAuditLogger.LogSecurityEvent(ctx, level, message, details)
}

func LogSystemEvent(ctx context.Context, level AuditLevel, message string, details map[string]interface{}) error {
	if GlobalAuditLogger == nil {
		return fmt.Errorf("audit logger not initialized")
	}
	return GlobalAuditLogger.LogSystemEvent(ctx, level, message, details)
}

func LogComplianceEvent(ctx context.Context, message string, details map[string]interface{}) error {
	if GlobalAuditLogger == nil {
		return fmt.Errorf("audit logger not initialized")
	}
	return GlobalAuditLogger.LogComplianceEvent(ctx, message, details)
}

func LogAdminEvent(ctx context.Context, userID, action, result string, details map[string]interface{}) error {
	if GlobalAuditLogger == nil {
		return fmt.Errorf("audit logger not initialized")
	}
	return GlobalAuditLogger.LogAdminEvent(ctx, userID, action, result, details)
}

func GetAuditEvents(filter *AuditFilter) []AuditEvent {
	if GlobalAuditLogger == nil {
		return nil
	}
	return GlobalAuditLogger.GetEvents(filter)
}

func GetAuditSummary(startTime, endTime time.Time) *AuditSummary {
	if GlobalAuditLogger == nil {
		return nil
	}
	return GlobalAuditLogger.GetAuditSummary(startTime, endTime)
}
