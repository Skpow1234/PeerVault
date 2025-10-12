package network

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// BandwidthManager manages bandwidth allocation and monitoring
type BandwidthManager struct {
	client    *client.Client
	configDir string
	policies  map[string]*BandwidthPolicy
	monitors  map[string]*BandwidthMonitor
	config    *BandwidthConfig
	stats     *BandwidthStats
	mu        sync.RWMutex
}

// BandwidthPolicy represents a bandwidth allocation policy
type BandwidthPolicy struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	UserID         string                 `json:"user_id"`
	MaxBandwidth   int64                  `json:"max_bandwidth"`   // in bytes per second
	BurstBandwidth int64                  `json:"burst_bandwidth"` // in bytes per second
	Priority       int                    `json:"priority"`        // 1-10, higher is more priority
	TimeWindow     time.Duration          `json:"time_window"`     // time window for rate limiting
	AllowedHours   []int                  `json:"allowed_hours"`   // hours of day when policy is active
	AllowedDays    []int                  `json:"allowed_days"`    // days of week when policy is active
	IsActive       bool                   `json:"is_active"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// BandwidthMonitor represents bandwidth usage monitoring
type BandwidthMonitor struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	CurrentUsage int64                  `json:"current_usage"` // current bandwidth usage in bytes per second
	PeakUsage    int64                  `json:"peak_usage"`    // peak bandwidth usage in bytes per second
	TotalUsage   int64                  `json:"total_usage"`   // total bandwidth used in bytes
	LastReset    time.Time              `json:"last_reset"`
	LastUpdated  time.Time              `json:"last_updated"`
	UsageHistory []BandwidthUsage       `json:"usage_history"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// BandwidthUsage represents bandwidth usage at a point in time
type BandwidthUsage struct {
	Timestamp  time.Time `json:"timestamp"`
	Usage      int64     `json:"usage"`       // in bytes per second
	TotalBytes int64     `json:"total_bytes"` // total bytes transferred
}

// BandwidthConfig represents bandwidth manager configuration
type BandwidthConfig struct {
	MonitoringInterval time.Duration `json:"monitoring_interval"`
	HistoryRetention   time.Duration `json:"history_retention"`
	DefaultPolicy      string        `json:"default_policy"`
	EnableThrottling   bool          `json:"enable_throttling"`
	ThrottleThreshold  float64       `json:"throttle_threshold"` // percentage of max bandwidth
	AlertThreshold     float64       `json:"alert_threshold"`    // percentage of max bandwidth
}

// BandwidthStats represents bandwidth statistics
type BandwidthStats struct {
	TotalPolicies      int       `json:"total_policies"`
	ActivePolicies     int       `json:"active_policies"`
	TotalMonitors      int       `json:"total_monitors"`
	TotalBandwidth     int64     `json:"total_bandwidth"`     // in bytes per second
	UsedBandwidth      int64     `json:"used_bandwidth"`      // in bytes per second
	AvailableBandwidth int64     `json:"available_bandwidth"` // in bytes per second
	UtilizationRate    float64   `json:"utilization_rate"`    // percentage
	ThrottledUsers     int       `json:"throttled_users"`
	AlertedUsers       int       `json:"alerted_users"`
	LastUpdated        time.Time `json:"last_updated"`
}

// BandwidthResult represents the result of a bandwidth operation
type BandwidthResult struct {
	Allowed      bool    `json:"allowed"`
	CurrentUsage int64   `json:"current_usage"`
	MaxBandwidth int64   `json:"max_bandwidth"`
	Utilization  float64 `json:"utilization"`
	Throttled    bool    `json:"throttled"`
	Message      string  `json:"message"`
}

