package monitoring

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Alerter manages performance alerts
type Alerter struct {
	alerts    map[string]*PerformanceAlert
	mu        sync.RWMutex
	config    *MonitoringConfig
	logger    *slog.Logger
	collector *Collector
	callbacks []AlertCallback
}

// AlertCallback is called when an alert is triggered
type AlertCallback func(alert *PerformanceAlert)

// NewAlerter creates a new alerter
func NewAlerter(config *MonitoringConfig, collector *Collector, logger *slog.Logger) *Alerter {
	return &Alerter{
		alerts:    make(map[string]*PerformanceAlert),
		config:    config,
		logger:    logger,
		collector: collector,
		callbacks: make([]AlertCallback, 0),
	}
}

// RegisterCallback registers a callback for alert notifications
func (a *Alerter) RegisterCallback(callback AlertCallback) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.callbacks = append(a.callbacks, callback)
}

// CheckPerformance checks performance metrics and triggers alerts if needed
func (a *Alerter) CheckPerformance() {
	snapshot := a.collector.GetPerformanceSnapshot(nil)
	
	// Check overall response time
	a.checkResponseTime(snapshot.OverallResponseTime, "")
	
	// Check per-endpoint performance
	for _, ep := range snapshot.EndpointStats {
		a.checkResponseTime(ep.ResponseTime, ep.Endpoint)
		a.checkErrorRate(ep.ErrorRate, ep.Endpoint)
		a.checkSaturation(ep.Saturation, ep.Endpoint)
	}
	
	// Check system load
	a.checkSystemLoad(snapshot.SystemLoad)
	
	// Check for anomalies
	a.checkAnomalies(snapshot)
	
	// Resolve alerts that are no longer triggering
	a.resolveStaleAlerts()
}

// checkResponseTime checks if response times exceed thresholds
func (a *Alerter) checkResponseTime(stats ResponseTimeStats, endpoint string) {
	thresholds := a.config.AlertThresholds
	
	// Check P95
	if float64(stats.P95.Milliseconds()) > thresholds.ResponseTimeP95Ms {
		a.triggerAlert(PerformanceAlert{
			Type:      AlertTypeResponseTime,
			Severity:  AlertSeverityWarning,
			Endpoint:  endpoint,
			Message:   fmt.Sprintf("P95 response time %.2fms exceeds threshold %.2fms", float64(stats.P95.Milliseconds()), thresholds.ResponseTimeP95Ms),
			Value:     float64(stats.P95.Milliseconds()),
			Threshold: thresholds.ResponseTimeP95Ms,
			Metadata: map[string]string{
				"metric": "p95",
			},
		})
	}
	
	// Check P99
	if float64(stats.P99.Milliseconds()) > thresholds.ResponseTimeP99Ms {
		a.triggerAlert(PerformanceAlert{
			Type:      AlertTypeResponseTime,
			Severity:  AlertSeverityCritical,
			Endpoint:  endpoint,
			Message:   fmt.Sprintf("P99 response time %.2fms exceeds threshold %.2fms", float64(stats.P99.Milliseconds()), thresholds.ResponseTimeP99Ms),
			Value:     float64(stats.P99.Milliseconds()),
			Threshold: thresholds.ResponseTimeP99Ms,
			Metadata: map[string]string{
				"metric": "p99",
			},
		})
	}
}

// checkErrorRate checks if error rate exceeds threshold
func (a *Alerter) checkErrorRate(errorRate float64, endpoint string) {
	threshold := a.config.AlertThresholds.ErrorRatePercent
	
	if errorRate > threshold {
		severity := AlertSeverityWarning
		if errorRate > threshold*2 {
			severity = AlertSeverityCritical
		}
		
		a.triggerAlert(PerformanceAlert{
			Type:      AlertTypeErrorRate,
			Severity:  severity,
			Endpoint:  endpoint,
			Message:   fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", errorRate, threshold),
			Value:     errorRate,
			Threshold: threshold,
		})
	}
}

// checkSaturation checks if endpoint is saturated
func (a *Alerter) checkSaturation(saturation float64, endpoint string) {
	threshold := a.config.AlertThresholds.SaturationPercent
	
	if saturation > threshold {
		severity := AlertSeverityWarning
		if saturation > 95 {
			severity = AlertSeverityCritical
		}
		
		a.triggerAlert(PerformanceAlert{
			Type:      AlertTypeSaturation,
			Severity:  severity,
			Endpoint:  endpoint,
			Message:   fmt.Sprintf("Endpoint saturation %.2f%% exceeds threshold %.2f%%", saturation, threshold),
			Value:     saturation,
			Threshold: threshold,
		})
	}
}

// checkSystemLoad checks if system resources are overloaded
func (a *Alerter) checkSystemLoad(load SystemLoad) {
	// Check memory
	if load.Memory > a.config.AlertThresholds.MemoryPercent {
		a.triggerAlert(PerformanceAlert{
			Type:      AlertTypeSystemLoad,
			Severity:  AlertSeverityWarning,
			Message:   fmt.Sprintf("Memory usage %.2f%% exceeds threshold %.2f%%", load.Memory, a.config.AlertThresholds.MemoryPercent),
			Value:     load.Memory,
			Threshold: a.config.AlertThresholds.MemoryPercent,
			Metadata: map[string]string{
				"resource": "memory",
			},
		})
	}
	
	// Check goroutine leak
	if load.Goroutines > 10000 {
		a.triggerAlert(PerformanceAlert{
			Type:     AlertTypeSystemLoad,
			Severity: AlertSeverityWarning,
			Message:  fmt.Sprintf("High goroutine count: %d (possible leak)", load.Goroutines),
			Value:    float64(load.Goroutines),
			Threshold: 10000,
			Metadata: map[string]string{
				"resource": "goroutines",
			},
		})
	}
}

