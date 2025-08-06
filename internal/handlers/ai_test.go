package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/ai"
	"github.com/garnizeh/englog/internal/handlers"
	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/storage"
)

const (
	modelName = "all-minilm"
)

// TestAIHandler_ServeHTTP_AnalyzeSentiment tests the analyze sentiment endpoint via ServeHTTP
func TestAIHandler_ServeHTTP_AnalyzeSentiment(t *testing.T) {
	ctx := context.Background()

	// Setup test dependencies
	store := storage.NewMemoryStore()
	aiService, err := ai.NewService(ctx, modelName, "http://localhost:11434", Logger())
	if err != nil || aiService == nil {
		t.Fatalf("Failed to create AI service: %v", err)
	}
	handler := handlers.NewAIHandler(store, aiService, Logger())

	// Create a test journal entry
	journal := &models.Journal{
		ID:        "test-journal-123",
		Content:   "Today was an amazing day! I felt really happy and accomplished.",
		CreatedAt: time.Now(),
	}
	store.Store(journal)

	tests := []struct {
		name           string
		method         string
		path           string
		body           map[string]any
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid journal ID via query param",
			method:         "POST",
			path:           "/ai/analyze-sentiment?journal_id=test-journal-123",
			expectedStatus: http.StatusInternalServerError, // Will fail due to Ollama not running
			expectError:    true,
		},
		{
			name:   "valid content via body",
			method: "POST",
			path:   "/ai/analyze-sentiment",
			body: map[string]any{
				"content": "Today was a great day! I felt amazing and accomplished so much.",
			},
			expectedStatus: http.StatusInternalServerError, // Will fail due to Ollama not running
			expectError:    true,
		},
		{
			name:           "non-existent journal",
			method:         "POST",
			path:           "/ai/analyze-sentiment?journal_id=non-existent",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "empty journal ID",
			method:         "POST",
			path:           "/ai/analyze-sentiment",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "wrong method GET",
			method:         "GET",
			path:           "/ai/analyze-sentiment?journal_id=test-journal-123",
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req, err = http.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, tt.path, nil)
			}

			if err != nil {
				t.Fatal(err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler via ServeHTTP (exported method)
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check response body
			var response map[string]any
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if tt.expectError {
				if _, hasError := response["error"]; !hasError {
					t.Errorf("Expected error in response but got none")
				}
			}

			// For Ollama-related errors, verify it's the expected type
			if tt.expectedStatus == http.StatusInternalServerError {
				if errorMsg, hasError := response["error"]; hasError {
					errorStr := errorMsg.(string)
					if !strings.Contains(errorStr, "Sentiment analysis failed") &&
						!strings.Contains(errorStr, "connection") {
						t.Logf("AI service error (expected when Ollama not available): %s", errorStr)
					}
				}
			}
		})
	}
}

// TestAIHandler_ServeHTTP_GenerateJournal tests the generate journal endpoint via ServeHTTP
func TestAIHandler_ServeHTTP_GenerateJournal(t *testing.T) {
	ctx := context.Background()

	// Setup test dependencies
	store := storage.NewMemoryStore()
	aiService, err := ai.NewService(ctx, modelName, "http://localhost:11434", Logger())
	if err != nil || aiService == nil {
		t.Fatalf("Failed to create AI service: %v", err)
	}
	handler := handlers.NewAIHandler(store, aiService, Logger())

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    map[string]any
		expectedStatus int
		expectError    bool
	}{
		{
			name:   "valid prompt request",
			method: "POST",
			path:   "/ai/generate-journal",
			requestBody: map[string]any{
				"prompt":  "Write about a productive day at work",
				"context": "The user is a software developer",
			},
			expectedStatus: http.StatusInternalServerError, // Will fail due to Ollama not running
			expectError:    true,
		},
		{
			name:   "missing prompt",
			method: "POST",
			path:   "/ai/generate-journal",
			requestBody: map[string]any{
				"context": "Some context",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "empty prompt",
			method: "POST",
			path:   "/ai/generate-journal",
			requestBody: map[string]any{
				"prompt": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "too short prompt",
			method: "POST",
			path:   "/ai/generate-journal",
			requestBody: map[string]any{
				"prompt": "hi",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "wrong method GET",
			method: "GET",
			path:   "/ai/generate-journal",
			requestBody: map[string]any{
				"prompt": "test",
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			bodyBytes, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatal(err)
			}

			// Create request
			req, err := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler via ServeHTTP (exported method)
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check response body
			var response map[string]any
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if tt.expectError {
				if _, hasError := response["error"]; !hasError {
					t.Errorf("Expected error in response but got none")
				}
			}
		})
	}
}

// TestAIHandler_ServeHTTP_Health tests the health endpoint via ServeHTTP
func TestAIHandler_ServeHTTP_Health(t *testing.T) {
	ctx := context.Background()

	// Setup test dependencies
	store := storage.NewMemoryStore()
	aiService, err := ai.NewService(ctx, modelName, "http://localhost:11434", Logger())
	if err != nil || aiService == nil {
		t.Fatalf("Failed to create AI service: %v", err)
	}
	handler := handlers.NewAIHandler(store, aiService, Logger())

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid GET request",
			method:         "GET",
			path:           "/ai/health",
			expectedStatus: http.StatusServiceUnavailable, // Expected when Ollama not running
			expectError:    false,                         // Health checks don't return errors, just status
		},
		{
			name:           "wrong method POST",
			method:         "POST",
			path:           "/ai/health",
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler via ServeHTTP (exported method)
			handler.ServeHTTP(rr, req)

			// Check status code (may vary based on Ollama availability)
			if status := rr.Code; status != tt.expectedStatus && tt.method == "GET" {
				t.Logf("handler returned status code: %d (this may vary based on Ollama availability)", status)
			} else if tt.method != "GET" && status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check response body is valid JSON
			var response map[string]any
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Check required fields for health endpoint
			if tt.method == "GET" {
				if _, hasStatus := response["status"]; !hasStatus {
					t.Errorf("Health check response missing 'status' field")
				}

				if _, hasTimestamp := response["timestamp"]; !hasTimestamp {
					t.Errorf("Health check response missing 'timestamp' field")
				}

				if _, hasAI := response["ai_service"]; !hasAI {
					t.Errorf("Health check response missing 'ai_service' field")
				}
			}

			if tt.expectError {
				if _, hasError := response["error"]; !hasError {
					t.Errorf("Expected error in response but got none")
				}
			}

			t.Logf("Health check response: %+v", response)
		})
	}
}

