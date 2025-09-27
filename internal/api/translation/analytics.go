package translation

import (
	"sync"
	"time"
)

// Analytics tracks translation analytics and metrics
type Analytics struct {
	mu sync.RWMutex

	// Translation metrics
	translations      map[string]*TranslationMetrics
	totalTranslations int
	totalErrors       int

	// Protocol metrics
	protocolStats map[string]*ProtocolMetrics

	// Time-based data
	hourlyData map[string]*HourlyData
	dailyData  map[string]*DailyData

	// Performance metrics
	avgTranslationTime time.Duration
	maxTranslationTime time.Duration
	minTranslationTime time.Duration

	// Error tracking
	errorCounts map[string]int
	lastError   time.Time
}

// TranslationMetrics tracks metrics for a specific translation engine
type TranslationMetrics struct {
	EngineName             string
	FromProtocol           string
	ToProtocol             string
	TotalRequests          int
	SuccessfulTranslations int
	FailedTranslations     int
	TotalBytesIn           int64
	TotalBytesOut          int64
	AvgLatency             time.Duration
	MaxLatency             time.Duration
	MinLatency             time.Duration
	LastActivity           time.Time
	CreatedAt              time.Time
}

// ProtocolMetrics tracks metrics for a specific protocol
type ProtocolMetrics struct {
	Protocol      string
	TotalRequests int
	TotalBytes    int64
	AvgLatency    time.Duration
	ErrorRate     float64
	LastActivity  time.Time
}

// HourlyData tracks hourly statistics
type HourlyData struct {
	Hour            int
	Translations    int
	Errors          int
	BytesTranslated int64
	AvgLatency      time.Duration
}

// DailyData tracks daily statistics
type DailyData struct {
	Date            time.Time
	Translations    int
	Errors          int
	BytesTranslated int64
	AvgLatency      time.Duration
	TopEngines      []string
}

// AnalyticsData represents the complete analytics data
type AnalyticsData struct {
	Summary            *AnalyticsSummary              `json:"summary"`
	TranslationEngines map[string]*TranslationMetrics `json:"translation_engines"`
	ProtocolStats      map[string]*ProtocolMetrics    `json:"protocol_stats"`
	HourlyData         map[string]*HourlyData         `json:"hourly_data"`
	DailyData          map[string]*DailyData          `json:"daily_data"`
	Performance        *PerformanceMetrics            `json:"performance"`
	ErrorAnalysis      *ErrorAnalysis                 `json:"error_analysis"`
	GeneratedAt        time.Time                      `json:"generated_at"`
}

// AnalyticsSummary provides a high-level overview
type AnalyticsSummary struct {
	TotalTranslations    int     `json:"total_translations"`
	TotalErrors          int     `json:"total_errors"`
	SuccessRate          float64 `json:"success_rate"`
	TotalBytesTranslated int64   `json:"total_bytes_translated"`
	ActiveEngines        int     `json:"active_engines"`
	ActiveProtocols      int     `json:"active_protocols"`
	Uptime               string  `json:"uptime"`
}

