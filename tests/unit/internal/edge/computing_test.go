package edge

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEdgeComputingManager(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.nodes)
	assert.NotNil(t, manager.tasks)
	assert.NotNil(t, manager.metrics)
	assert.Empty(t, manager.nodes)
	assert.Empty(t, manager.tasks)
}

func TestEdgeComputingManager_RegisterNode(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Test Node",
		Location: &Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			City:      "New York",
			Country:   "USA",
		},
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{
				Cores:        4,
				Frequency:    2.5,
				Architecture: "x86_64",
				Usage:        20.0,
			},
			Memory: &MemorySpec{
				Total:     8 * 1024 * 1024 * 1024, // 8GB
				Available: 6 * 1024 * 1024 * 1024, // 6GB
				Usage:     25.0,
			},
			Storage: &StorageSpec{
				Total:     100 * 1024 * 1024 * 1024, // 100GB
				Available: 80 * 1024 * 1024 * 1024,  // 80GB
				Usage:     20.0,
				Type:      "SSD",
			},
			Network: &NetworkSpec{
				Bandwidth: 1000 * 1024 * 1024, // 1Gbps
				Latency:   10.0,               // 10ms
				Protocols: []string{"TCP", "UDP"},
				IPAddress: "192.168.1.100",
			},
			Protocols: []string{"HTTP", "gRPC"},
			Services:  []string{"compute", "storage"},
		},
		Metadata: map[string]interface{}{
			"version": "1.0.0",
			"region":  "us-east-1",
		},
	}
	
	err := manager.RegisterNode(context.Background(), node)
	assert.NoError(t, err)
	
	// Verify node was registered
	registeredNode, err := manager.GetNode(context.Background(), "node-1")
	assert.NoError(t, err)
	assert.Equal(t, "active", registeredNode.Status)
	assert.NotZero(t, registeredNode.LastSeen)
	
	// Verify metrics were updated
	metrics, err := manager.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, metrics.TotalNodes)
	assert.Equal(t, 1, metrics.ActiveNodes)
}

func TestEdgeComputingManager_UnregisterNode(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Register a node first
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Test Node",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{Total: 8 * 1024 * 1024 * 1024, Usage: 25.0},
			Storage: &StorageSpec{Total: 100 * 1024 * 1024 * 1024, Usage: 20.0},
			Network: &NetworkSpec{Bandwidth: 1000 * 1024 * 1024, Latency: 10.0},
		},
	}
	
	err := manager.RegisterNode(context.Background(), node)
	require.NoError(t, err)
	
	// Unregister the node
	err = manager.UnregisterNode(context.Background(), "node-1")
	assert.NoError(t, err)
	
	// Verify node status changed
	registeredNode, err := manager.GetNode(context.Background(), "node-1")
	assert.NoError(t, err)
	assert.Equal(t, "inactive", registeredNode.Status)
	
	// Verify metrics were updated
	metrics, err := manager.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, metrics.TotalNodes)
	assert.Equal(t, 0, metrics.ActiveNodes)
}

func TestEdgeComputingManager_GetNode_NotFound(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	_, err := manager.GetNode(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")
}

func TestEdgeComputingManager_ListNodes(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Initially empty
	nodes, err := manager.ListNodes(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, nodes)
	
	// Register multiple nodes
	node1 := &EdgeNode{
		ID:   "node-1",
		Name: "Node 1",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{Total: 8 * 1024 * 1024 * 1024, Usage: 25.0},
			Storage: &StorageSpec{Total: 100 * 1024 * 1024 * 1024, Usage: 20.0},
			Network: &NetworkSpec{Bandwidth: 1000 * 1024 * 1024, Latency: 10.0},
		},
	}
	
	node2 := &EdgeNode{
		ID:   "node-2",
		Name: "Node 2",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 8, Usage: 30.0},
			Memory: &MemorySpec{Total: 16 * 1024 * 1024 * 1024, Usage: 40.0},
			Storage: &StorageSpec{Total: 200 * 1024 * 1024 * 1024, Usage: 30.0},
			Network: &NetworkSpec{Bandwidth: 2000 * 1024 * 1024, Latency: 5.0},
		},
	}
	
	err1 := manager.RegisterNode(context.Background(), node1)
	require.NoError(t, err1)
	err2 := manager.RegisterNode(context.Background(), node2)
	require.NoError(t, err2)
	
	// List nodes
	nodes, err = manager.ListNodes(context.Background())
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
}

