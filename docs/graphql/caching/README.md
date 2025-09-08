# GraphQL Advanced Caching

This document describes the advanced caching system implemented in PeerVault's GraphQL layer, providing intelligent query result caching with sophisticated invalidation strategies.

## Overview

The GraphQL caching system provides:

- **Query Result Caching**: Cache GraphQL query results with configurable TTL
- **Intelligent Invalidation**: Multiple invalidation strategies (TTL, event-based, hybrid)
- **Cache Warming**: Preload cache with common queries
- **Analytics & Insights**: Comprehensive cache performance monitoring
- **Pattern-based Invalidation**: Invalidate cache entries by patterns or tags

## Architecture

```bash
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GraphQL       │    │   Query Cache   │    │  Invalidation   │
│   Resolvers     │───▶│                 │◀───│   Manager       │
│                 │    │  - TTL Cache    │    │                 │
└─────────────────┘    │  - Compression  │    │  - Event-based  │
                       │  - Metrics      │    │  - Pattern      │
┌─────────────────┐    └─────────────────┘    │  - Hybrid       │
│   Cache         │                           └─────────────────┘
│   Analytics     │◀─────────────────────────────────────────────┘
│                 │
│  - Performance  │    ┌─────────────────┐
│  - Insights     │    │   Cache         │
│  - Trends       │    │   Warmer        │
└─────────────────┘    │                 │
                       │  - Scheduled    │
                       │  - On-demand    │
                       │  - Immediate    │
                       └─────────────────┘
```

## Cache Configuration

### Basic Configuration

```go
config := &cache.QueryCacheConfig{
    DefaultTTL:        5 * time.Minute,
    MaxCacheSize:      1000,
    EnableCompression: true,
    EnableMetrics:     true,
    InvalidationStrategy: cache.InvalidationStrategyHybrid,
}

queryCache, err := cache.NewQueryCache(config, logger)
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `DefaultTTL` | `time.Duration` | `5m` | Default time-to-live for cached entries |
| `MaxCacheSize` | `int` | `1000` | Maximum number of cached entries |
| `EnableCompression` | `bool` | `true` | Enable compression for cached data |
| `EnableMetrics` | `bool` | `true` | Enable cache performance metrics |
| `InvalidationStrategy` | `InvalidationStrategy` | `hybrid` | Cache invalidation strategy |

## Cache Operations

### Basic Cache Operations

```go
// Get cached result
result, found := queryCache.Get(ctx, query, variables)
if found {
    return result, nil
}

// Execute query and cache result
data, err := executeQuery(ctx, query, variables)
if err != nil {
    return nil, err
}

// Cache the result
err = queryCache.Set(ctx, query, variables, data, nil, 5*time.Minute)
return data, err
```

### Cache Invalidation

```go
// Invalidate by pattern
err := queryCache.InvalidateByPattern(ctx, "file.*")

// Invalidate by tags
err := queryCache.InvalidateByTags(ctx, []string{"user", "profile"})

// Clear all cache
err := queryCache.Clear(ctx)
```

## Invalidation Strategies

### TTL-based Invalidation

Cache entries are automatically invalidated after their TTL expires.

```go
config := &cache.QueryCacheConfig{
    InvalidationStrategy: cache.InvalidationStrategyTTL,
    DefaultTTL: 10 * time.Minute,
}
```

### Event-based Invalidation

Cache entries are invalidated based on events.

```go
// Create invalidation rule
rule := &cache.InvalidationRule{
    ID:       "file-update-rule",
    Name:     "File Update Invalidation",
    Pattern:  "file.*",
    Strategy: cache.InvalidationStrategyEvent,
    Conditions: []cache.InvalidationCondition{
        {
            Type:     cache.ConditionTypeEvent,
            Field:    "type",
            Operator: cache.OperatorEquals,
            Value:    "file.updated",
        },
    },
}

// Process invalidation event
event := &cache.InvalidationEvent{
    Type: "file.updated",
    Data: map[string]interface{}{
        "fileId": "file_123",
        "key":    "example.txt",
    },
    Tags: []string{"file", "storage"},
}

