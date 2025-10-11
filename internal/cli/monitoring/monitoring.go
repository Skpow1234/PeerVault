package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// Manager manages monitoring and alerting
type Manager struct {
	client    *client.Client
	formatter *formatter.Formatter
	alerts    []*Alert
	metrics   *MetricsCollector
	mu        sync.RWMutex
}

// Alert represents an alert rule
type Alert struct {
	ID            string
	Name          string
	Description   string
	Condition     AlertCondition
	Severity      AlertSeverity
	Enabled       bool
	LastTriggered time.Time
	TriggerCount  int
}

// AlertCondition represents an alert condition
type AlertCondition struct {
	Metric    string
	Operator  string
	Threshold float64
	Duration  time.Duration
}

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// MetricsCollector collects and stores metrics
type MetricsCollector struct {
	metrics map[string]*MetricSeries
	mu      sync.RWMutex
}

// MetricSeries represents a series of metric values
type MetricSeries struct {
	Name      string
	Values    []MetricValue
	MaxPoints int
}

// MetricValue represents a single metric value
type MetricValue struct {
	Timestamp time.Time
	Value     float64
	Labels    map[string]string
}

// Dashboard represents a monitoring dashboard
type Dashboard struct {
	ID          string
	Name        string
	Description string
	Widgets     []*Widget
	RefreshRate time.Duration
}

// Widget represents a dashboard widget
type Widget struct {
	ID       string
	Type     WidgetType
	Title    string
	Config   map[string]interface{}
	Position Position
	Size     Size
}

// WidgetType represents widget types
type WidgetType string

const (
	WidgetTypeGraph WidgetType = "graph"
	WidgetTypeGauge WidgetType = "gauge"
	WidgetTypeTable WidgetType = "table"
	WidgetTypeText  WidgetType = "text"
	WidgetTypeAlert WidgetType = "alert"
)

// Position represents widget position
type Position struct {
	X int
	Y int
}

// Size represents widget size
type Size struct {
	Width  int
	Height int
}

// New creates a new monitoring manager
func New(client *client.Client, formatter *formatter.Formatter) *Manager {
	return &Manager{
		client:    client,
		formatter: formatter,
		alerts:    make([]*Alert, 0),
		metrics: &MetricsCollector{
			metrics: make(map[string]*MetricSeries),
		},
	}
}

// StartMonitoring starts the monitoring system
func (m *Manager) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.formatter.PrintInfo("Starting monitoring system...")

	for {
		select {
		case <-ctx.Done():
			m.formatter.PrintInfo("Stopping monitoring system...")
			return
		case <-ticker.C:
			m.collectMetrics(ctx)
			m.checkAlerts(ctx)
		}
	}
}

// collectMetrics collects system metrics
func (m *Manager) collectMetrics(ctx context.Context) {
	// Collect system metrics
	health, err := m.client.GetHealth(ctx)
	if err != nil {
		m.formatter.PrintError(fmt.Errorf("failed to get health: %w", err))
		return
	}

	metrics, err := m.client.GetMetrics(ctx)
	if err != nil {
		m.formatter.PrintError(fmt.Errorf("failed to get metrics: %w", err))
		return
	}

	// Store metrics
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()

	// Health status
	m.addMetricValue("health_status", map[string]string{"service": "overall"},
		map[string]float64{"healthy": 1.0, "degraded": 0.5, "unhealthy": 0.0}[health.Status])

	// System metrics
	m.addMetricValue("files_stored", nil, float64(metrics.FilesStored))
	m.addMetricValue("network_traffic", nil, metrics.NetworkTraffic)
	m.addMetricValue("active_peers", nil, float64(metrics.ActivePeers))
	m.addMetricValue("storage_used", nil, float64(metrics.StorageUsed))

	// Service health
	for service, status := range health.Services {
		m.addMetricValue("service_health", map[string]string{"service": service},
			map[string]float64{"healthy": 1.0, "degraded": 0.5, "unhealthy": 0.0}[status])
	}
}

// addMetricValue adds a metric value to the series
func (m *Manager) addMetricValue(name string, labels map[string]string, value float64) {
	series, exists := m.metrics.metrics[name]
	if !exists {
		series = &MetricSeries{
			Name:      name,
			Values:    make([]MetricValue, 0),
			MaxPoints: 1000,
		}
		m.metrics.metrics[name] = series
	}

	metricValue := MetricValue{
		Timestamp: time.Now(),
		Value:     value,
		Labels:    labels,
	}

	series.Values = append(series.Values, metricValue)

	// Keep only last MaxPoints values
	if len(series.Values) > series.MaxPoints {
		series.Values = series.Values[len(series.Values)-series.MaxPoints:]
	}
}

// checkAlerts checks all alert conditions
func (m *Manager) checkAlerts(ctx context.Context) {
	m.mu.RLock()
	alerts := make([]*Alert, len(m.alerts))
	copy(alerts, m.alerts)
	m.mu.RUnlock()

	for _, alert := range alerts {
		if !alert.Enabled {
			continue
		}

		if m.evaluateAlertCondition(alert) {
			m.triggerAlert(alert)
		}
	}
}

