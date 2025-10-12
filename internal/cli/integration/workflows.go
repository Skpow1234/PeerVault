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

// Workflow represents an automation workflow
type Workflow struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Steps        []*WorkflowStep        `json:"steps"`
	Triggers     []*WorkflowTrigger     `json:"triggers"`
	Variables    map[string]interface{} `json:"variables"`
	IsActive     bool                   `json:"is_active"`
	LastRun      *time.Time             `json:"last_run,omitempty"`
	NextRun      *time.Time             `json:"next_run,omitempty"`
	RunCount     int                    `json:"run_count"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    string                 `json:"created_by"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// WorkflowStep represents a workflow step
type WorkflowStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // action, condition, loop, delay
	Config     map[string]interface{} `json:"config"`
	OnSuccess  string                 `json:"on_success,omitempty"` // next step ID
	OnFailure  string                 `json:"on_failure,omitempty"` // next step ID
	Timeout    time.Duration          `json:"timeout"`
	RetryCount int                    `json:"retry_count"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// WorkflowTrigger represents a workflow trigger
type WorkflowTrigger struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"` // event, schedule, manual, webhook
	Config        map[string]interface{} `json:"config"`
	IsActive      bool                   `json:"is_active"`
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
	TriggerCount  int                    `json:"trigger_count"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// WorkflowExecution represents a workflow execution
type WorkflowExecution struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      string                 `json:"status"` // running, completed, failed, cancelled
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Steps       []*StepExecution       `json:"steps"`
	Variables   map[string]interface{} `json:"variables"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StepExecution represents a step execution
type StepExecution struct {
	ID          string                 `json:"id"`
	StepID      string                 `json:"step_id"`
	Status      string                 `json:"status"` // pending, running, completed, failed, skipped
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WorkflowManager manages automation workflows
type WorkflowManager struct {
	mu         sync.RWMutex
	client     *client.Client
	configDir  string
	workflows  map[string]*Workflow
	executions map[string]*WorkflowExecution
	stats      *WorkflowStats
}

// WorkflowStats represents workflow statistics
type WorkflowStats struct {
	TotalWorkflows       int       `json:"total_workflows"`
	ActiveWorkflows      int       `json:"active_workflows"`
	TotalExecutions      int       `json:"total_executions"`
	SuccessfulExecutions int       `json:"successful_executions"`
	FailedExecutions     int       `json:"failed_executions"`
	RunningExecutions    int       `json:"running_executions"`
	LastUpdated          time.Time `json:"last_updated"`
}

// NewWorkflowManager creates a new workflow manager
func NewWorkflowManager(client *client.Client, configDir string) *WorkflowManager {
	wm := &WorkflowManager{
		client:     client,
		configDir:  configDir,
		workflows:  make(map[string]*Workflow),
		executions: make(map[string]*WorkflowExecution),
		stats:      &WorkflowStats{},
	}
	_ = wm.loadWorkflows()  // Ignore error for initialization
	_ = wm.loadExecutions() // Ignore error for initialization
	_ = wm.loadStats()      // Ignore error for initialization
	return wm
}

// CreateWorkflow creates a new workflow
func (wm *WorkflowManager) CreateWorkflow(ctx context.Context, name, description, createdBy string, steps []*WorkflowStep, triggers []*WorkflowTrigger) (*Workflow, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	workflow := &Workflow{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		Steps:        steps,
		Triggers:     triggers,
		Variables:    make(map[string]interface{}),
		IsActive:     true,
		RunCount:     0,
		SuccessCount: 0,
		FailureCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    createdBy,
		Metadata:     make(map[string]interface{}),
	}

	wm.workflows[workflow.ID] = workflow

	// Simulate API call - store workflow data as JSON
	workflowData, err := json.Marshal(workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow: %v", err)
	}

	tempFilePath := filepath.Join(wm.configDir, fmt.Sprintf("workflows/%s.json", workflow.ID))
	if err := os.WriteFile(tempFilePath, workflowData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write workflow data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = wm.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store workflow: %v", err)
	}

	wm.stats.TotalWorkflows++
	wm.stats.ActiveWorkflows++
	_ = wm.saveStats()
	_ = wm.saveWorkflows()
	return workflow, nil
}

// ListWorkflows returns all workflows
func (wm *WorkflowManager) ListWorkflows(ctx context.Context) ([]*Workflow, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workflows := make([]*Workflow, 0, len(wm.workflows))
	for _, workflow := range wm.workflows {
		workflows = append(workflows, workflow)
	}
	return workflows, nil
}

// GetWorkflow returns a workflow by ID
func (wm *WorkflowManager) GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workflow, exists := wm.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}
	return workflow, nil
}

// ExecuteWorkflow executes a workflow
func (wm *WorkflowManager) ExecuteWorkflow(ctx context.Context, workflowID string, variables map[string]interface{}) (*WorkflowExecution, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	workflow, exists := wm.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	if !workflow.IsActive {
		return nil, fmt.Errorf("workflow is not active: %s", workflowID)
	}

	execution := &WorkflowExecution{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     "running",
		StartedAt:  time.Now(),
		Steps:      make([]*StepExecution, 0),
		Variables:  variables,
		Metadata:   make(map[string]interface{}),
	}

	wm.executions[execution.ID] = execution

	// Simulate workflow execution
	go wm.runWorkflow(workflow, execution)

	wm.stats.TotalExecutions++
	wm.stats.RunningExecutions++
	_ = wm.saveStats()
	_ = wm.saveExecutions()
	return execution, nil
}

