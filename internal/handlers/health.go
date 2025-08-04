package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/garnizeh/englog/internal/storage"
)

// HealthHandler handles the health check endpoint
type HealthHandler struct {
	store *storage.MemoryStore
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(store *storage.MemoryStore) *HealthHandler {
	return &HealthHandler{
		store: store,
	}
}

// ServeHTTP implements the http.Handler interface for health checks
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "englog-api",
		"version":   "prototype-001",
		"storage": map[string]any{
			"type":          "memory",
			"journal_count": h.store.Count(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode health response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Health check requested",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"))
}
