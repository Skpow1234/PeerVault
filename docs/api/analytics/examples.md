# API Analytics Examples

This guide provides practical examples for using the PeerVault API Analytics system.

## Basic Usage

### Getting Started

First, make sure the REST API server is running with analytics enabled:

```bash
./bin/peervault-api
```

All analytics endpoints require authentication:

```bash
export AUTH_TOKEN="Bearer demo-token"
export API_BASE="http://localhost:8081/api/v1"
```

## Example Scenarios

### 1. Daily Operations Dashboard

Get a comprehensive overview of your API's health:

```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/dashboard?period=24h" | jq '.'
```

This returns:
- Overall metrics (requests, errors, latency)
- Top endpoints
- Recent trends
- Active users
- System health status

### 2. Performance Monitoring

#### Check Average Response Times
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?period=24h" | \
  jq '.average_duration / 1000000 | . | "Average latency: \(.)ms"'
```

#### Identify Slow Endpoints
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/summary" | \
  jq '.top_endpoints | sort_by(.average_duration) | reverse | .[0:5] | 
      .[] | {endpoint: "\(.method) \(.path)", latency_ms: (.average_duration / 1000000)}'
```

#### Track Latency Trends Over Week
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/trends?period=7d&interval=day" | \
  jq '.[] | {date: .timestamp, avg_latency_ms: .average_duration_ms}'
```

### 3. Error Analysis

#### Get Error Rate
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?period=24h" | \
  jq '{total: .total_requests, failed: .failed_requests, error_rate: .error_rate}'
```

#### Find Endpoints with High Error Rates
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/summary" | \
  jq '.top_endpoints | sort_by(.error_rate) | reverse | .[0:5] | 
      .[] | {endpoint: "\(.method) \(.path)", error_rate: .error_rate, failed: .failed_calls}'
```

#### Query Recent Errors
```bash
# Get last 50 5xx errors
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/calls?status_code=500&limit=50" | \
  jq '.[] | {time: .timestamp, path: .path, duration_ms: (.duration / 1000000)}'

# Get all 4xx errors in last hour
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/calls?period=hour&limit=100" | \
  jq '.[] | select(.status_code >= 400 and .status_code < 500) | 
      {time: .timestamp, status: .status_code, path: .path, user: .user_id}'
```

### 4. User Behavior Analysis

#### Analyze Specific User
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/user?user_id=user123&period=30d" | jq '.'
```

#### Find Most Active Users
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/popularity?period=7d" | \
  jq '.most_active_users | .[0:10] | 
      .[] | {user: .user_id, requests: .request_count, endpoints: .endpoint_count}'
```

#### Track User Activity Pattern
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/user?user_id=user123&period=7d" | \
  jq '{user: .user_id, total_requests: .total_requests, 
       peak_hour: .peak_hour, avg_hourly_rate: .average_request_rate}'
```

#### User's Favorite Endpoints
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/user?user_id=user123&period=30d" | \
  jq '.favorite_endpoints | .[] | 
      {endpoint: "\(.method) \(.path)", calls: .total_calls}'
```

### 5. API Popularity Tracking

#### Top 10 Endpoints
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/popularity?period=7d" | \
  jq '.top_endpoints | .[0:10] | .[] | 
      {endpoint: .endpoint, calls: .call_count, users: .unique_users}'
```

#### Trending Endpoints
```bash
# Trending up
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/popularity?period=7d" | \
  jq '.trending_up | .[] | 
      {endpoint: .endpoint, growth_rate: .growth_rate, 
       yesterday: .previous_day_count, today: .last_day_count}'

# Trending down
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/popularity?period=7d" | \
  jq '.trending_down | .[] | 
      {endpoint: .endpoint, decline_rate: .growth_rate, 
       yesterday: .previous_day_count, today: .last_day_count}'
```

### 6. Traffic Analysis

#### Requests by HTTP Method
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?period=24h" | \
  jq '.requests_by_method'
```

#### Requests by Path
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?period=24h" | \
  jq '.requests_by_path | to_entries | sort_by(.value) | reverse | 
      .[0:10] | .[] | {path: .key, count: .value}'
```

