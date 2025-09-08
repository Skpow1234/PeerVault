package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// CacheWarmer manages cache warming operations
type CacheWarmer struct {
	queryCache *QueryCache
	queries    []CacheWarmQuery
	mu         sync.RWMutex
	logger     *slog.Logger
	running    bool
	stopCh     chan struct{}
}

// WarmupStrategy defines how cache warming should be performed
type WarmupStrategy string

const (
	// WarmupStrategyImmediate - Warm cache immediately on startup
	WarmupStrategyImmediate WarmupStrategy = "immediate"
	// WarmupStrategyScheduled - Warm cache on a schedule
	WarmupStrategyScheduled WarmupStrategy = "scheduled"
	// WarmupStrategyOnDemand - Warm cache on demand
	WarmupStrategyOnDemand WarmupStrategy = "on_demand"
)

// WarmupConfig holds configuration for cache warming
type WarmupConfig struct {
	Strategy    WarmupStrategy `json:"strategy"`
	Interval    time.Duration  `json:"interval"`
	Concurrency int            `json:"concurrency"`
	MaxRetries  int            `json:"maxRetries"`
	RetryDelay  time.Duration  `json:"retryDelay"`
	Enabled     bool           `json:"enabled"`
}

// DefaultWarmupConfig returns the default warmup configuration
func DefaultWarmupConfig() *WarmupConfig {
	return &WarmupConfig{
		Strategy:    WarmupStrategyScheduled,
		Interval:    5 * time.Minute,
		Concurrency: 3,
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		Enabled:     true,
	}
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(queryCache *QueryCache, logger *slog.Logger) *CacheWarmer {
	return &CacheWarmer{
		queryCache: queryCache,
		queries:    make([]CacheWarmQuery, 0),
		logger:     logger,
		stopCh:     make(chan struct{}),
	}
}

// AddWarmupQuery adds a query to the warmup list
func (cw *CacheWarmer) AddWarmupQuery(query CacheWarmQuery) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// Set default values
	if query.TTL <= 0 {
		query.TTL = 5 * time.Minute
	}
	if query.Priority <= 0 {
		query.Priority = 1
	}

	cw.queries = append(cw.queries, query)
	cw.logger.Info("Added warmup query", "query", query.Query, "priority", query.Priority)
}

// RemoveWarmupQuery removes a query from the warmup list
func (cw *CacheWarmer) RemoveWarmupQuery(query string) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	for i, q := range cw.queries {
		if q.Query == query {
			cw.queries = append(cw.queries[:i], cw.queries[i+1:]...)
			cw.logger.Info("Removed warmup query", "query", query)
			break
		}
	}
}

// Start starts the cache warming process
func (cw *CacheWarmer) Start(ctx context.Context, config *WarmupConfig) error {
	if config == nil {
		config = DefaultWarmupConfig()
	}

	if !config.Enabled {
		cw.logger.Info("Cache warming is disabled")
		return nil
	}

	cw.mu.Lock()
	if cw.running {
		cw.mu.Unlock()
		return fmt.Errorf("cache warmer is already running")
	}
	cw.running = true
	cw.mu.Unlock()

	cw.logger.Info("Starting cache warmer",
		"strategy", config.Strategy,
		"interval", config.Interval,
		"concurrency", config.Concurrency)

	switch config.Strategy {
	case WarmupStrategyImmediate:
		return cw.warmupImmediate(ctx, config)
	case WarmupStrategyScheduled:
		return cw.warmupScheduled(ctx, config)
	case WarmupStrategyOnDemand:
		// On-demand warming doesn't need a background process
		return nil
	default:
		return fmt.Errorf("unknown warmup strategy: %s", config.Strategy)
	}
}

// Stop stops the cache warming process
func (cw *CacheWarmer) Stop() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if !cw.running {
		return
	}

	close(cw.stopCh)
	cw.running = false
	cw.logger.Info("Stopped cache warmer")
}

// WarmupNow performs immediate cache warming
func (cw *CacheWarmer) WarmupNow(ctx context.Context, config *WarmupConfig) error {
	if config == nil {
		config = DefaultWarmupConfig()
	}

	cw.mu.RLock()
	queries := make([]CacheWarmQuery, len(cw.queries))
	copy(queries, cw.queries)
	cw.mu.RUnlock()

	return cw.executeWarmup(ctx, queries, config)
}

// warmupImmediate performs immediate cache warming
func (cw *CacheWarmer) warmupImmediate(ctx context.Context, config *WarmupConfig) error {
	cw.mu.RLock()
	queries := make([]CacheWarmQuery, len(cw.queries))
	copy(queries, cw.queries)
	cw.mu.RUnlock()

	return cw.executeWarmup(ctx, queries, config)
}

// warmupScheduled performs scheduled cache warming
func (cw *CacheWarmer) warmupScheduled(ctx context.Context, config *WarmupConfig) error {
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Perform initial warmup
	if err := cw.warmupImmediate(ctx, config); err != nil {
		cw.logger.Error("Initial warmup failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cw.stopCh:
			return nil
		case <-ticker.C:
			cw.mu.RLock()
			queries := make([]CacheWarmQuery, len(cw.queries))
			copy(queries, cw.queries)
			cw.mu.RUnlock()

			if err := cw.executeWarmup(ctx, queries, config); err != nil {
				cw.logger.Error("Scheduled warmup failed", "error", err)
			}
		}
	}
}