// checkAnomalies checks for anomalous patterns
func (a *Alerter) checkAnomalies(snapshot PerformanceSnapshot) {
	// Check for sudden throughput drop
	for _, ep := range snapshot.EndpointStats {
		if ep.Throughput.PeakRPS > 0 {
			currentRPS := ep.Throughput.RequestsPerSecond
			dropPercent := (ep.Throughput.PeakRPS - currentRPS) / ep.Throughput.PeakRPS * 100
			
			if dropPercent > a.config.AlertThresholds.ThroughputDropPercent {
				a.triggerAlert(PerformanceAlert{
					Type:      AlertTypeThroughput,
					Severity:  AlertSeverityWarning,
					Endpoint:  ep.Endpoint,
					Message:   fmt.Sprintf("Throughput dropped %.2f%% from peak", dropPercent),
					Value:     currentRPS,
					Threshold: ep.Throughput.PeakRPS * (1 - a.config.AlertThresholds.ThroughputDropPercent/100),
					Metadata: map[string]string{
						"peak_rps":     fmt.Sprintf("%.2f", ep.Throughput.PeakRPS),
						"current_rps":  fmt.Sprintf("%.2f", currentRPS),
						"drop_percent": fmt.Sprintf("%.2f", dropPercent),
					},
				})
			}
		}
	}
}

// triggerAlert triggers or updates an alert
func (a *Alerter) triggerAlert(alert PerformanceAlert) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Generate alert key
	key := a.getAlertKey(alert)
	
	// Check if alert already exists
	existing, exists := a.alerts[key]
	if exists && existing.Status == AlertStatusActive {
		// Update existing alert
		existing.Value = alert.Value
		existing.TriggeredAt = time.Now()
		return
	}
	
	// Create new alert
	alert.ID = uuid.New().String()
	alert.TriggeredAt = time.Now()
	alert.Status = AlertStatusActive
	
	a.alerts[key] = &alert
	
	// Log alert
	a.logger.Warn("Performance alert triggered",
		"type", alert.Type,
		"severity", alert.Severity,
		"endpoint", alert.Endpoint,
		"message", alert.Message,
	)
	
	// Call callbacks
	for _, callback := range a.callbacks {
		go callback(&alert)
	}
}

// resolveStaleAlerts resolves alerts that are no longer triggering
func (a *Alerter) resolveStaleAlerts() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	now := time.Now()
	staleThreshold := 5 * time.Minute
	
	for _, alert := range a.alerts {
		if alert.Status == AlertStatusActive && now.Sub(alert.TriggeredAt) > staleThreshold {
			alert.Status = AlertStatusResolved
			a.logger.Info("Performance alert resolved",
				"type", alert.Type,
				"endpoint", alert.Endpoint,
			)
		}
	}
}

// GetActiveAlerts returns all active alerts
func (a *Alerter) GetActiveAlerts() []PerformanceAlert {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	var alerts []PerformanceAlert
	for _, alert := range a.alerts {
		if alert.Status == AlertStatusActive {
			alerts = append(alerts, *alert)
		}
	}
	
	return alerts
}

// GetAllAlerts returns all alerts
func (a *Alerter) GetAllAlerts() []PerformanceAlert {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	alerts := make([]PerformanceAlert, 0, len(a.alerts))
	for _, alert := range a.alerts {
		alerts = append(alerts, *alert)
	}
	
	return alerts
}

// SilenceAlert silences a specific alert
func (a *Alerter) SilenceAlert(alertID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	for _, alert := range a.alerts {
		if alert.ID == alertID {
			alert.Status = AlertStatusSilenced
			return nil
		}
	}
	
	return fmt.Errorf("alert not found: %s", alertID)
}

// ResolveAlert manually resolves an alert
func (a *Alerter) ResolveAlert(alertID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	for _, alert := range a.alerts {
		if alert.ID == alertID {
			alert.Status = AlertStatusResolved
			return nil
		}
	}
	
	return fmt.Errorf("alert not found: %s", alertID)
}

// ClearOldAlerts removes old resolved/silenced alerts
func (a *Alerter) ClearOldAlerts(olderThan time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	cutoff := time.Now().Add(-olderThan)
	
	for key, alert := range a.alerts {
		if (alert.Status == AlertStatusResolved || alert.Status == AlertStatusSilenced) &&
			alert.TriggeredAt.Before(cutoff) {
			delete(a.alerts, key)
		}
	}
}

// getAlertKey generates a unique key for an alert
func (a *Alerter) getAlertKey(alert PerformanceAlert) string {
	return fmt.Sprintf("%s:%s:%s", alert.Type, alert.Endpoint, alert.Severity)
}

// StartMonitoring starts the continuous monitoring loop
func (a *Alerter) StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for range ticker.C {
		a.CheckPerformance()
		a.ClearOldAlerts(24 * time.Hour)
	}
}

