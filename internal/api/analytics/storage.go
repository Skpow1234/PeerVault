package analytics

import (
	"sync"
	"time"
)

// Storage interface for analytics data persistence
type Storage interface {
	// RecordAPICall stores a single API call record
	RecordAPICall(call *APICall) error
	
	// GetAPICalls retrieves API calls based on query parameters
	GetAPICalls(query *AnalyticsQuery) ([]*APICall, error)
	
	// GetUsageMetrics calculates aggregated usage metrics
	GetUsageMetrics(startTime, endTime time.Time) (*UsageMetrics, error)
	
	// GetEndpointStats retrieves statistics for a specific endpoint
	GetEndpointStats(endpoint string, startTime, endTime time.Time) (*EndpointStats, error)
	
	// GetUserBehavior retrieves behavior metrics for a specific user
	GetUserBehavior(userID string, startTime, endTime time.Time) (*UserBehaviorMetrics, error)
	
	// GetUsageTrends retrieves usage trends over time
	GetUsageTrends(startTime, endTime time.Time, interval string) ([]*UsageTrend, error)
	
	// GetPopularityMetrics retrieves popularity metrics
	GetPopularityMetrics(startTime, endTime time.Time) (*PopularityMetrics, error)
	
	// GetAnalyticsSummary retrieves a comprehensive analytics summary
	GetAnalyticsSummary() (*AnalyticsSummary, error)
	
	// CleanupOldData removes old analytics data
	CleanupOldData(olderThan time.Time) error
}

// MemoryStorage implements in-memory storage for analytics data
type MemoryStorage struct {
	calls       []*APICall
	mu          sync.RWMutex
	maxEntries  int
	retentionDays int
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage(maxEntries, retentionDays int) *MemoryStorage {
	return &MemoryStorage{
		calls:         make([]*APICall, 0, maxEntries),
		maxEntries:    maxEntries,
		retentionDays: retentionDays,
	}
}

// RecordAPICall stores a single API call record
func (s *MemoryStorage) RecordAPICall(call *APICall) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// If we've reached max entries, remove the oldest
	if len(s.calls) >= s.maxEntries {
		s.calls = s.calls[1:]
	}
	
	s.calls = append(s.calls, call)
	return nil
}

// GetAPICalls retrieves API calls based on query parameters
func (s *MemoryStorage) GetAPICalls(query *AnalyticsQuery) ([]*APICall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var filtered []*APICall
	
	for _, call := range s.calls {
		// Apply filters
		if !query.StartTime.IsZero() && call.Timestamp.Before(query.StartTime) {
			continue
		}
		if !query.EndTime.IsZero() && call.Timestamp.After(query.EndTime) {
			continue
		}
		if query.Endpoint != "" && call.Endpoint != query.Endpoint {
			continue
		}
		if query.Method != "" && call.Method != query.Method {
			continue
		}
		if query.UserID != "" && call.UserID != query.UserID {
			continue
		}
		if query.StatusCode != 0 && call.StatusCode != query.StatusCode {
			continue
		}
		if query.MinDuration > 0 && call.Duration < query.MinDuration {
			continue
		}
		if query.MaxDuration > 0 && call.Duration > query.MaxDuration {
			continue
		}
		
		filtered = append(filtered, call)
	}
	
	// Apply limit and offset
	if query.Offset > 0 && query.Offset < len(filtered) {
		filtered = filtered[query.Offset:]
	}
	if query.Limit > 0 && query.Limit < len(filtered) {
		filtered = filtered[:query.Limit]
	}
	
	return filtered, nil
}

