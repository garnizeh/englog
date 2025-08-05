package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/garnizeh/englog/internal/ai"
	"github.com/garnizeh/englog/internal/handlers"
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

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

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

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(store)

	// Initialize AI service
	aiService, err := ai.NewService(ctx, modelName, ollamaURL)
	if err != nil {
		slog.Error("Failed to create AI service", "error", err)
		os.Exit(1)
	}

	// Initialize AI worker for synchronous processing
	aiWorker := worker.NewInMemoryWorker(aiService)

	// Initialize journal handler with AI worker
	journalHandler := handlers.NewJournalHandler(store, aiWorker)

	aiHandler := handlers.NewAIHandler(store, aiService)

	// Setup HTTP server and routes
	mux := http.NewServeMux()

	// Add middleware for logging and basic error handling
	mux.Handle("/health", loggingMiddleware(healthHandler))
	mux.Handle("/journals", loggingMiddleware(journalHandler))
	mux.Handle("/journals/", loggingMiddleware(journalHandler)) // For /journals/{id} paths

	// AI endpoints
	mux.Handle("/ai/analyze-sentiment", loggingMiddleware(aiHandler))
	mux.Handle("/ai/generate-journal", loggingMiddleware(aiHandler))
	mux.Handle("/ai/health", loggingMiddleware(aiHandler))

	mux.Handle("/", loggingMiddleware(http.HandlerFunc(defaultHandler)))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  600 * time.Second,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		slog.Info("Starting EngLog API server",
			"port", port,
			"version", "prototype-004",
			"storage", "memory",
			"ai_integration", "ollama",
			"features", []string{"synchronous_ai_processing", "sentiment_analysis"})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Server is ready to handle requests", "port", port)

	// Wait for interrupt signal
	<-quit
	slog.Info("Server is shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}

// loggingMiddleware adds request logging to handlers
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		slog.Info("HTTP request processed",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.Header.Get("User-Agent"))
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// defaultHandler handles requests to unknown endpoints
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"message": "EngLog API - Phase 0 (Dev Prototype)",
		"version": "prototype-004",
		"status":  "active",
		"features": []string{
			"Journal CRUD operations",
			"Synchronous AI sentiment analysis",
			"In-memory storage",
			"Ollama integration",
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
		slog.Error("Failed to encode default response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
