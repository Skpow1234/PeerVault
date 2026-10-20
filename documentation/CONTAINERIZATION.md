# Shugo Containerization Guide

This document describes how to run Shugo in containers using Docker and Docker Compose.

## Overview

Shugo supports two containerization approaches:

1. **All-in-one container**: Single container running all nodes (original behavior)
2. **Multi-container deployment**: Separate containers for each node (production-ready)

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 1GB of available memory
- Ports 3000, 5000, and 7000 available (or configurable)

## Quick Start

### Option 1: All-in-One Container (Development)

For development and testing, use the all-in-one container:

```bash
# Build and run
docker-compose -f docker-compose.dev.yml up --build

# Or using the original Dockerfile
docker build -t shugo .
docker run --rm -p 3000:3000 -p 5000:5000 -p 7000:7000 shugo
```

### Option 2: Multi-Container Deployment (Production)

For production-like environments with separate containers:

```bash
# Build and run all services
docker-compose up --build

# Run in background
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

## Multi-Container Architecture

The multi-container setup consists of:

### Services

1. **shugo-node1** (Port 3000)
   - Bootstrap/seed node
   - No dependencies on other nodes
   - Exposes port 3000

2. **shugo-node2** (Port 7000)
   - Bootstrap/seed node
   - No dependencies on other nodes
   - Exposes port 7000

3. **shugo-node3** (Port 5000)
   - Client node that connects to both bootstrap nodes
   - Depends on node1 and node2
   - Exposes port 5000

4. **shugo-demo** (No exposed ports)
   - Demo client that runs store/get operations
   - Connects to the network via node3
   - Runs once and exits

### Network

All containers communicate via a custom bridge network (`shugo-network`).

### Volumes

Each node has its own persistent volume:

- `shugo-node1-data`
- `shugo-node2-data`
- `shugo-node3-data`
- `shugo-demo-data`

## Configuration

### Environment Variables

- `LOG_LEVEL`: Set logging level (debug, info, warn, error)
- Default: `info`

### Command Line Options

#### Node Options (`shugo-node`)

```bash
--listen string        Listen address (default ":3000")
--bootstrap string     Comma-separated list of bootstrap node addresses
--log-level string     Log level (default "info")
--storage-prefix string Storage directory prefix (default "shugo")
```

#### Demo Client Options (`shugo-demo`)

```bash
--target string        Target node to connect to (default "localhost:5000")
--bootstrap string     Comma-separated list of bootstrap node addresses
--log-level string     Log level (default "info")
--iterations int       Number of store/get iterations (default 20)
```

## Examples

### Running Individual Nodes

```bash
# Build the node image
docker build -f Dockerfile.node -t shugo-node .

# Run a bootstrap node
docker run -d --name node1 -p 3000:3000 shugo-node --listen :3000

# Run a client node that connects to the bootstrap node
docker run -d --name node2 -p 5000:5000 shugo-node \
  --listen :5000 \
  --bootstrap node1:3000
```

### Custom Configuration

```bash
# Run with custom log level
docker run shugo-node --log-level debug

# Run with custom storage prefix
docker run shugo-node --storage-prefix mycluster

# Run demo with custom iterations
docker run shugo-demo --iterations 50
```

## Development

### Building Images

```bash
# Build node image
docker build -f Dockerfile.node -t shugo-node .

# Build demo image
docker build -f Dockerfile.demo -t shugo-demo .

# Build all-in-one image
docker build -f Dockerfile -t shugo .
```

### Testing

```bash
# Test multi-container setup
docker-compose up --build
docker-compose logs -f shugo-demo

# Test individual components
docker run --rm shugo-node --help
docker run --rm shugo-demo --help
```

## Troubleshooting

### Common Issues

1. **Port conflicts**

   ```bash
   # Check if ports are in use
   netstat -tulpn | grep :3000
   
   # Use different ports in docker-compose.yml
   ports:
     - "3001:3000"  # Map host port 3001 to container port 3000
   ```

2. **Container networking**

   ```bash
   # Check network connectivity
   docker exec shugo-node1 ping shugo-node2
   
   # Inspect network
   docker network inspect shugo_shugo-network
   ```

3. **Volume permissions**

   ```bash
   # Check volume data
   docker volume ls
   docker volume inspect shugo_shugo-node1-data
   ```

### Logs

```bash
# View all logs
docker-compose logs

# View specific service logs
docker-compose logs shugo-node1

# Follow logs in real-time
docker-compose logs -f shugo-demo
```

### Cleanup

```bash
# Stop and remove containers
docker-compose down

# Remove volumes (WARNING: This deletes all data)
docker-compose down -v

# Remove images
docker rmi shugo-node shugo-demo shugo
```

## Production Considerations

### Security

- Use non-root users in containers
- Implement proper secrets management
- Use private networks for inter-container communication
- Regularly update base images

### Performance

- Monitor resource usage: `docker stats`
- Adjust resource limits in docker-compose.yml
- Use host networking for high-performance scenarios
- Consider using Docker Swarm or Kubernetes for orchestration

### Monitoring

- Implement health checks
- Use logging aggregation (ELK stack, etc.)
- Monitor container metrics
- Set up alerting for failures

### Scaling

- Use load balancers for multiple client nodes
- Implement service discovery
- Consider using Docker Swarm or Kubernetes
- Use external storage for persistence

## File Structure

```bash
.
├── Dockerfile              # All-in-one container
├── Dockerfile.node         # Individual node container
├── Dockerfile.demo         # Demo client container
├── docker-compose.yml      # Multi-container deployment
├── docker-compose.dev.yml  # Development setup
├── cmd/
│   ├── shugo/         # Original all-in-one entrypoint
│   ├── shugo-node/    # Individual node entrypoint
│   └── shugo-demo/    # Demo client entrypoint
└── CONTAINERIZATION.md    # This file
