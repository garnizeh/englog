package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/garnizeh/englog/internal/ai/ollama"
	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/models"
)

// AIService interface defines the methods that any AI service must implement
type AIService interface {
	ProcessJournalSentiment(ctx context.Context, journal *models.Journal) (*models.SentimentResult, error)
	GenerateStructuredJournal(ctx context.Context, req *models.PromptRequest) (*models.GeneratedJournal, error)
	ValidateJournalContent(content string) error
	ValidatePromptRequest(req *models.PromptRequest) error
	HealthCheck(ctx context.Context) error
}

// Service provides AI processing capabilities
type Service struct {
	ollamaClient *ollama.Client
	logger       *logging.Logger
}

// Ensure Service implements AIService interface
var _ AIService = (*Service)(nil)

// NewService creates a new AI service
func NewService(ctx context.Context, modelName, baseURL string, logger *logging.Logger) (*Service, error) {
	ollamaClient, err := ollama.New(ctx, modelName, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	return &Service{
		ollamaClient: ollamaClient,
		logger:       logger,
	}, nil
}

// ProcessJournalSentiment analyzes the sentiment of a journal entry
func (s *Service) ProcessJournalSentiment(ctx context.Context, journal *models.Journal) (*models.SentimentResult, error) {
	if journal == nil {
		return nil, fmt.Errorf("journal cannot be nil")
	}

	if strings.TrimSpace(journal.Content) == "" {
		return nil, fmt.Errorf("journal content cannot be empty")
	}

	s.logger.Info("processing journal sentiment",
		"journal_id", journal.ID,
		"content_length", len(journal.Content),
	)

	start := time.Now()
	result, err := s.ollamaClient.AnalyzeSentiment(ctx, journal.Content)
	if err != nil {
		s.logger.Error("sentiment analysis failed",
			"journal_id", journal.ID,
			"error", err,
			"duration", time.Since(start),
		)
		return nil, fmt.Errorf("sentiment analysis failed for journal %s: %w", journal.ID, err)
	}

	s.logger.Info("sentiment analysis completed",
		"journal_id", journal.ID,
		"sentiment_score", result.Score,
		"sentiment_label", result.Label,
		"confidence", result.Confidence,
		"duration", time.Since(start),
	)

	return result, nil
}

// GenerateStructuredJournal creates a structured journal entry from a prompt
func (s *Service) GenerateStructuredJournal(ctx context.Context, req *models.PromptRequest) (*models.GeneratedJournal, error) {
	if req == nil {
		return nil, fmt.Errorf("prompt request cannot be nil")
	}

	if strings.TrimSpace(req.Prompt) == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	s.logger.Info("generating structured journal",
		"prompt_length", len(req.Prompt),
		"has_context", req.Context != "",
	)

	start := time.Now()
	result, err := s.ollamaClient.GenerateJournal(ctx, req)
	if err != nil {
		s.logger.Error("journal generation failed",
			"error", err,
			"duration", time.Since(start),
		)
		return nil, fmt.Errorf("journal generation failed: %w", err)
	}

	s.logger.Info("journal generation completed",
		"content_length", len(result.Content),
		"themes_count", len(result.Metadata.Themes),
		"entities_count", len(result.Metadata.Entities),
		"duration", time.Since(start),
	)

	return result, nil
}

// ValidateJournalContent performs basic validation on journal content
func (s *Service) ValidateJournalContent(content string) error {
	content = strings.TrimSpace(content)

	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	if len(content) < 10 {
		return fmt.Errorf("content too short (minimum 10 characters)")
	}

	if len(content) > 50000 {
		return fmt.Errorf("content too long (maximum 50,000 characters)")
	}

	return nil
}

// ValidatePromptRequest performs basic validation on prompt requests
func (s *Service) ValidatePromptRequest(req *models.PromptRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if len(prompt) < 5 {
		return fmt.Errorf("prompt too short (minimum 5 characters)")
	}

	if len(prompt) > 5000 {
		return fmt.Errorf("prompt too long (maximum 5,000 characters)")
	}

	return nil
}

// HealthCheck verifies that the AI service is operational
func (s *Service) HealthCheck(ctx context.Context) error {
	// Simple health check: try to analyze sentiment of a test message
	testJournal := &models.Journal{
		ID:      "health_check",
		Content: "This is a simple test message for health checking.",
	}

	_, err := s.ProcessJournalSentiment(ctx, testJournal)
	if err != nil {
		return fmt.Errorf("AI service health check failed: %w", err)
	}

	return nil
}
