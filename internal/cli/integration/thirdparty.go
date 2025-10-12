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

// ThirdPartyIntegration represents a third-party service integration
type ThirdPartyIntegration struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Service      string                 `json:"service"` // aws, azure, gcp, slack, discord, etc.
	Type         string                 `json:"type"`    // storage, messaging, monitoring, auth
	Config       map[string]interface{} `json:"config"`
	Credentials  map[string]interface{} `json:"credentials"`
	IsActive     bool                   `json:"is_active"`
	LastSync     *time.Time             `json:"last_sync,omitempty"`
	SyncCount    int                    `json:"sync_count"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    string                 `json:"created_by"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// APIGateway represents an API gateway configuration
type APIGateway struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	BaseURL      string                 `json:"base_url"`
	Routes       []*APIRoute            `json:"routes"`
	Middleware   []*APIMiddleware       `json:"middleware"`
	RateLimit    *RateLimit             `json:"rate_limit"`
	Auth         *APIAuth               `json:"auth"`
	IsActive     bool                   `json:"is_active"`
	LastRequest  *time.Time             `json:"last_request,omitempty"`
	RequestCount int                    `json:"request_count"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    string                 `json:"created_by"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// APIRoute represents an API route
type APIRoute struct {
	ID         string                 `json:"id"`
	Path       string                 `json:"path"`
	Method     string                 `json:"method"` // GET, POST, PUT, DELETE
	Handler    string                 `json:"handler"`
	Middleware []string               `json:"middleware"`
	Auth       bool                   `json:"auth"`
	RateLimit  *RateLimit             `json:"rate_limit"`
	Config     map[string]interface{} `json:"config"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// APIMiddleware represents API middleware
type APIMiddleware struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"` // auth, rate_limit, logging, cors
	Config   map[string]interface{} `json:"config"`
	Order    int                    `json:"order"`
	IsActive bool                   `json:"is_active"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	Requests   int           `json:"requests"`
	Window     time.Duration `json:"window"`
	Burst      int           `json:"burst"`
	SkipOnFail bool          `json:"skip_on_fail"`
}

// APIAuth represents API authentication configuration
type APIAuth struct {
	Type      string                 `json:"type"` // api_key, jwt, oauth, basic
	Config    map[string]interface{} `json:"config"`
	Required  bool                   `json:"required"`
	SkipPaths []string               `json:"skip_paths"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// IntegrationManager manages third-party integrations and API gateway
type IntegrationManager struct {
	mu           sync.RWMutex
	client       *client.Client
	configDir    string
	integrations map[string]*ThirdPartyIntegration
	gateways     map[string]*APIGateway
	stats        *IntegrationStats
}

// IntegrationStats represents integration statistics
type IntegrationStats struct {
	TotalIntegrations  int       `json:"total_integrations"`
	ActiveIntegrations int       `json:"active_integrations"`
	TotalGateways      int       `json:"total_gateways"`
	ActiveGateways     int       `json:"active_gateways"`
	TotalRequests      int       `json:"total_requests"`
	SuccessfulRequests int       `json:"successful_requests"`
	FailedRequests     int       `json:"failed_requests"`
	LastUpdated        time.Time `json:"last_updated"`
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(client *client.Client, configDir string) *IntegrationManager {
	im := &IntegrationManager{
		client:       client,
		configDir:    configDir,
		integrations: make(map[string]*ThirdPartyIntegration),
		gateways:     make(map[string]*APIGateway),
		stats:        &IntegrationStats{},
	}
	_ = im.loadIntegrations() // Ignore error for initialization
	_ = im.loadGateways()     // Ignore error for initialization
	_ = im.loadStats()        // Ignore error for initialization
	return im
}

// CreateIntegration creates a new third-party integration
func (im *IntegrationManager) CreateIntegration(ctx context.Context, name, description, service, integrationType, createdBy string, config, credentials map[string]interface{}) (*ThirdPartyIntegration, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	integration := &ThirdPartyIntegration{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		Service:      service,
		Type:         integrationType,
		Config:       config,
		Credentials:  credentials,
		IsActive:     true,
		SyncCount:    0,
		SuccessCount: 0,
		FailureCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    createdBy,
		Metadata:     make(map[string]interface{}),
	}

	im.integrations[integration.ID] = integration

	// Simulate API call - store integration data as JSON
	integrationData, err := json.Marshal(integration)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal integration: %v", err)
	}

	tempFilePath := filepath.Join(im.configDir, fmt.Sprintf("integrations/%s.json", integration.ID))
	if err := os.WriteFile(tempFilePath, integrationData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write integration data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = im.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store integration: %v", err)
	}

	im.stats.TotalIntegrations++
	im.stats.ActiveIntegrations++
	_ = im.saveStats()
	_ = im.saveIntegrations()
	return integration, nil
}

// ListIntegrations returns all integrations
func (im *IntegrationManager) ListIntegrations(ctx context.Context) ([]*ThirdPartyIntegration, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	integrations := make([]*ThirdPartyIntegration, 0, len(im.integrations))
	for _, integration := range im.integrations {
		integrations = append(integrations, integration)
	}
	return integrations, nil
}

// GetIntegration returns an integration by ID
func (im *IntegrationManager) GetIntegration(ctx context.Context, integrationID string) (*ThirdPartyIntegration, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	integration, exists := im.integrations[integrationID]
	if !exists {
		return nil, fmt.Errorf("integration not found: %s", integrationID)
	}
	return integration, nil
}

// SyncIntegration syncs data with a third-party service
func (im *IntegrationManager) SyncIntegration(ctx context.Context, integrationID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	integration, exists := im.integrations[integrationID]
	if !exists {
		return fmt.Errorf("integration not found: %s", integrationID)
	}

	if !integration.IsActive {
		return fmt.Errorf("integration is not active: %s", integrationID)
	}

	// Simulate sync operation
	time.Sleep(200 * time.Millisecond)

	now := time.Now()
	integration.LastSync = &now
	integration.SyncCount++
	integration.SuccessCount++
	integration.UpdatedAt = now

	im.stats.TotalRequests++
	im.stats.SuccessfulRequests++

	_ = im.saveIntegrations()
	_ = im.saveStats()
	return nil
}

// CreateAPIGateway creates a new API gateway
func (im *IntegrationManager) CreateAPIGateway(ctx context.Context, name, description, baseURL, createdBy string, routes []*APIRoute, middleware []*APIMiddleware, rateLimit *RateLimit, auth *APIAuth) (*APIGateway, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	gateway := &APIGateway{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		BaseURL:      baseURL,
		Routes:       routes,
		Middleware:   middleware,
		RateLimit:    rateLimit,
		Auth:         auth,
		IsActive:     true,
		RequestCount: 0,
		SuccessCount: 0,
		FailureCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    createdBy,
		Metadata:     make(map[string]interface{}),
	}

	im.gateways[gateway.ID] = gateway

	// Simulate API call - store gateway data as JSON
	gatewayData, err := json.Marshal(gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gateway: %v", err)
	}

	tempFilePath := filepath.Join(im.configDir, fmt.Sprintf("gateways/%s.json", gateway.ID))
	if err := os.WriteFile(tempFilePath, gatewayData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write gateway data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = im.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store gateway: %v", err)
	}

	im.stats.TotalGateways++
	im.stats.ActiveGateways++
	_ = im.saveStats()
	_ = im.saveGateways()
	return gateway, nil
}

// ListGateways returns all API gateways
func (im *IntegrationManager) ListGateways(ctx context.Context) ([]*APIGateway, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	gateways := make([]*APIGateway, 0, len(im.gateways))
	for _, gateway := range im.gateways {
		gateways = append(gateways, gateway)
	}
	return gateways, nil
}

// GetGateway returns a gateway by ID
func (im *IntegrationManager) GetGateway(ctx context.Context, gatewayID string) (*APIGateway, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	gateway, exists := im.gateways[gatewayID]
	if !exists {
		return nil, fmt.Errorf("gateway not found: %s", gatewayID)
	}
	return gateway, nil
}

// ProcessRequest processes a request through the API gateway
func (im *IntegrationManager) ProcessRequest(ctx context.Context, gatewayID, path, method string, headers map[string]string, body []byte) (map[string]interface{}, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	gateway, exists := im.gateways[gatewayID]
	if !exists {
		return nil, fmt.Errorf("gateway not found: %s", gatewayID)
	}

	if !gateway.IsActive {
		return nil, fmt.Errorf("gateway is not active: %s", gatewayID)
	}

	// Find matching route
	var matchedRoute *APIRoute
	for _, route := range gateway.Routes {
		if route.Path == path && route.Method == method {
			matchedRoute = route
			break
		}
	}

	if matchedRoute == nil {
		return nil, fmt.Errorf("no route found for %s %s", method, path)
	}

	// Simulate request processing
	time.Sleep(50 * time.Millisecond)

	now := time.Now()
	gateway.LastRequest = &now
	gateway.RequestCount++
	gateway.SuccessCount++
	gateway.UpdatedAt = now

	im.stats.TotalRequests++
	im.stats.SuccessfulRequests++

	_ = im.saveGateways()
	_ = im.saveStats()

	// Return simulated response
	response := map[string]interface{}{
		"status":    "success",
		"message":   "Request processed successfully",
		"route":     matchedRoute.ID,
		"timestamp": now.Unix(),
	}

	return response, nil
}

// UpdateIntegrationStatus updates integration status
func (im *IntegrationManager) UpdateIntegrationStatus(ctx context.Context, integrationID string, isActive bool) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	integration, exists := im.integrations[integrationID]
	if !exists {
		return fmt.Errorf("integration not found: %s", integrationID)
	}

	integration.IsActive = isActive
	integration.UpdatedAt = time.Now()

	if isActive {
		im.stats.ActiveIntegrations++
	} else {
		im.stats.ActiveIntegrations--
	}

	_ = im.saveIntegrations()
	_ = im.saveStats()
	return nil
}

