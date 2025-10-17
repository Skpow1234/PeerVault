package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/analytics"
	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// AnalyticsCommand handles analytics and reporting operations
type AnalyticsCommand struct {
	BaseCommand
	dashboardManager     *analytics.DashboardManager
	visualizationManager *analytics.VisualizationManager
}

// NewAnalyticsCommand creates a new analytics command
func NewAnalyticsCommand(client *client.Client, formatter *formatter.Formatter, dashboardManager *analytics.DashboardManager, visualizationManager *analytics.VisualizationManager) *AnalyticsCommand {
	return &AnalyticsCommand{
		BaseCommand: BaseCommand{
			name:        "analytics",
			description: "Analytics and reporting operations",
			usage:       "analytics [command] [options]",
			client:      client,
			formatter:   formatter,
		},
		dashboardManager:     dashboardManager,
		visualizationManager: visualizationManager,
	}
}

// Execute executes the analytics command
func (c *AnalyticsCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp()
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "dashboard":
		return c.handleDashboardCommand(ctx, subArgs)
	case "viz", "visualization":
		return c.handleVisualizationCommand(ctx, subArgs)
	case "metric":
		return c.handleMetricCommand(ctx, subArgs)
	case "report":
		return c.handleReportCommand(ctx, subArgs)
	case "ml", "model":
		return c.handleMLCommand(ctx, subArgs)
	case "alert":
		return c.handleAlertCommand(ctx, subArgs)
	case "stats":
		return c.handleStatsCommand(ctx, subArgs)
	case "help":
		return c.showHelp()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// Dashboard commands
func (c *AnalyticsCommand) handleDashboardCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics dashboard [create|list|get|add-widget] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createDashboard(ctx, subArgs)
	case "list":
		return c.listDashboards(ctx, subArgs)
	case "get":
		return c.getDashboard(ctx, subArgs)
	case "add-widget":
		return c.addWidget(ctx, subArgs)
	default:
		return fmt.Errorf("unknown dashboard subcommand: %s", subcommand)
	}
}

// Visualization commands
func (c *AnalyticsCommand) handleVisualizationCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics viz [create|list|get] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createVisualization(ctx, subArgs)
	case "list":
		return c.listVisualizations(ctx, subArgs)
	case "get":
		return c.getVisualization(ctx, subArgs)
	default:
		return fmt.Errorf("unknown visualization subcommand: %s", subcommand)
	}
}

// Metric commands
func (c *AnalyticsCommand) handleMetricCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics metric [create|update|list] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createMetric(ctx, subArgs)
	case "update":
		return c.updateMetric(ctx, subArgs)
	case "list":
		return c.listMetrics(ctx, subArgs)
	default:
		return fmt.Errorf("unknown metric subcommand: %s", subcommand)
	}
}

// Report commands
func (c *AnalyticsCommand) handleReportCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics report [create|list|get] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createReport(ctx, subArgs)
	case "list":
		return c.listReports(ctx, subArgs)
	case "get":
		return c.getReport(ctx, subArgs)
	default:
		return fmt.Errorf("unknown report subcommand: %s", subcommand)
	}
}

// ML commands
func (c *AnalyticsCommand) handleMLCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics ml [create|train|predict|list] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createMLModel(ctx, subArgs)
	case "train":
		return c.trainModel(ctx, subArgs)
	case "predict":
		return c.makePrediction(ctx, subArgs)
	case "list":
		return c.listModels(ctx, subArgs)
	default:
		return fmt.Errorf("unknown ML subcommand: %s", subcommand)
	}
}

// Alert commands
func (c *AnalyticsCommand) handleAlertCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics alert [create|list|get] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createAlert(ctx, subArgs)
	case "list":
		return c.listAlerts(ctx, subArgs)
	case "get":
		return c.getAlert(ctx, subArgs)
	default:
		return fmt.Errorf("unknown alert subcommand: %s", subcommand)
	}
}

// Stats command
func (c *AnalyticsCommand) handleStatsCommand(ctx context.Context, _ []string) error {
	return c.getAnalyticsStats(ctx)
}

