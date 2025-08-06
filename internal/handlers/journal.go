package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/storage"
	"github.com/garnizeh/englog/internal/worker"
	"github.com/google/uuid"
)

// JournalHandler handles journal-related HTTP requests
type JournalHandler struct {
	store  *storage.MemoryStore
	worker *worker.InMemoryWorker
	logger *logging.Logger
}

// NewJournalHandler creates a new journal handler
func NewJournalHandler(store *storage.MemoryStore, worker *worker.InMemoryWorker, logger *logging.Logger) *JournalHandler {
	return &JournalHandler{
		store:  store,
		worker: worker,
		logger: logger,
	}
}

// ServeHTTP implements the http.Handler interface for journal operations
func (h *JournalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create logger with request context
	requestLogger := h.logger.WithContext(r.Context())

	// Log the incoming request
	requestLogger.LogHTTPRequest(
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		r.Header.Get("User-Agent"),
		r.ContentLength,
	)

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
		requestLogger := h.logger.WithContext(r.Context())
		requestLogger.Error("Failed to decode create journal request", "error", err)
		h.sendValidationErrorResponse(w, []models.ValidationError{
			{
				Field:   "body",
				Message: "Invalid JSON format: " + err.Error(),
				Code:    "INVALID_JSON",
			},
		})
		return
	}

	// Validate request using schema validation
	if validationErrors := req.Validate(); validationErrors.HasErrors() {
		requestLogger := h.logger.WithContext(r.Context())
		requestLogger.LogValidationError("create_journal", validationErrors)
		h.sendValidationErrorResponse(w, validationErrors)
		return
	}

	// Create new journal entry with validated and trimmed content
	now := time.Now()
	journal := &models.Journal{
		ID:        uuid.New().String(),
		Content:   strings.TrimSpace(req.Content),
		Timestamp: now,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  req.Metadata,
	}

	// Process journal with AI synchronously (with graceful failure handling)
	if h.worker != nil {
		h.logger.LogAIProcessingStart(journal.ID, journal.Content, len(journal.Content))

		h.worker.ProcessJournalWithGracefulFailure(r.Context(), journal)

		if journal.ProcessingResult != nil {
			var durationMs int64
			if journal.ProcessingResult.ProcessingTime != nil {
				durationMs = journal.ProcessingResult.ProcessingTime.Nanoseconds() / int64(time.Millisecond)
			}
			h.logger.LogAIProcessingComplete(journal.ID,
				durationMs,
				journal.ProcessingResult.Status == "completed",
				"")
		}
	} else {
		h.logger.WithContext(r.Context()).Warn("No AI worker available, skipping processing",
			"journal_id", journal.ID)
	}

	// Store the journal (with processing results if available)
	if err := h.store.Store(journal); err != nil {
		h.logger.LogStorageOperation("store", "journal", journal.ID, false, err.Error())
		h.sendErrorResponse(w, "Failed to create journal entry", http.StatusInternalServerError)
		return
	}

	h.logger.WithContext(r.Context()).Info("Journal created successfully",
		"journal_id", journal.ID,
		"content_length", len(journal.Content),
		"ai_processed", journal.ProcessingResult != nil,
		"processing_status", func() string {
			if journal.ProcessingResult != nil {
				return string(journal.ProcessingResult.Status)
			}
			return "not_processed"
		}())

	// Return the created journal with processing results
	h.sendJSONResponse(w, journal, http.StatusCreated)
}

// getAllJournals handles GET /journals
func (h *JournalHandler) getAllJournals(w http.ResponseWriter, r *http.Request) {
	journals, err := h.store.GetAll()
	if err != nil {
		h.logger.LogStorageOperation("get_all", "journal", "all", false, err.Error())
		h.sendErrorResponse(w, "Failed to retrieve journals", http.StatusInternalServerError)
		return
	}

	h.logger.WithContext(r.Context()).Info("Retrieved all journals", "count", len(journals))

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
		h.logger.WithContext(r.Context()).Info("Journal not found", "journal_id", id, "error", err)
		h.sendErrorResponse(w, "Journal not found", http.StatusNotFound)
		return
	}

	h.logger.WithContext(r.Context()).Info("Retrieved journal by ID", "journal_id", id)

	h.sendJSONResponse(w, journal, http.StatusOK)
}

// sendJSONResponse sends a JSON response with the given data and status code
func (h *JournalHandler) sendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
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
		h.logger.Error("Failed to encode error response", "error", err)
		// Fallback to plain text error
		http.Error(w, message, statusCode)
	}
}

// sendValidationErrorResponse sends a structured validation error response
func (h *JournalHandler) sendValidationErrorResponse(w http.ResponseWriter, validationErrors models.ValidationErrors) {
	errorResponse := map[string]any{
		"error":             "Validation failed",
		"status":            http.StatusBadRequest,
		"timestamp":         time.Now().UTC(),
		"validation_errors": validationErrors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.logger.Error("Failed to encode validation error response", "error", err)
		// Fallback to simple error response
		h.sendErrorResponse(w, "Validation failed", http.StatusBadRequest)
	}
}
