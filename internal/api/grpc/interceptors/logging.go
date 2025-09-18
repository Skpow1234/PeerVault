package interceptors

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	LogRequests    bool
	LogResponses   bool
	LogErrors      bool
	LogDuration    bool
	LogUserID      bool
	LogMetadata    bool
	LogPayload     bool
	MaxPayloadSize int
	SkipMethods    []string
	LogLevel       slog.Level
}

// DefaultLoggingConfig returns the default logging configuration
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		LogRequests:    true,
		LogResponses:   true,
		LogErrors:      true,
		LogDuration:    true,
		LogUserID:      true,
		LogMetadata:    false,
		LogPayload:     false,
		MaxPayloadSize: 1024,
		SkipMethods:    []string{"/peervault.PeerVaultService/HealthCheck"},
		LogLevel:       slog.LevelInfo,
	}
}

// LoggingInterceptor provides logging for gRPC services
type LoggingInterceptor struct {
	config *LoggingConfig
	logger *slog.Logger
}

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(config *LoggingConfig, logger *slog.Logger) *LoggingInterceptor {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &LoggingInterceptor{
		config: config,
		logger: logger,
	}
}

// UnaryLoggingInterceptor returns a unary server interceptor for logging
func (li *LoggingInterceptor) UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if method should skip logging
		if li.shouldSkipLogging(info.FullMethod) {
			return handler(ctx, req)
		}

		start := time.Now()
		requestID := li.generateRequestID()

		// Log request
		if li.config.LogRequests {
			li.logRequest(ctx, requestID, info.FullMethod, req)
		}

		// Call handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Log response
		if li.config.LogResponses {
			li.logResponse(ctx, requestID, info.FullMethod, resp, err, duration)
		}

		// Log error if present
		if err != nil && li.config.LogErrors {
			li.logError(ctx, requestID, info.FullMethod, err, duration)
		}

		return resp, err
	}
}

// StreamLoggingInterceptor returns a stream server interceptor for logging
func (li *LoggingInterceptor) StreamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Check if method should skip logging
		if li.shouldSkipLogging(info.FullMethod) {
			return handler(srv, ss)
		}

		start := time.Now()
		requestID := li.generateRequestID()

		// Log stream start
		if li.config.LogRequests {
			li.logStreamStart(ss.Context(), requestID, info.FullMethod)
		}

		// Wrap stream for logging
		wrappedStream := &loggingServerStream{
			ServerStream: ss,
			interceptor:  li,
			requestID:    requestID,
			method:       info.FullMethod,
			startTime:    start,
		}

		// Call handler
		err := handler(srv, wrappedStream)

		duration := time.Since(start)

		// Log stream end
		if li.config.LogResponses {
			li.logStreamEnd(ss.Context(), requestID, info.FullMethod, err, duration)
		}

		// Log error if present
		if err != nil && li.config.LogErrors {
			li.logError(ss.Context(), requestID, info.FullMethod, err, duration)
		}

		return err
	}
}

// shouldSkipLogging checks if the method should skip logging
func (li *LoggingInterceptor) shouldSkipLogging(method string) bool {
	for _, skipMethod := range li.config.SkipMethods {
		if method == skipMethod {
			return true
		}
	}
	return false
}

// logRequest logs a request
func (li *LoggingInterceptor) logRequest(ctx context.Context, requestID, method string, req interface{}) {
	attrs := []slog.Attr{
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("type", "request"),
	}

	// Add user ID if available
	if li.config.LogUserID {
		if userID, ok := ctx.Value("user_id").(string); ok {
			attrs = append(attrs, slog.String("user_id", userID))
		}
	}

	// Add metadata if enabled
	if li.config.LogMetadata {
		if md, ok := ctx.Value("metadata").(map[string]string); ok {
			attrs = append(attrs, slog.Any("metadata", md))
		}
	}

	// Add payload if enabled
	if li.config.LogPayload {
		payload := li.truncatePayload(fmt.Sprintf("%+v", req))
		attrs = append(attrs, slog.String("payload", payload))
	}

	li.logger.LogAttrs(ctx, li.config.LogLevel, "gRPC request", attrs...)
}

// logResponse logs a response
func (li *LoggingInterceptor) logResponse(ctx context.Context, requestID, method string, resp interface{}, err error, duration time.Duration) {
	attrs := []slog.Attr{
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("type", "response"),
	}

	// Add duration if enabled
	if li.config.LogDuration {
		attrs = append(attrs, slog.Duration("duration", duration))
	}

	// Add user ID if available
	if li.config.LogUserID {
		if userID, ok := ctx.Value("user_id").(string); ok {
			attrs = append(attrs, slog.String("user_id", userID))
		}
	}

	// Add status code
	if err != nil {
		if st, ok := status.FromError(err); ok {
			attrs = append(attrs, slog.String("status_code", st.Code().String()))
			attrs = append(attrs, slog.String("status_message", st.Message()))
		} else {
			attrs = append(attrs, slog.String("status_code", codes.Unknown.String()))
			attrs = append(attrs, slog.String("status_message", err.Error()))
		}
	} else {
		attrs = append(attrs, slog.String("status_code", codes.OK.String()))
	}

	// Add payload if enabled
	if li.config.LogPayload && resp != nil {
		payload := li.truncatePayload(fmt.Sprintf("%+v", resp))
		attrs = append(attrs, slog.String("payload", payload))
	}

	li.logger.LogAttrs(ctx, li.config.LogLevel, "gRPC response", attrs...)
}

