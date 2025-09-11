package stitching

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
)

// SchemaStitcher manages dynamic schema composition and stitching
type SchemaStitcher struct {
	schemas     map[string]*SchemaDefinition
	stitched    *StitchedSchema
	mu          sync.RWMutex
	logger      *slog.Logger
	version     int64
	lastUpdated time.Time
}

// SchemaDefinition represents a GraphQL schema definition
type SchemaDefinition struct {
	ID            string                          `json:"id"`
	Name          string                          `json:"name"`
	Version       string                          `json:"version"`
	Schema        string                          `json:"schema"`
	Types         map[string]*TypeDefinition      `json:"types"`
	Queries       map[string]*FieldDefinition     `json:"queries"`
	Mutations     map[string]*FieldDefinition     `json:"mutations"`
	Subscriptions map[string]*FieldDefinition     `json:"subscriptions"`
	Directives    map[string]*DirectiveDefinition `json:"directives"`
	Metadata      map[string]interface{}          `json:"metadata"`
	CreatedAt     time.Time                       `json:"createdAt"`
	UpdatedAt     time.Time                       `json:"updatedAt"`
}

// StitchedSchema represents the final stitched schema
type StitchedSchema struct {
	Schema        string                          `json:"schema"`
	Types         map[string]*TypeDefinition      `json:"types"`
	Queries       map[string]*FieldDefinition     `json:"queries"`
	Mutations     map[string]*FieldDefinition     `json:"mutations"`
	Subscriptions map[string]*FieldDefinition     `json:"subscriptions"`
	Directives    map[string]*DirectiveDefinition `json:"directives"`
	Version       int64                           `json:"version"`
	CreatedAt     time.Time                       `json:"createdAt"`
	UpdatedAt     time.Time                       `json:"updatedAt"`
	Sources       []string                        `json:"sources"`
}

// TypeDefinition represents a GraphQL type definition
type TypeDefinition struct {
	Name          string                           `json:"name"`
	Kind          TypeKind                         `json:"kind"`
	Description   string                           `json:"description,omitempty"`
	Fields        map[string]*FieldDefinition      `json:"fields,omitempty"`
	Interfaces    []string                         `json:"interfaces,omitempty"`
	PossibleTypes []string                         `json:"possibleTypes,omitempty"`
	EnumValues    []*EnumValueDefinition           `json:"enumValues,omitempty"`
	InputFields   map[string]*InputValueDefinition `json:"inputFields,omitempty"`
	OfType        *TypeDefinition                  `json:"ofType,omitempty"`
	Source        string                           `json:"source,omitempty"`
	Metadata      map[string]interface{}           `json:"metadata,omitempty"`
}

// FieldDefinition represents a GraphQL field definition
type FieldDefinition struct {
	Name              string                  `json:"name"`
	Description       string                  `json:"description,omitempty"`
	Type              *TypeDefinition         `json:"type"`
	Args              []*InputValueDefinition `json:"args,omitempty"`
	IsDeprecated      bool                    `json:"isDeprecated"`
	DeprecationReason string                  `json:"deprecationReason,omitempty"`
	Source            string                  `json:"source,omitempty"`
	Metadata          map[string]interface{}  `json:"metadata,omitempty"`
}

// InputValueDefinition represents a GraphQL input value definition
type InputValueDefinition struct {
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	Type         *TypeDefinition `json:"type"`
	DefaultValue interface{}     `json:"defaultValue,omitempty"`
	Source       string          `json:"source,omitempty"`
}

// EnumValueDefinition represents a GraphQL enum value definition
type EnumValueDefinition struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason,omitempty"`
	Source            string `json:"source,omitempty"`
}

// DirectiveDefinition represents a GraphQL directive definition
type DirectiveDefinition struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Locations   []DirectiveLocation     `json:"locations"`
	Args        []*InputValueDefinition `json:"args,omitempty"`
	Source      string                  `json:"source,omitempty"`
}

// TypeKind represents the kind of GraphQL type
type TypeKind string

