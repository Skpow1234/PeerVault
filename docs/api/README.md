# PeerVault REST API Documentation

This directory contains comprehensive documentation for the PeerVault REST API, including OpenAPI/Swagger specifications and usage examples.

## 📁 File Structure

```bash
docs/api/
├── README.md                    # This file - API documentation overview
├── peervault-rest-api.yaml     # OpenAPI 3.0 specification (Swagger)
└── examples/                   # API usage examples (coming soon)
    ├── curl/                   # cURL examples
    ├── javascript/             # JavaScript/Node.js examples
    └── python/                 # Python examples
```

## 🚀 Quick Start

### Interactive Documentation

The easiest way to explore the API is through the interactive Swagger UI:

1. **Start the REST API server via Docker compose.**

2. **Open Swagger UI in your browser:**

   ```bash
   http://localhost:8081/docs
   ```

3. **Explore and test endpoints directly in the browser!**

### Machine-Readable Specification

For programmatic access, the OpenAPI specification is available at:

```bash
http://localhost:8081/swagger.json
```

## 📋 API Overview

The PeerVault REST API provides a comprehensive interface for managing the distributed file storage system:

### 🔧 Core Features

- **File Management**: Upload, download, delete, and manage files with metadata
- **Peer Management**: Discover, monitor, and manage peer nodes in the network
- **System Monitoring**: Real-time metrics, health checks, and system information
- **Webhook Support**: Event-driven notifications for system events

### 🏗️ Architecture

The API follows a clean layered architecture with organized types:

```bash
internal/api/rest/
├── types/            # Consolidated types, entities, DTOs, and mappers
│   ├── entities.go   # Core business entities
│   ├── requests/     # API request DTOs
│   │   ├── file_requests.go
│   │   ├── peer_requests.go
│   │   └── system_requests.go
│   ├── responses/    # API response DTOs
│   │   ├── file_responses.go
│   │   ├── peer_responses.go
│   │   └── system_responses.go
│   └── mappers.go    # Entity-DTO mapping functions
├── endpoints/        # HTTP handlers with proper error handling
├── services/         # Business logic interfaces
└── implementations/  # Service implementations
```

### 🔐 Authentication

API endpoints support optional token-based authentication:

Use an API client and include the `Authorization: Bearer <token>` header.

### ⚡ Rate Limiting

API requests are rate-limited to **100 requests per minute** per client IP address.

### 🌐 CORS Support

All endpoints support Cross-Origin Resource Sharing (CORS) for web applications.

## 📚 API Endpoints

### System Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | Get API information |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | System metrics |
| `GET` | `/docs` | Interactive documentation |
| `GET` | `/swagger.json` | OpenAPI specification |

### File Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/files` | List all files |
| `POST` | `/api/v1/files` | Upload a file |
| `GET` | `/api/v1/files/{key}` | Get file by key |
| `DELETE` | `/api/v1/files/{key}` | Delete a file |
| `PUT` | `/api/v1/files/{key}/metadata` | Update file metadata |

### Peer Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/peers` | List all peers |
| `POST` | `/api/v1/peers` | Add a new peer |
| `GET` | `/api/v1/peers/{peerID}` | Get peer by ID |
| `DELETE` | `/api/v1/peers/{peerID}` | Remove a peer |

### System Information

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/system/info` | Get system information |
| `POST` | `/api/v1/webhook` | Webhook endpoint |

## 🔍 OpenAPI Specification

The complete API specification is available in the `peervault-rest-api.yaml` file, which includes:

### 📖 Detailed Documentation

- **Comprehensive descriptions** for all endpoints
- **Request/response schemas** with examples
- **Error handling** documentation
- **Authentication** and security information
- **Rate limiting** details

### 🏷️ Schema Definitions

The specification includes detailed schemas for:

- **FileResponse**: File information with metadata and replicas
- **PeerResponse**: Peer node information and status
- **SystemInfoResponse**: System statistics and metrics
- **HealthResponse**: Health check status
- **ErrorResponse**: Standardized error format

### 🎯 Operation IDs

Each endpoint has a unique operation ID for easy reference:

- `getApiInfo` - Get API information
- `healthCheck` - Health check
- `getMetrics` - Get system metrics
- `listFiles` - List files
- `uploadFile` - Upload file
- `getFile` - Get file by key
- `deleteFile` - Delete file
- `updateFileMetadata` - Update file metadata
- `listPeers` - List peers
- `addPeer` - Add peer
- `getPeer` - Get peer by ID
- `removePeer` - Remove peer
- `getSystemInfo` - Get system information
- `webhook` - Webhook endpoint

## 🛠️ Usage Examples

### Health Check

Use `GET /health` to check service status.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

### List Files

Use `GET /api/v1/files` to list files.

**Response:**

```json
{
  "files": [
    {
      "key": "file_1234567890",
      "name": "document.pdf",
      "size": 1024,
      "content_type": "application/pdf",
      "hash": "abc123def456...",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "metadata": {
        "owner": "user1",
        "category": "documents"
      },
      "replicas": [
        {
          "peer_id": "peer1",
          "status": "active",
          "created_at": "2024-01-15T10:30:00Z"
        }
      ]
    }
  ],
  "total": 1
}
```

### Upload File

Use `POST /api/v1/files` with multipart form data to upload a file.

### Add Peer

Use `POST /api/v1/peers` with JSON to add a peer.

## 🔧 Development

### Running the API Server

Start the REST API via Docker compose.

### Configuration Options

Configure the API via Docker compose environment variables.

### Testing the API

Integration tests run in CI using containerized workflows.

## 📖 Additional Resources

### Related Documentation

- **[Main README](../../README.md)** - Project overview and setup
- **[GraphQL API](../graphql/README.md)** - GraphQL API documentation
- **[Architecture Documentation](../../documentation/)** - Detailed architecture guides

### Code Examples

- **[Integration Tests](../../tests/integration/rest/)** - Complete API test examples
- **[Service Implementations](../../internal/api/rest/implementations/)** - Business logic examples
- **[DTOs](../../internal/api/rest/dto/)** - Data transfer object definitions

### Tools and Utilities

- **Swagger UI**: Interactive API documentation at `/docs`
- **OpenAPI Spec**: Machine-readable specification at `/swagger.json`
- **Health Check**: System health monitoring at `/health`
- **Metrics**: Performance monitoring at `/metrics`

## 🤝 Contributing

When contributing to the API:

1. **Update the OpenAPI specification** in `peervault-rest-api.yaml`
2. **Add comprehensive examples** in the `examples/` directory
3. **Update this README** with new endpoint documentation
4. **Write integration tests** for new endpoints
5. **Follow the layered architecture** pattern

## 📄 License

This API documentation is part of the PeerVault project and is licensed under the MIT License.

---

For questions or support, please refer to the [main project documentation](../../README.md) or open an issue on GitHub.
