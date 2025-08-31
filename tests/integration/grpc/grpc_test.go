package grpc_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/Skpow1234/Peervault/internal/api/grpc"
)

func TestGRPCServerCreation(t *testing.T) {
	config := grpc.DefaultConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := grpc.NewServer(config, logger)

	if server == nil {
		t.Fatal("Failed to create gRPC server")
	}
}

func TestGRPCHealthCheck(t *testing.T) {
	config := grpc.DefaultConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := grpc.NewServer(config, logger)

	ctx := context.Background()
	response, err := server.HealthCheck(ctx)

	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
	}
}

func TestGRPCSystemInfo(t *testing.T) {
	config := grpc.DefaultConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := grpc.NewServer(config, logger)

	ctx := context.Background()
	response, err := server.GetSystemInfo(ctx)

	if err != nil {
		t.Fatalf("Get system info failed: %v", err)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
	}

	if response.PeerCount <= 0 {
		t.Errorf("Expected positive peer count, got %d", response.PeerCount)
	}

	if response.FileCount <= 0 {
		t.Errorf("Expected positive file count, got %d", response.FileCount)
	}
}

func TestGRPCMetrics(t *testing.T) {
	config := grpc.DefaultConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := grpc.NewServer(config, logger)

	ctx := context.Background()
	response, err := server.GetMetrics(ctx)

	if err != nil {
		t.Fatalf("Get metrics failed: %v", err)
	}

	if response.RequestsTotal <= 0 {
		t.Errorf("Expected positive requests total, got %d", response.RequestsTotal)
	}

	if response.RequestsPerMinute <= 0 {
		t.Errorf("Expected positive requests per minute, got %f", response.RequestsPerMinute)
	}

	if response.ActiveConnections <= 0 {
		t.Errorf("Expected positive active connections, got %d", response.ActiveConnections)
	}
}

func TestGRPCConfig(t *testing.T) {
	config := grpc.DefaultConfig()

	if config.Port != ":8082" {
		t.Errorf("Expected port ':8082', got '%s'", config.Port)
	}

	if config.AuthToken != "demo-token" {
		t.Errorf("Expected auth token 'demo-token', got '%s'", config.AuthToken)
	}

	if config.MaxConcurrentStreams != 100 {
		t.Errorf("Expected max concurrent streams 100, got %d", config.MaxConcurrentStreams)
	}
}