func TestEdgeComputingManager_FindNearestNodes(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Register nodes at different locations
	node1 := &EdgeNode{
		ID:   "node-1",
		Name: "New York Node",
		Location: &Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			City:      "New York",
		},
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{Total: 8 * 1024 * 1024 * 1024, Usage: 25.0},
			Storage: &StorageSpec{Total: 100 * 1024 * 1024 * 1024, Usage: 20.0},
			Network: &NetworkSpec{Bandwidth: 1000 * 1024 * 1024, Latency: 10.0},
		},
	}
	
	node2 := &EdgeNode{
		ID:   "node-2",
		Name: "Los Angeles Node",
		Location: &Location{
			Latitude:  34.0522,
			Longitude: -118.2437,
			City:      "Los Angeles",
		},
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 8, Usage: 30.0},
			Memory: &MemorySpec{Total: 16 * 1024 * 1024 * 1024, Usage: 40.0},
			Storage: &StorageSpec{Total: 200 * 1024 * 1024 * 1024, Usage: 30.0},
			Network: &NetworkSpec{Bandwidth: 2000 * 1024 * 1024, Latency: 5.0},
		},
	}
	
	err1 := manager.RegisterNode(context.Background(), node1)
	require.NoError(t, err1)
	err2 := manager.RegisterNode(context.Background(), node2)
	require.NoError(t, err2)
	
	// Search from New York
	searchLocation := &Location{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	
	// Find nodes within 1000km
	nodes, err := manager.FindNearestNodes(context.Background(), searchLocation, 1000, 0)
	assert.NoError(t, err)
	assert.Len(t, nodes, 1) // Only New York node should be within 1000km
	
	// Find nodes within 5000km
	nodes, err = manager.FindNearestNodes(context.Background(), searchLocation, 5000, 0)
	assert.NoError(t, err)
	assert.Len(t, nodes, 2) // Both nodes should be within 5000km
	
	// Test with limit
	nodes, err = manager.FindNearestNodes(context.Background(), searchLocation, 5000, 1)
	assert.NoError(t, err)
	assert.Len(t, nodes, 1) // Should be limited to 1 result
}

func TestEdgeComputingManager_SubmitTask(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Register a node first
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Test Node",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{
				Total:     8 * 1024 * 1024 * 1024,
				Available: 6 * 1024 * 1024 * 1024,
				Usage:     25.0,
			},
			Storage: &StorageSpec{
				Total:     100 * 1024 * 1024 * 1024,
				Available: 80 * 1024 * 1024 * 1024,
				Usage:     20.0,
			},
			Network: &NetworkSpec{
				Bandwidth: 1000 * 1024 * 1024,
				Latency:   10.0,
			},
		},
	}
	
	err := manager.RegisterNode(context.Background(), node)
	require.NoError(t, err)
	
	// Submit a task
	task := &EdgeTask{
		ID:   "task-1",
		Name: "Test Task",
		Type: "compute",
		Requirements: &TaskRequirements{
			CPU:     2.0,
			Memory:  1024 * 1024 * 1024, // 1GB
			Storage: 10 * 1024 * 1024,   // 10MB
			Network: 100 * 1024 * 1024,  // 100Mbps
			Latency: 50.0,               // 50ms
			Duration: 5 * time.Minute,
		},
		Input: map[string]interface{}{
			"data": "test input",
		},
		Metadata: map[string]interface{}{
			"priority": "high",
		},
	}
	
	err = manager.SubmitTask(context.Background(), task)
	assert.NoError(t, err)
	
	// Verify task was submitted
	submittedTask, err := manager.GetTask(context.Background(), "task-1")
	assert.NoError(t, err)
	assert.Equal(t, "assigned", submittedTask.Status)
	assert.Equal(t, "node-1", submittedTask.AssignedNode)
	assert.NotZero(t, submittedTask.CreatedAt)
}

func TestEdgeComputingManager_SubmitTask_NoSuitableNode(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Register a node with limited capabilities
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Limited Node",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 1, Usage: 20.0},
			Memory: &MemorySpec{
				Total:     1 * 1024 * 1024 * 1024,
				Available: 512 * 1024 * 1024,
				Usage:     50.0,
			},
			Storage: &StorageSpec{
				Total:     10 * 1024 * 1024 * 1024,
				Available: 5 * 1024 * 1024 * 1024,
				Usage:     50.0,
			},
			Network: &NetworkSpec{
				Bandwidth: 100 * 1024 * 1024,
				Latency:   100.0,
			},
		},
	}
	
	err := manager.RegisterNode(context.Background(), node)
	require.NoError(t, err)
	
	// Submit a task with high requirements
	task := &EdgeTask{
		ID:   "task-1",
		Name: "High Resource Task",
		Type: "compute",
		Requirements: &TaskRequirements{
			CPU:     8.0,                // More than available
			Memory:  2 * 1024 * 1024 * 1024, // More than available
			Storage: 20 * 1024 * 1024 * 1024, // More than available
			Network: 1000 * 1024 * 1024, // More than available
			Latency: 10.0,               // Less than available
		},
	}
	
	err = manager.SubmitTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign task")
}