// GetUsageMetrics calculates aggregated usage metrics
func (s *MemoryStorage) GetUsageMetrics(startTime, endTime time.Time) (*UsageMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics := &UsageMetrics{
		RequestsByMethod: make(map[string]int64),
		RequestsByPath:   make(map[string]int64),
		RequestsByStatus: make(map[int]int64),
		TimeWindow: TimeWindow{
			Start: startTime,
			End:   endTime,
		},
	}
	
	var totalDuration time.Duration
	endpointStatsMap := make(map[string]*EndpointStats)
	
	for _, call := range s.calls {
		if call.Timestamp.Before(startTime) || call.Timestamp.After(endTime) {
			continue
		}
		
		metrics.TotalRequests++
		if call.StatusCode >= 200 && call.StatusCode < 400 {
			metrics.SuccessfulRequests++
		} else {
			metrics.FailedRequests++
		}
		
		totalDuration += call.Duration
		metrics.TotalDataIn += call.RequestSize
		metrics.TotalDataOut += call.ResponseSize
		
		metrics.RequestsByMethod[call.Method]++
		metrics.RequestsByPath[call.Path]++
		metrics.RequestsByStatus[call.StatusCode]++
		
		// Track endpoint statistics
		key := call.Method + " " + call.Endpoint
		if stats, exists := endpointStatsMap[key]; exists {
			stats.TotalCalls++
			if call.StatusCode >= 200 && call.StatusCode < 400 {
				stats.SuccessfulCalls++
			} else {
				stats.FailedCalls++
			}
			stats.TotalDataIn += call.RequestSize
			stats.TotalDataOut += call.ResponseSize
			
			// Update min/max duration
			if call.Duration < stats.MinDuration || stats.MinDuration == 0 {
				stats.MinDuration = call.Duration
			}
			if call.Duration > stats.MaxDuration {
				stats.MaxDuration = call.Duration
			}
			
			if call.Timestamp.After(stats.LastCalled) {
				stats.LastCalled = call.Timestamp
			}
		} else {
			endpointStatsMap[key] = &EndpointStats{
				Path:            call.Path,
				Method:          call.Method,
				TotalCalls:      1,
				SuccessfulCalls: 0,
				FailedCalls:     0,
				MinDuration:     call.Duration,
				MaxDuration:     call.Duration,
				TotalDataIn:     call.RequestSize,
				TotalDataOut:    call.ResponseSize,
				LastCalled:      call.Timestamp,
			}
			if call.StatusCode >= 200 && call.StatusCode < 400 {
				endpointStatsMap[key].SuccessfulCalls++
			} else {
				endpointStatsMap[key].FailedCalls++
			}
		}
	}
	
	if metrics.TotalRequests > 0 {
		metrics.AverageDuration = totalDuration / time.Duration(metrics.TotalRequests)
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
		
		// Calculate average duration and error rate for each endpoint
		for _, stats := range endpointStatsMap {
			stats.AverageDuration = time.Duration(int64(totalDuration) / int64(stats.TotalCalls))
			if stats.TotalCalls > 0 {
				stats.ErrorRate = float64(stats.FailedCalls) / float64(stats.TotalCalls) * 100
			}
		}
	}
	
	// Convert endpoint stats map to slice and get top endpoints
	for _, stats := range endpointStatsMap {
		metrics.TopEndpoints = append(metrics.TopEndpoints, *stats)
	}
	
	// Sort by total calls (simple bubble sort for top 10)
	for i := 0; i < len(metrics.TopEndpoints) && i < 10; i++ {
		for j := i + 1; j < len(metrics.TopEndpoints); j++ {
			if metrics.TopEndpoints[j].TotalCalls > metrics.TopEndpoints[i].TotalCalls {
				metrics.TopEndpoints[i], metrics.TopEndpoints[j] = metrics.TopEndpoints[j], metrics.TopEndpoints[i]
			}
		}
	}
	
	// Limit to top 10
	if len(metrics.TopEndpoints) > 10 {
		metrics.TopEndpoints = metrics.TopEndpoints[:10]
	}
	
	return metrics, nil
}

// GetEndpointStats retrieves statistics for a specific endpoint
func (s *MemoryStorage) GetEndpointStats(endpoint string, startTime, endTime time.Time) (*EndpointStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	stats := &EndpointStats{
		Path: endpoint,
	}
	
	var totalDuration time.Duration
	
	for _, call := range s.calls {
		if call.Endpoint != endpoint {
			continue
		}
		if call.Timestamp.Before(startTime) || call.Timestamp.After(endTime) {
			continue
		}
		
		stats.TotalCalls++
		if call.StatusCode >= 200 && call.StatusCode < 400 {
			stats.SuccessfulCalls++
		} else {
			stats.FailedCalls++
		}
		
		totalDuration += call.Duration
		stats.TotalDataIn += call.RequestSize
		stats.TotalDataOut += call.ResponseSize
		
		if stats.MinDuration == 0 || call.Duration < stats.MinDuration {
			stats.MinDuration = call.Duration
		}
		if call.Duration > stats.MaxDuration {
			stats.MaxDuration = call.Duration
		}
		
		if call.Timestamp.After(stats.LastCalled) {
			stats.LastCalled = call.Timestamp
		}
	}
	
	if stats.TotalCalls > 0 {
		stats.AverageDuration = totalDuration / time.Duration(stats.TotalCalls)
		stats.ErrorRate = float64(stats.FailedCalls) / float64(stats.TotalCalls) * 100
	}
	
	return stats, nil
}

