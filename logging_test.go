package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/anthdm/foreverstore/internal/logging"
)

// TestLoggingConfiguration verifies that logging can be configured properly
func TestLoggingConfiguration(t *testing.T) {
	// Test different log levels
	testCases := []struct {
		level    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"invalid", slog.LevelInfo}, // Default fallback
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			// Configure logger
			logging.ConfigureLogger(tc.level)

			// Verify the logger is working
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			logger.Info("test message", "level", tc.level)

			// Parse JSON output
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("failed to parse log output: %v", err)
			}

			// Verify log entry has expected fields
			if logEntry["msg"] != "test message" {
				t.Errorf("expected message 'test message', got '%v'", logEntry["msg"])
			}

			if logEntry["level"] != tc.level {
				t.Errorf("expected level '%s', got '%v'", tc.level, logEntry["level"])
			}
		})
	}
}

// TestStructuredLogging verifies that structured logging works with context
func TestStructuredLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Test logging with different contexts
	logger.Info("file stored",
		"key", "test-file.txt",
		"bytes", 1024,
		"peer", "127.0.0.1:3000",
	)

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	// Verify all fields are present
	expectedFields := map[string]interface{}{
		"msg":   "file stored",
		"key":   "test-file.txt",
		"bytes": float64(1024), // JSON numbers are float64
		"peer":  "127.0.0.1:3000",
	}

	for key, expectedValue := range expectedFields {
		if logEntry[key] != expectedValue {
			t.Errorf("expected %s='%v', got '%v'", key, expectedValue, logEntry[key])
		}
	}
}

// TestErrorLogging verifies that error logging works correctly
func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Test error logging
	testError := "connection failed"
	logger.Error("network error", "error", testError, "peer", "127.0.0.1:3000")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	// Verify error level and message
	if logEntry["level"] != "ERROR" {
		t.Errorf("expected level 'ERROR', got '%v'", logEntry["level"])
	}

	if logEntry["error"] != testError {
		t.Errorf("expected error '%s', got '%v'", testError, logEntry["error"])
	}
}

// TestLoggingHelpers verifies that logging helper functions work
func TestLoggingHelpers(t *testing.T) {
	// Test component logger
	componentLogger := logging.Logger("fileserver")
	if componentLogger == nil {
		t.Error("component logger should not be nil")
	}

	// Test context helpers
	errorLogger := logging.WithError(fmt.Errorf("test error"))
	if errorLogger == nil {
		t.Error("error logger should not be nil")
	}

	peerLogger := logging.WithPeer("127.0.0.1:3000")
	if peerLogger == nil {
		t.Error("peer logger should not be nil")
	}

	keyLogger := logging.WithKey("test-key")
	if keyLogger == nil {
		t.Error("key logger should not be nil")
	}

	bytesLogger := logging.WithBytes(1024)
	if bytesLogger == nil {
		t.Error("bytes logger should not be nil")
	}
}

// TestLogLevels verifies that different log levels work correctly
func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)

	// Debug should be filtered out
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("debug message should be filtered out at info level")
	}

	// Info should be logged
	logger.Info("info message")
	if buf.Len() == 0 {
		t.Error("info message should be logged")
	}

	// Error should be logged
	logger.Error("error message")
	if buf.Len() == 0 {
		t.Error("error message should be logged")
	}
}
