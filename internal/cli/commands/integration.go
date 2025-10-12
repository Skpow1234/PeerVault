package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/integration"
)

// IntegrationCommand handles integration and automation operations
type IntegrationCommand struct {
	BaseCommand
	webhookManager     *integration.WebhookManager
	workflowManager    *integration.WorkflowManager
	integrationManager *integration.IntegrationManager
}

// NewIntegrationCommand creates a new integration command
func NewIntegrationCommand(client *client.Client, formatter *formatter.Formatter, webhookManager *integration.WebhookManager, workflowManager *integration.WorkflowManager, integrationManager *integration.IntegrationManager) *IntegrationCommand {
	return &IntegrationCommand{
		BaseCommand: BaseCommand{
			name:        "integration",
			description: "Integration and automation operations",
			usage:       "integration [command] [options]",
			client:      client,
			formatter:   formatter,
		},
		webhookManager:     webhookManager,
		workflowManager:    workflowManager,
		integrationManager: integrationManager,
	}
}

// Execute executes the integration command
func (c *IntegrationCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp()
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "webhook":
		return c.handleWebhookCommand(ctx, subArgs)
	case "workflow":
		return c.handleWorkflowCommand(ctx, subArgs)
	case "api", "gateway":
		return c.handleAPIGatewayCommand(ctx, subArgs)
	case "sync":
		return c.handleSyncCommand(ctx, subArgs)
	case "stats":
		return c.handleStatsCommand(ctx, subArgs)
	case "help":
		return c.showHelp()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// Webhook commands
func (c *IntegrationCommand) handleWebhookCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: integration webhook [create|list|get|trigger|status] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createWebhook(ctx, subArgs)
	case "list":
		return c.listWebhooks(ctx, subArgs)
	case "get":
		return c.getWebhook(ctx, subArgs)
	case "trigger":
		return c.triggerWebhook(ctx, subArgs)
	case "status":
		return c.updateWebhookStatus(ctx, subArgs)
	default:
		return fmt.Errorf("unknown webhook subcommand: %s", subcommand)
	}
}

// Workflow commands
func (c *IntegrationCommand) handleWorkflowCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: integration workflow [create|list|get|execute|status] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createWorkflow(ctx, subArgs)
	case "list":
		return c.listWorkflows(ctx, subArgs)
	case "get":
		return c.getWorkflow(ctx, subArgs)
	case "execute":
		return c.executeWorkflow(ctx, subArgs)
	case "status":
		return c.updateWorkflowStatus(ctx, subArgs)
	default:
		return fmt.Errorf("unknown workflow subcommand: %s", subcommand)
	}
}

// API Gateway commands
func (c *IntegrationCommand) handleAPIGatewayCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: integration api [create|list|get|request] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createAPIGateway(ctx, subArgs)
	case "list":
		return c.listGateways(ctx, subArgs)
	case "get":
		return c.getGateway(ctx, subArgs)
	case "request":
		return c.processRequest(ctx, subArgs)
	default:
		return fmt.Errorf("unknown API gateway subcommand: %s", subcommand)
	}
}

// Sync commands
func (c *IntegrationCommand) handleSyncCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: integration sync [integration_id]")
	}

	integrationID := args[0]
	return c.syncIntegration(ctx, integrationID)
}

// Stats command
func (c *IntegrationCommand) handleStatsCommand(ctx context.Context, _ []string) error {
	return c.getIntegrationStats(ctx)
}