// GetUserBehavior retrieves behavior metrics for a specific user
func (s *MemoryStorage) GetUserBehavior(userID string, startTime, endTime time.Time) (*UserBehaviorMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics := &UserBehaviorMetrics{
		UserID:      userID,
		UserAgents:  make(map[string]int64),
		IPAddresses: make(map[string]int64),
	}
	
	uniqueEndpoints := make(map[string]bool)
	hourlyRequests := make(map[int]int64)
	endpointStats := make(map[string]*EndpointStats)
	
	for _, call := range s.calls {
		if call.UserID != userID {
			continue
		}
		if call.Timestamp.Before(startTime) || call.Timestamp.After(endTime) {
			continue
		}
		
		metrics.TotalRequests++
		uniqueEndpoints[call.Endpoint] = true
		
		hour := call.Timestamp.Hour()
		hourlyRequests[hour]++
		
		if call.UserAgent != "" {
			metrics.UserAgents[call.UserAgent]++
		}
		if call.IPAddress != "" {
			metrics.IPAddresses[call.IPAddress]++
		}
		
		if metrics.FirstSeen.IsZero() || call.Timestamp.Before(metrics.FirstSeen) {
			metrics.FirstSeen = call.Timestamp
		}
		if call.Timestamp.After(metrics.LastSeen) {
			metrics.LastSeen = call.Timestamp
		}
		
		if call.StatusCode >= 400 {
			metrics.ErrorCount++
		}
		
		// Track endpoint usage
		key := call.Method + " " + call.Endpoint
		if stats, exists := endpointStats[key]; exists {
			stats.TotalCalls++
		} else {
			endpointStats[key] = &EndpointStats{
				Path:       call.Path,
				Method:     call.Method,
				TotalCalls: 1,
			}
		}
	}
	
	metrics.UniqueEndpoints = len(uniqueEndpoints)
	
	// Calculate average request rate
	if !metrics.FirstSeen.IsZero() && !metrics.LastSeen.IsZero() {
		hours := metrics.LastSeen.Sub(metrics.FirstSeen).Hours()
		if hours > 0 {
			metrics.AverageRequestRate = float64(metrics.TotalRequests) / hours
		}
	}
	
	// Find peak hour
	var maxRequests int64
	for hour, count := range hourlyRequests {
		if count > maxRequests {
			maxRequests = count
			metrics.PeakHour = hour
		}
	}
	
	// Get favorite endpoints (top 5)
	for _, stats := range endpointStats {
		metrics.FavoriteEndpoints = append(metrics.FavoriteEndpoints, *stats)
	}
	
	// Simple sort for top 5
	for i := 0; i < len(metrics.FavoriteEndpoints) && i < 5; i++ {
		for j := i + 1; j < len(metrics.FavoriteEndpoints); j++ {
			if metrics.FavoriteEndpoints[j].TotalCalls > metrics.FavoriteEndpoints[i].TotalCalls {
				metrics.FavoriteEndpoints[i], metrics.FavoriteEndpoints[j] = metrics.FavoriteEndpoints[j], metrics.FavoriteEndpoints[i]
			}
		}
	}
	
	if len(metrics.FavoriteEndpoints) > 5 {
		metrics.FavoriteEndpoints = metrics.FavoriteEndpoints[:5]
	}
	
	return metrics, nil
}

// GetUsageTrends retrieves usage trends over time
func (s *MemoryStorage) GetUsageTrends(startTime, endTime time.Time, interval string) ([]*UsageTrend, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Determine interval duration
	var intervalDuration time.Duration
	switch interval {
	case "hour":
		intervalDuration = time.Hour
	case "day":
		intervalDuration = 24 * time.Hour
	case "week":
		intervalDuration = 7 * 24 * time.Hour
	default:
		intervalDuration = time.Hour
	}
	
	trendsMap := make(map[time.Time]*UsageTrend)
	
	for _, call := range s.calls {
		if call.Timestamp.Before(startTime) || call.Timestamp.After(endTime) {
			continue
		}
		
		// Round timestamp to interval
		timestamp := call.Timestamp.Truncate(intervalDuration)
		
		if trend, exists := trendsMap[timestamp]; exists {
			trend.RequestCount++
			if call.StatusCode >= 200 && call.StatusCode < 400 {
				trend.SuccessCount++
			} else {
				trend.ErrorCount++
			}
			trend.DataIn += call.RequestSize
			trend.DataOut += call.ResponseSize
			trend.AverageDuration = (trend.AverageDuration*float64(trend.RequestCount-1) + float64(call.Duration.Milliseconds())) / float64(trend.RequestCount)
		} else {
			trend := &UsageTrend{
				Timestamp:       timestamp,
				RequestCount:    1,
				AverageDuration: float64(call.Duration.Milliseconds()),
				DataIn:          call.RequestSize,
				DataOut:         call.ResponseSize,
			}
			if call.StatusCode >= 200 && call.StatusCode < 400 {
				trend.SuccessCount = 1
			} else {
				trend.ErrorCount = 1
			}
			trendsMap[timestamp] = trend
		}
	}
	
	// Convert map to slice and sort by timestamp
	var trends []*UsageTrend
	for _, trend := range trendsMap {
		trends = append(trends, trend)
	}
	
	// Simple bubble sort by timestamp
	for i := 0; i < len(trends); i++ {
		for j := i + 1; j < len(trends); j++ {
			if trends[j].Timestamp.Before(trends[i].Timestamp) {
				trends[i], trends[j] = trends[j], trends[i]
			}
		}
	}
	
	return trends, nil
}

