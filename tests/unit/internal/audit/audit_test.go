package audit

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuditConstants(t *testing.T) {
	// Test audit level constants
	assert.Equal(t, AuditLevel("info"), AuditLevelInfo)
	assert.Equal(t, AuditLevel("warning"), AuditLevelWarning)
	assert.Equal(t, AuditLevel("error"), AuditLevelError)
	assert.Equal(t, AuditLevel("critical"), AuditLevelCritical)

	// Test audit event type constants
	assert.Equal(t, AuditEventType("authentication"), AuditEventTypeAuth)
	assert.Equal(t, AuditEventType("access"), AuditEventTypeAccess)
	assert.Equal(t, AuditEventType("data"), AuditEventTypeData)
	assert.Equal(t, AuditEventType("system"), AuditEventTypeSystem)
	assert.Equal(t, AuditEventType("security"), AuditEventTypeSecurity)
	assert.Equal(t, AuditEventType("compliance"), AuditEventTypeCompliance)
	assert.Equal(t, AuditEventType("administrative"), AuditEventTypeAdmin)
}

func TestAuditEvent(t *testing.T) {
	event := &AuditEvent{
		ID:        "test-id",
		Type:      AuditEventTypeAuth,
		Level:     AuditLevelInfo,
		UserID:    "user123",
		SessionID: "session456",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Resource:  "/api/users",
		Action:    "login",
		Result:    "success",
		Message:   "User logged in successfully",
		Details:   map[string]interface{}{"method": "password"},
		Metadata:  map[string]interface{}{"version": "1.0"},
		Timestamp: time.Now(),
		Source:    "auth-service",
		Category:  "authentication",
		Tags:      []string{"auth", "login", "success"},
	}

	assert.Equal(t, "test-id", event.ID)
	assert.Equal(t, AuditEventTypeAuth, event.Type)
	assert.Equal(t, AuditLevelInfo, event.Level)
	assert.Equal(t, "user123", event.UserID)
	assert.Equal(t, "session456", event.SessionID)
	assert.Equal(t, "192.168.1.1", event.IPAddress)
	assert.Equal(t, "Mozilla/5.0", event.UserAgent)
	assert.Equal(t, "/api/users", event.Resource)
	assert.Equal(t, "login", event.Action)
	assert.Equal(t, "success", event.Result)
	assert.Equal(t, "User logged in successfully", event.Message)
	assert.Equal(t, "auth-service", event.Source)
	assert.Equal(t, "authentication", event.Category)
	assert.Len(t, event.Tags, 3)
}

func TestNewAuditLogger(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.events)
	assert.NotNil(t, logger.buffer)
	assert.Equal(t, logPath, logger.logPath)
	assert.Equal(t, 100, logger.batchSize)
	assert.Equal(t, 30*time.Second, logger.flushInterval)
	assert.NotNil(t, logger.stopChan)

	// Clean up
	err = logger.Close()
	assert.NoError(t, err)
}

func TestNewAuditLogger_InvalidPath(t *testing.T) {
	// Test with invalid path (parent directory doesn't exist and can't be created)
	// This test might pass on some systems, so we'll make it more robust
	logger, err := NewAuditLogger("/root/restricted/audit.log")
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, logger)
		assert.Contains(t, err.Error(), "failed to create log directory")
	} else {
		// If it succeeds, clean up
		if logger != nil {
			_ = logger.Close()
		}
	}
}

func TestAuditLogger_LogEvent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	// Test logging an event with all fields
	event := &AuditEvent{
		ID:        "test-event-1",
		Type:      AuditEventTypeAuth,
		Level:     AuditLevelInfo,
		UserID:    "user123",
		Message:   "Test event",
		Timestamp: time.Now(),
	}

	err = logger.LogEvent(ctx, event)
	assert.NoError(t, err)

	// Test logging an event with minimal fields (should set defaults)
	event2 := &AuditEvent{
		Message: "Test event 2",
	}

	err = logger.LogEvent(ctx, event2)
	assert.NoError(t, err)

	// Verify events are stored
	events := logger.GetEvents(nil)
	assert.Len(t, events, 2)

	// Check first event
	assert.Equal(t, "test-event-1", events[0].ID)
	assert.Equal(t, AuditEventTypeAuth, events[0].Type)
	assert.Equal(t, AuditLevelInfo, events[0].Level)

	// Check second event (should have defaults set)
	assert.NotEmpty(t, events[1].ID)
	assert.Equal(t, AuditLevelInfo, events[1].Level) // Default level
	assert.NotZero(t, events[1].Timestamp)           // Default timestamp
}

