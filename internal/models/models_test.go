package models_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/garnizeh/englog/internal/models"
)

// Test ValidationError (exported type)
func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      models.ValidationError
		expected string
	}{
		{
			name: "basic validation error",
			err: models.ValidationError{
				Field:   "content",
				Message: "Content is required",
				Code:    "REQUIRED",
			},
			expected: "validation error in field 'content': Content is required",
		},
		{
			name: "validation error without code",
			err: models.ValidationError{
				Field:   "email",
				Message: "Invalid email format",
			},
			expected: "validation error in field 'email': Invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test ValidationErrors (exported type)
func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   models.ValidationErrors
		expected string
	}{
		{
			name:     "empty errors",
			errors:   models.ValidationErrors{},
			expected: "no validation errors",
		},
		{
			name: "single error",
			errors: models.ValidationErrors{
				{Field: "content", Message: "Content is required", Code: "REQUIRED"},
			},
			expected: "validation error in field 'content': Content is required",
		},
		{
			name: "multiple errors",
			errors: models.ValidationErrors{
				{Field: "content", Message: "Content is required", Code: "REQUIRED"},
				{Field: "metadata", Message: "Invalid metadata", Code: "INVALID_FORMAT"},
			},
			expected: "validation error in field 'content': Content is required and 1 more validation errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   models.ValidationErrors
		expected bool
	}{
		{
			name:     "empty errors",
			errors:   models.ValidationErrors{},
			expected: false,
		},
		{
			name: "has errors",
			errors: models.ValidationErrors{
				{Field: "content", Message: "Content is required", Code: "REQUIRED"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.HasErrors()
			if result != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidationErrors_ToJSON(t *testing.T) {
	tests := []struct {
		name     string
		errors   models.ValidationErrors
		expected string
	}{
		{
			name:     "empty errors",
			errors:   models.ValidationErrors{},
			expected: `[]`,
		},
		{
			name: "single error",
			errors: models.ValidationErrors{
				{Field: "content", Message: "Content is required", Code: "REQUIRED"},
			},
			expected: `[{"field":"content","message":"Content is required","code":"REQUIRED"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.ToJSON()
			if string(result) != tt.expected {
				t.Errorf("ToJSON() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test CreateJournalRequest (exported type)
func TestCreateJournalRequest_Validate(t *testing.T) {
	tests := []struct {
		name                  string
		request               models.CreateJournalRequest
		expectedHasErrors     bool
		expectedErrorsContain []string
	}{
		{
			name: "valid request - minimal",
			request: models.CreateJournalRequest{
				Content: "This is valid content",
			},
			expectedHasErrors: false,
		},
		{
			name: "valid request - complete",
			request: models.CreateJournalRequest{
				Content: "This is valid content with good length",
				Metadata: map[string]interface{}{
					"mood":     8,
					"location": "home",
					"weather":  "sunny",
				},
			},
			expectedHasErrors: false,
		},
		{
			name: "invalid - empty content",
			request: models.CreateJournalRequest{
				Content: "",
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Content is required"},
		},
		{
			name: "invalid - whitespace only content",
			request: models.CreateJournalRequest{
				Content: "   \n\t   ",
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Content cannot be only whitespace"},
		},
		{
			name: "invalid - content too short",
			request: models.CreateJournalRequest{
				Content: "Hi",
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Content must be at least 10 characters long"},
		},
		{
			name: "invalid - content too long",
			request: models.CreateJournalRequest{
				Content: strings.Repeat("a", 50001),
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Content exceeds maximum length"},
		},
		{
			name: "invalid - metadata key too long",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					strings.Repeat("a", 101): "value",
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"exceeds maximum length of 100 characters"},
		},
		{
			name: "invalid - metadata string value too long",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"key": strings.Repeat("a", 1001),
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"exceeds maximum length of 1000 characters"},
		},
		{
			name: "invalid - unsupported metadata type",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"key": make(chan int),
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"has unsupported type"},
		},
		{
			name: "invalid - too many metadata fields",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: func() map[string]interface{} {
					m := make(map[string]interface{})
					for i := 0; i < 21; i++ {
						m[fmt.Sprintf("key%d", i)] = "value"
					}
					return m
				}(),
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Metadata cannot have more than 20 fields"},
		},
		{
			name: "valid - metadata with arrays",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"tags":   []interface{}{"work", "productivity"},
					"scores": []interface{}{8, 9, 7},
					"flags":  []interface{}{true, false, true},
					"nulls":  []interface{}{nil, "value", nil},
				},
			},
			expectedHasErrors: false,
		},
		{
			name: "valid - metadata with nested objects",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"location": map[string]interface{}{
						"city":    "San Francisco",
						"country": "USA",
						"lat":     37.7749,
						"lng":     -122.4194,
					},
				},
			},
			expectedHasErrors: false,
		},
		{
			name: "invalid - metadata array too many items",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"big_array": func() []interface{} {
						arr := make([]interface{}, 51)
						for i := range arr {
							arr[i] = fmt.Sprintf("item%d", i)
						}
						return arr
					}(),
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"cannot have more than 50 elements"},
		},
		{
			name: "invalid - metadata array item too long",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"tags": []interface{}{strings.Repeat("a", 501)},
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"exceeds maximum length of 500 characters"},
		},
		{
			name: "invalid - metadata array with unsupported type",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"tags": []interface{}{"valid", []interface{}{"nested", "array"}},
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"has unsupported type"},
		},
		{
			name: "invalid - metadata object too many fields",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"location": func() map[string]interface{} {
						obj := make(map[string]interface{})
						for i := 0; i < 11; i++ {
							obj[fmt.Sprintf("field%d", i)] = "value"
						}
						return obj
					}(),
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"cannot have more than 10 fields"},
		},
		{
			name: "invalid - empty metadata key",
			request: models.CreateJournalRequest{
				Content: "Valid content here",
				Metadata: map[string]interface{}{
					"": "value",
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Metadata keys cannot be empty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.request.Validate()

			hasErrors := errors.HasErrors()
			if hasErrors != tt.expectedHasErrors {
				t.Errorf("HasErrors() = %v, want %v", hasErrors, tt.expectedHasErrors)
			}

			if tt.expectedHasErrors {
				errorStr := errors.Error()
				for _, expectedCode := range tt.expectedErrorsContain {
					if !strings.Contains(errorStr, expectedCode) {
						t.Errorf("Expected error to contain %q, but got: %q", expectedCode, errorStr)
					}
				}
			}
		})
	}
}

// Test PromptRequest (exported type)
func TestPromptRequest_Validate(t *testing.T) {
	tests := []struct {
		name                  string
		request               models.PromptRequest
		expectedHasErrors     bool
		expectedErrorsContain []string
	}{
		{
			name: "valid request",
			request: models.PromptRequest{
				Prompt: "What was my mood like this week?",
			},
			expectedHasErrors: false,
		},
		{
			name: "invalid - empty prompt",
			request: models.PromptRequest{
				Prompt: "",
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Prompt is required"},
		},
		{
			name: "invalid - whitespace only prompt",
			request: models.PromptRequest{
				Prompt: "   \n\t   ",
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Prompt cannot be only whitespace"},
		},
		{
			name: "invalid - prompt too short",
			request: models.PromptRequest{
				Prompt: "Hi",
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Prompt must be at least 3 characters long"},
		},
		{
			name: "invalid - prompt too long",
			request: models.PromptRequest{
				Prompt: strings.Repeat("a", 2001),
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Prompt exceeds maximum length of 2,000 characters"},
		},
		{
			name: "valid request with context",
			request: models.PromptRequest{
				Prompt:  "What was my mood like this week?",
				Context: "I've been working on mindfulness practices lately",
			},
			expectedHasErrors: false,
		},
		{
			name: "invalid - context too long",
			request: models.PromptRequest{
				Prompt:  "Valid prompt here",
				Context: strings.Repeat("a", 5001),
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Context exceeds maximum length of 5,000 characters"},
		},
		{
			name: "valid request with metadata",
			request: models.PromptRequest{
				Prompt: "Write about gratitude",
				Metadata: map[string]interface{}{
					"mood_preference": "positive",
					"length":          "medium",
					"style":           "reflective",
				},
			},
			expectedHasErrors: false,
		},
		{
			name: "invalid - too many metadata fields",
			request: models.PromptRequest{
				Prompt: "Valid prompt here",
				Metadata: func() map[string]interface{} {
					m := make(map[string]interface{})
					for i := 0; i < 11; i++ {
						m[fmt.Sprintf("key%d", i)] = "value"
					}
					return m
				}(),
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Metadata cannot have more than 10 fields"},
		},
		{
			name: "invalid - empty metadata key",
			request: models.PromptRequest{
				Prompt: "Valid prompt here",
				Metadata: map[string]interface{}{
					"": "value",
				},
			},
			expectedHasErrors:     true,
			expectedErrorsContain: []string{"Metadata keys cannot be empty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.request.Validate()

			hasErrors := errors.HasErrors()
			if hasErrors != tt.expectedHasErrors {
				t.Errorf("HasErrors() = %v, want %v", hasErrors, tt.expectedHasErrors)
			}

			if tt.expectedHasErrors {
				errorStr := errors.Error()
				for _, expectedCode := range tt.expectedErrorsContain {
					if !strings.Contains(errorStr, expectedCode) {
						t.Errorf("Expected error to contain %q, but got: %q", expectedCode, errorStr)
					}
				}
			}
		})
	}
}

// Test exported constants and types for proper JSON marshaling
func TestJSONMarshaling(t *testing.T) {
	t.Run("ProcessingStatus values", func(t *testing.T) {
		statuses := []models.ProcessingStatus{
			models.ProcessingStatusPending,
			models.ProcessingStatusProcessing,
			models.ProcessingStatusCompleted,
			models.ProcessingStatusFailed,
		}

		for _, status := range statuses {
			data, err := json.Marshal(status)
			if err != nil {
				t.Errorf("Failed to marshal ProcessingStatus %q: %v", status, err)
			}

			var unmarshaled models.ProcessingStatus
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("Failed to unmarshal ProcessingStatus %q: %v", status, err)
			}

			if unmarshaled != status {
				t.Errorf("Marshaling roundtrip failed: got %q, want %q", unmarshaled, status)
			}
		}
	})

	t.Run("Journal struct", func(t *testing.T) {
		journal := models.Journal{
			ID:      "test-id",
			Content: "Test content",
			Metadata: map[string]interface{}{
				"mood":     8,
				"location": "home",
			},
			ProcessingStatus: models.ProcessingStatusCompleted,
		}

		data, err := json.Marshal(journal)
		if err != nil {
			t.Errorf("Failed to marshal Journal: %v", err)
		}

		var unmarshaled models.Journal
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal Journal: %v", err)
		}

		if unmarshaled.ID != journal.ID {
			t.Errorf("ID mismatch: got %q, want %q", unmarshaled.ID, journal.ID)
		}
		if unmarshaled.Content != journal.Content {
			t.Errorf("Content mismatch: got %q, want %q", unmarshaled.Content, journal.Content)
		}
		if unmarshaled.ProcessingStatus != journal.ProcessingStatus {
			t.Errorf("ProcessingStatus mismatch: got %q, want %q", unmarshaled.ProcessingStatus, journal.ProcessingStatus)
		}
	})
}

// Test edge cases for exported validation
func TestValidationEdgeCases(t *testing.T) {
	t.Run("unicode content handling", func(t *testing.T) {
		// Test with various unicode characters
		unicodeContent := "Today I felt 😊 happy! 🌟 Testing unicode: 你好世界 здравствуй мир"

		request := models.CreateJournalRequest{
			Content: unicodeContent,
		}

		errors := request.Validate()
		if errors.HasErrors() {
			t.Errorf("Valid unicode content should not produce errors, got: %v", errors)
		}

		// Verify the content length is calculated correctly with UTF-8
		if utf8.RuneCountInString(unicodeContent) < 10 {
			t.Errorf("Unicode content should be long enough to pass validation")
		}
	})

	t.Run("boundary length validation", func(t *testing.T) {
		// Test exactly at boundaries
		exactMinContent := strings.Repeat("a", 10) // Exactly min length
		request := models.CreateJournalRequest{Content: exactMinContent}
		errors := request.Validate()
		if errors.HasErrors() {
			t.Errorf("Content at exactly min length should be valid, got: %v", errors)
		}

		exactMaxContent := strings.Repeat("a", 50000) // Exactly max length
		request = models.CreateJournalRequest{Content: exactMaxContent}
		errors = request.Validate()
		if errors.HasErrors() {
			t.Errorf("Content at exactly max length should be valid, got: %v", errors)
		}
	})
}
