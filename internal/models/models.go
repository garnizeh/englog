package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// ProcessingStatus represents the status of AI processing
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

// ProcessingResult contains the results of AI processing for a journal entry
// Schema: Complete AI analysis results including sentiment, timing, and error information
type ProcessingResult struct {
	// Status indicates the current state of AI processing
	Status ProcessingStatus `json:"status" example:"completed"`

	// SentimentResult contains sentiment analysis if processing was successful
	SentimentResult *SentimentResult `json:"sentiment_result,omitempty"`

	// ProcessedAt timestamp when AI processing was completed (only set if successful)
	ProcessedAt *time.Time `json:"processed_at,omitempty" example:"2025-08-05T10:30:20Z"`

	// ProcessingTime duration taken for AI processing (only set if completed)
	ProcessingTime *time.Duration `json:"processing_time,omitempty" example:"2.5s"`

	// Error contains error message if processing failed (only set if status is "failed")
	Error string `json:"error,omitempty" example:"AI service temporarily unavailable"`
}

// Journal represents a journal entry in the system
// Schema: Represents the complete state of a journal entry including content,
// metadata, timestamps, and optional AI processing results.
type Journal struct {
	// ID is a unique identifier for the journal entry (UUID v4 format)
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Content is the main text content of the journal entry
	// Must be between 1 and 50,000 characters
	Content string `json:"content" example:"Today was a wonderful day filled with new experiences..."`

	// ProcessingStatus indicates the current state of AI processing for this journal entry
	ProcessingStatus ProcessingStatus `json:"processing_status" example:"completed"`

	// Timestamp represents when the journal entry was originally created by the user
	// This is different from CreatedAt which represents when it was stored in the system
	Timestamp time.Time `json:"timestamp" example:"2025-08-05T10:30:00Z"`

	// CreatedAt represents when the journal entry was created in the system
	CreatedAt time.Time `json:"created_at" example:"2025-08-05T10:30:15Z"`

	// UpdatedAt represents when the journal entry was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2025-08-05T10:30:15Z"`

	// Metadata contains additional structured data associated with the journal entry
	// Can include mood ratings, tags, location data, etc.
	// Maximum 20 fields, each key max 100 chars, each string value max 1000 chars
	Metadata map[string]any `json:"metadata,omitempty" example:"{\"mood\": 8, \"tags\": [\"work\", \"productivity\"], \"location\": \"home\"}"`

	// ProcessingResult contains AI analysis results if processing has been completed
	ProcessingResult *ProcessingResult `json:"processing_result,omitempty"`
}

// CreateJournalRequest represents the request body for creating a journal
// Schema: Defines the required and optional fields for creating a new journal entry
type CreateJournalRequest struct {
	// Content is the main text content of the journal entry
	// Required field, must be between 1 and 50,000 characters after trimming whitespace
	Content string `json:"content" binding:"required" example:"Today I learned something amazing about Go programming..."`

	// Metadata contains additional structured data for the journal entry
	// Optional field, maximum 20 fields allowed
	// Supported value types: string, number, boolean, null, array (flat), object (one level deep)
	Metadata map[string]any `json:"metadata,omitempty" example:"{\"mood\": 7, \"tags\": [\"learning\", \"tech\"], \"location\": \"office\"}"`
}

// SentimentResult represents the result of sentiment analysis
// Schema: Structured sentiment analysis with score, label, confidence, and timing
type SentimentResult struct {
	// Score represents sentiment polarity from -1.0 (very negative) to 1.0 (very positive)
	// 0.0 represents neutral sentiment
	Score float64 `json:"score" example:"0.75"`

	// Label provides a human-readable sentiment classification
	// Valid values: "positive", "negative", "neutral"
	Label string `json:"label" example:"positive" enum:"positive,negative,neutral"`

	// Confidence represents the AI model's confidence in this sentiment analysis
	// Range: 0.0 (no confidence) to 1.0 (maximum confidence)
	Confidence float64 `json:"confidence" example:"0.92"`

	// ProcessedAt timestamp when sentiment analysis was performed
	ProcessedAt time.Time `json:"processed_at" example:"2025-08-05T10:30:18Z"`
}

