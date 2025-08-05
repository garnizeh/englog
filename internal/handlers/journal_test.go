package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/englog/internal/handlers"
	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/storage"
)

func TestJournalHandlers(t *testing.T) {
	// Setup
	store := storage.NewMemoryStore()
	handler := handlers.NewJournalHandler(store)

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
}
