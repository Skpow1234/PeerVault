package protocol

import (
	"context"
	"fmt"
	"sync"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/config"
	"github.com/Skpow1234/Peervault/internal/cli/graphql"
	"github.com/Skpow1234/Peervault/internal/cli/grpc"
)

// Type represents the protocol type
type Type string

const (
	REST    Type = "rest"
	GraphQL Type = "graphql"
	GRPC    Type = "grpc"
)

// Manager manages multiple protocol clients
type Manager struct {
	restClient    *client.Client
	graphqlClient *graphql.Client
	grpcClient    *grpc.Client
	currentType   Type
	mu            sync.RWMutex
}

// New creates a new protocol manager
func New(serverURL string) *Manager {
	config := &config.Config{ServerURL: serverURL}
	return &Manager{
		restClient:    client.New(config),
		graphqlClient: graphql.New(serverURL),
		grpcClient:    grpc.New(serverURL),
		currentType:   REST, // Default to REST
	}
}

// SetProtocol sets the current protocol
func (m *Manager) SetProtocol(protocol Type) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch protocol {
	case REST, GraphQL, GRPC:
		m.currentType = protocol
		return nil
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// GetProtocol returns the current protocol
func (m *Manager) GetProtocol() Type {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentType
}

// GetRESTClient returns the REST client
func (m *Manager) GetRESTClient() *client.Client {
	return m.restClient
}

// GetGraphQLClient returns the GraphQL client
func (m *Manager) GetGraphQLClient() *graphql.Client {
	return m.graphqlClient
}

// GetGRPCClient returns the gRPC client
func (m *Manager) GetGRPCClient() *grpc.Client {
	return m.grpcClient
}

// Connect connects using the current protocol
func (m *Manager) Connect(ctx context.Context) error {
	m.mu.RLock()
	protocol := m.currentType
	m.mu.RUnlock()

	switch protocol {
	case REST:
		return m.restClient.Connect(ctx)
	case GraphQL:
		// GraphQL uses HTTP, so we can use REST client for connection test
		return m.restClient.Connect(ctx)
	case GRPC:
		return m.grpcClient.Connect(ctx)
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// Disconnect disconnects from all protocols
func (m *Manager) Disconnect() error {
	var errs []error

	m.restClient.Disconnect()

	if err := m.grpcClient.Disconnect(); err != nil {
		errs = append(errs, fmt.Errorf("gRPC disconnect error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("disconnect errors: %v", errs)
	}

	return nil
}

// IsConnected returns whether the current protocol is connected
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	protocol := m.currentType
	m.mu.RUnlock()

	switch protocol {
	case REST:
		return m.restClient.IsConnected()
	case GraphQL:
		return m.restClient.IsConnected() // GraphQL uses HTTP
	case GRPC:
		return m.grpcClient.IsConnected()
	default:
		return false
	}
}

// SetServerURL sets the server URL for all protocols
func (m *Manager) SetServerURL(url string) {
	m.restClient.SetServerURL(url)
	m.graphqlClient = graphql.New(url)
	m.grpcClient.SetServer(url)
}

// SetAuthToken sets the authentication token for all protocols
func (m *Manager) SetAuthToken(token string) {
	m.restClient.SetAuthToken(token)
	m.graphqlClient.SetAuthToken(token)
	// gRPC would use different auth mechanism
}

// GetSupportedProtocols returns the list of supported protocols
func GetSupportedProtocols() []Type {
	return []Type{REST, GraphQL, GRPC}
}

// GetProtocolDescription returns a description of the protocol
func GetProtocolDescription(protocol Type) string {
	switch protocol {
	case REST:
		return "REST API - Standard HTTP/HTTPS API with JSON"
	case GraphQL:
		return "GraphQL - Query language for APIs with flexible data fetching"
	case GRPC:
		return "gRPC - High-performance RPC framework with Protocol Buffers"
	default:
		return "Unknown protocol"
	}
}

// GetProtocolFeatures returns features of the protocol
func GetProtocolFeatures(protocol Type) []string {
	switch protocol {
	case REST:
		return []string{
			"Simple HTTP requests",
			"JSON data format",
			"Standard HTTP status codes",
			"Easy to debug",
			"Wide browser support",
		}
	case GraphQL:
		return []string{
			"Single endpoint",
			"Flexible queries",
			"Real-time subscriptions",
			"Strong typing",
			"Introspection",
		}
	case GRPC:
		return []string{
			"High performance",
			"Binary protocol",
			"Streaming support",
			"Code generation",
			"Load balancing",
		}
	default:
		return []string{}
	}
}