// GetPopularityMetrics retrieves popularity metrics
func (s *MemoryStorage) GetPopularityMetrics(startTime, endTime time.Time) (*PopularityMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics := &PopularityMetrics{
		TimeWindow: TimeWindow{
			Start: startTime,
			End:   endTime,
		},
	}
	
	endpointPopularityMap := make(map[string]*EndpointPopularity)
	userActivityMap := make(map[string]*UserActivity)
	
	// Calculate midpoint for trend analysis
	midpoint := startTime.Add(endTime.Sub(startTime) / 2)
	
	for _, call := range s.calls {
		if call.Timestamp.Before(startTime) || call.Timestamp.After(endTime) {
			continue
		}
		
		// Track endpoint popularity
		key := call.Method + " " + call.Endpoint
		if pop, exists := endpointPopularityMap[key]; exists {
			pop.CallCount++
			if call.Timestamp.After(midpoint) {
				pop.LastDayCount++
			} else {
				pop.PreviousDayCount++
			}
		} else {
			pop := &EndpointPopularity{
				Endpoint: key,
				CallCount: 1,
			}
			if call.Timestamp.After(midpoint) {
				pop.LastDayCount = 1
			} else {
				pop.PreviousDayCount = 1
			}
			endpointPopularityMap[key] = pop
		}
		
		// Track user activity
		if call.UserID != "" {
			if activity, exists := userActivityMap[call.UserID]; exists {
				activity.RequestCount++
				if call.Timestamp.After(activity.LastActivity) {
					activity.LastActivity = call.Timestamp
				}
			} else {
				userActivityMap[call.UserID] = &UserActivity{
					UserID:       call.UserID,
					RequestCount: 1,
					LastActivity: call.Timestamp,
				}
			}
		}
	}
	
	// Calculate growth rate and trend score for endpoints
	for _, pop := range endpointPopularityMap {
		if pop.PreviousDayCount > 0 {
			pop.GrowthRate = (float64(pop.LastDayCount) - float64(pop.PreviousDayCount)) / float64(pop.PreviousDayCount) * 100
		} else if pop.LastDayCount > 0 {
			pop.GrowthRate = 100 // New endpoint
		}
		pop.TrendScore = float64(pop.LastDayCount) * (1 + pop.GrowthRate/100)
	}
	
	// Get top endpoints
	for _, pop := range endpointPopularityMap {
		metrics.TopEndpoints = append(metrics.TopEndpoints, *pop)
	}
	
	// Sort by call count
	for i := 0; i < len(metrics.TopEndpoints) && i < 10; i++ {
		for j := i + 1; j < len(metrics.TopEndpoints); j++ {
			if metrics.TopEndpoints[j].CallCount > metrics.TopEndpoints[i].CallCount {
				metrics.TopEndpoints[i], metrics.TopEndpoints[j] = metrics.TopEndpoints[j], metrics.TopEndpoints[i]
			}
		}
	}
	if len(metrics.TopEndpoints) > 10 {
		metrics.TopEndpoints = metrics.TopEndpoints[:10]
	}
	
	// Get trending up (positive growth rate)
	for _, pop := range endpointPopularityMap {
		if pop.GrowthRate > 0 {
			metrics.TrendingUp = append(metrics.TrendingUp, *pop)
		}
	}
	
	// Sort trending up by growth rate
	for i := 0; i < len(metrics.TrendingUp) && i < 10; i++ {
		for j := i + 1; j < len(metrics.TrendingUp); j++ {
			if metrics.TrendingUp[j].GrowthRate > metrics.TrendingUp[i].GrowthRate {
				metrics.TrendingUp[i], metrics.TrendingUp[j] = metrics.TrendingUp[j], metrics.TrendingUp[i]
			}
		}
	}
	if len(metrics.TrendingUp) > 10 {
		metrics.TrendingUp = metrics.TrendingUp[:10]
	}
	
	// Get trending down (negative growth rate)
	for _, pop := range endpointPopularityMap {
		if pop.GrowthRate < 0 {
			metrics.TrendingDown = append(metrics.TrendingDown, *pop)
		}
	}
	
	// Sort trending down by growth rate (most negative first)
	for i := 0; i < len(metrics.TrendingDown) && i < 10; i++ {
		for j := i + 1; j < len(metrics.TrendingDown); j++ {
			if metrics.TrendingDown[j].GrowthRate < metrics.TrendingDown[i].GrowthRate {
				metrics.TrendingDown[i], metrics.TrendingDown[j] = metrics.TrendingDown[j], metrics.TrendingDown[i]
			}
		}
	}
	if len(metrics.TrendingDown) > 10 {
		metrics.TrendingDown = metrics.TrendingDown[:10]
	}
	
	// Get most active users
	for _, activity := range userActivityMap {
		metrics.MostActiveUsers = append(metrics.MostActiveUsers, *activity)
	}
	
	// Sort by request count
	for i := 0; i < len(metrics.MostActiveUsers) && i < 10; i++ {
		for j := i + 1; j < len(metrics.MostActiveUsers); j++ {
			if metrics.MostActiveUsers[j].RequestCount > metrics.MostActiveUsers[i].RequestCount {
				metrics.MostActiveUsers[i], metrics.MostActiveUsers[j] = metrics.MostActiveUsers[j], metrics.MostActiveUsers[i]
			}
		}
	}
	if len(metrics.MostActiveUsers) > 10 {
		metrics.MostActiveUsers = metrics.MostActiveUsers[:10]
	}
	
	return metrics, nil
}

