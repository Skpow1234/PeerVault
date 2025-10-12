package analytics

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

// Dashboard represents an analytics dashboard
type Dashboard struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Widgets     []*Widget              `json:"widgets"`
	Layout      *Layout                `json:"layout"`
	Filters     map[string]interface{} `json:"filters"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	IsPublic    bool                   `json:"is_public"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // chart, table, metric, text
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Data        map[string]interface{} `json:"data"`
	Position    *Position              `json:"position"`
	Size        *Size                  `json:"size"`
	RefreshRate time.Duration          `json:"refresh_rate"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Layout represents dashboard layout configuration
type Layout struct {
	Columns    int      `json:"columns"`
	Rows       int      `json:"rows"`
	GridSize   int      `json:"grid_size"`
	Theme      string   `json:"theme"`
	Background string   `json:"background"`
	Colors     []string `json:"colors"`
}

// Position represents widget position
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Size represents widget size
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ChartData represents chart data
type ChartData struct {
	Labels   []string               `json:"labels"`
	Datasets []*Dataset             `json:"datasets"`
	Options  map[string]interface{} `json:"options"`
	Type     string                 `json:"type"` // line, bar, pie, scatter
}

// Dataset represents chart dataset
type Dataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor string    `json:"background_color"`
	BorderColor     string    `json:"border_color"`
	Fill            bool      `json:"fill"`
}

// Metric represents a key performance indicator
type Metric struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Change      float64                `json:"change"`
	ChangeType  string                 `json:"change_type"` // increase, decrease, neutral
	Trend       []float64              `json:"trend"`
	Thresholds  map[string]float64     `json:"thresholds"`
	Status      string                 `json:"status"` // good, warning, critical
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Report represents an analytics report
type Report struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // scheduled, on-demand, real-time
	Template    string                 `json:"template"`
	Parameters  map[string]interface{} `json:"parameters"`
	Schedule    *Schedule              `json:"schedule"`
	Recipients  []string               `json:"recipients"`
	Format      string                 `json:"format"` // pdf, html, csv, json
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     *time.Time             `json:"next_run,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Schedule represents report scheduling
type Schedule struct {
	Frequency string `json:"frequency"` // daily, weekly, monthly, custom
	Cron      string `json:"cron"`
	Timezone  string `json:"timezone"`
	Enabled   bool   `json:"enabled"`
}

// DashboardManager manages analytics dashboards
type DashboardManager struct {
	mu         sync.RWMutex
	client     *client.Client
	configDir  string
	dashboards map[string]*Dashboard
	widgets    map[string]*Widget
	metrics    map[string]*Metric
	reports    map[string]*Report
	stats      *AnalyticsStats
}

// AnalyticsStats represents analytics statistics
type AnalyticsStats struct {
	TotalDashboards  int       `json:"total_dashboards"`
	TotalWidgets     int       `json:"total_widgets"`
	TotalMetrics     int       `json:"total_metrics"`
	TotalReports     int       `json:"total_reports"`
	ActiveDashboards int       `json:"active_dashboards"`
	LastUpdated      time.Time `json:"last_updated"`
}

// NewDashboardManager creates a new dashboard manager
func NewDashboardManager(client *client.Client, configDir string) *DashboardManager {
	dm := &DashboardManager{
		client:     client,
		configDir:  configDir,
		dashboards: make(map[string]*Dashboard),
		widgets:    make(map[string]*Widget),
		metrics:    make(map[string]*Metric),
		reports:    make(map[string]*Report),
		stats:      &AnalyticsStats{},
	}
	_ = dm.loadDashboards() // Ignore error for initialization
	_ = dm.loadWidgets()    // Ignore error for initialization
	_ = dm.loadMetrics()    // Ignore error for initialization
	_ = dm.loadReports()    // Ignore error for initialization
	_ = dm.loadStats()      // Ignore error for initialization
	return dm
}

// CreateDashboard creates a new analytics dashboard
func (dm *DashboardManager) CreateDashboard(ctx context.Context, name, description, createdBy string, isPublic bool, tags []string) (*Dashboard, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dashboard := &Dashboard{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Widgets:     make([]*Widget, 0),
		Layout: &Layout{
			Columns:    12,
			Rows:       8,
			GridSize:   1,
			Theme:      "default",
			Background: "#ffffff",
			Colors:     []string{"#3498db", "#e74c3c", "#2ecc71", "#f39c12", "#9b59b6"},
		},
		Filters:   make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: createdBy,
		IsPublic:  isPublic,
		Tags:      tags,
		Metadata:  make(map[string]interface{}),
	}

	dm.dashboards[dashboard.ID] = dashboard

	// Simulate API call - store dashboard data as JSON
	dashboardData, err := json.Marshal(dashboard)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dashboard: %v", err)
	}

	tempFilePath := filepath.Join(dm.configDir, fmt.Sprintf("dashboards/%s.json", dashboard.ID))
	if err := os.WriteFile(tempFilePath, dashboardData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write dashboard data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = dm.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store dashboard: %v", err)
	}

	dm.stats.TotalDashboards++
	dm.stats.ActiveDashboards++
	_ = dm.saveStats()
	_ = dm.saveDashboards()
	return dashboard, nil
}