// Webhook operations
func (c *IntegrationCommand) createWebhook(ctx context.Context, args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: integration webhook create <name> <description> <url> <method> <created_by> <events...>")
	}

	name := args[0]
	description := args[1]
	url := args[2]
	method := args[3]
	createdBy := args[4]
	events := args[5:]

	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "PeerVault-Webhook/1.0",
	}

	retryCount := 3
	timeout := 30 * time.Second

	webhook, err := c.webhookManager.CreateWebhook(ctx, name, description, url, method, createdBy, events, headers, retryCount, timeout)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Webhook created successfully: %s", webhook.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", webhook.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  URL: %s", webhook.URL))
	c.formatter.PrintInfo(fmt.Sprintf("  Method: %s", webhook.Method))
	c.formatter.PrintInfo(fmt.Sprintf("  Events: %v", webhook.Events))
	c.formatter.PrintInfo(fmt.Sprintf("  Secret: %s", webhook.Secret))
	c.formatter.PrintInfo(fmt.Sprintf("  Retry Count: %d", webhook.RetryCount))

	return nil
}

func (c *IntegrationCommand) listWebhooks(ctx context.Context, _ []string) error {
	webhooks, err := c.webhookManager.ListWebhooks(ctx)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %v", err)
	}

	if len(webhooks) == 0 {
		c.formatter.PrintInfo("No webhooks found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d webhooks:", len(webhooks)))
	for _, webhook := range webhooks {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s",
			webhook.Name, webhook.ID[:8], webhook.URL))
		c.formatter.PrintInfo(fmt.Sprintf("    Method: %s", webhook.Method))
		c.formatter.PrintInfo(fmt.Sprintf("    Events: %v", webhook.Events))
		c.formatter.PrintInfo(fmt.Sprintf("    Active: %v", webhook.IsActive))
		c.formatter.PrintInfo(fmt.Sprintf("    Success: %d, Failures: %d", webhook.SuccessCount, webhook.FailureCount))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", webhook.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *IntegrationCommand) getWebhook(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: integration webhook get <webhook_id>")
	}

	webhookID := args[0]
	webhook, err := c.webhookManager.GetWebhook(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Webhook Details: %s", webhook.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  ID: %s", webhook.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Description: %s", webhook.Description))
	c.formatter.PrintInfo(fmt.Sprintf("  URL: %s", webhook.URL))
	c.formatter.PrintInfo(fmt.Sprintf("  Method: %s", webhook.Method))
	c.formatter.PrintInfo(fmt.Sprintf("  Events: %v", webhook.Events))
	c.formatter.PrintInfo(fmt.Sprintf("  Secret: %s", webhook.Secret))
	c.formatter.PrintInfo(fmt.Sprintf("  Retry Count: %d", webhook.RetryCount))
	c.formatter.PrintInfo(fmt.Sprintf("  Timeout: %v", webhook.Timeout))
	c.formatter.PrintInfo(fmt.Sprintf("  Active: %v", webhook.IsActive))
	c.formatter.PrintInfo(fmt.Sprintf("  Success Count: %d", webhook.SuccessCount))
	c.formatter.PrintInfo(fmt.Sprintf("  Failure Count: %d", webhook.FailureCount))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", webhook.CreatedAt.Format(time.RFC3339)))

	return nil
}

func (c *IntegrationCommand) triggerWebhook(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: integration webhook trigger <webhook_id> <event_type> <payload>")
	}

	webhookID := args[0]
	eventType := args[1]
	payloadStr := args[2]

	payload := map[string]interface{}{
		"event":     eventType,
		"data":      payloadStr,
		"timestamp": time.Now().Unix(),
	}

	event, err := c.webhookManager.TriggerWebhook(ctx, webhookID, eventType, payload)
	if err != nil {
		return fmt.Errorf("failed to trigger webhook: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Webhook triggered successfully: %s", event.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Webhook: %s", event.WebhookID))
	c.formatter.PrintInfo(fmt.Sprintf("  Event Type: %s", event.EventType))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", event.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", event.CreatedAt.Format(time.RFC3339)))

	return nil
}

func (c *IntegrationCommand) updateWebhookStatus(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: integration webhook status <webhook_id> <active|inactive>")
	}

	webhookID := args[0]
	status := args[1]
	isActive := status == "active"

	err := c.webhookManager.UpdateWebhookStatus(ctx, webhookID, isActive)
	if err != nil {
		return fmt.Errorf("failed to update webhook status: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Webhook status updated successfully: %s -> %s", webhookID, status))
	return nil
}

