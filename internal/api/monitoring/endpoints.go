package monitoring

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Handler provides HTTP handlers for monitoring endpoints
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new monitoring handler
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// HandleGetSnapshot returns the current performance snapshot
func (h *Handler) HandleGetSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot := h.service.GetPerformanceSnapshot(r.Context())
	h.respondJSON(w, snapshot)
}

// HandleGetRealTimeMetrics returns real-time performance metrics
func (h *Handler) HandleGetRealTimeMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.service.GetRealTimeMetrics()
	h.respondJSON(w, metrics)
}

// HandleGetResponseTimeStats returns response time statistics
func (h *Handler) HandleGetResponseTimeStats(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	method := r.URL.Query().Get("method")
	
	stats := h.service.GetResponseTimeStats(endpoint, method)
	h.respondJSON(w, stats)
}

// HandleGetThroughputStats returns throughput statistics
func (h *Handler) HandleGetThroughputStats(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	method := r.URL.Query().Get("method")
	
	stats := h.service.GetThroughputStats(endpoint, method)
	h.respondJSON(w, stats)
}

// HandleGetEndpointPerformance returns performance stats for an endpoint
func (h *Handler) HandleGetEndpointPerformance(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		http.Error(w, "Endpoint parameter is required", http.StatusBadRequest)
		return
	}
	
	method := r.URL.Query().Get("method")
	
	performance := h.service.GetEndpointPerformance(endpoint, method)
	h.respondJSON(w, performance)
}

// HandleGetPerformanceTrend returns performance trend data
func (h *Handler) HandleGetPerformanceTrend(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	
	bucketSizeStr := r.URL.Query().Get("bucket_size")
	bucketSize := 1 * time.Minute
	if bucketSizeStr != "" {
		duration, err := time.ParseDuration(bucketSizeStr)
		if err == nil {
			bucketSize = duration
		}
	}
	
	trend := h.service.GetPerformanceTrend(endpoint, bucketSize)
	h.respondJSON(w, trend)
}

// HandleGetPerformanceReport generates a comprehensive performance report
func (h *Handler) HandleGetPerformanceReport(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "5m"
	}
	
	report := h.service.GetPerformanceReport(period)
	h.respondJSON(w, report)
}

// HandleGetAlerts returns performance alerts
func (h *Handler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"
	
	var alerts []PerformanceAlert
	if activeOnly {
		alerts = h.service.GetActiveAlerts()
	} else {
		alerts = h.service.GetAllAlerts()
	}
	
	h.respondJSON(w, alerts)
}

// HandleSilenceAlert silences a specific alert
func (h *Handler) HandleSilenceAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.URL.Query().Get("alert_id")
	if alertID == "" {
		http.Error(w, "Alert ID parameter is required", http.StatusBadRequest)
		return
	}
	
	err := h.service.SilenceAlert(alertID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	h.respondJSON(w, map[string]string{"status": "silenced", "alert_id": alertID})
}

// HandleResolveAlert manually resolves an alert
func (h *Handler) HandleResolveAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.URL.Query().Get("alert_id")
	if alertID == "" {
		http.Error(w, "Alert ID parameter is required", http.StatusBadRequest)
		return
	}
	
	err := h.service.ResolveAlert(alertID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	h.respondJSON(w, map[string]string{"status": "resolved", "alert_id": alertID})
}

// HandleGetRecommendations returns optimization recommendations
func (h *Handler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	recommendations := h.service.GenerateRecommendations()
	
	// Filter by priority if requested
	priorityFilter := r.URL.Query().Get("priority")
	if priorityFilter != "" {
		filtered := make([]OptimizationRecommendation, 0)
		for _, rec := range recommendations {
			if string(rec.Priority) == priorityFilter {
				filtered = append(filtered, rec)
			}
		}
		recommendations = filtered
	}
	
	// Filter by type if requested
	typeFilter := r.URL.Query().Get("type")
	if typeFilter != "" {
		filtered := make([]OptimizationRecommendation, 0)
		for _, rec := range recommendations {
			if string(rec.Type) == typeFilter {
				filtered = append(filtered, rec)
			}
		}
		recommendations = filtered
	}
	
	// Limit results
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 && limit < len(recommendations) {
			recommendations = recommendations[:limit]
		}
	}
	
	h.respondJSON(w, recommendations)
}

