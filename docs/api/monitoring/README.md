# API Performance Monitoring

Real-time API performance monitoring with alerting and optimization recommendations for the PeerVault REST API.

## Overview

The Performance Monitoring system provides comprehensive real-time tracking of API performance metrics, automated alerting for degraded performance, and actionable optimization recommendations.

## Features

### 1. Response Time Monitoring
- Real-time response time tracking
- P50, P95, P99 percentile calculations
- Per-endpoint performance analysis
- Min/max/mean/standard deviation metrics
- Response time trend analysis

### 2. Throughput Tracking
- Requests per second (RPS) measurement
- Peak throughput detection
- Throughput trends over time
- Data transfer rate monitoring
- Per-endpoint throughput analysis

### 3. Performance Alerting
- Automated alert generation
- Configurable thresholds
- Multiple severity levels (info, warning, critical)
- Alert types:
  - Response time alerts (P95/P99 thresholds)
  - Error rate alerts
  - Throughput degradation alerts
  - System resource alerts
  - Saturation alerts
  - Anomaly detection

### 4. Optimization Recommendations
- Automated recommendation generation
- Prioritized suggestions (high/medium/low)
- Recommendation types:
  - Caching recommendations
  - Code optimization suggestions
  - Query optimization hints
  - Scaling recommendations
  - Configuration improvements
- Impact estimates and effort assessments
- Actionable steps for each recommendation

## API Endpoints

### Core Monitoring

```http
GET /api/v1/monitoring/snapshot
```
Returns current performance snapshot with all metrics.

**Response:**
```json
{
  "timestamp": "2025-11-27T10:30:00Z",
  "overall_response_time": {
    "count": 15234,
    "min": 12000000,
    "max": 850000000,
    "mean": 45000000,
    "median": 38000000,
    "p95": 95000000,
    "p99": 180000000
  },
  "overall_throughput": {
    "requests_per_second": 25.4,
    "requests_per_minute": 1524,
    "peak_rps": 45.2
  },
  "endpoint_stats": [...],
  "system_load": {
    "memory_percent": 68.5,
    "goroutines": 1250,
    "active_connections": 45
  },
  "active_alerts": [],
  "health_score": 92.5
}
```

```http
GET /api/v1/monitoring/realtime
```
Returns real-time metrics for the last minute.

**Response:**
```json
{
  "timestamp": "2025-11-27T10:30:00Z",
  "current_rps": 28.5,
  "average_latency_ms": 42.3,
  "active_requests": 12,
  "error_count": 3,
  "success_count": 1697,
  "last_minute_requests": 1700
}
```

### Detailed Metrics

```http
GET /api/v1/monitoring/response-time?endpoint=/api/v1/files&method=GET
```
Get response time statistics for a specific endpoint.

```http
GET /api/v1/monitoring/throughput?endpoint=/api/v1/files
```
Get throughput statistics for a specific endpoint.

```http
GET /api/v1/monitoring/endpoint?endpoint=/api/v1/files
```
Get comprehensive performance metrics for an endpoint.

**Response:**
```json
{
  "endpoint": "/api/v1/files",
  "method": "GET",
  "response_time": {
    "mean": 45000000,
    "p95": 95000000,
    "p99": 180000000
  },
  "throughput": {
    "requests_per_second": 12.5
  },
  "error_rate": 1.8,
  "success_rate": 98.2,
  "saturation": 45.2
}
```

### Trends & Reports

```http
GET /api/v1/monitoring/trend?endpoint=/api/v1/files&bucket_size=1m
```
Get performance trend data over time.

**Response:**
```json
{
  "endpoint": "/api/v1/files",
  "period": "5m",
  "data_points": [
    {
      "timestamp": "2025-11-27T10:25:00Z",
      "response_time_ms": 42.5,
      "throughput_rps": 12.3,
      "error_rate": 1.5
    }
  ],
  "trend": "stable",
  "volatility": 15.2
}
```

```http
GET /api/v1/monitoring/report?period=5m
```
Generate a comprehensive performance report.

**Response:**
```json
{
  "generated_at": "2025-11-27T10:30:00Z",
  "period": "5m",
  "summary": {
    "total_requests": 15234,
    "average_response_time": 45000000,
    "p95_response_time": 95000000,
    "throughput_rps": 25.4,
    "health_score": 92.5,
    "performance_grade": "A"
  },
  "endpoint_performance": [...],
  "trends": [...],
  "alerts": [...],
  "recommendations": [...]
}
```

### Alerts

```http
GET /api/v1/monitoring/alerts?active_only=true
```
Get performance alerts (active, resolved, or all).

```http
POST /api/v1/monitoring/alerts/silence?alert_id=<alert-id>
```
Silence a specific alert.

```http
POST /api/v1/monitoring/alerts/resolve?alert_id=<alert-id>
```
Manually resolve an alert.

### Recommendations

```http
GET /api/v1/monitoring/recommendations?priority=high&limit=10
```
Get optimization recommendations.

**Query Parameters:**
- `priority`: Filter by priority (high, medium, low)
- `type`: Filter by type (caching, code_optimization, scaling, etc.)
- `limit`: Maximum number of recommendations