// GetAnalyticsSummary retrieves a comprehensive analytics summary
func (s *MemoryStorage) GetAnalyticsSummary() (*AnalyticsSummary, error) {
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	last7days := now.Add(-7 * 24 * time.Hour)
	
	// Get overall metrics for last 24 hours
	overallMetrics, err := s.GetUsageMetrics(last24h, now)
	if err != nil {
		return nil, err
	}
	
	// Get recent trends (hourly for last 24 hours)
	trends, err := s.GetUsageTrends(last24h, now, "hour")
	if err != nil {
		return nil, err
	}
	
	// Count unique users
	s.mu.RLock()
	uniqueUsers := make(map[string]bool)
	newUsersToday := make(map[string]time.Time)
	
	for _, call := range s.calls {
		if call.UserID != "" {
			if call.Timestamp.After(last24h) {
				uniqueUsers[call.UserID] = true
			}
			if call.Timestamp.After(last7days) {
				if firstSeen, exists := newUsersToday[call.UserID]; !exists || call.Timestamp.Before(firstSeen) {
					newUsersToday[call.UserID] = call.Timestamp
				}
			}
		}
	}
	s.mu.RUnlock()
	
	// Count new users today
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	newToday := 0
	for _, firstSeen := range newUsersToday {
		if firstSeen.After(todayStart) {
			newToday++
		}
	}
	
	// Determine system health
	health := SystemHealth{
		Status:         "healthy",
		ErrorRate:      overallMetrics.ErrorRate,
		AverageLatency: float64(overallMetrics.AverageDuration.Milliseconds()),
	}
	
	if overallMetrics.TotalRequests > 0 {
		health.RequestRate = float64(overallMetrics.TotalRequests) / 24.0 / 3600.0 // requests per second
	}
	
	if health.ErrorRate > 10 {
		health.Status = "degraded"
		health.Warnings = append(health.Warnings, "High error rate detected")
	}
	if health.ErrorRate > 25 {
		health.Status = "unhealthy"
	}
	if health.AverageLatency > 1000 {
		health.Status = "degraded"
		health.Warnings = append(health.Warnings, "High latency detected")
	}
	
	summary := &AnalyticsSummary{
		OverallMetrics: *overallMetrics,
		TopEndpoints:   overallMetrics.TopEndpoints,
		RecentTrends:   trends,
		ActiveUsers:    len(uniqueUsers),
		NewUsersToday:  newToday,
		SystemHealth:   health,
		GeneratedAt:    now,
	}
	
	return summary, nil
}

// CleanupOldData removes old analytics data
func (s *MemoryStorage) CleanupOldData(olderThan time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var filtered []*APICall
	for _, call := range s.calls {
		if call.Timestamp.After(olderThan) {
			filtered = append(filtered, call)
		}
	}
	
	s.calls = filtered
	return nil
}

