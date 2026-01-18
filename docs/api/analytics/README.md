# API Analytics & Monitoring

Comprehensive API usage analytics with detailed metrics and insights for the PeerVault REST API.

## Overview

The API Analytics system provides real-time tracking and analysis of API usage patterns, user behavior, endpoint performance, and system health. It enables data-driven decisions about API design, performance optimization, and resource allocation.

## Features

### 1. Usage Metrics
- **Total Requests**: Track overall API request volume
- **Success/Error Rates**: Monitor API reliability
- **Request Distribution**: Analyze requests by method, path, and status code
- **Data Transfer**: Track incoming and outgoing data volumes
- **Response Times**: Measure average, min, and max latencies

### 2. User Behavior Analysis
- **User Activity**: Track individual user patterns
- **Peak Hours**: Identify when users are most active
- **Favorite Endpoints**: Discover most-used API endpoints per user
- **User Agents**: Analyze client applications and devices
- **Geographic Data**: Track user IP addresses and locations

### 3. API Popularity Tracking
- **Top Endpoints**: Identify most popular API endpoints
- **Trending Up/Down**: Detect emerging or declining endpoint usage
- **Growth Rates**: Calculate endpoint popularity changes over time
- **Active Users**: Monitor unique users accessing the API
- **User Engagement**: Track endpoint usage per user

### 4. Usage Trends
- **Time-Series Analysis**: View usage patterns over time
- **Hourly/Daily/Weekly Trends**: Analyze at different granularities
- **Error Trends**: Track error rates over time
- **Performance Trends**: Monitor latency trends
- **Growth Analysis**: Measure API adoption rates

## API Endpoints

### Get Analytics Summary
```http
GET /api/v1/analytics/summary
```

Returns a comprehensive overview of all analytics data.

**Response:**
```json
{
  "overall_metrics": {
    "total_requests": 15234,
    "successful_requests": 14890,
    "failed_requests": 344,
    "average_duration": 45000000,
    "error_rate": 2.26
  },
  "top_endpoints": [...],
  "recent_trends": [...],
  "active_users": 127,
  "new_users_today": 12,
  "system_health": {
    "status": "healthy",
    "error_rate": 2.26,
    "average_latency_ms": 45,
    "request_rate": 0.176
  },
  "generated_at": "2024-11-27T10:30:00Z"
}
```

### Get Usage Metrics
```http
GET /api/v1/analytics/usage?period=24h
GET /api/v1/analytics/usage?start_time=2024-11-27T00:00:00Z&end_time=2024-11-27T23:59:59Z
```

Returns aggregated usage metrics for a time period.

**Query Parameters:**
- `period`: Time window (hour, day, 24h, week, 7d, month, 30d, year)
- `start_time`: Start time (RFC3339 format or Unix timestamp)
- `end_time`: End time (RFC3339 format or Unix timestamp)

**Response:**
```json
{
  "total_requests": 15234,
  "successful_requests": 14890,
  "failed_requests": 344,
  "average_duration": 45000000,
  "total_data_in": 52428800,
  "total_data_out": 104857600,
  "requests_by_method": {
    "GET": 10234,
    "POST": 3456,
    "PUT": 987,
    "DELETE": 557
  },
  "requests_by_path": {
    "/api/v1/files": 8765,
    "/api/v1/peers": 4321,
    "/health": 2148
  },
  "top_endpoints": [
    {
      "path": "/api/v1/files",
      "method": "GET",
      "total_calls": 5432,
      "successful_calls": 5400,
      "failed_calls": 32,
      "average_duration": 42000000,
      "error_rate": 0.59
    }
  ],
  "error_rate": 2.26,
  "time_window": {
    "start": "2024-11-26T10:30:00Z",
    "end": "2024-11-27T10:30:00Z"
  }
}
```

### Get Endpoint Statistics
```http
GET /api/v1/analytics/endpoint?endpoint=/api/v1/files&period=7d
```

Returns detailed statistics for a specific endpoint.

**Query Parameters:**
- `endpoint`: The endpoint path to analyze (required)
- `period`: Time window (default: 24h)

