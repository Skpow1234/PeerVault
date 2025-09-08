package federation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// FederationGateway represents a GraphQL federation gateway
type FederationGateway struct {
	services   map[string]*FederatedService
	mu         sync.RWMutex
	logger     *slog.Logger
	httpClient *http.Client
}

// FederatedService represents a federated GraphQL service
type FederatedService struct {
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	Schema       string            `json:"schema"`
	HealthCheck  string            `json:"healthCheck,omitempty"`
	LastSeen     time.Time         `json:"lastSeen"`
	IsHealthy    bool              `json:"isHealthy"`
	Capabilities map[string]bool   `json:"capabilities"`
	Metadata     map[string]string `json:"metadata"`
}

// FederationConfig holds configuration for the federation gateway
type FederationConfig struct {
	GatewayPort         int           `json:"gatewayPort"`
	ServiceTimeout      time.Duration `json:"serviceTimeout"`
	HealthCheckInterval time.Duration `json:"healthCheckInterval"`
	EnableHealthChecks  bool          `json:"enableHealthChecks"`
}

// DefaultFederationConfig returns the default federation configuration
func DefaultFederationConfig() *FederationConfig {
	return &FederationConfig{
		GatewayPort:         8081,
		ServiceTimeout:      30 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		EnableHealthChecks:  true,
	}
}

// NewFederationGateway creates a new GraphQL federation gateway
func NewFederationGateway(logger *slog.Logger) *FederationGateway {
	return &FederationGateway{
		services: make(map[string]*FederatedService),
		logger:   logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegisterService registers a new federated service
func (fg *FederationGateway) RegisterService(service *FederatedService) error {
	fg.mu.Lock()
	defer fg.mu.Unlock()

	// Validate service
	if service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if service.URL == "" {
		return fmt.Errorf("service URL is required")
	}

	// Set default values
	if service.Capabilities == nil {
		service.Capabilities = make(map[string]bool)
	}
	if service.Metadata == nil {
		service.Metadata = make(map[string]string)
	}
	service.LastSeen = time.Now()
	service.IsHealthy = true

	fg.services[service.Name] = service
	fg.logger.Info("Registered federated service", "name", service.Name, "url", service.URL)

	return nil
}

// UnregisterService unregisters a federated service
func (fg *FederationGateway) UnregisterService(serviceName string) error {
	fg.mu.Lock()
	defer fg.mu.Unlock()

	if _, exists := fg.services[serviceName]; !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	delete(fg.services, serviceName)
	fg.logger.Info("Unregistered federated service", "name", serviceName)

	return nil
}

// GetService returns a federated service by name
func (fg *FederationGateway) GetService(serviceName string) (*FederatedService, bool) {
	fg.mu.RLock()
	defer fg.mu.RUnlock()
	service, exists := fg.services[serviceName]
	return service, exists
}

// ListServices returns all registered services
func (fg *FederationGateway) ListServices() []*FederatedService {
	fg.mu.RLock()
	defer fg.mu.RUnlock()

	services := make([]*FederatedService, 0, len(fg.services))
	for _, service := range fg.services {
		services = append(services, service)
	}
	return services
}

// GetHealthyServices returns all healthy services
func (fg *FederationGateway) GetHealthyServices() []*FederatedService {
	fg.mu.RLock()
	defer fg.mu.RUnlock()

	var healthyServices []*FederatedService
	for _, service := range fg.services {
		if service.IsHealthy {
			healthyServices = append(healthyServices, service)
		}
	}
	return healthyServices
}

// StartHealthChecks starts health checking for all services
func (fg *FederationGateway) StartHealthChecks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fg.logger.Info("Health checks stopped")
			return
		case <-ticker.C:
			fg.checkAllServicesHealth()
		}
	}
}

// checkAllServicesHealth checks the health of all registered services
func (fg *FederationGateway) checkAllServicesHealth() {
	fg.mu.RLock()
	services := make([]*FederatedService, 0, len(fg.services))
	for _, service := range fg.services {
		services = append(services, service)
	}
	fg.mu.RUnlock()

	for _, service := range services {
		go fg.checkServiceHealth(service)
	}
}

// checkServiceHealth checks the health of a specific service
func (fg *FederationGateway) checkServiceHealth(service *FederatedService) {
	healthURL := service.HealthCheck
	if healthURL == "" {
		healthURL = service.URL + "/health"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		fg.updateServiceHealth(service, false)
		return
	}

	resp, err := fg.httpClient.Do(req)
	if err != nil {
		fg.updateServiceHealth(service, false)
		return
	}
	defer resp.Body.Close()

	isHealthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	fg.updateServiceHealth(service, isHealthy)
}

// updateServiceHealth updates the health status of a service
func (fg *FederationGateway) updateServiceHealth(service *FederatedService, isHealthy bool) {
	fg.mu.Lock()
	defer fg.mu.Unlock()

	if service.IsHealthy != isHealthy {
		service.IsHealthy = isHealthy
		service.LastSeen = time.Now()

		status := "unhealthy"
		if isHealthy {
			status = "healthy"
		}

		fg.logger.Info("Service health status changed",
			"name", service.Name,
			"status", status,
			"url", service.URL)
	}
}

