# Docker Setup for EngLog

This document explains how to use the Docker configuration for EngLog development and deployment.

## Overview

The Docker setup provides:

- **Dockerfile**: Builds the Go application in a secure, multi-stage container
- **docker-compose.yml**: Production-ready orchestration with Ollama
- **docker-compose.dev.yml**: Development mode with hot-reload
- **docker-setup.sh**: Automated setup script

## Quick Start

### 1. Automated Setup (Recommended)

```bash
# Production mode
./scripts/docker-setup.sh

# Development mode with hot-reload
./scripts/docker-setup.sh --dev

# Custom model
./scripts/docker-setup.sh --model llama3.2

# Skip automatic model download
./scripts/docker-setup.sh --skip-model
```

### 2. Manual Setup

#### Production Mode

```bash
# Start services
docker-compose up -d

# Check logs
docker-compose logs -f

# Stop services
docker-compose down
```

#### Development Mode

```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop services
docker-compose -f docker-compose.dev.yml down
```

## Services

### API Service

- **Port**: 8080
- **Health Check**: http://localhost:8080/health
- **Environment Variables**:
  - `PORT`: Server port (default: 8080)
  - `OLLAMA_SERVER_URL`: Ollama connection URL
  - `OLLAMA_MODEL_NAME`: Model to use for AI processing
  - `LOG_LEVEL`: Logging verbosity (debug, info, warn, error)
  - `LOG_FORMAT`: Log format (text, json)

### Ollama Service

- **Port**: 11434
- **API Endpoint**: http://localhost:11434/api/tags
- **Model Storage**: Persistent volume `ollama_data`
- **Health Check**: Automatic model availability verification

## Configuration

### Environment Variables

Create a `.env` file to customize configuration:

```env
# API Configuration
PORT=8080
LOG_LEVEL=info
LOG_FORMAT=json

# Ollama Configuration
OLLAMA_MODEL_NAME=deepseek-r1:1.5b
OLLAMA_SERVER_URL=http://ollama:11434

# Development Configuration (dev mode only)
ENABLE_HOT_RELOAD=true
```

### Custom Models

To use different Ollama models:

```bash
# Download model in running container
docker exec englog-ollama ollama pull llama3.2

# Or specify during setup
./scripts/docker-setup.sh --model llama3.2
```

Available models:

- `deepseek-r1:1.5b` (default, fast)
- `llama3.2` (medium quality/speed)
- `mistral` (good for general text)
- `codegemma` (optimized for code)

## Development Features

### Hot Reload (Development Mode)

In development mode, the API service automatically reloads when source code changes:

```bash
# Start development environment
./scripts/docker-setup.sh --dev

# Make changes to Go files
# The API will automatically restart
```

### Debugging

#### View Container Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f ollama
```

#### Execute Commands in Container

```bash
# API container
docker exec -it englog-api sh

# Ollama container
docker exec -it englog-ollama sh

# Check models
docker exec englog-ollama ollama list
```

#### Health Checks

```bash
# API health
curl http://localhost:8080/health

# Ollama health
curl http://localhost:11434/api/tags

# Test API functionality
curl -X POST http://localhost:8080/journals \
  -H "Content-Type: application/json" \
  -d '{"content": "Docker test entry"}'
```

## Production Considerations

### Security

The Docker setup follows security best practices:

- Non-root user in containers
- Read-only root filesystem where possible
- Minimal attack surface with Alpine Linux
- Health checks for service reliability

### Performance

- Multi-stage builds minimize image size
- Build caching speeds up subsequent builds
- Volume caching for Go modules in dev mode
- Proper resource limits can be added

### Networking

Services communicate through Docker networks:

- `englog-network` (production)
- `englog-dev-network` (development)

External access:

- API: localhost:8080
- Ollama: localhost:11434

### Persistence

- `ollama_data`: Ollama models and configuration
- `go_mod_cache`: Go module cache (dev mode)

## Troubleshooting

### Common Issues

#### Ollama Model Download Fails

```bash
# Check Ollama logs
docker-compose logs ollama

# Manually download model
docker exec englog-ollama ollama pull deepseek-r1:1.5b
```

#### API Cannot Connect to Ollama

```bash
# Verify Ollama is running
docker-compose ps

# Check network connectivity
docker exec englog-api wget -O- http://ollama:11434/api/tags
```

#### Port Already in Use

```bash
# Change ports in docker-compose.yml
ports:
  - "8081:8080"  # API
  - "11435:11434"  # Ollama
```

#### Permission Issues

```bash
# Reset container permissions
docker-compose down
docker-compose up --build -d
```

### Performance Issues

#### Slow Build Times

```bash
# Use build cache
docker-compose build --parallel

# Clean and rebuild
docker system prune -f
docker-compose build --no-cache
```

#### High Memory Usage

```bash
# Check resource usage
docker stats

# Limit container resources
# Add to docker-compose.yml:
deploy:
  resources:
    limits:
      cpus: '1.0'
      memory: 1G
```

## Makefile Integration

Use the provided Makefile commands:

```bash
# Build Docker image
make docker-build

# Start production environment
make docker-run

# Start development environment
make docker-dev

# View logs
make docker-logs

# Stop services
make docker-stop

# Clean everything
make docker-clean
```

## Testing with Docker

### Unit Tests in Container

```bash
# Run tests in clean environment
docker run --rm englog:latest go test ./...

# Run with test dependencies
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### Integration Testing

```bash
# Test full stack
./scripts/docker-setup.sh
sleep 30  # Wait for services

# Run integration tests
curl -f http://localhost:8080/health
curl -X POST http://localhost:8080/journals \
  -H "Content-Type: application/json" \
  -d '{"content": "Integration test"}'
```

## Migration from Local Development

To migrate from local development to Docker:

1. **Backup your data** (if using local storage)
2. **Stop local services** (API, Ollama)
3. **Run Docker setup**: `./scripts/docker-setup.sh`
4. **Test functionality** with existing API calls
5. **Update any local scripts** to use `localhost:8080`

## Next Steps

After completing PROTOTYPE-007:

- Services run in isolated containers
- Development environment with hot-reload
- Production-ready orchestration
- Comprehensive documentation and troubleshooting

The Docker setup enables:

- Consistent development environments
- Easy deployment and scaling
- Isolated service testing
- Simplified onboarding for new developers
