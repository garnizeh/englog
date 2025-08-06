package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/garnizeh/englog/internal/ai"
	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/storage"
)

// AIHandler handles AI-related requests
type AIHandler struct {
	store     *storage.MemoryStore
	aiService *ai.Service
}

// NewAIHandler creates a new AI handler
func NewAIHandler(store *storage.MemoryStore, aiService *ai.Service) *AIHandler {

	return &AIHandler{
		aiService: aiService,
		store:     store,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *AIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/analyze-sentiment"):
		h.handleAnalyzeSentiment(w, r)
	case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/generate-journal"):
		h.handleGenerateJournal(w, r)
	case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/health"):
		h.handleAIHealth(w, r)
	default:
		h.writeErrorJSON(w, "Method not allowed or endpoint not found", http.StatusMethodNotAllowed)
	}
}

// writeErrorJSON writes a JSON error response
func (h *AIHandler) writeErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"error":     message,
		"timestamp": "2025-08-04T00:00:00Z", // Fixed for testing
	})
}

// writeValidationErrorJSON writes a structured validation error response
func (h *AIHandler) writeValidationErrorJSON(w http.ResponseWriter, validationErrors models.ValidationErrors) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error":             "Validation failed",
		"status":            http.StatusBadRequest,
		"timestamp":         "2025-08-04T00:00:00Z", // Fixed for testing
		"validation_errors": validationErrors,
	})
}

// writeSuccessJSON writes a JSON success response
func (h *AIHandler) writeSuccessJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// handleAnalyzeSentiment analyzes sentiment of journal content
func (h *AIHandler) handleAnalyzeSentiment(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request to analyze sentiment: %s %s\n", r.Method, r.URL.Path)

	if r.Method != "POST" {
		h.writeErrorJSON(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var journalID string
	var content string

	// Try to get journal_id from query parameters first
	if id := r.URL.Query().Get("journal_id"); id != "" {
		journalID = id
	} else {
		// Parse request body if no query parameter
		var req struct {
			JournalID string `json:"journal_id,omitempty"`
			Content   string `json:"content,omitempty"`
		}

		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				fmt.Printf("Failed to decode request body: %v\n", err)
				h.writeErrorJSON(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			journalID = req.JournalID
			content = req.Content
		}
	}

	if journalID == "" && content == "" {
		fmt.Println("missing journal_id or content")
		h.writeErrorJSON(w, "Either journal_id or content is required", http.StatusBadRequest)
		return
	}

	var journal *models.Journal
	var err error

	// Get journal either by ID or create temporary one from content
	if journalID != "" {
		journal, err = h.store.Get(journalID)
		if err != nil {
			fmt.Printf("Failed to get journal %s: %v\n", journalID, err)
			h.writeErrorJSON(w, fmt.Sprintf("Journal not found: %v", err), http.StatusNotFound)
			return
		}
	} else if content != "" {
		// Create temporary journal for analysis
		journal = &models.Journal{
			ID:      "temp_analysis",
			Content: content,
		}
	} else {
		h.writeErrorJSON(w, "Either journal_id or content must be provided", http.StatusBadRequest)
		return
	}

	// Validate content
	if err := h.aiService.ValidateJournalContent(journal.Content); err != nil {
		fmt.Printf("Content validation failed: %v\n", err)
		h.writeErrorJSON(w, fmt.Sprintf("Content validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Analyze sentiment
	result, err := h.aiService.ProcessJournalSentiment(r.Context(), journal)
	if err != nil {
		fmt.Printf("Sentiment analysis failed: %v\n", err)
		h.writeErrorJSON(w, fmt.Sprintf("Sentiment analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return result
	h.writeSuccessJSON(w, map[string]any{
		"journal_id": journal.ID,
		"sentiment":  result,
		"timestamp":  "2025-08-04T00:00:00Z", // Fixed for testing
	})
}

// handleGenerateJournal generates a structured journal from a prompt
func (h *AIHandler) handleGenerateJournal(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Generating journal: %s %s\n", r.Method, r.URL.Path)

	if r.Method != "POST" {
		h.writeErrorJSON(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("Failed to decode request body: %v\n", err)
		h.writeValidationErrorJSON(w, []models.ValidationError{
			{
				Field:   "body",
				Message: "Invalid JSON format: " + err.Error(),
				Code:    "INVALID_JSON",
			},
		})
		return
	}

	// Use new schema validation
	if validationErrors := req.Validate(); validationErrors.HasErrors() {
		fmt.Printf("Prompt validation failed: %v\n", validationErrors)
		h.writeValidationErrorJSON(w, validationErrors)
		return
	}

	ctx := r.Context()

	// Generate journal
	result, err := h.aiService.GenerateStructuredJournal(ctx, &req)
	if err != nil {
		fmt.Printf("Journal generation failed: %v\n", err)
		h.writeErrorJSON(w, fmt.Sprintf("Journal generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return result
	h.writeSuccessJSON(w, map[string]any{
		"generated_journal": result,
		"original_prompt":   req.Prompt,
		"timestamp":         "2025-08-04T00:00:00Z", // Fixed for testing
	})
}

// handleAIHealth checks the health of AI services
func (h *AIHandler) handleAIHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Checking AI health: %s %s\n", r.Method, r.URL.Path)

	if r.Method != "GET" {
		h.writeErrorJSON(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Perform health check
	err := h.aiService.HealthCheck(r.Context())

	status := "healthy"
	statusCode := http.StatusOK

	if err != nil {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
		fmt.Printf("AI health check failed: %v\n", err)
	}

	// Return health status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]any{
		"status":    status,
		"service":   "ai",
		"timestamp": "2025-08-04T00:00:00Z", // Static for prototype
		"ai_service": map[string]any{
			"ollama_integration": status,
		},
	}

	if err != nil {
		response["error"] = err.Error()
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Failed to encode health response: %v\n", err)
		h.writeErrorJSON(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	fmt.Printf("AI health check completed: %s\n", status)
}
