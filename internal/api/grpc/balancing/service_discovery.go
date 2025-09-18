package balancing

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ServiceDiscoveryConfig represents service discovery configuration
type ServiceDiscoveryConfig struct {
	Provider        string // "static", "consul", "etcd", "kubernetes"
	RefreshInterval time.Duration
	Timeout         time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
}

// DefaultServiceDiscoveryConfig returns the default service discovery configuration
func DefaultServiceDiscoveryConfig() *ServiceDiscoveryConfig {
	return &ServiceDiscoveryConfig{
		Provider:        "static",
		RefreshInterval: 30 * time.Second,
		Timeout:         10 * time.Second,
		MaxRetries:      3,
		RetryDelay:      time.Second,
	}
}

// ServiceDiscovery provides service discovery functionality
type ServiceDiscovery struct {
	config      *ServiceDiscoveryConfig
	logger      *slog.Logger
	services    map[string]*Service
	servicesMux sync.RWMutex
	providers   map[string]ServiceProvider
	stopChan    chan struct{}
}

// Service represents a discovered service
type Service struct {
	Name      string
	Instances []*ServiceInstance
	LastSeen  time.Time
	mutex     sync.RWMutex
}

// ServiceInstance represents a service instance
type ServiceInstance struct {
	ID       string
	Address  string
	Port     int
	Weight   int
	Tags     map[string]string
	Health   HealthStatus
	LastSeen time.Time
}

// ServiceProvider interface for different service discovery providers
type ServiceProvider interface {
	DiscoverServices(ctx context.Context) (map[string]*Service, error)
	RegisterService(ctx context.Context, service *Service) error
	DeregisterService(ctx context.Context, serviceID string) error
	WatchServices(ctx context.Context, callback func(*Service)) error
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(logger *slog.Logger) *ServiceDiscovery {
	if logger == nil {
		logger = slog.Default()
	}

	return &ServiceDiscovery{
		config:    DefaultServiceDiscoveryConfig(),
		logger:    logger,
		services:  make(map[string]*Service),
		providers: make(map[string]ServiceProvider),
		stopChan:  make(chan struct{}),
	}
}

// SetConfig sets the service discovery configuration
func (sd *ServiceDiscovery) SetConfig(config *ServiceDiscoveryConfig) {
	sd.config = config
}

// RegisterProvider registers a service discovery provider
func (sd *ServiceDiscovery) RegisterProvider(name string, provider ServiceProvider) {
	sd.providers[name] = provider
	sd.logger.Info("Registered service discovery provider", "provider", name)
}

// Start starts the service discovery
func (sd *ServiceDiscovery) Start() error {
	sd.logger.Info("Starting service discovery", "provider", sd.config.Provider)

	// Register default providers
	sd.registerDefaultProviders()

	// Start discovery loop
	go sd.discoveryLoop()

	return nil
}

// Stop stops the service discovery
func (sd *ServiceDiscovery) Stop() {
	sd.logger.Info("Stopping service discovery")
	close(sd.stopChan)
}

// discoveryLoop runs the service discovery loop
func (sd *ServiceDiscovery) discoveryLoop() {
	ticker := time.NewTicker(sd.config.RefreshInterval)
	defer ticker.Stop()

	// Perform initial discovery
	sd.discoverServices()

	for {
		select {
		case <-ticker.C:
			sd.discoverServices()
		case <-sd.stopChan:
			sd.logger.Info("Service discovery loop stopped")
			return
		}
	}
}

// discoverServices discovers services using the configured provider
func (sd *ServiceDiscovery) discoverServices() {
	provider, exists := sd.providers[sd.config.Provider]
	if !exists {
		sd.logger.Error("Service discovery provider not found", "provider", sd.config.Provider)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), sd.config.Timeout)
	defer cancel()

	services, err := provider.DiscoverServices(ctx)
	if err != nil {
		sd.logger.Error("Failed to discover services", "error", err)
		return
	}

	sd.servicesMux.Lock()
	defer sd.servicesMux.Unlock()

	// Update services
	for name, service := range services {
		sd.services[name] = service
		sd.logger.Debug("Discovered service", "service_name", name, "instances", len(service.Instances))
	}

	sd.logger.Info("Service discovery completed", "services", len(services))
}

// GetService returns a service by name
func (sd *ServiceDiscovery) GetService(name string) (*Service, bool) {
	sd.servicesMux.RLock()
	defer sd.servicesMux.RUnlock()

	service, exists := sd.services[name]
	return service, exists
}

// GetServices returns all discovered services
func (sd *ServiceDiscovery) GetServices() map[string]*Service {
	sd.servicesMux.RLock()
	defer sd.servicesMux.RUnlock()

	services := make(map[string]*Service)
	for name, service := range sd.services {
		services[name] = service
	}

	return services
}

// GetServiceInstances returns instances for a specific service
func (sd *ServiceDiscovery) GetServiceInstances(serviceName string) ([]*ServiceInstance, error) {
	service, exists := sd.GetService(serviceName)
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	service.mutex.RLock()
	defer service.mutex.RUnlock()

	instances := make([]*ServiceInstance, len(service.Instances))
	copy(instances, service.Instances)

	return instances, nil
}