// evaluateAlertCondition evaluates an alert condition
func (m *Manager) evaluateAlertCondition(alert *Alert) bool {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	series, exists := m.metrics.metrics[alert.Condition.Metric]
	if !exists {
		return false
	}

	// Get recent values within the duration
	cutoff := time.Now().Add(-alert.Condition.Duration)
	var recentValues []float64

	for _, value := range series.Values {
		if value.Timestamp.After(cutoff) {
			recentValues = append(recentValues, value.Value)
		}
	}

	if len(recentValues) == 0 {
		return false
	}

	// Calculate average of recent values
	var sum float64
	for _, value := range recentValues {
		sum += value
	}
	avg := sum / float64(len(recentValues))

	// Check condition
	switch alert.Condition.Operator {
	case ">":
		return avg > alert.Condition.Threshold
	case ">=":
		return avg >= alert.Condition.Threshold
	case "<":
		return avg < alert.Condition.Threshold
	case "<=":
		return avg <= alert.Condition.Threshold
	case "==":
		return avg == alert.Condition.Threshold
	case "!=":
		return avg != alert.Condition.Threshold
	default:
		return false
	}
}

// triggerAlert triggers an alert
func (m *Manager) triggerAlert(alert *Alert) {
	alert.LastTriggered = time.Now()
	alert.TriggerCount++

	// Format alert message based on severity
	var emoji string
	switch alert.Severity {
	case SeverityInfo:
		emoji = "ℹ️"
	case SeverityWarning:
		emoji = "⚠️"
	case SeverityCritical:
		emoji = "🚨"
	}

	message := fmt.Sprintf("%s ALERT [%s]: %s", emoji, alert.Severity, alert.Description)

	switch alert.Severity {
	case SeverityInfo:
		m.formatter.PrintInfo(message)
	case SeverityWarning:
		m.formatter.PrintWarning(message)
	case SeverityCritical:
		m.formatter.PrintError(fmt.Errorf("%s", message))
	}
}

// AddAlert adds a new alert
func (m *Manager) AddAlert(alert *Alert) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = append(m.alerts, alert)
}

// RemoveAlert removes an alert by ID
func (m *Manager) RemoveAlert(alertID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, alert := range m.alerts {
		if alert.ID == alertID {
			m.alerts = append(m.alerts[:i], m.alerts[i+1:]...)
			break
		}
	}
}

// GetAlerts returns all alerts
func (m *Manager) GetAlerts() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]*Alert, len(m.alerts))
	copy(alerts, m.alerts)
	return alerts
}

// GetMetrics returns metrics for a given name
func (m *Manager) GetMetrics(name string) *MetricSeries {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	if series, exists := m.metrics.metrics[name]; exists {
		// Return a copy
		values := make([]MetricValue, len(series.Values))
		copy(values, series.Values)
		return &MetricSeries{
			Name:      series.Name,
			Values:    values,
			MaxPoints: series.MaxPoints,
		}
	}
	return nil
}

// GetAllMetrics returns all metrics
func (m *Manager) GetAllMetrics() map[string]*MetricSeries {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	result := make(map[string]*MetricSeries)
	for name, series := range m.metrics.metrics {
		values := make([]MetricValue, len(series.Values))
		copy(values, series.Values)
		result[name] = &MetricSeries{
			Name:      series.Name,
			Values:    values,
			MaxPoints: series.MaxPoints,
		}
	}
	return result
}

// CreateDashboard creates a new dashboard
func (m *Manager) CreateDashboard(id, name, description string) *Dashboard {
	return &Dashboard{
		ID:          id,
		Name:        name,
		Description: description,
		Widgets:     make([]*Widget, 0),
		RefreshRate: 30 * time.Second,
	}
}

// AddWidget adds a widget to a dashboard
func (d *Dashboard) AddWidget(widget *Widget) {
	d.Widgets = append(d.Widgets, widget)
}

// GetDashboardData returns data for dashboard rendering
func (m *Manager) GetDashboardData(dashboard *Dashboard) map[string]interface{} {
	data := make(map[string]interface{})

	for _, widget := range dashboard.Widgets {
		widgetData := m.getWidgetData(widget)
		data[widget.ID] = widgetData
	}

	return data
}

// getWidgetData returns data for a specific widget
func (m *Manager) getWidgetData(widget *Widget) map[string]interface{} {
	switch widget.Type {
	case WidgetTypeGraph:
		metricName := widget.Config["metric"].(string)
		series := m.GetMetrics(metricName)
		if series != nil {
			return map[string]interface{}{
				"type":  "graph",
				"data":  series.Values,
				"title": widget.Title,
			}
		}
	case WidgetTypeGauge:
		metricName := widget.Config["metric"].(string)
		series := m.GetMetrics(metricName)
		if series != nil && len(series.Values) > 0 {
			latest := series.Values[len(series.Values)-1]
			return map[string]interface{}{
				"type":  "gauge",
				"value": latest.Value,
				"title": widget.Title,
			}
		}
	case WidgetTypeAlert:
		alerts := m.GetAlerts()
		return map[string]interface{}{
			"type":   "alert",
			"alerts": alerts,
			"title":  widget.Title,
		}
	}

	return map[string]interface{}{
		"type":  "text",
		"text":  "No data available",
		"title": widget.Title,
	}
}