const (
	TypeKindScalar      TypeKind = "SCALAR"
	TypeKindObject      TypeKind = "OBJECT"
	TypeKindInterface   TypeKind = "INTERFACE"
	TypeKindUnion       TypeKind = "UNION"
	TypeKindEnum        TypeKind = "ENUM"
	TypeKindInputObject TypeKind = "INPUT_OBJECT"
	TypeKindList        TypeKind = "LIST"
	TypeKindNonNull     TypeKind = "NON_NULL"
)

// DirectiveLocation represents where a directive can be used
type DirectiveLocation string

const (
	DirectiveLocationQuery                DirectiveLocation = "QUERY"
	DirectiveLocationMutation             DirectiveLocation = "MUTATION"
	DirectiveLocationSubscription         DirectiveLocation = "SUBSCRIPTION"
	DirectiveLocationField                DirectiveLocation = "FIELD"
	DirectiveLocationFragmentDefinition   DirectiveLocation = "FRAGMENT_DEFINITION"
	DirectiveLocationFragmentSpread       DirectiveLocation = "FRAGMENT_SPREAD"
	DirectiveLocationInlineFragment       DirectiveLocation = "INLINE_FRAGMENT"
	DirectiveLocationSchema               DirectiveLocation = "SCHEMA"
	DirectiveLocationScalar               DirectiveLocation = "SCALAR"
	DirectiveLocationObject               DirectiveLocation = "OBJECT"
	DirectiveLocationFieldDefinition      DirectiveLocation = "FIELD_DEFINITION"
	DirectiveLocationArgumentDefinition   DirectiveLocation = "ARGUMENT_DEFINITION"
	DirectiveLocationInterface            DirectiveLocation = "INTERFACE"
	DirectiveLocationUnion                DirectiveLocation = "UNION"
	DirectiveLocationEnum                 DirectiveLocation = "ENUM"
	DirectiveLocationEnumValue            DirectiveLocation = "ENUM_VALUE"
	DirectiveLocationInputObject          DirectiveLocation = "INPUT_OBJECT"
	DirectiveLocationInputFieldDefinition DirectiveLocation = "INPUT_FIELD_DEFINITION"
)

// StitchingConfig holds configuration for schema stitching
type StitchingConfig struct {
	ConflictResolution ConflictResolutionStrategy `json:"conflictResolution"`
	TypePrefixing      bool                       `json:"typePrefixing"`
	FieldPrefixing     bool                       `json:"fieldPrefixing"`
	MergeDirectives    bool                       `json:"mergeDirectives"`
	ValidateSchemas    bool                       `json:"validateSchemas"`
	AutoUpdate         bool                       `json:"autoUpdate"`
	UpdateInterval     time.Duration              `json:"updateInterval"`
}

// ConflictResolutionStrategy defines how to resolve schema conflicts
type ConflictResolutionStrategy string

const (
	ConflictResolutionLastWins  ConflictResolutionStrategy = "last_wins"
	ConflictResolutionFirstWins ConflictResolutionStrategy = "first_wins"
	ConflictResolutionMerge     ConflictResolutionStrategy = "merge"
	ConflictResolutionError     ConflictResolutionStrategy = "error"
)

// DefaultStitchingConfig returns the default stitching configuration
func DefaultStitchingConfig() *StitchingConfig {
	return &StitchingConfig{
		ConflictResolution: ConflictResolutionLastWins,
		TypePrefixing:      false,
		FieldPrefixing:     false,
		MergeDirectives:    true,
		ValidateSchemas:    true,
		AutoUpdate:         true,
		UpdateInterval:     1 * time.Minute,
	}
}

// NewSchemaStitcher creates a new schema stitcher
func NewSchemaStitcher(logger *slog.Logger) *SchemaStitcher {
	return &SchemaStitcher{
		schemas:     make(map[string]*SchemaDefinition),
		logger:      logger,
		version:     1,
		lastUpdated: time.Now(),
	}
}

// RegisterSchema registers a new schema for stitching
func (ss *SchemaStitcher) RegisterSchema(schema *SchemaDefinition) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Validate schema
	if err := ss.validateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Set metadata
	schema.CreatedAt = time.Now()
	schema.UpdatedAt = time.Now()

	// Register schema
	ss.schemas[schema.ID] = schema
	ss.version++

	ss.logger.Info("Registered schema",
		"id", schema.ID,
		"name", schema.Name,
		"version", schema.Version)

	// Auto-stitch if enabled
	if ss.shouldAutoStitch() {
		go func() {
			if err := ss.StitchSchemas(); err != nil {
				ss.logger.Error("Failed to auto-stitch schemas after registration", "error", err)
			}
		}()
	}

	return nil
}