func TestEdgeComputingManager_GetTask_NotFound(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	_, err := manager.GetTask(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestEdgeComputingManager_ListTasks(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Initially empty
	tasks, err := manager.ListTasks(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, tasks)
	
	// Register a node and submit tasks
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Test Node",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{Total: 8 * 1024 * 1024 * 1024, Available: 6 * 1024 * 1024 * 1024, Usage: 25.0},
			Storage: &StorageSpec{Total: 100 * 1024 * 1024 * 1024, Available: 80 * 1024 * 1024 * 1024, Usage: 20.0},
			Network: &NetworkSpec{Bandwidth: 1000 * 1024 * 1024, Latency: 10.0},
		},
	}
	
	err = manager.RegisterNode(context.Background(), node)
	require.NoError(t, err)
	
	task1 := &EdgeTask{
		ID:   "task-1",
		Name: "Task 1",
		Requirements: &TaskRequirements{
			CPU: 1.0, Memory: 1024 * 1024 * 1024, Storage: 10 * 1024 * 1024, Network: 100 * 1024 * 1024,
		},
	}
	
	task2 := &EdgeTask{
		ID:   "task-2",
		Name: "Task 2",
		Requirements: &TaskRequirements{
			CPU: 1.0, Memory: 1024 * 1024 * 1024, Storage: 10 * 1024 * 1024, Network: 100 * 1024 * 1024,
		},
	}
	
	err = manager.SubmitTask(context.Background(), task1)
	require.NoError(t, err)
	err = manager.SubmitTask(context.Background(), task2)
	require.NoError(t, err)
	
	// List tasks
	tasks, err = manager.ListTasks(context.Background())
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestEdgeComputingManager_UpdateTaskStatus(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Register a node and submit a task
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Test Node",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{Total: 8 * 1024 * 1024 * 1024, Available: 6 * 1024 * 1024 * 1024, Usage: 25.0},
			Storage: &StorageSpec{Total: 100 * 1024 * 1024 * 1024, Available: 80 * 1024 * 1024 * 1024, Usage: 20.0},
			Network: &NetworkSpec{Bandwidth: 1000 * 1024 * 1024, Latency: 10.0},
		},
	}
	
	err := manager.RegisterNode(context.Background(), node)
	require.NoError(t, err)
	
	task := &EdgeTask{
		ID:   "task-1",
		Name: "Test Task",
		Requirements: &TaskRequirements{
			CPU: 1.0, Memory: 1024 * 1024 * 1024, Storage: 10 * 1024 * 1024, Network: 100 * 1024 * 1024,
		},
	}
	
	err = manager.SubmitTask(context.Background(), task)
	require.NoError(t, err)
	
	// Update task status to running
	err = manager.UpdateTaskStatus(context.Background(), "task-1", "running")
	assert.NoError(t, err)
	
	// Verify status was updated
	updatedTask, err := manager.GetTask(context.Background(), "task-1")
	assert.NoError(t, err)
	assert.Equal(t, "running", updatedTask.Status)
	assert.NotNil(t, updatedTask.StartedAt)
	
	// Update task status to completed
	err = manager.UpdateTaskStatus(context.Background(), "task-1", "completed")
	assert.NoError(t, err)
	
	// Verify status was updated
	updatedTask, err = manager.GetTask(context.Background(), "task-1")
	assert.NoError(t, err)
	assert.Equal(t, "completed", updatedTask.Status)
	assert.NotNil(t, updatedTask.CompletedAt)
}

