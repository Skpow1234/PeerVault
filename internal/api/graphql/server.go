package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/graphql/subscriptions"
	"github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/websocket"
)

// Server represents the GraphQL server
type Server struct {
	server               *fileserver.Server
	config               *Config
	logger               *slog.Logger
	hub                  *websocket.Hub
	subscriptionManager  *websocket.SubscriptionManager
	subscriptionResolver *subscriptions.SubscriptionResolver
}

// Config holds the configuration for the GraphQL server
type Config struct {
	Port             int
	PlaygroundPath   string
	GraphQLPath      string
	WebSocketPath    string
	AllowedOrigins   []string
	EnablePlayground bool
	EnableWebSocket  bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:             8080,
		PlaygroundPath:   "/playground",
		GraphQLPath:      "/graphql",
		WebSocketPath:    "/ws",
		AllowedOrigins:   []string{"*"},
		EnablePlayground: true,
		EnableWebSocket:  true,
	}
}

// NewServer creates a new GraphQL server
func NewServer(fileserver *fileserver.Server, config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := websocket.NewHub(logger)
	subscriptionManager := websocket.NewSubscriptionManager(hub, logger)
	subscriptionResolver := subscriptions.NewSubscriptionResolver(hub, subscriptionManager, logger)

	server := &Server{
		server:               fileserver,
		config:               config,
		logger:               logger,
		hub:                  hub,
		subscriptionManager:  subscriptionManager,
		subscriptionResolver: subscriptionResolver,
	}

	// Start the WebSocket hub
	go hub.Run(context.Background())

	return server
}

// Start starts the GraphQL server
func (s *Server) Start(config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}

	mux := http.NewServeMux()

	// GraphQL endpoint
	mux.HandleFunc(config.GraphQLPath, s.CORSMiddleware(s.GraphQLHandler))

	// WebSocket endpoint for GraphQL subscriptions
	if config.EnableWebSocket {
		wsHandler := websocket.NewGraphQLSubscriptionHandler(s.hub, s.logger)
		mux.HandleFunc(config.WebSocketPath, s.CORSMiddleware(wsHandler.ServeHTTP))
	}

	// GraphQL Playground
	if config.EnablePlayground {
		mux.HandleFunc(config.PlaygroundPath, s.CORSMiddleware(s.PlaygroundHandler))
	}

	// Health check endpoint
	mux.HandleFunc("/health", s.HealthHandler)

	// Metrics endpoint
	mux.HandleFunc("/metrics", s.MetricsHandler)

	s.logger.Info("Starting GraphQL server",
		"port", config.Port,
		"playground", config.PlaygroundPath,
		"graphql", config.GraphQLPath,
		"websocket", config.WebSocketPath,
		"websocketEnabled", config.EnableWebSocket,
	)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.Port),
		Handler:           mux,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return server.ListenAndServe()
}

// CORSMiddleware adds CORS headers to responses
func (s *Server) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []ErrorLocation        `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// ErrorLocation represents the location of an error
type ErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// GraphQLHandler handles GraphQL requests
func (s *Server) GraphQLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// For now, return a simple response
	response := GraphQLResponse{
		Data: map[string]interface{}{
			"health": map[string]interface{}{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// PlaygroundHandler serves the GraphQL Playground
func (s *Server) PlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	playgroundHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>GraphQL Playground</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.42/build/static/css/index.css" />
    <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.42/build/static/js/middleware.js"></script>
</head>
<body>
    <div id="root"></div>
    <script>
        window.addEventListener('load', function (event) {
            GraphQLPlayground.init(document.getElementById('root'), {
                endpoint: '/graphql'
            })
        })
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(playgroundHTML)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// HealthHandler handles health check requests
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "peervault-graphql",
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// MetricsHandler handles metrics requests
func (s *Server) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// TODO: Implement actual metrics collection
	metrics := map[string]interface{}{
		"uptime":    time.Since(time.Now()).Seconds(),
		"requests":  0,
		"errors":    0,
		"timestamp": time.Now().UTC(),
	}

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleFileUpload handles file upload requests
func (s *Server) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Error("Failed to close file", "error", err)
		}
	}()

	// Get key from form
	key := r.FormValue("key")
	if key == "" {
		key = header.Filename
	}

	// TODO: Implement actual file upload logic using the fileserver
	s.logger.Info("File upload request",
		"filename", header.Filename,
		"size", header.Size,
		"key", key,
	)

	// For now, just return success
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"key":     key,
		"size":    header.Size,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleFileDownload handles file download requests
func (s *Server) HandleFileDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter required", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual file download logic using the fileserver
	s.logger.Info("File download request", "key", key)

	// For now, return not found
	http.Error(w, "File not found", http.StatusNotFound)
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	s.logger.Info("Stopping GraphQL server")
	// TODO: Implement graceful shutdown
	return nil
}
