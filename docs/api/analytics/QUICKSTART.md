# API Analytics - Quick Start Guide

Get started with API Analytics & Monitoring in under 5 minutes!

## Step 1: Start the API Server

Start the REST API via Docker:

```bash
docker compose -f docker-compose.apis.yml up -d --build
```

The server will start on `http://localhost:8081` with analytics automatically enabled.

## Step 2: Generate Some Traffic

Use any REST client to make a few API calls against `http://localhost:8081/api/v1`.

## Step 3: View Analytics

Now check out your analytics:

### Get a Summary
Use `/analytics/summary` to view overall health and metrics.

This shows:
- Total requests
- Success/error rates
- Average latency
- System health
- Top endpoints

### View the Dashboard
Use `/analytics/dashboard` to view a comprehensive analytics overview.

The dashboard provides a comprehensive overview combining:
- Summary metrics
- Usage statistics
- Popularity data
- Trends

### Check Usage Trends
Use `/analytics/trends` to view time-series trends.

## Step 4: Run the Demo

Use the analytics endpoints listed below to explore the data.

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

Use `/analytics/usage`, `/analytics/summary`, and `/analytics/trends` to monitor error rates, slow endpoints, and growth.

## Testing

Analytics tests run in CI using containerized workflows.

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
- Review [integration guide](./INTEGRATION.md)

## Resources

- [Full Documentation](./README.md)
- [Examples](./examples.md)
- [Testing Guide](./testing.md)
- [Integration Guide](./INTEGRATION.md)

---

**That's it!** You now have comprehensive API analytics running. Start exploring your API usage patterns!