#### Data Transfer Statistics
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?period=24h" | \
  jq '{data_in_mb: (.total_data_in / 1048576), 
       data_out_mb: (.total_data_out / 1048576),
       total_mb: ((.total_data_in + .total_data_out) / 1048576)}'
```

### 7. Trend Analysis

#### Hourly Trends for Last 24 Hours
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/trends?period=24h&interval=hour" | \
  jq '.[] | {hour: .timestamp, requests: .request_count, 
             success: .success_count, errors: .error_count}'
```

#### Daily Trends for Last Month
```bash
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/trends?period=30d&interval=day" | \
  jq '.[] | {date: .timestamp, requests: .request_count, 
             error_rate: ((.error_count / .request_count) * 100)}'
```

#### Compare Current vs Previous Period
```bash
# Get current week stats
CURRENT=$(curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?period=7d" | jq '.total_requests')

# Get previous week stats
START=$(date -d '14 days ago' -u +"%Y-%m-%dT%H:%M:%SZ")
END=$(date -d '7 days ago' -u +"%Y-%m-%dT%H:%M:%SZ")
PREVIOUS=$(curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?start_time=$START&end_time=$END" | \
  jq '.total_requests')

# Calculate growth
echo "Current: $CURRENT, Previous: $PREVIOUS"
echo "Growth: $(echo "scale=2; (($CURRENT - $PREVIOUS) / $PREVIOUS) * 100" | bc)%"
```

### 8. Endpoint Deep Dive

#### Analyze Specific Endpoint
```bash
ENDPOINT="/api/v1/files"
curl -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/endpoint?endpoint=$ENDPOINT&period=7d" | \
  jq '{endpoint: .path, total_calls: .total_calls, 
       success_rate: ((.successful_calls / .total_calls) * 100),
       avg_latency_ms: (.average_duration / 1000000),
       min_latency_ms: (.min_duration / 1000000),
       max_latency_ms: (.max_duration / 1000000)}'
```

#### Compare Multiple Endpoints
```bash
for endpoint in "/api/v1/files" "/api/v1/peers"; do
  echo "=== $endpoint ==="
  curl -s -H "Authorization: $AUTH_TOKEN" \
    "$API_BASE/analytics/endpoint?endpoint=$endpoint&period=24h" | \
    jq '{calls: .total_calls, error_rate: .error_rate, 
         avg_latency_ms: (.average_duration / 1000000)}'
  echo
done
```

### 9. Real-time Monitoring

#### Monitor Recent Activity (Last 5 Minutes)
```bash
while true; do
  clear
  echo "=== Last 5 Minutes ==="
  date
  curl -s -H "Authorization: $AUTH_TOKEN" \
    "$API_BASE/analytics/usage?period=hour" | \
    jq '{requests: .total_requests, errors: .failed_requests, 
         error_rate: .error_rate, avg_latency_ms: (.average_duration / 1000000)}'
  sleep 30
done
```

#### Alert on High Error Rate
```bash
#!/bin/bash
THRESHOLD=5.0

ERROR_RATE=$(curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?period=hour" | jq '.error_rate')

if (( $(echo "$ERROR_RATE > $THRESHOLD" | bc -l) )); then
  echo "ALERT: Error rate is $ERROR_RATE% (threshold: $THRESHOLD%)"
  # Send notification (email, Slack, etc.)
fi
```

### 10. Reporting

#### Generate Daily Report
```bash
#!/bin/bash
DATE=$(date +%Y-%m-%d)
REPORT_FILE="analytics-report-$DATE.json"

echo "Generating analytics report for $DATE..."

curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/dashboard?period=24h" > "$REPORT_FILE"

echo "Report saved to $REPORT_FILE"

# Extract key metrics
jq '{
  date: "'$DATE'",
  total_requests: .metrics.total_requests,
  error_rate: .metrics.error_rate,
  avg_latency_ms: (.metrics.average_duration / 1000000),
  active_users: .summary.active_users,
  top_endpoint: .metrics.top_endpoints[0].path,
  health_status: .summary.system_health.status
}' "$REPORT_FILE"
```

