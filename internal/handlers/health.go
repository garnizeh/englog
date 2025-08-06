package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/garnizeh/englog/internal/ai"
	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/storage"
)

var startTime = time.Now() // Application start time

// HealthHandler handles health check and status endpoints
type HealthHandler struct {
	store     *storage.MemoryStore
	aiService ai.AIService
	logger    *logging.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(store *storage.MemoryStore, aiService ai.AIService, logger *logging.Logger) *HealthHandler {
	return &HealthHandler{
		store:     store,
		aiService: aiService,
		logger:    logger,
	}
}

// ServeHTTP implements the http.Handler interface for health checks and status endpoints
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Route based on path
	path := strings.TrimPrefix(r.URL.Path, "/")
	switch path {
	case "health":
		h.handleHealth(w, r)
	case "status":
		h.handleStatus(w, r)
	case "status/ollama":
		h.handleOllamaStatus(w, r)
	default:
		h.sendErrorResponse(w, "Not found", http.StatusNotFound)
	}
}

// handleHealth handles the basic health check endpoint
func (h *HealthHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	requestLogger := h.logger.WithContext(r.Context())
	start := time.Now()

	response := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "englog-api",
		"version":   "prototype-009",
		"storage": map[string]any{
			"type":          "memory",
			"journal_count": h.store.Count(),
		},
		"response_time_ms": time.Since(start).Milliseconds(),
	}

	h.sendJSONResponse(w, response, http.StatusOK)

	requestLogger.Debug("Health check completed successfully",
		"response_time_ms", time.Since(start).Milliseconds(),
	)
}

// handleStatus handles the system status endpoint with detailed information
func (h *HealthHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	requestLogger := h.logger.WithContext(r.Context())
	start := time.Now()

	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get journal statistics
	journalStats := h.store.GetStats()

	uptime := time.Since(startTime)

	response := map[string]any{
		"status":         "healthy",
		"timestamp":      time.Now().UTC(),
		"service":        "englog-api",
		"version":        "prototype-009",
		"uptime_seconds": uptime.Seconds(),
		"uptime_human":   uptime.String(),
		"memory": map[string]any{
			"allocated_bytes":       memStats.Alloc,
			"allocated_mb":          float64(memStats.Alloc) / 1024 / 1024,
			"total_allocated_bytes": memStats.TotalAlloc,
			"total_allocated_mb":    float64(memStats.TotalAlloc) / 1024 / 1024,
			"heap_objects":          memStats.HeapObjects,
			"gc_cycles":             memStats.NumGC,
		},
		"storage": map[string]any{
			"type":                   "memory",
			"journal_count":          journalStats.TotalJournals,
			"processed_count":        journalStats.ProcessedJournals,
			"avg_processing_time_ms": journalStats.AvgProcessingTimeMS,
		},
		"response_time_ms": time.Since(start).Milliseconds(),
	}

	h.sendJSONResponse(w, response, http.StatusOK)

	requestLogger.Debug("Status check completed successfully",
		"journal_count", journalStats.TotalJournals,
		"memory_mb", float64(memStats.Alloc)/1024/1024,
		"uptime_seconds", uptime.Seconds(),
		"response_time_ms", time.Since(start).Milliseconds(),
	)
}

// handleOllamaStatus handles the Ollama connectivity check endpoint
func (h *HealthHandler) handleOllamaStatus(w http.ResponseWriter, r *http.Request) {
	requestLogger := h.logger.WithContext(r.Context())
	start := time.Now()

	// Test Ollama connectivity
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	ollamaErr := h.aiService.HealthCheck(ctx)
	isHealthy := ollamaErr == nil
	statusCode := http.StatusOK

	response := map[string]any{
		"status":           "healthy",
		"timestamp":        time.Now().UTC(),
		"service":          "ollama-integration",
		"connected":        isHealthy,
		"response_time_ms": time.Since(start).Milliseconds(),
	}

	if !isHealthy {
		response["status"] = "unhealthy"
		response["error"] = ollamaErr.Error()
		statusCode = http.StatusServiceUnavailable

		requestLogger.Error("Ollama health check failed",
			"error", ollamaErr,
			"response_time_ms", time.Since(start).Milliseconds(),
		)
	} else {
		requestLogger.Debug("Ollama health check completed successfully",
			"response_time_ms", time.Since(start).Milliseconds(),
		)
	}

	h.sendJSONResponse(w, response, statusCode)
}

// sendJSONResponse sends a JSON response with the given data and status code
func (h *HealthHandler) sendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// sendErrorResponse sends an error response with the given message and status code
func (h *HealthHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]any{
		"status":    statusCode,
		"error":     message,
		"timestamp": time.Now().UTC(),
	}

	h.sendJSONResponse(w, response, statusCode)
}
