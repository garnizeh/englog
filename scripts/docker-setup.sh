#!/bin/bash
set -e

echo "🚀 Setting up EngLog Docker environment..."

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if Docker is installed and running
if ! command_exists docker; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if Docker Compose is available
if ! command_exists docker-compose && ! docker compose version >/dev/null 2>&1; then
    echo "❌ Docker Compose is not available. Please install Docker Compose."
    exit 1
fi

# Use docker compose if available, otherwise fall back to docker-compose
COMPOSE_CMD="docker compose"
if ! docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker-compose"
fi

# Default values
MODE="production"
MODEL_NAME="deepseek-r1:1.5b"
SETUP_MODEL=true

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dev)
            MODE="development"
            shift
            ;;
        --model)
            MODEL_NAME="$2"
            shift 2
            ;;
        --skip-model)
            SETUP_MODEL=false
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --dev          Run in development mode with hot reload"
            echo "  --model NAME   Specify Ollama model name (default: deepseek-r1:1.5b)"
            echo "  --skip-model   Skip automatic model download"
            echo "  --help, -h     Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    # Run in production mode"
            echo "  $0 --dev             # Run in development mode"
            echo "  $0 --model llama3.2  # Use different model"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information."
            exit 1
            ;;
    esac
done

# Set the appropriate compose file
if [ "$MODE" = "development" ]; then
    COMPOSE_FILE="docker-compose.dev.yml"
    echo "🔧 Running in development mode with hot reload"
else
    COMPOSE_FILE="docker-compose.yml"
    echo "🏭 Running in production mode"
fi

echo "📦 Starting services with $COMPOSE_CMD..."

# Start the services
$COMPOSE_CMD -f $COMPOSE_FILE up -d

echo "⏳ Waiting for Ollama to be ready..."

# Wait for Ollama to be healthy
max_attempts=30
attempt=1
while [ $attempt -le $max_attempts ]; do
    if docker exec englog-ollama-$([ "$MODE" = "development" ] && echo "dev" || echo "") curl -s http://localhost:11434/api/tags >/dev/null 2>&1; then
        echo "✅ Ollama is ready!"
        break
    fi

    if [ $attempt -eq $max_attempts ]; then
        echo "❌ Timeout waiting for Ollama to be ready"
        echo "📋 Check logs with: $COMPOSE_CMD -f $COMPOSE_FILE logs ollama"
        exit 1
    fi

    echo "🔄 Attempt $attempt/$max_attempts - Waiting for Ollama..."
    sleep 10
    ((attempt++))
done

# Download the model if requested
if [ "$SETUP_MODEL" = true ]; then
    echo "📥 Downloading model: $MODEL_NAME"
    echo "⚠️  This may take several minutes depending on the model size..."

    OLLAMA_CONTAINER="englog-ollama"
    if [ "$MODE" = "development" ]; then
        OLLAMA_CONTAINER="englog-ollama-dev"
    fi

    if docker exec $OLLAMA_CONTAINER ollama pull $MODEL_NAME; then
        echo "✅ Model $MODEL_NAME downloaded successfully!"
    else
        echo "⚠️  Failed to download model $MODEL_NAME"
        echo "📝 You can download it manually later with:"
        echo "   docker exec $OLLAMA_CONTAINER ollama pull $MODEL_NAME"
    fi
fi

echo ""
echo "🎉 EngLog is now running!"
echo ""
echo "📊 Services:"
echo "  • API Server: http://localhost:8080"
echo "  • Ollama: http://localhost:11434"
echo ""
echo "🔧 Useful commands:"
echo "  • View logs: $COMPOSE_CMD -f $COMPOSE_FILE logs -f"
echo "  • Stop services: $COMPOSE_CMD -f $COMPOSE_FILE down"
echo "  • Restart API: $COMPOSE_CMD -f $COMPOSE_FILE restart api$([ "$MODE" = "development" ] && echo "-dev" || echo "")"
echo ""
echo "🧪 Test the API:"
echo "  curl http://localhost:8080/health"
echo "  curl http://localhost:8080/journals"
echo ""

# Show model info if it was downloaded
if [ "$SETUP_MODEL" = true ]; then
    echo "🤖 Current Ollama models:"
    docker exec $OLLAMA_CONTAINER ollama list 2>/dev/null || echo "   (Run 'docker exec $OLLAMA_CONTAINER ollama list' to see models)"
    echo ""
fi
