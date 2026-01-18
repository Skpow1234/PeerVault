# Analytics Integration Guide

This guide explains how to integrate the API Analytics & Monitoring system into your PeerVault deployment.

## Quick Start

### 1. Enable Analytics

Analytics is enabled by default. To configure it, modify your server configuration:

```go
import (
    "github.com/Skpow1234/Peervault/internal/api/analytics"
    "github.com/Skpow1234/Peervault/internal/api/rest"
)

config := rest.DefaultConfig()
config.AnalyticsConfig = &analytics.Config{
    Enabled:           true,
    MaxEntries:        100000,
    RetentionDays:     30,
    CleanupInterval:   24 * time.Hour,
    EnableUserTracking: true,
}

server := rest.NewServer(config, logger)
```

### 2. Access Analytics

Once the server is running, analytics are available at:

- Summary: `GET /api/v1/analytics/summary`
- Dashboard: `GET /api/v1/analytics/dashboard`
- Usage Metrics: `GET /api/v1/analytics/usage`
- All endpoints: See [API Documentation](./README.md)

### 3. Verify Integration

Verification runs in CI using containerized workflows.

## Integration Patterns

### Pattern 1: Real-time Monitoring Dashboard

Create a dashboard that displays real-time analytics:

```javascript
// Example: React Dashboard Component
import React, { useState, useEffect } from 'react';

function AnalyticsDashboard() {
  const [data, setData] = useState(null);
  
  useEffect(() => {
    const fetchAnalytics = async () => {
      const response = await fetch('/api/v1/analytics/dashboard?period=24h', {
        headers: { 'Authorization': 'Bearer ' + token }
      });
      const data = await response.json();
      setData(data);
    };
    
    // Fetch initially
    fetchAnalytics();
    
    // Refresh every 30 seconds
    const interval = setInterval(fetchAnalytics, 30000);
    return () => clearInterval(interval);
  }, []);
  
  if (!data) return <div>Loading...</div>;
  
  return (
    <div className="dashboard">
      <div className="metric-card">
        <h3>Total Requests</h3>
        <p>{data.metrics.total_requests}</p>
      </div>
      <div className="metric-card">
        <h3>Error Rate</h3>
        <p>{data.metrics.error_rate.toFixed(2)}%</p>
      </div>
      <div className="metric-card">
        <h3>Active Users</h3>
        <p>{data.summary.active_users}</p>
      </div>
      <div className="metric-card">
        <h3>Health Status</h3>
        <p className={data.summary.system_health.status}>
          {data.summary.system_health.status}
        </p>
      </div>
    </div>
  );
}
```

### Pattern 2: Alerting System

Set up automated alerts based on analytics:

```go
package monitoring

import (
    "time"
    "github.com/Skpow1234/Peervault/internal/api/analytics"
)

type AlertConfig struct {
    ErrorRateThreshold    float64
    LatencyThresholdMs    float64
    RequestRateMin        float64
    CheckInterval         time.Duration
}

type Alerter struct {
    service *analytics.Service
    config  *AlertConfig
}

func (a *Alerter) MonitorHealth() {
    ticker := time.NewTicker(a.config.CheckInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        summary, err := a.service.GetAnalyticsSummary()
        if err != nil {
            continue
        }
        
        // Check error rate
        if summary.SystemHealth.ErrorRate > a.config.ErrorRateThreshold {
            a.sendAlert("High Error Rate", 
                fmt.Sprintf("Error rate is %.2f%%, threshold is %.2f%%",
                    summary.SystemHealth.ErrorRate,
                    a.config.ErrorRateThreshold))
        }
        
        // Check latency
        if summary.SystemHealth.AverageLatency > a.config.LatencyThresholdMs {
            a.sendAlert("High Latency",
                fmt.Sprintf("Average latency is %.2fms, threshold is %.2fms",
                    summary.SystemHealth.AverageLatency,
                    a.config.LatencyThresholdMs))
        }
        
        // Check if system is unhealthy
        if summary.SystemHealth.Status == "unhealthy" {
            a.sendAlert("System Unhealthy",
                fmt.Sprintf("System status: %s, warnings: %v",
                    summary.SystemHealth.Status,
                    summary.SystemHealth.Warnings))
        }
    }
}

func (a *Alerter) sendAlert(title, message string) {
    // Send to Slack, email, PagerDuty, etc.
    log.Printf("ALERT: %s - %s", title, message)
}
```