// NewBandwidthManager creates a new bandwidth manager
func NewBandwidthManager(client *client.Client, configDir string) *BandwidthManager {
	bm := &BandwidthManager{
		client:    client,
		configDir: configDir,
		policies:  make(map[string]*BandwidthPolicy),
		monitors:  make(map[string]*BandwidthMonitor),
		config:    getDefaultBandwidthConfig(),
		stats:     &BandwidthStats{},
	}

	bm.loadConfig()
	bm.loadPolicies()
	bm.loadMonitors()
	bm.loadStats()

	// Start monitoring routine
	go bm.startMonitoringRoutine()

	return bm
}

// CreatePolicy creates a new bandwidth policy
func (bm *BandwidthManager) CreatePolicy(policy *BandwidthPolicy) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Check if policy already exists
	if _, exists := bm.policies[policy.ID]; exists {
		return fmt.Errorf("policy with ID %s already exists", policy.ID)
	}

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	bm.policies[policy.ID] = policy
	bm.savePolicies()
	bm.updateStats()

	return nil
}

// UpdatePolicy updates an existing bandwidth policy
func (bm *BandwidthManager) UpdatePolicy(policyID string, updates *BandwidthPolicy) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	policy, exists := bm.policies[policyID]
	if !exists {
		return fmt.Errorf("policy with ID %s not found", policyID)
	}

	// Update fields
	if updates.Name != "" {
		policy.Name = updates.Name
	}
	if updates.Description != "" {
		policy.Description = updates.Description
	}
	if updates.MaxBandwidth > 0 {
		policy.MaxBandwidth = updates.MaxBandwidth
	}
	if updates.BurstBandwidth > 0 {
		policy.BurstBandwidth = updates.BurstBandwidth
	}
	if updates.Priority > 0 {
		policy.Priority = updates.Priority
	}
	if updates.TimeWindow > 0 {
		policy.TimeWindow = updates.TimeWindow
	}
	if len(updates.AllowedHours) > 0 {
		policy.AllowedHours = updates.AllowedHours
	}
	if len(updates.AllowedDays) > 0 {
		policy.AllowedDays = updates.AllowedDays
	}

	policy.IsActive = updates.IsActive
	policy.UpdatedAt = time.Now()

	bm.savePolicies()
	bm.updateStats()

	return nil
}

// DeletePolicy deletes a bandwidth policy
func (bm *BandwidthManager) DeletePolicy(policyID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.policies[policyID]; !exists {
		return fmt.Errorf("policy with ID %s not found", policyID)
	}

	delete(bm.policies, policyID)
	bm.savePolicies()
	bm.updateStats()

	return nil
}

// GetPolicy returns a bandwidth policy by ID
func (bm *BandwidthManager) GetPolicy(policyID string) (*BandwidthPolicy, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	policy, exists := bm.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy with ID %s not found", policyID)
	}

	// Return a copy
	policyCopy := *policy
	return &policyCopy, nil
}

// ListPolicies returns all bandwidth policies
func (bm *BandwidthManager) ListPolicies() []*BandwidthPolicy {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var policies []*BandwidthPolicy
	for _, policy := range bm.policies {
		// Return a copy
		policyCopy := *policy
		policies = append(policies, &policyCopy)
	}

	return policies
}

// GetUserPolicy returns the active policy for a user
func (bm *BandwidthManager) GetUserPolicy(userID string) (*BandwidthPolicy, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Find the highest priority active policy for the user
	var bestPolicy *BandwidthPolicy
	highestPriority := 0

	for _, policy := range bm.policies {
		if policy.UserID == userID && policy.IsActive {
			if policy.Priority > highestPriority {
				highestPriority = policy.Priority
				bestPolicy = policy
			}
		}
	}

	if bestPolicy == nil {
		// Return default policy if no user-specific policy found
		if defaultPolicy, exists := bm.policies[bm.config.DefaultPolicy]; exists {
			policyCopy := *defaultPolicy
			return &policyCopy, nil
		}
		return nil, fmt.Errorf("no policy found for user %s", userID)
	}

	// Return a copy
	policyCopy := *bestPolicy
	return &policyCopy, nil
}