// GeneratedJournal represents an AI-generated journal entry
type GeneratedJournal struct {
	Content         string            `json:"content"`          // Structured text optimized for semantic analysis
	Metadata        GeneratedMetadata `json:"metadata"`         // Comprehensive metadata
	SemanticMarkers []string          `json:"semantic_markers"` // Prepared for future embedding generation
	ProcessingHints map[string]any    `json:"processing_hints"` // Optimization flags for Phase 2 vectorization
	GeneratedAt     time.Time         `json:"generated_at"`
}

// GeneratedMetadata contains comprehensive metadata for generated journal entries
type GeneratedMetadata struct {
	Mood             string   `json:"mood"`              // Overall mood assessment
	EmotionalContext string   `json:"emotional_context"` // Detailed emotional state
	Themes           []string `json:"themes"`            // Main themes identified
	Entities         []string `json:"entities"`          // People, places, objects mentioned
	KeyPhrases       []string `json:"key_phrases"`       // Important phrases for semantic analysis
	Tags             []string `json:"tags"`              // Categorization tags
}

// PromptRequest represents a request to generate a journal entry from a prompt
// Schema: Defines the structure for AI-assisted journal generation requests
type PromptRequest struct {
	// Prompt is the main input text for generating a journal entry
	// Required field, must be between 3 and 2,000 characters
	Prompt string `json:"prompt" binding:"required" example:"Write about a day when I felt grateful"`

	// Context provides additional background information for better generation
	// Optional field, maximum 5,000 characters
	Context string `json:"context,omitempty" example:"I've been working on mindfulness practices lately"`

	// Metadata contains hints and preferences for journal generation
	// Optional field, maximum 10 fields allowed
	Metadata map[string]any `json:"metadata,omitempty" example:"{\"mood_preference\": \"positive\", \"length\": \"medium\"}"`
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	if len(ves) == 1 {
		return ves[0].Error()
	}
	return fmt.Sprintf("%s and %d more validation errors", ves[0].Error(), len(ves)-1)
}

// HasErrors checks if there are any validation errors
func (ves ValidationErrors) HasErrors() bool {
	return len(ves) > 0
}

// ToJSON converts validation errors to JSON format
func (ves ValidationErrors) ToJSON() []byte {
	if len(ves) == 0 {
		return []byte("[]")
	}
	data, _ := json.Marshal(ves)
	return data
}

// Validate validates a CreateJournalRequest
func (req *CreateJournalRequest) Validate() ValidationErrors {
	var errors ValidationErrors

	// Validate content
	if req.Content == "" {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: "Content is required and cannot be empty",
			Code:    "REQUIRED",
		})
	} else {
		trimmed := strings.TrimSpace(req.Content)
		if trimmed == "" {
			errors = append(errors, ValidationError{
				Field:   "content",
				Message: "Content cannot be only whitespace",
				Code:    "INVALID_FORMAT",
			})
		} else if utf8.RuneCountInString(trimmed) > 50000 {
			errors = append(errors, ValidationError{
				Field:   "content",
				Message: "Content exceeds maximum length of 50,000 characters",
				Code:    "MAX_LENGTH_EXCEEDED",
			})
		} else if utf8.RuneCountInString(trimmed) < 10 {
			errors = append(errors, ValidationError{
				Field:   "content",
				Message: "Content must be at least 10 characters long",
				Code:    "MIN_LENGTH_NOT_MET",
			})
		}
	}

	// Validate metadata if provided
	if req.Metadata != nil {
		if len(req.Metadata) > 20 {
			errors = append(errors, ValidationError{
				Field:   "metadata",
				Message: "Metadata cannot have more than 20 fields",
				Code:    "TOO_MANY_FIELDS",
			})
		}

		for key, value := range req.Metadata {
			if key == "" {
				errors = append(errors, ValidationError{
					Field:   "metadata",
					Message: "Metadata keys cannot be empty",
					Code:    "INVALID_KEY",
				})
				continue
			}

			if utf8.RuneCountInString(key) > 100 {
				errors = append(errors, ValidationError{
					Field:   "metadata",
					Message: fmt.Sprintf("Metadata key '%s' exceeds maximum length of 100 characters", key),
					Code:    "INVALID_KEY",
				})
			}

			// Validate value type and size
			if err := validateMetadataValue(key, value); err != nil {
				errors = append(errors, ValidationError{
					Field:   "metadata",
					Message: err.Error(),
					Code:    "INVALID_VALUE",
				})
			}
		}
	}

	return errors
}

