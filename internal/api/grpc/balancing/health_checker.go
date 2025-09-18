package balancing

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// HealthChecker provides health checking functionality for servers
type HealthChecker struct {
	interval    time.Duration
	timeout     time.Duration
	logger      *slog.Logger
	checkers    map[string]*ServerHealthChecker
	checkersMux sync.RWMutex
	stopChan    chan struct{}
}

// ServerHealthChecker represents a health checker for a specific server
type ServerHealthChecker struct {
	Server     *Server
	Interval   time.Duration
	Timeout    time.Duration
	Logger     *slog.Logger
	UpdateFunc func(string, HealthStatus)
	stopChan   chan struct{}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval, timeout time.Duration, logger *slog.Logger) *HealthChecker {
	if logger == nil {
		logger = slog.Default()
	}

	return &HealthChecker{
		interval: interval,
		timeout:  timeout,
		logger:   logger,
		checkers: make(map[string]*ServerHealthChecker),
		stopChan: make(chan struct{}),
	}
}

// StartHealthCheck starts health checking for a server
func (hc *HealthChecker) StartHealthCheck(server *Server, updateFunc func(string, HealthStatus)) {
	hc.checkersMux.Lock()
	defer hc.checkersMux.Unlock()

	// Check if checker already exists
	if _, exists := hc.checkers[server.ID]; exists {
		hc.logger.Warn("Health checker already exists for server", "server_id", server.ID)
		return
	}

	// Create server health checker
	checker := &ServerHealthChecker{
		Server:     server,
		Interval:   hc.interval,
		Timeout:    hc.timeout,
		Logger:     hc.logger.With("server_id", server.ID),
		UpdateFunc: updateFunc,
		stopChan:   make(chan struct{}),
	}

	hc.checkers[server.ID] = checker

	// Start health checking
	go checker.start()

	hc.logger.Info("Started health checker for server", "server_id", server.ID)
}

// StopHealthCheck stops health checking for a server
func (hc *HealthChecker) StopHealthCheck(serverID string) {
	hc.checkersMux.Lock()
	defer hc.checkersMux.Unlock()

	checker, exists := hc.checkers[serverID]
	if !exists {
		return
	}

	checker.stop()
	delete(hc.checkers, serverID)

	hc.logger.Info("Stopped health checker for server", "server_id", serverID)
}

// Stop stops all health checkers
func (hc *HealthChecker) Stop() {
	hc.checkersMux.Lock()
	defer hc.checkersMux.Unlock()

	for serverID, checker := range hc.checkers {
		checker.stop()
		hc.logger.Info("Stopped health checker for server", "server_id", serverID)
	}

	hc.checkers = make(map[string]*ServerHealthChecker)
	close(hc.stopChan)
}

// start starts the health checking loop
func (shc *ServerHealthChecker) start() {
	ticker := time.NewTicker(shc.Interval)
	defer ticker.Stop()

	// Perform initial health check
	shc.performHealthCheck()

	for {
		select {
		case <-ticker.C:
			shc.performHealthCheck()
		case <-shc.stopChan:
			shc.Logger.Info("Health checker stopped")
			return
		}
	}
}

// stop stops the health checking loop
func (shc *ServerHealthChecker) stop() {
	close(shc.stopChan)
}

// performHealthCheck performs a health check on the server
func (shc *ServerHealthChecker) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), shc.Timeout)
	defer cancel()

	// Create a simple health check request
	// In a real implementation, you would use the actual gRPC health check service
	conn := shc.Server.Conn
	if conn == nil {
		shc.Logger.Error("Server connection is nil")
		shc.UpdateFunc(shc.Server.ID, HealthStatusUnhealthy)
		return
	}

	// Check connection state
	state := conn.GetState()
	if state.String() != "READY" {
		shc.Logger.Warn("Server connection not ready", "state", state.String())
		shc.UpdateFunc(shc.Server.ID, HealthStatusUnhealthy)
		return
	}

	// Perform a simple ping test
	// This is a simplified health check - in production, use the actual health check service
	err := shc.pingServer(ctx, conn)
	if err != nil {
		shc.Logger.Error("Health check failed", "error", err)
		shc.UpdateFunc(shc.Server.ID, HealthStatusUnhealthy)
		return
	}

	shc.Logger.Debug("Health check passed")
	shc.UpdateFunc(shc.Server.ID, HealthStatusHealthy)
}

