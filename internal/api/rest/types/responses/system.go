package responses

import "time"

// SystemInfoResponse represents system information response
type SystemInfoResponse struct {
	Version      string        `json:"version"`
	Uptime       time.Duration `json:"uptime"`
	StartTime    time.Time     `json:"start_time"`
	StorageUsed  int64         `json:"storage_used"`
	StorageTotal int64         `json:"storage_total"`
	PeerCount    int           `json:"peer_count"`
	FileCount    int           `json:"file_count"`
}

// MetricsResponse represents metrics response
type MetricsResponse struct {
	RequestsTotal     int64     `json:"requests_total"`
	RequestsPerMinute float64   `json:"requests_per_minute"`
	ActiveConnections int       `json:"active_connections"`
	StorageUsage      float64   `json:"storage_usage_percent"`
	LastUpdated       time.Time `json:"last_updated"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
