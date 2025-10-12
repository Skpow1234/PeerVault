package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/edge"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// EdgeCommand handles edge computing commands
type EdgeCommand struct {
	BaseCommand
	edgeManager *edge.EdgeManager
}

// NewEdgeCommand creates a new edge command
func NewEdgeCommand(client *client.Client, formatter *formatter.Formatter, edgeManager *edge.EdgeManager) *EdgeCommand {
	return &EdgeCommand{
		BaseCommand: BaseCommand{
			name:        "edge",
			description: "Edge computing operations",
			usage:       "edge [command] [options]",
			client:      client,
			formatter:   formatter,
		},
		edgeManager: edgeManager,
	}
}

// Execute executes the edge command
func (c *EdgeCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp()
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "add-node":
		return c.addEdgeNode(ctx, subArgs)
	case "remove-node":
		return c.removeEdgeNode(ctx, subArgs)
	case "list-nodes":
		return c.listEdgeNodes(ctx, subArgs)
	case "get-node":
		return c.getEdgeNode(ctx, subArgs)
	case "update-status":
		return c.updateEdgeNodeStatus(ctx, subArgs)
	case "update-metrics":
		return c.updateEdgeNodeMetrics(ctx, subArgs)
	case "create-task":
		return c.createEdgeTask(ctx, subArgs)
	case "get-task":
		return c.getEdgeTask(ctx, subArgs)
	case "list-tasks":
		return c.listEdgeTasks(ctx, subArgs)
	case "get-tasks-by-node":
		return c.getEdgeTasksByNode(ctx, subArgs)
	case "update-task-status":
		return c.updateEdgeTaskStatus(ctx, subArgs)
	case "create-workload":
		return c.createEdgeWorkload(ctx, subArgs)
	case "get-workload":
		return c.getEdgeWorkload(ctx, subArgs)
	case "list-workloads":
		return c.listEdgeWorkloads(ctx, subArgs)
	case "schedule-workload":
		return c.scheduleWorkload(ctx, subArgs)
	case "stats":
		return c.getStatistics(ctx, subArgs)
	default:
		return c.showHelp()
	}
}

// addEdgeNode adds a new edge node
func (c *EdgeCommand) addEdgeNode(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: edge add-node <id> <name> <location> [ip] [status]")
	}

	node := &edge.EdgeNode{
		ID:             args[0],
		Name:           args[1],
		Location:       args[2],
		IPAddress:      "0.0.0.0",
		Status:         "active",
		CPUUsage:       0.0,
		MemoryUsage:    0.0,
		DiskUsage:      0.0,
		NetworkLatency: 0.0,
		Capabilities:   []string{"compute", "storage"},
		Metadata:       make(map[string]string),
	}

	if len(args) > 3 {
		node.IPAddress = args[3]
	}
	if len(args) > 4 {
		node.Status = args[4]
	}

	err := c.edgeManager.AddEdgeNode(ctx, node)
	if err != nil {
		return fmt.Errorf("failed to add edge node: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Edge node '%s' added successfully", node.ID))
	return nil
}

// removeEdgeNode removes an edge node
func (c *EdgeCommand) removeEdgeNode(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: edge remove-node <node-id>")
	}

	nodeID := args[0]
	err := c.edgeManager.RemoveEdgeNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to remove edge node: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Edge node '%s' removed successfully", nodeID))
	return nil
}

// listEdgeNodes lists all edge nodes
func (c *EdgeCommand) listEdgeNodes(ctx context.Context, _ []string) error {
	nodes, err := c.edgeManager.ListEdgeNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list edge nodes: %v", err)
	}

	if len(nodes) == 0 {
		c.formatter.PrintInfo("No edge nodes found")
		return nil
	}

	headers := []string{"ID", "Name", "Location", "IP Address", "Status", "CPU%", "Memory%", "Disk%", "Latency"}
	rows := make([][]string, len(nodes))

	for i, node := range nodes {
		rows[i] = []string{
			node.ID,
			node.Name,
			node.Location,
			node.IPAddress,
			node.Status,
			fmt.Sprintf("%.1f", node.CPUUsage),
			fmt.Sprintf("%.1f", node.MemoryUsage),
			fmt.Sprintf("%.1f", node.DiskUsage),
			fmt.Sprintf("%.1fms", node.NetworkLatency),
		}
	}

	c.formatter.PrintTable(headers, rows)
	return nil
}

