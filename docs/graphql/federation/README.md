# GraphQL Federation

This document describes the GraphQL federation implementation in PeerVault, which enables multi-service composition and distributed GraphQL schemas.

## Overview

GraphQL Federation allows multiple GraphQL services to be composed into a single, unified GraphQL API. This enables:

- **Service Composition**: Multiple services can contribute to a single GraphQL schema
- **Distributed Development**: Teams can develop and deploy services independently
- **Schema Stitching**: Automatic combination of schemas from different services
- **Service Discovery**: Automatic discovery and health checking of federated services

## Architecture

```bash
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Service A     │    │   Service B     │    │   Service C     │
│  (Files)        │    │  (Nodes)        │    │  (Analytics)    │
│                 │    │                 │    │                 │
│ GraphQL Schema  │    │ GraphQL Schema  │    │ GraphQL Schema  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │   Federation Gateway      │
                    │                           │
                    │  - Service Discovery      │
                    │  - Health Checking        │
                    │  - Query Planning         │
                    │  - Schema Composition     │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │      Client               │
                    │                           │
                    │  Single GraphQL Endpoint  │
                    └───────────────────────────┘
```

## Federation Gateway

The federation gateway is the central component that:

1. **Discovers Services**: Automatically discovers and registers GraphQL services
2. **Health Monitoring**: Continuously monitors service health
3. **Query Planning**: Determines which services are needed for each query
4. **Schema Composition**: Combines schemas from all services
5. **Request Routing**: Routes queries to appropriate services

### Starting the Gateway

```bash
# Start with default configuration
peervault-federation

# Start on custom port
peervault-federation -port 9090

# Start with verbose logging
peervault-federation -verbose

# Start with custom health check interval
peervault-federation -health-check-interval 1m
```

### Configuration

The federation gateway can be configured with the following options:

```go
config := &federation.FederationConfig{
    GatewayPort:         8081,           // Port for the gateway
    ServiceTimeout:      30 * time.Second, // Timeout for service requests
    HealthCheckInterval: 30 * time.Second, // Health check interval
    EnableHealthChecks:  true,           // Enable health checks
}
```

## Service Registration

### Automatic Registration

Services can be automatically registered by the gateway on startup:

```go
// Register a service
service := &federation.FederatedService{
    Name:     "peervault-files",
    URL:      "http://localhost:8080/graphql",
    HealthCheck: "http://localhost:8080/health",
    Capabilities: map[string]bool{
        "files":   true,
        "storage": true,
    },
    Metadata: map[string]string{
        "version": "1.0.0",
        "region":  "us-east-1",
    },
}

gateway.RegisterService(service)
```

### Manual Registration via API

Services can also be registered via the REST API:

```bash
# Register a new service
curl -X POST http://localhost:8081/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "peervault-analytics",
    "url": "http://localhost:8082/graphql",
    "healthCheck": "http://localhost:8082/health",
    "capabilities": {
      "analytics": true,
      "metrics": true
    },
    "metadata": {
      "version": "1.0.0",
      "region": "us-east-1"
    }
  }'
```

## API Endpoints

### GraphQL Endpoint

**POST** `/graphql`

The main GraphQL endpoint that executes queries across all federated services.

```graphql
query {
  # This query might span multiple services
  files {
    id
    key
    size
    owner {
      id
      address
      status
    }
  }
  
  systemMetrics {
    storage {
      totalSpace
      usedSpace
    }
    performance {
      averageResponseTime
      cpuUsage
    }
  }
}
```

### Service Management

**GET** `/services`

List all registered services.

```json
[
  {
    "name": "peervault-main",
    "url": "http://localhost:8080/graphql",
    "isHealthy": true,
    "lastSeen": "2024-01-15T10:30:00Z",
    "capabilities": {
      "files": true,
      "nodes": true,
      "storage": true
    }
  }
]
```

**POST** `/services`

Register a new service.

**GET** `/services/{name}`

Get details of a specific service.

**DELETE** `/services/{name}`

Unregister a service.

### Schema and Health

**GET** `/schema`

Get the combined federated schema.

**GET** `/health`

Health check for the federation gateway.

**GET** `/metrics`

Get metrics for all services.

## Service Development