// UnregisterSchema unregisters a schema
func (ss *SchemaStitcher) UnregisterSchema(schemaID string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if _, exists := ss.schemas[schemaID]; !exists {
		return fmt.Errorf("schema %s not found", schemaID)
	}

	delete(ss.schemas, schemaID)
	ss.version++

	ss.logger.Info("Unregistered schema", "id", schemaID)

	// Auto-stitch if enabled
	if ss.shouldAutoStitch() {
		go func() {
			if err := ss.StitchSchemas(); err != nil {
				ss.logger.Error("Failed to auto-stitch schemas after unregistration", "error", err)
			}
		}()
	}

	return nil
}

// UpdateSchema updates an existing schema
func (ss *SchemaStitcher) UpdateSchema(schema *SchemaDefinition) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if _, exists := ss.schemas[schema.ID]; !exists {
		return fmt.Errorf("schema %s not found", schema.ID)
	}

	// Validate schema
	if err := ss.validateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Update schema
	schema.UpdatedAt = time.Now()
	ss.schemas[schema.ID] = schema
	ss.version++

	ss.logger.Info("Updated schema",
		"id", schema.ID,
		"name", schema.Name,
		"version", schema.Version)

	// Auto-stitch if enabled
	if ss.shouldAutoStitch() {
		go func() {
			if err := ss.StitchSchemas(); err != nil {
				ss.logger.Error("Failed to auto-stitch schemas after update", "error", err)
			}
		}()
	}

	return nil
}

// StitchSchemas stitches all registered schemas into a single schema
func (ss *SchemaStitcher) StitchSchemas() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.logger.Info("Starting schema stitching", "schemas", len(ss.schemas))

	// Create new stitched schema
	stitched := &StitchedSchema{
		Types:         make(map[string]*TypeDefinition),
		Queries:       make(map[string]*FieldDefinition),
		Mutations:     make(map[string]*FieldDefinition),
		Subscriptions: make(map[string]*FieldDefinition),
		Directives:    make(map[string]*DirectiveDefinition),
		Version:       ss.version,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Sources:       make([]string, 0, len(ss.schemas)),
	}

	// Process each schema
	for _, schema := range ss.schemas {
		stitched.Sources = append(stitched.Sources, schema.ID)

		// Merge types
		if err := ss.mergeTypes(stitched, schema); err != nil {
			return fmt.Errorf("failed to merge types from schema %s: %w", schema.ID, err)
		}

		// Merge queries
		if err := ss.mergeFields(stitched.Queries, schema.Queries, schema.ID); err != nil {
			return fmt.Errorf("failed to merge queries from schema %s: %w", schema.ID, err)
		}

		// Merge mutations
		if err := ss.mergeFields(stitched.Mutations, schema.Mutations, schema.ID); err != nil {
			return fmt.Errorf("failed to merge mutations from schema %s: %w", schema.ID, err)
		}

		// Merge subscriptions
		if err := ss.mergeFields(stitched.Subscriptions, schema.Subscriptions, schema.ID); err != nil {
			return fmt.Errorf("failed to merge subscriptions from schema %s: %w", schema.ID, err)
		}

		// Merge directives
		if err := ss.mergeDirectives(stitched, schema); err != nil {
			return fmt.Errorf("failed to merge directives from schema %s: %w", schema.ID, err)
		}
	}

	// Generate final schema string
	schemaString, err := ss.generateSchemaString(stitched)
	if err != nil {
		return fmt.Errorf("failed to generate schema string: %w", err)
	}
	stitched.Schema = schemaString

	// Update stitched schema
	ss.stitched = stitched
	ss.lastUpdated = time.Now()

	ss.logger.Info("Schema stitching completed",
		"types", len(stitched.Types),
		"queries", len(stitched.Queries),
		"mutations", len(stitched.Mutations),
		"subscriptions", len(stitched.Subscriptions),
		"directives", len(stitched.Directives))

	return nil
}

