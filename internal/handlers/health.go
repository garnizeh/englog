package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/storage"
)

// HealthHandler handles the health check endpoint
type HealthHandler struct {
	store  *storage.MemoryStore
	logger *logging.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(store *storage.MemoryStore, logger *logging.Logger) *HealthHandler {
	return &HealthHandler{
		store:  store,
		logger: logger,
	}
}

// ServeHTTP implements the http.Handler interface for health checks
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create logger with request context
	requestLogger := h.logger.WithContext(r.Context())

	response := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "englog-api",
		"version":   "prototype-006",
		"storage": map[string]any{
			"type":          "memory",
			"journal_count": h.store.Count(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		requestLogger.Error("Failed to encode health response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	requestLogger.Debug("Health check completed successfully",
		"journal_count", h.store.Count(),
	)
}