// pingServer performs a simple ping to test server connectivity
func (shc *ServerHealthChecker) pingServer(ctx context.Context, conn *grpc.ClientConn) error {
	// This is a simplified implementation
	// In a real implementation, you would use the gRPC health check service

	// Try to get the connection state
	state := conn.GetState()
	if state.String() != "READY" {
		return fmt.Errorf("connection not ready: %s", state.String())
	}

	// Wait for connection to be ready
	ready := conn.WaitForStateChange(ctx, state)
	if !ready {
		return fmt.Errorf("connection not ready within timeout")
	}

	return nil
}

// GetHealthStatus returns the current health status of all servers
func (hc *HealthChecker) GetHealthStatus() map[string]HealthStatus {
	hc.checkersMux.RLock()
	defer hc.checkersMux.RUnlock()

	status := make(map[string]HealthStatus)
	for serverID, checker := range hc.checkers {
		status[serverID] = checker.Server.HealthStatus
	}

	return status
}

// GetHealthStats returns health statistics
func (hc *HealthChecker) GetHealthStats() map[string]interface{} {
	hc.checkersMux.RLock()
	defer hc.checkersMux.RUnlock()

	stats := make(map[string]interface{})
	stats["total_servers"] = len(hc.checkers)
	stats["healthy_servers"] = 0
	stats["unhealthy_servers"] = 0
	stats["unknown_servers"] = 0

	serverStats := make([]map[string]interface{}, 0)
	for serverID, checker := range hc.checkers {
		serverStat := map[string]interface{}{
			"server_id":     serverID,
			"health_status": checker.Server.HealthStatus,
			"last_health":   checker.Server.LastHealth,
			"address":       checker.Server.Address,
			"port":          checker.Server.Port,
		}
		serverStats = append(serverStats, serverStat)

		switch checker.Server.HealthStatus {
		case HealthStatusHealthy:
			stats["healthy_servers"] = stats["healthy_servers"].(int) + 1
		case HealthStatusUnhealthy:
			stats["unhealthy_servers"] = stats["unhealthy_servers"].(int) + 1
		case HealthStatusUnknown:
			stats["unknown_servers"] = stats["unknown_servers"].(int) + 1
		}
	}

	stats["servers"] = serverStats
	stats["check_interval"] = hc.interval
	stats["check_timeout"] = hc.timeout

	return stats
}

// SetHealthCheckInterval updates the health check interval for all servers
func (hc *HealthChecker) SetHealthCheckInterval(interval time.Duration) {
	hc.checkersMux.Lock()
	defer hc.checkersMux.Unlock()

	hc.interval = interval
	for _, checker := range hc.checkers {
		checker.Interval = interval
	}

	hc.logger.Info("Updated health check interval", "interval", interval)
}

// SetHealthCheckTimeout updates the health check timeout for all servers
func (hc *HealthChecker) SetHealthCheckTimeout(timeout time.Duration) {
	hc.checkersMux.Lock()
	defer hc.checkersMux.Unlock()

	hc.timeout = timeout
	for _, checker := range hc.checkers {
		checker.Timeout = timeout
	}

	hc.logger.Info("Updated health check timeout", "timeout", timeout)
}

// ForceHealthCheck forces an immediate health check for a specific server
func (hc *HealthChecker) ForceHealthCheck(serverID string) error {
	hc.checkersMux.RLock()
	defer hc.checkersMux.RUnlock()

	checker, exists := hc.checkers[serverID]
	if !exists {
		return fmt.Errorf("health checker not found for server %s", serverID)
	}

	// Perform immediate health check
	checker.performHealthCheck()

	hc.logger.Info("Forced health check for server", "server_id", serverID)
	return nil
}

// ForceHealthCheckAll forces an immediate health check for all servers
func (hc *HealthChecker) ForceHealthCheckAll() {
	hc.checkersMux.RLock()
	defer hc.checkersMux.RUnlock()

	for serverID, checker := range hc.checkers {
		checker.performHealthCheck()
		hc.logger.Debug("Forced health check for server", "server_id", serverID)
	}

	hc.logger.Info("Forced health check for all servers")
}
