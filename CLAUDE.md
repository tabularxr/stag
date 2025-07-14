# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

**Development:**
- `make run` - Run the application locally
- `make build` - Build the binary
- `make deps` - Install dependencies and tidy modules

**Database:**
- `make docker-up-db` - Start only ArangoDB container
- `make docker-up` - Start full stack (ArangoDB + STAG)
- `make docker-down` - Stop all containers

**Testing:**
- `make test` - Run full test suite (unit + integration)
- `make test-unit` - Run unit tests only
- `make test-integration` - Run integration tests only
- `make test-with-db` - Run tests with temporary database

**Code Quality:**
- `make fmt` - Format Go code
- `make lint` - Run golangci-lint
- `make clean` - Clean build artifacts

## Required Environment Variables

Before running locally, set these environment variables:
```bash
export ARANGO_URL=http://localhost:8529
export ARANGO_DATABASE=stag
export ARANGO_USERNAME=root
export ARANGO_PASSWORD=stagpassword
export STAG_PORT=8080
export LOG_LEVEL=info
```

## Architecture Overview

STAG is a Go-based spatial data service with the following key components:

**Core Structure:**
- `cmd/stag/` - Application entry point
- `internal/` - Private application code
- `pkg/` - Public APIs and utilities
- `tests/` - Unit and integration tests

**Key Internal Packages:**
- `internal/config/` - Configuration management using environment variables
- `internal/database/` - ArangoDB connection and migrations
- `internal/server/` - Gin-based HTTP server with handlers and middleware
- `internal/spatial/` - Spatial data repository layer
- `internal/compression/` - Draco mesh compression
- `internal/metrics/` - Prometheus metrics collection

**API Types:**
- Core types defined in `pkg/api/types.go`
- `SpatialEvent` - Primary data structure containing anchors and meshes
- `Anchor` - Spatial positioning data with pose and metadata
- `Mesh` - Compressed geometry data linked to anchors
- WebSocket protocol for real-time updates

**Database:**
- Uses ArangoDB as the primary database
- Spatial indexing for efficient querying
- Graph relationships between spatial entities

**Key Features:**
- Real-time spatial data ingestion via REST API
- WebSocket support for live updates
- Draco compression for efficient mesh storage
- Prometheus metrics for monitoring
- Radius-based spatial queries
- Time-range and session-based filtering

## Development Notes

- Go 1.22+ required
- Uses Gin framework for HTTP routing
- ArangoDB must be running before tests/development
- WebSocket connections require session_id parameter
- Mesh data is compressed using Draco format
- All timestamps are Unix milliseconds (int64)