// Dashboard operations
func (c *AnalyticsCommand) createDashboard(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: analytics dashboard create <name> <description> <created_by> [public] [tags...]")
	}

	name := args[0]
	description := args[1]
	createdBy := args[2]
	isPublic := len(args) > 3 && args[3] == "public"
	tags := []string{}
	if len(args) > 4 {
		tags = args[4:]
	}

	dashboard, err := c.dashboardManager.CreateDashboard(ctx, name, description, createdBy, isPublic, tags)
	if err != nil {
		return fmt.Errorf("failed to create dashboard: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Dashboard created successfully: %s", dashboard.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", dashboard.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Description: %s", dashboard.Description))
	c.formatter.PrintInfo(fmt.Sprintf("  Created By: %s", dashboard.CreatedBy))
	c.formatter.PrintInfo(fmt.Sprintf("  Public: %v", dashboard.IsPublic))
	c.formatter.PrintInfo(fmt.Sprintf("  Tags: %v", dashboard.Tags))
	c.formatter.PrintInfo(fmt.Sprintf("  Layout: %dx%d grid", dashboard.Layout.Columns, dashboard.Layout.Rows))

	return nil
}

func (c *AnalyticsCommand) listDashboards(ctx context.Context, _ []string) error {
	dashboards, err := c.dashboardManager.ListDashboards(ctx)
	if err != nil {
		return fmt.Errorf("failed to list dashboards: %v", err)
	}

	if len(dashboards) == 0 {
		c.formatter.PrintInfo("No dashboards found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d dashboards:", len(dashboards)))
	for _, dashboard := range dashboards {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s",
			dashboard.Name, dashboard.ID[:8], dashboard.Description))
		c.formatter.PrintInfo(fmt.Sprintf("    Created By: %s", dashboard.CreatedBy))
		c.formatter.PrintInfo(fmt.Sprintf("    Public: %v", dashboard.IsPublic))
		c.formatter.PrintInfo(fmt.Sprintf("    Widgets: %d", len(dashboard.Widgets)))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", dashboard.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *AnalyticsCommand) getDashboard(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics dashboard get <dashboard_id>")
	}

	dashboardID := args[0]
	dashboard, err := c.dashboardManager.GetDashboard(ctx, dashboardID)
	if err != nil {
		return fmt.Errorf("failed to get dashboard: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Dashboard Details: %s", dashboard.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  ID: %s", dashboard.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Description: %s", dashboard.Description))
	c.formatter.PrintInfo(fmt.Sprintf("  Created By: %s", dashboard.CreatedBy))
	c.formatter.PrintInfo(fmt.Sprintf("  Public: %v", dashboard.IsPublic))
	c.formatter.PrintInfo(fmt.Sprintf("  Tags: %v", dashboard.Tags))
	c.formatter.PrintInfo(fmt.Sprintf("  Layout: %dx%d grid (%s theme)",
		dashboard.Layout.Columns, dashboard.Layout.Rows, dashboard.Layout.Theme))
	c.formatter.PrintInfo(fmt.Sprintf("  Widgets: %d", len(dashboard.Widgets)))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", dashboard.CreatedAt.Format(time.RFC3339)))
	c.formatter.PrintInfo(fmt.Sprintf("  Updated: %s", dashboard.UpdatedAt.Format(time.RFC3339)))

	return nil
}

func (c *AnalyticsCommand) addWidget(ctx context.Context, args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: analytics dashboard add-widget <dashboard_id> <type> <title> <description> <x> <y> <width> <height>")
	}

	dashboardID := args[0]
	widgetType := args[1]
	title := args[2]
	description := args[3]
	x, _ := strconv.Atoi(args[4])
	y, _ := strconv.Atoi(args[5])
	width, _ := strconv.Atoi(args[6])
	height, _ := strconv.Atoi(args[7])

	config := map[string]interface{}{
		"chart_type": "line",
		"color":      "#3498db",
	}

	position := &analytics.Position{X: x, Y: y}
	size := &analytics.Size{Width: width, Height: height}

	widget, err := c.dashboardManager.AddWidget(ctx, dashboardID, widgetType, title, description, config, position, size)
	if err != nil {
		return fmt.Errorf("failed to add widget: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Widget added successfully: %s", widget.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Type: %s", widget.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Title: %s", widget.Title))
	c.formatter.PrintInfo(fmt.Sprintf("  Position: (%d, %d)", widget.Position.X, widget.Position.Y))
	c.formatter.PrintInfo(fmt.Sprintf("  Size: %dx%d", widget.Size.Width, widget.Size.Height))

	return nil
}