// logError logs an error
func (li *LoggingInterceptor) logError(ctx context.Context, requestID, method string, err error, duration time.Duration) {
	attrs := []slog.Attr{
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("type", "error"),
		slog.String("error", err.Error()),
	}

	// Add duration if enabled
	if li.config.LogDuration {
		attrs = append(attrs, slog.Duration("duration", duration))
	}

	// Add user ID if available
	if li.config.LogUserID {
		if userID, ok := ctx.Value("user_id").(string); ok {
			attrs = append(attrs, slog.String("user_id", userID))
		}
	}

	// Add status code
	if st, ok := status.FromError(err); ok {
		attrs = append(attrs, slog.String("status_code", st.Code().String()))
		attrs = append(attrs, slog.String("status_message", st.Message()))
	} else {
		attrs = append(attrs, slog.String("status_code", codes.Unknown.String()))
	}

	li.logger.LogAttrs(ctx, slog.LevelError, "gRPC error", attrs...)
}

// logStreamStart logs the start of a stream
func (li *LoggingInterceptor) logStreamStart(ctx context.Context, requestID, method string) {
	attrs := []slog.Attr{
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("type", "stream_start"),
	}

	// Add user ID if available
	if li.config.LogUserID {
		if userID, ok := ctx.Value("user_id").(string); ok {
			attrs = append(attrs, slog.String("user_id", userID))
		}
	}

	li.logger.LogAttrs(ctx, li.config.LogLevel, "gRPC stream started", attrs...)
}

// logStreamEnd logs the end of a stream
func (li *LoggingInterceptor) logStreamEnd(ctx context.Context, requestID, method string, err error, duration time.Duration) {
	attrs := []slog.Attr{
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("type", "stream_end"),
	}

	// Add duration if enabled
	if li.config.LogDuration {
		attrs = append(attrs, slog.Duration("duration", duration))
	}

	// Add user ID if available
	if li.config.LogUserID {
		if userID, ok := ctx.Value("user_id").(string); ok {
			attrs = append(attrs, slog.String("user_id", userID))
		}
	}

	// Add status code
	if err != nil {
		if st, ok := status.FromError(err); ok {
			attrs = append(attrs, slog.String("status_code", st.Code().String()))
			attrs = append(attrs, slog.String("status_message", st.Message()))
		} else {
			attrs = append(attrs, slog.String("status_code", codes.Unknown.String()))
			attrs = append(attrs, slog.String("status_message", err.Error()))
		}
	} else {
		attrs = append(attrs, slog.String("status_code", codes.OK.String()))
	}

	li.logger.LogAttrs(ctx, li.config.LogLevel, "gRPC stream ended", attrs...)
}

// truncatePayload truncates payload to maximum size
func (li *LoggingInterceptor) truncatePayload(payload string) string {
	if len(payload) <= li.config.MaxPayloadSize {
		return payload
	}
	return payload[:li.config.MaxPayloadSize] + "..."
}

// generateRequestID generates a unique request ID
func (li *LoggingInterceptor) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// loggingServerStream wraps a ServerStream for logging
type loggingServerStream struct {
	grpc.ServerStream
	interceptor  *LoggingInterceptor
	requestID    string
	method       string
	startTime    time.Time
	messageCount int
}

// SendMsg logs the message being sent
func (lss *loggingServerStream) SendMsg(m interface{}) error {
	lss.messageCount++

	if lss.interceptor.config.LogPayload {
		payload := lss.interceptor.truncatePayload(fmt.Sprintf("%+v", m))
		lss.interceptor.logger.LogAttrs(lss.Context(), lss.interceptor.config.LogLevel, "gRPC stream send",
			slog.String("request_id", lss.requestID),
			slog.String("method", lss.method),
			slog.String("type", "stream_send"),
			slog.Int("message_count", lss.messageCount),
			slog.String("payload", payload),
		)
	}

	return lss.ServerStream.SendMsg(m)
}

// RecvMsg logs the message being received
func (lss *loggingServerStream) RecvMsg(m interface{}) error {
	err := lss.ServerStream.RecvMsg(m)

	if err == nil {
		lss.messageCount++

		if lss.interceptor.config.LogPayload {
			payload := lss.interceptor.truncatePayload(fmt.Sprintf("%+v", m))
			lss.interceptor.logger.LogAttrs(lss.Context(), lss.interceptor.config.LogLevel, "gRPC stream receive",
				slog.String("request_id", lss.requestID),
				slog.String("method", lss.method),
				slog.String("type", "stream_receive"),
				slog.Int("message_count", lss.messageCount),
				slog.String("payload", payload),
			)
		}
	}

	return err
}
