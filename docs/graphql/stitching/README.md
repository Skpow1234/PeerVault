# GraphQL Schema Stitching

This document describes the dynamic schema stitching implementation in PeerVault, which enables runtime schema composition and versioning.

## Overview

Schema stitching allows you to:

- **Dynamic Composition**: Combine multiple GraphQL schemas at runtime
- **Runtime Updates**: Add, remove, or modify schemas without restarting
- **Version Management**: Track and manage schema versions
- **Migration Support**: Migrate between schema versions safely
- **Conflict Resolution**: Handle conflicts between schemas automatically

## Architecture

```bash
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Schema A      │    │   Schema B      │    │   Schema C      │
│  (Files)        │    │  (Users)        │    │  (Analytics)    │
│                 │    │                 │    │                 │
│ GraphQL Schema  │    │ GraphQL Schema  │    │ GraphQL Schema  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │   Schema Stitcher        │
                    │                           │
                    │  - Schema Registration    │
                    │  - Conflict Resolution    │
                    │  - Dynamic Composition    │
                    │  - Version Management     │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │   Stitched Schema        │
                    │                           │
                    │  Combined GraphQL API     │
                    └───────────────────────────┘
```

## Basic Usage

### Schema Registration

```go
// Create schema stitcher
stitcher := stitching.NewSchemaStitcher(logger)

// Register a schema
schema := &stitching.SchemaDefinition{
    ID:     "files-service",
    Name:   "Files Service",
    Version: "1.0.0",
    Schema: `
        type Query {
            files: [File!]!
            file(id: ID!): File
        }
        
        type File {
            id: ID!
            name: String!
            size: Int!
        }
    `,
    Types: map[string]*stitching.TypeDefinition{
        "File": {
            Name: "File",
            Kind: stitching.TypeKindObject,
            Fields: map[string]*stitching.FieldDefinition{
                "id": {
                    Name: "id",
                    Type: &stitching.TypeDefinition{
                        Name: "ID",
                        Kind: stitching.TypeKindNonNull,
                        OfType: &stitching.TypeDefinition{
                            Name: "ID",
                            Kind: stitching.TypeKindScalar,
                        },
                    },
                },
                "name": {
                    Name: "name",
                    Type: &stitching.TypeDefinition{
                        Name: "String",
                        Kind: stitching.TypeKindNonNull,
                        OfType: &stitching.TypeDefinition{
                            Name: "String",
                            Kind: stitching.TypeKindScalar,
                        },
                    },
                },
            },
        },
    },
    Queries: map[string]*stitching.FieldDefinition{
        "files": {
            Name: "files",
            Type: &stitching.TypeDefinition{
                Name: "File",
                Kind: stitching.TypeKindList,
                OfType: &stitching.TypeDefinition{
                    Name: "File",
                    Kind: stitching.TypeKindNonNull,
                    OfType: &stitching.TypeDefinition{
                        Name: "File",
                        Kind: stitching.TypeKindObject,
                    },
                },
            },
        },
    },
}

err := stitcher.RegisterSchema(schema)
```

### Schema Stitching

```go
// Stitch all registered schemas
err := stitcher.StitchSchemas()
if err != nil {
    log.Fatal("Failed to stitch schemas:", err)
}

// Get the stitched schema
stitched := stitcher.GetStitchedSchema()
fmt.Println("Stitched Schema:")
fmt.Println(stitched.Schema)
```

### Schema Updates

```go
// Update an existing schema
updatedSchema := &stitching.SchemaDefinition{
    ID:     "files-service",
    Name:   "Files Service",
    Version: "1.1.0",
    Schema: `
        type Query {
            files: [File!]!
            file(id: ID!): File
            fileCount: Int!
        }
        
        type File {
            id: ID!
            name: String!
            size: Int!
            createdAt: String!
        }
    `,
    // ... updated type definitions
}

err := stitcher.UpdateSchema(updatedSchema)
```

## Schema Versioning

### Version Management

