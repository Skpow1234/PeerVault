package analytics

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Handler provides HTTP handlers for analytics endpoints
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new analytics handler
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// HandleGetSummary returns analytics summary
func (h *Handler) HandleGetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetAnalyticsSummary()
	if err != nil {
		h.logger.Error("Failed to get analytics summary", "error", err)
		http.Error(w, "Failed to get analytics summary", http.StatusInternalServerError)
		return
	}
	
	h.respondJSON(w, summary)
}

// HandleGetUsageMetrics returns usage metrics for a time period
func (h *Handler) HandleGetUsageMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse time range from query parameters
	startTime, endTime, err := h.parseTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	metrics, err := h.service.GetUsageMetrics(startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get usage metrics", "error", err)
		http.Error(w, "Failed to get usage metrics", http.StatusInternalServerError)
		return
	}
	
	h.respondJSON(w, metrics)
}

// HandleGetEndpointStats returns statistics for a specific endpoint
func (h *Handler) HandleGetEndpointStats(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		http.Error(w, "Endpoint parameter is required", http.StatusBadRequest)
		return
	}
	
	startTime, endTime, err := h.parseTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	stats, err := h.service.GetEndpointStats(endpoint, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get endpoint stats", "error", err, "endpoint", endpoint)
		http.Error(w, "Failed to get endpoint stats", http.StatusInternalServerError)
		return
	}
	
	h.respondJSON(w, stats)
}

// HandleGetUserBehavior returns behavior metrics for a specific user
func (h *Handler) HandleGetUserBehavior(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID parameter is required", http.StatusBadRequest)
		return
	}
	
	startTime, endTime, err := h.parseTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	behavior, err := h.service.GetUserBehavior(userID, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get user behavior", "error", err, "user_id", userID)
		http.Error(w, "Failed to get user behavior", http.StatusInternalServerError)
		return
	}
	
	h.respondJSON(w, behavior)
}

// HandleGetUsageTrends returns usage trends over time
func (h *Handler) HandleGetUsageTrends(w http.ResponseWriter, r *http.Request) {
	startTime, endTime, err := h.parseTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "hour"
	}
	
	trends, err := h.service.GetUsageTrends(startTime, endTime, interval)
	if err != nil {
		h.logger.Error("Failed to get usage trends", "error", err)
		http.Error(w, "Failed to get usage trends", http.StatusInternalServerError)
		return
	}
	
	h.respondJSON(w, trends)
}

// HandleGetPopularityMetrics returns popularity metrics
func (h *Handler) HandleGetPopularityMetrics(w http.ResponseWriter, r *http.Request) {
	startTime, endTime, err := h.parseTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	metrics, err := h.service.GetPopularityMetrics(startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get popularity metrics", "error", err)
		http.Error(w, "Failed to get popularity metrics", http.StatusInternalServerError)
		return
	}
	
	h.respondJSON(w, metrics)
}

// HandleQueryAPICalls returns API calls matching query parameters
func (h *Handler) HandleQueryAPICalls(w http.ResponseWriter, r *http.Request) {
	query := &AnalyticsQuery{}
	
	// Parse query parameters
	startTime, endTime, err := h.parseTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	query.StartTime = startTime
	query.EndTime = endTime
	query.Endpoint = r.URL.Query().Get("endpoint")
	query.Method = r.URL.Query().Get("method")
	query.UserID = r.URL.Query().Get("user_id")
	
	if statusStr := r.URL.Query().Get("status_code"); statusStr != "" {
		statusCode, err := strconv.Atoi(statusStr)
		if err == nil {
			query.StatusCode = statusCode
		}
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil {
			query.Limit = limit
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil {
			query.Offset = offset
		}
	}
	
	calls, err := h.service.QueryAPICalls(query)
	if err != nil {
		h.logger.Error("Failed to query API calls", "error", err)
		http.Error(w, "Failed to query API calls", http.StatusInternalServerError)
		return
	}
	
	h.respondJSON(w, calls)
}

// HandleGetDashboard returns a comprehensive dashboard view
func (h *Handler) HandleGetDashboard(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}
	
	startTime, endTime := GetTimeWindow(period)
	
	// Get various metrics in parallel
	type result struct {
		Summary    *AnalyticsSummary    `json:"summary"`
		Metrics    *UsageMetrics        `json:"metrics"`
		Popularity *PopularityMetrics   `json:"popularity"`
		Trends     []*UsageTrend        `json:"trends"`
	}
	
	res := &result{}
	
	// Get summary
	summary, err := h.service.GetAnalyticsSummary()
	if err != nil {
		h.logger.Error("Failed to get summary", "error", err)
	} else {
		res.Summary = summary
	}
	
	// Get metrics
	metrics, err := h.service.GetUsageMetrics(startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get metrics", "error", err)
	} else {
		res.Metrics = metrics
	}
	
	// Get popularity
	popularity, err := h.service.GetPopularityMetrics(startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get popularity", "error", err)
	} else {
		res.Popularity = popularity
	}
	
	// Get trends
	trends, err := h.service.GetUsageTrends(startTime, endTime, "hour")
	if err != nil {
		h.logger.Error("Failed to get trends", "error", err)
	} else {
		res.Trends = trends
	}
	
	h.respondJSON(w, res)
}

// parseTimeRange parses start and end time from query parameters
func (h *Handler) parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	// Check for period parameter (e.g., "24h", "7d", "30d")
	if period := r.URL.Query().Get("period"); period != "" {
		start, end := GetTimeWindow(period)
		return start, end, nil
	}
	
	// Parse individual start and end times
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	endTime := now
	
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		parsed, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			// Try Unix timestamp
			timestamp, err := strconv.ParseInt(startStr, 10, 64)
			if err == nil {
				startTime = time.Unix(timestamp, 0)
			}
		} else {
			startTime = parsed
		}
	}
	
	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		parsed, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			// Try Unix timestamp
			timestamp, err := strconv.ParseInt(endStr, 10, 64)
			if err == nil {
				endTime = time.Unix(timestamp, 0)
			}
		} else {
			endTime = parsed
		}
	}
	
	return startTime, endTime, nil
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