// HandleGetDashboard returns comprehensive dashboard data
func (h *Handler) HandleGetDashboard(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "5m"
	}
	
	type Dashboard struct {
		Snapshot        PerformanceSnapshot            `json:"snapshot"`
		RealTimeMetrics RealTimeMetrics                `json:"real_time_metrics"`
		Alerts          []PerformanceAlert             `json:"alerts"`
		Recommendations []OptimizationRecommendation   `json:"recommendations"`
		Report          PerformanceReport              `json:"report"`
		GeneratedAt     time.Time                      `json:"generated_at"`
	}
	
	dashboard := Dashboard{
		Snapshot:        h.service.GetPerformanceSnapshot(r.Context()),
		RealTimeMetrics: h.service.GetRealTimeMetrics(),
		Alerts:          h.service.GetActiveAlerts(),
		Recommendations: h.service.GenerateRecommendations(),
		Report:          h.service.GetPerformanceReport(period),
		GeneratedAt:     time.Now(),
	}
	
	h.respondJSON(w, dashboard)
}

// HandleGetHealthScore returns the current health score
func (h *Handler) HandleGetHealthScore(w http.ResponseWriter, r *http.Request) {
	snapshot := h.service.GetPerformanceSnapshot(r.Context())
	
	response := map[string]interface{}{
		"health_score": snapshot.HealthScore,
		"grade":        calculateGrade(snapshot.HealthScore),
		"status":       getHealthStatus(snapshot.HealthScore),
		"timestamp":    time.Now(),
		"details": map[string]interface{}{
			"avg_response_time_ms": snapshot.OverallResponseTime.Mean.Milliseconds(),
			"p95_response_time_ms": snapshot.OverallResponseTime.P95.Milliseconds(),
			"throughput_rps":       snapshot.OverallThroughput.RequestsPerSecond,
			"active_alerts":        len(snapshot.ActiveAlerts),
			"goroutines":           snapshot.SystemLoad.Goroutines,
			"memory_percent":       snapshot.SystemLoad.Memory,
		},
	}
	
	h.respondJSON(w, response)
}

// HandleGetMetricsSummary returns a summary of key metrics
func (h *Handler) HandleGetMetricsSummary(w http.ResponseWriter, r *http.Request) {
	snapshot := h.service.GetPerformanceSnapshot(r.Context())
	realtime := h.service.GetRealTimeMetrics()
	
	summary := map[string]interface{}{
		"timestamp": time.Now(),
		"performance": map[string]interface{}{
			"avg_response_time_ms": snapshot.OverallResponseTime.Mean.Milliseconds(),
			"p50_response_time_ms": snapshot.OverallResponseTime.Median.Milliseconds(),
			"p95_response_time_ms": snapshot.OverallResponseTime.P95.Milliseconds(),
			"p99_response_time_ms": snapshot.OverallResponseTime.P99.Milliseconds(),
			"min_response_time_ms": snapshot.OverallResponseTime.Min.Milliseconds(),
			"max_response_time_ms": snapshot.OverallResponseTime.Max.Milliseconds(),
		},
		"throughput": map[string]interface{}{
			"current_rps":          realtime.CurrentRPS,
			"requests_per_minute":  snapshot.OverallThroughput.RequestsPerMinute,
			"peak_rps":             snapshot.OverallThroughput.PeakRPS,
			"bytes_per_second":     snapshot.OverallThroughput.BytesPerSecond,
		},
		"system": map[string]interface{}{
			"active_requests": realtime.ActiveRequests,
			"goroutines":      snapshot.SystemLoad.Goroutines,
			"memory_percent":  snapshot.SystemLoad.Memory,
		},
		"health": map[string]interface{}{
			"score":         snapshot.HealthScore,
			"grade":         calculateGrade(snapshot.HealthScore),
			"status":        getHealthStatus(snapshot.HealthScore),
			"active_alerts": len(snapshot.ActiveAlerts),
		},
	}
	
	h.respondJSON(w, summary)
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Helper functions

func getHealthStatus(score float64) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 75:
		return "good"
	case score >= 60:
		return "fair"
	case score >= 40:
		return "poor"
	default:
		return "critical"
	}
}