// CheckBandwidth checks if a user can use the requested bandwidth
func (bm *BandwidthManager) CheckBandwidth(userID string, requestedBandwidth int64) (*BandwidthResult, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Get user policy
	policy, err := bm.GetUserPolicy(userID)
	if err != nil {
		return nil, err
	}

	// Get or create monitor for user
	monitor, exists := bm.monitors[userID]
	if !exists {
		monitor = &BandwidthMonitor{
			ID:           userID,
			UserID:       userID,
			CurrentUsage: 0,
			PeakUsage:    0,
			TotalUsage:   0,
			LastReset:    time.Now(),
			LastUpdated:  time.Now(),
			UsageHistory: make([]BandwidthUsage, 0),
			Metadata:     make(map[string]interface{}),
		}
		bm.monitors[userID] = monitor
	}

	// Check if within allowed time window
	if !bm.isWithinAllowedTime(policy) {
		return &BandwidthResult{
			Allowed:      false,
			CurrentUsage: monitor.CurrentUsage,
			MaxBandwidth: policy.MaxBandwidth,
			Utilization:  float64(monitor.CurrentUsage) / float64(policy.MaxBandwidth) * 100,
			Throttled:    false,
			Message:      "outside allowed time window",
		}, nil
	}

	// Check bandwidth limits
	totalUsage := monitor.CurrentUsage + requestedBandwidth
	utilization := float64(totalUsage) / float64(policy.MaxBandwidth) * 100

	allowed := totalUsage <= policy.MaxBandwidth
	throttled := utilization > bm.config.ThrottleThreshold*100

	if !allowed && policy.BurstBandwidth > 0 {
		// Check burst bandwidth
		allowed = totalUsage <= policy.BurstBandwidth
	}

	// Update monitor
	monitor.CurrentUsage = totalUsage
	if totalUsage > monitor.PeakUsage {
		monitor.PeakUsage = totalUsage
	}
	monitor.TotalUsage += requestedBandwidth
	monitor.LastUpdated = time.Now()

	// Add to usage history
	usage := BandwidthUsage{
		Timestamp:  time.Now(),
		Usage:      totalUsage,
		TotalBytes: monitor.TotalUsage,
	}
	monitor.UsageHistory = append(monitor.UsageHistory, usage)

	// Keep only recent history
	if len(monitor.UsageHistory) > 1000 {
		monitor.UsageHistory = monitor.UsageHistory[len(monitor.UsageHistory)-1000:]
	}

	bm.saveMonitors()
	bm.updateStats()

	return &BandwidthResult{
		Allowed:      allowed,
		CurrentUsage: totalUsage,
		MaxBandwidth: policy.MaxBandwidth,
		Utilization:  utilization,
		Throttled:    throttled,
		Message:      bm.getMessage(allowed, throttled),
	}, nil
}

// RecordUsage records bandwidth usage for a user
func (bm *BandwidthManager) RecordUsage(userID string, bytesUsed int64) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	monitor, exists := bm.monitors[userID]
	if !exists {
		monitor = &BandwidthMonitor{
			ID:           userID,
			UserID:       userID,
			CurrentUsage: 0,
			PeakUsage:    0,
			TotalUsage:   0,
			LastReset:    time.Now(),
			LastUpdated:  time.Now(),
			UsageHistory: make([]BandwidthUsage, 0),
			Metadata:     make(map[string]interface{}),
		}
		bm.monitors[userID] = monitor
	}

	monitor.TotalUsage += bytesUsed
	monitor.LastUpdated = time.Now()

	bm.saveMonitors()
	bm.updateStats()

	return nil
}

