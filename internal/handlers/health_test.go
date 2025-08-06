package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
			// Setup test store
			testStore := storage.NewMemoryStore()
			testHandler := handlers.NewHealthHandler(testStore, Logger())

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
				if version, ok := response["version"].(string); !ok || version != "prototype-006" {
					t.Errorf("Expected version 'prototype-006', got: %v", response["version"])
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
			if tt.expectedStatus == http.StatusMethodNotAllowed {
				responseBody := rr.Body.String()
				if responseBody != "Method not allowed\n" {
					t.Errorf("Expected 'Method not allowed' error message, got: %s", responseBody)
				}
			}
		})
	}
}

func TestHealthHandler_ResponseStructure(t *testing.T) {
	// Test response structure with various store states
	store := storage.NewMemoryStore()
	handler := handlers.NewHealthHandler(store, Logger())

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
	handler := handlers.NewHealthHandler(store, Logger())

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
	handler := handlers.NewHealthHandler(store, Logger())

	if handler == nil {
		t.Fatal("NewHealthHandler returned nil")
	}
}

// Benchmark test for health endpoint performance
func BenchmarkHealthHandler_ServeHTTP(b *testing.B) {
	store := storage.NewMemoryStore()
	handler := handlers.NewHealthHandler(store, Logger())

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