// getEdgeNode gets a specific edge node
func (c *EdgeCommand) getEdgeNode(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: edge get-node <node-id>")
	}

	nodeID := args[0]
	node, err := c.edgeManager.GetEdgeNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get edge node: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Edge Node: %s", node.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", node.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Location: %s", node.Location))
	c.formatter.PrintInfo(fmt.Sprintf("  IP Address: %s", node.IPAddress))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", node.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  CPU Usage: %.1f%%", node.CPUUsage))
	c.formatter.PrintInfo(fmt.Sprintf("  Memory Usage: %.1f%%", node.MemoryUsage))
	c.formatter.PrintInfo(fmt.Sprintf("  Disk Usage: %.1f%%", node.DiskUsage))
	c.formatter.PrintInfo(fmt.Sprintf("  Network Latency: %.1fms", node.NetworkLatency))
	c.formatter.PrintInfo(fmt.Sprintf("  Capabilities: %s", strings.Join(node.Capabilities, ", ")))
	c.formatter.PrintInfo(fmt.Sprintf("  Last Seen: %s", node.LastSeen.Format("2006-01-02 15:04:05")))

	return nil
}

// updateEdgeNodeStatus updates edge node status
func (c *EdgeCommand) updateEdgeNodeStatus(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: edge update-status <node-id> <status>")
	}

	nodeID := args[0]
	status := args[1]

	err := c.edgeManager.UpdateEdgeNodeStatus(ctx, nodeID, status)
	if err != nil {
		return fmt.Errorf("failed to update edge node status: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Edge node '%s' status updated to '%s'", nodeID, status))
	return nil
}

// updateEdgeNodeMetrics updates edge node metrics
func (c *EdgeCommand) updateEdgeNodeMetrics(ctx context.Context, args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: edge update-metrics <node-id> <cpu-usage> <memory-usage> <disk-usage> <network-latency>")
	}

	nodeID := args[0]
	cpuUsage, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid CPU usage: %v", err)
	}
	memoryUsage, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid memory usage: %v", err)
	}
	diskUsage, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return fmt.Errorf("invalid disk usage: %v", err)
	}
	networkLatency, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return fmt.Errorf("invalid network latency: %v", err)
	}

	err = c.edgeManager.UpdateEdgeNodeMetrics(ctx, nodeID, cpuUsage, memoryUsage, diskUsage, networkLatency)
	if err != nil {
		return fmt.Errorf("failed to update edge node metrics: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Edge node '%s' metrics updated: CPU=%.1f%%, Memory=%.1f%%, Disk=%.1f%%, Latency=%.1fms",
		nodeID, cpuUsage, memoryUsage, diskUsage, networkLatency))
	return nil
}

// createEdgeTask creates a new edge task
func (c *EdgeCommand) createEdgeTask(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: edge create-task <id> <node-id> <name> <type> [priority]")
	}

	taskID := args[0]
	nodeID := args[1]
	name := args[2]
	taskType := args[3]
	priority := 1

	if len(args) > 4 {
		var err error
		priority, err = strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid priority: %v", err)
		}
	}

	task := &edge.EdgeTask{
		ID:          taskID,
		NodeID:      nodeID,
		Name:        name,
		Type:        taskType,
		Status:      "pending",
		Priority:    priority,
		CPUUsage:    0.0,
		MemoryUsage: 0.0,
		Input:       make(map[string]string),
		Output:      make(map[string]string),
	}

	err := c.edgeManager.CreateEdgeTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create edge task: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Edge task '%s' created successfully", taskID))
	return nil
}

// getEdgeTask gets a specific edge task
func (c *EdgeCommand) getEdgeTask(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: edge get-task <task-id>")
	}

	taskID := args[0]
	task, err := c.edgeManager.GetEdgeTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get edge task: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Edge Task: %s", task.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Node ID: %s", task.NodeID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", task.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Type: %s", task.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", task.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  Priority: %d", task.Priority))
	c.formatter.PrintInfo(fmt.Sprintf("  CPU Usage: %.1f%%", task.CPUUsage))
	c.formatter.PrintInfo(fmt.Sprintf("  Memory Usage: %.1f%%", task.MemoryUsage))
	c.formatter.PrintInfo(fmt.Sprintf("  Start Time: %s", task.StartTime.Format("2006-01-02 15:04:05")))
	if task.EndTime != nil {
		c.formatter.PrintInfo(fmt.Sprintf("  End Time: %s", task.EndTime.Format("2006-01-02 15:04:05")))
	}
	if task.Error != "" {
		c.formatter.PrintInfo(fmt.Sprintf("  Error: %s", task.Error))
	}

	return nil
}

