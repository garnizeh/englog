package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/storage"
	"github.com/google/uuid"
)

// JournalHandler handles journal-related HTTP requests
type JournalHandler struct {
	store *storage.MemoryStore
}

// NewJournalHandler creates a new journal handler
func NewJournalHandler(store *storage.MemoryStore) *JournalHandler {
	return &JournalHandler{
		store: store,
	}
}

// ServeHTTP implements the http.Handler interface for journal operations
func (h *JournalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request
	slog.Info("Journal request received",
		"method", r.Method,
		"path", r.URL.Path,
		"content_length", r.ContentLength)

	switch r.Method {
	case http.MethodPost:
		h.createJournal(w, r)
	case http.MethodGet:
		// Check if this is a request for a specific journal (has ID in path)
		path := strings.TrimPrefix(r.URL.Path, "/journals")
		if path != "" && path != "/" {
			// Extract ID from path (format: /journals/{id})
			id := strings.Trim(path, "/")
			h.getJournalByID(w, r, id)
		} else {
			h.getAllJournals(w, r)
		}
	default:
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createJournal handles POST /journals
func (h *JournalHandler) createJournal(w http.ResponseWriter, r *http.Request) {
	var req models.CreateJournalRequest

	// Parse and validate JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode create journal request", "error", err)
		h.sendErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Content) == "" {
		h.sendErrorResponse(w, "Content field is required and cannot be empty", http.StatusBadRequest)
		return
	}

	// Create new journal entry
	now := time.Now()
	journal := &models.Journal{
		ID:        uuid.New().String(),
		Content:   strings.TrimSpace(req.Content),
		Timestamp: now,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  req.Metadata,
	}

	// Store the journal
	if err := h.store.Store(journal); err != nil {
		slog.Error("Failed to store journal", "error", err, "journal_id", journal.ID)
		h.sendErrorResponse(w, "Failed to create journal entry", http.StatusInternalServerError)
		return
	}

	slog.Info("Journal created successfully",
		"journal_id", journal.ID,
		"content_length", len(journal.Content))

	// Return the created journal
	h.sendJSONResponse(w, journal, http.StatusCreated)
}

// getAllJournals handles GET /journals
func (h *JournalHandler) getAllJournals(w http.ResponseWriter, r *http.Request) {
	journals, err := h.store.GetAll()
	if err != nil {
		slog.Error("Failed to retrieve journals", "error", err)
		h.sendErrorResponse(w, "Failed to retrieve journals", http.StatusInternalServerError)
		return
	}

	slog.Info("Retrieved all journals", "count", len(journals))

	// Create response with journals and metadata
	response := map[string]any{
		"journals":     journals,
		"count":        len(journals),
		"retrieved_at": time.Now().UTC(),
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// getJournalByID handles GET /journals/{id}
func (h *JournalHandler) getJournalByID(w http.ResponseWriter, r *http.Request, id string) {
	// Validate ID format (basic UUID validation)
	if id == "" {
		h.sendErrorResponse(w, "Journal ID is required", http.StatusBadRequest)
		return
	}

	journal, err := h.store.Get(id)
	if err != nil {
		slog.Info("Journal not found", "journal_id", id, "error", err)
		h.sendErrorResponse(w, "Journal not found", http.StatusNotFound)
		return
	}

	slog.Info("Retrieved journal by ID", "journal_id", id)

	h.sendJSONResponse(w, journal, http.StatusOK)
}

// sendJSONResponse sends a JSON response with the given data and status code
func (h *JournalHandler) sendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
		// If we fail to encode the response, we can't send another JSON response
		// so we just log the error
	}
}

// sendErrorResponse sends a JSON error response
func (h *JournalHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := map[string]any{
		"error":     message,
		"status":    statusCode,
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		slog.Error("Failed to encode error response", "error", err)
		// Fallback to plain text error
		http.Error(w, message, statusCode)
	}
}
