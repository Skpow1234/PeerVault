package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AuditManager manages audit logging
type AuditManager struct {
	configDir string
	logFile   string
	mu        sync.Mutex
}

// AuditEvent represents an audit log event
type AuditEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"` // "success", "failure", "denied"
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Details   map[string]interface{} `json:"details"`
	RiskLevel string                 `json:"risk_level"` // "low", "medium", "high", "critical"
	Category  string                 `json:"category"`   // "auth", "file", "system", "security"
}

// NewAuditManager creates a new audit manager
func NewAuditManager(configDir string) *AuditManager {
	return &AuditManager{
		configDir: configDir,
		logFile:   filepath.Join(configDir, "audit.log"),
	}
}

// LogEvent logs an audit event
func (am *AuditManager) LogEvent(event *AuditEvent) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Generate event ID if not provided
	if event.ID == "" {
		event.ID = am.generateEventID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Determine risk level if not provided
	if event.RiskLevel == "" {
		event.RiskLevel = am.determineRiskLevel(event)
	}

	// Create log entry
	logEntry := map[string]interface{}{
		"id":         event.ID,
		"timestamp":  event.Timestamp.Format(time.RFC3339),
		"user_id":    event.UserID,
		"username":   event.Username,
		"action":     event.Action,
		"resource":   event.Resource,
		"result":     event.Result,
		"ip_address": event.IPAddress,
		"user_agent": event.UserAgent,
		"details":    event.Details,
		"risk_level": event.RiskLevel,
		"category":   event.Category,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Append to log file
	file, err := os.OpenFile(am.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer func() { _ = file.Close() }()

	_, err = file.WriteString(string(jsonData) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	return nil
}

// LogAuthentication logs authentication events
func (am *AuditManager) LogAuthentication(userID, username, action, result, ipAddress, userAgent string, details map[string]interface{}) error {
	event := &AuditEvent{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Resource:  "authentication",
		Result:    result,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
		Category:  "auth",
	}

	return am.LogEvent(event)
}

// LogFileOperation logs file operation events
func (am *AuditManager) LogFileOperation(userID, username, action, resource, result, ipAddress string, details map[string]interface{}) error {
	event := &AuditEvent{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Resource:  resource,
		Result:    result,
		IPAddress: ipAddress,
		Details:   details,
		Category:  "file",
	}

	return am.LogEvent(event)
}

// LogSystemOperation logs system operation events
func (am *AuditManager) LogSystemOperation(userID, username, action, resource, result, ipAddress string, details map[string]interface{}) error {
	event := &AuditEvent{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Resource:  resource,
		Result:    result,
		IPAddress: ipAddress,
		Details:   details,
		Category:  "system",
	}

	return am.LogEvent(event)
}

// LogSecurityEvent logs security-related events
func (am *AuditManager) LogSecurityEvent(userID, username, action, resource, result, ipAddress string, details map[string]interface{}) error {
	event := &AuditEvent{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Resource:  resource,
		Result:    result,
		IPAddress: ipAddress,
		Details:   details,
		Category:  "security",
	}

	return am.LogEvent(event)
}

// QueryEvents queries audit events with filters
func (am *AuditManager) QueryEvents(filters map[string]interface{}) ([]*AuditEvent, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Read all events from log file
	events, err := am.readAllEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	// Apply filters
	var filteredEvents []*AuditEvent
	for _, event := range events {
		if am.matchesFilters(event, filters) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// GetEventsByUser returns events for a specific user
func (am *AuditManager) GetEventsByUser(userID string, limit int) ([]*AuditEvent, error) {
	filters := map[string]interface{}{
		"user_id": userID,
	}

	events, err := am.QueryEvents(filters)
	if err != nil {
		return nil, err
	}

	// Limit results
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

// GetEventsByTimeRange returns events within a time range
func (am *AuditManager) GetEventsByTimeRange(start, end time.Time) ([]*AuditEvent, error) {
	filters := map[string]interface{}{
		"start_time": start,
		"end_time":   end,
	}

	return am.QueryEvents(filters)
}

// GetHighRiskEvents returns high-risk events
func (am *AuditManager) GetHighRiskEvents(limit int) ([]*AuditEvent, error) {
	filters := map[string]interface{}{
		"risk_level": []string{"high", "critical"},
	}

	events, err := am.QueryEvents(filters)
	if err != nil {
		return nil, err
	}

	// Limit results
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

// GetFailedLogins returns failed login attempts
func (am *AuditManager) GetFailedLogins(limit int) ([]*AuditEvent, error) {
	filters := map[string]interface{}{
		"category": "auth",
		"action":   "login",
		"result":   "failure",
	}

	events, err := am.QueryEvents(filters)
	if err != nil {
		return nil, err
	}

	// Limit results
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

// GenerateReport generates an audit report
func (am *AuditManager) GenerateReport(start, end time.Time) (*AuditReport, error) {
	events, err := am.GetEventsByTimeRange(start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	report := &AuditReport{
		StartTime:   start,
		EndTime:     end,
		TotalEvents: len(events),
	}

	// Analyze events
	report.AnalyzeEvents(events)

	return report, nil
}

// Utility functions
func (am *AuditManager) generateEventID() string {
	return fmt.Sprintf("audit_%d", time.Now().UnixNano())
}

func (am *AuditManager) determineRiskLevel(event *AuditEvent) string {
	// High-risk actions
	highRiskActions := []string{
		"delete", "revoke", "deactivate", "admin_access",
		"certificate_generate", "user_create", "role_modify",
	}

	for _, action := range highRiskActions {
		if event.Action == action {
			return "high"
		}
	}

	// Failed operations are medium risk
	if event.Result == "failure" || event.Result == "denied" {
		return "medium"
	}

	// Authentication failures are high risk
	if event.Category == "auth" && event.Result == "failure" {
		return "high"
	}

	// Default to low risk
	return "low"
}

func (am *AuditManager) readAllEvents() ([]*AuditEvent, error) {
	if _, err := os.Stat(am.logFile); os.IsNotExist(err) {
		return []*AuditEvent{}, nil
	}

	data, err := os.ReadFile(am.logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit log: %w", err)
	}

	var events []*AuditEvent
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var event AuditEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip malformed entries
		}

		events = append(events, &event)
	}

	return events, nil
}

func (am *AuditManager) matchesFilters(event *AuditEvent, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "user_id":
			if event.UserID != value {
				return false
			}
		case "username":
			if event.Username != value {
				return false
			}
		case "action":
			if event.Action != value {
				return false
			}
		case "result":
			if event.Result != value {
				return false
			}
		case "category":
			if event.Category != value {
				return false
			}
		case "risk_level":
			if riskLevels, ok := value.([]string); ok {
				found := false
				for _, level := range riskLevels {
					if event.RiskLevel == level {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			} else if event.RiskLevel != value {
				return false
			}
		case "start_time":
			if startTime, ok := value.(time.Time); ok {
				if event.Timestamp.Before(startTime) {
					return false
				}
			}
		case "end_time":
			if endTime, ok := value.(time.Time); ok {
				if event.Timestamp.After(endTime) {
					return false
				}
			}
		}
	}

	return true
}

// AuditReport represents an audit report
type AuditReport struct {
	StartTime      time.Time      `json:"start_time"`
	EndTime        time.Time      `json:"end_time"`
	TotalEvents    int            `json:"total_events"`
	EventsByUser   map[string]int `json:"events_by_user"`
	EventsByAction map[string]int `json:"events_by_action"`
	EventsByResult map[string]int `json:"events_by_result"`
	EventsByRisk   map[string]int `json:"events_by_risk"`
	FailedLogins   int            `json:"failed_logins"`
	HighRiskEvents int            `json:"high_risk_events"`
}

// AnalyzeEvents analyzes events for the report
func (ar *AuditReport) AnalyzeEvents(events []*AuditEvent) {
	ar.EventsByUser = make(map[string]int)
	ar.EventsByAction = make(map[string]int)
	ar.EventsByResult = make(map[string]int)
	ar.EventsByRisk = make(map[string]int)

	for _, event := range events {
		// Count by user
		ar.EventsByUser[event.Username]++

		// Count by action
		ar.EventsByAction[event.Action]++

		// Count by result
		ar.EventsByResult[event.Result]++

		// Count by risk level
		ar.EventsByRisk[event.RiskLevel]++

		// Count failed logins
		if event.Category == "auth" && event.Result == "failure" {
			ar.FailedLogins++
		}

		// Count high-risk events
		if event.RiskLevel == "high" || event.RiskLevel == "critical" {
			ar.HighRiskEvents++
		}
	}
}