// Workflow operations
func (c *IntegrationCommand) createWorkflow(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: integration workflow create <name> <description> <created_by>")
	}

	name := args[0]
	description := args[1]
	createdBy := args[2]

	// Create sample workflow steps
	steps := []*integration.WorkflowStep{
		{
			ID:         "step1",
			Name:       "Initialize",
			Type:       "action",
			Config:     map[string]interface{}{"action": "initialize"},
			Timeout:    30 * time.Second,
			RetryCount: 2,
			Metadata:   make(map[string]interface{}),
		},
		{
			ID:         "step2",
			Name:       "Process Data",
			Type:       "action",
			Config:     map[string]interface{}{"action": "process"},
			Timeout:    60 * time.Second,
			RetryCount: 3,
			Metadata:   make(map[string]interface{}),
		},
		{
			ID:         "step3",
			Name:       "Finalize",
			Type:       "action",
			Config:     map[string]interface{}{"action": "finalize"},
			Timeout:    15 * time.Second,
			RetryCount: 1,
			Metadata:   make(map[string]interface{}),
		},
	}

	// Create sample triggers
	triggers := []*integration.WorkflowTrigger{
		{
			ID:           "trigger1",
			Type:         "event",
			Config:       map[string]interface{}{"event": "file_uploaded"},
			IsActive:     true,
			TriggerCount: 0,
			Metadata:     make(map[string]interface{}),
		},
	}

	workflow, err := c.workflowManager.CreateWorkflow(ctx, name, description, createdBy, steps, triggers)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Workflow created successfully: %s", workflow.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", workflow.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Description: %s", workflow.Description))
	c.formatter.PrintInfo(fmt.Sprintf("  Steps: %d", len(workflow.Steps)))
	c.formatter.PrintInfo(fmt.Sprintf("  Triggers: %d", len(workflow.Triggers)))
	c.formatter.PrintInfo(fmt.Sprintf("  Active: %v", workflow.IsActive))

	return nil
}

func (c *IntegrationCommand) listWorkflows(ctx context.Context, _ []string) error {
	workflows, err := c.workflowManager.ListWorkflows(ctx)
	if err != nil {
		return fmt.Errorf("failed to list workflows: %v", err)
	}

	if len(workflows) == 0 {
		c.formatter.PrintInfo("No workflows found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d workflows:", len(workflows)))
	for _, workflow := range workflows {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s",
			workflow.Name, workflow.ID[:8], workflow.Description))
		c.formatter.PrintInfo(fmt.Sprintf("    Steps: %d", len(workflow.Steps)))
		c.formatter.PrintInfo(fmt.Sprintf("    Triggers: %d", len(workflow.Triggers)))
		c.formatter.PrintInfo(fmt.Sprintf("    Active: %v", workflow.IsActive))
		c.formatter.PrintInfo(fmt.Sprintf("    Runs: %d (Success: %d, Failed: %d)",
			workflow.RunCount, workflow.SuccessCount, workflow.FailureCount))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", workflow.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *IntegrationCommand) getWorkflow(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: integration workflow get <workflow_id>")
	}

	workflowID := args[0]
	workflow, err := c.workflowManager.GetWorkflow(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Workflow Details: %s", workflow.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  ID: %s", workflow.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Description: %s", workflow.Description))
	c.formatter.PrintInfo(fmt.Sprintf("  Steps: %d", len(workflow.Steps)))
	c.formatter.PrintInfo(fmt.Sprintf("  Triggers: %d", len(workflow.Triggers)))
	c.formatter.PrintInfo(fmt.Sprintf("  Active: %v", workflow.IsActive))
	c.formatter.PrintInfo(fmt.Sprintf("  Run Count: %d", workflow.RunCount))
	c.formatter.PrintInfo(fmt.Sprintf("  Success Count: %d", workflow.SuccessCount))
	c.formatter.PrintInfo(fmt.Sprintf("  Failure Count: %d", workflow.FailureCount))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", workflow.CreatedAt.Format(time.RFC3339)))

	return nil
}