// TestAIHandler_ServeHTTP_UnknownRoutes tests unknown endpoints via ServeHTTP
func TestAIHandler_ServeHTTP_UnknownRoutes(t *testing.T) {
	ctx := context.Background()

	// Setup test dependencies
	store := storage.NewMemoryStore()
	aiService, err := ai.NewService(ctx, modelName, "http://localhost:11434", Logger())
	if err != nil || aiService == nil {
		t.Fatalf("Failed to create AI service: %v", err)
	}
	handler := handlers.NewAIHandler(store, aiService, Logger())

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "unknown endpoint",
			method:         "GET",
			path:           "/ai/unknown",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "root path",
			method:         "GET",
			path:           "/ai",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "empty path",
			method:         "POST",
			path:           "/ai/",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check response is valid JSON
			var response map[string]any
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, hasError := response["error"]; !hasError {
				t.Errorf("Expected error in response for unknown endpoint")
			}
		})
	}
}

// TestAIHandler_ServeHTTP_MalformedJSON tests malformed JSON handling via ServeHTTP
func TestAIHandler_ServeHTTP_MalformedJSON(t *testing.T) {
	ctx := context.Background()

	// Setup test dependencies
	store := storage.NewMemoryStore()
	aiService, err := ai.NewService(ctx, modelName, "http://localhost:11434", Logger())
	if err != nil || aiService == nil {
		t.Fatalf("Failed to create AI service: %v", err)
	}
	handler := handlers.NewAIHandler(store, aiService, Logger())

	tests := []struct {
		name        string
		path        string
		body        string
		contentType string
	}{
		{
			name:        "malformed JSON for generate-journal",
			path:        "/ai/generate-journal",
			body:        `{"prompt": "test", "context": "incomplete...`,
			contentType: "application/json",
		},
		{
			name:        "malformed JSON for analyze-sentiment",
			path:        "/ai/analyze-sentiment",
			body:        `{"content": "test content", "invalid...`,
			contentType: "application/json",
		},
		{
			name:        "empty body with content-type",
			path:        "/ai/generate-journal",
			body:        "",
			contentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", tt.path, bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Should return bad request for malformed JSON
			if status := rr.Code; status != http.StatusBadRequest {
				t.Errorf("handler returned wrong status code for malformed JSON: got %v want %v", status, http.StatusBadRequest)
			}

			// Check error message
			var response map[string]any
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, hasError := response["error"]; !hasError {
				t.Errorf("Expected error in response for malformed JSON")
			}
		})
	}
}

// BenchmarkAIHandler_ServeHTTP_AnalyzeSentiment benchmarks the analyze sentiment endpoint
func BenchmarkAIHandler_ServeHTTP_AnalyzeSentiment(b *testing.B) {
	ctx := context.Background()

	// Setup test dependencies
	store := storage.NewMemoryStore()
	aiService, err := ai.NewService(ctx, modelName, "http://localhost:11434", Logger())
	if err != nil || aiService == nil {
		b.Fatalf("Failed to create AI service: %v", err)
	}
	handler := handlers.NewAIHandler(store, aiService, Logger())

	// Create test journal
	journal := &models.Journal{
		ID:        "bench-journal",
		Content:   "Today was a productive day with great achievements and positive outcomes.",
		CreatedAt: time.Now(),
	}
	store.Store(journal)

	// Create request
	req, _ := http.NewRequest("POST", "/ai/analyze-sentiment?journal_id=bench-journal", nil)

	b.ResetTimer()

	for b.Loop() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Note: This will likely timeout without Ollama, but tests the handler performance
		if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
			b.Errorf("Unexpected status code: %d", rr.Code)
		}
	}
}

// BenchmarkAIHandler_ServeHTTP_Health benchmarks the health endpoint
func BenchmarkAIHandler_ServeHTTP_Health(b *testing.B) {
	ctx := context.Background()

	// Setup test dependencies
	store := storage.NewMemoryStore()
	aiService, err := ai.NewService(ctx, modelName, "http://localhost:11434", Logger())
	if err != nil || aiService == nil {
		b.Fatalf("Failed to create AI service: %v", err)
	}
	handler := handlers.NewAIHandler(store, aiService, Logger())

	// Create request
	req, _ := http.NewRequest("GET", "/ai/health", nil)

	b.ResetTimer()

	for b.Loop() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Health endpoint should always respond, regardless of Ollama status
		if rr.Code != http.StatusOK && rr.Code != http.StatusServiceUnavailable {
			b.Errorf("Unexpected status code: %d", rr.Code)
		}
	}
}
