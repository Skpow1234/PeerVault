package network

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// LoadBalancer manages load balancing across multiple servers
type LoadBalancer struct {
	client    *client.Client
	configDir string
	servers   []*Server
	algorithm LoadBalancingAlgorithm
	stats     *LoadBalancerStats
	mu        sync.RWMutex
}

// Server represents a backend server
type Server struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	URL               string    `json:"url"`
	Weight            int       `json:"weight"`
	HealthCheck       string    `json:"health_check"`
	IsHealthy         bool      `json:"is_healthy"`
	LastCheck         time.Time `json:"last_check"`
	ResponseTime      int64     `json:"response_time"` // in milliseconds
	ActiveConnections int       `json:"active_connections"`
	TotalRequests     int64     `json:"total_requests"`
	FailedRequests    int64     `json:"failed_requests"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// LoadBalancingAlgorithm represents different load balancing strategies
type LoadBalancingAlgorithm string

const (
	RoundRobin         LoadBalancingAlgorithm = "round_robin"
	WeightedRoundRobin LoadBalancingAlgorithm = "weighted_round_robin"
	LeastConnections   LoadBalancingAlgorithm = "least_connections"
	LeastResponseTime  LoadBalancingAlgorithm = "least_response_time"
	Random             LoadBalancingAlgorithm = "random"
	IPHash             LoadBalancingAlgorithm = "ip_hash"
)

// LoadBalancerStats represents load balancer statistics
type LoadBalancerStats struct {
	TotalRequests       int64     `json:"total_requests"`
	SuccessfulRequests  int64     `json:"successful_requests"`
	FailedRequests      int64     `json:"failed_requests"`
	AverageResponseTime float64   `json:"average_response_time"`
	ActiveConnections   int       `json:"active_connections"`
	HealthyServers      int       `json:"healthy_servers"`
	UnhealthyServers    int       `json:"unhealthy_servers"`
	LastUpdated         time.Time `json:"last_updated"`
}

// LoadBalancerConfig represents load balancer configuration
type LoadBalancerConfig struct {
	Algorithm           LoadBalancingAlgorithm `json:"algorithm"`
	HealthCheckInterval time.Duration          `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration          `json:"health_check_timeout"`
	MaxRetries          int                    `json:"max_retries"`
	RetryDelay          time.Duration          `json:"retry_delay"`
	StickySession       bool                   `json:"sticky_session"`
	SessionTimeout      time.Duration          `json:"session_timeout"`
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(client *client.Client, configDir string) *LoadBalancer {
	lb := &LoadBalancer{
		client:    client,
		configDir: configDir,
		servers:   make([]*Server, 0),
		algorithm: RoundRobin,
		stats:     &LoadBalancerStats{},
	}

	_ = lb.loadServers() // Ignore error for initialization
	_ = lb.loadConfig()  // Ignore error for initialization
	_ = lb.loadStats()   // Ignore error for initialization

	// Start health check routine
	go lb.startHealthChecks()

	return lb
}

// AddServer adds a new server to the load balancer
func (lb *LoadBalancer) AddServer(server *Server) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Check if server already exists
	for _, existingServer := range lb.servers {
		if existingServer.ID == server.ID {
			return fmt.Errorf("server with ID %s already exists", server.ID)
		}
	}

	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()
	server.IsHealthy = true // Assume healthy initially

	lb.servers = append(lb.servers, server)
	_ = lb.saveServers() // Ignore error for demo purposes

	return nil
}

// RemoveServer removes a server from the load balancer
func (lb *LoadBalancer) RemoveServer(serverID string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, server := range lb.servers {
		if server.ID == serverID {
			lb.servers = append(lb.servers[:i], lb.servers[i+1:]...)
			_ = lb.saveServers() // Ignore error for demo purposes
			return nil
		}
	}

	return fmt.Errorf("server with ID %s not found", serverID)
}

// GetServer returns a server by ID
func (lb *LoadBalancer) GetServer(serverID string) (*Server, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	for _, server := range lb.servers {
		if server.ID == serverID {
			// Return a copy
			serverCopy := *server
			return &serverCopy, nil
		}
	}

	return nil, fmt.Errorf("server with ID %s not found", serverID)
}

