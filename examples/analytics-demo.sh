#!/bin/bash
# API Analytics Demo Script
# This script demonstrates the API Analytics & Monitoring features

set -e

API_BASE="http://localhost:8081"
API_V1="$API_BASE/api/v1"
AUTH="Authorization: Bearer demo-token"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================================"
echo "API Analytics & Monitoring - Demo"
echo "================================================"
echo ""

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is not installed. Please install jq to run this demo."
    echo "  macOS: brew install jq"
    echo "  Ubuntu/Debian: sudo apt-get install jq"
    exit 1
fi

# Check if server is running
echo -e "${BLUE}Checking if API server is running...${NC}"
if ! curl -s "$API_BASE/health" > /dev/null 2>&1; then
    echo "Error: API server is not running on $API_BASE"
    echo "Please start the server first: ./bin/peervault-api"
    exit 1
fi
echo -e "${GREEN}Server is running${NC}"
echo ""

# Demo 1: Generate sample traffic
echo -e "${BLUE}Demo 1: Generating sample API traffic...${NC}"
echo "Creating 30 API requests with mixed success/failure..."
for i in {1..30}; do
    # 80% success, 20% errors
    if [ $((i % 5)) -eq 0 ]; then
        curl -s -H "$AUTH" "$API_V1/nonexistent" > /dev/null 2>&1 || true
    else
        curl -s -H "$AUTH" "$API_V1/files" > /dev/null 2>&1 || true
    fi
    printf "."
done
echo ""
echo -e "${GREEN}Sample traffic generated${NC}"
sleep 1
echo ""

# Demo 2: View Analytics Summary
echo -e "${BLUE}Demo 2: Viewing Analytics Summary${NC}"
SUMMARY=$(curl -s -H "$AUTH" "$API_V1/analytics/summary")
echo "Current API Health:"
echo "-------------------"
echo "$SUMMARY" | jq -r '
  "Total Requests: \(.overall_metrics.total_requests)",
  "Successful: \(.overall_metrics.successful_requests)",
  "Failed: \(.overall_metrics.failed_requests)",
  "Error Rate: \(.overall_metrics.error_rate)%",
  "Avg Latency: \(.overall_metrics.average_duration / 1000000)ms",
  "Active Users: \(.active_users)",
  "System Status: \(.system_health.status)"
'
echo ""

# Demo 3: Top Endpoints
echo -e "${BLUE}Demo 3: Most Popular Endpoints${NC}"
echo "Top 5 endpoints by call count:"
echo "-------------------------------"
curl -s -H "$AUTH" "$API_V1/analytics/summary" | \
  jq -r '.top_endpoints[:5] | .[] | 
    "\(.method) \(.path) - \(.total_calls) calls (Error Rate: \(.error_rate)%)"'
echo ""

# Demo 4: Usage Trends
echo -e "${BLUE}Demo 4: Hourly Usage Trends (Last 6 Hours)${NC}"
echo "Recent traffic patterns:"
echo "------------------------"
curl -s -H "$AUTH" "$API_V1/analytics/trends?period=6h&interval=hour" | \
  jq -r '.[-6:] | .[] | 
    "\(.timestamp | split("T")[1] | split(".")[0]): \(.request_count) requests, \(.error_count) errors"'
echo ""

# Demo 5: Performance Analysis
echo -e "${BLUE}Demo 5: Performance Metrics${NC}"
METRICS=$(curl -s -H "$AUTH" "$API_V1/analytics/usage?period=24h")
echo "Last 24 Hours Performance:"
echo "--------------------------"
echo "$METRICS" | jq -r '
  "Requests per Method:",
  (.requests_by_method | to_entries[] | "  \(.key): \(.value)"),
  "",
  "Data Transfer:",
  "  In: \(.total_data_in / 1048576 | floor) MB",
  "  Out: \(.total_data_out / 1048576 | floor) MB",
  "",
  "Average Response Time: \(.average_duration / 1000000) ms"
'
echo ""

# Demo 6: Error Analysis
echo -e "${BLUE}Demo 6: Error Analysis${NC}"
ERROR_RATE=$(curl -s -H "$AUTH" "$API_V1/analytics/usage?period=24h" | jq -r '.error_rate')
echo "Current error rate: ${ERROR_RATE}%"