func (c *IntegrationCommand) executeWorkflow(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: integration workflow execute <workflow_id> [variables...]")
	}

	workflowID := args[0]
	variables := make(map[string]interface{})

	// Parse variables from args
	for i := 1; i < len(args); i += 2 {
		if i+1 < len(args) {
			variables[args[i]] = args[i+1]
		}
	}

	execution, err := c.workflowManager.ExecuteWorkflow(ctx, workflowID, variables)
	if err != nil {
		return fmt.Errorf("failed to execute workflow: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Workflow execution started: %s", execution.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Workflow: %s", execution.WorkflowID))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", execution.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  Started: %s", execution.StartedAt.Format(time.RFC3339)))
	c.formatter.PrintInfo(fmt.Sprintf("  Variables: %v", execution.Variables))

	return nil
}

func (c *IntegrationCommand) updateWorkflowStatus(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: integration workflow status <workflow_id> <active|inactive>")
	}

	workflowID := args[0]
	status := args[1]
	isActive := status == "active"

	err := c.workflowManager.UpdateWorkflowStatus(ctx, workflowID, isActive)
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Workflow status updated successfully: %s -> %s", workflowID, status))
	return nil
}

// API Gateway operations
func (c *IntegrationCommand) createAPIGateway(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: integration api create <name> <description> <base_url> <created_by>")
	}

	name := args[0]
	description := args[1]
	baseURL := args[2]
	createdBy := args[3]

	// Create sample routes
	routes := []*integration.APIRoute{
		{
			ID:       "route1",
			Path:     "/api/v1/files",
			Method:   "GET",
			Handler:  "listFiles",
			Auth:     true,
			Config:   map[string]interface{}{"cache": true},
			Metadata: make(map[string]interface{}),
		},
		{
			ID:       "route2",
			Path:     "/api/v1/files",
			Method:   "POST",
			Handler:  "uploadFile",
			Auth:     true,
			Config:   map[string]interface{}{"max_size": "100MB"},
			Metadata: make(map[string]interface{}),
		},
	}

	// Create sample middleware
	middleware := []*integration.APIMiddleware{
		{
			ID:       "middleware1",
			Name:     "Authentication",
			Type:     "auth",
			Config:   map[string]interface{}{"type": "jwt"},
			Order:    1,
			IsActive: true,
			Metadata: make(map[string]interface{}),
		},
		{
			ID:       "middleware2",
			Name:     "Rate Limiting",
			Type:     "rate_limit",
			Config:   map[string]interface{}{"requests": 100, "window": "1m"},
			Order:    2,
			IsActive: true,
			Metadata: make(map[string]interface{}),
		},
	}

	// Create rate limit
	rateLimit := &integration.RateLimit{
		Requests:   1000,
		Window:     time.Minute,
		Burst:      100,
		SkipOnFail: false,
	}

	// Create auth config
	auth := &integration.APIAuth{
		Type:      "jwt",
		Config:    map[string]interface{}{"secret": "your-secret-key"},
		Required:  true,
		SkipPaths: []string{"/health", "/metrics"},
		Metadata:  make(map[string]interface{}),
	}

	gateway, err := c.integrationManager.CreateAPIGateway(ctx, name, description, baseURL, createdBy, routes, middleware, rateLimit, auth)
	if err != nil {
		return fmt.Errorf("failed to create API gateway: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("API Gateway created successfully: %s", gateway.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", gateway.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Base URL: %s", gateway.BaseURL))
	c.formatter.PrintInfo(fmt.Sprintf("  Routes: %d", len(gateway.Routes)))
	c.formatter.PrintInfo(fmt.Sprintf("  Middleware: %d", len(gateway.Middleware)))
	c.formatter.PrintInfo(fmt.Sprintf("  Rate Limit: %d requests/%v", gateway.RateLimit.Requests, gateway.RateLimit.Window))
	c.formatter.PrintInfo(fmt.Sprintf("  Auth Type: %s", gateway.Auth.Type))

	return nil
}

func (c *IntegrationCommand) listGateways(ctx context.Context, _ []string) error {
	gateways, err := c.integrationManager.ListGateways(ctx)
	if err != nil {
		return fmt.Errorf("failed to list gateways: %v", err)
	}

	if len(gateways) == 0 {
		c.formatter.PrintInfo("No API gateways found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d API gateways:", len(gateways)))
	for _, gateway := range gateways {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s",
			gateway.Name, gateway.ID[:8], gateway.BaseURL))
		c.formatter.PrintInfo(fmt.Sprintf("    Routes: %d", len(gateway.Routes)))
		c.formatter.PrintInfo(fmt.Sprintf("    Middleware: %d", len(gateway.Middleware)))
		c.formatter.PrintInfo(fmt.Sprintf("    Active: %v", gateway.IsActive))
		c.formatter.PrintInfo(fmt.Sprintf("    Requests: %d (Success: %d, Failed: %d)",
			gateway.RequestCount, gateway.SuccessCount, gateway.FailureCount))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", gateway.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *IntegrationCommand) getGateway(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: integration api get <gateway_id>")
	}

	gatewayID := args[0]
	gateway, err := c.integrationManager.GetGateway(ctx, gatewayID)
	if err != nil {
		return fmt.Errorf("failed to get gateway: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("API Gateway Details: %s", gateway.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  ID: %s", gateway.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Description: %s", gateway.Description))
	c.formatter.PrintInfo(fmt.Sprintf("  Base URL: %s", gateway.BaseURL))
	c.formatter.PrintInfo(fmt.Sprintf("  Routes: %d", len(gateway.Routes)))
	c.formatter.PrintInfo(fmt.Sprintf("  Middleware: %d", len(gateway.Middleware)))
	c.formatter.PrintInfo(fmt.Sprintf("  Rate Limit: %d requests/%v", gateway.RateLimit.Requests, gateway.RateLimit.Window))
	c.formatter.PrintInfo(fmt.Sprintf("  Auth Type: %s", gateway.Auth.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Active: %v", gateway.IsActive))
	c.formatter.PrintInfo(fmt.Sprintf("  Request Count: %d", gateway.RequestCount))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", gateway.CreatedAt.Format(time.RFC3339)))

	return nil
}

func (c *IntegrationCommand) processRequest(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: integration api request <gateway_id> <path> <method> [headers...]")
	}

	gatewayID := args[0]
	path := args[1]
	method := args[2]

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = "PeerVault-CLI/1.0"

	// Parse additional headers
	for i := 3; i < len(args); i += 2 {
		if i+1 < len(args) {
			headers[args[i]] = args[i+1]
		}
	}

	body := []byte(`{"test": "data"}`)

	response, err := c.integrationManager.ProcessRequest(ctx, gatewayID, path, method, headers, body)
	if err != nil {
		return fmt.Errorf("failed to process request: %v", err)
	}

	c.formatter.PrintSuccess("Request processed successfully")
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %v", response["status"]))
	c.formatter.PrintInfo(fmt.Sprintf("  Message: %v", response["message"]))
	c.formatter.PrintInfo(fmt.Sprintf("  Route: %v", response["route"]))
	c.formatter.PrintInfo(fmt.Sprintf("  Timestamp: %v", response["timestamp"]))

	return nil
}

// Sync operations
func (c *IntegrationCommand) syncIntegration(ctx context.Context, integrationID string) error {
	err := c.integrationManager.SyncIntegration(ctx, integrationID)
	if err != nil {
		return fmt.Errorf("failed to sync integration: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Integration synced successfully: %s", integrationID))
	return nil
}

// Stats operation
func (c *IntegrationCommand) getIntegrationStats(ctx context.Context) error {
	webhookStats, err := c.webhookManager.GetWebhookStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get webhook stats: %v", err)
	}

	workflowStats, err := c.workflowManager.GetWorkflowStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workflow stats: %v", err)
	}

	integrationStats, err := c.integrationManager.GetIntegrationStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get integration stats: %v", err)
	}

	c.formatter.PrintInfo("Integration Statistics:")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Webhooks:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Webhooks: %d", webhookStats.TotalWebhooks))
	c.formatter.PrintInfo(fmt.Sprintf("  Active Webhooks: %d", webhookStats.ActiveWebhooks))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Events: %d", webhookStats.TotalEvents))
	c.formatter.PrintInfo(fmt.Sprintf("  Successful Events: %d", webhookStats.SuccessfulEvents))
	c.formatter.PrintInfo(fmt.Sprintf("  Failed Events: %d", webhookStats.FailedEvents))
	c.formatter.PrintInfo(fmt.Sprintf("  Pending Events: %d", webhookStats.PendingEvents))
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Workflows:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Workflows: %d", workflowStats.TotalWorkflows))
	c.formatter.PrintInfo(fmt.Sprintf("  Active Workflows: %d", workflowStats.ActiveWorkflows))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Executions: %d", workflowStats.TotalExecutions))
	c.formatter.PrintInfo(fmt.Sprintf("  Successful Executions: %d", workflowStats.SuccessfulExecutions))
	c.formatter.PrintInfo(fmt.Sprintf("  Failed Executions: %d", workflowStats.FailedExecutions))
	c.formatter.PrintInfo(fmt.Sprintf("  Running Executions: %d", workflowStats.RunningExecutions))
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Third-Party Integrations:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Integrations: %d", integrationStats.TotalIntegrations))
	c.formatter.PrintInfo(fmt.Sprintf("  Active Integrations: %d", integrationStats.ActiveIntegrations))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Gateways: %d", integrationStats.TotalGateways))
	c.formatter.PrintInfo(fmt.Sprintf("  Active Gateways: %d", integrationStats.ActiveGateways))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Requests: %d", integrationStats.TotalRequests))
	c.formatter.PrintInfo(fmt.Sprintf("  Successful Requests: %d", integrationStats.SuccessfulRequests))
	c.formatter.PrintInfo(fmt.Sprintf("  Failed Requests: %d", integrationStats.FailedRequests))
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo(fmt.Sprintf("Last Updated: %s", webhookStats.LastUpdated.Format(time.RFC3339)))

	return nil
}