### Pattern 3: Export to External Systems

Export analytics to Prometheus, Grafana, or other monitoring tools:

```go
package exporters

import (
    "fmt"
    "time"
    "github.com/Skpow1234/Peervault/internal/api/analytics"
    "github.com/prometheus/client_golang/prometheus"
)

type PrometheusExporter struct {
    service *analytics.Service
    
    requestsTotal *prometheus.CounterVec
    errorRate     prometheus.Gauge
    latency       prometheus.Histogram
}

func NewPrometheusExporter(service *analytics.Service) *PrometheusExporter {
    exporter := &PrometheusExporter{
        service: service,
        requestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "api_requests_total",
                Help: "Total number of API requests",
            },
            []string{"method", "endpoint", "status"},
        ),
        errorRate: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "api_error_rate",
                Help: "Current API error rate percentage",
            },
        ),
        latency: prometheus.NewHistogram(
            prometheus.HistogramOpts{
                Name:    "api_latency_seconds",
                Help:    "API request latency in seconds",
                Buckets: prometheus.DefBuckets,
            },
        ),
    }
    
    prometheus.MustRegister(exporter.requestsTotal)
    prometheus.MustRegister(exporter.errorRate)
    prometheus.MustRegister(exporter.latency)
    
    return exporter
}

func (e *PrometheusExporter) UpdateMetrics() error {
    now := time.Now()
    metrics, err := e.service.GetUsageMetrics(now.Add(-5*time.Minute), now)
    if err != nil {
        return err
    }
    
    // Update error rate
    e.errorRate.Set(metrics.ErrorRate)
    
    // Update request counts by method and endpoint
    for endpoint, stats := range metrics.RequestsByPath {
        e.requestsTotal.WithLabelValues("GET", endpoint, "200").Add(float64(stats))
    }
    
    return nil
}
```

### Pattern 4: Custom Analytics Queries

Build custom analytics queries for specific use cases:

```go
package custom

import (
    "time"
    "github.com/Skpow1234/Peervault/internal/api/analytics"
)

type CustomAnalytics struct {
    service *analytics.Service
}

// GetSlowEndpoints returns endpoints that are slower than threshold
func (c *CustomAnalytics) GetSlowEndpoints(thresholdMs float64, period string) ([]*analytics.EndpointStats, error) {
    startTime, endTime := analytics.GetTimeWindow(period)
    metrics, err := c.service.GetUsageMetrics(startTime, endTime)
    if err != nil {
        return nil, err
    }
    
    var slowEndpoints []*analytics.EndpointStats
    for _, endpoint := range metrics.TopEndpoints {
        latencyMs := float64(endpoint.AverageDuration.Milliseconds())
        if latencyMs > thresholdMs {
            ep := endpoint
            slowEndpoints = append(slowEndpoints, &ep)
        }
    }
    
    return slowEndpoints, nil
}

// GetUserActivity returns activity summary for multiple users
func (c *CustomAnalytics) GetUserActivity(userIDs []string, period string) (map[string]*analytics.UserBehaviorMetrics, error) {
    startTime, endTime := analytics.GetTimeWindow(period)
    
    results := make(map[string]*analytics.UserBehaviorMetrics)
    for _, userID := range userIDs {
        behavior, err := c.service.GetUserBehavior(userID, startTime, endTime)
        if err != nil {
            continue
        }
        results[userID] = behavior
    }
    
    return results, nil
}

// GetEndpointComparison compares performance of multiple endpoints
func (c *CustomAnalytics) GetEndpointComparison(endpoints []string, period string) (map[string]*analytics.EndpointStats, error) {
    startTime, endTime := analytics.GetTimeWindow(period)
    
    results := make(map[string]*analytics.EndpointStats)
    for _, endpoint := range endpoints {
        stats, err := c.service.GetEndpointStats(endpoint, startTime, endTime)
        if err != nil {
            continue
        }
        results[endpoint] = stats
    }
    
    return results, nil
}
```

