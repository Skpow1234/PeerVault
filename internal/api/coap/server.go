package coap

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/app/fileserver"
)

// Server represents the CoAP server
type Server struct {
	fileserver *fileserver.Server
	config     *ServerConfig
	logger     *slog.Logger

	// Resource management
	resources   map[string]*Resource
	resourcesMu sync.RWMutex

	// Observation management
	observers   map[string][]*Observer
	observersMu sync.RWMutex

	// Client management
	clients   map[string]*Client
	clientsMu sync.RWMutex

	// Statistics
	stats   *ServerStats
	statsMu sync.RWMutex

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// ServerConfig holds the configuration for the CoAP server
type ServerConfig struct {
	Port           int
	Host           string
	EnableDTLS     bool
	DTLSPort       int
	MaxConnections int
	MaxMessageSize int
	BlockSize      int
	MaxAge         int
	EnableObserve  bool
	ObserveTimeout time.Duration
}

// ServerStats holds server statistics
type ServerStats struct {
	StartTime         time.Time
	TotalConnections  int
	ActiveConnections int
	TotalRequests     int
	TotalResponses    int
	TotalResources    int
	TotalObservers    int
	BytesReceived     int64
	BytesSent         int64
}

// NewServer creates a new CoAP server
func NewServer(fileserver *fileserver.Server, config *ServerConfig, logger *slog.Logger) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		fileserver: fileserver,
		config:     config,
		logger:     logger,
		resources:  make(map[string]*Resource),
		observers:  make(map[string][]*Observer),
		clients:    make(map[string]*Client),
		stats: &ServerStats{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Register default resources
	server.registerDefaultResources()

	// Start background tasks
	go server.startBackgroundTasks()

	return server
}

// ServeUDP starts the UDP CoAP server
func (s *Server) ServeUDP(ctx context.Context, conn *net.UDPConn) error {
	buffer := make([]byte, s.config.MaxMessageSize)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Set read timeout
		if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			s.logger.Error("Failed to set read deadline", "error", err)
			continue
		}

		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return err
		}

		// Handle message in goroutine
		go s.handleMessage(conn, clientAddr, buffer[:n])
	}
}

// ServeDTLS starts the DTLS CoAP server
func (s *Server) ServeDTLS(ctx context.Context, addr string) error {
	// TODO: Implement DTLS support
	// This would involve setting up a DTLS listener and handling encrypted connections
	s.logger.Info("DTLS support not yet implemented", "addr", addr)
	return nil
}

// handleMessage handles a received CoAP message
func (s *Server) handleMessage(conn *net.UDPConn, clientAddr *net.UDPAddr, data []byte) {
	// Parse CoAP message
	message, err := ParseMessage(data)
	if err != nil {
		s.logger.Error("Failed to parse CoAP message", "error", err, "client", clientAddr)
		return
	}

	// Update statistics
	s.updateStats(func(stats *ServerStats) {
		stats.TotalRequests++
		stats.BytesReceived += int64(len(data))
	})

	// Get or create client
	client := s.getOrCreateClient(clientAddr)

	// Handle the message
	response, err := s.handleRequest(message, client)
	if err != nil {
		s.logger.Error("Failed to handle CoAP request", "error", err, "client", clientAddr)
		// Send error response
		errorResponse := s.createErrorResponse(message, InternalServerError)
		if err := s.sendResponse(conn, clientAddr, errorResponse); err != nil {
			s.logger.Error("Failed to send error response", "error", err)
		}
		return
	}

	// Send response if not nil (some requests don't require responses)
	if response != nil {
		if err := s.sendResponse(conn, clientAddr, response); err != nil {
			s.logger.Error("Failed to send CoAP response", "error", err, "client", clientAddr)
		}
	}
}

// handleRequest handles a CoAP request
func (s *Server) handleRequest(message *Message, client *Client) (*Message, error) {
	// Find the resource
	resource, exists := s.getResource(message.GetPath())
	if !exists {
		return s.createErrorResponse(message, NotFound), nil
	}

	// Handle the request based on method
	switch MethodCode(message.Code) {
	case GET:
		return s.handleGet(message, resource, client)
	case POST:
		return s.handlePost(message, resource, client)
	case PUT:
		return s.handlePut(message, resource, client)
	case DELETE:
		return s.handleDelete(message, resource, client)
	default:
		return s.createErrorResponse(message, MethodNotAllowed), nil
	}
}