func TestAuditLogger_LogAuthEvent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	details := map[string]interface{}{
		"method":   "password",
		"provider": "local",
	}

	err = logger.LogAuthEvent(ctx, "user123", "login", "success", "192.168.1.1", "Mozilla/5.0", details)
	assert.NoError(t, err)

	events := logger.GetEvents(nil)
	assert.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, AuditEventTypeAuth, event.Type)
	assert.Equal(t, AuditLevelInfo, event.Level)
	assert.Equal(t, "user123", event.UserID)
	assert.Equal(t, "login", event.Action)
	assert.Equal(t, "success", event.Result)
	assert.Equal(t, "192.168.1.1", event.IPAddress)
	assert.Equal(t, "Mozilla/5.0", event.UserAgent)
	assert.Equal(t, "authentication", event.Category)
	assert.Contains(t, event.Tags, "auth")
	assert.Contains(t, event.Tags, "login")
	assert.Contains(t, event.Tags, "success")
	assert.Equal(t, "password", event.Details["method"])
	assert.Equal(t, "local", event.Details["provider"])
}

func TestAuditLogger_LogAccessEvent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	details := map[string]interface{}{
		"permission":    "read",
		"resource_type": "file",
	}

	err = logger.LogAccessEvent(ctx, "user123", "/api/files/123", "read", "success", details)
	assert.NoError(t, err)

	events := logger.GetEvents(nil)
	assert.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, AuditEventTypeAccess, event.Type)
	assert.Equal(t, AuditLevelInfo, event.Level)
	assert.Equal(t, "user123", event.UserID)
	assert.Equal(t, "/api/files/123", event.Resource)
	assert.Equal(t, "read", event.Action)
	assert.Equal(t, "success", event.Result)
	assert.Equal(t, "access_control", event.Category)
	assert.Contains(t, event.Tags, "access")
	assert.Contains(t, event.Tags, "read")
	assert.Contains(t, event.Tags, "success")
}

func TestAuditLogger_LogDataEvent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	details := map[string]interface{}{
		"data_type": "user_profile",
		"operation": "update",
	}

	err = logger.LogDataEvent(ctx, "user123", "/api/users/123", "update", "success", details)
	assert.NoError(t, err)

	events := logger.GetEvents(nil)
	assert.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, AuditEventTypeData, event.Type)
	assert.Equal(t, AuditLevelInfo, event.Level)
	assert.Equal(t, "user123", event.UserID)
	assert.Equal(t, "/api/users/123", event.Resource)
	assert.Equal(t, "update", event.Action)
	assert.Equal(t, "success", event.Result)
	assert.Equal(t, "data_access", event.Category)
	assert.Contains(t, event.Tags, "data")
	assert.Contains(t, event.Tags, "update")
	assert.Contains(t, event.Tags, "success")
}

func TestAuditLogger_LogSecurityEvent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	details := map[string]interface{}{
		"threat_type": "brute_force",
		"attempts":    5,
	}

	err = logger.LogSecurityEvent(ctx, AuditLevelWarning, "Multiple failed login attempts detected", details)
	assert.NoError(t, err)

	events := logger.GetEvents(nil)
	assert.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, AuditEventTypeSecurity, event.Type)
	assert.Equal(t, AuditLevelWarning, event.Level)
	assert.Equal(t, "Multiple failed login attempts detected", event.Message)
	assert.Equal(t, "security", event.Category)
	assert.Contains(t, event.Tags, "security")
	assert.Contains(t, event.Tags, "warning")
}