### Creating a Federated Service

To create a service that can be federated:

1. **Implement GraphQL Schema**: Define your service's GraphQL schema
2. **Add Health Endpoint**: Implement a `/health` endpoint
3. **Register with Gateway**: Register your service with the federation gateway

Example service:

```go
package main

import (
    "net/http"
    "github.com/graphql-go/graphql"
)

func main() {
    // Define your GraphQL schema
    schema, _ := graphql.NewSchema(graphql.SchemaConfig{
        Query: graphql.NewObject(graphql.ObjectConfig{
            Name: "Query",
            Fields: graphql.Fields{
                "files": &graphql.Field{
                    Type: graphql.NewList(fileType),
                    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                        // Your resolver logic
                        return getFiles(), nil
                    },
                },
            },
        }),
    })

    // Setup HTTP handlers
    http.HandleFunc("/graphql", graphqlHandler(schema))
    http.HandleFunc("/health", healthHandler)
    
    // Start server
    http.ListenAndServe(":8080", nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"healthy"}`))
}
```

### Service Capabilities

Services can declare their capabilities to help with query planning:

```go
capabilities := map[string]bool{
    "files":      true,  // Can handle file operations
    "storage":    true,  // Can handle storage operations
    "analytics":  true,  // Can handle analytics queries
    "metrics":    true,  // Can provide metrics
    "subscriptions": true, // Supports real-time subscriptions
}
```

## Query Planning

The federation gateway automatically determines which services are needed for each query:

1. **Parse Query**: Analyze the GraphQL query to identify required fields
2. **Map to Services**: Determine which services can provide the required fields
3. **Plan Execution**: Create an execution plan across multiple services
4. **Execute**: Execute the plan and combine results

### Example Query Planning

```graphql
query {
  files {           # Requires 'files' service
    id
    key
    owner {         # Requires 'nodes' service
      id
      address
    }
  }
  systemMetrics {   # Requires 'analytics' service
    storage {
      totalSpace
    }
  }
}
```

This query would be planned across three services:

- `peervault-files` for file data
- `peervault-nodes` for node/owner data  
- `peervault-analytics` for system metrics

## Health Monitoring

The federation gateway continuously monitors service health:

- **Health Checks**: Regular HTTP requests to service health endpoints
- **Service Status**: Track healthy/unhealthy status for each service
- **Automatic Recovery**: Automatically retry failed services
- **Metrics**: Collect health metrics and statistics

### Health Check Configuration

```go
config := &federation.FederationConfig{
    EnableHealthChecks:  true,
    HealthCheckInterval: 30 * time.Second,
    ServiceTimeout:      10 * time.Second,
}
```

## Error Handling

The federation gateway handles errors gracefully:

- **Service Unavailable**: Return partial results if some services are down
- **Timeout Handling**: Implement timeouts for service requests
- **Error Aggregation**: Combine errors from multiple services
- **Fallback Strategies**: Implement fallback mechanisms for critical services

## Performance Considerations

- **Connection Pooling**: Reuse HTTP connections to services
- **Caching**: Cache service schemas and health status
- **Load Balancing**: Distribute load across multiple service instances
- **Query Optimization**: Optimize queries to minimize service calls

## Security

- **Service Authentication**: Authenticate requests to federated services
- **CORS Configuration**: Configure CORS for cross-origin requests
- **Rate Limiting**: Implement rate limiting for service requests
- **TLS**: Use TLS for secure communication between services

## Monitoring and Observability

The federation gateway provides comprehensive monitoring:

- **Service Metrics**: Track requests, errors, and response times per service
- **Health Status**: Monitor service health and availability
- **Query Metrics**: Track query performance and execution times
- **Error Tracking**: Monitor and alert on service errors

Access metrics at: `GET /metrics`

## Best Practices

1. **Service Design**: Design services with clear boundaries and responsibilities
2. **Schema Design**: Use consistent naming and types across services
3. **Error Handling**: Implement robust error handling and fallback mechanisms
4. **Monitoring**: Set up comprehensive monitoring and alerting
5. **Testing**: Test federation scenarios thoroughly
6. **Documentation**: Document service capabilities and schemas
7. **Versioning**: Implement proper versioning for service schemas