// ResetUsage resets bandwidth usage for a user
func (bm *BandwidthManager) ResetUsage(userID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	monitor, exists := bm.monitors[userID]
	if !exists {
		return fmt.Errorf("monitor for user %s not found", userID)
	}

	monitor.CurrentUsage = 0
	monitor.PeakUsage = 0
	monitor.TotalUsage = 0
	monitor.LastReset = time.Now()
	monitor.LastUpdated = time.Now()
	monitor.UsageHistory = make([]BandwidthUsage, 0)

	bm.saveMonitors()
	bm.updateStats()

	return nil
}

// GetMonitor returns bandwidth monitor for a user
func (bm *BandwidthManager) GetMonitor(userID string) (*BandwidthMonitor, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	monitor, exists := bm.monitors[userID]
	if !exists {
		return nil, fmt.Errorf("monitor for user %s not found", userID)
	}

	// Return a copy
	monitorCopy := *monitor
	return &monitorCopy, nil
}

// ListMonitors returns all bandwidth monitors
func (bm *BandwidthManager) ListMonitors() []*BandwidthMonitor {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var monitors []*BandwidthMonitor
	for _, monitor := range bm.monitors {
		// Return a copy
		monitorCopy := *monitor
		monitors = append(monitors, &monitorCopy)
	}

	return monitors
}

// GetStats returns bandwidth statistics
func (bm *BandwidthManager) GetStats() *BandwidthStats {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Update current stats
	bm.updateStats()

	// Return a copy
	stats := *bm.stats
	return &stats
}

// UpdateConfig updates bandwidth manager configuration
func (bm *BandwidthManager) UpdateConfig(config *BandwidthConfig) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.config = config
	bm.saveConfig()

	return nil
}

// GetConfig returns current bandwidth manager configuration
func (bm *BandwidthManager) GetConfig() *BandwidthConfig {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Return a copy
	config := *bm.config
	return &config
}