func TestEdgeComputingManager_UpdateTaskStatus_NotFound(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	err := manager.UpdateTaskStatus(context.Background(), "non-existent", "running")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestEdgeComputingManager_GetMetrics(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Initially empty metrics
	metrics, err := manager.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, metrics.TotalNodes)
	assert.Equal(t, 0, metrics.ActiveNodes)
	assert.Equal(t, 0, metrics.TotalTasks)
	assert.Equal(t, 0, metrics.CompletedTasks)
	assert.Equal(t, 0, metrics.FailedTasks)
	
	// Register nodes and submit tasks
	node1 := &EdgeNode{
		ID:   "node-1",
		Name: "Node 1",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{Total: 8 * 1024 * 1024 * 1024, Available: 6 * 1024 * 1024 * 1024, Usage: 25.0},
			Storage: &StorageSpec{Total: 100 * 1024 * 1024 * 1024, Available: 80 * 1024 * 1024 * 1024, Usage: 20.0},
			Network: &NetworkSpec{Bandwidth: 1000 * 1024 * 1024, Latency: 10.0},
		},
	}
	
	node2 := &EdgeNode{
		ID:   "node-2",
		Name: "Node 2",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 8, Usage: 30.0},
			Memory: &MemorySpec{Total: 16 * 1024 * 1024 * 1024, Available: 12 * 1024 * 1024 * 1024, Usage: 25.0},
			Storage: &StorageSpec{Total: 200 * 1024 * 1024 * 1024, Available: 160 * 1024 * 1024 * 1024, Usage: 20.0},
			Network: &NetworkSpec{Bandwidth: 2000 * 1024 * 1024, Latency: 5.0},
		},
	}
	
	err = manager.RegisterNode(context.Background(), node1)
	require.NoError(t, err)
	err = manager.RegisterNode(context.Background(), node2)
	require.NoError(t, err)
	
	// Submit tasks
	task1 := &EdgeTask{
		ID:   "task-1",
		Name: "Task 1",
		Requirements: &TaskRequirements{
			CPU: 1.0, Memory: 1024 * 1024 * 1024, Storage: 10 * 1024 * 1024, Network: 100 * 1024 * 1024,
		},
	}
	
	task2 := &EdgeTask{
		ID:   "task-2",
		Name: "Task 2",
		Requirements: &TaskRequirements{
			CPU: 1.0, Memory: 1024 * 1024 * 1024, Storage: 10 * 1024 * 1024, Network: 100 * 1024 * 1024,
		},
	}
	
	err = manager.SubmitTask(context.Background(), task1)
	require.NoError(t, err)
	err = manager.SubmitTask(context.Background(), task2)
	require.NoError(t, err)
	
	// Update task statuses
	err = manager.UpdateTaskStatus(context.Background(), "task-1", "completed")
	require.NoError(t, err)
	err = manager.UpdateTaskStatus(context.Background(), "task-2", "failed")
	require.NoError(t, err)
	
	// Check metrics
	metrics, err = manager.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, metrics.TotalNodes)
	assert.Equal(t, 2, metrics.ActiveNodes)
	assert.Equal(t, 2, metrics.TotalTasks)
	assert.Equal(t, 1, metrics.CompletedTasks)
	assert.Equal(t, 1, metrics.FailedTasks)
	assert.Greater(t, metrics.AverageLatency, 0.0)
	assert.Greater(t, metrics.ResourceUtilization, 0.0)
}

func TestEdgeComputingManager_OptimizeResourceAllocation(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Register a node
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Test Node",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{Total: 8 * 1024 * 1024 * 1024, Available: 6 * 1024 * 1024 * 1024, Usage: 25.0},
			Storage: &StorageSpec{Total: 100 * 1024 * 1024 * 1024, Available: 80 * 1024 * 1024 * 1024, Usage: 20.0},
			Network: &NetworkSpec{Bandwidth: 1000 * 1024 * 1024, Latency: 10.0},
		},
	}
	
	err := manager.RegisterNode(context.Background(), node)
	require.NoError(t, err)
	
	// Submit a task that can't be assigned initially (no suitable node)
	task := &EdgeTask{
		ID:   "task-1",
		Name: "Test Task",
		Requirements: &TaskRequirements{
			CPU: 8.0, // More than available
			Memory: 10 * 1024 * 1024 * 1024, // More than available
			Storage: 200 * 1024 * 1024 * 1024, // More than available
			Network: 2000 * 1024 * 1024, // More than available
		},
	}
	
	// Manually add task to pending state
	manager.mu.Lock()
	task.Status = "pending"
	manager.tasks[task.ID] = task
	manager.mu.Unlock()
	
	// Optimize resource allocation
	err = manager.OptimizeResourceAllocation(context.Background())
	assert.NoError(t, err)
	
	// Task should still be pending (no suitable node)
	updatedTask, err := manager.GetTask(context.Background(), "task-1")
	assert.NoError(t, err)
	assert.Equal(t, "pending", updatedTask.Status)
}