// Visualization operations
func (c *AnalyticsCommand) createVisualization(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: analytics viz create <name> <description> <type> <created_by> [public] [tags...]")
	}

	name := args[0]
	description := args[1]
	vizType := args[2]
	createdBy := args[3]
	isPublic := len(args) > 4 && args[4] == "public"
	tags := []string{}
	if len(args) > 5 {
		tags = args[5:]
	}

	// Create sample chart data
	chartData := &analytics.ChartData{
		Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
		Datasets: []*analytics.Dataset{
			{
				Label:           "Sales",
				Data:            []float64{12, 19, 3, 5, 2, 3},
				BackgroundColor: "#3498db",
				BorderColor:     "#2980b9",
				Fill:            false,
			},
		},
		Type: vizType,
	}

	config := map[string]interface{}{
		"responsive": true,
		"scales": map[string]interface{}{
			"y": map[string]interface{}{
				"beginAtZero": true,
			},
		},
	}

	visualization, err := c.visualizationManager.CreateVisualization(ctx, name, description, vizType, createdBy, chartData, config, isPublic, tags)
	if err != nil {
		return fmt.Errorf("failed to create visualization: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Visualization created successfully: %s", visualization.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", visualization.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Type: %s", visualization.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Created By: %s", visualization.CreatedBy))
	c.formatter.PrintInfo(fmt.Sprintf("  Public: %v", visualization.IsPublic))
	c.formatter.PrintInfo(fmt.Sprintf("  Data Points: %d", len(chartData.Labels)))

	return nil
}

func (c *AnalyticsCommand) listVisualizations(ctx context.Context, _ []string) error {
	visualizations, err := c.visualizationManager.ListVisualizations(ctx)
	if err != nil {
		return fmt.Errorf("failed to list visualizations: %v", err)
	}

	if len(visualizations) == 0 {
		c.formatter.PrintInfo("No visualizations found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d visualizations:", len(visualizations)))
	for _, viz := range visualizations {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s",
			viz.Name, viz.ID[:8], viz.Type))
		c.formatter.PrintInfo(fmt.Sprintf("    Created By: %s", viz.CreatedBy))
		c.formatter.PrintInfo(fmt.Sprintf("    Public: %v", viz.IsPublic))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", viz.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *AnalyticsCommand) getVisualization(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics viz get <visualization_id>")
	}

	vizID := args[0]
	// Note: This would need a GetVisualization method in the manager
	c.formatter.PrintInfo(fmt.Sprintf("Visualization details for: %s", vizID))
	c.formatter.PrintInfo("  (Get visualization method not implemented in manager)")

	return nil
}

// Metric operations
func (c *AnalyticsCommand) createMetric(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: analytics metric create <name> <unit> <value> [warning_threshold] [critical_threshold]")
	}

	name := args[0]
	unit := args[1]
	value, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid value: %v", err)
	}

	thresholds := make(map[string]float64)
	if len(args) > 3 {
		if warning, err := strconv.ParseFloat(args[3], 64); err == nil {
			thresholds["warning"] = warning
		}
	}
	if len(args) > 4 {
		if critical, err := strconv.ParseFloat(args[4], 64); err == nil {
			thresholds["critical"] = critical
		}
	}

	metric, err := c.dashboardManager.CreateMetric(ctx, name, unit, value, thresholds)
	if err != nil {
		return fmt.Errorf("failed to create metric: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Metric created successfully: %s", metric.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", metric.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Value: %.2f %s", metric.Value, metric.Unit))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", metric.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  Thresholds: %v", metric.Thresholds))

	return nil
}

func (c *AnalyticsCommand) updateMetric(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: analytics metric update <metric_id> <new_value>")
	}

	metricID := args[0]
	value, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid value: %v", err)
	}

	err = c.dashboardManager.UpdateMetric(ctx, metricID, value)
	if err != nil {
		return fmt.Errorf("failed to update metric: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Metric updated successfully: %s", metricID))
	c.formatter.PrintInfo(fmt.Sprintf("  New Value: %.2f", value))

	return nil
}

func (c *AnalyticsCommand) listMetrics(ctx context.Context, _ []string) error {
	// Note: This would need a ListMetrics method in the manager
	c.formatter.PrintInfo("Metrics listing not implemented in manager")
	return nil
}

