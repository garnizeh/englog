package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/handlers"
	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/storage"
	"github.com/garnizeh/englog/internal/worker"
)

// mockAIProcessor is a mock implementation for testing
type mockAIProcessor struct {
	shouldFail      bool
	sentimentResult *models.SentimentResult
}

func (m *mockAIProcessor) ProcessJournalSentiment(ctx context.Context, journal *models.Journal) (*models.SentimentResult, error) {
	if m.shouldFail {
		return nil, errors.New("mock AI processing error")
	}

	if m.sentimentResult != nil {
		return m.sentimentResult, nil
	}

	// Default successful response
	return &models.SentimentResult{
		Score:       0.7,
		Label:       "positive",
		Confidence:  0.85,
		ProcessedAt: time.Now(),
	}, nil
}

func TestJournalHandlers(t *testing.T) {
	// Setup
	store := storage.NewMemoryStore()
	mockAI := &mockAIProcessor{
		sentimentResult: &models.SentimentResult{
			Score:       0.8,
			Label:       "positive",
			Confidence:  0.9,
			ProcessedAt: time.Now(),
		},
	}
	aiWorker := worker.NewInMemoryWorker(mockAI, Logger())
	handler := handlers.NewJournalHandler(store, aiWorker, Logger())

	t.Run("CreateJournal", func(t *testing.T) {
		createReq := models.CreateJournalRequest{
			Content: "This is my first journal entry for testing.",
			Metadata: map[string]any{
				"mood":     8,
				"location": "home",
				"tags":     []string{"test", "first-entry"},
			},
		}

		jsonData, err := json.Marshal(createReq)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/journals", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var journal models.Journal
		if err := json.NewDecoder(w.Body).Decode(&journal); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if journal.ID == "" {
			t.Error("Expected journal ID to be generated")
		}

		if journal.Content != createReq.Content {
			t.Errorf("Expected content '%s', got '%s'", createReq.Content, journal.Content)
		}

		// Check that AI processing was performed
		if journal.ProcessingResult == nil {
			t.Error("Expected processing result to be set")
		} else {
			if journal.ProcessingResult.Status != models.ProcessingStatusCompleted {
				t.Errorf("Expected processing status to be completed, got %v", journal.ProcessingResult.Status)
			}

			if journal.ProcessingResult.SentimentResult == nil {
				t.Error("Expected sentiment result to be set")
			} else {
				if journal.ProcessingResult.SentimentResult.Score != 0.8 {
					t.Errorf("Expected sentiment score 0.8, got %v", journal.ProcessingResult.SentimentResult.Score)
				}
				if journal.ProcessingResult.SentimentResult.Label != "positive" {
					t.Errorf("Expected sentiment label 'positive', got %v", journal.ProcessingResult.SentimentResult.Label)
				}
			}

			if journal.ProcessingResult.ProcessedAt == nil {
				t.Error("Expected processed_at to be set")
			}

			if journal.ProcessingResult.ProcessingTime == nil {
				t.Error("Expected processing_time to be set")
			}
		}
	})

	t.Run("GetJournals", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/journals", nil)

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]any
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if _, ok := response["journals"]; !ok {
			t.Error("Expected 'journals' field in response")
		}

		if _, ok := response["count"]; !ok {
			t.Error("Expected 'count' field in response")
		}
	})

	t.Run("CreateJournalWithEmptyContent", func(t *testing.T) {
		createReq := models.CreateJournalRequest{
			Content: "",
		}

		jsonData, err := json.Marshal(createReq)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/journals", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d for empty content, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CreateJournalWithInvalidJSON", func(t *testing.T) {
		invalidJSON := `{"content": "test", "invalid": json}`

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/journals", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("GetJournalByID", func(t *testing.T) {
		// First create a journal
		createReq := models.CreateJournalRequest{
			Content: "Test journal for ID retrieval",
		}

		jsonData, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/journals", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		handler.ServeHTTP(w, req)

		var created models.Journal
		json.NewDecoder(w.Body).Decode(&created)

		// Now get by ID
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/journals/"+created.ID, nil)

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var retrieved models.Journal
		if err := json.NewDecoder(w.Body).Decode(&retrieved); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}
	})

	t.Run("GetNonExistentJournal", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/journals/non-existent-id", nil)

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d for non-existent journal, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/journals", nil)

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d for DELETE method, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("CreateJournalWithAIProcessingFailure", func(t *testing.T) {
		// Setup handler with failing AI processor
		failingMockAI := &mockAIProcessor{
			shouldFail: true,
		}
		failingWorker := worker.NewInMemoryWorker(failingMockAI, Logger())
		failingHandler := handlers.NewJournalHandler(store, failingWorker, Logger())

		createReq := models.CreateJournalRequest{
			Content: "This journal will have AI processing failure.",
		}

		jsonData, err := json.Marshal(createReq)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/journals", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		failingHandler.ServeHTTP(w, req)

		// Journal creation should still succeed even if AI processing fails
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var journal models.Journal
		if err := json.NewDecoder(w.Body).Decode(&journal); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check that processing result shows failure but journal is still created
		if journal.ProcessingResult == nil {
			t.Error("Expected processing result to be set")
		} else {
			if journal.ProcessingResult.Status != models.ProcessingStatusFailed {
				t.Errorf("Expected processing status to be failed, got %v", journal.ProcessingResult.Status)
			}

			if journal.ProcessingResult.Error == "" {
				t.Error("Expected error message to be set")
			}

			if journal.ProcessingResult.SentimentResult != nil {
				t.Error("Expected sentiment result to be nil on failure")
			}
		}

		// Journal content should still be preserved
		if journal.Content != createReq.Content {
			t.Errorf("Expected content '%s', got '%s'", createReq.Content, journal.Content)
		}
	})

	t.Run("CreateJournalWithoutWorker", func(t *testing.T) {
		// Setup handler without worker (nil worker)
		handlerWithoutWorker := handlers.NewJournalHandler(store, nil, Logger())

		createReq := models.CreateJournalRequest{
			Content: "This journal will not have AI processing.",
		}

		jsonData, err := json.Marshal(createReq)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/journals", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		handlerWithoutWorker.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var journal models.Journal
		if err := json.NewDecoder(w.Body).Decode(&journal); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should not have processing result when no worker is available
		if journal.ProcessingResult != nil {
			t.Error("Expected processing result to be nil when no worker is available")
		}
	})
}
