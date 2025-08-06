package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/ai"
	"github.com/garnizeh/englog/internal/handlers"
	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/storage"
)

func TestHealthHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		setupData      func(*storage.MemoryStore)
	}{
		{
			name:           "valid GET request",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			setupData:      nil,
		},
		{
			name:           "GET request with journal data",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			setupData: func(s *storage.MemoryStore) {
				// Add some test journals
				journal1 := &models.Journal{
					ID:        "test-1",
					Content:   "Test journal 1",
					CreatedAt: time.Now(),
				}
				journal2 := &models.Journal{
					ID:        "test-2",
					Content:   "Test journal 2",
					CreatedAt: time.Now(),
				}
				s.Store(journal1)
				s.Store(journal2)
			},
		},
		{
			name:           "POST method not allowed",
			method:         "POST",
			path:           "/health",
			expectedStatus: http.StatusMethodNotAllowed,
			setupData:      nil,
		},
		{
			name:           "PUT method not allowed",
			method:         "PUT",
			path:           "/health",
			expectedStatus: http.StatusMethodNotAllowed,
			setupData:      nil,
		},
		{
			name:           "DELETE method not allowed",
			method:         "DELETE",
			path:           "/health",
			expectedStatus: http.StatusMethodNotAllowed,
			setupData:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test store and AI service
			testStore := storage.NewMemoryStore()
			mockAI := ai.NewMockAIProvider()
			testHandler := handlers.NewHealthHandler(testStore, mockAI, Logger())

			if tt.setupData != nil {
				tt.setupData(testStore)
			}

			// Create request
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			testHandler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// For successful requests, validate response structure
			if tt.expectedStatus == http.StatusOK {
				// Check Content-Type header
				expectedContentType := "application/json"
				if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
					t.Errorf("handler returned wrong content type: got %v want %v", contentType, expectedContentType)
				}

				// Parse response body
				var response map[string]any
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}

				// Validate required fields
				requiredFields := []string{"status", "timestamp", "service", "version", "storage"}
				for _, field := range requiredFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Response missing required field: %s", field)
					}
				}

				// Validate status field
				if status, ok := response["status"].(string); !ok || status != "healthy" {
					t.Errorf("Expected status 'healthy', got: %v", response["status"])
				}

				// Validate service field
				if service, ok := response["service"].(string); !ok || service != "englog-api" {
					t.Errorf("Expected service 'englog-api', got: %v", response["service"])
				}

				// Validate version field
				if version, ok := response["version"].(string); !ok || version != "prototype-009" {
					t.Errorf("Expected version 'prototype-009', got: %v", response["version"])
				}

				// Validate timestamp field
				if timestampStr, ok := response["timestamp"].(string); ok {
					_, err := time.Parse(time.RFC3339, timestampStr)
					if err != nil {
						t.Errorf("Invalid timestamp format: %v", timestampStr)
					}
				} else {
					t.Errorf("Expected timestamp to be a string, got: %T", response["timestamp"])
				}

				// Validate storage field
				if storage, ok := response["storage"].(map[string]any); ok {
					// Check storage type
					if storageType, ok := storage["type"].(string); !ok || storageType != "memory" {
						t.Errorf("Expected storage type 'memory', got: %v", storage["type"])
					}

					// Check journal count
					expectedCount := testStore.Count()
					if journalCount, ok := storage["journal_count"].(float64); ok {
						if int(journalCount) != expectedCount {
							t.Errorf("Expected journal count %d, got: %v", expectedCount, journalCount)
						}
					} else {
						t.Errorf("Expected journal_count to be a number, got: %T", storage["journal_count"])
					}
				} else {
					t.Errorf("Expected storage to be an object, got: %T", response["storage"])
				}
			}

			// For method not allowed, check error message
			// Validate error responses for method not allowed
			if tt.expectedStatus == http.StatusMethodNotAllowed {
				var errorResponse map[string]any
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
					t.Fatalf("Failed to parse error response JSON: %v", err)
				}

				if errorMsg, ok := errorResponse["error"].(string); !ok || errorMsg != "Method not allowed" {
					t.Errorf("Expected error message 'Method not allowed', got: %v", errorResponse["error"])
				}

				if status, ok := errorResponse["status"].(float64); !ok || int(status) != tt.expectedStatus {
					t.Errorf("Expected status %d in error response, got: %v", tt.expectedStatus, errorResponse["status"])
				}
			}
		})
	}
}