func TestAuditLogger_LogSystemEvent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	details := map[string]interface{}{
		"component": "database",
		"status":    "healthy",
	}

	err = logger.LogSystemEvent(ctx, AuditLevelInfo, "Database connection established", details)
	assert.NoError(t, err)

	events := logger.GetEvents(nil)
	assert.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, AuditEventTypeSystem, event.Type)
	assert.Equal(t, AuditLevelInfo, event.Level)
	assert.Equal(t, "Database connection established", event.Message)
	assert.Equal(t, "system", event.Category)
	assert.Contains(t, event.Tags, "system")
	assert.Contains(t, event.Tags, "info")
}

func TestAuditLogger_LogComplianceEvent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	details := map[string]interface{}{
		"regulation": "GDPR",
		"article":    "Article 32",
	}

	err = logger.LogComplianceEvent(ctx, "Data processing activity logged for GDPR compliance", details)
	assert.NoError(t, err)

	events := logger.GetEvents(nil)
	assert.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, AuditEventTypeCompliance, event.Type)
	assert.Equal(t, AuditLevelInfo, event.Level)
	assert.Equal(t, "Data processing activity logged for GDPR compliance", event.Message)
	assert.Equal(t, "compliance", event.Category)
	assert.Contains(t, event.Tags, "compliance")
}

func TestAuditLogger_LogAdminEvent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	details := map[string]interface{}{
		"target_user": "user456",
		"permission":  "admin",
	}

	err = logger.LogAdminEvent(ctx, "admin123", "grant_permission", "success", details)
	assert.NoError(t, err)

	events := logger.GetEvents(nil)
	assert.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, AuditEventTypeAdmin, event.Type)
	assert.Equal(t, AuditLevelInfo, event.Level)
	assert.Equal(t, "admin123", event.UserID)
	assert.Equal(t, "grant_permission", event.Action)
	assert.Equal(t, "success", event.Result)
	assert.Equal(t, "administrative", event.Category)
	assert.Contains(t, event.Tags, "admin")
	assert.Contains(t, event.Tags, "grant_permission")
	assert.Contains(t, event.Tags, "success")
}

func TestAuditLogger_GetEvents_WithFilter(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	// Log multiple events
	_ = logger.LogAuthEvent(ctx, "user1", "login", "success", "192.168.1.1", "Mozilla/5.0", nil)
	_ = logger.LogAuthEvent(ctx, "user2", "login", "failure", "192.168.1.2", "Chrome/5.0", nil)
	_ = logger.LogAccessEvent(ctx, "user1", "/api/files", "read", "success", nil)
	_ = logger.LogSecurityEvent(ctx, AuditLevelWarning, "Security alert", nil)

	// Test filter by user ID
	filter := &AuditFilter{UserID: "user1"}
	events := logger.GetEvents(filter)
	assert.Len(t, events, 2)

	// Test filter by type
	filter = &AuditFilter{Type: AuditEventTypeAuth}
	events = logger.GetEvents(filter)
	assert.Len(t, events, 2)

	// Test filter by level
	filter = &AuditFilter{Level: AuditLevelWarning}
	events = logger.GetEvents(filter)
	assert.Len(t, events, 1)

	// Test filter by action
	filter = &AuditFilter{Action: "login"}
	events = logger.GetEvents(filter)
	assert.Len(t, events, 2)

	// Test filter by result
	filter = &AuditFilter{Result: "success"}
	events = logger.GetEvents(filter)
	assert.Len(t, events, 2)

	// Test filter by resource
	filter = &AuditFilter{Resource: "/api/files"}
	events = logger.GetEvents(filter)
	assert.Len(t, events, 1)

	// Test filter by tags
	filter = &AuditFilter{Tags: []string{"auth"}}
	events = logger.GetEvents(filter)
	assert.Len(t, events, 2)

	// Test filter by time range
	now := time.Now()
	filter = &AuditFilter{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now.Add(1 * time.Hour),
	}
	events = logger.GetEvents(filter)
	assert.Len(t, events, 4)

	// Test filter with limit
	filter = &AuditFilter{Limit: 2}
	events = logger.GetEvents(filter)
	assert.Len(t, events, 2)
}

