package balancing

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LoadBalancerConfig represents load balancer configuration
type LoadBalancerConfig struct {
	Strategy            string // "round_robin", "random", "weighted", "least_connections"
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	MaxRetries          int
	RetryDelay          time.Duration
	FailoverTimeout     time.Duration
	StickySession       bool
	SessionTimeout      time.Duration
}

// DefaultLoadBalancerConfig returns the default load balancer configuration
func DefaultLoadBalancerConfig() *LoadBalancerConfig {
	return &LoadBalancerConfig{
		Strategy:            "round_robin",
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		MaxRetries:          3,
		RetryDelay:          time.Second,
		FailoverTimeout:     10 * time.Second,
		StickySession:       false,
		SessionTimeout:      5 * time.Minute,
	}
}

// LoadBalancer provides client-side load balancing for gRPC services
type LoadBalancer struct {
	config           *LoadBalancerConfig
	logger           *slog.Logger
	servers          []*Server
	serverMux        sync.RWMutex
	healthChecker    *HealthChecker
	serviceDiscovery *ServiceDiscovery
	roundRobinIndex  int
	random           *rand.Rand
	sessions         map[string]*Session
	sessionMux       sync.RWMutex
}

// Server represents a gRPC server instance
type Server struct {
	ID           string
	Address      string
	Port         int
	Weight       int
	Connections  int64
	LastHealth   time.Time
	HealthStatus HealthStatus
	Conn         *grpc.ClientConn
}