// Utility methods
func (bm *BandwidthManager) isWithinAllowedTime(policy *BandwidthPolicy) bool {
	now := time.Now()

	// Check allowed hours
	if len(policy.AllowedHours) > 0 {
		hour := now.Hour()
		allowed := false
		for _, allowedHour := range policy.AllowedHours {
			if hour == allowedHour {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// Check allowed days (0 = Sunday, 1 = Monday, etc.)
	if len(policy.AllowedDays) > 0 {
		day := int(now.Weekday())
		allowed := false
		for _, allowedDay := range policy.AllowedDays {
			if day == allowedDay {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	return true
}

func (bm *BandwidthManager) getMessage(allowed, throttled bool) string {
	if !allowed {
		return "bandwidth limit exceeded"
	}
	if throttled {
		return "bandwidth throttled"
	}
	return "bandwidth available"
}

func (bm *BandwidthManager) updateStats() {
	bm.stats.TotalPolicies = len(bm.policies)
	bm.stats.ActivePolicies = 0
	bm.stats.TotalMonitors = len(bm.monitors)
	bm.stats.TotalBandwidth = 0
	bm.stats.UsedBandwidth = 0
	bm.stats.ThrottledUsers = 0
	bm.stats.AlertedUsers = 0

	for _, policy := range bm.policies {
		if policy.IsActive {
			bm.stats.ActivePolicies++
			bm.stats.TotalBandwidth += policy.MaxBandwidth
		}
	}

	for _, monitor := range bm.monitors {
		bm.stats.UsedBandwidth += monitor.CurrentUsage

		// Check if user is throttled or alerted
		policy, err := bm.GetUserPolicy(monitor.UserID)
		if err == nil {
			utilization := float64(monitor.CurrentUsage) / float64(policy.MaxBandwidth) * 100
			if utilization > bm.config.ThrottleThreshold*100 {
				bm.stats.ThrottledUsers++
			}
			if utilization > bm.config.AlertThreshold*100 {
				bm.stats.AlertedUsers++
			}
		}
	}

	bm.stats.AvailableBandwidth = bm.stats.TotalBandwidth - bm.stats.UsedBandwidth
	if bm.stats.TotalBandwidth > 0 {
		bm.stats.UtilizationRate = float64(bm.stats.UsedBandwidth) / float64(bm.stats.TotalBandwidth) * 100
	}

	bm.stats.LastUpdated = time.Now()
}

// Monitoring routine
func (bm *BandwidthManager) startMonitoringRoutine() {
	ticker := time.NewTicker(bm.config.MonitoringInterval)
	defer ticker.Stop()

	for range ticker.C {
		bm.performMonitoring()
	}
}

func (bm *BandwidthManager) performMonitoring() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Clean up old usage history
	cutoff := time.Now().Add(-bm.config.HistoryRetention)
	for _, monitor := range bm.monitors {
		var newHistory []BandwidthUsage
		for _, usage := range monitor.UsageHistory {
			if usage.Timestamp.After(cutoff) {
				newHistory = append(newHistory, usage)
			}
		}
		monitor.UsageHistory = newHistory
	}

	bm.saveMonitors()
	bm.saveStats()
}

// Configuration management
func getDefaultBandwidthConfig() *BandwidthConfig {
	return &BandwidthConfig{
		MonitoringInterval: 1 * time.Minute,
		HistoryRetention:   24 * time.Hour,
		DefaultPolicy:      "default",
		EnableThrottling:   true,
		ThrottleThreshold:  0.8, // 80%
		AlertThreshold:     0.9, // 90%
	}
}

func (bm *BandwidthManager) loadConfig() error {
	configFile := filepath.Join(bm.configDir, "bandwidth.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil // Use default config
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config BandwidthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	bm.config = &config
	return nil
}

func (bm *BandwidthManager) saveConfig() error {
	configFile := filepath.Join(bm.configDir, "bandwidth.json")

	data, err := json.MarshalIndent(bm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configFile, data, 0644)
}

func (bm *BandwidthManager) loadPolicies() error {
	policiesFile := filepath.Join(bm.configDir, "bandwidth_policies.json")
	if _, err := os.Stat(policiesFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty policies
	}

	data, err := os.ReadFile(policiesFile)
	if err != nil {
		return fmt.Errorf("failed to read policies file: %w", err)
	}

	var policies map[string]*BandwidthPolicy
	if err := json.Unmarshal(data, &policies); err != nil {
		return fmt.Errorf("failed to unmarshal policies: %w", err)
	}

	bm.policies = policies
	return nil
}

func (bm *BandwidthManager) savePolicies() error {
	policiesFile := filepath.Join(bm.configDir, "bandwidth_policies.json")

	data, err := json.MarshalIndent(bm.policies, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policies: %w", err)
	}

	return os.WriteFile(policiesFile, data, 0644)
}

func (bm *BandwidthManager) loadMonitors() error {
	monitorsFile := filepath.Join(bm.configDir, "bandwidth_monitors.json")
	if _, err := os.Stat(monitorsFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty monitors
	}

	data, err := os.ReadFile(monitorsFile)
	if err != nil {
		return fmt.Errorf("failed to read monitors file: %w", err)
	}

	var monitors map[string]*BandwidthMonitor
	if err := json.Unmarshal(data, &monitors); err != nil {
		return fmt.Errorf("failed to unmarshal monitors: %w", err)
	}

	bm.monitors = monitors
	return nil
}

func (bm *BandwidthManager) saveMonitors() error {
	monitorsFile := filepath.Join(bm.configDir, "bandwidth_monitors.json")

	data, err := json.MarshalIndent(bm.monitors, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal monitors: %w", err)
	}

	return os.WriteFile(monitorsFile, data, 0644)
}

func (bm *BandwidthManager) loadStats() error {
	statsFile := filepath.Join(bm.configDir, "bandwidth_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil // Use default stats
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats BandwidthStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	bm.stats = &stats
	return nil
}

func (bm *BandwidthManager) saveStats() error {
	statsFile := filepath.Join(bm.configDir, "bandwidth_stats.json")

	data, err := json.MarshalIndent(bm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