func TestAuditLogger_GetEvents_NoFilter(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	// Log some events
	_ = logger.LogAuthEvent(ctx, "user1", "login", "success", "192.168.1.1", "Mozilla/5.0", nil)
	_ = logger.LogAccessEvent(ctx, "user1", "/api/files", "read", "success", nil)

	// Get all events without filter
	events := logger.GetEvents(nil)
	assert.Len(t, events, 2)
}

func TestAuditLogger_GetAuditSummary(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	// Log various types of events
	_ = logger.LogAuthEvent(ctx, "user1", "login", "success", "192.168.1.1", "Mozilla/5.0", nil)
	_ = logger.LogAuthEvent(ctx, "user2", "login", "failure", "192.168.1.2", "Chrome/5.0", nil)
	_ = logger.LogAccessEvent(ctx, "user1", "/api/files", "read", "success", nil)
	_ = logger.LogDataEvent(ctx, "user1", "/api/users", "update", "success", nil)
	_ = logger.LogSecurityEvent(ctx, AuditLevelWarning, "Security alert", nil)
	_ = logger.LogSystemEvent(ctx, AuditLevelError, "System error", nil)
	_ = logger.LogComplianceEvent(ctx, "Compliance event", nil)
	_ = logger.LogAdminEvent(ctx, "admin1", "create_user", "success", nil)

	// Get summary
	now := time.Now()
	summary := logger.GetAuditSummary(now.Add(-1*time.Hour), now.Add(1*time.Hour))

	assert.NotNil(t, summary)
	assert.Equal(t, 8, summary.TotalEvents)
	assert.Equal(t, 6, summary.InfoCount)    // auth, access, data, compliance, admin events
	assert.Equal(t, 1, summary.WarningCount) // security event
	assert.Equal(t, 1, summary.ErrorCount)   // system event
	assert.Equal(t, 0, summary.CriticalCount)
	assert.Equal(t, 2, summary.AuthEvents)
	assert.Equal(t, 1, summary.AccessEvents)
	assert.Equal(t, 1, summary.DataEvents)
	assert.Equal(t, 1, summary.SystemEvents)
	assert.Equal(t, 1, summary.SecurityEvents)
	assert.Equal(t, 1, summary.ComplianceEvents)
	assert.Equal(t, 1, summary.AdminEvents)
	assert.Equal(t, 4, summary.SuccessCount) // success results
	assert.Equal(t, 1, summary.FailureCount) // failure results
	assert.Equal(t, 0, summary.DeniedCount)
}

func TestAuditLogger_ConcurrentAccess(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	// Test concurrent logging
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			// Log different types of events concurrently
			switch i % 4 {
			case 0:
				_ = logger.LogAuthEvent(ctx, "user1", "login", "success", "192.168.1.1", "Mozilla/5.0", nil)
			case 1:
				_ = logger.LogAccessEvent(ctx, "user1", "/api/files", "read", "success", nil)
			case 2:
				_ = logger.LogSecurityEvent(ctx, AuditLevelInfo, "Security event", nil)
			case 3:
				_ = logger.LogSystemEvent(ctx, AuditLevelInfo, "System event", nil)
			}

			// Test concurrent access to GetEvents
			events := logger.GetEvents(nil)
			assert.GreaterOrEqual(t, len(events), 1)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all events were logged
	events := logger.GetEvents(nil)
	assert.Len(t, events, 10)
}

func TestGlobalAuditLogger(t *testing.T) {
	// Test that global logger is initially nil
	assert.Nil(t, GlobalAuditLogger)

	// Test initialization
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	err := InitializeAuditLogger(logPath)
	assert.NoError(t, err)
	assert.NotNil(t, GlobalAuditLogger)

	// Clean up
	err = GlobalAuditLogger.Close()
	assert.NoError(t, err)
}

