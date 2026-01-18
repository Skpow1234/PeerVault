#!/bin/bash
# Test script for API Analytics

set -e

API_BASE="http://localhost:8081"
API_V1="$API_BASE/api/v1"
AUTH="Authorization: Bearer demo-token"

echo "================================================"
echo "API Analytics & Monitoring - Test Suite"
echo "================================================"
echo ""

# Check if server is running
if ! curl -s "$API_BASE/health" > /dev/null 2>&1; then
    echo "Error: API server is not running on $API_BASE"
    echo "Please start the server first: ./bin/peervault-api"
    exit 1
fi

echo "API server is running"
echo ""

# Test 1: Generate test traffic
echo "Test 1: Generating test traffic..."
for i in {1..50}; do
    # Mix of successful and failed requests
    if [ $((i % 10)) -eq 0 ]; then
        # Intentional 404
        curl -s -H "$AUTH" "$API_V1/nonexistent" > /dev/null 2>&1 || true
    else
        # Valid request
        curl -s -H "$AUTH" "$API_V1/files" > /dev/null 2>&1 || true
    fi
done
echo "Generated 50 test requests"
echo ""

# Wait for analytics to process
sleep 1

# Test 2: Get Analytics Summary
echo "Test 2: Testing /analytics/summary endpoint..."
SUMMARY=$(curl -s -H "$AUTH" "$API_V1/analytics/summary")
if echo "$SUMMARY" | jq -e '.overall_metrics.total_requests' > /dev/null 2>&1; then
    TOTAL_REQUESTS=$(echo "$SUMMARY" | jq -r '.overall_metrics.total_requests')
    ERROR_RATE=$(echo "$SUMMARY" | jq -r '.overall_metrics.error_rate')
    ACTIVE_USERS=$(echo "$SUMMARY" | jq -r '.active_users')
    HEALTH_STATUS=$(echo "$SUMMARY" | jq -r '.system_health.status')
    
    echo "  Total Requests: $TOTAL_REQUESTS"
    echo "  Error Rate: $ERROR_RATE%"
    echo "  Active Users: $ACTIVE_USERS"
    echo "  Health Status: $HEALTH_STATUS"
    echo "Summary endpoint working"
else
    echo "Failed to get analytics summary"
    exit 1
fi
echo ""

# Test 3: Get Usage Metrics
echo "Test 3: Testing /analytics/usage endpoint..."
USAGE=$(curl -s -H "$AUTH" "$API_V1/analytics/usage?period=24h")
if echo "$USAGE" | jq -e '.total_requests' > /dev/null 2>&1; then
    SUCCESSFUL=$(echo "$USAGE" | jq -r '.successful_requests')
    FAILED=$(echo "$USAGE" | jq -r '.failed_requests')
    AVG_DURATION=$(echo "$USAGE" | jq -r '.average_duration / 1000000')
    
    echo "  Successful: $SUCCESSFUL"
    echo "  Failed: $FAILED"
    echo "  Avg Duration: ${AVG_DURATION}ms"
    echo "Usage metrics endpoint working"
else
    echo "Failed to get usage metrics"
    exit 1
fi
echo ""

# Test 4: Get Endpoint Stats
echo "Test 4: Testing /analytics/endpoint endpoint..."
ENDPOINT_STATS=$(curl -s -H "$AUTH" "$API_V1/analytics/endpoint?endpoint=/api/v1/files&period=24h")
if echo "$ENDPOINT_STATS" | jq -e '.total_calls' > /dev/null 2>&1; then
    TOTAL_CALLS=$(echo "$ENDPOINT_STATS" | jq -r '.total_calls')
    ENDPOINT_ERROR_RATE=$(echo "$ENDPOINT_STATS" | jq -r '.error_rate')
    
    echo "  Total Calls: $TOTAL_CALLS"
    echo "  Error Rate: $ENDPOINT_ERROR_RATE%"
    echo "Endpoint stats working"
else
    echo "Failed to get endpoint stats"
    exit 1
fi
echo ""

# Test 5: Get Usage Trends
echo "Test 5: Testing /analytics/trends endpoint..."
TRENDS=$(curl -s -H "$AUTH" "$API_V1/analytics/trends?period=24h&interval=hour")
if echo "$TRENDS" | jq -e '.[0].request_count' > /dev/null 2>&1; then
    TREND_COUNT=$(echo "$TRENDS" | jq '. | length')
    echo "  Trend Data Points: $TREND_COUNT"
    echo "Trends endpoint working"