// Report operations
func (c *AnalyticsCommand) createReport(ctx context.Context, args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: analytics report create <name> <description> <type> <template> <format> <created_by> [frequency] [cron]")
	}

	name := args[0]
	description := args[1]
	reportType := args[2]
	template := args[3]
	format := args[4]
	createdBy := args[5]

	frequency := "daily"
	if len(args) > 6 {
		frequency = args[6]
	}
	cron := "0 0 * * *"
	if len(args) > 7 {
		cron = args[7]
	}

	schedule := &analytics.Schedule{
		Frequency: frequency,
		Cron:      cron,
		Timezone:  "UTC",
		Enabled:   true,
	}

	recipients := []string{createdBy}

	report, err := c.dashboardManager.CreateReport(ctx, name, description, reportType, template, format, createdBy, schedule, recipients)
	if err != nil {
		return fmt.Errorf("failed to create report: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Report created successfully: %s", report.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", report.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Type: %s", report.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Format: %s", report.Format))
	c.formatter.PrintInfo(fmt.Sprintf("  Schedule: %s (%s)", report.Schedule.Frequency, report.Schedule.Cron))
	c.formatter.PrintInfo(fmt.Sprintf("  Recipients: %v", report.Recipients))

	return nil
}

func (c *AnalyticsCommand) listReports(ctx context.Context, _ []string) error {
	reports, err := c.dashboardManager.ListReports(ctx)
	if err != nil {
		return fmt.Errorf("failed to list reports: %v", err)
	}

	if len(reports) == 0 {
		c.formatter.PrintInfo("No reports found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d reports:", len(reports)))
	for _, report := range reports {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s",
			report.Name, report.ID[:8], report.Type))
		c.formatter.PrintInfo(fmt.Sprintf("    Format: %s", report.Format))
		c.formatter.PrintInfo(fmt.Sprintf("    Schedule: %s", report.Schedule.Frequency))
		c.formatter.PrintInfo(fmt.Sprintf("    Status: %s", report.Status))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", report.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *AnalyticsCommand) getReport(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics report get <report_id>")
	}

	reportID := args[0]
	c.formatter.PrintInfo(fmt.Sprintf("Report details for: %s", reportID))
	c.formatter.PrintInfo("  (Get report method not implemented in manager)")

	return nil
}

// ML operations
func (c *AnalyticsCommand) createMLModel(ctx context.Context, args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: analytics ml create <name> <description> <type> <algorithm> <version> <created_by>")
	}

	name := args[0]
	description := args[1]
	modelType := args[2]
	algorithm := args[3]
	version := args[4]
	createdBy := args[5]

	parameters := map[string]interface{}{
		"learning_rate": 0.01,
		"epochs":        100,
		"batch_size":    32,
	}

	model, err := c.visualizationManager.CreateMLModel(ctx, name, description, modelType, algorithm, version, createdBy, parameters)
	if err != nil {
		return fmt.Errorf("failed to create ML model: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("ML Model created successfully: %s", model.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", model.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Type: %s", model.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Algorithm: %s", model.Algorithm))
	c.formatter.PrintInfo(fmt.Sprintf("  Version: %s", model.Version))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", model.Status))

	return nil
}

func (c *AnalyticsCommand) trainModel(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: analytics ml train <model_id> <training_data>")
	}

	modelID := args[0]
	trainingData := args[1]

	err := c.visualizationManager.TrainModel(ctx, modelID, trainingData)
	if err != nil {
		return fmt.Errorf("failed to train model: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Model trained successfully: %s", modelID))
	c.formatter.PrintInfo("  Status: ready")
	c.formatter.PrintInfo("  Training completed")

	return nil
}

func (c *AnalyticsCommand) makePrediction(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: analytics ml predict <model_id> <input_data>")
	}

	modelID := args[0]
	inputData := args[1]

	input := map[string]interface{}{
		"data":      inputData,
		"timestamp": time.Now().Unix(),
	}

	prediction, err := c.visualizationManager.MakePrediction(ctx, modelID, input)
	if err != nil {
		return fmt.Errorf("failed to make prediction: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Prediction made successfully: %s", prediction.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Model: %s", prediction.ModelID))
	c.formatter.PrintInfo(fmt.Sprintf("  Confidence: %.2f", prediction.Confidence))
	c.formatter.PrintInfo(fmt.Sprintf("  Output: %v", prediction.Output))
	c.formatter.PrintInfo(fmt.Sprintf("  Timestamp: %s", prediction.Timestamp.Format(time.RFC3339)))

	return nil
}

func (c *AnalyticsCommand) listModels(ctx context.Context, _ []string) error {
	models, err := c.visualizationManager.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %v", err)
	}

	if len(models) == 0 {
		c.formatter.PrintInfo("No ML models found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d ML models:", len(models)))
	for _, model := range models {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s",
			model.Name, model.ID[:8], model.Type))
		c.formatter.PrintInfo(fmt.Sprintf("    Algorithm: %s", model.Algorithm))
		c.formatter.PrintInfo(fmt.Sprintf("    Status: %s", model.Status))
		c.formatter.PrintInfo(fmt.Sprintf("    Accuracy: %.2f", model.Accuracy))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", model.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

