package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/garnizeh/englog/internal/ai"
	"github.com/garnizeh/englog/internal/handlers"
	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/middleware"
	"github.com/garnizeh/englog/internal/storage"
	"github.com/garnizeh/englog/internal/worker"
)

const (
	defaultPort      = "8080"
	defaultModelName = "deepseek-r1:1.5b"
	defaultOllamaURL = "http://localhost:11434"
)

func main() {
	ctx := context.Background()

	// Setup structured logging from environment
	logger := logging.NewLoggerFromEnv()

	// Initialize in-memory storage
	store := storage.NewMemoryStore()

	// Get ollama model name from environment or use default
	modelName := os.Getenv("OLLAMA_MODEL_NAME")
	if modelName == "" {
		modelName = defaultModelName
	}

	// Get ollama server URL from environment or use default
	ollamaURL := os.Getenv("OLLAMA_SERVER_URL")
	if ollamaURL == "" {
		ollamaURL = defaultOllamaURL
	}

	// Log startup configuration
	logger.LogSystemEvent("application_startup", map[string]any{
		"version":     "prototype-006",
		"storage":     "memory",
		"ai_provider": "ollama",
		"model_name":  modelName,
		"ollama_url":  ollamaURL,
		"log_level":   os.Getenv("LOG_LEVEL"),
		"log_format":  os.Getenv("LOG_FORMAT"),
	})

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(store, logger)

	// Initialize AI service
	aiService, err := ai.NewService(ctx, modelName, ollamaURL, logger)
	if err != nil {
		logger.Error("Failed to create AI service", "error", err)
		os.Exit(1)
	}

	// Initialize AI worker for synchronous processing
	aiWorker := worker.NewInMemoryWorker(aiService, logger)

	// Initialize journal handler with AI worker
	journalHandler := handlers.NewJournalHandler(store, aiWorker, logger)

	aiHandler := handlers.NewAIHandler(store, aiService, logger)

	// Setup HTTP server and routes
	mux := http.NewServeMux()

	// Create middleware instance
	requestMiddleware := middleware.NewRequestMiddleware(logger)

	// Add comprehensive middleware stack with new logging middleware
	var handler http.Handler = mux

	// Add our new middleware stack in reverse order (last added = first executed)
	handler = requestMiddleware.RecoveryMiddleware(handler)
	handler = requestMiddleware.PerformanceMiddleware(handler)
	handler = requestMiddleware.LoggingMiddleware(handler)

	// Add routes without the old middleware (new middleware handles all requests)
	mux.Handle("/health", healthHandler)
	mux.Handle("/journals", journalHandler)
	mux.Handle("/journals/", journalHandler) // For /journals/{id} paths

	// AI endpoints
	mux.Handle("/ai/analyze-sentiment", aiHandler)
	mux.Handle("/ai/generate-journal", aiHandler)
	mux.Handle("/ai/health", aiHandler)

	mux.Handle("/", http.HandlerFunc(defaultHandler))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler, // Use our middleware-wrapped handler
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  600 * time.Second,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.WithContext(ctx).Info("Starting EngLog API server",
			"port", port,
			"version", "prototype-006",
			"storage", "memory",
			"ai_integration", "ollama",
			"ollama_model", modelName,
			"ollama_url", ollamaURL,
			"features", []string{"synchronous_ai_processing", "sentiment_analysis", "structured_logging"})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	logger.WithContext(ctx).Info("Server is ready to handle requests", "port", port)

	// Wait for interrupt signal
	<-quit
	logger.WithContext(ctx).Info("Server is shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.WithContext(ctx).Info("Server stopped gracefully")
}

// defaultHandler handles requests to unknown endpoints
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"message": "EngLog API - Phase 0 (Dev Prototype)",
		"version": "prototype-006",
		"status":  "active",
		"features": []string{
			"Journal CRUD operations",
			"Synchronous AI sentiment analysis",
			"In-memory storage",
			"Ollama integration",
			"Structured logging and observability",
		},
		"endpoints": map[string]string{
			"health":            "/health",
			"create_journal":    "POST /journals",
			"get_all_journals":  "GET /journals",
			"get_journal_by_id": "GET /journals/{id}",
			"ai_analyze":        "POST /ai/analyze-sentiment",
			"ai_generate":       "POST /ai/generate-journal",
			"ai_health":         "GET /ai/health",
		},
		"documentation": "https://github.com/garnizeh/englog",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Note: logging is handled by middleware, but this is a fallback for encoding errors
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