// handleGet handles GET requests
func (s *Server) handleGet(message *Message, resource *Resource, client *Client) (*Message, error) {
	// Check for observe option
	if message.HasOption(Observe) {
		return s.handleObserve(message, resource, client)
	}

	// Handle regular GET request
	response := s.createResponse(message, byte(Content), resource.GetContent())

	// Add content format option
	if resource.ContentFormat != nil {
		response.AddOption(ContentFormat, uint16(*resource.ContentFormat))
	}

	// Add max age option
	response.AddOption(MaxAge, uint32(s.config.MaxAge))

	return response, nil
}

// handlePost handles POST requests
func (s *Server) handlePost(message *Message, resource *Resource, client *Client) (*Message, error) {
	if resource.PostHandler == nil {
		return s.createErrorResponse(message, MethodNotAllowed), nil
	}

	// Call the POST handler
	response, err := resource.PostHandler(message, client)
	if err != nil {
		return s.createErrorResponse(message, InternalServerError), nil
	}

	return response, nil
}

// handlePut handles PUT requests
func (s *Server) handlePut(message *Message, resource *Resource, client *Client) (*Message, error) {
	if resource.PutHandler == nil {
		return s.createErrorResponse(message, MethodNotAllowed), nil
	}

	// Call the PUT handler
	response, err := resource.PutHandler(message, client)
	if err != nil {
		return s.createErrorResponse(message, InternalServerError), nil
	}

	return response, nil
}

// handleDelete handles DELETE requests
func (s *Server) handleDelete(message *Message, resource *Resource, client *Client) (*Message, error) {
	if resource.DeleteHandler == nil {
		return s.createErrorResponse(message, MethodNotAllowed), nil
	}

	// Call the DELETE handler
	response, err := resource.DeleteHandler(message, client)
	if err != nil {
		return s.createErrorResponse(message, InternalServerError), nil
	}

	return response, nil
}

// handleObserve handles observation requests
func (s *Server) handleObserve(message *Message, resource *Resource, client *Client) (*Message, error) {
	if !s.config.EnableObserve {
		return s.createErrorResponse(message, MethodNotAllowed), nil
	}

	// Create observer
	observer := &Observer{
		Client:    client,
		Token:     message.Token,
		Resource:  resource,
		CreatedAt: time.Now(),
	}

	// Add observer to resource
	s.addObserver(message.GetPath(), observer)

	// Send initial response
	response := s.createResponse(message, byte(Content), resource.GetContent())
	response.AddOption(Observe, uint32(time.Now().Unix()))

	if resource.ContentFormat != nil {
		response.AddOption(ContentFormat, uint16(*resource.ContentFormat))
	}

	return response, nil
}

// getOrCreateClient gets or creates a client
func (s *Server) getOrCreateClient(addr *net.UDPAddr) *Client {
	clientID := addr.String()

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	client, exists := s.clients[clientID]
	if !exists {
		client = &Client{
			ID:        clientID,
			Address:   addr,
			CreatedAt: time.Now(),
		}
		s.clients[clientID] = client
		s.updateStats(func(stats *ServerStats) {
			stats.TotalConnections++
			stats.ActiveConnections++
		})
	}

	client.LastSeen = time.Now()
	return client
}

// getResource gets a resource by path
func (s *Server) getResource(path string) (*Resource, bool) {
	s.resourcesMu.RLock()
	defer s.resourcesMu.RUnlock()

	resource, exists := s.resources[path]
	return resource, exists
}

// addObserver adds an observer to a resource
func (s *Server) addObserver(path string, observer *Observer) {
	s.observersMu.Lock()
	defer s.observersMu.Unlock()

	s.observers[path] = append(s.observers[path], observer)
	s.updateStats(func(stats *ServerStats) {
		stats.TotalObservers++
	})

	s.logger.Debug("Observer added",
		"path", path,
		"client", observer.Client.ID,
		"totalObservers", len(s.observers[path]),
	)
}

// registerDefaultResources registers default CoAP resources
func (s *Server) registerDefaultResources() {
	// Well-known core resource
	s.registerResource("/.well-known/core", &Resource{
		Name:          "Core",
		Description:   "Core resource for resource discovery",
		Content:       s.getCoreResourceContent(),
		ContentFormat: &[]CoAPContentFormat{ContentFormatApplicationLinkFormat}[0],
	})

	// Server info resource
	s.registerResource("/server", &Resource{
		Name:          "Server Info",
		Description:   "Server information and statistics",
		Content:       s.getServerInfoContent(),
		ContentFormat: &[]CoAPContentFormat{ContentFormatApplicationJSON}[0],
	})

	// Health check resource
	s.registerResource("/health", &Resource{
		Name:          "Health Check",
		Description:   "Server health status",
		Content:       []byte(`{"status":"healthy"}`),
		ContentFormat: &[]CoAPContentFormat{ContentFormatApplicationJSON}[0],
	})
}