// Alert operations
func (c *AnalyticsCommand) createAlert(ctx context.Context, args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: analytics alert create <name> <description> <type> <severity> <created_by> <condition>")
	}

	name := args[0]
	description := args[1]
	alertType := args[2]
	severity := args[3]
	createdBy := args[4]
	condition := args[5]

	conditionMap := map[string]interface{}{
		"field":     "value",
		"operator":  ">",
		"threshold": condition,
	}

	recipients := []string{createdBy}
	channels := []string{"email"}

	alert, err := c.visualizationManager.CreateAlert(ctx, name, description, alertType, severity, createdBy, conditionMap, recipients, channels)
	if err != nil {
		return fmt.Errorf("failed to create alert: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Alert created successfully: %s", alert.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", alert.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Type: %s", alert.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Severity: %s", alert.Severity))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", alert.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  Recipients: %v", alert.Recipients))

	return nil
}

func (c *AnalyticsCommand) listAlerts(ctx context.Context, _ []string) error {
	// Note: This would need a ListAlerts method in the manager
	c.formatter.PrintInfo("Alerts listing not implemented in manager")
	return nil
}

func (c *AnalyticsCommand) getAlert(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: analytics alert get <alert_id>")
	}

	alertID := args[0]
	c.formatter.PrintInfo(fmt.Sprintf("Alert details for: %s", alertID))
	c.formatter.PrintInfo("  (Get alert method not implemented in manager)")

	return nil
}

// Stats operation
func (c *AnalyticsCommand) getAnalyticsStats(ctx context.Context) error {
	dashboardStats, err := c.dashboardManager.GetAnalyticsStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dashboard stats: %v", err)
	}

	visualizationStats, err := c.visualizationManager.GetVisualizationStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get visualization stats: %v", err)
	}

	c.formatter.PrintInfo("Analytics Statistics:")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Dashboards:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Dashboards: %d", dashboardStats.TotalDashboards))
	c.formatter.PrintInfo(fmt.Sprintf("  Active Dashboards: %d", dashboardStats.ActiveDashboards))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Widgets: %d", dashboardStats.TotalWidgets))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Metrics: %d", dashboardStats.TotalMetrics))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Reports: %d", dashboardStats.TotalReports))
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Visualizations & ML:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Visualizations: %d", visualizationStats.TotalVisualizations))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Data Sources: %d", visualizationStats.TotalDataSources))
	c.formatter.PrintInfo(fmt.Sprintf("  Total ML Models: %d", visualizationStats.TotalModels))
	c.formatter.PrintInfo(fmt.Sprintf("  Active Models: %d", visualizationStats.ActiveModels))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Predictions: %d", visualizationStats.TotalPredictions))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Alerts: %d", visualizationStats.TotalAlerts))
	c.formatter.PrintInfo(fmt.Sprintf("  Triggered Alerts: %d", visualizationStats.TriggeredAlerts))
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo(fmt.Sprintf("Last Updated: %s", dashboardStats.LastUpdated.Format(time.RFC3339)))

	return nil
}

// Help
func (c *AnalyticsCommand) showHelp() error {
	c.formatter.PrintInfo("Analytics Command Help:")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Usage: analytics [command] [options]")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Commands:")
	c.formatter.PrintInfo("  dashboard [create|list|get|add-widget]  - Dashboard management")
	c.formatter.PrintInfo("  viz [create|list|get]                  - Data visualization")
	c.formatter.PrintInfo("  metric [create|update|list]            - Metrics management")
	c.formatter.PrintInfo("  report [create|list|get]               - Report generation")
	c.formatter.PrintInfo("  ml [create|train|predict|list]         - Machine learning")
	c.formatter.PrintInfo("  alert [create|list|get]                - Alert management")
	c.formatter.PrintInfo("  stats                                  - Show analytics statistics")
	c.formatter.PrintInfo("  help                                   - Show this help")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Examples:")
	c.formatter.PrintInfo("  analytics dashboard create 'Sales Dashboard' 'Monthly sales overview' user123")
	c.formatter.PrintInfo("  analytics viz create 'Sales Chart' 'Monthly sales trend' line user123")
	c.formatter.PrintInfo("  analytics metric create 'Revenue' 'USD' 15000.50 20000 25000")
	c.formatter.PrintInfo("  analytics ml create 'Sales Predictor' 'Predicts sales' regression linear 1.0 user123")
	c.formatter.PrintInfo("  analytics alert create 'High CPU' 'CPU usage alert' threshold high user123 80")
	c.formatter.PrintInfo("  analytics stats")

	return nil
}