err := invalidationManager.ProcessEvent(ctx, event)
```

### Hybrid Invalidation

Combines TTL and event-based invalidation.

```go
config := &cache.QueryCacheConfig{
    InvalidationStrategy: cache.InvalidationStrategyHybrid,
    DefaultTTL: 5 * time.Minute,
}
```

## Cache Warming

### Warmup Configuration

```go
warmupConfig := &cache.WarmupConfig{
    Strategy:    cache.WarmupStrategyScheduled,
    Interval:    5 * time.Minute,
    Concurrency: 3,
    MaxRetries:  3,
    RetryDelay:  1 * time.Second,
    Enabled:     true,
}
```

### Warmup Strategies

#### Immediate Warming

```go
config := &cache.WarmupConfig{
    Strategy: cache.WarmupStrategyImmediate,
}
```

#### Scheduled Warming

```go
config := &cache.WarmupConfig{
    Strategy: cache.WarmupStrategyScheduled,
    Interval: 5 * time.Minute,
}
```

#### On-demand Warming

```go
config := &cache.WarmupConfig{
    Strategy: cache.WarmupStrategyOnDemand,
}
```

### Adding Warmup Queries

```go
// Add warmup query
warmupQuery := cache.CacheWarmQuery{
    Query: `
        query {
            systemMetrics {
                storage {
                    totalSpace
                    usedSpace
                }
                performance {
                    averageResponseTime
                }
            }
        }
    `,
    TTL:      2 * time.Minute,
    Priority: 10,
}

cacheWarmer.AddWarmupQuery(warmupQuery)
```

## Cache Analytics

### Performance Metrics

```go
// Get cache metrics
metrics := cacheAnalytics.GetMetrics()

fmt.Printf("Hit Rate: %.2f%%\n", metrics.HitRate*100)
fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
fmt.Printf("Average Response Time: %v\n", metrics.AverageResponseTime)
```

### Performance Report

```go
// Generate comprehensive report
report := cacheAnalytics.GetPerformanceReport()

fmt.Printf("Cache Efficiency Score: %.2f\n", 
    report["efficiency"].(map[string]interface{})["score"])
```

### Insights

```go
// Get cache insights
insights := cacheAnalytics.GenerateInsights()

for _, insight := range insights {
    fmt.Printf("%s: %s\n", insight.Severity, insight.Title)
    fmt.Printf("  %s\n", insight.Description)
    fmt.Printf("  Recommendation: %s\n", insight.Recommendation)
}
```

## Cache Patterns

### Query-based Caching

```go
// Cache query results
func (r *Resolver) Files(ctx context.Context, limit *int) ([]*File, error) {
    // Generate cache key
    cacheKey := fmt.Sprintf("files:limit:%d", *limit)
    
    // Try to get from cache
    if cached, found := queryCache.Get(ctx, cacheKey, nil); found {
        return cached.Data.([]*File), nil
    }
    
    // Execute query
    files, err := r.executeFilesQuery(ctx, limit)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    queryCache.Set(ctx, cacheKey, nil, files, nil, 5*time.Minute)
    
    return files, nil
}
```

### Field-level Caching

```go
// Cache individual fields
func (r *Resolver) File(ctx context.Context, key string) (*File, error) {
    // Cache file metadata
    metadataKey := fmt.Sprintf("file:metadata:%s", key)
    if cached, found := queryCache.Get(ctx, metadataKey, nil); found {
        return cached.Data.(*File), nil
    }
    
    // Execute query
    file, err := r.executeFileQuery(ctx, key)
    if err != nil {
        return nil, err
    }
    
    // Cache with different TTL for metadata
    queryCache.Set(ctx, metadataKey, nil, file, nil, 10*time.Minute)
    
    return file, nil
}
```

### Conditional Caching

```go
// Cache based on conditions
func (r *Resolver) SystemMetrics(ctx context.Context) (*SystemMetrics, error) {
    // Don't cache if system is under high load
    if r.isSystemUnderLoad() {
        return r.executeSystemMetricsQuery(ctx)
    }
    
    // Normal caching logic
    if cached, found := queryCache.Get(ctx, "system:metrics", nil); found {
        return cached.Data.(*SystemMetrics), nil
    }
    
    metrics, err := r.executeSystemMetricsQuery(ctx)
    if err != nil {
        return nil, err
    }
    
    queryCache.Set(ctx, "system:metrics", nil, metrics, nil, 1*time.Minute)
    return metrics, nil
}
```

## Best Practices

### 1. TTL Selection

- **Static Data**: Use longer TTL (15-30 minutes)
- **Dynamic Data**: Use shorter TTL (1-5 minutes)
- **Real-time Data**: Use very short TTL (30 seconds) or no caching

### 2. Cache Key Design

```go
// Good: Descriptive and unique
cacheKey := fmt.Sprintf("user:profile:%s", userID)

