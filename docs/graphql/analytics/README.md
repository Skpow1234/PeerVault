# GraphQL Analytics

This document describes the comprehensive GraphQL analytics and monitoring system implemented in PeerVault, providing detailed insights into query performance, system metrics, and operational intelligence.

## Overview

The GraphQL analytics system provides:

- **Query Performance Tracking**: Monitor query execution times, complexity, and frequency
- **System Performance Monitoring**: Track memory usage, CPU utilization, and throughput
- **Error Analysis**: Detailed error tracking and analysis
- **Real-time Dashboard**: Web-based dashboard for live monitoring
- **Performance Insights**: Automated insights and recommendations
- **Data Export**: Export analytics data in multiple formats

## Architecture

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GraphQL       │    │   Query         │    │   Performance   │
│   Resolvers     │───▶│   Analytics     │    │   Monitor       │
│                 │    │                 │    │                 │
└─────────────────┘    │  - Query Stats  │    │  - System Stats │
                       │  - Complexity   │    │  - Memory Usage │
┌─────────────────┐    │  - Error Track  │    │  - CPU Usage    │
│   Analytics     │◀───┼─────────────────┼───▶│  - Throughput   │
│   Dashboard     │    │                 │    │                 │
│                 │    └─────────────────┘    └─────────────────┘
│  - Real-time    │
│  - Insights     │    ┌─────────────────┐
│  - Export       │    │   Data Export   │
│  - Reports      │    │                 │
└─────────────────┘    │  - JSON         │
                       │  - CSV          │
                       │  - Reports      │
                       └─────────────────┘
```

## Query Analytics

### Basic Usage

```go
// Create query analytics
queryAnalytics := analytics.NewQueryAnalytics(logger)

// Record a query execution
execution := &analytics.QueryExecution{
    Query:      "query { files { id name } }",
    Variables:  map[string]interface{}{"limit": 10},
    StartTime:  time.Now(),
    EndTime:    time.Now().Add(100 * time.Millisecond),
    Duration:   100 * time.Millisecond,
    Complexity: 5,
    Depth:      2,
    FieldCount: 2,
    Source:     "web",
    UserAgent:  "Mozilla/5.0...",
    IPAddress:  "192.168.1.100",
}

queryAnalytics.RecordQuery(execution)
```

### Query Metrics

```go
// Get query statistics
stats := queryAnalytics.GetQueryStats()
fmt.Printf("Total queries: %d\n", stats["totalExecutions"])
fmt.Printf("Average response time: %v\n", stats["averageTime"])
fmt.Printf("Error rate: %.2f%%\n", stats["errorRate"].(float64)*100)

// Get top queries
topQueries := queryAnalytics.GetTopQueries(10)
for _, query := range topQueries {
    fmt.Printf("Query: %s, Count: %d, Avg Time: %v\n", 
        query.Query, query.Count, query.AverageTime)
}

// Get slowest queries
slowQueries := queryAnalytics.GetSlowestQueries(10)
for _, query := range slowQueries {
    fmt.Printf("Slow Query: %s, Avg Time: %v\n", 
        query.Query, query.AverageTime)
}
```

### Query Insights

```go
// Get automated insights
insights := queryAnalytics.GetQueryInsights()
for _, insight := range insights {
    fmt.Printf("%s: %s\n", insight.Severity, insight.Title)
    fmt.Printf("  %s\n", insight.Description)
    fmt.Printf("  Recommendation: %s\n", insight.Recommendation)
}
```

## Performance Monitoring

### Basic Usage

```go
// Create performance monitor
config := analytics.DefaultPerformanceMonitorConfig()
performanceMonitor := analytics.NewPerformanceMonitor(config, logger)

// Record a request
performanceMonitor.RecordRequest(
    150*time.Millisecond, // duration
    true,                 // success
    "",                   // error type
    1024,                 // request size
    10,                   // query complexity
    3,                    // query depth
    5,                    // field count
)
```

### Performance Metrics

```go
// Get performance metrics
metrics := performanceMonitor.GetMetrics()
fmt.Printf("Requests per second: %.2f\n", metrics.RequestsPerSecond)
fmt.Printf("Memory usage: %d MB\n", metrics.MemoryUsage/(1024*1024))
fmt.Printf("CPU usage: %.2f%%\n", metrics.CPUUsage)
fmt.Printf("Error rate: %.2f%%\n", metrics.ErrorRate*100)
```

### Performance Report

```go
// Get comprehensive performance report
report := performanceMonitor.GetPerformanceReport()
fmt.Printf("Performance Report: %+v\n", report)

