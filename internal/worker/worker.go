package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/garnizeh/englog/internal/models"
)

// AIProcessor interface defines the contract for AI processing services
type AIProcessor interface {
	ProcessJournalSentiment(ctx context.Context, journal *models.Journal) (*models.SentimentResult, error)
}

// InMemoryWorker handles synchronous AI processing of journal entries
type InMemoryWorker struct {
	aiService AIProcessor
	logger    *slog.Logger
}

// NewInMemoryWorker creates a new in-memory worker instance
func NewInMemoryWorker(aiService AIProcessor) *InMemoryWorker {
	return &InMemoryWorker{
		aiService: aiService,
		logger:    slog.Default().With("component", "in_memory_worker"),
	}
}

// ProcessJournal performs synchronous AI processing on a journal entry
func (w *InMemoryWorker) ProcessJournal(ctx context.Context, journal *models.Journal) {
	if journal == nil {
		w.logger.Error("cannot process nil journal")
		return
	}

	w.logger.Info("starting journal processing",
		"journal_id", journal.ID,
		"content_length", len(journal.Content))

	start := time.Now()

	// Initialize processing result with pending status
	journal.ProcessingResult = &models.ProcessingResult{
		Status: models.ProcessingStatusPending,
	}

	// Set timeout for AI processing to prevent hanging requests
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Perform sentiment analysis
	sentimentResult, err := w.aiService.ProcessJournalSentiment(ctx, journal)
	processingTime := time.Since(start)

	if err != nil {
		w.logger.Error("journal processing failed",
			"journal_id", journal.ID,
			"error", err,
			"processing_time", processingTime)

		// Update processing result with error
		journal.ProcessingResult.Status = models.ProcessingStatusFailed
		journal.ProcessingResult.Error = err.Error()
		processingTimePtr := processingTime
		journal.ProcessingResult.ProcessingTime = &processingTimePtr
		return
	}

	// Update processing result with success
	processedAt := time.Now()
	processingTimePtr := processingTime
	journal.ProcessingResult = &models.ProcessingResult{
		Status:          models.ProcessingStatusCompleted,
		SentimentResult: sentimentResult,
		ProcessedAt:     &processedAt,
		ProcessingTime:  &processingTimePtr,
	}

	w.logger.Info("journal processing completed successfully",
		"journal_id", journal.ID,
		"sentiment_score", sentimentResult.Score,
		"sentiment_label", sentimentResult.Label,
		"confidence", sentimentResult.Confidence,
		"processing_time", processingTime)
}

// ProcessJournalWithGracefulFailure processes a journal entry with graceful degradation
// If processing fails, the journal is still considered valid but without AI results
func (w *InMemoryWorker) ProcessJournalWithGracefulFailure(ctx context.Context, journal *models.Journal) {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error("journal processing panicked",
				"journal_id", journal.ID,
				"panic", r)

			// Ensure we have a processing result even in case of panic
			if journal.ProcessingResult == nil {
				journal.ProcessingResult = &models.ProcessingResult{}
			}
			journal.ProcessingResult.Status = models.ProcessingStatusFailed
			journal.ProcessingResult.Error = "processing panicked"
		}
	}()

	w.ProcessJournal(ctx, journal)
}