// HealthStatus represents the health status of a server
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Session represents a sticky session
type Session struct {
	ID        string
	ServerID  string
	CreatedAt time.Time
	LastUsed  time.Time
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(config *LoadBalancerConfig, logger *slog.Logger) *LoadBalancer {
	if config == nil {
		config = DefaultLoadBalancerConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &LoadBalancer{
		config:           config,
		logger:           logger,
		servers:          make([]*Server, 0),
		healthChecker:    NewHealthChecker(config.HealthCheckInterval, config.HealthCheckTimeout, logger),
		serviceDiscovery: NewServiceDiscovery(logger),
		roundRobinIndex:  0,
		random:           rand.New(rand.NewSource(time.Now().UnixNano())),
		sessions:         make(map[string]*Session),
	}
}

// AddServer adds a server to the load balancer
func (lb *LoadBalancer) AddServer(id, address string, port int, weight int) error {
	lb.serverMux.Lock()
	defer lb.serverMux.Unlock()

	// Check if server already exists
	for _, server := range lb.servers {
		if server.ID == id {
			return fmt.Errorf("server with ID %s already exists", id)
		}
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", address, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %w", id, err)
	}

	server := &Server{
		ID:           id,
		Address:      address,
		Port:         port,
		Weight:       weight,
		Connections:  0,
		LastHealth:   time.Now(),
		HealthStatus: HealthStatusUnknown,
		Conn:         conn,
	}

	lb.servers = append(lb.servers, server)

	// Start health checking for this server
	go lb.healthChecker.StartHealthCheck(server, lb.updateServerHealth)

	lb.logger.Info("Added server to load balancer", "server_id", id, "address", address, "port", port, "weight", weight)
	return nil
}

// RemoveServer removes a server from the load balancer
func (lb *LoadBalancer) RemoveServer(id string) error {
	lb.serverMux.Lock()
	defer lb.serverMux.Unlock()

	for i, server := range lb.servers {
		if server.ID == id {
			// Close connection
			if server.Conn != nil {
				if err := server.Conn.Close(); err != nil {
					lb.logger.Warn("Failed to close server connection", "server_id", id, "error", err)
				}
			}

			// Remove from slice
			lb.servers = append(lb.servers[:i], lb.servers[i+1:]...)

			lb.logger.Info("Removed server from load balancer", "server_id", id)
			return nil
		}
	}

	return fmt.Errorf("server with ID %s not found", id)
}

// GetServer returns a server based on the load balancing strategy
func (lb *LoadBalancer) GetServer(sessionID string) (*Server, error) {
	lb.serverMux.RLock()
	defer lb.serverMux.RUnlock()

	// Check for sticky session
	if lb.config.StickySession && sessionID != "" {
		if server := lb.getStickyServer(sessionID); server != nil {
			return server, nil
		}
	}

	// Get healthy servers
	healthyServers := lb.getHealthyServers()
	if len(healthyServers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}

	// Select server based on strategy
	var selectedServer *Server
	switch lb.config.Strategy {
	case "round_robin":
		selectedServer = lb.selectRoundRobin(healthyServers)
	case "random":
		selectedServer = lb.selectRandom(healthyServers)
	case "weighted":
		selectedServer = lb.selectWeighted(healthyServers)
	case "least_connections":
		selectedServer = lb.selectLeastConnections(healthyServers)
	default:
		selectedServer = lb.selectRoundRobin(healthyServers)
	}

	// Create sticky session if enabled
	if lb.config.StickySession && sessionID != "" {
		lb.createStickySession(sessionID, selectedServer.ID)
	}

	return selectedServer, nil
}

// getHealthyServers returns a list of healthy servers
func (lb *LoadBalancer) getHealthyServers() []*Server {
	var healthyServers []*Server
	for _, server := range lb.servers {
		if server.HealthStatus == HealthStatusHealthy {
			healthyServers = append(healthyServers, server)
		}
	}
	return healthyServers
}

// selectRoundRobin selects a server using round-robin strategy
func (lb *LoadBalancer) selectRoundRobin(servers []*Server) *Server {
	if len(servers) == 0 {
		return nil
	}

	server := servers[lb.roundRobinIndex%len(servers)]
	lb.roundRobinIndex++
	return server
}

// selectRandom selects a server using random strategy
func (lb *LoadBalancer) selectRandom(servers []*Server) *Server {
	if len(servers) == 0 {
		return nil
	}

	index := lb.random.Intn(len(servers))
	return servers[index]
}

// selectWeighted selects a server using weighted strategy
func (lb *LoadBalancer) selectWeighted(servers []*Server) *Server {
	if len(servers) == 0 {
		return nil
	}

	totalWeight := 0
	for _, server := range servers {
		totalWeight += server.Weight
	}

	if totalWeight == 0 {
		return lb.selectRandom(servers)
	}

	randomWeight := lb.random.Intn(totalWeight)
	currentWeight := 0

	for _, server := range servers {
		currentWeight += server.Weight
		if randomWeight < currentWeight {
			return server
		}
	}

	return servers[0]
}

// selectLeastConnections selects a server with the least connections
func (lb *LoadBalancer) selectLeastConnections(servers []*Server) *Server {
	if len(servers) == 0 {
		return nil
	}

	selectedServer := servers[0]
	minConnections := selectedServer.Connections

	for _, server := range servers {
		if server.Connections < minConnections {
			selectedServer = server
			minConnections = server.Connections
		}
	}

	return selectedServer
}

// getStickyServer returns the server for a sticky session
func (lb *LoadBalancer) getStickyServer(sessionID string) *Server {
	lb.sessionMux.RLock()
	defer lb.sessionMux.RUnlock()

	session, exists := lb.sessions[sessionID]
	if !exists {
		return nil
	}

	// Check if session is expired
	if time.Since(session.LastUsed) > lb.config.SessionTimeout {
		delete(lb.sessions, sessionID)
		return nil
	}

	// Find server by ID
	for _, server := range lb.servers {
		if server.ID == session.ServerID && server.HealthStatus == HealthStatusHealthy {
			session.LastUsed = time.Now()
			return server
		}
	}

	// Server not found or unhealthy, remove session
	delete(lb.sessions, sessionID)
	return nil
}

// createStickySession creates a sticky session
func (lb *LoadBalancer) createStickySession(sessionID, serverID string) {
	lb.sessionMux.Lock()
	defer lb.sessionMux.Unlock()

	lb.sessions[sessionID] = &Session{
		ID:        sessionID,
		ServerID:  serverID,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}
}

// updateServerHealth updates the health status of a server
func (lb *LoadBalancer) updateServerHealth(serverID string, status HealthStatus) {
	lb.serverMux.Lock()
	defer lb.serverMux.Unlock()

	for _, server := range lb.servers {
		if server.ID == serverID {
			server.HealthStatus = status
			server.LastHealth = time.Now()
			lb.logger.Debug("Updated server health", "server_id", serverID, "status", status)
			break
		}
	}
}

// IncrementConnections increments the connection count for a server
func (lb *LoadBalancer) IncrementConnections(serverID string) {
	lb.serverMux.Lock()
	defer lb.serverMux.Unlock()

	for _, server := range lb.servers {
		if server.ID == serverID {
			server.Connections++
			break
		}
	}
}

// DecrementConnections decrements the connection count for a server
func (lb *LoadBalancer) DecrementConnections(serverID string) {
	lb.serverMux.Lock()
	defer lb.serverMux.Unlock()

	for _, server := range lb.servers {
		if server.ID == serverID {
			server.Connections--
			if server.Connections < 0 {
				server.Connections = 0
			}
			break
		}
	}
}

// GetServerStats returns statistics for all servers
func (lb *LoadBalancer) GetServerStats() map[string]interface{} {
	lb.serverMux.RLock()
	defer lb.serverMux.RUnlock()

	stats := make(map[string]interface{})
	stats["total_servers"] = len(lb.servers)
	stats["healthy_servers"] = 0
	stats["unhealthy_servers"] = 0
	stats["unknown_servers"] = 0
	stats["total_connections"] = int64(0)

	serverStats := make([]map[string]interface{}, 0)
	for _, server := range lb.servers {
		serverStat := map[string]interface{}{
			"id":            server.ID,
			"address":       server.Address,
			"port":          server.Port,
			"weight":        server.Weight,
			"connections":   server.Connections,
			"health_status": server.HealthStatus,
			"last_health":   server.LastHealth,
		}
		serverStats = append(serverStats, serverStat)

		switch server.HealthStatus {
		case HealthStatusHealthy:
			stats["healthy_servers"] = stats["healthy_servers"].(int) + 1
		case HealthStatusUnhealthy:
			stats["unhealthy_servers"] = stats["unhealthy_servers"].(int) + 1
		case HealthStatusUnknown:
			stats["unknown_servers"] = stats["unknown_servers"].(int) + 1
		}

		stats["total_connections"] = stats["total_connections"].(int64) + server.Connections
	}

	stats["servers"] = serverStats
	stats["strategy"] = lb.config.Strategy
	stats["sticky_session"] = lb.config.StickySession
	stats["active_sessions"] = len(lb.sessions)

	return stats
}

// HealthCheck performs a health check on all servers
func (lb *LoadBalancer) HealthCheck() map[string]interface{} {
	lb.serverMux.RLock()
	defer lb.serverMux.RUnlock()

	health := make(map[string]interface{})
	health["status"] = "healthy"
	health["timestamp"] = time.Now()

	healthyCount := 0
	totalCount := len(lb.servers)

	for _, server := range lb.servers {
		if server.HealthStatus == HealthStatusHealthy {
			healthyCount++
		}
	}

	health["healthy_servers"] = healthyCount
	health["total_servers"] = totalCount
	health["health_percentage"] = float64(healthyCount) / float64(totalCount) * 100

	if healthyCount == 0 {
		health["status"] = "unhealthy"
	} else if healthyCount < totalCount {
		health["status"] = "degraded"
	}

	return health
}

// Close closes the load balancer and all connections
func (lb *LoadBalancer) Close() error {
	lb.serverMux.Lock()
	defer lb.serverMux.Unlock()

	// Stop health checker
	lb.healthChecker.Stop()

	// Close all server connections
	for _, server := range lb.servers {
		if server.Conn != nil {
			if err := server.Conn.Close(); err != nil {
				lb.logger.Warn("Failed to close server connection", "server_id", server.ID, "error", err)
			}
		}
	}

	// Clear servers
	lb.servers = make([]*Server, 0)

	// Clear sessions
	lb.sessionMux.Lock()
	lb.sessions = make(map[string]*Session)
	lb.sessionMux.Unlock()

	lb.logger.Info("Load balancer closed")
	return nil
}
