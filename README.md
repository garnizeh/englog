# englog

Welcome to **englog**, a project built through a unique collaboration between a human developer and an artificial intelligence assistant. Our goal is to create an exceptional product by combining human creativity with AI-driven development.

## Vision

**englog** is a system designed to empower users by helping them collect, process, and analyze their personal journals using the power of artificial intelligence. By transforming daily entries into structured insights, we aim to provide a tool for self-reflection and personal growth.

This repository is a living testament to a new way of building software, where human-AI partnership is at the core of the creative process.

## Current Status: Prototype Phase 0 - FEATURE COMPLETE

We have **successfully completed Phase 0** prototype development with all 9 planned tasks finished. The prototype demonstrates a fully functional AI-powered journaling system with comprehensive testing and documentation.

### Completed Prototypes:

- ✅ **PROTOTYPE-001:** Basic API structure and health endpoints
- ✅ **PROTOTYPE-002:** Journal CRUD operations with in-memory storage
- ✅ **PROTOTYPE-003:** Ollama AI integration for sentiment analysis and journal generation
- ✅ **PROTOTYPE-004:** Synchronous AI processing with graceful failure handling
- ✅ **PROTOTYPE-005:** JSON schema design and data validation
- ✅ **PROTOTYPE-006:** Structured logging and request debugging
- ✅ **PROTOTYPE-007:** Docker development environment setup
- ✅ **PROTOTYPE-008:** Manual testing documentation and API examples
- ✅ **PROTOTYPE-009:** Health and status monitoring endpoints

### Current Features (Prototype Phase 0 Complete):

- **Complete AI Processing Pipeline:** Journal entries are automatically analyzed for sentiment, themes, and insights
- **Synchronous Processing:** AI analysis completes before API response is returned with graceful failure handling
- **Comprehensive Monitoring:** Health checks, system status, and Ollama connectivity monitoring
- **Docker Development Environment:** Full containerized setup with hot-reload and automated scripts
- **Extensive Testing Documentation:** 20+ use cases with Bruno collection and curl examples
- **Structured Logging:** Production-ready JSON logging with request tracing and performance metrics
- **Flexible Data Models:** JSON schema validation with extensible metadata support

## Core Components

The system is architected around three main components:

1.  **API Service (Go):** A robust backend service with in-memory storage and AI integration

    - RESTful endpoints for journal management
    - Synchronous AI processing pipeline
    - Comprehensive error handling and logging

2.  **AI Worker (Go):** An in-memory processing system that enriches journal entries

    - Sentiment analysis using Ollama integration
    - Processing status tracking (pending, completed, failed)
    - Configurable timeouts and retry logic

3.  **Web Application:** _(Planned for future prototypes)_

## Technology Stack

**Current Phase 0 Implementation:**

- **Backend:** Go 1.21+ with Gin-style HTTP handling and structured JSON logging
- **AI Integration:** Ollama local LLM with configurable models (tested with deepseek-r1:1.5b)
- **Storage:** In-memory with comprehensive statistics and data persistence simulation
- **Testing:** Unit tests, integration tests, and extensive manual testing documentation
- **Development:** Docker Compose with hot-reload, automated setup scripts, and environment management
- **Monitoring:** Health checks, system status monitoring, and performance metrics

**Planned for Phase 1 (MVP):**

- **Database:** PostgreSQL with JSONB for flexible schema and pgvector for embeddings
- **Authentication:** JWT tokens with OAuth 2.0 (Google, GitHub) and email OTP
- **Frontend:** Vanilla JavaScript SPA with modern CSS frameworks
- **Deployment:** Kubernetes with auto-scaling, monitoring, and production-ready infrastructure

## API Endpoints

Complete API surface (Phase 0 - All Prototypes):

**Core Journal Management:**

- `POST /journals` - Create journal with automatic AI processing and validation
- `GET /journals` - List all journals with AI results and metadata
- `GET /journals/{id}` - Get specific journal with comprehensive AI analysis
- `PUT /journals/{id}` - Update journal content with re-processing
- `DELETE /journals/{id}` - Remove journal and associated AI data

**AI Processing & Analysis:**

- `POST /ai/analyze-sentiment` - Direct sentiment analysis endpoint
- `POST /ai/generate-journal` - AI-powered journal generation with prompts
- `GET /ai/health` - AI service health check and model availability

**System Monitoring & Health:**

- `GET /health` - Basic API health check with response time metrics
- `GET /status` - Comprehensive system status (uptime, memory, journal statistics)
- `GET /status/ollama` - Ollama connectivity and model availability check

**Development & Testing:**

- Complete Bruno API collection with 15+ organized requests
- Comprehensive curl examples for all endpoints
- Error handling scenarios and validation examples

## Getting Started

### Prerequisites

- Go 1.21 or later
- Ollama installed and running locally
- A compatible Ollama model (e.g., `deepseek-r1:1.5b`)

### Quick Start

1. **Clone the repository:**

   ```bash
   git clone https://github.com/garnizeh/englog.git
   cd englog
   ```

2. **Install dependencies:**

   ```bash
   go mod download
   ```

3. **Start Ollama and pull a model:**

   ```bash
   ollama serve
   ollama pull deepseek-r1:1.5b  # Or your preferred model
   ```

