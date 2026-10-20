package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Skpow1234/Peervault/internal/edge"
)

func main() {
	var (
		command = flag.String("command", "help", "Command to execute (node, task, metrics, help)")
		nodeID  = flag.String("node-id", "", "Node ID")
		taskID  = flag.String("task-id", "", "Task ID")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help || *command == "help" {
		showHelp()
		return
	}

	// Create edge computing manager
	edgeManager := edge.NewEdgeComputingManager()
	ctx := context.Background()

	switch *command {
	case "node":
		handleNodeCommand(ctx, edgeManager, *nodeID)
	case "task":
		handleTaskCommand(ctx, edgeManager, *taskID)
	case "metrics":
		handleMetricsCommand(ctx, edgeManager)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func handleNodeCommand(ctx context.Context, edgeManager *edge.EdgeComputingManager, nodeID string) {
	// Create sample edge nodes
	sampleNodes := []*edge.EdgeNode{
		{
			ID:   "node-1",
			Name: "Edge Node 1",
			Location: &edge.Location{
				Latitude:  37.7749,
				Longitude: -122.4194,
				Altitude:  10.0,
				Address:   "123 Main St",
				City:      "San Francisco",
				Country:   "USA",
			},
			Capabilities: &edge.NodeCapabilities{
				CPU: &edge.CPUSpec{
					Cores:        4,
					Frequency:    2.4,
					Architecture: "x86_64",
					Usage:        25.0,
				},
				Memory: &edge.MemorySpec{
					Total:     8589934592, // 8GB
					Available: 6442450944, // 6GB
					Usage:     25.0,
				},
				Storage: &edge.StorageSpec{
					Total:     107374182400, // 100GB
					Available: 85899345920,  // 80GB
					Usage:     20.0,
					Type:      "SSD",
				},
				Network: &edge.NetworkSpec{
					Bandwidth: 1000000000, // 1Gbps
					Latency:   5.0,
					Protocols: []string{"TCP", "UDP", "HTTP", "HTTPS"},
					IPAddress: "192.168.1.100",
				},
				Protocols: []string{"HTTP", "HTTPS", "gRPC", "WebSocket"},
				Services:  []string{"compute", "storage", "network"},
			},
			Status: "active",
		},
		{
			ID:   "node-2",
			Name: "Edge Node 2",
			Location: &edge.Location{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Altitude:  5.0,
				Address:   "456 Broadway",
				City:      "New York",
				Country:   "USA",
			},
			Capabilities: &edge.NodeCapabilities{
				CPU: &edge.CPUSpec{
					Cores:        8,
					Frequency:    3.2,
					Architecture: "x86_64",
					Usage:        45.0,
				},
				Memory: &edge.MemorySpec{
					Total:     17179869184, // 16GB
					Available: 10737418240, // 10GB
					Usage:     37.5,
				},
				Storage: &edge.StorageSpec{
					Total:     214748364800, // 200GB
					Available: 171798691840, // 160GB
					Usage:     20.0,
					Type:      "NVMe",
				},
				Network: &edge.NetworkSpec{
					Bandwidth: 10000000000, // 10Gbps
					Latency:   2.0,
					Protocols: []string{"TCP", "UDP", "HTTP", "HTTPS", "gRPC"},
					IPAddress: "192.168.1.101",
				},
				GPU: &edge.GPUSpec{
					Model:             "NVIDIA RTX 3080",
					Memory:            10737418240, // 10GB
					Usage:             30.0,
					ComputeCapability: "8.6",
				},
				Protocols: []string{"HTTP", "HTTPS", "gRPC", "WebSocket", "MQTT"},
				Services:  []string{"compute", "storage", "network", "gpu"},
			},
			Status: "active",
		},
	}

	// Register sample nodes
	for _, node := range sampleNodes {
		err := edgeManager.RegisterNode(ctx, node)
		if err != nil {
			log.Printf("Failed to register node %s: %v", node.ID, err)
		}
	}

	// List all nodes
	nodes, err := edgeManager.ListNodes(ctx)
	if err != nil {
		log.Fatalf("Failed to list nodes: %v", err)
	}

	fmt.Printf("Edge Computing Nodes:\n")
	for _, node := range nodes {
		fmt.Printf("  ID: %s\n", node.ID)
		fmt.Printf("  Name: %s\n", node.Name)
		fmt.Printf("  Status: %s\n", node.Status)
		fmt.Printf("  Location: %s, %s\n", node.Location.City, node.Location.Country)
		fmt.Printf("  CPU: %d cores @ %.1f GHz (%.1f%% usage)\n",
			node.Capabilities.CPU.Cores,
			node.Capabilities.CPU.Frequency,
			node.Capabilities.CPU.Usage)
		fmt.Printf("  Memory: %.1f GB / %.1f GB (%.1f%% usage)\n",
			float64(node.Capabilities.Memory.Available)/1024/1024/1024,
			float64(node.Capabilities.Memory.Total)/1024/1024/1024,
			node.Capabilities.Memory.Usage)
		fmt.Printf("  Storage: %.1f GB / %.1f GB (%.1f%% usage)\n",
			float64(node.Capabilities.Storage.Available)/1024/1024/1024,
			float64(node.Capabilities.Storage.Total)/1024/1024/1024,
			node.Capabilities.Storage.Usage)
		fmt.Printf("  Network: %.1f Gbps, %.1f ms latency\n",
			float64(node.Capabilities.Network.Bandwidth)/1000000000,
			node.Capabilities.Network.Latency)
		if node.Capabilities.GPU != nil {
			fmt.Printf("  GPU: %s (%.1f GB, %.1f%% usage)\n",
				node.Capabilities.GPU.Model,
				float64(node.Capabilities.GPU.Memory)/1024/1024/1024,
				node.Capabilities.GPU.Usage)
		}
		fmt.Printf("  Services: %v\n", node.Capabilities.Services)
		fmt.Printf("  ---\n")
	}

	// Find nearest nodes to a location
	sfLocation := &edge.Location{
		Latitude:  37.7749,
		Longitude: -122.4194,
		Altitude:  0.0,
		Address:   "San Francisco, CA",
		City:      "San Francisco",
		Country:   "USA",
	}

	nearestNodes, err := edgeManager.FindNearestNodes(ctx, sfLocation, 1000.0, 5)
	if err != nil {
		log.Fatalf("Failed to find nearest nodes: %v", err)
	}

	fmt.Printf("\nNearest Nodes to San Francisco (within 1000km):\n")
	for _, node := range nearestNodes {
		// Calculate distance manually since the method is private
		distance := calculateDistance(sfLocation, node.Location)
		fmt.Printf("  %s: %.1f km away\n", node.Name, distance)
	}
}

func handleTaskCommand(ctx context.Context, edgeManager *edge.EdgeComputingManager, taskID string) {
	// Create sample tasks
	sampleTasks := []*edge.EdgeTask{
		{
			ID:       "task-1",
			Name:     "Image Processing Task",
			Type:     "compute",
			Priority: 1,
			Requirements: &edge.TaskRequirements{
				CPU:      2.0,
				Memory:   4294967296, // 4GB
				Storage:  1073741824, // 1GB
				Network:  100000000,  // 100Mbps
				GPU:      true,
				IoT:      false,
				Latency:  10.0,
				Duration: 5 * time.Minute,
			},
			Input: map[string]interface{}{
				"image_path": "/data/image.jpg",
				"operation":  "resize",
				"width":      1920,
				"height":     1080,
			},
			Status: "pending",
		},
		{
			ID:       "task-2",
			Name:     "Data Analysis Task",
			Type:     "compute",
			Priority: 2,
			Requirements: &edge.TaskRequirements{
				CPU:      4.0,
				Memory:   8589934592, // 8GB
				Storage:  2147483648, // 2GB
				Network:  1000000000, // 1Gbps
				GPU:      false,
				IoT:      false,
				Latency:  50.0,
				Duration: 15 * time.Minute,
			},
			Input: map[string]interface{}{
				"dataset_path": "/data/dataset.csv",
				"algorithm":    "linear_regression",
				"parameters":   map[string]interface{}{"learning_rate": 0.01},
			},
			Status: "pending",
		},
		{
			ID:       "task-3",
			Name:     "IoT Data Collection",
			Type:     "iot",
			Priority: 3,
			Requirements: &edge.TaskRequirements{
				CPU:      1.0,
				Memory:   1073741824, // 1GB
				Storage:  536870912,  // 512MB
				Network:  10000000,   // 10Mbps
				GPU:      false,
				IoT:      true,
				Latency:  100.0,
				Duration: 30 * time.Minute,
			},
			Input: map[string]interface{}{
				"sensor_type": "temperature",
				"interval":    "1m",
				"duration":    "30m",
			},
			Status: "pending",
		},
	}

	// Submit tasks
	for _, task := range sampleTasks {
		err := edgeManager.SubmitTask(ctx, task)
		if err != nil {
			log.Printf("Failed to submit task %s: %v", task.ID, err)
		}
	}

	// List all tasks
	tasks, err := edgeManager.ListTasks(ctx)
	if err != nil {
		log.Fatalf("Failed to list tasks: %v", err)
	}

	fmt.Printf("Edge Computing Tasks:\n")
	for _, task := range tasks {
		fmt.Printf("  ID: %s\n", task.ID)
		fmt.Printf("  Name: %s\n", task.Name)
		fmt.Printf("  Type: %s\n", task.Type)
		fmt.Printf("  Priority: %d\n", task.Priority)
		fmt.Printf("  Status: %s\n", task.Status)
		fmt.Printf("  Assigned Node: %s\n", task.AssignedNode)
		fmt.Printf("  CPU Required: %.1f cores\n", task.Requirements.CPU)
		fmt.Printf("  Memory Required: %.1f GB\n", float64(task.Requirements.Memory)/1024/1024/1024)
		fmt.Printf("  Storage Required: %.1f GB\n", float64(task.Requirements.Storage)/1024/1024/1024)
		fmt.Printf("  Network Required: %.1f Mbps\n", float64(task.Requirements.Network)/1000000)
		fmt.Printf("  GPU Required: %t\n", task.Requirements.GPU)
		fmt.Printf("  IoT Required: %t\n", task.Requirements.IoT)
		fmt.Printf("  Max Latency: %.1f ms\n", task.Requirements.Latency)
		fmt.Printf("  Estimated Duration: %v\n", task.Requirements.Duration)
		fmt.Printf("  Created At: %s\n", task.CreatedAt.Format(time.RFC3339))
		if task.StartedAt != nil {
			fmt.Printf("  Started At: %s\n", task.StartedAt.Format(time.RFC3339))
		}
		if task.CompletedAt != nil {
			fmt.Printf("  Completed At: %s\n", task.CompletedAt.Format(time.RFC3339))
		}
		fmt.Printf("  ---\n")
	}

	// Optimize resource allocation
	err = edgeManager.OptimizeResourceAllocation(ctx)
	if err != nil {
		log.Printf("Failed to optimize resource allocation: %v", err)
	} else {
		fmt.Printf("\nResource allocation optimized successfully!\n")
	}
}

func handleMetricsCommand(ctx context.Context, edgeManager *edge.EdgeComputingManager) {
	// Get metrics
	metrics, err := edgeManager.GetMetrics(ctx)
	if err != nil {
		log.Fatalf("Failed to get metrics: %v", err)
	}

	fmt.Printf("Edge Computing Metrics:\n")
	fmt.Printf("  Total Nodes: %d\n", metrics.TotalNodes)
	fmt.Printf("  Active Nodes: %d\n", metrics.ActiveNodes)
	fmt.Printf("  Total Tasks: %d\n", metrics.TotalTasks)
	fmt.Printf("  Completed Tasks: %d\n", metrics.CompletedTasks)
	fmt.Printf("  Failed Tasks: %d\n", metrics.FailedTasks)
	fmt.Printf("  Average Latency: %.2f ms\n", metrics.AverageLatency)
	fmt.Printf("  Resource Utilization: %.2f%%\n", metrics.ResourceUtilization)

	// Calculate success rate
	if metrics.TotalTasks > 0 {
		successRate := float64(metrics.CompletedTasks) / float64(metrics.TotalTasks) * 100
		fmt.Printf("  Success Rate: %.2f%%\n", successRate)
	}

	// Calculate node utilization
	if metrics.TotalNodes > 0 {
		nodeUtilization := float64(metrics.ActiveNodes) / float64(metrics.TotalNodes) * 100
		fmt.Printf("  Node Utilization: %.2f%%\n", nodeUtilization)
	}
}

func showHelp() {
	fmt.Printf("PeerVault Edge Computing Tool\n\n")
	fmt.Printf("Usage: peervault-edge -command <command> [options]\n\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  node      Manage edge computing nodes\n")
	fmt.Printf("  task      Manage edge computing tasks\n")
	fmt.Printf("  metrics   Show edge computing metrics\n")
	fmt.Printf("  help      Show this help message\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -node-id <id>     Node ID (for node-specific operations)\n")
	fmt.Printf("  -task-id <id>     Task ID (for task-specific operations)\n")
	fmt.Printf("  -help             Show this help message\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  peervault-edge -command node\n")
	fmt.Printf("  peervault-edge -command task\n")
	fmt.Printf("  peervault-edge -command metrics\n")
}

// calculateDistance calculates the distance between two locations using Haversine formula
func calculateDistance(loc1, loc2 *edge.Location) float64 {
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