// Help
func (c *IntegrationCommand) showHelp() error {
	c.formatter.PrintInfo("Integration Command Help:")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Usage: integration [command] [options]")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Commands:")
	c.formatter.PrintInfo("  webhook [create|list|get|trigger|status]  - Webhook management")
	c.formatter.PrintInfo("  workflow [create|list|get|execute|status] - Workflow automation")
	c.formatter.PrintInfo("  api [create|list|get|request]             - API gateway management")
	c.formatter.PrintInfo("  sync <integration_id>                     - Sync third-party integration")
	c.formatter.PrintInfo("  stats                                     - Show integration statistics")
	c.formatter.PrintInfo("  help                                      - Show this help")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Examples:")
	c.formatter.PrintInfo("  integration webhook create 'File Upload' 'Notify on file upload' https://api.example.com/webhook POST user123 file_uploaded")
	c.formatter.PrintInfo("  integration workflow create 'Data Processing' 'Process uploaded files' user123")
	c.formatter.PrintInfo("  integration api create 'File API' 'File management API' https://api.peervault.com user123")
	c.formatter.PrintInfo("  integration api request gateway123 /api/v1/files GET")
	c.formatter.PrintInfo("  integration sync integration123")
	c.formatter.PrintInfo("  integration stats")

	return nil
}
