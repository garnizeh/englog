package models

import (
	"time"
)

// ProcessingStatus represents the status of AI processing
type ProcessingStatus string

const (
	ProcessingStatusPending   ProcessingStatus = "pending"
	ProcessingStatusCompleted ProcessingStatus = "completed"
	ProcessingStatusFailed    ProcessingStatus = "failed"
)

// ProcessingResult contains the results of AI processing for a journal entry
type ProcessingResult struct {
	Status          ProcessingStatus `json:"status"`
	SentimentResult *SentimentResult `json:"sentiment_result,omitempty"`
	ProcessedAt     *time.Time       `json:"processed_at,omitempty"`
	ProcessingTime  *time.Duration   `json:"processing_time,omitempty"`
	Error           string           `json:"error,omitempty"`
}

// Journal represents a journal entry in the system
type Journal struct {
	ID               string            `json:"id"`
	Content          string            `json:"content"`
	Timestamp        time.Time         `json:"timestamp"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	Metadata         map[string]any    `json:"metadata,omitempty"`
	ProcessingResult *ProcessingResult `json:"processing_result,omitempty"`
}

// CreateJournalRequest represents the request body for creating a journal
type CreateJournalRequest struct {
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// SentimentResult represents the result of sentiment analysis
type SentimentResult struct {
	Score       float64   `json:"score"`      // Sentiment score from -1.0 (negative) to 1.0 (positive)
	Label       string    `json:"label"`      // "positive", "negative", or "neutral"
	Confidence  float64   `json:"confidence"` // Confidence level from 0.0 to 1.0
	ProcessedAt time.Time `json:"processed_at"`
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
type PromptRequest struct {
	Prompt   string         `json:"prompt"`             // User's input prompt
	Context  string         `json:"context,omitempty"`  // Optional context for better generation
	Metadata map[string]any `json:"metadata,omitempty"` // Additional metadata hints
}