// listEdgeTasks lists all edge tasks
func (c *EdgeCommand) listEdgeTasks(ctx context.Context, _ []string) error {
	tasks, err := c.edgeManager.ListEdgeTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to list edge tasks: %v", err)
	}

	if len(tasks) == 0 {
		c.formatter.PrintInfo("No edge tasks found")
		return nil
	}

	headers := []string{"ID", "Node ID", "Name", "Type", "Status", "Priority", "CPU%", "Memory%", "Start Time"}
	rows := make([][]string, len(tasks))

	for i, task := range tasks {
		rows[i] = []string{
			task.ID,
			task.NodeID,
			task.Name,
			task.Type,
			task.Status,
			fmt.Sprintf("%d", task.Priority),
			fmt.Sprintf("%.1f", task.CPUUsage),
			fmt.Sprintf("%.1f", task.MemoryUsage),
			task.StartTime.Format("2006-01-02 15:04:05"),
		}
	}

	c.formatter.PrintTable(headers, rows)
	return nil
}

// getEdgeTasksByNode gets tasks for a specific edge node
func (c *EdgeCommand) getEdgeTasksByNode(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: edge get-tasks-by-node <node-id>")
	}

	nodeID := args[0]
	tasks, err := c.edgeManager.GetEdgeTasksByNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get edge tasks by node: %v", err)
	}

	if len(tasks) == 0 {
		c.formatter.PrintInfo("No edge tasks found for this node")
		return nil
	}

	headers := []string{"ID", "Name", "Type", "Status", "Priority", "CPU%", "Memory%", "Start Time"}
	rows := make([][]string, len(tasks))

	for i, task := range tasks {
		rows[i] = []string{
			task.ID,
			task.Name,
			task.Type,
			task.Status,
			fmt.Sprintf("%d", task.Priority),
			fmt.Sprintf("%.1f", task.CPUUsage),
			fmt.Sprintf("%.1f", task.MemoryUsage),
			task.StartTime.Format("2006-01-02 15:04:05"),
		}
	}

	c.formatter.PrintTable(headers, rows)
	return nil
}

// updateEdgeTaskStatus updates edge task status
func (c *EdgeCommand) updateEdgeTaskStatus(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: edge update-task-status <task-id> <status>")
	}

	taskID := args[0]
	status := args[1]

	err := c.edgeManager.UpdateEdgeTaskStatus(ctx, taskID, status)
	if err != nil {
		return fmt.Errorf("failed to update edge task status: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Edge task '%s' status updated to '%s'", taskID, status))
	return nil
}

// createEdgeWorkload creates a new edge workload
func (c *EdgeCommand) createEdgeWorkload(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: edge create-workload <id> <name> <type> [priority]")
	}

	workloadID := args[0]
	name := args[1]
	workloadType := args[2]
	priority := 1

	if len(args) > 3 {
		var err error
		priority, err = strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("invalid priority: %v", err)
		}
	}

	workload := &edge.EdgeWorkload{
		ID:           workloadID,
		Name:         name,
		Type:         workloadType,
		Priority:     priority,
		Requirements: make(map[string]string),
		Tasks:        []edge.EdgeTask{},
		Status:       "pending",
	}

	err := c.edgeManager.CreateEdgeWorkload(ctx, workload)
	if err != nil {
		return fmt.Errorf("failed to create edge workload: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Edge workload '%s' created successfully", workloadID))
	return nil
}

