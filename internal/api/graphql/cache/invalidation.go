package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// InvalidationManager manages cache invalidation strategies
type InvalidationManager struct {
	queryCache *QueryCache
	strategies map[string]InvalidationStrategy
	mu         sync.RWMutex
	logger     *slog.Logger
}

// InvalidationRule defines when and how to invalidate cache entries
type InvalidationRule struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	Pattern    string                  `json:"pattern"`
	Tags       []string                `json:"tags"`
	Conditions []InvalidationCondition `json:"conditions"`
	Strategy   InvalidationStrategy    `json:"strategy"`
	Enabled    bool                    `json:"enabled"`
	CreatedAt  time.Time               `json:"createdAt"`
	UpdatedAt  time.Time               `json:"updatedAt"`
}

// InvalidationCondition defines a condition for cache invalidation
type InvalidationCondition struct {
	Type     ConditionType          `json:"type"`
	Field    string                 `json:"field"`
	Operator ConditionOperator      `json:"operator"`
	Value    interface{}            `json:"value"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ConditionType defines the type of invalidation condition
type ConditionType string

const (
	ConditionTypeTime    ConditionType = "time"
	ConditionTypeEvent   ConditionType = "event"
	ConditionTypeData    ConditionType = "data"
	ConditionTypePattern ConditionType = "pattern"
	ConditionTypeTag     ConditionType = "tag"
)

// ConditionOperator defines the operator for conditions
type ConditionOperator string

const (
	OperatorEquals    ConditionOperator = "equals"
	OperatorNotEquals ConditionOperator = "not_equals"
	OperatorContains  ConditionOperator = "contains"
	OperatorMatches   ConditionOperator = "matches"
	OperatorGreater   ConditionOperator = "greater"
	OperatorLess      ConditionOperator = "less"
	OperatorAfter     ConditionOperator = "after"
	OperatorBefore    ConditionOperator = "before"
)

// InvalidationEvent represents an event that can trigger cache invalidation
type InvalidationEvent struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      []string               `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewInvalidationManager creates a new invalidation manager
func NewInvalidationManager(queryCache *QueryCache, logger *slog.Logger) *InvalidationManager {
	return &InvalidationManager{
		queryCache: queryCache,
		strategies: make(map[string]InvalidationStrategy),
		logger:     logger,
	}
}

// RegisterStrategy registers an invalidation strategy
func (im *InvalidationManager) RegisterStrategy(name string, strategy InvalidationStrategy) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.strategies[name] = strategy
	im.logger.Info("Registered invalidation strategy", "name", name)
}

// AddRule adds a new invalidation rule
func (im *InvalidationManager) AddRule(rule *InvalidationRule) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if rule.ID == "" {
		rule.ID = generateRuleID()
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// Validate the rule
	if err := im.validateRule(rule); err != nil {
		return err
	}

	// Store the rule (in a real implementation, this would be persisted)
	im.logger.Info("Added invalidation rule", "id", rule.ID, "name", rule.Name)
	return nil
}

// RemoveRule removes an invalidation rule
func (im *InvalidationManager) RemoveRule(ruleID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Remove the rule (in a real implementation, this would be persisted)
	im.logger.Info("Removed invalidation rule", "id", ruleID)
	return nil
}

// ProcessEvent processes an invalidation event
func (im *InvalidationManager) ProcessEvent(ctx context.Context, event *InvalidationEvent) error {
	im.mu.RLock()
	rules := im.getMatchingRules(event)
	im.mu.RUnlock()

	for _, rule := range rules {
		if err := im.executeRule(ctx, rule, event); err != nil {
			im.logger.Error("Failed to execute invalidation rule",
				"ruleId", rule.ID,
				"eventType", event.Type,
				"error", err)
		}
	}

	return nil
}

// getMatchingRules returns rules that match the given event
func (im *InvalidationManager) getMatchingRules(event *InvalidationEvent) []*InvalidationRule {
	// This is a simplified implementation
	// In a real implementation, you would evaluate conditions against the event

	var matchingRules []*InvalidationRule

	// For now, return rules that match by event type or tags
	for _, rule := range im.getAllRules() {
		if !rule.Enabled {
			continue
		}

		// Check if rule matches event type
		if rule.Pattern != "" && matchesPattern(event.Type, rule.Pattern) {
			matchingRules = append(matchingRules, rule)
			continue
		}

		// Check if rule matches event tags
		if len(rule.Tags) > 0 && hasMatchingTags(event.Tags, rule.Tags) {
			matchingRules = append(matchingRules, rule)
			continue
		}
	}

	return matchingRules
}

// executeRule executes an invalidation rule
func (im *InvalidationManager) executeRule(ctx context.Context, rule *InvalidationRule, event *InvalidationEvent) error {
	im.logger.Info("Executing invalidation rule",
		"ruleId", rule.ID,
		"ruleName", rule.Name,
		"eventType", event.Type)

	switch rule.Strategy {
	case InvalidationStrategyTTL:
		// TTL-based invalidation is handled automatically by the cache
		return nil

	case InvalidationStrategyEvent:
		return im.executeEventBasedInvalidation(ctx, rule, event)

	case InvalidationStrategyHybrid:
		// Execute both TTL and event-based invalidation
		if err := im.executeEventBasedInvalidation(ctx, rule, event); err != nil {
			return err
		}
		return nil

	default:
		return im.executeEventBasedInvalidation(ctx, rule, event)
	}
}