// GetHealthyServiceInstances returns healthy instances for a specific service
func (sd *ServiceDiscovery) GetHealthyServiceInstances(serviceName string) ([]*ServiceInstance, error) {
	instances, err := sd.GetServiceInstances(serviceName)
	if err != nil {
		return nil, err
	}

	var healthyInstances []*ServiceInstance
	for _, instance := range instances {
		if instance.Health == HealthStatusHealthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	return healthyInstances, nil
}

// RegisterService registers a service with the service discovery
func (sd *ServiceDiscovery) RegisterService(service *Service) error {
	provider, exists := sd.providers[sd.config.Provider]
	if !exists {
		return fmt.Errorf("service discovery provider not found: %s", sd.config.Provider)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sd.config.Timeout)
	defer cancel()

	err := provider.RegisterService(ctx, service)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	sd.logger.Info("Registered service", "service_name", service.Name)
	return nil
}

// DeregisterService deregisters a service from the service discovery
func (sd *ServiceDiscovery) DeregisterService(serviceID string) error {
	provider, exists := sd.providers[sd.config.Provider]
	if !exists {
		return fmt.Errorf("service discovery provider not found: %s", sd.config.Provider)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sd.config.Timeout)
	defer cancel()

	err := provider.DeregisterService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	sd.logger.Info("Deregistered service", "service_id", serviceID)
	return nil
}

// GetStats returns service discovery statistics
func (sd *ServiceDiscovery) GetStats() map[string]interface{} {
	sd.servicesMux.RLock()
	defer sd.servicesMux.RUnlock()

	stats := make(map[string]interface{})
	stats["provider"] = sd.config.Provider
	stats["total_services"] = len(sd.services)
	stats["refresh_interval"] = sd.config.RefreshInterval
	stats["timeout"] = sd.config.Timeout

	totalInstances := 0
	healthyInstances := 0
	unhealthyInstances := 0

	serviceStats := make([]map[string]interface{}, 0)
	for name, service := range sd.services {
		service.mutex.RLock()
		instanceCount := len(service.Instances)
		service.mutex.RUnlock()

		serviceStat := map[string]interface{}{
			"name":      name,
			"instances": instanceCount,
			"last_seen": service.LastSeen,
		}
		serviceStats = append(serviceStats, serviceStat)

		totalInstances += instanceCount

		// Count healthy/unhealthy instances
		service.mutex.RLock()
		for _, instance := range service.Instances {
			switch instance.Health {
			case HealthStatusHealthy:
				healthyInstances++
			case HealthStatusUnhealthy:
				unhealthyInstances++
			}
		}
		service.mutex.RUnlock()
	}

	stats["total_instances"] = totalInstances
	stats["healthy_instances"] = healthyInstances
	stats["unhealthy_instances"] = unhealthyInstances
	stats["services"] = serviceStats

	return stats
}

// registerDefaultProviders registers default service discovery providers
func (sd *ServiceDiscovery) registerDefaultProviders() {
	// Register static provider
	sd.RegisterProvider("static", NewStaticServiceProvider(sd.logger))

	// Register Consul provider (if available)
	// sd.RegisterProvider("consul", NewConsulServiceProvider(sd.logger))

	// Register etcd provider (if available)
	// sd.RegisterProvider("etcd", NewEtcdServiceProvider(sd.logger))

	// Register Kubernetes provider (if available)
	// sd.RegisterProvider("kubernetes", NewKubernetesServiceProvider(sd.logger))
}

// StaticServiceProvider provides static service discovery
type StaticServiceProvider struct {
	logger   *slog.Logger
	services map[string]*Service
}

// NewStaticServiceProvider creates a new static service provider
func NewStaticServiceProvider(logger *slog.Logger) *StaticServiceProvider {
	return &StaticServiceProvider{
		logger:   logger,
		services: make(map[string]*Service),
	}
}

// DiscoverServices discovers services from static configuration
func (sp *StaticServiceProvider) DiscoverServices(ctx context.Context) (map[string]*Service, error) {
	services := make(map[string]*Service)
	for name, service := range sp.services {
		services[name] = service
	}
	return services, nil
}

// RegisterService registers a service with the static provider
func (sp *StaticServiceProvider) RegisterService(ctx context.Context, service *Service) error {
	sp.services[service.Name] = service
	sp.logger.Info("Registered service with static provider", "service_name", service.Name)
	return nil
}

// DeregisterService deregisters a service from the static provider
func (sp *StaticServiceProvider) DeregisterService(ctx context.Context, serviceID string) error {
	delete(sp.services, serviceID)
	sp.logger.Info("Deregistered service from static provider", "service_id", serviceID)
	return nil
}

// WatchServices watches for service changes (not implemented for static provider)
func (sp *StaticServiceProvider) WatchServices(ctx context.Context, callback func(*Service)) error {
	// Static provider doesn't support watching
	return fmt.Errorf("watching not supported for static provider")
}

// AddStaticService adds a static service
func (sp *StaticServiceProvider) AddStaticService(name string, instances []*ServiceInstance) {
	service := &Service{
		Name:      name,
		Instances: instances,
		LastSeen:  time.Now(),
	}
	sp.services[name] = service
	sp.logger.Info("Added static service", "service_name", name, "instances", len(instances))
}