// GetStitchedSchema returns the current stitched schema
func (ss *SchemaStitcher) GetStitchedSchema() *StitchedSchema {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.stitched
}

// GetSchema returns a specific schema by ID
func (ss *SchemaStitcher) GetSchema(schemaID string) (*SchemaDefinition, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	schema, exists := ss.schemas[schemaID]
	return schema, exists
}

// ListSchemas returns all registered schemas
func (ss *SchemaStitcher) ListSchemas() []*SchemaDefinition {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	schemas := make([]*SchemaDefinition, 0, len(ss.schemas))
	for _, schema := range ss.schemas {
		schemas = append(schemas, schema)
	}

	// Sort by name
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Name < schemas[j].Name
	})

	return schemas
}

// GetVersion returns the current schema version
func (ss *SchemaStitcher) GetVersion() int64 {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.version
}

// validateSchema validates a schema definition
func (ss *SchemaStitcher) validateSchema(schema *SchemaDefinition) error {
	if schema.ID == "" {
		return fmt.Errorf("schema ID is required")
	}
	if schema.Name == "" {
		return fmt.Errorf("schema name is required")
	}
	if schema.Schema == "" {
		return fmt.Errorf("schema string is required")
	}

	// Basic validation - in a real implementation, you would parse and validate the GraphQL schema
	return nil
}

// shouldAutoStitch determines if schemas should be auto-stitched
func (ss *SchemaStitcher) shouldAutoStitch() bool {
	// This would be based on configuration
	return true
}

// mergeTypes merges types from a schema into the stitched schema
func (ss *SchemaStitcher) mergeTypes(stitched *StitchedSchema, schema *SchemaDefinition) error {
	for name, typeDef := range schema.Types {
		// Check for conflicts
		if existing, exists := stitched.Types[name]; exists {
			if err := ss.resolveTypeConflict(stitched, name, existing, typeDef, schema.ID); err != nil {
				return err
			}
		} else {
			// Add new type
			newType := ss.copyTypeDefinition(typeDef)
			newType.Source = schema.ID
			stitched.Types[name] = newType
		}
	}
	return nil
}

// mergeFields merges fields from a schema into the stitched schema
func (ss *SchemaStitcher) mergeFields(target map[string]*FieldDefinition, source map[string]*FieldDefinition, schemaID string) error {
	for name, field := range source {
		// Check for conflicts
		if existing, exists := target[name]; exists {
			if err := ss.resolveFieldConflict(target, name, existing, field, schemaID); err != nil {
				return err
			}
		} else {
			// Add new field
			newField := ss.copyFieldDefinition(field)
			newField.Source = schemaID
			target[name] = newField
		}
	}
	return nil
}

// mergeDirectives merges directives from a schema into the stitched schema
func (ss *SchemaStitcher) mergeDirectives(stitched *StitchedSchema, schema *SchemaDefinition) error {
	for name, directive := range schema.Directives {
		// Check for conflicts
		if existing, exists := stitched.Directives[name]; exists {
			if err := ss.resolveDirectiveConflict(stitched, name, existing, directive, schema.ID); err != nil {
				return err
			}
		} else {
			// Add new directive
			newDirective := ss.copyDirectiveDefinition(directive)
			newDirective.Source = schema.ID
			stitched.Directives[name] = newDirective
		}
	}
	return nil
}

// resolveTypeConflict resolves conflicts between types
func (ss *SchemaStitcher) resolveTypeConflict(stitched *StitchedSchema, name string, existing, new *TypeDefinition, schemaID string) error {
	// This is a simplified conflict resolution
	// In a real implementation, you would have more sophisticated logic

	ss.logger.Warn("Type conflict detected",
		"type", name,
		"existingSource", existing.Source,
		"newSource", schemaID)

	// For now, use the last-wins strategy
	newType := ss.copyTypeDefinition(new)
	newType.Source = schemaID
	stitched.Types[name] = newType

	return nil
}

