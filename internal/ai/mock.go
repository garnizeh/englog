package ai

import (
	"context"
	"time"

	"github.com/garnizeh/englog/internal/models"
)

// MockAIProvider is a mock implementation of AIService for testing
type MockAIProvider struct {
	ProcessJournalSentimentFunc   func(ctx context.Context, journal *models.Journal) (*models.SentimentResult, error)
	GenerateStructuredJournalFunc func(ctx context.Context, req *models.PromptRequest) (*models.GeneratedJournal, error)
	ValidateJournalContentFunc    func(content string) error
	ValidatePromptRequestFunc     func(req *models.PromptRequest) error
	HealthCheckFunc               func(ctx context.Context) error
}

// Ensure MockAIProvider implements AIService interface
var _ AIService = (*MockAIProvider)(nil)

// ProcessJournalSentiment mocks sentiment analysis
func (m *MockAIProvider) ProcessJournalSentiment(ctx context.Context, journal *models.Journal) (*models.SentimentResult, error) {
	if m.ProcessJournalSentimentFunc != nil {
		return m.ProcessJournalSentimentFunc(ctx, journal)
	}

	// Default mock response for positive sentiment
	return &models.SentimentResult{
		Score:       0.75,
		Label:       "positive",
		Confidence:  0.92,
		ProcessedAt: time.Now(),
	}, nil
}

// GenerateStructuredJournal mocks journal generation
func (m *MockAIProvider) GenerateStructuredJournal(ctx context.Context, req *models.PromptRequest) (*models.GeneratedJournal, error) {
	if m.GenerateStructuredJournalFunc != nil {
		return m.GenerateStructuredJournalFunc(ctx, req)
	}

	// Default mock response
	return &models.GeneratedJournal{
		Content: "This is a mock generated journal entry about a productive day at work.",
		Metadata: models.GeneratedMetadata{
			Mood:             "positive",
			EmotionalContext: "productive and focused",
			Themes:           []string{"work", "productivity"},
			Entities:         []string{"work", "team", "project"},
			KeyPhrases:       []string{"productive day", "team collaboration"},
			Tags:             []string{"work", "productivity", "ai-generated"},
		},
		SemanticMarkers: []string{"productivity", "collaboration", "achievement"},
		ProcessingHints: map[string]any{
			"source":     "mock",
			"confidence": 0.95,
		},
		GeneratedAt: time.Now(),
	}, nil
}

// ValidateJournalContent mocks content validation
func (m *MockAIProvider) ValidateJournalContent(content string) error {
	if m.ValidateJournalContentFunc != nil {
		return m.ValidateJournalContentFunc(content)
	}

	// Default validation - always pass
	return nil
}

// ValidatePromptRequest mocks prompt request validation
func (m *MockAIProvider) ValidatePromptRequest(req *models.PromptRequest) error {
	if m.ValidatePromptRequestFunc != nil {
		return m.ValidatePromptRequestFunc(req)
	}

	// Default validation - always pass
	return nil
}

// HealthCheck mocks health check
func (m *MockAIProvider) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}

	// Default health check - always healthy
	return nil
}

// NewMockAIProvider creates a new mock AI provider with default implementations
func NewMockAIProvider() *MockAIProvider {
	return &MockAIProvider{}
}

// NewMockAIProviderWithDefaults creates a mock with realistic default responses
func NewMockAIProviderWithDefaults() *MockAIProvider {
	mock := &MockAIProvider{}

	mock.ProcessJournalSentimentFunc = func(ctx context.Context, journal *models.Journal) (*models.SentimentResult, error) {
		// Analyze content for basic sentiment
		score := 0.0
		label := "neutral"

		if len(journal.Content) > 0 {
			content := journal.Content
			positiveWords := []string{"good", "great", "happy", "amazing", "wonderful", "excellent", "productive"}
			negativeWords := []string{"bad", "sad", "awful", "terrible", "horrible", "stressed", "anxious"}

			positiveCount := 0
			negativeCount := 0

			for _, word := range positiveWords {
				if containsWord(content, word) {
					positiveCount++
				}
			}

			for _, word := range negativeWords {
				if containsWord(content, word) {
					negativeCount++
				}
			}

			if positiveCount > negativeCount {
				score = 0.75
				label = "positive"
			} else if negativeCount > positiveCount {
				score = -0.65
				label = "negative"
			}
		}

		return &models.SentimentResult{
			Score:       score,
			Label:       label,
			Confidence:  0.88,
			ProcessedAt: time.Now(),
		}, nil
	}

	return mock
}

// containsWord checks if a string contains a specific word (case-insensitive)
func containsWord(text, word string) bool {
	return len(text) > 0 && len(word) > 0 &&
		(text == word ||
			text[0:len(word)] == word ||
			text[len(text)-len(word):] == word ||
			false) // simplified word matching for mock
}
