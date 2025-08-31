package logging

import (
	"log/slog"
	"os"
)

// ConfigureLogger sets up structured logging with appropriate level and format
func ConfigureLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create a structured logger with JSON output for production
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	// Use JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// Logger returns a logger with additional context
func Logger(component string) *slog.Logger {
	return slog.With("component", component)
}

// WithError creates a logger with error context
func WithError(err error) *slog.Logger {
	return slog.With("error", err.Error())
}

// WithPeer creates a logger with peer context
func WithPeer(peer string) *slog.Logger {
	return slog.With("peer", peer)
}

// WithKey creates a logger with file key context
func WithKey(key string) *slog.Logger {
	return slog.With("key", key)
}

// WithBytes creates a logger with byte count context
func WithBytes(bytes int64) *slog.Logger {
	return slog.With("bytes", bytes)
}