# PeerVault Docker Setup

This directory contains Docker configurations for running PeerVault services in containers.

## Available Services

### Core Services

- **peervault-node**: The main P2P node service
- **peervault-rest-api**: REST API service (port 8080)
- **peervault-grpc-api**: gRPC API service (port 8082)
- **peervault-graphql-api**: GraphQL API service (port 8081)
- **peervault-demo**: Demo client for testing

## Docker Compose Files

### docker-compose.yml

Contains the original P2P network setup with bootstrap nodes and client nodes.

### docker-compose.apis.yml

Contains all API services plus the P2P network nodes for a complete setup.

## Quick Start

### Run P2P Network Only

```bash
docker-compose -f docker/docker-compose.yml up -d
```

### Run Complete Setup (APIs + P2P Network)

```bash
docker-compose -f docker/docker-compose.apis.yml up -d
```

### Run Individual Services

#### REST API

```bash
docker build -f docker/Dockerfile.rest-api -t peervault-rest-api .
docker run -p 8080:8080 peervault-rest-api
```

#### gRPC API

```bash
docker build -f docker/Dockerfile.grpc-api -t peervault-grpc-api .
docker run -p 8082:8082 peervault-grpc-api
```

#### GraphQL API

```bash
docker build -f docker/Dockerfile.graphql-api -t peervault-graphql-api .
docker run -p 8081:8081 peervault-graphql-api
```

#### P2P Node

```bash
docker build -f docker/Dockerfile.node -t peervault-node .
docker run -p 3000:3000 peervault-node
```

## Service Endpoints

### REST API (Port 8080)

- Health: `GET /health`
- Metrics: `GET /metrics`
- System Info: `GET /system`
- Files: `GET /api/v1/files`
- Peers: `GET /api/v1/peers`
- Documentation: `GET /docs`
- Swagger: `GET /swagger.json`

### gRPC API (Port 8082)

- gRPC endpoints for file and peer operations
- HTTP endpoints for health, metrics, and system info

### GraphQL API (Port 8081)

- GraphQL endpoint: `POST /graphql`
- Playground: `GET /playground`
- Health: `GET /health`
- Metrics: `GET /metrics`

### P2P Network

- Node 1: Port 3000 (Bootstrap)
- Node 2: Port 7000 (Bootstrap)
- Node 3: Port 5000 (Client)

## Environment Variables

All services support the following environment variables:

- `LOG_LEVEL`: Logging level (debug, info, warn, error)

## Volumes

Each service creates persistent volumes for data storage:

- `peervault-rest-data`
- `peervault-grpc-data`
- `peervault-graphql-data`
- `peervault-node1-data`
- `peervault-node2-data`
- `peervault-node3-data`

## Networks

Services communicate over the `peervault-api-network` bridge network.

## Development

For development, use the `docker-compose.dev.yml` file which includes additional development tools and configurations.

## Troubleshooting

1. **Port conflicts**: Ensure the required ports (3000, 5000, 7000, 8080, 8081, 8082) are available
2. **Build issues**: Make sure all dependencies are properly installed and the proto files are generated
3. **Network issues**: Check that the Docker network is properly configured

## Security Notes

- The gRPC API uses a demo authentication token by default
- All services run with minimal privileges in Alpine Linux containers
- TLS certificates are included for secure communication