// getEdgeWorkload gets a specific edge workload
func (c *EdgeCommand) getEdgeWorkload(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: edge get-workload <workload-id>")
	}

	workloadID := args[0]
	workload, err := c.edgeManager.GetEdgeWorkload(ctx, workloadID)
	if err != nil {
		return fmt.Errorf("failed to get edge workload: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Edge Workload: %s", workload.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", workload.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Type: %s", workload.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Priority: %d", workload.Priority))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", workload.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  Tasks: %d", len(workload.Tasks)))
	c.formatter.PrintInfo(fmt.Sprintf("  Created At: %s", workload.CreatedAt.Format("2006-01-02 15:04:05")))
	if workload.ScheduledAt.After(workload.CreatedAt) {
		c.formatter.PrintInfo(fmt.Sprintf("  Scheduled At: %s", workload.ScheduledAt.Format("2006-01-02 15:04:05")))
	}
	if workload.CompletedAt != nil {
		c.formatter.PrintInfo(fmt.Sprintf("  Completed At: %s", workload.CompletedAt.Format("2006-01-02 15:04:05")))
	}

	return nil
}

// listEdgeWorkloads lists all edge workloads
func (c *EdgeCommand) listEdgeWorkloads(ctx context.Context, _ []string) error {
	workloads, err := c.edgeManager.ListEdgeWorkloads(ctx)
	if err != nil {
		return fmt.Errorf("failed to list edge workloads: %v", err)
	}

	if len(workloads) == 0 {
		c.formatter.PrintInfo("No edge workloads found")
		return nil
	}

	headers := []string{"ID", "Name", "Type", "Priority", "Status", "Tasks", "Created At"}
	rows := make([][]string, len(workloads))

	for i, workload := range workloads {
		rows[i] = []string{
			workload.ID,
			workload.Name,
			workload.Type,
			fmt.Sprintf("%d", workload.Priority),
			workload.Status,
			fmt.Sprintf("%d", len(workload.Tasks)),
			workload.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.formatter.PrintTable(headers, rows)
	return nil
}

// scheduleWorkload schedules a workload
func (c *EdgeCommand) scheduleWorkload(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: edge schedule-workload <workload-id>")
	}

	workloadID := args[0]
	err := c.edgeManager.ScheduleWorkload(ctx, workloadID)
	if err != nil {
		return fmt.Errorf("failed to schedule workload: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Edge workload '%s' scheduled successfully", workloadID))
	return nil
}

// getStatistics gets edge computing statistics
func (c *EdgeCommand) getStatistics(ctx context.Context, _ []string) error {
	stats, err := c.edgeManager.GetEdgeStatistics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get statistics: %v", err)
	}

	c.formatter.PrintInfo("Edge Computing Statistics:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Nodes: %v", stats["total_nodes"]))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Tasks: %v", stats["total_tasks"]))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Workloads: %v", stats["total_workloads"]))

	if nodesByStatus, ok := stats["nodes_by_status"].(map[string]int); ok {
		c.formatter.PrintInfo("  Nodes by Status:")
		for status, count := range nodesByStatus {
			c.formatter.PrintInfo(fmt.Sprintf("    %s: %d", status, count))
		}
	}

	if tasksByStatus, ok := stats["tasks_by_status"].(map[string]int); ok {
		c.formatter.PrintInfo("  Tasks by Status:")
		for status, count := range tasksByStatus {
			c.formatter.PrintInfo(fmt.Sprintf("    %s: %d", status, count))
		}
	}

	if avgCPU, ok := stats["avg_cpu_usage"].(float64); ok {
		c.formatter.PrintInfo(fmt.Sprintf("  Average CPU Usage: %.1f%%", avgCPU))
	}
	if avgMemory, ok := stats["avg_memory_usage"].(float64); ok {
		c.formatter.PrintInfo(fmt.Sprintf("  Average Memory Usage: %.1f%%", avgMemory))
	}
	if avgDisk, ok := stats["avg_disk_usage"].(float64); ok {
		c.formatter.PrintInfo(fmt.Sprintf("  Average Disk Usage: %.1f%%", avgDisk))
	}

	return nil
}

// showHelp shows help information
func (c *EdgeCommand) showHelp() error {
	c.formatter.PrintInfo("Edge Computing Commands:")
	c.formatter.PrintInfo("  add-node <id> <name> <location> [ip] [status] - Add a new edge node")
	c.formatter.PrintInfo("  remove-node <node-id> - Remove an edge node")
	c.formatter.PrintInfo("  list-nodes - List all edge nodes")
	c.formatter.PrintInfo("  get-node <node-id> - Get edge node details")
	c.formatter.PrintInfo("  update-status <node-id> <status> - Update edge node status")
	c.formatter.PrintInfo("  update-metrics <node-id> <cpu> <memory> <disk> <latency> - Update edge node metrics")
	c.formatter.PrintInfo("  create-task <id> <node-id> <name> <type> [priority] - Create an edge task")
	c.formatter.PrintInfo("  get-task <task-id> - Get edge task details")
	c.formatter.PrintInfo("  list-tasks - List all edge tasks")
	c.formatter.PrintInfo("  get-tasks-by-node <node-id> - Get tasks for a specific node")
	c.formatter.PrintInfo("  update-task-status <task-id> <status> - Update edge task status")
	c.formatter.PrintInfo("  create-workload <id> <name> <type> [priority] - Create an edge workload")
	c.formatter.PrintInfo("  get-workload <workload-id> - Get edge workload details")
	c.formatter.PrintInfo("  list-workloads - List all edge workloads")
	c.formatter.PrintInfo("  schedule-workload <workload-id> - Schedule a workload")
	c.formatter.PrintInfo("  stats - Show edge computing statistics")
	return nil
}