// validateMetadataValue validates metadata values
func validateMetadataValue(key string, value any) error {
	switch v := value.(type) {
	case string:
		if utf8.RuneCountInString(v) > 1000 {
			return fmt.Errorf("metadata value for key '%s' exceeds maximum length of 1000 characters", key)
		}
	case float64, int, int32, int64:
		// JSON numbers are always float64, but we also accept Go integer types
		// No additional validation needed for numeric values
	case bool:
		// Booleans are always valid
	case nil:
		// Null values are allowed
	case []any:
		if len(v) > 50 {
			return fmt.Errorf("metadata array for key '%s' cannot have more than 50 elements", key)
		}
		for i, item := range v {
			if err := validateMetadataArrayItem(key, i, item); err != nil {
				return err
			}
		}
	case map[string]any:
		if len(v) > 10 {
			return fmt.Errorf("metadata object for key '%s' cannot have more than 10 fields", key)
		}
		for subKey, subValue := range v {
			if err := validateMetadataValue(fmt.Sprintf("%s.%s", key, subKey), subValue); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("metadata value for key '%s' has unsupported type", key)
	}
	return nil
}

// validateMetadataArrayItem validates items in metadata arrays
func validateMetadataArrayItem(key string, index int, value any) error {
	switch v := value.(type) {
	case string:
		if utf8.RuneCountInString(v) > 500 {
			return fmt.Errorf("metadata array item %d for key '%s' exceeds maximum length of 500 characters", index, key)
		}
	case float64, int, int32, int64, bool, nil:
		// These types are always valid in arrays
	default:
		return fmt.Errorf("metadata array item %d for key '%s' has unsupported type (nested arrays/objects not allowed)", index, key)
	}
	return nil
}

// Validate validates a PromptRequest
func (req *PromptRequest) Validate() ValidationErrors {
	var errors ValidationErrors

	// Validate prompt
	if req.Prompt == "" {
		errors = append(errors, ValidationError{
			Field:   "prompt",
			Message: "Prompt is required and cannot be empty",
			Code:    "REQUIRED",
		})
	} else {
		trimmed := strings.TrimSpace(req.Prompt)
		if trimmed == "" {
			errors = append(errors, ValidationError{
				Field:   "prompt",
				Message: "Prompt cannot be only whitespace",
				Code:    "INVALID_FORMAT",
			})
		} else if utf8.RuneCountInString(trimmed) > 2000 {
			errors = append(errors, ValidationError{
				Field:   "prompt",
				Message: "Prompt exceeds maximum length of 2,000 characters",
				Code:    "MAX_LENGTH_EXCEEDED",
			})
		} else if utf8.RuneCountInString(trimmed) < 3 {
			errors = append(errors, ValidationError{
				Field:   "prompt",
				Message: "Prompt must be at least 3 characters long",
				Code:    "MIN_LENGTH_NOT_MET",
			})
		}
	}

	// Validate context if provided
	if req.Context != "" {
		if utf8.RuneCountInString(req.Context) > 5000 {
			errors = append(errors, ValidationError{
				Field:   "context",
				Message: "Context exceeds maximum length of 5,000 characters",
				Code:    "MAX_LENGTH_EXCEEDED",
			})
		}
	}

	// Validate metadata if provided (reuse the same validation as CreateJournalRequest)
	if req.Metadata != nil {
		if len(req.Metadata) > 10 {
			errors = append(errors, ValidationError{
				Field:   "metadata",
				Message: "Metadata cannot have more than 10 fields",
				Code:    "TOO_MANY_FIELDS",
			})
		}

		for key, value := range req.Metadata {
			if key == "" {
				errors = append(errors, ValidationError{
					Field:   "metadata",
					Message: "Metadata keys cannot be empty",
					Code:    "INVALID_KEY",
				})
				continue
			}

			if err := validateMetadataValue(key, value); err != nil {
				errors = append(errors, ValidationError{
					Field:   "metadata",
					Message: err.Error(),
					Code:    "INVALID_VALUE",
				})
			}
		}
	}

	return errors
}
