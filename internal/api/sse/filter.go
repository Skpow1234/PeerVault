package sse

import (
	"log/slog"
	"sync"
	"time"
)

// FilterManager manages event filtering for SSE connections
type FilterManager struct {
	// Event filters by topic
	filters map[string]*EventFilter

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger
	logger *slog.Logger
}

// EventFilter represents a filter for events
type EventFilter struct {
	// Topic this filter applies to
	Topic string

	// Event types to include (empty means all)
	IncludeTypes []string

	// Event types to exclude
	ExcludeTypes []string

	// Data field filters
	DataFilters map[string]interface{}

	// Rate limiting
	RateLimit *RateLimit

	// Created timestamp
	CreatedAt time.Time
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	// Maximum events per second
	MaxEventsPerSecond int

	// Maximum events per minute
	MaxEventsPerMinute int

	// Last event timestamp
	LastEvent time.Time

	// Event counters
	EventsThisSecond int
	EventsThisMinute int

	// Counter reset timestamps
	LastSecondReset time.Time
	LastMinuteReset time.Time
}

// NewFilterManager creates a new filter manager
func NewFilterManager(logger *slog.Logger) *FilterManager {
	return &FilterManager{
		filters: make(map[string]*EventFilter),
		logger:  logger,
	}
}

// AddFilter adds a filter for a topic
func (fm *FilterManager) AddFilter(topic string, filter *EventFilter) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	filter.Topic = topic
	filter.CreatedAt = time.Now()
	fm.filters[topic] = filter

	fm.logger.Info("Event filter added", "topic", topic, "filter", filter)
}

// RemoveFilter removes a filter for a topic
func (fm *FilterManager) RemoveFilter(topic string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	delete(fm.filters, topic)
	fm.logger.Info("Event filter removed", "topic", topic)
}

// GetFilter returns a filter for a topic
func (fm *FilterManager) GetFilter(topic string) (*EventFilter, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	filter, exists := fm.filters[topic]
	return filter, exists
}

// ShouldSendEvent determines if an event should be sent based on filters
func (fm *FilterManager) ShouldSendEvent(topic string, event *Event) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	filter, exists := fm.filters[topic]
	if !exists {
		// No filter means send all events
		return true
	}

	// Check event type filters
	if !fm.checkEventTypeFilter(event.Type, filter) {
		return false
	}

	// Check data filters
	if !fm.checkDataFilters(event.Data, filter.DataFilters) {
		return false
	}

	// Check rate limiting
	if !fm.checkRateLimit(filter.RateLimit) {
		return false
	}

	return true
}

// checkEventTypeFilter checks if the event type matches the filter
func (fm *FilterManager) checkEventTypeFilter(eventType string, filter *EventFilter) bool {
	// If include types are specified, event must be in the list
	if len(filter.IncludeTypes) > 0 {
		for _, includeType := range filter.IncludeTypes {
			if includeType == eventType {
				goto checkExclude
			}
		}
		return false
	}

checkExclude:
	// Check exclude types
	for _, excludeType := range filter.ExcludeTypes {
		if excludeType == eventType {
			return false
		}
	}

	return true
}

// checkDataFilters checks if the event data matches the data filters
func (fm *FilterManager) checkDataFilters(data interface{}, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}

	// Convert data to map for filtering
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// If data is not a map, can't apply data filters
		return true
	}

	// Check each filter
	for key, expectedValue := range filters {
		actualValue, exists := dataMap[key]
		if !exists {
			return false
		}

		if actualValue != expectedValue {
			return false
		}
	}

	return true
}

// checkRateLimit checks if the event passes rate limiting
func (fm *FilterManager) checkRateLimit(rateLimit *RateLimit) bool {
	if rateLimit == nil {
		return true
	}

	now := time.Now()

	// Reset counters if needed
	if now.Sub(rateLimit.LastSecondReset) >= time.Second {
		rateLimit.EventsThisSecond = 0
		rateLimit.LastSecondReset = now
	}

	if now.Sub(rateLimit.LastMinuteReset) >= time.Minute {
		rateLimit.EventsThisMinute = 0
		rateLimit.LastMinuteReset = now
	}

	// Check rate limits
	if rateLimit.MaxEventsPerSecond > 0 && rateLimit.EventsThisSecond >= rateLimit.MaxEventsPerSecond {
		return false
	}

	if rateLimit.MaxEventsPerMinute > 0 && rateLimit.EventsThisMinute >= rateLimit.MaxEventsPerMinute {
		return false
	}

	// Update counters
	rateLimit.EventsThisSecond++
	rateLimit.EventsThisMinute++
	rateLimit.LastEvent = now

	return true
}

// GetFilterStats returns statistics about filters
func (fm *FilterManager) GetFilterStats() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_filters": len(fm.filters),
		"filters":       make([]map[string]interface{}, 0, len(fm.filters)),
	}

	for topic, filter := range fm.filters {
		filterStats := map[string]interface{}{
			"topic":         topic,
			"include_types": filter.IncludeTypes,
			"exclude_types": filter.ExcludeTypes,
			"data_filters":  filter.DataFilters,
			"created_at":    filter.CreatedAt,
		}

		if filter.RateLimit != nil {
			filterStats["rate_limit"] = map[string]interface{}{
				"max_events_per_second": filter.RateLimit.MaxEventsPerSecond,
				"max_events_per_minute": filter.RateLimit.MaxEventsPerMinute,
				"events_this_second":    filter.RateLimit.EventsThisSecond,
				"events_this_minute":    filter.RateLimit.EventsThisMinute,
				"last_event":            filter.RateLimit.LastEvent,
			}
		}

		stats["filters"] = append(stats["filters"].([]map[string]interface{}), filterStats)
	}

	return stats
}

// CreateDefaultFilter creates a default filter for a topic
func CreateDefaultFilter(topic string) *EventFilter {
	return &EventFilter{
		Topic:        topic,
		IncludeTypes: []string{},
		ExcludeTypes: []string{},
		DataFilters:  make(map[string]interface{}),
		RateLimit: &RateLimit{
			MaxEventsPerSecond: 10,
			MaxEventsPerMinute: 100,
		},
		CreatedAt: time.Now(),
	}
}

// CreateTypeFilter creates a filter that only includes specific event types
func CreateTypeFilter(topic string, includeTypes []string) *EventFilter {
	filter := CreateDefaultFilter(topic)
	filter.IncludeTypes = includeTypes
	return filter
}

// CreateExcludeFilter creates a filter that excludes specific event types
func CreateExcludeFilter(topic string, excludeTypes []string) *EventFilter {
	filter := CreateDefaultFilter(topic)
	filter.ExcludeTypes = excludeTypes
	return filter
}

// CreateDataFilter creates a filter that filters by data fields
func CreateDataFilter(topic string, dataFilters map[string]interface{}) *EventFilter {
	filter := CreateDefaultFilter(topic)
	filter.DataFilters = dataFilters
	return filter
}

// CreateRateLimitFilter creates a filter with custom rate limiting
func CreateRateLimitFilter(topic string, maxPerSecond, maxPerMinute int) *EventFilter {
	filter := CreateDefaultFilter(topic)
	filter.RateLimit = &RateLimit{
		MaxEventsPerSecond: maxPerSecond,
		MaxEventsPerMinute: maxPerMinute,
	}
	return filter
}