// ListServers returns all servers
func (lb *LoadBalancer) ListServers() []*Server {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	servers := make([]*Server, len(lb.servers))
	for i, server := range lb.servers {
		// Return copies
		serverCopy := *server
		servers[i] = &serverCopy
	}

	return servers
}

// SelectServer selects the best server based on the load balancing algorithm
func (lb *LoadBalancer) SelectServer() (*Server, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	// Filter healthy servers
	var healthyServers []*Server
	for _, server := range lb.servers {
		if server.IsHealthy {
			healthyServers = append(healthyServers, server)
		}
	}

	if len(healthyServers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}

	// Select server based on algorithm
	switch lb.algorithm {
	case RoundRobin:
		return lb.selectRoundRobin(healthyServers)
	case WeightedRoundRobin:
		return lb.selectWeightedRoundRobin(healthyServers)
	case LeastConnections:
		return lb.selectLeastConnections(healthyServers)
	case LeastResponseTime:
		return lb.selectLeastResponseTime(healthyServers)
	case Random:
		return lb.selectRandom(healthyServers)
	case IPHash:
		return lb.selectIPHash(healthyServers)
	default:
		return lb.selectRoundRobin(healthyServers)
	}
}

// UpdateServerHealth updates the health status of a server
func (lb *LoadBalancer) UpdateServerHealth(serverID string, isHealthy bool, responseTime int64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, server := range lb.servers {
		if server.ID == serverID {
			server.IsHealthy = isHealthy
			server.LastCheck = time.Now()
			server.ResponseTime = responseTime
			server.UpdatedAt = time.Now()
			break
		}
	}

	_ = lb.saveServers() // Ignore error for demo purposes
}

// RecordRequest records a request to a server
func (lb *LoadBalancer) RecordRequest(serverID string, success bool, responseTime int64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Update server stats
	for _, server := range lb.servers {
		if server.ID == serverID {
			server.TotalRequests++
			if !success {
				server.FailedRequests++
			}
			server.UpdatedAt = time.Now()
			break
		}
	}

	// Update load balancer stats
	lb.stats.TotalRequests++
	if success {
		lb.stats.SuccessfulRequests++
	} else {
		lb.stats.FailedRequests++
	}

	// Update average response time
	if lb.stats.TotalRequests > 0 {
		lb.stats.AverageResponseTime = (lb.stats.AverageResponseTime*float64(lb.stats.TotalRequests-1) + float64(responseTime)) / float64(lb.stats.TotalRequests)
	}

	lb.stats.LastUpdated = time.Now()
	_ = lb.saveStats() // Ignore error for demo purposes
}

// GetStats returns load balancer statistics
func (lb *LoadBalancer) GetStats() *LoadBalancerStats {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Update current stats
	healthyCount := 0
	unhealthyCount := 0
	activeConnections := 0

	for _, server := range lb.servers {
		if server.IsHealthy {
			healthyCount++
		} else {
			unhealthyCount++
		}
		activeConnections += server.ActiveConnections
	}

	lb.stats.HealthyServers = healthyCount
	lb.stats.UnhealthyServers = unhealthyCount
	lb.stats.ActiveConnections = activeConnections

	// Return a copy
	stats := *lb.stats
	return &stats
}

// UpdateConfig updates load balancer configuration
func (lb *LoadBalancer) UpdateConfig(config *LoadBalancerConfig) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.algorithm = config.Algorithm
	_ = lb.saveConfig() // Ignore error for demo purposes

	return nil
}

// GetConfig returns current load balancer configuration
func (lb *LoadBalancer) GetConfig() *LoadBalancerConfig {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return &LoadBalancerConfig{
		Algorithm:           lb.algorithm,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
		StickySession:       false,
		SessionTimeout:      30 * time.Minute,
	}
}

// Selection algorithms
func (lb *LoadBalancer) selectRoundRobin(servers []*Server) (*Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	// Simple round robin - select first server for now
	// In a real implementation, you'd maintain a counter
	return servers[0], nil
}