#### Generate CSV Export
```bash
curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/trends?period=7d&interval=day" | \
  jq -r '["Date","Requests","Success","Errors","Error Rate","Avg Latency (ms)"], 
         (.[] | [.timestamp, .request_count, .success_count, .error_count, 
                 ((.error_count / .request_count) * 100), .average_duration_ms]) | 
         @csv' > trends-report.csv

echo "CSV report saved to trends-report.csv"
```

## Integration Examples

### Prometheus Metrics Export

```bash
#!/bin/bash
# Export analytics to Prometheus format

curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/usage?period=hour" | \
  jq -r '
    "# HELP api_requests_total Total API requests",
    "# TYPE api_requests_total counter",
    "api_requests_total \(.total_requests)",
    "",
    "# HELP api_requests_failed Failed API requests",
    "# TYPE api_requests_failed counter",
    "api_requests_failed \(.failed_requests)",
    "",
    "# HELP api_error_rate_percent API error rate percentage",
    "# TYPE api_error_rate_percent gauge",
    "api_error_rate_percent \(.error_rate)",
    "",
    "# HELP api_latency_seconds Average API latency in seconds",
    "# TYPE api_latency_seconds gauge",
    "api_latency_seconds \(.average_duration / 1000000000)"
  '
```

### Slack Notification

```bash
#!/bin/bash
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

SUMMARY=$(curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/summary" | \
  jq '{
    requests: .overall_metrics.total_requests,
    error_rate: .overall_metrics.error_rate,
    active_users: .active_users,
    status: .system_health.status
  }')

curl -X POST "$SLACK_WEBHOOK" \
  -H 'Content-Type: application/json' \
  -d "{
    \"text\": \"Daily API Analytics Report\",
    \"attachments\": [{
      \"color\": \"good\",
      \"fields\": [
        {\"title\": \"Total Requests\", \"value\": \"$(echo $SUMMARY | jq -r '.requests')\", \"short\": true},
        {\"title\": \"Error Rate\", \"value\": \"$(echo $SUMMARY | jq -r '.error_rate')%\", \"short\": true},
        {\"title\": \"Active Users\", \"value\": \"$(echo $SUMMARY | jq -r '.active_users')\", \"short\": true},
        {\"title\": \"Status\", \"value\": \"$(echo $SUMMARY | jq -r '.status')\", \"short\": true}
      ]
    }]
  }"
```

## Advanced Queries

### Find Slowest Requests
```bash
curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/calls?limit=1000" | \
  jq 'sort_by(.duration) | reverse | .[0:10] | 
      .[] | {time: .timestamp, path: .path, method: .method, 
             duration_ms: (.duration / 1000000), user: .user_id}'
```

### Analyze User Agent Distribution
```bash
curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/calls?limit=1000" | \
  jq 'group_by(.user_agent) | 
      .[] | {user_agent: .[0].user_agent, count: length} | 
      select(.count > 0)' | \
  jq -s 'sort_by(.count) | reverse'
```

### Peak Usage Hours
```bash
curl -s -H "Authorization: $AUTH_TOKEN" \
  "$API_BASE/analytics/trends?period=7d&interval=hour" | \
  jq 'group_by(.timestamp | split("T")[1] | split(":")[0]) | 
      .[] | {hour: .[0].timestamp | split("T")[1] | split(":")[0], 
             avg_requests: ([.[] | .request_count] | add / length)} | 
      select(.avg_requests > 0)' | \
  jq -s 'sort_by(.avg_requests) | reverse'
```

## Tips

1. **Use `jq` for JSON processing**: All examples use `jq` to parse and format JSON responses
2. **Store credentials securely**: Never hardcode API tokens in scripts
3. **Implement caching**: Cache analytics results for dashboard displays
4. **Schedule reports**: Use cron jobs for automated reporting
5. **Monitor continuously**: Set up continuous monitoring for critical metrics

