package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/grpc/services"
	"github.com/Skpow1234/Peervault/internal/api/grpc/types"
)

type Server struct {
	config        *Config
	logger        *slog.Logger
	fileService   *services.FileService
	peerService   *services.PeerService
	systemService *services.SystemService
	listener      net.Listener
}

type Config struct {
	Port                 string
	AuthToken            string
	MaxConcurrentStreams int
}

func DefaultConfig() *Config {
	return &Config{
		Port:                 ":8082",
		AuthToken:            "demo-token",
		MaxConcurrentStreams: 100,
	}
}

func NewServer(config *Config, logger *slog.Logger) *Server {
	// Initialize services
	fileService := services.NewFileService(logger)
	peerService := services.NewPeerService(logger)
	systemService := services.NewSystemService(logger)

	return &Server{
		config:        config,
		logger:        logger,
		fileService:   fileService,
		peerService:   peerService,
		systemService: systemService,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.config.Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.listener = lis
	s.logger.Info("Starting gRPC server", "port", s.config.Port)

	// For now, just log that the server is ready
	// TODO: Implement actual gRPC server with protobuf
	s.logger.Info("gRPC server ready (protobuf implementation pending)", "port", s.config.Port)

	// Keep the server running
	select {}
}

func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	s.logger.Info("gRPC server stopped")
}

// Mock implementations for now - these will be replaced with actual gRPC methods
func (s *Server) HealthCheck(ctx context.Context) (*types.HealthResponse, error) {
	return &types.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}, nil
}

func (s *Server) GetSystemInfo(ctx context.Context) (*types.SystemInfoResponse, error) {
	return &types.SystemInfoResponse{
		Version:       "1.0.0",
		UptimeSeconds: int64(time.Since(time.Now().Add(-time.Hour)).Seconds()),
		StartTime:     time.Now().Add(-time.Hour),
		StorageUsed:   1024 * 1024,      // 1MB
		StorageTotal:  10 * 1024 * 1024, // 10MB
		PeerCount:     3,
		FileCount:     5,
	}, nil
}

func (s *Server) GetMetrics(ctx context.Context) (*types.MetricsResponse, error) {
	return &types.MetricsResponse{
		RequestsTotal:       100,
		RequestsPerMinute:   10.5,
		ActiveConnections:   5,
		StorageUsagePercent: 10.0,
		LastUpdated:         time.Now(),
	}, nil
}
