# englog

Welcome to **englog**, a project built through a unique collaboration between a human developer and an artificial intelligence assistant. Our goal is to create an exceptional product by combining human creativity with AI-driven development.

## Vision

**englog** is a system designed to empower users by helping them collect, process, and analyze their personal journals using the power of artificial intelligence. By transforming daily entries into structured insights, we aim to provide a tool for self-reflection and personal growth.

This repository is a living testament to a new way of building software, where human-AI partnership is at the core of the creative process.

## Current Status: Prototype Phase 0

We're currently in the **prototype development phase**, building and testing core functionalities. Recent achievements:

- ✅ **PROTOTYPE-001:** Basic API structure and health endpoints
- ✅ **PROTOTYPE-002:** Journal CRUD operations with in-memory storage
- ✅ **PROTOTYPE-003:** Ollama AI integration for sentiment analysis and journal generation
- ✅ **PROTOTYPE-004:** Synchronous AI processing with graceful failure handling

### Latest Features (Prototype-004):

- **Automatic AI Processing:** Journal entries are automatically analyzed for sentiment when created
- **Synchronous Processing:** AI analysis completes before API response is returned
- **Graceful Failure:** Journal creation succeeds even when AI processing fails
- **Real-time Results:** GET requests return both journal content and AI analysis results

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

- **Backend:** Go with structured logging
- **AI Integration:** Ollama (local LLM inference)
- **Storage:** In-memory (prototype phase)
- **Testing:** Comprehensive unit and integration tests
- **Future:** NoSQL database, cloud deployment, web interface

## API Endpoints

Current endpoints (Prototype-004):

- `GET /health` - System health check
- `POST /journals` - Create journal with automatic AI processing
- `GET /journals` - List all journals with AI results
- `GET /journals/{id}` - Get specific journal with AI analysis
- `POST /ai/analyze-sentiment` - Direct sentiment analysis
- `POST /ai/generate-journal` - AI-powered journal generation
- `GET /ai/health` - AI service health check

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
   # Create a journal with automatic AI processing
   curl -X POST http://localhost:8080/journals \
     -H "Content-Type: application/json" \
     -d '{
       "content": "Today was a wonderful day! I learned so much.",
       "metadata": {"mood": 8, "location": "home"}
     }'

   # Get all journals with AI results
   curl http://localhost:8080/journals
   ```

### Configuration

Environment variables:

- `PORT`: Server port (default: 8080)
- `OLLAMA_SERVER_URL`: Ollama server URL (default: http://localhost:11434)
- `OLLAMA_MODEL_NAME`: Model to use (default: deepseek-r1:1.5b)

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test packages
go test ./internal/worker/...
go test ./internal/handlers/...
```

## Contribution

This project follows a unique human-AI collaborative model. All contributions are a result of the synergy between the human project lead and the AI assistant.

---

Let's begin this exciting journey!
