package interceptors

import (
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

// InterceptorManager manages all gRPC interceptors
type InterceptorManager struct {
	authInterceptor           *AuthInterceptor
	loggingInterceptor        *LoggingInterceptor
	monitoringInterceptor     *MonitoringInterceptor
	rateLimitInterceptor      *RateLimitInterceptor
	validationInterceptor     *ValidationInterceptor
	cacheInterceptor          *CacheInterceptor
	circuitBreakerInterceptor *CircuitBreakerInterceptor
	logger                    *slog.Logger
}

// NewInterceptorManager creates a new interceptor manager
func NewInterceptorManager(logger *slog.Logger) *InterceptorManager {
	if logger == nil {
		logger = slog.Default()
	}

	return &InterceptorManager{
		logger: logger,
	}
}

// SetAuthInterceptor sets the authentication interceptor
func (im *InterceptorManager) SetAuthInterceptor(config *AuthConfig) {
	im.authInterceptor = NewAuthInterceptor(config, im.logger)
}

// SetLoggingInterceptor sets the logging interceptor
func (im *InterceptorManager) SetLoggingInterceptor(config *LoggingConfig) {
	im.loggingInterceptor = NewLoggingInterceptor(config, im.logger)
}

// SetMonitoringInterceptor sets the monitoring interceptor
func (im *InterceptorManager) SetMonitoringInterceptor(config *MonitoringConfig) {
	im.monitoringInterceptor = NewMonitoringInterceptor(config, im.logger)
}

// SetRateLimitInterceptor sets the rate limit interceptor
func (im *InterceptorManager) SetRateLimitInterceptor(requestsPerSecond, burstSize int) {
	im.rateLimitInterceptor = NewRateLimitInterceptor(requestsPerSecond, burstSize, im.logger)
}

// SetValidationInterceptor sets the validation interceptor
func (im *InterceptorManager) SetValidationInterceptor() {
	im.validationInterceptor = NewValidationInterceptor(im.logger)
}

// SetCacheInterceptor sets the cache interceptor
func (im *InterceptorManager) SetCacheInterceptor(cacheTTL time.Duration) {
	im.cacheInterceptor = NewCacheInterceptor(cacheTTL, im.logger)
}

// SetCircuitBreakerInterceptor sets the circuit breaker interceptor
func (im *InterceptorManager) SetCircuitBreakerInterceptor(failureThreshold int, timeout time.Duration) {
	im.circuitBreakerInterceptor = NewCircuitBreakerInterceptor(failureThreshold, timeout, im.logger)
}

// GetUnaryInterceptors returns all unary server interceptors
func (im *InterceptorManager) GetUnaryInterceptors() []grpc.UnaryServerInterceptor {
	var interceptors []grpc.UnaryServerInterceptor

	// Add interceptors in order of execution
	if im.circuitBreakerInterceptor != nil {
		interceptors = append(interceptors, im.circuitBreakerInterceptor.UnaryCircuitBreakerInterceptor())
	}

	if im.rateLimitInterceptor != nil {
		interceptors = append(interceptors, im.rateLimitInterceptor.UnaryRateLimitInterceptor())
	}

	if im.authInterceptor != nil {
		interceptors = append(interceptors, im.authInterceptor.UnaryAuthInterceptor())
	}

	if im.validationInterceptor != nil {
		interceptors = append(interceptors, im.validationInterceptor.UnaryValidationInterceptor())
	}

	if im.cacheInterceptor != nil {
		interceptors = append(interceptors, im.cacheInterceptor.UnaryCacheInterceptor())
	}

	if im.loggingInterceptor != nil {
		interceptors = append(interceptors, im.loggingInterceptor.UnaryLoggingInterceptor())
	}

	if im.monitoringInterceptor != nil {
		interceptors = append(interceptors, im.monitoringInterceptor.UnaryMonitoringInterceptor())
	}

	return interceptors
}

// GetStreamInterceptors returns all stream server interceptors
func (im *InterceptorManager) GetStreamInterceptors() []grpc.StreamServerInterceptor {
	var interceptors []grpc.StreamServerInterceptor

	// Add interceptors in order of execution
	if im.circuitBreakerInterceptor != nil {
		interceptors = append(interceptors, im.circuitBreakerInterceptor.StreamCircuitBreakerInterceptor())
	}

	if im.rateLimitInterceptor != nil {
		interceptors = append(interceptors, im.rateLimitInterceptor.StreamRateLimitInterceptor())
	}

	if im.authInterceptor != nil {
		interceptors = append(interceptors, im.authInterceptor.StreamAuthInterceptor())
	}

	if im.validationInterceptor != nil {
		interceptors = append(interceptors, im.validationInterceptor.StreamValidationInterceptor())
	}

	if im.loggingInterceptor != nil {
		interceptors = append(interceptors, im.loggingInterceptor.StreamLoggingInterceptor())
	}

	if im.monitoringInterceptor != nil {
		interceptors = append(interceptors, im.monitoringInterceptor.StreamMonitoringInterceptor())
	}

	return interceptors
}

// GetAuthInterceptor returns the authentication interceptor
func (im *InterceptorManager) GetAuthInterceptor() *AuthInterceptor {
	return im.authInterceptor
}

// GetLoggingInterceptor returns the logging interceptor
func (im *InterceptorManager) GetLoggingInterceptor() *LoggingInterceptor {
	return im.loggingInterceptor
}

// GetMonitoringInterceptor returns the monitoring interceptor
func (im *InterceptorManager) GetMonitoringInterceptor() *MonitoringInterceptor {
	return im.monitoringInterceptor
}

// GetRateLimitInterceptor returns the rate limit interceptor
func (im *InterceptorManager) GetRateLimitInterceptor() *RateLimitInterceptor {
	return im.rateLimitInterceptor
}

// GetValidationInterceptor returns the validation interceptor
func (im *InterceptorManager) GetValidationInterceptor() *ValidationInterceptor {
	return im.validationInterceptor
}

// GetCacheInterceptor returns the cache interceptor
func (im *InterceptorManager) GetCacheInterceptor() *CacheInterceptor {
	return im.cacheInterceptor
}

// GetCircuitBreakerInterceptor returns the circuit breaker interceptor
func (im *InterceptorManager) GetCircuitBreakerInterceptor() *CircuitBreakerInterceptor {
	return im.circuitBreakerInterceptor
}

// EnableInterceptor enables a specific interceptor
func (im *InterceptorManager) EnableInterceptor(interceptorType string) {
	switch interceptorType {
	case "auth":
		if im.authInterceptor != nil {
			im.authInterceptor.RequireAuth = true
		}
	case "logging":
		if im.loggingInterceptor != nil {
			im.loggingInterceptor.config.LogRequests = true
		}
	case "monitoring":
		if im.monitoringInterceptor != nil {
			im.monitoringInterceptor.config.EnableMetrics = true
		}
	case "rate_limit":
		if im.rateLimitInterceptor != nil {
			im.rateLimitInterceptor.Enable()
		}
	case "validation":
		if im.validationInterceptor != nil {
			im.validationInterceptor.Enable()
		}
	case "cache":
		if im.cacheInterceptor != nil {
			im.cacheInterceptor.Enable()
		}
	case "circuit_breaker":
		if im.circuitBreakerInterceptor != nil {
			im.circuitBreakerInterceptor.Enable()
		}
	}
}

// DisableInterceptor disables a specific interceptor
func (im *InterceptorManager) DisableInterceptor(interceptorType string) {
	switch interceptorType {
	case "auth":
		if im.authInterceptor != nil {
			im.authInterceptor.RequireAuth = false
		}
	case "logging":
		if im.loggingInterceptor != nil {
			im.loggingInterceptor.config.LogRequests = false
		}
	case "monitoring":
		if im.monitoringInterceptor != nil {
			im.monitoringInterceptor.config.EnableMetrics = false
		}
	case "rate_limit":
		if im.rateLimitInterceptor != nil {
			im.rateLimitInterceptor.Disable()
		}
	case "validation":
		if im.validationInterceptor != nil {
			im.validationInterceptor.Disable()
		}
	case "cache":
		if im.cacheInterceptor != nil {
			im.cacheInterceptor.Disable()
		}
	case "circuit_breaker":
		if im.circuitBreakerInterceptor != nil {
			im.circuitBreakerInterceptor.Disable()
		}
	}
}

// GetInterceptorStatus returns the status of all interceptors
func (im *InterceptorManager) GetInterceptorStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// Auth interceptor status
	if im.authInterceptor != nil {
		status["auth"] = map[string]interface{}{
			"enabled":  im.authInterceptor.RequireAuth,
			"issuer":   im.authInterceptor.config.Issuer,
			"audience": im.authInterceptor.config.Audience,
		}
	}

	// Logging interceptor status
	if im.loggingInterceptor != nil {
		status["logging"] = map[string]interface{}{
			"log_requests":  im.loggingInterceptor.config.LogRequests,
			"log_responses": im.loggingInterceptor.config.LogResponses,
			"log_errors":    im.loggingInterceptor.config.LogErrors,
			"log_duration":  im.loggingInterceptor.config.LogDuration,
		}
	}

	// Monitoring interceptor status
	if im.monitoringInterceptor != nil {
		status["monitoring"] = map[string]interface{}{
			"enable_metrics": im.monitoringInterceptor.config.EnableMetrics,
			"enable_tracing": im.monitoringInterceptor.config.EnableTracing,
			"sample_rate":    im.monitoringInterceptor.config.SampleRate,
		}
	}

	// Rate limit interceptor status
	if im.rateLimitInterceptor != nil {
		status["rate_limit"] = map[string]interface{}{
			"enabled":             im.rateLimitInterceptor.IsEnabled(),
			"requests_per_second": im.rateLimitInterceptor.requestsPerSecond,
			"burst_size":          im.rateLimitInterceptor.burstSize,
		}
	}

	// Validation interceptor status
	if im.validationInterceptor != nil {
		status["validation"] = map[string]interface{}{
			"enabled":    im.validationInterceptor.IsEnabled(),
			"validators": len(im.validationInterceptor.validators),
		}
	}

	// Cache interceptor status
	if im.cacheInterceptor != nil {
		status["cache"] = map[string]interface{}{
			"enabled":    im.cacheInterceptor.IsEnabled(),
			"cache_ttl":  im.cacheInterceptor.cacheTTL,
			"cache_size": len(im.cacheInterceptor.cache),
		}
	}

	// Circuit breaker interceptor status
	if im.circuitBreakerInterceptor != nil {
		status["circuit_breaker"] = map[string]interface{}{
			"enabled":           im.circuitBreakerInterceptor.IsEnabled(),
			"state":             im.circuitBreakerInterceptor.GetState(),
			"failure_count":     im.circuitBreakerInterceptor.GetFailureCount(),
			"failure_threshold": im.circuitBreakerInterceptor.failureThreshold,
		}
	}

	return status
}

// ResetAllMetrics resets all metrics in monitoring interceptor
func (im *InterceptorManager) ResetAllMetrics() {
	if im.monitoringInterceptor != nil {
		im.monitoringInterceptor.ResetMetrics()
	}
}

// ResetAllTraces resets all traces in monitoring interceptor
func (im *InterceptorManager) ResetAllTraces() {
	if im.monitoringInterceptor != nil {
		im.monitoringInterceptor.ResetTraces()
	}
}

// GetMetrics returns metrics from monitoring interceptor
func (im *InterceptorManager) GetMetrics() map[string]interface{} {
	if im.monitoringInterceptor != nil {
		return im.monitoringInterceptor.GetMetrics()
	}
	return make(map[string]interface{})
}

// GetTraces returns traces from monitoring interceptor
func (im *InterceptorManager) GetTraces() map[string]*Trace {
	if im.monitoringInterceptor != nil {
		return im.monitoringInterceptor.GetTraces()
	}
	return make(map[string]*Trace)
}