func TestHealthHandler_ResponseStructure(t *testing.T) {
	// Test response structure with various store states
	store := storage.NewMemoryStore()
	mockAI := ai.NewMockAIProvider()
	handler := handlers.NewHealthHandler(store, mockAI, Logger())

	// Test with empty store
	t.Run("empty store", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		var response map[string]any
		json.Unmarshal(rr.Body.Bytes(), &response)

		storage := response["storage"].(map[string]any)
		if journalCount := storage["journal_count"].(float64); journalCount != 0 {
			t.Errorf("Expected journal count 0 for empty store, got: %v", journalCount)
		}
	})

	// Test with populated store
	t.Run("populated store", func(t *testing.T) {
		// Add test data
		for i := 0; i < 5; i++ {
			journal := &models.Journal{
				ID:        string(rune('a' + i)),
				Content:   "Test content",
				CreatedAt: time.Now(),
			}
			store.Store(journal)
		}

		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		var response map[string]any
		json.Unmarshal(rr.Body.Bytes(), &response)

		storage := response["storage"].(map[string]any)
		if journalCount := storage["journal_count"].(float64); journalCount != 5 {
			t.Errorf("Expected journal count 5, got: %v", journalCount)
		}
	})
}

func TestHealthHandler_ConcurrentRequests(t *testing.T) {
	// Test handler behavior under concurrent requests
	store := storage.NewMemoryStore()
	mockAI := ai.NewMockAIProvider()
	handler := handlers.NewHealthHandler(store, mockAI, Logger())

	// Add some test data
	for i := 0; i < 10; i++ {
		journal := &models.Journal{
			ID:        string(rune('a' + i)),
			Content:   "Concurrent test content",
			CreatedAt: time.Now(),
		}
		store.Store(journal)
	}

	// Make concurrent requests
	numRequests := 10
	responses := make(chan *httptest.ResponseRecorder, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req, _ := http.NewRequest("GET", "/health", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			responses <- rr
		}()
	}

	// Collect and validate all responses
	for i := 0; i < numRequests; i++ {
		rr := <-responses

		if rr.Code != http.StatusOK {
			t.Errorf("Concurrent request %d failed with status: %d", i, rr.Code)
		}

		var response map[string]any
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse concurrent response %d: %v", i, err)
		}

		// All responses should have the same journal count
		storage := response["storage"].(map[string]any)
		if journalCount := storage["journal_count"].(float64); journalCount != 10 {
			t.Errorf("Concurrent request %d returned wrong journal count: %v", i, journalCount)
		}
	}
}

func TestHealthHandler_NewHealthHandler(t *testing.T) {
	// Test handler creation
	store := storage.NewMemoryStore()
	mockAI := ai.NewMockAIProvider()
	handler := handlers.NewHealthHandler(store, mockAI, Logger())

	if handler == nil {
		t.Fatal("NewHealthHandler returned nil")
	}
}

// Benchmark test for health endpoint performance
func BenchmarkHealthHandler_ServeHTTP(b *testing.B) {
	store := storage.NewMemoryStore()
	mockAI := ai.NewMockAIProvider()
	handler := handlers.NewHealthHandler(store, mockAI, Logger())

	// Add some test data
	for i := 0; i < 100; i++ {
		journal := &models.Journal{
			ID:        string(rune('a'+i%26)) + string(rune('0'+i/26)),
			Content:   "Benchmark test content",
			CreatedAt: time.Now(),
		}
		store.Store(journal)
	}

	req, _ := http.NewRequest("GET", "/health", nil)

	b.ResetTimer()

	for b.Loop() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Errorf("Unexpected status code: %d", rr.Code)
		}
	}
}

// Tests for new endpoints in PROTOTYPE-009