**Response:**
```json
{
  "path": "/api/v1/files",
  "method": "GET",
  "total_calls": 5432,
  "successful_calls": 5400,
  "failed_calls": 32,
  "average_duration": 42000000,
  "min_duration": 12000000,
  "max_duration": 350000000,
  "total_data_in": 26214400,
  "total_data_out": 52428800,
  "error_rate": 0.59,
  "last_called": "2024-11-27T10:28:45Z"
}
```

### Get User Behavior
```http
GET /api/v1/analytics/user?user_id=user123&period=30d
```

Returns behavior analysis for a specific user.

**Query Parameters:**
- `user_id`: User identifier (required)
- `period`: Time window (default: 24h)

**Response:**
```json
{
  "user_id": "user123",
  "total_requests": 1234,
  "unique_endpoints": 15,
  "average_request_rate": 12.5,
  "peak_hour": 14,
  "favorite_endpoints": [
    {
      "path": "/api/v1/files",
      "method": "GET",
      "total_calls": 456
    }
  ],
  "user_agents": {
    "Mozilla/5.0": 890,
    "curl/7.68.0": 344
  },
  "ip_addresses": {
    "192.168.1.100": 1234
  },
  "first_seen": "2024-10-28T10:30:00Z",
  "last_seen": "2024-11-27T10:28:45Z",
  "error_count": 23
}
```

### Get Usage Trends
```http
GET /api/v1/analytics/trends?period=7d&interval=day
```

Returns usage trends over time.

**Query Parameters:**
- `period`: Time window (default: 24h)
- `interval`: Aggregation interval (hour, day, week; default: hour)

**Response:**
```json
[
  {
    "timestamp": "2024-11-27T00:00:00Z",
    "request_count": 2345,
    "success_count": 2290,
    "error_count": 55,
    "average_duration_ms": 45.3,
    "data_in": 7340032,
    "data_out": 14680064
  },
  {
    "timestamp": "2024-11-27T01:00:00Z",
    "request_count": 2123,
    "success_count": 2080,
    "error_count": 43,
    "average_duration_ms": 42.8,
    "data_in": 6815744,
    "data_out": 13631488
  }
]
```

### Get Popularity Metrics
```http
GET /api/v1/analytics/popularity?period=7d
```

Returns endpoint popularity and trending analysis.

**Query Parameters:**
- `period`: Time window (default: 24h)

**Response:**
```json
{
  "top_endpoints": [
    {
      "endpoint": "GET /api/v1/files",
      "call_count": 5432,
      "unique_users": 89,
      "growth_rate": 15.3,
      "trend_score": 6263.56,
      "last_day_count": 2890,
      "previous_day_count": 2542
    }
  ],
  "trending_up": [
    {
      "endpoint": "POST /api/v1/files",
      "call_count": 1234,
      "growth_rate": 45.2
    }
  ],
  "trending_down": [
    {
      "endpoint": "GET /api/v1/peers",
      "call_count": 987,
      "growth_rate": -12.5
    }
  ],
  "most_active_users": [
    {
      "user_id": "user123",
      "request_count": 1234,
      "endpoint_count": 15,
      "last_activity": "2024-11-27T10:28:45Z"
    }
  ],
  "time_window": {
    "start": "2024-11-20T10:30:00Z",
    "end": "2024-11-27T10:30:00Z"
  }
}
```

### Query API Calls
```http
GET /api/v1/analytics/calls?method=GET&status_code=200&limit=100
```

Returns raw API call records matching query parameters.

**Query Parameters:**
- `start_time`: Start time filter
- `end_time`: End time filter
- `endpoint`: Endpoint filter
- `method`: HTTP method filter
- `user_id`: User ID filter
- `status_code`: Status code filter
- `limit`: Maximum results (default: 100)
- `offset`: Results offset for pagination

**Response:**
```json
[
  {
    "id": "call-uuid",
    "timestamp": "2024-11-27T10:28:45Z",
    "method": "GET",
    "path": "/api/v1/files",
    "status_code": 200,
    "duration": 42000000,
    "user_id": "user123",
    "user_agent": "Mozilla/5.0",
    "ip_address": "192.168.1.100",
    "request_size": 1024,
    "response_size": 5120,
    "api_version": "v1",
    "endpoint": "/api/v1/files"
  }
]
```