// GetIntegrationStats returns integration statistics
func (im *IntegrationManager) GetIntegrationStats(ctx context.Context) (*IntegrationStats, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Update stats
	im.stats.LastUpdated = time.Now()
	return im.stats, nil
}

// File operations
func (im *IntegrationManager) loadIntegrations() error {
	integrationsFile := filepath.Join(im.configDir, "integrations.json")
	if _, err := os.Stat(integrationsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(integrationsFile)
	if err != nil {
		return fmt.Errorf("failed to read integrations file: %w", err)
	}

	var integrations []*ThirdPartyIntegration
	if err := json.Unmarshal(data, &integrations); err != nil {
		return fmt.Errorf("failed to unmarshal integrations: %w", err)
	}

	for _, integration := range integrations {
		im.integrations[integration.ID] = integration
		if integration.IsActive {
			im.stats.ActiveIntegrations++
		}
	}
	return nil
}

func (im *IntegrationManager) saveIntegrations() error {
	integrationsFile := filepath.Join(im.configDir, "integrations.json")

	var integrations []*ThirdPartyIntegration
	for _, integration := range im.integrations {
		integrations = append(integrations, integration)
	}

	data, err := json.MarshalIndent(integrations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal integrations: %w", err)
	}

	return os.WriteFile(integrationsFile, data, 0644)
}

func (im *IntegrationManager) loadGateways() error {
	gatewaysFile := filepath.Join(im.configDir, "gateways.json")
	if _, err := os.Stat(gatewaysFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(gatewaysFile)
	if err != nil {
		return fmt.Errorf("failed to read gateways file: %w", err)
	}

	var gateways []*APIGateway
	if err := json.Unmarshal(data, &gateways); err != nil {
		return fmt.Errorf("failed to unmarshal gateways: %w", err)
	}

	for _, gateway := range gateways {
		im.gateways[gateway.ID] = gateway
		if gateway.IsActive {
			im.stats.ActiveGateways++
		}
	}
	return nil
}

func (im *IntegrationManager) saveGateways() error {
	gatewaysFile := filepath.Join(im.configDir, "gateways.json")

	var gateways []*APIGateway
	for _, gateway := range im.gateways {
		gateways = append(gateways, gateway)
	}

	data, err := json.MarshalIndent(gateways, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal gateways: %w", err)
	}

	return os.WriteFile(gatewaysFile, data, 0644)
}

func (im *IntegrationManager) loadStats() error {
	statsFile := filepath.Join(im.configDir, "integration_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats IntegrationStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	im.stats = &stats
	return nil
}

func (im *IntegrationManager) saveStats() error {
	statsFile := filepath.Join(im.configDir, "integration_stats.json")

	data, err := json.MarshalIndent(im.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