// ListDashboards returns all dashboards
func (dm *DashboardManager) ListDashboards(ctx context.Context) ([]*Dashboard, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dashboards := make([]*Dashboard, 0, len(dm.dashboards))
	for _, dashboard := range dm.dashboards {
		dashboards = append(dashboards, dashboard)
	}
	return dashboards, nil
}

// GetDashboard returns a dashboard by ID
func (dm *DashboardManager) GetDashboard(ctx context.Context, dashboardID string) (*Dashboard, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dashboard, exists := dm.dashboards[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard not found: %s", dashboardID)
	}
	return dashboard, nil
}

// AddWidget adds a widget to a dashboard
func (dm *DashboardManager) AddWidget(ctx context.Context, dashboardID, widgetType, title, description string, config map[string]interface{}, position *Position, size *Size) (*Widget, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dashboard, exists := dm.dashboards[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard not found: %s", dashboardID)
	}

	widget := &Widget{
		ID:          uuid.New().String(),
		Type:        widgetType,
		Title:       title,
		Description: description,
		Config:      config,
		Data:        make(map[string]interface{}),
		Position:    position,
		Size:        size,
		RefreshRate: 30 * time.Second,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	dashboard.Widgets = append(dashboard.Widgets, widget)
	dm.widgets[widget.ID] = widget
	dashboard.UpdatedAt = time.Now()

	dm.stats.TotalWidgets++
	_ = dm.saveStats()
	_ = dm.saveDashboards()
	_ = dm.saveWidgets()
	return widget, nil
}

// CreateMetric creates a new metric
func (dm *DashboardManager) CreateMetric(ctx context.Context, name, unit string, value float64, thresholds map[string]float64) (*Metric, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	metric := &Metric{
		ID:          uuid.New().String(),
		Name:        name,
		Value:       value,
		Unit:        unit,
		Change:      0.0,
		ChangeType:  "neutral",
		Trend:       make([]float64, 0),
		Thresholds:  thresholds,
		Status:      "good",
		LastUpdated: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	dm.metrics[metric.ID] = metric

	// Simulate API call - store metric data as JSON
	metricData, err := json.Marshal(metric)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric: %v", err)
	}

	tempFilePath := filepath.Join(dm.configDir, fmt.Sprintf("metrics/%s.json", metric.ID))
	if err := os.WriteFile(tempFilePath, metricData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write metric data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = dm.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store metric: %v", err)
	}

	dm.stats.TotalMetrics++
	_ = dm.saveStats()
	_ = dm.saveMetrics()
	return metric, nil
}

// UpdateMetric updates a metric value
func (dm *DashboardManager) UpdateMetric(ctx context.Context, metricID string, value float64) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	metric, exists := dm.metrics[metricID]
	if !exists {
		return fmt.Errorf("metric not found: %s", metricID)
	}

	oldValue := metric.Value
	metric.Value = value
	metric.Change = value - oldValue
	metric.ChangeType = "neutral"
	if metric.Change > 0 {
		metric.ChangeType = "increase"
	} else if metric.Change < 0 {
		metric.ChangeType = "decrease"
	}

	// Update trend
	metric.Trend = append(metric.Trend, value)
	if len(metric.Trend) > 100 { // Keep last 100 values
		metric.Trend = metric.Trend[1:]
	}

	// Update status based on thresholds
	metric.Status = "good"
	if warning, exists := metric.Thresholds["warning"]; exists && value >= warning {
		metric.Status = "warning"
	}
	if critical, exists := metric.Thresholds["critical"]; exists && value >= critical {
		metric.Status = "critical"
	}

	metric.LastUpdated = time.Now()

	_ = dm.saveMetrics()
	return nil
}

// CreateReport creates a new analytics report
func (dm *DashboardManager) CreateReport(ctx context.Context, name, description, reportType, template, format, createdBy string, schedule *Schedule, recipients []string) (*Report, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	report := &Report{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Type:        reportType,
		Template:    template,
		Parameters:  make(map[string]interface{}),
		Schedule:    schedule,
		Recipients:  recipients,
		Format:      format,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Metadata:    make(map[string]interface{}),
	}

	dm.reports[report.ID] = report

	// Simulate API call - store report data as JSON
	reportData, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report: %v", err)
	}

	tempFilePath := filepath.Join(dm.configDir, fmt.Sprintf("reports/%s.json", report.ID))
	if err := os.WriteFile(tempFilePath, reportData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write report data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = dm.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store report: %v", err)
	}

	dm.stats.TotalReports++
	_ = dm.saveStats()
	_ = dm.saveReports()
	return report, nil
}

