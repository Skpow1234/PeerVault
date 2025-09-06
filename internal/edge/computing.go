package edge

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// EdgeNode represents an edge computing node
type EdgeNode struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Location     *Location              `json:"location"`
	Capabilities *NodeCapabilities      `json:"capabilities"`
	Status       string                 `json:"status"`
	LastSeen     time.Time              `json:"last_seen"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Location represents a geographical location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
}

// NodeCapabilities represents the capabilities of an edge node
type NodeCapabilities struct {
	CPU       *CPUSpec     `json:"cpu"`
	Memory    *MemorySpec  `json:"memory"`
	Storage   *StorageSpec `json:"storage"`
	Network   *NetworkSpec `json:"network"`
	GPU       *GPUSpec     `json:"gpu,omitempty"`
	IoT       *IoTSpec     `json:"iot,omitempty"`
	Protocols []string     `json:"protocols"`
	Services  []string     `json:"services"`
}

// CPUSpec represents CPU specifications
type CPUSpec struct {
	Cores        int     `json:"cores"`
	Frequency    float64 `json:"frequency"` // GHz
	Architecture string  `json:"architecture"`
	Usage        float64 `json:"usage"` // Percentage
}

// MemorySpec represents memory specifications
type MemorySpec struct {
	Total     int64   `json:"total"`     // Bytes
	Available int64   `json:"available"` // Bytes
	Usage     float64 `json:"usage"`     // Percentage
}

// StorageSpec represents storage specifications
type StorageSpec struct {
	Total     int64   `json:"total"`     // Bytes
	Available int64   `json:"available"` // Bytes
	Usage     float64 `json:"usage"`     // Percentage
	Type      string  `json:"type"`      // SSD, HDD, NVMe, etc.
}

// NetworkSpec represents network specifications
type NetworkSpec struct {
	Bandwidth int64    `json:"bandwidth"` // bps
	Latency   float64  `json:"latency"`   // ms
	Protocols []string `json:"protocols"`
	IPAddress string   `json:"ip_address"`
}

// GPUSpec represents GPU specifications
type GPUSpec struct {
	Model             string  `json:"model"`
	Memory            int64   `json:"memory"` // Bytes
	Usage             float64 `json:"usage"`  // Percentage
	ComputeCapability string  `json:"compute_capability"`
}

// IoTSpec represents IoT device specifications
type IoTSpec struct {
	DeviceType string       `json:"device_type"`
	Sensors    []string     `json:"sensors"`
	Actuators  []string     `json:"actuators"`
	Protocols  []string     `json:"protocols"`
	Battery    *BatterySpec `json:"battery,omitempty"`
}

// BatterySpec represents battery specifications
type BatterySpec struct {
	Level       float64 `json:"level"`       // Percentage
	Voltage     float64 `json:"voltage"`     // Volts
	Current     float64 `json:"current"`     // Amperes
	Temperature float64 `json:"temperature"` // Celsius
}

// EdgeTask represents a task to be executed on an edge node
type EdgeTask struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Priority     int                    `json:"priority"`
	Requirements *TaskRequirements      `json:"requirements"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	Status       string                 `json:"status"`
	AssignedNode string                 `json:"assigned_node"`
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TaskRequirements represents the requirements for a task
type TaskRequirements struct {
	CPU      float64       `json:"cpu"`      // CPU cores required
	Memory   int64         `json:"memory"`   // Memory required in bytes
	Storage  int64         `json:"storage"`  // Storage required in bytes
	Network  int64         `json:"network"`  // Network bandwidth required in bps
	GPU      bool          `json:"gpu"`      // GPU required
	IoT      bool          `json:"iot"`      // IoT capabilities required
	Latency  float64       `json:"latency"`  // Maximum latency in ms
	Duration time.Duration `json:"duration"` // Estimated duration
}

// EdgeComputingManager provides edge computing functionality
type EdgeComputingManager struct {
	nodes   map[string]*EdgeNode
	tasks   map[string]*EdgeTask
	mu      sync.RWMutex
	metrics *EdgeMetrics
}

// EdgeMetrics represents edge computing metrics
type EdgeMetrics struct {
	TotalNodes          int     `json:"total_nodes"`
	ActiveNodes         int     `json:"active_nodes"`
	TotalTasks          int     `json:"total_tasks"`
	CompletedTasks      int     `json:"completed_tasks"`
	FailedTasks         int     `json:"failed_tasks"`
	AverageLatency      float64 `json:"average_latency"`
	ResourceUtilization float64 `json:"resource_utilization"`
}