```go
// Create version manager
versionManager := stitching.NewSchemaVersionManager(stitcher, logger)

// Create a new version
version, err := versionManager.CreateVersion(
    "files-service",
    "1.1.0",
    "Added fileCount query and createdAt field",
    "developer@example.com",
)

// Activate a version
err = versionManager.ActivateVersion("files-service", "1.1.0")
```

### Schema Migration

```go
// Create a migration
migration, err := versionManager.CreateMigration(
    "1.0.0",
    "1.1.0",
    "files-service",
    "developer@example.com",
)

// Execute the migration
err = versionManager.ExecuteMigration(migration)
```

### Migration Types

The system automatically detects different types of migrations:

- **Add Migration**: New types or fields added
- **Remove Migration**: Types or fields removed
- **Modify Migration**: Existing types or fields modified
- **Breaking Migration**: Changes that break existing clients

## Conflict Resolution

### Configuration

```go
config := &stitching.StitchingConfig{
    ConflictResolution: stitching.ConflictResolutionLastWins,
    TypePrefixing:      false,
    FieldPrefixing:     false,
    MergeDirectives:    true,
    ValidateSchemas:    true,
    AutoUpdate:         true,
    UpdateInterval:     1 * time.Minute,
}
```

### Resolution Strategies

#### Last Wins

```go
config := &stitching.StitchingConfig{
    ConflictResolution: stitching.ConflictResolutionLastWins,
}
```

#### First Wins

```go
config := &stitching.StitchingConfig{
    ConflictResolution: stitching.ConflictResolutionFirstWins,
}
```

#### Merge

```go
config := &stitching.StitchingConfig{
    ConflictResolution: stitching.ConflictResolutionMerge,
}
```

#### Error on Conflict

```go
config := &stitching.StitchingConfig{
    ConflictResolution: stitching.ConflictResolutionError,
}
```

## Advanced Features

### Schema Validation

```go
// Validate schema before registration
err := stitcher.ValidateSchema(schema)
if err != nil {
    log.Printf("Schema validation failed: %v", err)
}
```

### Schema Introspection

```go
// Get schema statistics
stats := stitcher.GetStitchingStats()
fmt.Printf("Total schemas: %d\n", stats["totalSchemas"])
fmt.Printf("Stitched types: %d\n", stats["stitchedTypes"])
fmt.Printf("Stitched queries: %d\n", stats["stitchedQueries"])
```

### Schema Monitoring

```go
// Monitor schema changes
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        stats := stitcher.GetStitchingStats()
        if stats["hasStitched"].(bool) {
            fmt.Printf("Schema version: %d\n", stats["version"])
        }
    }
}()
```

## Schema Patterns

### Microservice Schema

```go
// Each microservice provides its own schema
userSchema := &stitching.SchemaDefinition{
    ID:     "user-service",
    Name:   "User Service",
    Schema: `
        type Query {
            users: [User!]!
            user(id: ID!): User
        }
        
        type User {
            id: ID!
            name: String!
            email: String!
        }
    `,
}

fileSchema := &stitching.SchemaDefinition{
    ID:     "file-service",
    Name:   "File Service",
    Schema: `
        type Query {
            files: [File!]!
            file(id: ID!): File
        }
        
        type File {
            id: ID!
            name: String!
            owner: User!
        }
    `,
}
```

### Layered Schema

```go
// Base schema with common types
baseSchema := &stitching.SchemaDefinition{
    ID:     "base-schema",
    Name:   "Base Schema",
    Schema: `
        scalar DateTime
        
        type Query {
            health: String!
        }
    `,
}

// Feature-specific schemas
featureSchema := &stitching.SchemaDefinition{
    ID:     "feature-schema",
    Name:   "Feature Schema",
    Schema: `
        type Query {
            features: [Feature!]!
        }
        
        type Feature {
            id: ID!
            name: String!
            enabled: Boolean!
            createdAt: DateTime!
        }
    `,
}
```

### Plugin Schema