func TestEdgeComputingManager_NodeMeetsRequirements(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Test Node",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Cores: 4, Usage: 20.0},
			Memory: &MemorySpec{
				Total:     8 * 1024 * 1024 * 1024,
				Available: 6 * 1024 * 1024 * 1024,
				Usage:     25.0,
			},
			Storage: &StorageSpec{
				Total:     100 * 1024 * 1024 * 1024,
				Available: 80 * 1024 * 1024 * 1024,
				Usage:     20.0,
			},
			Network: &NetworkSpec{
				Bandwidth: 1000 * 1024 * 1024,
				Latency:   10.0,
			},
		},
	}
	
	tests := []struct {
		name         string
		requirements *TaskRequirements
		expected     bool
	}{
		{
			name:         "no requirements",
			requirements: nil,
			expected:     true,
		},
		{
			name: "meets all requirements",
			requirements: &TaskRequirements{
				CPU:     2.0,
				Memory:  1024 * 1024 * 1024,
				Storage: 10 * 1024 * 1024,
				Network: 100 * 1024 * 1024,
			},
			expected: true,
		},
		{
			name: "exceeds CPU requirements",
			requirements: &TaskRequirements{
				CPU: 8.0, // More than available cores
			},
			expected: false,
		},
		{
			name: "exceeds memory requirements",
			requirements: &TaskRequirements{
				Memory: 10 * 1024 * 1024 * 1024, // More than available
			},
			expected: false,
		},
		{
			name: "exceeds storage requirements",
			requirements: &TaskRequirements{
				Storage: 200 * 1024 * 1024 * 1024, // More than available
			},
			expected: false,
		},
		{
			name: "exceeds network requirements",
			requirements: &TaskRequirements{
				Network: 2000 * 1024 * 1024, // More than available
			},
			expected: false,
		},
		{
			name: "requires GPU but node doesn't have one",
			requirements: &TaskRequirements{
				GPU: true,
			},
			expected: false,
		},
		{
			name: "requires IoT but node doesn't have one",
			requirements: &TaskRequirements{
				IoT: true,
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.nodeMeetsRequirements(node, tt.requirements)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEdgeComputingManager_CalculateNodeScore(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	node := &EdgeNode{
		ID:   "node-1",
		Name: "Test Node",
		Capabilities: &NodeCapabilities{
			CPU: &CPUSpec{Usage: 20.0},
			Memory: &MemorySpec{Usage: 25.0},
			Storage: &StorageSpec{Usage: 30.0},
			Network: &NetworkSpec{Latency: 10.0},
		},
	}
	
	score := manager.calculateNodeScore(node)
	
	// Score should be calculated based on resource usage and latency
	// Lower usage and latency should result in higher scores
	expectedScore := (100-20)*0.3 + (100-25)*0.3 + (100-30)*0.2 + (100-10)*0.2
	assert.Equal(t, expectedScore, score)
}

func TestEdgeComputingManager_CalculateDistance(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Test with nil locations
	distance := manager.calculateDistance(nil, nil)
	assert.True(t, distance == math.Inf(1))
	
	// Test with one nil location
	location1 := &Location{Latitude: 40.7128, Longitude: -74.0060}
	distance = manager.calculateDistance(location1, nil)
	assert.True(t, distance == math.Inf(1))
	
	// Test with same location
	location2 := &Location{Latitude: 40.7128, Longitude: -74.0060}
	distance = manager.calculateDistance(location1, location2)
	assert.Equal(t, 0.0, distance)
	
	// Test with different locations (New York to Los Angeles)
	location3 := &Location{Latitude: 34.0522, Longitude: -118.2437}
	distance = manager.calculateDistance(location1, location3)
	assert.Greater(t, distance, 0.0)
	assert.Less(t, distance, 5000.0) // Should be less than 5000km
}

func TestEdgeComputingManager_Concurrency(t *testing.T) {
	manager := NewEdgeComputingManager()
	
	// Test concurrent node registration
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			node := &EdgeNode{
				ID:   fmt.Sprintf("node-%d", id),
				Name: fmt.Sprintf("Node %d", id),
				Capabilities: &NodeCapabilities{
					CPU: &CPUSpec{Cores: 4, Usage: 20.0},
					Memory: &MemorySpec{Total: 8 * 1024 * 1024 * 1024, Available: 6 * 1024 * 1024 * 1024, Usage: 25.0},
					Storage: &StorageSpec{Total: 100 * 1024 * 1024 * 1024, Available: 80 * 1024 * 1024 * 1024, Usage: 20.0},
					Network: &NetworkSpec{Bandwidth: 1000 * 1024 * 1024, Latency: 10.0},
				},
			}
			
			err := manager.RegisterNode(context.Background(), node)
			assert.NoError(t, err)
			
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for goroutines")
		}
	}
	
	// Verify all nodes were registered
	nodes, err := manager.ListNodes(context.Background())
	assert.NoError(t, err)
	assert.Len(t, nodes, 10)
}