func TestConvenienceFunctions(t *testing.T) {
	// Initialize global logger
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	err := InitializeAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = GlobalAuditLogger.Close() }()

	ctx := context.Background()

	// Test LogEvent
	event := &AuditEvent{
		Type:    AuditEventTypeAuth,
		Level:   AuditLevelInfo,
		Message: "Test event",
	}

	err = LogEvent(ctx, event)
	assert.NoError(t, err)

	// Test LogAuthEvent
	err = LogAuthEvent(ctx, "user123", "login", "success", "192.168.1.1", "Mozilla/5.0", nil)
	assert.NoError(t, err)

	// Test LogAccessEvent
	err = LogAccessEvent(ctx, "user123", "/api/files", "read", "success", nil)
	assert.NoError(t, err)

	// Test LogDataEvent
	err = LogDataEvent(ctx, "user123", "/api/users", "update", "success", nil)
	assert.NoError(t, err)

	// Test LogSecurityEvent
	err = LogSecurityEvent(ctx, AuditLevelWarning, "Security alert", nil)
	assert.NoError(t, err)

	// Test LogSystemEvent
	err = LogSystemEvent(ctx, AuditLevelInfo, "System event", nil)
	assert.NoError(t, err)

	// Test LogComplianceEvent
	err = LogComplianceEvent(ctx, "Compliance event", nil)
	assert.NoError(t, err)

	// Test LogAdminEvent
	err = LogAdminEvent(ctx, "admin123", "create_user", "success", nil)
	assert.NoError(t, err)

	// Test GetAuditEvents
	events := GetAuditEvents(nil)
	assert.Len(t, events, 8)

	// Test GetAuditSummary
	now := time.Now()
	summary := GetAuditSummary(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	assert.NotNil(t, summary)
	assert.Equal(t, 8, summary.TotalEvents)
}

func TestConvenienceFunctions_NoGlobalLogger(t *testing.T) {
	// Reset global logger
	GlobalAuditLogger = nil

	ctx := context.Background()

	// Test that convenience functions return error when global logger is nil
	event := &AuditEvent{Message: "Test"}
	err := LogEvent(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit logger not initialized")

	err = LogAuthEvent(ctx, "user123", "login", "success", "192.168.1.1", "Mozilla/5.0", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit logger not initialized")

	err = LogAccessEvent(ctx, "user123", "/api/files", "read", "success", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit logger not initialized")

	err = LogDataEvent(ctx, "user123", "/api/users", "update", "success", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit logger not initialized")

	err = LogSecurityEvent(ctx, AuditLevelWarning, "Security alert", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit logger not initialized")

	err = LogSystemEvent(ctx, AuditLevelInfo, "System event", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit logger not initialized")

	err = LogComplianceEvent(ctx, "Compliance event", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit logger not initialized")

	err = LogAdminEvent(ctx, "admin123", "create_user", "success", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit logger not initialized")

	// Test that GetAuditEvents returns nil when global logger is nil
	events := GetAuditEvents(nil)
	assert.Nil(t, events)

	// Test that GetAuditSummary returns nil when global logger is nil
	summary := GetAuditSummary(time.Now(), time.Now())
	assert.Nil(t, summary)
}

func TestAuditLogger_EdgeCases(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	logger, err := NewAuditLogger(logPath)
	assert.NoError(t, err)
	defer func() { _ = logger.Close() }()

	ctx := context.Background()

	// Test logging event with empty fields
	event := &AuditEvent{
		Message: "Test event with empty fields",
	}

	err = logger.LogEvent(ctx, event)
	assert.NoError(t, err)

	// Test logging another event
	event2 := &AuditEvent{
		Message: "Test event 2",
	}

	err = logger.LogEvent(ctx, event2)
	assert.NoError(t, err)

	// Test filter with empty values
	filter := &AuditFilter{
		UserID:   "",
		Type:     "",
		Level:    "",
		Resource: "",
		Action:   "",
		Result:   "",
		Tags:     []string{},
	}

	events := logger.GetEvents(filter)
	assert.Len(t, events, 2) // Should return all events when filter is empty

	// Test summary with zero time range
	summary := logger.GetAuditSummary(time.Time{}, time.Time{})
	assert.NotNil(t, summary)
	assert.Equal(t, 2, summary.TotalEvents)
}
