# EngLog API - Prototype 001

This is the implementation of **PROTOTYPE-001**: Basic REST API Foundation with in-memory storage.

## Overview

A simple Go REST API server that provides the foundation for the EngLog journal management system. This prototype uses in-memory storage (maps/slices) for rapid development and testing without database dependencies.

## Features

✅ **Completed (PROTOTYPE-001)**

- Go HTTP server running on port 8080
- In-memory data structures for journal storage (thread-safe)
- Basic JSON request/response handling
- Graceful server startup and shutdown
- Structured logging with request information
- Basic error handling with appropriate HTTP status codes
- Health check endpoint (`/health`)
- No authentication required
- Data persists only during server runtime

## Quick Start

### Prerequisites

- Go 1.21 or later
- Make (optional, for convenience commands)

### Running the Server

```bash
# Clone and navigate to the project
cd englog

# Run directly
make run
# or
go run ./cmd/api

# Build binary
make build
./bin/englog-api

# Run tests
make test

# Run all checks (format, vet, test)
make check
```

### Testing the API

```bash
# Health check
curl http://localhost:8080/health

# API info
curl http://localhost:8080/

# Test error handling
curl -X POST http://localhost:8080/health  # Should return 405
```

## API Endpoints

| Method | Endpoint  | Description                             |
| ------ | --------- | --------------------------------------- |
| GET    | `/`       | API information and available endpoints |
| GET    | `/health` | Health check with storage status        |

## Architecture

```
cmd/api/           # Main application entry point
├── main.go        # HTTP server setup and configuration
└── main_test.go   # Basic server tests

internal/
├── models/        # Data models
│   └── journal.go # Journal entry structure
├── storage/       # Storage implementations
│   └── memory.go  # In-memory storage with thread safety
└── handlers/      # HTTP handlers
    └── health.go  # Health check endpoint
```

## Technical Specifications

- **Language**: Go
- **Storage**: In-memory (map[string]\*Journal with sync.RWMutex)
- **Logging**: Structured JSON logging (slog)
- **Server**: HTTP server with graceful shutdown
- **Port**: 8080 (configurable via PORT environment variable)
- **Startup Time**: < 5 seconds (typically ~100ms)

## Logging

The server outputs structured JSON logs including:

- Server startup/shutdown events
- HTTP request details (method, path, status, duration, user agent)
- Health check requests
- Error conditions

Example log entry:

```json
{
  "time": "2025-08-04T20:12:18.772539673Z",
  "level": "INFO",
  "msg": "HTTP request processed",
  "method": "GET",
  "path": "/health",
  "status_code": 200,
  "duration_ms": 0,
  "remote_addr": "[::1]:45438",
  "user_agent": "curl/8.5.0"
}
```

## Development

This is a development prototype focusing on establishing the foundation. The in-memory storage is intentionally simple and will be replaced with persistent storage in future prototypes.

### Next Steps (Future Prototypes)

- Add journal CRUD endpoints
- Implement persistent storage (database)
- Add authentication and authorization
- Add input validation and sanitization
- Add comprehensive test coverage
- Add API documentation (OpenAPI/Swagger)

## Task Status

**PROTOTYPE-001**: ✅ **COMPLETED**

All acceptance criteria have been met:

- [x] Go HTTP server starts and responds on port 8080
- [x] In-memory data structures store journal entries
- [x] Basic JSON request/response handling
- [x] Graceful server startup and shutdown
- [x] Structured logging with request information
- [x] Basic error handling with HTTP status codes
- [x] No authentication required
- [x] Data persists only during runtime