// ExecuteQuery executes a GraphQL query across federated services
func (fg *FederationGateway) ExecuteQuery(ctx context.Context, query string, variables map[string]interface{}) (*FederationResult, error) {
	// Parse the query to determine which services are needed
	requiredServices, err := fg.parseQueryForServices(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Check if all required services are healthy
	healthyServices := fg.GetHealthyServices()
	availableServices := make(map[string]bool)
	for _, service := range healthyServices {
		availableServices[service.Name] = true
	}

	for _, serviceName := range requiredServices {
		if !availableServices[serviceName] {
			return nil, fmt.Errorf("required service %s is not available", serviceName)
		}
	}

	// Execute the query
	result := &FederationResult{
		Data:   make(map[string]interface{}),
		Errors: []FederationError{},
	}

	// For now, execute on the first available service
	// In a real implementation, this would involve query planning and execution across multiple services
	if len(healthyServices) > 0 {
		service := healthyServices[0]
		serviceResult, err := fg.executeQueryOnService(ctx, service, query, variables)
		if err != nil {
			result.Errors = append(result.Errors, FederationError{
				Message: err.Error(),
				Service: service.Name,
			})
		} else {
			result.Data = serviceResult.Data
			result.Errors = append(result.Errors, serviceResult.Errors...)
		}
	}

	return result, nil
}

// parseQueryForServices parses a GraphQL query to determine which services are needed
func (fg *FederationGateway) parseQueryForServices(_ string) ([]string, error) {
	// This is a simplified implementation
	// In a real implementation, you would parse the GraphQL query and determine
	// which services are needed based on the schema and field resolvers

	// For now, return all available services
	fg.mu.RLock()
	services := make([]string, 0, len(fg.services))
	for serviceName := range fg.services {
		services = append(services, serviceName)
	}
	fg.mu.RUnlock()

	return services, nil
}

// executeQueryOnService executes a GraphQL query on a specific service
func (fg *FederationGateway) executeQueryOnService(ctx context.Context, service *FederatedService, query string, variables map[string]interface{}) (*FederationResult, error) {
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", service.URL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := fg.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var result FederationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// FederationResult represents the result of a federated GraphQL query
type FederationResult struct {
	Data   map[string]interface{} `json:"data"`
	Errors []FederationError      `json:"errors,omitempty"`
}

// FederationError represents an error from a federated service
type FederationError struct {
	Message    string                 `json:"message"`
	Service    string                 `json:"service,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GetFederatedSchema returns the combined schema from all services
func (fg *FederationGateway) GetFederatedSchema() (string, error) {
	fg.mu.RLock()
	defer fg.mu.RUnlock()

	// In a real implementation, this would introspect all services and combine their schemas
	// For now, return a basic federated schema
	schema := `
		type Query {
			# File operations
			file(key: String!): File
			files(limit: Int, offset: Int): [File!]!
			
			# Node operations
			node(id: ID!): Node
			nodes: [Node!]!
			
			# System operations
			systemMetrics: SystemMetrics!
		}

		type Mutation {
			# File operations
			uploadFile(key: String!, data: String!): File!
			deleteFile(key: String!): Boolean!
			
			# Node operations
			addPeer(address: String!, port: Int!): Node!
			removePeer(id: ID!): Boolean!
		}

		type Subscription {
			# Real-time events
			fileUploaded: File!
			peerConnected: Node!
			systemMetricsUpdated: SystemMetrics!
		}

		# Include all the types from the main schema
		scalar Time
		scalar Upload

		type File {
			id: ID!
			key: String!
			hashedKey: String!
			size: Int!
			createdAt: Time!
			updatedAt: Time!
			owner: Node!
			replicas: [FileReplica!]!
			metadata: FileMetadata
		}

		type FileMetadata {
			contentType: String
			checksum: String
			tags: [String!]
			customFields: String
		}

		type FileReplica {
			node: Node!
			status: ReplicaStatus!
			lastSync: Time
			size: Int
		}

		enum ReplicaStatus {
			SYNCED
			SYNCING
			FAILED
			PENDING
		}

		type Node {
			id: ID!
			address: String!
			port: Int!
			status: NodeStatus!
			lastSeen: Time
			health: NodeHealth
			capabilities: [String!]
		}

		type NodeHealth {
			isHealthy: Boolean!
			lastHeartbeat: Time
			responseTime: Float
			uptime: Float
			errors: [String!]
		}

		enum NodeStatus {
			ONLINE
			OFFLINE
			DEGRADED
			UNKNOWN
		}

		type SystemMetrics {
			storage: StorageMetrics!
			network: NetworkMetrics!
			performance: PerformanceMetrics!
			uptime: Float!
		}

		type StorageMetrics {
			totalSpace: Int!
			usedSpace: Int!
			availableSpace: Int!
			fileCount: Int!
			replicationFactor: Float
		}

		type NetworkMetrics {
			activeConnections: Int!
			totalBytesTransferred: Int!
			averageBandwidth: Float
			errorRate: Float
		}

		type PerformanceMetrics {
			averageResponseTime: Float
			requestsPerSecond: Float
			errorRate: Float
			memoryUsage: Float
			cpuUsage: Float
		}
	`

	return schema, nil
}

// GetServiceMetrics returns metrics for all services
func (fg *FederationGateway) GetServiceMetrics() map[string]interface{} {
	fg.mu.RLock()
	defer fg.mu.RUnlock()

	metrics := map[string]interface{}{
		"totalServices":     len(fg.services),
		"healthyServices":   0,
		"unhealthyServices": 0,
		"services":          make(map[string]interface{}),
	}

	services := make(map[string]interface{})
	for name, service := range fg.services {
		serviceMetrics := map[string]interface{}{
			"name":         service.Name,
			"url":          service.URL,
			"isHealthy":    service.IsHealthy,
			"lastSeen":     service.LastSeen,
			"capabilities": service.Capabilities,
			"metadata":     service.Metadata,
		}
		services[name] = serviceMetrics

		if service.IsHealthy {
			metrics["healthyServices"] = metrics["healthyServices"].(int) + 1
		} else {
			metrics["unhealthyServices"] = metrics["unhealthyServices"].(int) + 1
		}
	}

	metrics["services"] = services
	return metrics
}