// executeWarmup executes the warmup process
func (cw *CacheWarmer) executeWarmup(ctx context.Context, queries []CacheWarmQuery, config *WarmupConfig) error {
	if len(queries) == 0 {
		cw.logger.Debug("No warmup queries to execute")
		return nil
	}

	// Sort queries by priority (higher priority first)
	sortedQueries := cw.sortQueriesByPriority(queries)

	cw.logger.Info("Starting cache warmup", "queries", len(sortedQueries))

	// Execute warmup queries with concurrency control
	semaphore := make(chan struct{}, config.Concurrency)
	var wg sync.WaitGroup
	var errors []error
	var mu sync.Mutex

	for _, query := range sortedQueries {
		wg.Add(1)
		go func(q CacheWarmQuery) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-semaphore }()

			// Execute warmup query with retries
			if err := cw.executeWarmupQueryWithRetry(ctx, q, config); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("warmup query failed: %w", err))
				mu.Unlock()
			}
		}(query)
	}

	wg.Wait()

	if len(errors) > 0 {
		cw.logger.Error("Cache warmup completed with errors", "errors", len(errors))
		for _, err := range errors {
			cw.logger.Error("Warmup error", "error", err)
		}
		return fmt.Errorf("warmup completed with %d errors", len(errors))
	}

	cw.logger.Info("Cache warmup completed successfully")
	return nil
}

// executeWarmupQueryWithRetry executes a warmup query with retry logic
func (cw *CacheWarmer) executeWarmupQueryWithRetry(ctx context.Context, query CacheWarmQuery, config *WarmupConfig) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		if err := cw.executeWarmupQuery(ctx, query); err != nil {
			lastErr = err
			cw.logger.Warn("Warmup query failed",
				"query", query.Query,
				"attempt", attempt,
				"maxRetries", config.MaxRetries,
				"error", err)

			if attempt < config.MaxRetries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(config.RetryDelay):
					continue
				}
			}
		} else {
			cw.logger.Debug("Warmup query succeeded", "query", query.Query, "attempt", attempt)
			return nil
		}
	}

	return fmt.Errorf("warmup query failed after %d attempts: %w", config.MaxRetries, lastErr)
}

// executeWarmupQuery executes a single warmup query
func (cw *CacheWarmer) executeWarmupQuery(ctx context.Context, query CacheWarmQuery) error {
	// In a real implementation, this would:
	// 1. Parse the GraphQL query
	// 2. Execute the query against the GraphQL engine
	// 3. Cache the result

	// For now, we'll simulate the execution
	cw.logger.Debug("Executing warmup query",
		"query", query.Query,
		"variables", query.Variables,
		"ttl", query.TTL)

	// Simulate query execution time
	time.Sleep(100 * time.Millisecond)

	// Simulate successful execution
	cw.logger.Debug("Warmup query executed successfully", "query", query.Query)
	return nil
}

// sortQueriesByPriority sorts queries by priority (higher priority first)
func (cw *CacheWarmer) sortQueriesByPriority(queries []CacheWarmQuery) []CacheWarmQuery {
	// Simple bubble sort by priority
	sorted := make([]CacheWarmQuery, len(queries))
	copy(sorted, queries)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Priority < sorted[j+1].Priority {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// GetWarmupQueries returns the current warmup queries
func (cw *CacheWarmer) GetWarmupQueries() []CacheWarmQuery {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	queries := make([]CacheWarmQuery, len(cw.queries))
	copy(queries, cw.queries)
	return queries
}

// IsRunning returns whether the cache warmer is currently running
func (cw *CacheWarmer) IsRunning() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.running
}

// CreateDefaultWarmupQueries creates default warmup queries
func (cw *CacheWarmer) CreateDefaultWarmupQueries() {
	defaultQueries := []CacheWarmQuery{
		{
			Query: `
				query {
					systemMetrics {
						storage {
							totalSpace
							usedSpace
							availableSpace
						}
						performance {
							averageResponseTime
							requestsPerSecond
						}
					}
				}
			`,
			TTL:      2 * time.Minute,
			Priority: 10,
		},
		{
			Query: `
				query {
					nodes {
						id
						address
						port
						status
						health {
							isHealthy
							responseTime
						}
					}
				}
			`,
			TTL:      1 * time.Minute,
			Priority: 8,
		},
		{
			Query: `
				query {
					files(limit: 100) {
						id
						key
						size
						createdAt
						owner {
							id
							address
						}
					}
				}
			`,
			TTL:      3 * time.Minute,
			Priority: 5,
		},
		{
			Query: `
				query {
					peerNetwork {
						nodes {
							id
							address
							status
						}
						connections {
							from {
								id
							}
							to {
								id
							}
							status
							latency
						}
					}
				}
			`,
			TTL:      1 * time.Minute,
			Priority: 7,
		},
	}

	for _, query := range defaultQueries {
		cw.AddWarmupQuery(query)
	}

	cw.logger.Info("Created default warmup queries", "count", len(defaultQueries))
}

// GetWarmupStats returns statistics about cache warming
func (cw *CacheWarmer) GetWarmupStats() map[string]interface{} {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	queries := cw.queries
	totalQueries := len(queries)

	// Calculate priority distribution
	priorityCounts := make(map[int]int)
	for _, query := range queries {
		priorityCounts[query.Priority]++
	}

	// Calculate TTL distribution
	ttlRanges := map[string]int{
		"< 1min":  0,
		"1-5min":  0,
		"5-15min": 0,
		"> 15min": 0,
	}

	for _, query := range queries {
		switch {
		case query.TTL < time.Minute:
			ttlRanges["< 1min"]++
		case query.TTL <= 5*time.Minute:
			ttlRanges["1-5min"]++
		case query.TTL <= 15*time.Minute:
			ttlRanges["5-15min"]++
		default:
			ttlRanges["> 15min"]++
		}
	}

	return map[string]interface{}{
		"totalQueries":   totalQueries,
		"isRunning":      cw.running,
		"priorityCounts": priorityCounts,
		"ttlRanges":      ttlRanges,
		"queries":        queries,
	}
}