// registerResource registers a resource
func (s *Server) registerResource(path string, resource *Resource) {
	s.resourcesMu.Lock()
	defer s.resourcesMu.Unlock()

	s.resources[path] = resource
	s.updateStats(func(stats *ServerStats) {
		stats.TotalResources++
	})

	s.logger.Debug("Resource registered", "path", path, "name", resource.Name)
}

// getCoreResourceContent returns the content for the core resource
func (s *Server) getCoreResourceContent() []byte {
	s.resourcesMu.RLock()
	defer s.resourcesMu.RUnlock()

	var links []string
	for path, resource := range s.resources {
		link := fmt.Sprintf("<%s>;ct=%d", path, uint16(*resource.ContentFormat))
		links = append(links, link)
	}

	return []byte(strings.Join(links, ","))
}

// getServerInfoContent returns the server information content
func (s *Server) getServerInfoContent() []byte {
	stats := s.GetStats()
	info := map[string]interface{}{
		"name":            "PeerVault CoAP Server",
		"version":         "1.0.0",
		"uptime":          time.Since(stats.StartTime).String(),
		"total_requests":  stats.TotalRequests,
		"total_resources": stats.TotalResources,
		"active_clients":  stats.ActiveConnections,
	}

	// Convert to JSON (simplified)
	return []byte(fmt.Sprintf(`{"name":"%s","version":"%s","uptime":"%s","total_requests":%d,"total_resources":%d,"active_clients":%d}`,
		info["name"], info["version"], info["uptime"],
		info["total_requests"], info["total_resources"], info["active_clients"]))
}

// createResponse creates a response message
func (s *Server) createResponse(request *Message, code byte, payload []byte) *Message {
	response := &Message{
		Type:      Acknowledgement,
		Code:      code,
		MessageID: request.MessageID,
		Token:     request.Token,
		Payload:   payload,
	}

	return response
}

// createErrorResponse creates an error response message
func (s *Server) createErrorResponse(request *Message, code ResponseCode) *Message {
	response := &Message{
		Type:      Acknowledgement,
		Code:      byte(code),
		MessageID: request.MessageID,
		Token:     request.Token,
		Payload:   []byte{},
	}

	return response
}

// sendResponse sends a response message
func (s *Server) sendResponse(conn *net.UDPConn, addr *net.UDPAddr, message *Message) error {
	data, err := message.Encode()
	if err != nil {
		return err
	}

	_, err = conn.WriteToUDP(data, addr)
	if err != nil {
		return err
	}

	// Update statistics
	s.updateStats(func(stats *ServerStats) {
		stats.TotalResponses++
		stats.BytesSent += int64(len(data))
	})

	return nil
}

// updateStats updates server statistics
func (s *Server) updateStats(updater func(*ServerStats)) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	updater(s.stats)
}

// GetStats returns server statistics
func (s *Server) GetStats() *ServerStats {
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()

	// Return a copy
	stats := *s.stats
	return &stats
}

// startBackgroundTasks starts background maintenance tasks
func (s *Server) startBackgroundTasks() {
	// Cleanup task
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.ctx.Done():
			return
		}
	}
}

// cleanup performs periodic cleanup tasks
func (s *Server) cleanup() {
	// Remove inactive clients
	s.clientsMu.Lock()
	for clientID, client := range s.clients {
		if time.Since(client.LastSeen) > 5*time.Minute {
			delete(s.clients, clientID)
			s.updateStats(func(stats *ServerStats) {
				stats.ActiveConnections--
			})
		}
	}
	s.clientsMu.Unlock()

	// Remove expired observers
	s.observersMu.Lock()
	for path, observers := range s.observers {
		var activeObservers []*Observer
		for _, observer := range observers {
			if time.Since(observer.CreatedAt) < s.config.ObserveTimeout {
				activeObservers = append(activeObservers, observer)
			}
		}
		s.observers[path] = activeObservers
	}
	s.observersMu.Unlock()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.logger.Info("Shutting down CoAP server...")
	s.cancel()

	// Close all client connections
	s.clientsMu.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMu.Unlock()

	s.logger.Info("CoAP server shutdown complete")
}