// resolveFieldConflict resolves conflicts between fields
func (ss *SchemaStitcher) resolveFieldConflict(target map[string]*FieldDefinition, name string, existing, new *FieldDefinition, schemaID string) error {
	ss.logger.Warn("Field conflict detected",
		"field", name,
		"existingSource", existing.Source,
		"newSource", schemaID)

	// For now, use the last-wins strategy
	newField := ss.copyFieldDefinition(new)
	newField.Source = schemaID
	target[name] = newField

	return nil
}

// resolveDirectiveConflict resolves conflicts between directives
func (ss *SchemaStitcher) resolveDirectiveConflict(stitched *StitchedSchema, name string, existing, new *DirectiveDefinition, schemaID string) error {
	ss.logger.Warn("Directive conflict detected",
		"directive", name,
		"existingSource", existing.Source,
		"newSource", schemaID)

	// For now, use the last-wins strategy
	newDirective := ss.copyDirectiveDefinition(new)
	newDirective.Source = schemaID
	stitched.Directives[name] = newDirective

	return nil
}

// copyTypeDefinition creates a deep copy of a type definition
func (ss *SchemaStitcher) copyTypeDefinition(original *TypeDefinition) *TypeDefinition {
	copy := &TypeDefinition{
		Name:        original.Name,
		Kind:        original.Kind,
		Description: original.Description,
		Source:      original.Source,
		Metadata:    make(map[string]interface{}),
	}

	// Copy metadata
	for k, v := range original.Metadata {
		copy.Metadata[k] = v
	}

	// Copy fields
	if original.Fields != nil {
		copy.Fields = make(map[string]*FieldDefinition)
		for name, field := range original.Fields {
			copy.Fields[name] = ss.copyFieldDefinition(field)
		}
	}

	// Copy other fields as needed
	copy.Interfaces = append([]string{}, original.Interfaces...)
	copy.PossibleTypes = append([]string{}, original.PossibleTypes...)
	copy.InputFields = make(map[string]*InputValueDefinition)
	for name, input := range original.InputFields {
		copy.InputFields[name] = ss.copyInputValueDefinition(input)
	}

	if original.OfType != nil {
		copy.OfType = ss.copyTypeDefinition(original.OfType)
	}

	return copy
}

// copyFieldDefinition creates a deep copy of a field definition
func (ss *SchemaStitcher) copyFieldDefinition(original *FieldDefinition) *FieldDefinition {
	copy := &FieldDefinition{
		Name:              original.Name,
		Description:       original.Description,
		IsDeprecated:      original.IsDeprecated,
		DeprecationReason: original.DeprecationReason,
		Source:            original.Source,
		Metadata:          make(map[string]interface{}),
	}

	// Copy metadata
	for k, v := range original.Metadata {
		copy.Metadata[k] = v
	}

	// Copy type
	if original.Type != nil {
		copy.Type = ss.copyTypeDefinition(original.Type)
	}

	// Copy args
	copy.Args = make([]*InputValueDefinition, len(original.Args))
	for i, arg := range original.Args {
		copy.Args[i] = ss.copyInputValueDefinition(arg)
	}

	return copy
}

// copyInputValueDefinition creates a deep copy of an input value definition
func (ss *SchemaStitcher) copyInputValueDefinition(original *InputValueDefinition) *InputValueDefinition {
	copy := &InputValueDefinition{
		Name:         original.Name,
		Description:  original.Description,
		DefaultValue: original.DefaultValue,
		Source:       original.Source,
	}

	if original.Type != nil {
		copy.Type = ss.copyTypeDefinition(original.Type)
	}

	return copy
}

// copyDirectiveDefinition creates a deep copy of a directive definition
func (ss *SchemaStitcher) copyDirectiveDefinition(original *DirectiveDefinition) *DirectiveDefinition {
	copy := &DirectiveDefinition{
		Name:        original.Name,
		Description: original.Description,
		Locations:   append([]DirectiveLocation{}, original.Locations...),
		Source:      original.Source,
	}

	// Copy args
	copy.Args = make([]*InputValueDefinition, len(original.Args))
	for i, arg := range original.Args {
		copy.Args[i] = ss.copyInputValueDefinition(arg)
	}

	return copy
}