// executeEventBasedInvalidation executes event-based cache invalidation
func (im *InvalidationManager) executeEventBasedInvalidation(ctx context.Context, rule *InvalidationRule, event *InvalidationEvent) error {
	// Determine what to invalidate based on the rule
	if rule.Pattern != "" {
		if err := im.queryCache.InvalidateByPattern(ctx, rule.Pattern); err != nil {
			return err
		}
	}

	if len(rule.Tags) > 0 {
		if err := im.queryCache.InvalidateByTags(ctx, rule.Tags); err != nil {
			return err
		}
	}

	// Publish invalidation event
	im.queryCache.PublishInvalidation(rule.Pattern)

	im.logger.Info("Executed event-based invalidation",
		"ruleId", rule.ID,
		"pattern", rule.Pattern,
		"tags", rule.Tags)

	return nil
}

// validateRule validates an invalidation rule
func (im *InvalidationManager) validateRule(rule *InvalidationRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.Pattern == "" && len(rule.Tags) == 0 {
		return fmt.Errorf("rule must have either pattern or tags")
	}

	// Validate strategy
	if _, exists := im.strategies[string(rule.Strategy)]; !exists {
		return fmt.Errorf("invalid strategy: %s", rule.Strategy)
	}

	// Validate conditions
	for i, condition := range rule.Conditions {
		if err := im.validateCondition(&condition); err != nil {
			return fmt.Errorf("invalid condition at index %d: %w", i, err)
		}
	}

	return nil
}

// validateCondition validates an invalidation condition
func (im *InvalidationManager) validateCondition(condition *InvalidationCondition) error {
	if condition.Type == "" {
		return fmt.Errorf("condition type is required")
	}

	if condition.Field == "" {
		return fmt.Errorf("condition field is required")
	}

	if condition.Operator == "" {
		return fmt.Errorf("condition operator is required")
	}

	// Validate operator for condition type
	switch condition.Type {
	case ConditionTypeTime:
		if condition.Operator != OperatorAfter && condition.Operator != OperatorBefore {
			return fmt.Errorf("invalid operator for time condition: %s", condition.Operator)
		}
	case ConditionTypeEvent:
		if condition.Operator != OperatorEquals && condition.Operator != OperatorContains {
			return fmt.Errorf("invalid operator for event condition: %s", condition.Operator)
		}
	case ConditionTypeData:
		// Most operators are valid for data conditions
	case ConditionTypePattern:
		if condition.Operator != OperatorMatches {
			return fmt.Errorf("invalid operator for pattern condition: %s", condition.Operator)
		}
	case ConditionTypeTag:
		if condition.Operator != OperatorEquals && condition.Operator != OperatorContains {
			return fmt.Errorf("invalid operator for tag condition: %s", condition.Operator)
		}
	}

	return nil
}

// getAllRules returns all invalidation rules
func (im *InvalidationManager) getAllRules() []*InvalidationRule {
	// In a real implementation, this would fetch from persistent storage
	// For now, return some example rules

	return []*InvalidationRule{
		{
			ID:       "file-update-rule",
			Name:     "File Update Invalidation",
			Pattern:  "file.*",
			Strategy: InvalidationStrategyEvent,
			Enabled:  true,
			Conditions: []InvalidationCondition{
				{
					Type:     ConditionTypeEvent,
					Field:    "type",
					Operator: OperatorEquals,
					Value:    "file.updated",
				},
			},
		},
		{
			ID:       "node-change-rule",
			Name:     "Node Change Invalidation",
			Tags:     []string{"node", "peer"},
			Strategy: InvalidationStrategyEvent,
			Enabled:  true,
			Conditions: []InvalidationCondition{
				{
					Type:     ConditionTypeEvent,
					Field:    "type",
					Operator: OperatorContains,
					Value:    "node",
				},
			},
		},
		{
			ID:       "system-metrics-rule",
			Name:     "System Metrics Invalidation",
			Pattern:  "system.*",
			Strategy: InvalidationStrategyTTL,
			Enabled:  true,
			Conditions: []InvalidationCondition{
				{
					Type:     ConditionTypeTime,
					Field:    "timestamp",
					Operator: OperatorAfter,
					Value:    time.Now().Add(-5 * time.Minute),
				},
			},
		},
	}
}

// Helper functions

// generateRuleID generates a unique rule ID
func generateRuleID() string {
	return fmt.Sprintf("rule_%d", time.Now().UnixNano())
}

// matchesPattern checks if a string matches a pattern
func matchesPattern(str, pattern string) bool {
	// This is a simplified pattern matching implementation
	// In a real implementation, you would use proper regex or glob matching
	return str == pattern || contains(str, pattern)
}

// hasMatchingTags checks if two tag slices have any matching tags
func hasMatchingTags(tags1, tags2 []string) bool {
	for _, tag1 := range tags1 {
		for _, tag2 := range tags2 {
			if tag1 == tag2 {
				return true
			}
		}
	}
	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

// indexOf returns the index of the first occurrence of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// CreateDefaultRules creates default invalidation rules
func (im *InvalidationManager) CreateDefaultRules() error {
	defaultRules := []*InvalidationRule{
		{
			ID:       "file-operations",
			Name:     "File Operations Invalidation",
			Pattern:  "file.*",
			Strategy: InvalidationStrategyEvent,
			Enabled:  true,
		},
		{
			ID:       "node-operations",
			Name:     "Node Operations Invalidation",
			Tags:     []string{"node", "peer"},
			Strategy: InvalidationStrategyEvent,
			Enabled:  true,
		},
		{
			ID:       "system-metrics",
			Name:     "System Metrics Invalidation",
			Pattern:  "system.*",
			Strategy: InvalidationStrategyTTL,
			Enabled:  true,
		},
	}

	for _, rule := range defaultRules {
		if err := im.AddRule(rule); err != nil {
			im.logger.Error("Failed to add default rule", "ruleId", rule.ID, "error", err)
		}
	}

	return nil
}