4. **Run the API server:**

   ```bash
   go run cmd/api/main.go
   # Or build and run:
   # go build -o bin/englog-api cmd/api/main.go
   # ./bin/englog-api
   ```

5. **Test the API:**

   ```bash
   # Check system health and status
   curl http://localhost:8080/health
   curl http://localhost:8080/status
   curl http://localhost:8080/status/ollama

   # Create a journal with automatic AI processing
   curl -X POST http://localhost:8080/journals \
     -H "Content-Type: application/json" \
     -d '{
       "content": "Today was a wonderful day! I learned so much about AI and programming.",
       "metadata": {"mood": 8, "location": "home", "tags": ["learning", "positive"]}
     }'

   # Get all journals with AI results
   curl http://localhost:8080/journals

   # Test AI processing directly
   curl -X POST http://localhost:8080/ai/analyze-sentiment \
     -H "Content-Type: application/json" \
     -d '{"content": "I feel excited about the future!"}'
   ```

### Docker Setup (Optional)

For consistent development environments and easier setup, you can run the entire stack using Docker:

1. **Quick Start with Docker:**

   ```bash
   # Run the automated setup script
   ./scripts/docker-setup.sh

   # Or start manually
   docker-compose up -d
   ```

2. **Development Mode with Hot Reload:**

   ```bash
   # Use development configuration
   ./scripts/docker-setup.sh --dev

   # Or manually
   docker-compose -f docker-compose.dev.yml up -d
   ```

3. **Custom Model:**

   ```bash
   # Use a different Ollama model
   ./scripts/docker-setup.sh --model llama3.2
   ```

4. **Docker Services:**

   - **API Server:** http://localhost:8080
   - **Ollama:** http://localhost:11434
   - **Health Checks:** Automatic service health monitoring

5. **Docker Commands:**

   ```bash
   # View logs
   docker-compose logs -f

   # Stop services
   docker-compose down

   # Restart API only
   docker-compose restart api

   # Clean up (removes volumes)
   docker-compose down -v
   ```

The Docker setup automatically:

- Downloads and configures Ollama with the specified model
- Sets up networking between services
- Configures health checks and proper startup order
- Provides optional hot-reload for development

### Configuration

Environment variables (Phase 0 complete configuration):

**Core Server Configuration:**

- `PORT`: Server port (default: 8080)
- `OLLAMA_SERVER_URL`: Ollama server URL (default: http://localhost:11434)
- `OLLAMA_MODEL_NAME`: Model to use (default: deepseek-r1:1.5b)

**Logging Configuration:**

- `LOG_LEVEL`: Logging level (debug, info, warn, error - default: info)
- `LOG_FORMAT`: Log format (text, json - default: json for structured logging)

**AI Processing Configuration:**

- `AI_TIMEOUT`: AI processing timeout in seconds (default: 30s)
- `AI_RETRY_ATTEMPTS`: Number of retry attempts for failed AI requests (default: 3)

**Development Configuration:**

- `ENVIRONMENT`: Environment name (development, staging, production)
- `DEBUG_MODE`: Enable debug features and verbose logging (default: false)

### Running Tests

```bash
# Run all tests with coverage
go test ./... -v -cover

# Run specific test packages
go test ./internal/worker/... -v
go test ./internal/handlers/... -v
go test ./internal/storage/... -v

# Run with race detection
go test -race ./...

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Manual Testing & Validation

Phase 0 includes comprehensive manual testing resources:

1. **Bruno API Collection:** Import `/bruno-collection/` for full API testing
2. **Use Case Documentation:** `/docs/hands-on/PROTOTYPE-005-USE-CASES.md` - 20+ detailed test scenarios
3. **Docker Testing:** `/docs/hands-on/DOCKER.md` - Containerized testing environment
4. **Performance Validation:** Response time benchmarks and load testing examples

### Next Steps (Phase 1 - MVP)

With Phase 0 prototype complete, the next phase involves:

- **Database Integration:** Migration from in-memory to PostgreSQL with JSONB
- **Authentication System:** OAuth 2.0 + email OTP implementation
- **Web Frontend:** Vanilla JavaScript SPA for user interface
- **Worker Pool Architecture:** Distributed asynchronous AI processing
- **Production Deployment:** Kubernetes deployment with monitoring

See `/docs/planning/ROADMAP.md` for complete development roadmap.

## Contribution

This project follows a unique human-AI collaborative model. All contributions are a result of the synergy between the human project lead and the AI assistant.

### Phase 0 Achievement

**🎉 MILESTONE REACHED:** Phase 0 prototype development is now **COMPLETE** with all 9 planned tasks successfully implemented and tested. This represents a fully functional proof-of-concept that validates:

- ✅ **Technical Feasibility:** AI-powered journal analysis works reliably with local LLMs
- ✅ **Architecture Viability:** Clean separation between API, storage, and AI processing layers
- ✅ **Development Workflow:** Effective human-AI collaboration for rapid feature development
- ✅ **Quality Standards:** Comprehensive testing, documentation, and monitoring capabilities

The prototype successfully demonstrates automatic sentiment analysis, theme extraction, and insight generation from natural language journal entries, with response times under 3 seconds and comprehensive error handling.

**Ready for Phase 1:** With Phase 0 complete, the project is now ready to evolve into the MVP phase with database persistence, authentication, and web interface development.

---

Let's begin this exciting journey!