```go
// Core schema
coreSchema := &stitching.SchemaDefinition{
    ID:     "core-schema",
    Name:   "Core Schema",
    Schema: `
        type Query {
            system: System!
        }
        
        type System {
            version: String!
            plugins: [Plugin!]!
        }
    `,
}

// Plugin schemas are dynamically registered
pluginSchema := &stitching.SchemaDefinition{
    ID:     "analytics-plugin",
    Name:   "Analytics Plugin",
    Schema: `
        extend type System {
            analytics: Analytics!
        }
        
        type Analytics {
            metrics: [Metric!]!
        }
    `,
}
```

## Best Practices

### 1. Schema Design

- **Clear Boundaries**: Design schemas with clear service boundaries
- **Consistent Naming**: Use consistent naming conventions across schemas
- **Type Safety**: Ensure type compatibility between schemas
- **Documentation**: Document schema changes and migrations

### 2. Version Management

```go
// Use semantic versioning
version := "1.2.3" // major.minor.patch

// Document changes
description := "Added pagination support to files query"

// Create versions for all changes
version, err := versionManager.CreateVersion(
    schemaID,
    version,
    description,
    author,
)
```

### 3. Migration Strategy

```go
// Plan migrations carefully
migration, err := versionManager.CreateMigration(
    "1.0.0",
    "1.1.0",
    schemaID,
    author,
)

// Test migrations in staging
if err := versionManager.ExecuteMigration(migration); err != nil {
    // Rollback on failure
    versionManager.RollbackMigration(migration)
}
```

### 4. Conflict Resolution

```go
// Use appropriate conflict resolution strategy
config := &stitching.StitchingConfig{
    ConflictResolution: stitching.ConflictResolutionLastWins,
    ValidateSchemas:    true,
    AutoUpdate:         false, // Manual control for production
}
```

### 5. Monitoring

```go
// Monitor schema health
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        stats := stitcher.GetStitchingStats()
        if !stats["hasStitched"].(bool) {
            logger.Error("Schema stitching failed")
        }
    }
}()
```

## API Endpoints

### Schema Management

- `GET /schemas` - List all registered schemas
- `POST /schemas` - Register a new schema
- `GET /schemas/{id}` - Get schema details
- `PUT /schemas/{id}` - Update a schema
- `DELETE /schemas/{id}` - Unregister a schema

### Schema Stitching API

- `POST /stitch` - Stitch all schemas
- `GET /stitched` - Get stitched schema
- `GET /stitched/schema` - Get stitched schema as GraphQL SDL

### Version Management API

- `GET /schemas/{id}/versions` - List schema versions
- `POST /schemas/{id}/versions` - Create new version
- `PUT /schemas/{id}/versions/{version}/activate` - Activate version
- `GET /schemas/{id}/versions/{version}` - Get version details

### Migration Management

- `GET /schemas/{id}/migrations` - List migrations
- `POST /schemas/{id}/migrations` - Create migration
- `POST /migrations/{id}/execute` - Execute migration
- `POST /migrations/{id}/rollback` - Rollback migration

## Troubleshooting

### Common Issues

1. **Schema Conflicts**
   - Check conflict resolution strategy
   - Review schema naming conventions
   - Use type prefixing if needed

2. **Migration Failures**
   - Validate schemas before migration
   - Test migrations in staging
   - Implement rollback procedures

3. **Performance Issues**
   - Monitor schema complexity
   - Use schema caching
   - Optimize query resolution

### Debugging

```go
// Enable debug logging
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Get detailed stitching stats
stats := stitcher.GetStitchingStats()
fmt.Printf("Stitching Stats: %+v\n", stats)

// Check version status
versionStats := versionManager.GetVersioningStats()
fmt.Printf("Version Stats: %+v\n", versionStats)
```

## Performance Considerations

- **Schema Complexity**: Keep schemas focused and not overly complex
- **Caching**: Cache stitched schemas to avoid recomputation
- **Lazy Loading**: Load schemas on demand when possible
- **Validation**: Validate schemas efficiently
- **Memory Usage**: Monitor memory usage for large schemas

## Security

- **Schema Validation**: Validate all incoming schemas
- **Access Control**: Control who can register/update schemas
- **Audit Logging**: Log all schema changes
- **Version Control**: Track all schema modifications