// PerformanceMetrics tracks performance statistics
type PerformanceMetrics struct {
	AvgTranslationTime  time.Duration `json:"avg_translation_time"`
	MaxTranslationTime  time.Duration `json:"max_translation_time"`
	MinTranslationTime  time.Duration `json:"min_translation_time"`
	P95Latency          time.Duration `json:"p95_latency"`
	P99Latency          time.Duration `json:"p99_latency"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
}

// ErrorAnalysis provides error analysis
type ErrorAnalysis struct {
	TotalErrors     int            `json:"total_errors"`
	ErrorRate       float64        `json:"error_rate"`
	ErrorTypes      map[string]int `json:"error_types"`
	MostCommonError string         `json:"most_common_error"`
	LastErrorTime   time.Time      `json:"last_error_time"`
	ErrorTrend      []ErrorTrend   `json:"error_trend"`
}

// ErrorTrend tracks error trends over time
type ErrorTrend struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
	Rate      float64   `json:"rate"`
}

// NewAnalytics creates a new analytics instance
func NewAnalytics() *Analytics {
	return &Analytics{
		translations:       make(map[string]*TranslationMetrics),
		protocolStats:      make(map[string]*ProtocolMetrics),
		hourlyData:         make(map[string]*HourlyData),
		dailyData:          make(map[string]*DailyData),
		errorCounts:        make(map[string]int),
		minTranslationTime: time.Hour, // Initialize with a large value
	}
}

// RecordTranslation records a translation event
func (a *Analytics) RecordTranslation(engineName, fromProtocol, toProtocol string, success bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()

	// Update translation metrics
	if metrics, exists := a.translations[engineName]; exists {
		metrics.TotalRequests++
		if success {
			metrics.SuccessfulTranslations++
		} else {
			metrics.FailedTranslations++
		}
		metrics.LastActivity = now
	} else {
		a.translations[engineName] = &TranslationMetrics{
			EngineName:             engineName,
			FromProtocol:           fromProtocol,
			ToProtocol:             toProtocol,
			TotalRequests:          1,
			SuccessfulTranslations: 1,
			CreatedAt:              now,
			LastActivity:           now,
		}
		if !success {
			a.translations[engineName].SuccessfulTranslations = 0
			a.translations[engineName].FailedTranslations = 1
		}
	}

	// Update protocol stats
	a.updateProtocolStats(fromProtocol, success, now)
	a.updateProtocolStats(toProtocol, success, now)

	// Update totals
	a.totalTranslations++
	if !success {
		a.totalErrors++
	}

	// Update hourly data
	a.updateHourlyData(now, success)

	// Update daily data
	a.updateDailyData(now, success)
}

// RecordLatency records translation latency
func (a *Analytics) RecordLatency(engineName string, latency time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update engine metrics
	if metrics, exists := a.translations[engineName]; exists {
		// Calculate running average
		totalRequests := metrics.SuccessfulTranslations + metrics.FailedTranslations
		if totalRequests > 0 {
			metrics.AvgLatency = (metrics.AvgLatency*time.Duration(totalRequests-1) + latency) / time.Duration(totalRequests)
		}

		// Update min/max
		if latency > metrics.MaxLatency {
			metrics.MaxLatency = latency
		}
		if latency < metrics.MinLatency || metrics.MinLatency == 0 {
			metrics.MinLatency = latency
		}
	}

	// Update global performance metrics
	a.updatePerformanceMetrics(latency)
}

// RecordError records an error
func (a *Analytics) RecordError(errorType string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.errorCounts[errorType]++
	a.lastError = time.Now()
}

// RecordBytes records bytes transferred
func (a *Analytics) RecordBytes(engineName string, bytesIn, bytesOut int64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if metrics, exists := a.translations[engineName]; exists {
		metrics.TotalBytesIn += bytesIn
		metrics.TotalBytesOut += bytesOut
	}
}

// GetData returns the complete analytics data
func (a *Analytics) GetData() *AnalyticsData {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Create summary
	summary := &AnalyticsSummary{
		TotalTranslations:    a.totalTranslations,
		TotalErrors:          a.totalErrors,
		TotalBytesTranslated: a.calculateTotalBytes(),
		ActiveEngines:        len(a.translations),
		ActiveProtocols:      len(a.protocolStats),
	}

	if a.totalTranslations > 0 {
		summary.SuccessRate = float64(a.totalTranslations-a.totalErrors) / float64(a.totalTranslations) * 100
	}

	// Create performance metrics
	performance := &PerformanceMetrics{
		AvgTranslationTime:  a.avgTranslationTime,
		MaxTranslationTime:  a.maxTranslationTime,
		MinTranslationTime:  a.minTranslationTime,
		ThroughputPerSecond: a.calculateThroughput(),
	}

	// Create error analysis
	errorAnalysis := &ErrorAnalysis{
		TotalErrors:   a.totalErrors,
		ErrorTypes:    make(map[string]int),
		LastErrorTime: a.lastError,
	}

	for errorType, count := range a.errorCounts {
		errorAnalysis.ErrorTypes[errorType] = count
	}

	if a.totalTranslations > 0 {
		errorAnalysis.ErrorRate = float64(a.totalErrors) / float64(a.totalTranslations) * 100
	}

	// Find most common error
	maxCount := 0
	for errorType, count := range a.errorCounts {
		if count > maxCount {
			maxCount = count
			errorAnalysis.MostCommonError = errorType
		}
	}

	return &AnalyticsData{
		Summary:            summary,
		TranslationEngines: a.copyTranslationMetrics(),
		ProtocolStats:      a.copyProtocolMetrics(),
		HourlyData:         a.copyHourlyData(),
		DailyData:          a.copyDailyData(),
		Performance:        performance,
		ErrorAnalysis:      errorAnalysis,
		GeneratedAt:        time.Now(),
	}
}

// Cleanup removes old data to prevent memory leaks
func (a *Analytics) Cleanup() {
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour) // Keep only last 24 hours of hourly data

	// Clean up old hourly data
	for key, data := range a.hourlyData {
		if time.Date(now.Year(), now.Month(), now.Day(), data.Hour, 0, 0, 0, now.Location()).Before(cutoff) {
			delete(a.hourlyData, key)
		}
	}

	// Clean up old daily data (keep last 30 days)
	dailyCutoff := now.Add(-30 * 24 * time.Hour)
	for key, data := range a.dailyData {
		if data.Date.Before(dailyCutoff) {
			delete(a.dailyData, key)
		}
	}
}

// Helper methods

func (a *Analytics) updateProtocolStats(protocol string, success bool, timestamp time.Time) {
	if stats, exists := a.protocolStats[protocol]; exists {
		stats.TotalRequests++
		stats.LastActivity = timestamp
		if !success {
			stats.ErrorRate = float64(stats.TotalRequests-1) / float64(stats.TotalRequests) * 100
		}
	} else {
		a.protocolStats[protocol] = &ProtocolMetrics{
			Protocol:      protocol,
			TotalRequests: 1,
			LastActivity:  timestamp,
		}
		if !success {
			a.protocolStats[protocol].ErrorRate = 100.0
		}
	}
}

func (a *Analytics) updateHourlyData(timestamp time.Time, success bool) {
	hour := timestamp.Hour()
	key := timestamp.Format("2006-01-02-15")

	if data, exists := a.hourlyData[key]; exists {
		data.Translations++
		if !success {
			data.Errors++
		}
	} else {
		a.hourlyData[key] = &HourlyData{
			Hour:         hour,
			Translations: 1,
		}
		if !success {
			a.hourlyData[key].Errors = 1
		}
	}
}

func (a *Analytics) updateDailyData(timestamp time.Time, success bool) {
	date := timestamp.Truncate(24 * time.Hour)
	key := date.Format("2006-01-02")

	if data, exists := a.dailyData[key]; exists {
		data.Translations++
		if !success {
			data.Errors++
		}
	} else {
		a.dailyData[key] = &DailyData{
			Date:         date,
			Translations: 1,
		}
		if !success {
			a.dailyData[key].Errors = 1
		}
	}
}

func (a *Analytics) updatePerformanceMetrics(latency time.Duration) {
	// Update global performance metrics
	if a.avgTranslationTime == 0 {
		a.avgTranslationTime = latency
	} else {
		a.avgTranslationTime = (a.avgTranslationTime + latency) / 2
	}

	if latency > a.maxTranslationTime {
		a.maxTranslationTime = latency
	}

	if latency < a.minTranslationTime {
		a.minTranslationTime = latency
	}
}

func (a *Analytics) calculateTotalBytes() int64 {
	total := int64(0)
	for _, metrics := range a.translations {
		total += metrics.TotalBytesOut
	}
	return total
}

func (a *Analytics) calculateThroughput() float64 {
	// Calculate throughput based on recent activity
	now := time.Now()
	recentTranslations := 0
	timeWindow := 1 * time.Minute

	for _, metrics := range a.translations {
		if now.Sub(metrics.LastActivity) < timeWindow {
			recentTranslations += metrics.TotalRequests
		}
	}

	return float64(recentTranslations) / timeWindow.Seconds()
}

func (a *Analytics) copyTranslationMetrics() map[string]*TranslationMetrics {
	result := make(map[string]*TranslationMetrics)
	for key, metrics := range a.translations {
		// Create a copy
		copy := *metrics
		result[key] = &copy
	}
	return result
}

func (a *Analytics) copyProtocolMetrics() map[string]*ProtocolMetrics {
	result := make(map[string]*ProtocolMetrics)
	for key, metrics := range a.protocolStats {
		// Create a copy
		copy := *metrics
		result[key] = &copy
	}
	return result
}

func (a *Analytics) copyHourlyData() map[string]*HourlyData {
	result := make(map[string]*HourlyData)
	for key, data := range a.hourlyData {
		// Create a copy
		copy := *data
		result[key] = &copy
	}
	return result
}

func (a *Analytics) copyDailyData() map[string]*DailyData {
	result := make(map[string]*DailyData)
	for key, data := range a.dailyData {
		// Create a copy
		copy := *data
		result[key] = &copy
	}
	return result
}