### Get Dashboard
```http
GET /api/v1/analytics/dashboard?period=24h
```

Returns comprehensive dashboard data combining multiple analytics views.

**Query Parameters:**
- `period`: Time window (default: 24h)

**Response:**
```json
{
  "summary": { /* Analytics summary */ },
  "metrics": { /* Usage metrics */ },
  "popularity": { /* Popularity metrics */ },
  "trends": [ /* Usage trends */ ]
}
```

## Configuration

Analytics can be configured in the REST API server configuration:

```go
config := &rest.Config{
    AnalyticsConfig: &analytics.Config{
        Enabled:           true,
        MaxEntries:        100000,
        RetentionDays:     30,
        CleanupInterval:   24 * time.Hour,
        EnableUserTracking: true,
    },
}
```

### Configuration Options

- **Enabled**: Enable/disable analytics tracking
- **MaxEntries**: Maximum number of API call records to store in memory
- **RetentionDays**: Number of days to retain analytics data
- **CleanupInterval**: Interval for automatic cleanup of old data
- **EnableUserTracking**: Enable/disable user-specific analytics

## Architecture

### Components

1. **Middleware**: Captures API call metadata automatically
2. **Storage**: In-memory storage with configurable retention
3. **Service**: Business logic for analytics calculations
4. **Handler**: HTTP endpoints for querying analytics
5. **API**: High-level analytics API for complex queries

### Data Flow

```
HTTP Request → Analytics Middleware → Record API Call
                                      ↓
                                  Storage (In-Memory)
                                      ↓
                      Service (Aggregation & Analysis)
                                      ↓
                      Handler (HTTP Endpoints)
                                      ↓
                      JSON Response
```

## Usage Examples

Use the `/api/v1/analytics/*` endpoints from any API client. Command-line examples are omitted in the Docker-only workflow.

### Get Popularity Report
Use `/api/v1/analytics/popularity` with your API client.

### Query Recent Errors
Use `/api/v1/analytics/calls` with your API client.

## Best Practices

### 1. Set Appropriate Retention
- Balance storage requirements with data needs
- Consider your traffic volume when setting `MaxEntries`
- Use longer retention for historical trend analysis

### 2. Monitor System Health
- Check the analytics summary regularly
- Set up alerts for high error rates
- Monitor latency trends

### 3. Optimize Performance
- Use time windows to limit query scope
- Apply filters to reduce result sets
- Use pagination for large result sets

### 4. Privacy Considerations
- Review user tracking requirements
- Consider anonymizing IP addresses
- Implement data retention policies

### 5. Integration
- Export analytics to external monitoring systems
- Use webhooks for real-time alerts
- Integrate with business intelligence tools

## Metrics Glossary

- **Total Requests**: Total number of API calls received
- **Success Rate**: Percentage of requests with 2xx or 3xx status codes
- **Error Rate**: Percentage of requests with 4xx or 5xx status codes
- **Average Duration**: Mean response time across all requests
- **P50/P95/P99 Latency**: 50th/95th/99th percentile response times
- **Growth Rate**: Percentage change in usage between time periods
- **Trend Score**: Weighted metric combining volume and growth
- **Active Users**: Unique users making requests in the time window

## Troubleshooting

### Analytics Not Recording
- Check if analytics is enabled in configuration
- Verify middleware is properly configured
- Check logs for error messages

### High Memory Usage
- Reduce `MaxEntries` configuration
- Decrease `RetentionDays`
- Increase `CleanupInterval` frequency

### Missing User Data
- Ensure `EnableUserTracking` is true
- Verify user ID is provided in requests (via header or token)
- Check middleware user extraction logic

## Future Enhancements

- Database-backed storage for larger datasets
- Real-time analytics streaming
- Advanced anomaly detection
- Custom metric definitions
- Export to Prometheus/Grafana
- Machine learning-based predictions

## Support

For issues or questions about API Analytics:
- See the main [API Documentation](../README.md)
- Check the [Testing Guide](../testing/README.md)
- Review example usage in `examples/`