func (lb *LoadBalancer) selectWeightedRoundRobin(servers []*Server) (*Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	// Calculate total weight
	totalWeight := 0
	for _, server := range servers {
		totalWeight += server.Weight
	}

	if totalWeight == 0 {
		return servers[0], nil
	}

	// Select based on weight (simplified implementation)
	random := rand.Intn(totalWeight)
	currentWeight := 0

	for _, server := range servers {
		currentWeight += server.Weight
		if random < currentWeight {
			return server, nil
		}
	}

	return servers[0], nil
}

func (lb *LoadBalancer) selectLeastConnections(servers []*Server) (*Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	bestServer := servers[0]
	minConnections := bestServer.ActiveConnections

	for _, server := range servers[1:] {
		if server.ActiveConnections < minConnections {
			bestServer = server
			minConnections = server.ActiveConnections
		}
	}

	return bestServer, nil
}

func (lb *LoadBalancer) selectLeastResponseTime(servers []*Server) (*Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	bestServer := servers[0]
	minResponseTime := bestServer.ResponseTime

	for _, server := range servers[1:] {
		if server.ResponseTime < minResponseTime {
			bestServer = server
			minResponseTime = server.ResponseTime
		}
	}

	return bestServer, nil
}

func (lb *LoadBalancer) selectRandom(servers []*Server) (*Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	index := rand.Intn(len(servers))
	return servers[index], nil
}

func (lb *LoadBalancer) selectIPHash(servers []*Server) (*Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	// Simplified IP hash - just use random for now
	// In a real implementation, you'd hash the client IP
	index := rand.Intn(len(servers))
	return servers[index], nil
}

// Health check routine
func (lb *LoadBalancer) startHealthChecks() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		lb.performHealthChecks()
	}
}

func (lb *LoadBalancer) performHealthChecks() {
	lb.mu.RLock()
	servers := make([]*Server, len(lb.servers))
	copy(servers, lb.servers)
	lb.mu.RUnlock()

	for _, server := range servers {
		go lb.checkServerHealth(server)
	}
}

func (lb *LoadBalancer) checkServerHealth(server *Server) {
	start := time.Now()

	// Simulate health check
	// In a real implementation, you'd make an HTTP request to the health check endpoint
	time.Sleep(time.Millisecond * 10) // Simulate network delay

	responseTime := time.Since(start).Milliseconds()
	isHealthy := responseTime < 1000 // Consider healthy if response time < 1 second

	lb.UpdateServerHealth(server.ID, isHealthy, responseTime)
}

// Data persistence
func (lb *LoadBalancer) loadServers() error {
	serversFile := filepath.Join(lb.configDir, "servers.json")
	if _, err := os.Stat(serversFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty servers
	}

	data, err := os.ReadFile(serversFile)
	if err != nil {
		return fmt.Errorf("failed to read servers file: %w", err)
	}

	var servers []*Server
	if err := json.Unmarshal(data, &servers); err != nil {
		return fmt.Errorf("failed to unmarshal servers: %w", err)
	}

	lb.servers = servers
	return nil
}

func (lb *LoadBalancer) saveServers() error {
	serversFile := filepath.Join(lb.configDir, "servers.json")

	data, err := json.MarshalIndent(lb.servers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal servers: %w", err)
	}

	return os.WriteFile(serversFile, data, 0644)
}

func (lb *LoadBalancer) loadConfig() error {
	configFile := filepath.Join(lb.configDir, "loadbalancer.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil // Use default config
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config LoadBalancerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	lb.algorithm = config.Algorithm
	return nil
}

func (lb *LoadBalancer) saveConfig() error {
	configFile := filepath.Join(lb.configDir, "loadbalancer.json")

	config := &LoadBalancerConfig{
		Algorithm:           lb.algorithm,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
		StickySession:       false,
		SessionTimeout:      30 * time.Minute,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configFile, data, 0644)
}

func (lb *LoadBalancer) loadStats() error {
	statsFile := filepath.Join(lb.configDir, "loadbalancer_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil // Use default stats
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats LoadBalancerStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	lb.stats = &stats
	return nil
}

func (lb *LoadBalancer) saveStats() error {
	statsFile := filepath.Join(lb.configDir, "loadbalancer_stats.json")

	data, err := json.MarshalIndent(lb.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
