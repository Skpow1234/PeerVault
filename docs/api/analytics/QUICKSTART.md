# API Analytics - Quick Start Guide

Get started with API Analytics & Monitoring in under 5 minutes!

## Step 1: Start the API Server

```bash
# Build if you haven't already
go build -o bin/ ./cmd/...

# Start the REST API server (analytics enabled by default)
./bin/peervault-api
```

The server will start on `http://localhost:8081` with analytics automatically enabled.

## Step 2: Generate Some Traffic

Open a new terminal and make a few API calls:

```bash
# Set your auth token
export AUTH="Authorization: Bearer demo-token"
export API="http://localhost:8081/api/v1"

# Make some test requests
for i in {1..20}; do
  curl -s -H "$AUTH" "$API/files" > /dev/null
  echo "Request $i sent"
done
```

## Step 3: View Analytics

Now check out your analytics:

### Get a Summary
```bash
curl -H "$AUTH" "$API/analytics/summary" | jq '.'
```

This shows:
- Total requests
- Success/error rates
- Average latency
- System health
- Top endpoints

### View the Dashboard
```bash
curl -H "$AUTH" "$API/analytics/dashboard?period=24h" | jq '.'
```

The dashboard provides a comprehensive overview combining:
- Summary metrics
- Usage statistics
- Popularity data
- Trends

### Check Usage Trends
```bash
curl -H "$AUTH" "$API/analytics/trends?period=hour&interval=5min" | jq '.'
```

## Step 4: Run the Demo

For an interactive demonstration:

```bash
./examples/analytics-demo.sh
```

This will:
1. Generate sample traffic
2. Display various analytics views
3. Show trending endpoints
4. Export data to CSV
5. Check system health

## Key Endpoints

| Endpoint | Purpose |
|----------|---------|
| `/analytics/summary` | Overall health and metrics |
| `/analytics/dashboard` | Comprehensive dashboard |
| `/analytics/usage` | Usage metrics for a period |
| `/analytics/trends` | Time-series trends |
| `/analytics/popularity` | Popular and trending endpoints |
| `/analytics/endpoint` | Stats for specific endpoint |
| `/analytics/user` | User behavior analysis |
| `/analytics/calls` | Query raw API calls |

## Query Parameters

All endpoints support:
- `period` - Time window: `hour`, `24h`, `7d`, `30d`
- `start_time` - Custom start time (RFC3339)
- `end_time` - Custom end time (RFC3339)

Additional parameters by endpoint:
- `/analytics/trends` - `interval`: `hour`, `day`, `week`
- `/analytics/endpoint` - `endpoint`: path to analyze
- `/analytics/user` - `user_id`: user to analyze
- `/analytics/calls` - `method`, `status_code`, `limit`, `offset`

## Configuration

Customize analytics in your server configuration:

```go
config := rest.DefaultConfig()
config.AnalyticsConfig = &analytics.Config{
    Enabled:           true,
    MaxEntries:        100000,
    RetentionDays:     30,
    CleanupInterval:   24 * time.Hour,
    EnableUserTracking: true,
}
```

## Real-World Examples

### Monitor Error Rate
```bash
# Check if error rate is acceptable
ERROR_RATE=$(curl -s -H "$AUTH" "$API/analytics/usage?period=hour" | jq -r '.error_rate')
echo "Current error rate: ${ERROR_RATE}%"

if (( $(echo "$ERROR_RATE > 5" | bc -l) )); then
  echo "⚠️  High error rate detected!"
fi
```

### Find Slow Endpoints
```bash
# Get top 5 slowest endpoints
curl -s -H "$AUTH" "$API/analytics/summary" | \
  jq -r '.top_endpoints | sort_by(.average_duration) | reverse | .[0:5] | 
         .[] | "\(.method) \(.path): \(.average_duration / 1000000)ms"'
```

### Track Daily Growth
```bash
# Compare today vs yesterday
TODAY=$(curl -s -H "$AUTH" "$API/analytics/usage?period=24h" | jq '.total_requests')
YESTERDAY=$(curl -s -H "$AUTH" "$API/analytics/usage?period=48h" | jq '.total_requests')
GROWTH=$(echo "scale=2; (($TODAY - ($YESTERDAY - $TODAY)) / ($YESTERDAY - $TODAY)) * 100" | bc)
echo "Daily growth: ${GROWTH}%"
```

## Testing

Run the comprehensive test suite:

```bash
./scripts/test-analytics.sh
```

This validates:
- All endpoints are working
- Data accuracy
- Performance
- Time windows
- Filters

## Next Steps

1. **Explore the Full Documentation**
   - [Complete API Reference](./README.md)
   - [60+ Examples](./examples.md)
   - [Integration Guide](./INTEGRATION.md)
   - [Testing Guide](./testing.md)

2. **Build a Dashboard**
   - Use the `/dashboard` endpoint
   - Create a web UI
   - Set up real-time updates
   - Add visualizations

3. **Set Up Alerts**
   - Monitor error rates
   - Track latency
   - Alert on anomalies
   - Integrate with Slack/email

4. **Export Data**
   - Send to Prometheus
   - Visualize in Grafana
   - Export to CSV
   - Build custom reports

## Troubleshooting

### No Data Showing?
- Wait a few seconds after making requests (async recording)
- Verify analytics is enabled in config
- Check server logs for errors

### Authentication Errors?
- Ensure you're using the correct token
- Check the `Authorization: Bearer <token>` header format
- Verify token matches server configuration

### Need Help?
- Check the [main documentation](./README.md)
- Run `./examples/analytics-demo.sh` for examples
- Review [integration guide](./INTEGRATION.md)

## Resources

- 📖 [Full Documentation](./README.md)
- 💡 [Examples](./examples.md)
- 🧪 [Testing Guide](./testing.md)
- 🔧 [Integration Guide](./INTEGRATION.md)
- 🚀 [Demo Script](../../examples/analytics-demo.sh)

---

**That's it!** You now have comprehensive API analytics running. Start exploring your API usage patterns! 🎉

