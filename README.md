# STAG - Spatial Tracking and Georeferencing Service

STAG is a high-performance spatial data ingestion and query service built in Go. It provides real-time spatial tracking capabilities with support for anchors, meshes, and WebSocket connectivity for live updates.

## Features

- **Spatial Data Ingestion**: Ingest spatial events with anchors and mesh data
- **Real-time Updates**: WebSocket support for live spatial data streaming
- **Compressed Storage**: Draco compression for efficient mesh storage
- **Spatial Querying**: Query anchors by session, radius, and time range
- **Metrics & Monitoring**: Prometheus metrics integration
- **Health Checks**: Built-in health endpoints for monitoring
- **ArangoDB Backend**: Graph database for spatial relationships

## Architecture

STAG is built with a modular architecture:

```
├── cmd/stag/           # Main application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── database/       # ArangoDB connection and migrations
│   ├── server/         # HTTP server and routing
│   │   ├── handlers/   # API endpoints
│   │   └── middleware/ # HTTP middleware
│   ├── spatial/        # Spatial data repository
│   ├── compression/    # Draco mesh compression
│   └── metrics/        # Prometheus metrics
├── pkg/
│   ├── api/           # API types and structures
│   ├── logger/        # Logging utilities
│   └── errors/        # Error handling
└── tests/
    ├── unit/          # Unit tests
    └── integration/   # Integration tests
```

## Quick Start

### Prerequisites

- Go 1.22 or higher
- ArangoDB 3.11+
- Docker and Docker Compose (for containerized deployment)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd stag
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Start ArangoDB**
   ```bash
   make docker-up-db
   ```

4. **Set environment variables**
   ```bash
   export ARANGO_URL=http://localhost:8529
   export ARANGO_DATABASE=stag
   export ARANGO_USERNAME=root
   export ARANGO_PASSWORD=stagpassword
   export STAG_PORT=8080
   export LOG_LEVEL=info
   ```

5. **Run the application**
   ```bash
   make run
   ```

The server will start on port 8080 by default.

### Docker Deployment

1. **Build and run with Docker Compose**
   ```bash
   make docker-up
   ```

This will start both ArangoDB and STAG in containers with proper networking.

## API Endpoints

### Core Endpoints

- `POST /api/v1/ingest` - Ingest spatial events
- `GET /api/v1/query` - Query spatial data
- `GET /api/v1/anchors/{id}` - Get specific anchor
- `GET /api/v1/health` - Health check
- `GET /api/v1/ws` - WebSocket connection

### Example Usage

#### Ingest Spatial Data

```bash
curl -X POST http://localhost:8080/api/v1/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "session-123",
    "event_id": "event-456",
    "timestamp": 1640995200000,
    "anchors": [{
      "id": "anchor-789",
      "session_id": "session-123",
      "pose": {
        "x": 10.5,
        "y": 20.7,
        "z": 30.9,
        "rotation": [0.1, 0.2, 0.3, 0.9]
      },
      "timestamp": 1640995200000
    }],
    "meshes": [{
      "id": "mesh-101",
      "anchor_id": "anchor-789",
      "vertices": "encoded-vertex-data",
      "faces": "encoded-face-data",
      "compression_level": 7,
      "timestamp": 1640995200000
    }]
  }'
```

#### Query Spatial Data

```bash
# Query by session
curl "http://localhost:8080/api/v1/query?session_id=session-123&include_meshes=true"

# Query by radius
curl "http://localhost:8080/api/v1/query?anchor_id=anchor-789&radius=50.0"

# Query with time range
curl "http://localhost:8080/api/v1/query?session_id=session-123&since=1640995200000&limit=100"
```

#### WebSocket Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?session_id=session-123');

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

ws.send(JSON.stringify({
  type: 'ping',
  timestamp: Date.now()
}));
```

## Configuration

STAG can be configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `STAG_PORT` | `8080` | Server port |
| `STAG_HOST` | `localhost` | Server host |
| `ARANGO_URL` | `http://localhost:8529` | ArangoDB connection URL |
| `ARANGO_DATABASE` | `stag` | ArangoDB database name |
| `ARANGO_USERNAME` | `root` | ArangoDB username |
| `ARANGO_PASSWORD` | *required* | ArangoDB password |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

## Testing

### Unit Tests

```bash
make test-unit
```

### Integration Tests

Integration tests require a running ArangoDB instance:

```bash
make test-integration
```

### Full Test Suite

```bash
make test
```

### Test with Database

Run tests with a temporary ArangoDB container:

```bash
make test-with-db
```

## Development

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean
```

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build
```

## Data Models

### Spatial Event

```go
type SpatialEvent struct {
    SessionID string    `json:"session_id"`
    EventID   string    `json:"event_id"`
    Timestamp int64     `json:"timestamp"`
    Anchors   []Anchor  `json:"anchors"`
    Meshes    []Mesh    `json:"meshes"`
}
```

### Anchor

```go
type Anchor struct {
    ID        string    `json:"id"`
    SessionID string    `json:"session_id"`
    Pose      Pose      `json:"pose"`
    Timestamp int64     `json:"timestamp"`
    Metadata  Metadata  `json:"metadata,omitempty"`
}
```

### Mesh

```go
type Mesh struct {
    ID               string `json:"id"`
    AnchorID         string `json:"anchor_id"`
    Vertices         []byte `json:"vertices"`
    Faces            []byte `json:"faces"`
    IsDelta          bool   `json:"is_delta"`
    BaseMeshID       string `json:"base_mesh_id,omitempty"`
    CompressionLevel int    `json:"compression_level"`
    Timestamp        int64  `json:"timestamp"`
}
```

## Monitoring

STAG includes Prometheus metrics at `/metrics` endpoint:

- HTTP request duration and count
- Database operation metrics
- WebSocket connection metrics
- Compression performance metrics

## WebSocket Protocol

### Message Types

- `anchor_update` - New anchor data
- `mesh_update` - New mesh data
- `ping/pong` - Keep-alive messages
- `error` - Error notifications

### Example Message

```json
{
  "type": "anchor_update",
  "session_id": "session-123",
  "data": {
    "id": "anchor-789",
    "pose": {
      "x": 10.5,
      "y": 20.7,
      "z": 30.9,
      "rotation": [0.1, 0.2, 0.3, 0.9]
    }
  },
  "timestamp": 1640995200000
}
```

## Performance Considerations

- **Compression**: Meshes are compressed using Draco for efficient storage
- **Database**: ArangoDB provides efficient spatial indexing
- **Caching**: Query results can be cached for better performance
- **Batch Processing**: Multiple anchors/meshes can be ingested in a single request

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Verify ArangoDB is running
   - Check connection credentials
   - Ensure database exists

2. **WebSocket Connection Issues**
   - Check firewall settings
   - Verify session ID parameter
   - Monitor connection logs

3. **High Memory Usage**
   - Adjust compression levels
   - Monitor mesh data sizes
   - Consider data retention policies

### Logging

Enable debug logging for detailed information:

```bash
export LOG_LEVEL=debug
```

## License

This project is licensed under the terms specified in the LICENSE file.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Support

For issues and questions, please refer to the project's issue tracker or documentation.