// runWorkflow simulates running a workflow
func (wm *WorkflowManager) runWorkflow(workflow *Workflow, execution *WorkflowExecution) {
	// Simulate workflow execution time
	time.Sleep(500 * time.Millisecond)

	// Create step executions
	for _, step := range workflow.Steps {
		stepExecution := &StepExecution{
			ID:        uuid.New().String(),
			StepID:    step.ID,
			Status:    "completed",
			StartedAt: time.Now(),
			Output:    make(map[string]interface{}),
			Attempts:  1,
			Metadata:  make(map[string]interface{}),
		}
		now := time.Now()
		stepExecution.CompletedAt = &now
		stepExecution.Duration = 100 * time.Millisecond
		stepExecution.Output["result"] = "success"

		execution.Steps = append(execution.Steps, stepExecution)
	}

	// Update execution
	wm.mu.Lock()
	execution.Status = "completed"
	now := time.Now()
	execution.CompletedAt = &now
	execution.Duration = 500 * time.Millisecond

	// Update workflow stats
	workflow.RunCount++
	workflow.SuccessCount++
	workflow.LastRun = &now

	wm.stats.RunningExecutions--
	wm.stats.SuccessfulExecutions++
	wm.mu.Unlock()

	_ = wm.saveExecutions()
	_ = wm.saveWorkflows()
	_ = wm.saveStats()
}

// UpdateWorkflowStatus updates workflow status
func (wm *WorkflowManager) UpdateWorkflowStatus(ctx context.Context, workflowID string, isActive bool) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	workflow, exists := wm.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	workflow.IsActive = isActive
	workflow.UpdatedAt = time.Now()

	if isActive {
		wm.stats.ActiveWorkflows++
	} else {
		wm.stats.ActiveWorkflows--
	}

	_ = wm.saveWorkflows()
	_ = wm.saveStats()
	return nil
}

// ListExecutions returns all workflow executions
func (wm *WorkflowManager) ListExecutions(ctx context.Context) ([]*WorkflowExecution, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	executions := make([]*WorkflowExecution, 0, len(wm.executions))
	for _, execution := range wm.executions {
		executions = append(executions, execution)
	}
	return executions, nil
}

// GetWorkflowStats returns workflow statistics
func (wm *WorkflowManager) GetWorkflowStats(ctx context.Context) (*WorkflowStats, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// Update stats
	wm.stats.LastUpdated = time.Now()
	return wm.stats, nil
}

// File operations
func (wm *WorkflowManager) loadWorkflows() error {
	workflowsFile := filepath.Join(wm.configDir, "workflows.json")
	if _, err := os.Stat(workflowsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(workflowsFile)
	if err != nil {
		return fmt.Errorf("failed to read workflows file: %w", err)
	}

	var workflows []*Workflow
	if err := json.Unmarshal(data, &workflows); err != nil {
		return fmt.Errorf("failed to unmarshal workflows: %w", err)
	}

	for _, workflow := range workflows {
		wm.workflows[workflow.ID] = workflow
		if workflow.IsActive {
			wm.stats.ActiveWorkflows++
		}
	}
	return nil
}

func (wm *WorkflowManager) saveWorkflows() error {
	workflowsFile := filepath.Join(wm.configDir, "workflows.json")

	var workflows []*Workflow
	for _, workflow := range wm.workflows {
		workflows = append(workflows, workflow)
	}

	data, err := json.MarshalIndent(workflows, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflows: %w", err)
	}

	return os.WriteFile(workflowsFile, data, 0644)
}

func (wm *WorkflowManager) loadExecutions() error {
	executionsFile := filepath.Join(wm.configDir, "workflow_executions.json")
	if _, err := os.Stat(executionsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(executionsFile)
	if err != nil {
		return fmt.Errorf("failed to read executions file: %w", err)
	}

	var executions []*WorkflowExecution
	if err := json.Unmarshal(data, &executions); err != nil {
		return fmt.Errorf("failed to unmarshal executions: %w", err)
	}

	for _, execution := range executions {
		wm.executions[execution.ID] = execution
		if execution.Status == "running" {
			wm.stats.RunningExecutions++
		} else if execution.Status == "completed" {
			wm.stats.SuccessfulExecutions++
		} else if execution.Status == "failed" {
			wm.stats.FailedExecutions++
		}
	}
	return nil
}

func (wm *WorkflowManager) saveExecutions() error {
	executionsFile := filepath.Join(wm.configDir, "workflow_executions.json")

	var executions []*WorkflowExecution
	for _, execution := range wm.executions {
		executions = append(executions, execution)
	}

	data, err := json.MarshalIndent(executions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal executions: %w", err)
	}

	return os.WriteFile(executionsFile, data, 0644)
}

func (wm *WorkflowManager) loadStats() error {
	statsFile := filepath.Join(wm.configDir, "workflow_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats WorkflowStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	wm.stats = &stats
	return nil
}

func (wm *WorkflowManager) saveStats() error {
	statsFile := filepath.Join(wm.configDir, "workflow_stats.json")

	data, err := json.MarshalIndent(wm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