// generateSchemaString generates the GraphQL schema string from the stitched schema
func (ss *SchemaStitcher) generateSchemaString(stitched *StitchedSchema) (string, error) {
	var builder strings.Builder

	// Add schema definition
	builder.WriteString("type Query {\n")
	for name, field := range stitched.Queries {
		builder.WriteString(fmt.Sprintf("  %s", name))
		if len(field.Args) > 0 {
			builder.WriteString("(")
			for i, arg := range field.Args {
				if i > 0 {
					builder.WriteString(", ")
				}
				builder.WriteString(fmt.Sprintf("%s: %s", arg.Name, ss.typeToString(arg.Type)))
			}
			builder.WriteString(")")
		}
		builder.WriteString(fmt.Sprintf(": %s\n", ss.typeToString(field.Type)))
	}
	builder.WriteString("}\n\n")

	// Add mutation definition
	if len(stitched.Mutations) > 0 {
		builder.WriteString("type Mutation {\n")
		for name, field := range stitched.Mutations {
			builder.WriteString(fmt.Sprintf("  %s", name))
			if len(field.Args) > 0 {
				builder.WriteString("(")
				for i, arg := range field.Args {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString(fmt.Sprintf("%s: %s", arg.Name, ss.typeToString(arg.Type)))
				}
				builder.WriteString(")")
			}
			builder.WriteString(fmt.Sprintf(": %s\n", ss.typeToString(field.Type)))
		}
		builder.WriteString("}\n\n")
	}

	// Add subscription definition
	if len(stitched.Subscriptions) > 0 {
		builder.WriteString("type Subscription {\n")
		for name, field := range stitched.Subscriptions {
			builder.WriteString(fmt.Sprintf("  %s", name))
			if len(field.Args) > 0 {
				builder.WriteString("(")
				for i, arg := range field.Args {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString(fmt.Sprintf("%s: %s", arg.Name, ss.typeToString(arg.Type)))
				}
				builder.WriteString(")")
			}
			builder.WriteString(fmt.Sprintf(": %s\n", ss.typeToString(field.Type)))
		}
		builder.WriteString("}\n\n")
	}

	// Add type definitions
	for name, typeDef := range stitched.Types {
		if name == "Query" || name == "Mutation" || name == "Subscription" {
			continue // Already handled above
		}

		builder.WriteString(fmt.Sprintf("type %s {\n", name))
		if typeDef.Fields != nil {
			for fieldName, field := range typeDef.Fields {
				builder.WriteString(fmt.Sprintf("  %s", fieldName))
				if len(field.Args) > 0 {
					builder.WriteString("(")
					for i, arg := range field.Args {
						if i > 0 {
							builder.WriteString(", ")
						}
						builder.WriteString(fmt.Sprintf("%s: %s", arg.Name, ss.typeToString(arg.Type)))
					}
					builder.WriteString(")")
				}
				builder.WriteString(fmt.Sprintf(": %s\n", ss.typeToString(field.Type)))
			}
		}
		builder.WriteString("}\n\n")
	}

	return builder.String(), nil
}

// typeToString converts a type definition to a string
func (ss *SchemaStitcher) typeToString(typeDef *TypeDefinition) string {
	if typeDef == nil {
		return "String"
	}

	switch typeDef.Kind {
	case TypeKindNonNull:
		return ss.typeToString(typeDef.OfType) + "!"
	case TypeKindList:
		return "[" + ss.typeToString(typeDef.OfType) + "]"
	default:
		return typeDef.Name
	}
}

// GetStitchingStats returns statistics about schema stitching
func (ss *SchemaStitcher) GetStitchingStats() map[string]interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	stats := map[string]interface{}{
		"totalSchemas": len(ss.schemas),
		"version":      ss.version,
		"lastUpdated":  ss.lastUpdated,
		"hasStitched":  ss.stitched != nil,
	}

	if ss.stitched != nil {
		stats["stitchedTypes"] = len(ss.stitched.Types)
		stats["stitchedQueries"] = len(ss.stitched.Queries)
		stats["stitchedMutations"] = len(ss.stitched.Mutations)
		stats["stitchedSubscriptions"] = len(ss.stitched.Subscriptions)
		stats["stitchedDirectives"] = len(ss.stitched.Directives)
		stats["stitchedSources"] = ss.stitched.Sources
	}

	return stats
}