// Get performance insights
insights := performanceMonitor.GetPerformanceInsights()
for _, insight := range insights {
    fmt.Printf("Performance Insight: %s - %s\n", 
        insight.Severity, insight.Title)
}
```

## Analytics Dashboard

### Starting the Dashboard

```go
// Create dashboard
dashboardConfig := analytics.DefaultDashboardConfig()
dashboard := analytics.NewAnalyticsDashboard(
    queryAnalytics,
    performanceMonitor,
    dashboardConfig,
    logger,
)

// Start dashboard server
go dashboard.Start(context.Background())
```

### Dashboard Features

The analytics dashboard provides:

- **Real-time Metrics**: Live updates of query and performance metrics
- **Query Analysis**: Top queries, slowest queries, and error analysis
- **Performance Charts**: Visual representation of response times and throughput
- **Insights Panel**: Automated insights and recommendations
- **Export Functionality**: Export data in JSON and CSV formats

### Dashboard Endpoints

- `GET /` - Main dashboard page
- `GET /api/metrics` - Comprehensive metrics
- `GET /api/queries` - Query analytics data
- `GET /api/performance` - Performance analytics data
- `GET /api/insights` - Insights and recommendations
- `GET /api/export` - Export analytics data
- `GET /api/health` - Health check

## Configuration

### Query Analytics Configuration

```go
config := &analytics.QueryAnalyticsConfig{
    EnableTracking:     true,
    MaxQueries:         10000,
    RetentionPeriod:    24 * time.Hour,
    EnableComplexity:   true,
    EnableDepth:        true,
    EnableFieldCount:   true,
    EnableUserTracking: false,
    EnableIPTracking:   false,
}
```

### Performance Monitor Configuration

```go
config := &analytics.PerformanceMonitorConfig{
    EnableSystemMetrics:   true,
    EnableQueryMetrics:    true,
    EnableErrorTracking:   true,
    EnableThroughputTracking: true,
    UpdateInterval:        1 * time.Second,
    RetentionPeriod:       1 * time.Hour,
    RPSWindowSize:         1 * time.Minute,
    ResponseTimeBuckets: []time.Duration{
        1 * time.Millisecond,
        10 * time.Millisecond,
        100 * time.Millisecond,
        1 * time.Second,
        10 * time.Second,
    },
}
```

### Dashboard Configuration

```go
config := &analytics.DashboardConfig{
    Port:            8082,
    EnableDashboard: true,
    RefreshInterval: 5 * time.Second,
    EnableExport:    true,
    EnableRealTime:  true,
}
```

## Integration with GraphQL Server

### Middleware Integration

```go
// Create analytics components
queryAnalytics := analytics.NewQueryAnalytics(logger)
performanceMonitor := analytics.NewPerformanceMonitor(config, logger)

// Create GraphQL server with analytics middleware
server := graphql.NewServer(fileserver, config)

// Add analytics middleware
server.UseAnalytics(queryAnalytics, performanceMonitor)
```

### Resolver Integration

```go
func (r *Resolver) Files(ctx context.Context, limit *int) ([]*File, error) {
    start := time.Now()
    
    // Execute query
    files, err := r.executeFilesQuery(ctx, limit)
    
    // Record analytics
    execution := &analytics.QueryExecution{
        Query:     "query { files { id name } }",
        StartTime: start,
        EndTime:   time.Now(),
        Duration:  time.Since(start),
        Error:     err,
        Complexity: 5,
        Depth:     2,
        FieldCount: 2,
    }
    
    queryAnalytics.RecordQuery(execution)
    performanceMonitor.RecordRequest(
        execution.Duration,
        err == nil,
        getErrorType(err),
        1024,
        execution.Complexity,
        execution.Depth,
        execution.FieldCount,
    )
    
    return files, err
}
```

## Advanced Features

### Custom Metrics

```go
// Add custom metrics to query execution
execution := &analytics.QueryExecution{
    Query:     "query { files { id name } }",
    StartTime: start,
    EndTime:   time.Now(),
    Duration:  time.Since(start),
    Metadata: map[string]interface{}{
        "user_id":    "user_123",
        "tenant_id":  "tenant_456",
        "api_version": "v1",
    },
}
```

### Query Complexity Analysis

```go
// Analyze query complexity
complexity := calculateQueryComplexity(query)
if complexity > 1000 {
    logger.Warn("High complexity query detected", 
        "complexity", complexity, 
        "query", query)
}
```

### Performance Alerts

```go
// Set up performance alerts
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        metrics := performanceMonitor.GetMetrics()
        
        // Alert on high error rate
        if metrics.ErrorRate > 0.05 {
            sendAlert("High error rate detected", 
                "errorRate", metrics.ErrorRate)
        }
        
        // Alert on slow response times
        if metrics.AverageResponseTime > 1*time.Second {
            sendAlert("Slow response times detected", 
                "avgResponseTime", metrics.AverageResponseTime)
        }
    }
}()
```

## Data Export

### JSON Export

```go
// Export analytics data as JSON
data, err := dashboard.exportJSON()
if err != nil {
    log.Fatal("Export failed:", err)
}