// Bad: Too generic
cacheKey := "user"
```

### 3. Invalidation Rules

```go
// Create specific invalidation rules
rules := []*cache.InvalidationRule{
    {
        ID:       "user-profile-update",
        Pattern:  "user:profile:*",
        Strategy: cache.InvalidationStrategyEvent,
        Conditions: []cache.InvalidationCondition{
            {
                Type:     cache.ConditionTypeEvent,
                Field:    "type",
                Operator: cache.OperatorEquals,
                Value:    "user.profile.updated",
            },
        },
    },
}
```

### 4. Cache Warming

```go
// Warm cache with high-priority queries
highPriorityQueries := []cache.CacheWarmQuery{
    {
        Query:    "query { systemMetrics { storage { totalSpace } } }",
        TTL:      2 * time.Minute,
        Priority: 10,
    },
    {
        Query:    "query { nodes { id status } }",
        TTL:      1 * time.Minute,
        Priority: 8,
    },
}
```

### 5. Monitoring

```go
// Monitor cache performance
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        metrics := cacheAnalytics.GetMetrics()
        if metrics.HitRate < 0.5 {
            logger.Warn("Low cache hit rate", "hitRate", metrics.HitRate)
        }
    }
}()
```

## Performance Optimization

### 1. Compression

Enable compression for large cached objects:

```go
config := &cache.QueryCacheConfig{
    EnableCompression: true,
}
```

### 2. Memory Management

Monitor cache size and implement eviction policies:

```go
// Check cache utilization
stats := queryCache.GetCacheStats()
utilization := float64(stats["size"].(int)) / float64(stats["config"].(map[string]interface{})["maxCacheSize"].(int))

if utilization > 0.9 {
    // Implement eviction or increase cache size
}
```

### 3. Concurrent Access

Use appropriate concurrency settings:

```go
warmupConfig := &cache.WarmupConfig{
    Concurrency: runtime.NumCPU(),
}
```

## Troubleshooting

### Common Issues

1. **Low Hit Rate**
   - Check TTL values
   - Verify cache key generation
   - Review invalidation rules

2. **High Memory Usage**
   - Reduce cache size
   - Enable compression
   - Implement eviction policies

3. **Stale Data**
   - Review invalidation events
   - Check TTL values
   - Verify invalidation rules

### Debugging

```go
// Enable debug logging
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Get detailed cache stats
stats := queryCache.GetCacheStats()
fmt.Printf("Cache Stats: %+v\n", stats)

// Get performance insights
insights := cacheAnalytics.GenerateInsights()
for _, insight := range insights {
    fmt.Printf("Insight: %+v\n", insight)
}
```

## API Endpoints

### Cache Management

- `GET /cache/metrics` - Get cache performance metrics
- `GET /cache/stats` - Get detailed cache statistics
- `POST /cache/invalidate` - Invalidate cache entries
- `GET /cache/insights` - Get cache performance insights

### Cache Warming

- `GET /cache/warmup/status` - Get warmup status
- `POST /cache/warmup/start` - Start cache warming
- `POST /cache/warmup/stop` - Stop cache warming
- `POST /cache/warmup/queries` - Add warmup queries

### Analytics

- `GET /cache/analytics/report` - Get performance report
- `GET /cache/analytics/export` - Export metrics
- `GET /cache/analytics/trends` - Get performance trends