**Response:**
```json
[
  {
    "id": "rec-uuid",
    "type": "caching",
    "priority": "high",
    "title": "Implement caching for /api/v1/files",
    "description": "Endpoint has high response time (245ms)...",
    "endpoint": "/api/v1/files",
    "impact_estimate": "Reduce latency by 50-80% (245ms → 75ms)",
    "effort": "Medium",
    "actions": [
      "Identify cacheable data",
      "Implement Redis cache",
      "Set appropriate TTL",
      "Add cache invalidation"
    ],
    "rationale": "Current P95 is 320ms...",
    "metrics": {
      "current_avg_ms": 245,
      "p95_ms": 320
    }
  }
]
```

### Dashboard & Summary

```http
GET /api/v1/monitoring/dashboard?period=5m
```
Comprehensive dashboard with all metrics, alerts, and recommendations.

```http
GET /api/v1/monitoring/health-score
```
Current health score and status.

**Response:**
```json
{
  "health_score": 92.5,
  "grade": "A",
  "status": "excellent",
  "timestamp": "2025-11-27T10:30:00Z",
  "details": {
    "avg_response_time_ms": 45,
    "p95_response_time_ms": 95,
    "throughput_rps": 25.4,
    "active_alerts": 0,
    "goroutines": 1250,
    "memory_percent": 68.5
  }
}
```

```http
GET /api/v1/monitoring/summary
```
Quick metrics summary for dashboards.

## Configuration

Configure monitoring in your REST API server:

```go
config := &rest.Config{
    MonitoringConfig: &monitoring.MonitoringConfig{
        Enabled:    true,
        SampleRate: 1.0,  // Sample 100% of requests
        WindowSize: 5 * time.Minute,
        AlertThresholds: monitoring.AlertThresholds{
            ResponseTimeP95Ms:     500,
            ResponseTimeP99Ms:     1000,
            ErrorRatePercent:      5,
            ThroughputDropPercent: 50,
            SaturationPercent:     80,
            MemoryPercent:         85,
        },
        RecommendationThresholds: monitoring.RecommendationThresholds{
            SlowResponseTimeMs:   200,
            HighErrorRatePercent: 3,
            HighMemoryPercent:    75,
        },
    },
}
```

### Configuration Options

- **Enabled**: Enable/disable monitoring
- **SampleRate**: Percentage of requests to monitor (0.0-1.0)
- **WindowSize**: Time window for metric collection
- **AlertThresholds**: Thresholds for triggering alerts
- **RecommendationThresholds**: Thresholds for generating recommendations

## Usage Examples

Use the `/api/v1/monitoring/*` endpoints from any API client. Command-line examples are omitted in the Docker-only workflow.

## Alert Types & Severity

### Alert Types
- **response_time**: Response time exceeds thresholds
- **throughput**: Throughput degradation detected
- **error_rate**: High error rate detected
- **saturation**: Endpoint approaching capacity
- **system_load**: System resources overloaded
- **anomalous**: Anomalous patterns detected

### Severity Levels
- **info**: Informational alerts
- **warning**: Performance degradation detected
- **critical**: Immediate attention required

## Health Score

The health score (0-100) is calculated based on:
- Response time performance
- Error rates
- Throughput stability
- System resource usage
- Active alerts

**Grading:**
- A (90-100): Excellent performance
- B (80-89): Good performance
- C (70-79): Fair performance
- D (60-69): Poor performance
- F (<60): Critical performance issues

## Integration

### Alert Callbacks

Register callbacks for alert notifications:

```go
service.RegisterAlertCallback(func(alert *monitoring.PerformanceAlert) {
    // Send to Slack, email, PagerDuty, etc.
    log.Printf("ALERT: %s - %s", alert.Type, alert.Message)
})
```

### Custom Metrics

The monitoring system automatically tracks all HTTP requests. No additional instrumentation required.

## Best Practices

1. **Set Appropriate Thresholds**: Adjust alert thresholds based on your SLAs
2. **Monitor Trends**: Watch for degrading trends before they become critical
3. **Act on Recommendations**: Implement high-priority recommendations first
4. **Review Regularly**: Check the dashboard daily
5. **Tune Sample Rate**: Adjust sample rate based on traffic volume

## Troubleshooting

### High Memory Usage
- Reduce `WindowSize` to keep fewer metrics in memory
- Lower `SampleRate` to track fewer requests
- Check for goroutine leaks

### Missing Metrics
- Verify `Enabled: true` in configuration
- Check middleware is properly installed
- Review logs for errors

### False Alerts
- Adjust alert thresholds higher
- Consider traffic patterns (spikes)
- Review sample data for anomalies

## Performance Impact

- **Overhead**: < 1ms per request
- **Memory**: ~100 bytes per tracked metric
- **CPU**: Minimal (async recording)
- **Storage**: In-memory with configurable retention

## Architecture

```
HTTP Request → Monitoring Middleware → Record Metric
                                      ↓
                                  Collector (In-Memory)
                                      ↓
                   ┌──────────────────┼──────────────────┐
                   ↓                  ↓                  ↓
               Alerter           Optimizer          Service
                   ↓                  ↓                  ↓
              Alerts          Recommendations      HTTP Endpoints
```

## See Also

- [API Analytics](../analytics/README.md) - Usage analytics and tracking
- [REST API](../README.md) - Main API documentation
- [Performance Testing](../performance/README.md) - Load testing

---

**Status:** Production Ready
**Version:** 1.0
**Last Updated:** November 27, 2025

