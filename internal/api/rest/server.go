package rest

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/rest/endpoints"
	"github.com/Skpow1234/Peervault/internal/api/rest/implementations"
	"github.com/Skpow1234/Peervault/internal/api/rest/ratelimit"
	"github.com/Skpow1234/Peervault/internal/api/rest/versioning"
)

type Server struct {
	config          *Config
	logger          *slog.Logger
	httpServer      *http.Server
	rateLimiter     *ratelimit.RateLimiter
	FileEndpoints   *endpoints.FileEndpoints
	PeerEndpoints   *endpoints.PeerEndpoints
	SystemEndpoints *endpoints.SystemEndpoints
}

type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxHeaderBytes  int
	AllowedOrigins  []string
	RateLimitPerMin int
	AuthToken       string
	VersionConfig   *versioning.VersionConfig
	RateLimitConfig *ratelimit.RateLimitConfig
}

func DefaultConfig() *Config {
	return &Config{
		Port:            ":8081",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		MaxHeaderBytes:  1 << 20,
		AllowedOrigins:  []string{"*"},
		RateLimitPerMin: 100,
		AuthToken:       "demo-token",
		VersionConfig:   versioning.NewVersionConfig(),
		RateLimitConfig: ratelimit.DefaultConfig(),
	}
}

func NewServer(config *Config, logger *slog.Logger) *Server {
	// Initialize services
	fileService := implementations.NewFileService()
	peerService := implementations.NewPeerService()
	systemService := implementations.NewSystemService()

	// Initialize rate limiter
	rateLimiter := ratelimit.NewRateLimiter(config.RateLimitConfig)

	// Initialize endpoints
	fileEndpoints := endpoints.NewFileEndpoints(fileService, logger)
	peerEndpoints := endpoints.NewPeerEndpoints(peerService, logger)
	systemEndpoints := endpoints.NewSystemEndpoints(systemService, logger)

	return &Server{
		config:          config,
		logger:          logger,
		rateLimiter:     rateLimiter,
		FileEndpoints:   fileEndpoints,
		PeerEndpoints:   peerEndpoints,
		SystemEndpoints: systemEndpoints,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Apply middleware
	versionMiddleware := versioning.VersionMiddleware(s.config.VersionConfig)
	rateLimitMiddleware := s.rateLimiter.Middleware()
	handler := s.CORSMiddleware(versionMiddleware(rateLimitMiddleware(s.authMiddleware(s.loggingMiddleware(mux)))))

	// API routes
	api := http.NewServeMux()
	api.HandleFunc("GET /files", s.FileEndpoints.HandleListFiles)
	api.HandleFunc("GET /files/get", s.FileEndpoints.HandleGetFile)
	api.HandleFunc("POST /files", s.FileEndpoints.HandleUploadFile)
	api.HandleFunc("DELETE /files", s.FileEndpoints.HandleDeleteFile)
	api.HandleFunc("PUT /files/metadata", s.FileEndpoints.HandleUpdateFileMetadata)

	api.HandleFunc("GET /peers", s.PeerEndpoints.HandleListPeers)
	api.HandleFunc("GET /peers/get", s.PeerEndpoints.HandleGetPeer)
	api.HandleFunc("POST /peers", s.PeerEndpoints.HandleAddPeer)
	api.HandleFunc("DELETE /peers", s.PeerEndpoints.HandleRemovePeer)

	// System routes
	mux.HandleFunc("GET /health", s.SystemEndpoints.HandleHealth)
	mux.HandleFunc("GET /metrics", s.SystemEndpoints.HandleMetrics)
	mux.HandleFunc("GET /system", s.SystemEndpoints.HandleSystemInfo)
	mux.HandleFunc("POST /webhook", s.SystemEndpoints.HandleWebhook)
	mux.HandleFunc("GET /api", s.SystemEndpoints.HandleRoot)
	mux.HandleFunc("GET /docs", s.SystemEndpoints.HandleDocs)
	mux.HandleFunc("GET /swagger.json", s.SystemEndpoints.HandleSwaggerJSON)

	// Mount API under /api/v1
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	s.httpServer = &http.Server{
		Addr:           s.config.Port,
		Handler:        handler,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}

	s.logger.Info("Starting REST API server", "port", s.config.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	// Stop rate limiter
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, allowedOrigin := range s.config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check and docs
		if r.URL.Path == "/health" || r.URL.Path == "/docs" || r.URL.Path == "/swagger.json" || r.URL.Path == "/api" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		expectedToken := "Bearer " + s.config.AuthToken
		if authHeader != expectedToken {
			http.Error(w, "Invalid authorization token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		s.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"duration", duration,
		)
	})
}