func TestHealthHandler_StatusEndpoint(t *testing.T) {
	store := storage.NewMemoryStore()
	mockAI := ai.NewMockAIProvider()
	handler := handlers.NewHealthHandler(store, mockAI, Logger())

	// Add some test journals with processing results
	processingTime1 := 150 * time.Millisecond
	processingTime2 := 200 * time.Millisecond
	processedAt1 := time.Now().Add(-30 * time.Minute)
	processedAt2 := time.Now().Add(-15 * time.Minute)

	journal1 := &models.Journal{
		ID:               "test-1",
		Content:          "Test journal 1",
		CreatedAt:        time.Now().Add(-1 * time.Hour),
		ProcessingStatus: models.ProcessingStatusCompleted,
		ProcessingResult: &models.ProcessingResult{
			Status:         models.ProcessingStatusCompleted,
			ProcessingTime: &processingTime1,
			ProcessedAt:    &processedAt1,
		},
	}
	journal2 := &models.Journal{
		ID:               "test-2",
		Content:          "Test journal 2",
		CreatedAt:        time.Now().Add(-30 * time.Minute),
		ProcessingStatus: models.ProcessingStatusCompleted,
		ProcessingResult: &models.ProcessingResult{
			Status:         models.ProcessingStatusCompleted,
			ProcessingTime: &processingTime2,
			ProcessedAt:    &processedAt2,
		},
	}
	store.Store(journal1)
	store.Store(journal2)

	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify response structure
	expectedFields := []string{"status", "timestamp", "service", "version", "uptime_seconds", "uptime_human", "memory", "storage", "response_time_ms"}
	for _, field := range expectedFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Response missing required field: %s", field)
		}
	}

	// Verify memory information
	memory, ok := response["memory"].(map[string]any)
	if !ok {
		t.Fatal("Memory field is not a map")
	}
	memoryFields := []string{"allocated_bytes", "allocated_mb", "total_allocated_bytes", "total_allocated_mb", "heap_objects", "gc_cycles"}
	for _, field := range memoryFields {
		if _, exists := memory[field]; !exists {
			t.Errorf("Memory section missing field: %s", field)
		}
	}

	// Verify storage information
	storageInfo, ok := response["storage"].(map[string]any)
	if !ok {
		t.Fatal("Storage field is not a map")
	}
	storageFields := []string{"type", "journal_count", "processed_count", "avg_processing_time_ms"}
	for _, field := range storageFields {
		if _, exists := storageInfo[field]; !exists {
			t.Errorf("Storage section missing field: %s", field)
		}
	}

	// Verify journal statistics
	if journalCount := storageInfo["journal_count"]; journalCount != float64(2) {
		t.Errorf("Wrong journal count: got %v want 2", journalCount)
	}
	if processedCount := storageInfo["processed_count"]; processedCount != float64(2) {
		t.Errorf("Wrong processed count: got %v want 2", processedCount)
	}
}

func TestHealthHandler_OllamaStatusEndpoint(t *testing.T) {
	tests := []struct {
		name            string
		healthCheckErr  error
		expectedStatus  int
		expectedHealthy bool
	}{
		{
			name:            "healthy Ollama",
			healthCheckErr:  nil,
			expectedStatus:  http.StatusOK,
			expectedHealthy: true,
		},
		{
			name:            "unhealthy Ollama",
			healthCheckErr:  fmt.Errorf("connection failed"),
			expectedStatus:  http.StatusServiceUnavailable,
			expectedHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStore()
			mockAI := ai.NewMockAIProvider()

			// Set up mock behavior
			mockAI.HealthCheckFunc = func(ctx context.Context) error {
				return tt.healthCheckErr
			}

			handler := handlers.NewHealthHandler(store, mockAI, Logger())

			req, err := http.NewRequest("GET", "/status/ollama", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Parse response
			var response map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response JSON: %v", err)
			}

			// Verify response structure
			expectedFields := []string{"status", "timestamp", "service", "connected", "response_time_ms"}
			for _, field := range expectedFields {
				if _, exists := response[field]; !exists {
					t.Errorf("Response missing required field: %s", field)
				}
			}

			// Verify connection status
			if connected := response["connected"]; connected != tt.expectedHealthy {
				t.Errorf("Wrong connection status: got %v want %v", connected, tt.expectedHealthy)
			}

			// Verify error field presence when unhealthy
			if !tt.expectedHealthy {
				if _, exists := response["error"]; !exists {
					t.Error("Response missing error field when unhealthy")
				}
			}
		})
	}
}

func TestHealthHandler_UnsupportedPaths(t *testing.T) {
	store := storage.NewMemoryStore()
	mockAI := ai.NewMockAIProvider()
	handler := handlers.NewHealthHandler(store, mockAI, Logger())

	unsupportedPaths := []string{"/status/unknown", "/health/extra", "/invalid"}

	for _, path := range unsupportedPaths {
		t.Run(path, func(t *testing.T) {
			req, err := http.NewRequest("GET", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Should return 404 for unsupported paths
			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("handler returned wrong status code for %s: got %v want %v", path, status, http.StatusNotFound)
			}
		})
	}
}