## Configuration Options

### Memory Configuration

For high-traffic systems, adjust memory settings:

```go
config.AnalyticsConfig = &analytics.Config{
    MaxEntries:    500000,  // Increase for more data retention
    RetentionDays: 90,      // Keep data longer
}
```

### Storage Options

For production deployments, consider implementing custom storage:

```go
// Implement the Storage interface for database persistence
type PostgresStorage struct {
    db *sql.DB
}

func (s *PostgresStorage) RecordAPICall(call *analytics.APICall) error {
    // Store in PostgreSQL
}

func (s *PostgresStorage) GetUsageMetrics(startTime, endTime time.Time) (*analytics.UsageMetrics, error) {
    // Query from PostgreSQL
}

// ... implement other methods
```

### User Tracking

Control user tracking based on privacy requirements:

```go
config.AnalyticsConfig = &analytics.Config{
    EnableUserTracking: false,  // Disable user-specific tracking
}
```

## Best Practices

### 1. Performance Optimization

- Use appropriate time windows to limit query scope
- Implement caching for frequently accessed analytics
- Consider read replicas for analytics queries in production

### 2. Security

- Protect analytics endpoints with authentication
- Implement role-based access control for sensitive metrics
- Sanitize user data before storing in analytics

### 3. Data Management

- Set appropriate retention periods based on compliance requirements
- Regularly backup analytics data
- Monitor storage usage and adjust MaxEntries as needed

### 4. Monitoring

- Set up alerts for critical metrics
- Monitor analytics system performance
- Track analytics API usage

### 5. Privacy Compliance

- Anonymize IP addresses if required
- Implement data deletion on user request
- Document data collection in privacy policy

## Troubleshooting

### High Memory Usage

**Problem**: Analytics consuming too much memory

**Solution**:
```go
// Reduce max entries
config.AnalyticsConfig.MaxEntries = 50000

// Reduce retention period
config.AnalyticsConfig.RetentionDays = 7

// Increase cleanup frequency
config.AnalyticsConfig.CleanupInterval = 6 * time.Hour
```

### Missing Analytics Data

**Problem**: No analytics data appearing

**Solution**:
1. Verify analytics is enabled: `config.AnalyticsConfig.Enabled = true`
2. Check middleware is applied to routes
3. Verify authentication tokens are correct
4. Check server logs for errors

### Slow Analytics Queries

**Problem**: Analytics endpoints are slow

**Solution**:
1. Use shorter time windows
2. Implement result caching
3. Add indexes if using database storage
4. Consider pre-aggregation for common queries

## Production Deployment

### Recommended Configuration

```go
config := rest.DefaultConfig()
config.AnalyticsConfig = &analytics.Config{
    Enabled:           true,
    MaxEntries:        200000,
    RetentionDays:     90,
    CleanupInterval:   12 * time.Hour,
    EnableUserTracking: true,
}
```

### Scaling Considerations

1. **High Traffic**: Use database-backed storage instead of in-memory
2. **Multiple Servers**: Centralize analytics collection
3. **Data Analysis**: Export to dedicated analytics platform
4. **Reporting**: Pre-generate reports for common queries

## Additional Resources

- [API Documentation](./README.md)
- [Examples](./examples.md)
- [Testing Guide](./testing.md)
- [Main REST API Documentation](../README.md)

