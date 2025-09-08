package stitching

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

// SchemaVersionManager manages schema versions and migrations
type SchemaVersionManager struct {
	versions map[string]*SchemaVersion
	stitcher *SchemaStitcher
	mu       sync.RWMutex
	logger   *slog.Logger
}

// SchemaVersion represents a version of a schema
type SchemaVersion struct {
	ID          string                 `json:"id"`
	SchemaID    string                 `json:"schemaId"`
	Version     string                 `json:"version"`
	Schema      *SchemaDefinition      `json:"schema"`
	Stitched    *StitchedSchema        `json:"stitched,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	CreatedBy   string                 `json:"createdBy"`
	Description string                 `json:"description"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SchemaMigration represents a migration between schema versions
type SchemaMigration struct {
	ID          string                 `json:"id"`
	FromVersion string                 `json:"fromVersion"`
	ToVersion   string                 `json:"toVersion"`
	SchemaID    string                 `json:"schemaId"`
	Type        MigrationType          `json:"type"`
	Changes     []SchemaChange         `json:"changes"`
	Script      string                 `json:"script,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	CreatedBy   string                 `json:"createdBy"`
	Status      MigrationStatus        `json:"status"`
	ExecutedAt  *time.Time             `json:"executedAt,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SchemaChange represents a change in a schema
type SchemaChange struct {
	Type        ChangeType             `json:"type"`
	Path        string                 `json:"path"`
	OldValue    interface{}            `json:"oldValue,omitempty"`
	NewValue    interface{}            `json:"newValue,omitempty"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// MigrationType defines the type of migration
type MigrationType string

const (
	MigrationTypeAdd       MigrationType = "add"
	MigrationTypeRemove    MigrationType = "remove"
	MigrationTypeModify    MigrationType = "modify"
	MigrationTypeRename    MigrationType = "rename"
	MigrationTypeDeprecate MigrationType = "deprecate"
	MigrationTypeBreaking  MigrationType = "breaking"
)

// MigrationStatus defines the status of a migration
type MigrationStatus string

const (
	MigrationStatusPending   MigrationStatus = "pending"
	MigrationStatusRunning   MigrationStatus = "running"
	MigrationStatusCompleted MigrationStatus = "completed"
	MigrationStatusFailed    MigrationStatus = "failed"
	MigrationStatusRollback  MigrationStatus = "rollback"
)

// ChangeType defines the type of schema change
type ChangeType string

const (
	ChangeTypeTypeAdded        ChangeType = "type_added"
	ChangeTypeTypeRemoved      ChangeType = "type_removed"
	ChangeTypeTypeModified     ChangeType = "type_modified"
	ChangeTypeFieldAdded       ChangeType = "field_added"
	ChangeTypeFieldRemoved     ChangeType = "field_removed"
	ChangeTypeFieldModified    ChangeType = "field_modified"
	ChangeTypeFieldRenamed     ChangeType = "field_renamed"
	ChangeTypeFieldDeprecated  ChangeType = "field_deprecated"
	ChangeTypeArgumentAdded    ChangeType = "argument_added"
	ChangeTypeArgumentRemoved  ChangeType = "argument_removed"
	ChangeTypeArgumentModified ChangeType = "argument_modified"
	ChangeTypeDirectiveAdded   ChangeType = "directive_added"
	ChangeTypeDirectiveRemoved ChangeType = "directive_removed"
)

// VersioningConfig holds configuration for schema versioning
type VersioningConfig struct {
	EnableVersioning    bool          `json:"enableVersioning"`
	MaxVersions         int           `json:"maxVersions"`
	RetentionPeriod     time.Duration `json:"retentionPeriod"`
	AutoMigration       bool          `json:"autoMigration"`
	BackupBeforeMigrate bool          `json:"backupBeforeMigrate"`
	ValidateMigrations  bool          `json:"validateMigrations"`
}

// DefaultVersioningConfig returns the default versioning configuration
func DefaultVersioningConfig() *VersioningConfig {
	return &VersioningConfig{
		EnableVersioning:    true,
		MaxVersions:         10,
		RetentionPeriod:     30 * 24 * time.Hour, // 30 days
		AutoMigration:       false,
		BackupBeforeMigrate: true,
		ValidateMigrations:  true,
	}
}

// NewSchemaVersionManager creates a new schema version manager
func NewSchemaVersionManager(stitcher *SchemaStitcher, logger *slog.Logger) *SchemaVersionManager {
	return &SchemaVersionManager{
		versions: make(map[string]*SchemaVersion),
		stitcher: stitcher,
		logger:   logger,
	}
}

// CreateVersion creates a new schema version
func (svm *SchemaVersionManager) CreateVersion(schemaID, version, description, createdBy string) (*SchemaVersion, error) {
	svm.mu.Lock()
	defer svm.mu.Unlock()

	// Get the current schema
	schema, exists := svm.stitcher.GetSchema(schemaID)
	if !exists {
		return nil, fmt.Errorf("schema %s not found", schemaID)
	}

	// Create version ID
	versionID := fmt.Sprintf("%s:%s", schemaID, version)

	// Check if version already exists
	if _, exists := svm.versions[versionID]; exists {
		return nil, fmt.Errorf("version %s already exists for schema %s", version, schemaID)
	}

	// Create new version
	schemaVersion := &SchemaVersion{
		ID:          versionID,
		SchemaID:    schemaID,
		Version:     version,
		Schema:      schema,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Description: description,
		IsActive:    false,
		Metadata:    make(map[string]interface{}),
	}

	// Store version
	svm.versions[versionID] = schemaVersion

	svm.logger.Info("Created schema version",
		"schemaId", schemaID,
		"version", version,
		"createdBy", createdBy)

	return schemaVersion, nil
}

// GetVersion returns a specific schema version
func (svm *SchemaVersionManager) GetVersion(schemaID, version string) (*SchemaVersion, bool) {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	versionID := fmt.Sprintf("%s:%s", schemaID, version)
	versionObj, exists := svm.versions[versionID]
	return versionObj, exists
}

// ListVersions returns all versions for a schema
func (svm *SchemaVersionManager) ListVersions(schemaID string) []*SchemaVersion {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	var versions []*SchemaVersion
	for _, version := range svm.versions {
		if version.SchemaID == schemaID {
			versions = append(versions, version)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].CreatedAt.After(versions[j].CreatedAt)
	})

	return versions
}

// ActivateVersion activates a specific schema version
func (svm *SchemaVersionManager) ActivateVersion(schemaID, version string) error {
	svm.mu.Lock()
	defer svm.mu.Unlock()

	versionID := fmt.Sprintf("%s:%s", schemaID, version)
	versionObj, exists := svm.versions[versionID]
	if !exists {
		return fmt.Errorf("version %s not found for schema %s", version, schemaID)
	}

	// Deactivate all other versions for this schema
	for _, v := range svm.versions {
		if v.SchemaID == schemaID {
			v.IsActive = false
		}
	}

	// Activate the specified version
	versionObj.IsActive = true

	// Update the schema in the stitcher
	if err := svm.stitcher.UpdateSchema(versionObj.Schema); err != nil {
		return fmt.Errorf("failed to update schema: %w", err)
	}

	svm.logger.Info("Activated schema version",
		"schemaId", schemaID,
		"version", version)

	return nil
}

// CreateMigration creates a migration between two schema versions
func (svm *SchemaVersionManager) CreateMigration(fromVersion, toVersion, schemaID, createdBy string) (*SchemaMigration, error) {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	// Get both versions
	fromVersionObj, exists := svm.GetVersion(schemaID, fromVersion)
	if !exists {
		return nil, fmt.Errorf("from version %s not found", fromVersion)
	}

	toVersionObj, exists := svm.GetVersion(schemaID, toVersion)
	if !exists {
		return nil, fmt.Errorf("to version %s not found", toVersion)
	}

	// Calculate changes
	changes, err := svm.calculateChanges(fromVersionObj.Schema, toVersionObj.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate changes: %w", err)
	}

	// Create migration
	migration := &SchemaMigration{
		ID:          fmt.Sprintf("%s:%s->%s", schemaID, fromVersion, toVersion),
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		SchemaID:    schemaID,
		Type:        svm.determineMigrationType(changes),
		Changes:     changes,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Status:      MigrationStatusPending,
		Metadata:    make(map[string]interface{}),
	}

	svm.logger.Info("Created schema migration",
		"schemaId", schemaID,
		"fromVersion", fromVersion,
		"toVersion", toVersion,
		"changes", len(changes))

	return migration, nil
}

// ExecuteMigration executes a schema migration
func (svm *SchemaVersionManager) ExecuteMigration(migration *SchemaMigration) error {
	svm.mu.Lock()
	defer svm.mu.Unlock()

	// Update migration status
	migration.Status = MigrationStatusRunning
	now := time.Now()
	migration.ExecutedAt = &now

	svm.logger.Info("Executing schema migration",
		"migrationId", migration.ID,
		"type", migration.Type,
		"changes", len(migration.Changes))

	// Execute migration steps
	for _, change := range migration.Changes {
		if err := svm.executeChange(change); err != nil {
			migration.Status = MigrationStatusFailed
			return fmt.Errorf("failed to execute change %s: %w", change.Type, err)
		}
	}

	// Activate the target version
	if err := svm.ActivateVersion(migration.SchemaID, migration.ToVersion); err != nil {
		migration.Status = MigrationStatusFailed
		return fmt.Errorf("failed to activate version %s: %w", migration.ToVersion, err)
	}

	// Mark migration as completed
	migration.Status = MigrationStatusCompleted

	svm.logger.Info("Schema migration completed",
		"migrationId", migration.ID,
		"toVersion", migration.ToVersion)

	return nil
}

// calculateChanges calculates the differences between two schemas
func (svm *SchemaVersionManager) calculateChanges(from, to *SchemaDefinition) ([]SchemaChange, error) {
	var changes []SchemaChange

	// Compare types
	changes = append(changes, svm.compareTypes(from.Types, to.Types)...)

	// Compare queries
	changes = append(changes, svm.compareFields(from.Queries, to.Queries, "Query")...)

	// Compare mutations
	changes = append(changes, svm.compareFields(from.Mutations, to.Mutations, "Mutation")...)

	// Compare subscriptions
	changes = append(changes, svm.compareFields(from.Subscriptions, to.Subscriptions, "Subscription")...)

	// Compare directives
	changes = append(changes, svm.compareDirectives(from.Directives, to.Directives)...)

	return changes, nil
}

// compareTypes compares two type maps and returns changes
func (svm *SchemaVersionManager) compareTypes(from, to map[string]*TypeDefinition) []SchemaChange {
	var changes []SchemaChange

	// Check for added types
	for name, typeDef := range to {
		if _, exists := from[name]; !exists {
			changes = append(changes, SchemaChange{
				Type:        ChangeTypeTypeAdded,
				Path:        fmt.Sprintf("types.%s", name),
				NewValue:    typeDef,
				Description: fmt.Sprintf("Type %s was added", name),
			})
		}
	}

	// Check for removed types
	for name, typeDef := range from {
		if _, exists := to[name]; !exists {
			changes = append(changes, SchemaChange{
				Type:        ChangeTypeTypeRemoved,
				Path:        fmt.Sprintf("types.%s", name),
				OldValue:    typeDef,
				Description: fmt.Sprintf("Type %s was removed", name),
			})
		}
	}

	// Check for modified types
	for name, fromType := range from {
		if toType, exists := to[name]; exists {
			fieldChanges := svm.compareFields(fromType.Fields, toType.Fields, fmt.Sprintf("types.%s", name))
			changes = append(changes, fieldChanges...)
		}
	}

	return changes
}

// compareFields compares two field maps and returns changes
func (svm *SchemaVersionManager) compareFields(from, to map[string]*FieldDefinition, path string) []SchemaChange {
	var changes []SchemaChange

	// Check for added fields
	for name, field := range to {
		if _, exists := from[name]; !exists {
			changes = append(changes, SchemaChange{
				Type:        ChangeTypeFieldAdded,
				Path:        fmt.Sprintf("%s.%s", path, name),
				NewValue:    field,
				Description: fmt.Sprintf("Field %s was added to %s", name, path),
			})
		}
	}

	// Check for removed fields
	for name, field := range from {
		if _, exists := to[name]; !exists {
			changes = append(changes, SchemaChange{
				Type:        ChangeTypeFieldRemoved,
				Path:        fmt.Sprintf("%s.%s", path, name),
				OldValue:    field,
				Description: fmt.Sprintf("Field %s was removed from %s", name, path),
			})
		}
	}

	// Check for modified fields
	for name, fromField := range from {
		if toField, exists := to[name]; exists {
			if svm.fieldsAreDifferent(fromField, toField) {
				changes = append(changes, SchemaChange{
					Type:        ChangeTypeFieldModified,
					Path:        fmt.Sprintf("%s.%s", path, name),
					OldValue:    fromField,
					NewValue:    toField,
					Description: fmt.Sprintf("Field %s was modified in %s", name, path),
				})
			}
		}
	}

	return changes
}

// compareDirectives compares two directive maps and returns changes
func (svm *SchemaVersionManager) compareDirectives(from, to map[string]*DirectiveDefinition) []SchemaChange {
	var changes []SchemaChange

	// Check for added directives
	for name, directive := range to {
		if _, exists := from[name]; !exists {
			changes = append(changes, SchemaChange{
				Type:        ChangeTypeDirectiveAdded,
				Path:        fmt.Sprintf("directives.%s", name),
				NewValue:    directive,
				Description: fmt.Sprintf("Directive %s was added", name),
			})
		}
	}

	// Check for removed directives
	for name, directive := range from {
		if _, exists := to[name]; !exists {
			changes = append(changes, SchemaChange{
				Type:        ChangeTypeDirectiveRemoved,
				Path:        fmt.Sprintf("directives.%s", name),
				OldValue:    directive,
				Description: fmt.Sprintf("Directive %s was removed", name),
			})
		}
	}

	return changes
}

// fieldsAreDifferent checks if two fields are different
func (svm *SchemaVersionManager) fieldsAreDifferent(field1, field2 *FieldDefinition) bool {
	// Compare basic properties
	if field1.Name != field2.Name ||
		field1.Description != field2.Description ||
		field1.IsDeprecated != field2.IsDeprecated ||
		field1.DeprecationReason != field2.DeprecationReason {
		return true
	}

	// Compare types
	if !svm.typesAreEqual(field1.Type, field2.Type) {
		return true
	}

	// Compare arguments
	if len(field1.Args) != len(field2.Args) {
		return true
	}

	for i, arg1 := range field1.Args {
		if i >= len(field2.Args) {
			return true
		}
		arg2 := field2.Args[i]
		if arg1.Name != arg2.Name || !svm.typesAreEqual(arg1.Type, arg2.Type) {
			return true
		}
	}

	return false
}

// typesAreEqual checks if two types are equal
func (svm *SchemaVersionManager) typesAreEqual(type1, type2 *TypeDefinition) bool {
	if type1 == nil && type2 == nil {
		return true
	}
	if type1 == nil || type2 == nil {
		return false
	}

	return type1.Name == type2.Name && type1.Kind == type2.Kind
}

// determineMigrationType determines the type of migration based on changes
func (svm *SchemaVersionManager) determineMigrationType(changes []SchemaChange) MigrationType {
	hasBreaking := false
	hasAdditions := false
	hasRemovals := false

	for _, change := range changes {
		switch change.Type {
		case ChangeTypeTypeRemoved, ChangeTypeFieldRemoved, ChangeTypeArgumentRemoved:
			hasRemovals = true
			hasBreaking = true
		case ChangeTypeTypeAdded, ChangeTypeFieldAdded, ChangeTypeArgumentAdded:
			hasAdditions = true
		case ChangeTypeFieldModified, ChangeTypeArgumentModified:
			hasBreaking = true
		}
	}

	if hasBreaking {
		return MigrationTypeBreaking
	}
	if hasRemovals {
		return MigrationTypeRemove
	}
	if hasAdditions {
		return MigrationTypeAdd
	}

	return MigrationTypeModify
}

// executeChange executes a single schema change
func (svm *SchemaVersionManager) executeChange(change SchemaChange) error {
	// This is a simplified implementation
	// In a real implementation, you would apply the actual changes to the schema

	svm.logger.Debug("Executing schema change",
		"type", change.Type,
		"path", change.Path,
		"description", change.Description)

	// For now, just log the change
	return nil
}

// GetMigrationHistory returns the migration history for a schema
func (svm *SchemaVersionManager) GetMigrationHistory(schemaID string) []*SchemaMigration {
	// This would return the migration history from persistent storage
	// For now, return an empty slice
	return []*SchemaMigration{}
}

// RollbackMigration rolls back a migration
func (svm *SchemaVersionManager) RollbackMigration(migration *SchemaMigration) error {
	svm.mu.Lock()
	defer svm.mu.Unlock()

	svm.logger.Info("Rolling back schema migration",
		"migrationId", migration.ID,
		"fromVersion", migration.FromVersion,
		"toVersion", migration.ToVersion)

	// Activate the previous version
	if err := svm.ActivateVersion(migration.SchemaID, migration.FromVersion); err != nil {
		return fmt.Errorf("failed to activate previous version: %w", err)
	}

	// Update migration status
	migration.Status = MigrationStatusRollback

	svm.logger.Info("Schema migration rolled back",
		"migrationId", migration.ID)

	return nil
}

// GetVersioningStats returns statistics about schema versioning
func (svm *SchemaVersionManager) GetVersioningStats() map[string]interface{} {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	stats := map[string]interface{}{
		"totalVersions":  len(svm.versions),
		"activeVersions": 0,
	}

	// Count active versions
	for _, version := range svm.versions {
		if version.IsActive {
			stats["activeVersions"] = stats["activeVersions"].(int) + 1
		}
	}

	return stats
}
