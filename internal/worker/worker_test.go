package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/worker"
	"github.com/google/uuid"
)

// mockAIProcessor is a mock implementation of AIProcessor for testing
type mockAIProcessor struct {
	shouldFail      bool
	delay           time.Duration
	sentimentResult *models.SentimentResult
}

func (m *mockAIProcessor) ProcessJournalSentiment(ctx context.Context, journal *models.Journal) (*models.SentimentResult, error) {
	if m.delay > 0 {
		// Respect context cancellation
		select {
		case <-time.After(m.delay):
			// Continue processing
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

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

func TestInMemoryWorker_ProcessJournal_Success(t *testing.T) {
	// Arrange
	mockAI := &mockAIProcessor{
		sentimentResult: &models.SentimentResult{
			Score:       0.8,
			Label:       "positive",
			Confidence:  0.9,
			ProcessedAt: time.Now(),
		},
	}
	worker := worker.NewInMemoryWorker(mockAI, logger())

	journal := &models.Journal{
		ID:      uuid.New().String(),
		Content: "Today was a wonderful day!",
	}

	// Act
	worker.ProcessJournal(context.Background(), journal)

	// Assert
	if journal.ProcessingResult == nil {
		t.Fatal("Expected processing result to be set")
	}

	if journal.ProcessingResult.Status != models.ProcessingStatusCompleted {
		t.Errorf("Expected status to be completed, got %v", journal.ProcessingResult.Status)
	}

	if journal.ProcessingResult.SentimentResult == nil {
		t.Fatal("Expected sentiment result to be set")
	}

	if journal.ProcessingResult.SentimentResult.Score != 0.8 {
		t.Errorf("Expected sentiment score 0.8, got %v", journal.ProcessingResult.SentimentResult.Score)
	}

	if journal.ProcessingResult.SentimentResult.Label != "positive" {
		t.Errorf("Expected sentiment label 'positive', got %v", journal.ProcessingResult.SentimentResult.Label)
	}

	if journal.ProcessingResult.ProcessedAt == nil {
		t.Error("Expected processed_at to be set")
	}

	if journal.ProcessingResult.ProcessingTime == nil {
		t.Error("Expected processing_time to be set")
	}
}

func TestInMemoryWorker_ProcessJournal_Failure(t *testing.T) {
	// Arrange
	mockAI := &mockAIProcessor{
		shouldFail: true,
	}
	worker := worker.NewInMemoryWorker(mockAI, logger())

	journal := &models.Journal{
		ID:      uuid.New().String(),
		Content: "Test content",
	}

	// Act
	worker.ProcessJournal(context.Background(), journal)

	// Assert
	if journal.ProcessingResult == nil {
		t.Fatal("Expected processing result to be set")
	}

	if journal.ProcessingResult.Status != models.ProcessingStatusFailed {
		t.Errorf("Expected status to be failed, got %v", journal.ProcessingResult.Status)
	}

	if journal.ProcessingResult.Error == "" {
		t.Error("Expected error message to be set")
	}

	if journal.ProcessingResult.SentimentResult != nil {
		t.Error("Expected sentiment result to be nil on failure")
	}

	if journal.ProcessingResult.ProcessingTime == nil {
		t.Error("Expected processing_time to be set even on failure")
	}
}

func TestInMemoryWorker_ProcessJournal_Timeout(t *testing.T) {
	// Arrange
	mockAI := &mockAIProcessor{
		delay: 20 * time.Second, // Longer than the 15-second timeout
	}
	worker := worker.NewInMemoryWorker(mockAI, logger())

	journal := &models.Journal{
		ID:      uuid.New().String(),
		Content: "Test content",
	}

	// Act
	start := time.Now()
	worker.ProcessJournal(context.Background(), journal)
	duration := time.Since(start)

	// Assert
	if duration >= 18*time.Second {
		t.Error("Expected processing to timeout around 15 seconds, took too long")
	}

	if journal.ProcessingResult == nil {
		t.Fatal("Expected processing result to be set")
	}

	if journal.ProcessingResult.Status != models.ProcessingStatusFailed {
		t.Errorf("Expected status to be failed due to timeout, got %v", journal.ProcessingResult.Status)
	}

	// Should have a timeout-related error
	if journal.ProcessingResult.Error == "" {
		t.Error("Expected error message to be set on timeout")
	}
}

func TestInMemoryWorker_ProcessJournal_NilJournal(t *testing.T) {
	// Arrange
	mockAI := &mockAIProcessor{}
	worker := worker.NewInMemoryWorker(mockAI, logger())

	// Act & Assert - should not panic
	worker.ProcessJournal(context.Background(), nil)
}

func TestInMemoryWorker_ProcessJournalWithGracefulFailure(t *testing.T) {
	// Arrange
	mockAI := &mockAIProcessor{
		shouldFail: true,
	}
	worker := worker.NewInMemoryWorker(mockAI, logger())

	journal := &models.Journal{
		ID:      uuid.New().String(),
		Content: "Test content",
	}

	// Act
	worker.ProcessJournalWithGracefulFailure(context.Background(), journal)

	// Assert
	if journal.ProcessingResult == nil {
		t.Fatal("Expected processing result to be set")
	}

	if journal.ProcessingResult.Status != models.ProcessingStatusFailed {
		t.Errorf("Expected status to be failed, got %v", journal.ProcessingResult.Status)
	}

	// Journal should still be valid for storage even if processing failed
	if journal.ID == "" {
		t.Error("Journal ID should be preserved")
	}

	if journal.Content == "" {
		t.Error("Journal content should be preserved")
	}
}

func logger() *logging.Logger {
	logConfig := logging.Config{
		Level:  logging.DebugLevel,
		Format: "json",
	}

	return logging.NewLogger(logConfig)
}