if (( $(echo "$ERROR_RATE > 5" | bc -l) )); then
    echo -e "${YELLOW}Warning: Error rate is above 5%${NC}"
    echo "Recent errors:"
    curl -s -H "$AUTH" "$API_V1/analytics/calls?limit=5" | \
      jq -r '.[] | select(.status_code >= 400) | 
        "  \(.timestamp | split("T")[1] | split(".")[0]) - \(.method) \(.path) - Status: \(.status_code)"'
else
    echo -e "${GREEN}Error rate is acceptable (< 5%)${NC}"
fi
echo ""

# Demo 7: Popularity Metrics
echo -e "${BLUE}Demo 7: API Popularity & Trends${NC}"
POPULARITY=$(curl -s -H "$AUTH" "$API_V1/analytics/popularity?period=7d")

echo "Trending Up:"
echo "$POPULARITY" | jq -r '.trending_up[:3] | .[] | 
  "  \(.endpoint) - Growth: +\(.growth_rate | floor)%"' 2>/dev/null || echo "  (No trending data yet)"
echo ""

echo "Most Active Users (Top 3):"
echo "$POPULARITY" | jq -r '.most_active_users[:3] | .[] | 
  "  User: \(.user_id) - \(.request_count) requests"' 2>/dev/null || echo "  (No user data available)"
echo ""

# Demo 8: Real-time Dashboard
echo -e "${BLUE}Demo 8: Dashboard View${NC}"
echo "Comprehensive dashboard data:"
echo "-----------------------------"
DASHBOARD=$(curl -s -H "$AUTH" "$API_V1/analytics/dashboard?period=24h")
echo "$DASHBOARD" | jq '{
  system_health: .summary.system_health.status,
  total_requests: .metrics.total_requests,
  error_rate: .metrics.error_rate,
  active_users: .summary.active_users,
  top_endpoint: .metrics.top_endpoints[0].path,
  trend_direction: (
    if (.trends | length) > 1 then
      if (.trends[-1].request_count > .trends[0].request_count) then "↑ Increasing"
      elif (.trends[-1].request_count < .trends[0].request_count) then "↓ Decreasing"
      else "→ Stable"
      end
    else "N/A"
    end
  )
}'
echo ""

# Demo 9: Export to CSV
echo -e "${BLUE}Demo 9: Exporting Data${NC}"
echo "Exporting trends to CSV..."
EXPORT_FILE="/tmp/analytics-export-$(date +%Y%m%d-%H%M%S).csv"
curl -s -H "$AUTH" "$API_V1/analytics/trends?period=24h&interval=hour" | \
  jq -r '["Timestamp","Requests","Success","Errors","Avg Latency (ms)"],
         (.[] | [.timestamp, .request_count, .success_count, .error_count, .average_duration_ms]) | 
         @csv' > "$EXPORT_FILE"
echo -e "${GREEN}Data exported to: $EXPORT_FILE${NC}"
echo ""

# Demo 10: Health Check Alert Simulation
echo -e "${BLUE}Demo 10: Alert Simulation${NC}"
HEALTH_STATUS=$(curl -s -H "$AUTH" "$API_V1/analytics/summary" | jq -r '.system_health.status')
echo "System Health: $HEALTH_STATUS"

if [ "$HEALTH_STATUS" = "healthy" ]; then
    echo -e "${GREEN}All systems operational${NC}"
elif [ "$HEALTH_STATUS" = "degraded" ]; then
    echo -e "${YELLOW}System degraded - review warnings${NC}"
else
    echo -e "${YELLOW}System unhealthy - immediate attention required${NC}"
fi
echo ""

# Summary
echo "================================================"
echo "Demo Complete!"
echo "================================================"
echo ""
echo "Available Analytics Endpoints:"
echo "  • Summary:     $API_V1/analytics/summary"
echo "  • Dashboard:   $API_V1/analytics/dashboard"
echo "  • Usage:       $API_V1/analytics/usage"
echo "  • Trends:      $API_V1/analytics/trends"
echo "  • Popularity:  $API_V1/analytics/popularity"
echo ""
echo "Documentation:"
echo "  • Full API Docs: docs/api/analytics/README.md"
echo "  • Examples:      docs/api/analytics/examples.md"
echo "  • Integration:   docs/api/analytics/INTEGRATION.md"
echo ""
echo -e "${GREEN}Try building your own analytics dashboard!${NC}"