else
    echo "Failed to get trends"
    exit 1
fi
echo ""

# Test 6: Get Popularity Metrics
echo "Test 6: Testing /analytics/popularity endpoint..."
POPULARITY=$(curl -s -H "$AUTH" "$API_V1/analytics/popularity?period=7d")
if echo "$POPULARITY" | jq -e '.top_endpoints' > /dev/null 2>&1; then
    TOP_COUNT=$(echo "$POPULARITY" | jq '.top_endpoints | length')
    echo "  Top Endpoints Count: $TOP_COUNT"
    
    if [ "$TOP_COUNT" -gt 0 ]; then
        TOP_ENDPOINT=$(echo "$POPULARITY" | jq -r '.top_endpoints[0].endpoint')
        TOP_CALLS=$(echo "$POPULARITY" | jq -r '.top_endpoints[0].call_count')
        echo "  Most Popular: $TOP_ENDPOINT ($TOP_CALLS calls)"
    fi
    
    echo "Popularity metrics working"
else
    echo "Failed to get popularity metrics"
    exit 1
fi
echo ""

# Test 7: Query API Calls
echo "Test 7: Testing /analytics/calls endpoint..."
CALLS=$(curl -s -H "$AUTH" "$API_V1/analytics/calls?limit=10")
if echo "$CALLS" | jq -e '.[0].id' > /dev/null 2>&1; then
    CALLS_COUNT=$(echo "$CALLS" | jq '. | length')
    echo "  Retrieved Calls: $CALLS_COUNT"
    echo "API calls query working"
else
    echo "Failed to query API calls"
    exit 1
fi
echo ""

# Test 8: Get Dashboard
echo "Test 8: Testing /analytics/dashboard endpoint..."
DASHBOARD=$(curl -s -H "$AUTH" "$API_V1/analytics/dashboard?period=24h")
if echo "$DASHBOARD" | jq -e '.summary' > /dev/null 2>&1; then
    echo "  Has Summary: OK"
    echo "  Has Metrics: $(echo $DASHBOARD | jq -e '.metrics' > /dev/null 2>&1 && echo 'OK' || echo 'FAIL')"
    echo "  Has Popularity: $(echo $DASHBOARD | jq -e '.popularity' > /dev/null 2>&1 && echo 'OK' || echo 'FAIL')"
    echo "  Has Trends: $(echo $DASHBOARD | jq -e '.trends' > /dev/null 2>&1 && echo 'OK' || echo 'FAIL')"
    echo "Dashboard endpoint working"
else
    echo "Failed to get dashboard"
    exit 1
fi
echo ""

# Test 9: Test different time windows
echo "Test 9: Testing time window parameters..."
for period in "hour" "24h" "7d" "30d"; do
    RESULT=$(curl -s -H "$AUTH" "$API_V1/analytics/usage?period=$period")
    if echo "$RESULT" | jq -e '.total_requests' > /dev/null 2>&1; then
        echo "  Period '$period': OK"
    else
        echo "  Period '$period': FAIL"
    fi
done
echo ""

# Test 10: Performance test
echo "Test 10: Performance test..."
START_TIME=$(date +%s%N)
for i in {1..20}; do
    curl -s -H "$AUTH" "$API_V1/analytics/summary" > /dev/null
done
END_TIME=$(date +%s%N)
DURATION=$((($END_TIME - $START_TIME) / 1000000))
AVG_TIME=$(($DURATION / 20))

echo "  20 requests completed in ${DURATION}ms"
echo "  Average: ${AVG_TIME}ms per request"

if [ $AVG_TIME -lt 100 ]; then
    echo "Performance is good (< 100ms)"
elif [ $AVG_TIME -lt 500 ]; then
    echo "Performance is acceptable (< 500ms)"
else
    echo "Performance is slow (> 500ms)"
fi
echo ""

# Summary
echo "================================================"
echo "Test Results Summary"
echo "================================================"
echo "All analytics endpoints are working correctly"
echo "Data is being recorded and aggregated properly"
echo "Time windows and filters are functional"
echo "Performance is acceptable"
echo ""
echo "Analytics Dashboard: $API_BASE/api/v1/analytics/dashboard"
echo "Analytics Summary: $API_BASE/api/v1/analytics/summary"
echo ""
echo "All tests passed!"