// Save to file
err = os.WriteFile("analytics.json", data, 0644)
```

### CSV Export

```go
// Export analytics data as CSV
data, err := dashboard.exportCSV()
if err != nil {
    log.Fatal("Export failed:", err)
}

// Save to file
err = os.WriteFile("analytics.csv", data, 0644)
```

### API Export

```bash
# Export via API
curl -o analytics.json http://localhost:8082/api/export?format=json
curl -o analytics.csv http://localhost:8082/api/export?format=csv
```

## Monitoring and Alerting

### Health Checks

```go
// Check analytics system health
func checkAnalyticsHealth() error {
    // Check if analytics are collecting data
    stats := queryAnalytics.GetQueryStats()
    if stats["totalExecutions"].(int64) == 0 {
        return fmt.Errorf("no query data collected")
    }
    
    // Check if performance monitor is working
    metrics := performanceMonitor.GetMetrics()
    if metrics.Uptime == 0 {
        return fmt.Errorf("performance monitor not working")
    }
    
    return nil
}
```

### Automated Alerts

```go
// Set up automated alerting
func setupAlerts() {
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        for range ticker.C {
            // Check error rate
            metrics := performanceMonitor.GetMetrics()
            if metrics.ErrorRate > 0.1 {
                sendAlert("High error rate", metrics.ErrorRate)
            }
            
            // Check response times
            if metrics.AverageResponseTime > 2*time.Second {
                sendAlert("Slow response times", metrics.AverageResponseTime)
            }
            
            // Check memory usage
            if metrics.MemoryUsage > 2*1024*1024*1024 { // 2GB
                sendAlert("High memory usage", metrics.MemoryUsage)
            }
        }
    }()
}
```

## Best Practices

### 1. Performance Impact

- **Minimal Overhead**: Analytics should have minimal impact on query performance
- **Async Processing**: Process analytics data asynchronously when possible
- **Sampling**: Consider sampling for high-volume scenarios
- **Caching**: Cache frequently accessed analytics data

### 2. Data Retention

```go
// Set appropriate retention periods
config := &analytics.QueryAnalyticsConfig{
    RetentionPeriod: 7 * 24 * time.Hour, // 7 days
}

// Clean up old data
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        queryAnalytics.Cleanup(7 * 24 * time.Hour)
    }
}()
```

### 3. Privacy and Security

```go
// Disable sensitive tracking in production
config := &analytics.QueryAnalyticsConfig{
    EnableUserTracking: false, // Don't track user data
    EnableIPTracking:   false, // Don't track IP addresses
}
```

### 4. Monitoring

```go
// Monitor analytics system itself
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        stats := queryAnalytics.GetQueryStats()
        if stats["totalExecutions"].(int64) == 0 {
            logger.Warn("No query data collected in the last minute")
        }
    }
}()
```

## Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Reduce retention period
   - Implement data sampling
   - Use data compression

2. **Performance Impact**
   - Use async processing
   - Implement sampling
   - Cache analytics data

3. **Missing Data**
   - Check analytics configuration
   - Verify middleware integration
   - Check for errors in analytics code

### Debugging

```go
// Enable debug logging
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Check analytics status
stats := queryAnalytics.GetQueryStats()
fmt.Printf("Analytics Status: %+v\n", stats)

// Check performance metrics
metrics := performanceMonitor.GetMetrics()
fmt.Printf("Performance Metrics: %+v\n", metrics)
```

## API Reference

### Query Analytics API

- `RecordQuery(execution *QueryExecution)` - Record a query execution
- `GetQueryStats() map[string]interface{}` - Get overall query statistics
- `GetTopQueries(limit int) []*QueryMetrics` - Get most frequent queries
- `GetSlowestQueries(limit int) []*QueryMetrics` - Get slowest queries
- `GetQueryInsights() []QueryInsight` - Get automated insights

### Performance Monitor API

- `RecordRequest(duration, success, errorType, requestSize, complexity, depth, fieldCount)` - Record a request
- `GetMetrics() *PerformanceMetrics` - Get performance metrics
- `GetPerformanceReport() map[string]interface{}` - Get comprehensive report
- `GetPerformanceInsights() []PerformanceInsight` - Get performance insights

### Dashboard API

- `Start(ctx context.Context) error` - Start the dashboard server
- `exportJSON() ([]byte, error)` - Export data as JSON
- `exportCSV() ([]byte, error)` - Export data as CSV