// ListReports returns all reports
func (dm *DashboardManager) ListReports(ctx context.Context) ([]*Report, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	reports := make([]*Report, 0, len(dm.reports))
	for _, report := range dm.reports {
		reports = append(reports, report)
	}
	return reports, nil
}

// GetAnalyticsStats returns analytics statistics
func (dm *DashboardManager) GetAnalyticsStats(ctx context.Context) (*AnalyticsStats, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Update stats
	dm.stats.LastUpdated = time.Now()
	return dm.stats, nil
}

// File operations
func (dm *DashboardManager) loadDashboards() error {
	dashboardsFile := filepath.Join(dm.configDir, "dashboards.json")
	if _, err := os.Stat(dashboardsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(dashboardsFile)
	if err != nil {
		return fmt.Errorf("failed to read dashboards file: %w", err)
	}

	var dashboards []*Dashboard
	if err := json.Unmarshal(data, &dashboards); err != nil {
		return fmt.Errorf("failed to unmarshal dashboards: %w", err)
	}

	for _, dashboard := range dashboards {
		dm.dashboards[dashboard.ID] = dashboard
	}
	return nil
}

func (dm *DashboardManager) saveDashboards() error {
	dashboardsFile := filepath.Join(dm.configDir, "dashboards.json")

	var dashboards []*Dashboard
	for _, dashboard := range dm.dashboards {
		dashboards = append(dashboards, dashboard)
	}

	data, err := json.MarshalIndent(dashboards, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dashboards: %w", err)
	}

	return os.WriteFile(dashboardsFile, data, 0644)
}

func (dm *DashboardManager) loadWidgets() error {
	widgetsFile := filepath.Join(dm.configDir, "widgets.json")
	if _, err := os.Stat(widgetsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(widgetsFile)
	if err != nil {
		return fmt.Errorf("failed to read widgets file: %w", err)
	}

	var widgets []*Widget
	if err := json.Unmarshal(data, &widgets); err != nil {
		return fmt.Errorf("failed to unmarshal widgets: %w", err)
	}

	for _, widget := range widgets {
		dm.widgets[widget.ID] = widget
	}
	return nil
}

func (dm *DashboardManager) saveWidgets() error {
	widgetsFile := filepath.Join(dm.configDir, "widgets.json")

	var widgets []*Widget
	for _, widget := range dm.widgets {
		widgets = append(widgets, widget)
	}

	data, err := json.MarshalIndent(widgets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal widgets: %w", err)
	}

	return os.WriteFile(widgetsFile, data, 0644)
}

func (dm *DashboardManager) loadMetrics() error {
	metricsFile := filepath.Join(dm.configDir, "metrics.json")
	if _, err := os.Stat(metricsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(metricsFile)
	if err != nil {
		return fmt.Errorf("failed to read metrics file: %w", err)
	}

	var metrics []*Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	for _, metric := range metrics {
		dm.metrics[metric.ID] = metric
	}
	return nil
}

func (dm *DashboardManager) saveMetrics() error {
	metricsFile := filepath.Join(dm.configDir, "metrics.json")

	var metrics []*Metric
	for _, metric := range dm.metrics {
		metrics = append(metrics, metric)
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return os.WriteFile(metricsFile, data, 0644)
}

func (dm *DashboardManager) loadReports() error {
	reportsFile := filepath.Join(dm.configDir, "reports.json")
	if _, err := os.Stat(reportsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(reportsFile)
	if err != nil {
		return fmt.Errorf("failed to read reports file: %w", err)
	}

	var reports []*Report
	if err := json.Unmarshal(data, &reports); err != nil {
		return fmt.Errorf("failed to unmarshal reports: %w", err)
	}

	for _, report := range reports {
		dm.reports[report.ID] = report
	}
	return nil
}

func (dm *DashboardManager) saveReports() error {
	reportsFile := filepath.Join(dm.configDir, "reports.json")

	var reports []*Report
	for _, report := range dm.reports {
		reports = append(reports, report)
	}

	data, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reports: %w", err)
	}

	return os.WriteFile(reportsFile, data, 0644)
}

func (dm *DashboardManager) loadStats() error {
	statsFile := filepath.Join(dm.configDir, "analytics_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats AnalyticsStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	dm.stats = &stats
	return nil
}

func (dm *DashboardManager) saveStats() error {
	statsFile := filepath.Join(dm.configDir, "analytics_stats.json")

	data, err := json.MarshalIndent(dm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