// NewEdgeComputingManager creates a new edge computing manager
func NewEdgeComputingManager() *EdgeComputingManager {
	return &EdgeComputingManager{
		nodes:   make(map[string]*EdgeNode),
		tasks:   make(map[string]*EdgeTask),
		metrics: &EdgeMetrics{},
	}
}

// RegisterNode registers an edge node
func (ecm *EdgeComputingManager) RegisterNode(ctx context.Context, node *EdgeNode) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	node.Status = "active"
	node.LastSeen = time.Now()
	ecm.nodes[node.ID] = node

	ecm.updateMetrics()

	return nil
}

// UnregisterNode unregisters an edge node
func (ecm *EdgeComputingManager) UnregisterNode(ctx context.Context, nodeID string) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	if node, exists := ecm.nodes[nodeID]; exists {
		node.Status = "inactive"
		ecm.updateMetrics()
	}

	return nil
}

// GetNode retrieves an edge node by ID
func (ecm *EdgeComputingManager) GetNode(ctx context.Context, nodeID string) (*EdgeNode, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	node, exists := ecm.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	return node, nil
}

// ListNodes lists all edge nodes
func (ecm *EdgeComputingManager) ListNodes(ctx context.Context) ([]*EdgeNode, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	nodes := make([]*EdgeNode, 0, len(ecm.nodes))
	for _, node := range ecm.nodes {
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// FindNearestNodes finds the nearest nodes to a given location
func (ecm *EdgeComputingManager) FindNearestNodes(ctx context.Context, location *Location, maxDistance float64, limit int) ([]*EdgeNode, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	var candidates []*EdgeNode

	for _, node := range ecm.nodes {
		if node.Status != "active" {
			continue
		}

		distance := ecm.calculateDistance(location, node.Location)
		if distance <= maxDistance {
			candidates = append(candidates, node)
		}
	}

	// Sort by distance
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			distI := ecm.calculateDistance(location, candidates[i].Location)
			distJ := ecm.calculateDistance(location, candidates[j].Location)
			if distI > distJ {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Limit results
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}

	return candidates, nil
}

// SubmitTask submits a task for execution on edge nodes
func (ecm *EdgeComputingManager) SubmitTask(ctx context.Context, task *EdgeTask) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	task.Status = "pending"
	task.CreatedAt = time.Now()
	ecm.tasks[task.ID] = task

	// Try to assign the task to a suitable node
	err := ecm.assignTask(task)
	if err != nil {
		return fmt.Errorf("failed to assign task: %w", err)
	}

	ecm.updateMetrics()

	return nil
}

// GetTask retrieves a task by ID
func (ecm *EdgeComputingManager) GetTask(ctx context.Context, taskID string) (*EdgeTask, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	task, exists := ecm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks lists all tasks
func (ecm *EdgeComputingManager) ListTasks(ctx context.Context) ([]*EdgeTask, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	tasks := make([]*EdgeTask, 0, len(ecm.tasks))
	for _, task := range ecm.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTaskStatus updates the status of a task
func (ecm *EdgeComputingManager) UpdateTaskStatus(ctx context.Context, taskID string, status string) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	task, exists := ecm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Status = status

	switch status {
	case "running":
		now := time.Now()
		task.StartedAt = &now
	case "completed", "failed":
		now := time.Now()
		task.CompletedAt = &now
	}

	ecm.updateMetrics()

	return nil
}

// GetMetrics returns edge computing metrics
func (ecm *EdgeComputingManager) GetMetrics(ctx context.Context) (*EdgeMetrics, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	return ecm.metrics, nil
}

// OptimizeResourceAllocation optimizes resource allocation across edge nodes
func (ecm *EdgeComputingManager) OptimizeResourceAllocation(ctx context.Context) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	// Reassign pending tasks to available nodes
	for _, task := range ecm.tasks {
		if task.Status == "pending" {
			err := ecm.assignTask(task)
			if err != nil {
				// Log error but continue with other tasks
				continue
			}
		}
	}

	return nil
}

// assignTask assigns a task to a suitable node
func (ecm *EdgeComputingManager) assignTask(task *EdgeTask) error {
	// Find suitable nodes based on requirements
	var suitableNodes []*EdgeNode

	for _, node := range ecm.nodes {
		if node.Status != "active" {
			continue
		}

		if ecm.nodeMeetsRequirements(node, task.Requirements) {
			suitableNodes = append(suitableNodes, node)
		}
	}

	if len(suitableNodes) == 0 {
		return fmt.Errorf("no suitable nodes found for task %s", task.ID)
	}

	// Select the best node (simplified selection)
	bestNode := suitableNodes[0]
	for _, node := range suitableNodes[1:] {
		if ecm.calculateNodeScore(node) > ecm.calculateNodeScore(bestNode) {
			bestNode = node
		}
	}

	task.AssignedNode = bestNode.ID
	task.Status = "assigned"

	return nil
}

// nodeMeetsRequirements checks if a node meets task requirements
func (ecm *EdgeComputingManager) nodeMeetsRequirements(node *EdgeNode, requirements *TaskRequirements) bool {
	if requirements == nil {
		return true
	}

	// Check CPU requirements
	if requirements.CPU > 0 && node.Capabilities.CPU.Cores < int(requirements.CPU) {
		return false
	}

	// Check memory requirements
	if requirements.Memory > 0 && node.Capabilities.Memory.Available < requirements.Memory {
		return false
	}

	// Check storage requirements
	if requirements.Storage > 0 && node.Capabilities.Storage.Available < requirements.Storage {
		return false
	}

	// Check network requirements
	if requirements.Network > 0 && node.Capabilities.Network.Bandwidth < requirements.Network {
		return false
	}

	// Check GPU requirements
	if requirements.GPU && node.Capabilities.GPU == nil {
		return false
	}

	// Check IoT requirements
	if requirements.IoT && node.Capabilities.IoT == nil {
		return false
	}

	return true
}

// calculateNodeScore calculates a score for node selection
func (ecm *EdgeComputingManager) calculateNodeScore(node *EdgeNode) float64 {
	score := 0.0

	// CPU score (lower usage is better)
	score += (100 - node.Capabilities.CPU.Usage) * 0.3

	// Memory score (lower usage is better)
	score += (100 - node.Capabilities.Memory.Usage) * 0.3

	// Storage score (lower usage is better)
	score += (100 - node.Capabilities.Storage.Usage) * 0.2

	// Network score (lower latency is better)
	score += (100 - node.Capabilities.Network.Latency) * 0.2

	return score
}

// calculateDistance calculates the distance between two locations
func (ecm *EdgeComputingManager) calculateDistance(loc1, loc2 *Location) float64 {
	if loc1 == nil || loc2 == nil {
		return math.Inf(1)
	}

	// Haversine formula for calculating distance between two points
	const earthRadius = 6371 // km

	lat1Rad := loc1.Latitude * math.Pi / 180
	lat2Rad := loc2.Latitude * math.Pi / 180
	deltaLat := (loc2.Latitude - loc1.Latitude) * math.Pi / 180
	deltaLon := (loc2.Longitude - loc1.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// updateMetrics updates the edge computing metrics
func (ecm *EdgeComputingManager) updateMetrics() {
	ecm.metrics.TotalNodes = len(ecm.nodes)
	ecm.metrics.ActiveNodes = 0
	ecm.metrics.TotalTasks = len(ecm.tasks)
	ecm.metrics.CompletedTasks = 0
	ecm.metrics.FailedTasks = 0

	totalLatency := 0.0
	latencyCount := 0
	totalUtilization := 0.0
	utilizationCount := 0

	for _, node := range ecm.nodes {
		if node.Status == "active" {
			ecm.metrics.ActiveNodes++
		}

		if node.Capabilities.Network.Latency > 0 {
			totalLatency += node.Capabilities.Network.Latency
			latencyCount++
		}

		utilization := (node.Capabilities.CPU.Usage + node.Capabilities.Memory.Usage + node.Capabilities.Storage.Usage) / 3
		totalUtilization += utilization
		utilizationCount++
	}

	for _, task := range ecm.tasks {
		switch task.Status {
		case "completed":
			ecm.metrics.CompletedTasks++
		case "failed":
			ecm.metrics.FailedTasks++
		}
	}

	if latencyCount > 0 {
		ecm.metrics.AverageLatency = totalLatency / float64(latencyCount)
	}

	if utilizationCount > 0 {
		ecm.metrics.ResourceUtilization = totalUtilization / float64(utilizationCount)
	}
}
