package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/models"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// TODO: change this to an interface to also accommodate other AI providers in the future

// Request represents the request structure for Ollama API
type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// Response represents the response structure from Ollama API
type Response struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// Client implements AI client using Ollama with langchaingo
type Client struct {
	baseURL   string
	modelName string
	llm       llms.Model
	logger    *logging.Logger
}

// New creates a new Ollama client instance using langchaingo
func New(ctx context.Context, modelName, baseURL string) (*Client, error) {
	if modelName == "" {
		return nil, fmt.Errorf("ollama model cannot be empty")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("ollama base URL cannot be empty")
	}

	logger := logging.NewLoggerFromEnv()
	logger.Info("Creating new Ollama service",
		"base_url", baseURL,
		"model", modelName,
	)

	// Create langchaingo Ollama LLM instance
	llm, err := ollama.New(
		ollama.WithServerURL(baseURL),
		ollama.WithModel(modelName),
	)
	if err != nil {
		logger.Error("Failed to create langchaingo Ollama LLM",
			"error", err,
			"base_url", baseURL,
			"model", modelName,
		)
		return nil, fmt.Errorf("failed to create Ollama LLM: %w", err)
	}

	logger.Info("Successfully created langchaingo Ollama LLM",
		"model", modelName,
	)

	return &Client{
		baseURL:   baseURL,
		modelName: modelName,
		llm:       llm,
		logger:    logger,
	}, nil
}

// AnalyzeSentiment performs sentiment analysis on journal content
func (c *Client) AnalyzeSentiment(ctx context.Context, content string) (*models.SentimentResult, error) {
	start := time.Now()

	c.logger.Info("Starting sentiment analysis",
		"content_length", len(content),
		"model", c.modelName,
	)

	prompt := c.buildSentimentPrompt(content)

	response, err := c.callOllamaWithRetry(ctx, prompt, 3)
	if err != nil {
		duration := time.Since(start)
		c.logger.Error("Sentiment analysis failed",
			"error", err,
			"duration", duration,
			"content_length", len(content),
		)
		return nil, fmt.Errorf("sentiment analysis failed: %w", err)
	}

	result, err := c.parseSentimentResponse(response)
	if err != nil {
		c.logger.Error("Failed to parse sentiment response",
			"error", err,
			"response", response,
			"response_length", len(response),
		)
		return nil, fmt.Errorf("failed to parse sentiment response: %w", err)
	}

	result.ProcessedAt = time.Now()
	duration := time.Since(start)

	c.logger.Info("Sentiment analysis completed",
		"duration", duration,
		"score", result.Score,
		"label", result.Label,
		"confidence", result.Confidence,
	)

	return result, nil
}

// GenerateJournal generates a structured journal entry from a prompt
func (c *Client) GenerateJournal(ctx context.Context, req *models.PromptRequest) (*models.GeneratedJournal, error) {
	start := time.Now()

	c.logger.Info("Starting journal generation",
		"prompt", req.Prompt,
		"context", req.Context,
		"model", c.modelName,
	)

	prompt := c.buildGenerationPrompt(req)
	c.logger.Debug("Journal generation prompt",
		"full_prompt", prompt,
	)

	response, err := c.callOllamaWithRetry(ctx, prompt, 3)
	if err != nil {
		duration := time.Since(start)
		c.logger.Error("Journal generation failed",
			"error", err,
			"duration", duration,
			"prompt", req.Prompt,
		)
		return nil, fmt.Errorf("journal generation failed: %w", err)
	}

	result, err := c.parseGenerationResponse(response)
	if err != nil {
		c.logger.Error("Failed to parse generation response",
			"error", err,
			"response", response,
			"response_length", len(response),
		)
		return nil, fmt.Errorf("failed to parse generation response: %w", err)
	}

	result.GeneratedAt = time.Now()
	duration := time.Since(start)

	c.logger.Info("Journal generation completed",
		"duration", duration,
		"content_length", len(result.Content),
		"themes_count", len(result.Metadata.Themes),
		"tags_count", len(result.Metadata.Tags),
	)

	return result, nil
}

// callOllamaWithRetry calls Ollama API with retry mechanism
func (c *Client) callOllamaWithRetry(ctx context.Context, prompt string, maxRetries int) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Check if context was canceled before each attempt
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		response, err := c.callOllama(ctx, prompt)
		if err == nil {
			c.logger.Debug("Ollama call succeeded",
				"attempt", attempt,
				"max_retries", maxRetries,
			)
			return response, nil
		}

		lastErr = err
		c.logger.Warn("Ollama call failed, retrying",
			"attempt", attempt,
			"max_retries", maxRetries,
			"error", err,
		)

		if attempt < maxRetries {
			// Exponential backoff respecting the context
			backoff := time.Duration(attempt) * time.Second
			c.logger.Debug("Backing off before retry",
				"backoff_duration", backoff,
				"attempt", attempt,
			)
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return "", ctx.Err()
			case <-timer.C:
				// Continue to next attempt
			}
		}
	}

	c.logger.Error("Ollama call failed after all retries",
		"max_retries", maxRetries,
		"final_error", lastErr,
	)
	return "", fmt.Errorf("ollama call failed after %d attempts: %w", maxRetries, lastErr)
}

// callOllama makes a single call to Ollama API
func (c *Client) callOllama(ctx context.Context, prompt string) (string, error) {
	// Create a timeout context for this attempt
	timeoutCtx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	c.logger.Debug("Calling Ollama API",
		"model", c.modelName,
		"timeout", "300s",
		"prompt_length", len(prompt),
	)

	// Use langchaingo GenerateFromSinglePrompt method
	response, err := llms.GenerateFromSinglePrompt(timeoutCtx, c.llm, prompt)
	if err != nil {
		c.logger.Error("Failed to call Ollama API",
			"error", err,
			"model", c.modelName,
		)
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}

	c.logger.Debug("Successfully called Ollama API",
		"response_length", len(response),
		"model", c.modelName,
	)

	return response, nil
}

// buildSentimentPrompt creates a prompt for sentiment analysis
func (c *Client) buildSentimentPrompt(content string) string {
	return fmt.Sprintf(`Analyze the sentiment of the following journal entry and respond ONLY with valid JSON in this exact format:
{
  "score": <float between -1.0 and 1.0>,
  "label": "<positive|negative|neutral>",
  "confidence": <float between 0.0 and 1.0>
}

Journal entry to analyze:
%s

Remember: Respond ONLY with the JSON object, no additional text or explanation.`, content)
}

// buildGenerationPrompt creates a prompt for journal generation
func (c *Client) buildGenerationPrompt(req *models.PromptRequest) string {
	context := ""
	if req.Context != "" {
		context = fmt.Sprintf("\nContext: %s", req.Context)
	}

	return fmt.Sprintf(`You are a journal writing assistant. Write a detailed journal entry and provide metadata in JSON format.

User prompt: %s%s

Respond with ONLY valid JSON in this exact structure (no extra text, no markdown, no explanations):

{
  "content": "Write a detailed journal entry here (3-5 sentences about the experience, emotions, and thoughts)",
  "metadata": {
    "mood": "overall mood assessment",
    "emotional_context": "detailed emotional state description",
    "themes": ["theme1", "theme2", "theme3"],
    "entities": ["entity1", "entity2"],
    "key_phrases": ["phrase1", "phrase2", "phrase3"],
    "tags": ["tag1", "tag2", "tag3"]
  },
  "semantic_markers": ["marker1", "marker2", "marker3"],
  "processing_hints": {
    "emotional_intensity": "low",
    "complexity": "moderate",
    "future_analysis_priority": "medium"
  }
}

Important: Return only the JSON object. No other text.`, req.Prompt, context)
}

// parseSentimentResponse parses the sentiment analysis response
func (c *Client) parseSentimentResponse(response string) (*models.SentimentResult, error) {
	var result models.SentimentResult

	// Clean the response (remove any potential markdown formatting)
	cleaned := cleanJSONResponse(response)

	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse sentiment JSON: %w", err)
	}

	// Validate the parsed result
	if result.Score < -1.0 || result.Score > 1.0 {
		return nil, fmt.Errorf("invalid sentiment score: %f (must be between -1.0 and 1.0)", result.Score)
	}

	if result.Confidence < 0.0 || result.Confidence > 1.0 {
		return nil, fmt.Errorf("invalid confidence: %f (must be between 0.0 and 1.0)", result.Confidence)
	}

	validLabels := map[string]bool{"positive": true, "negative": true, "neutral": true}
	if !validLabels[result.Label] {
		return nil, fmt.Errorf("invalid sentiment label: %s (must be positive, negative, or neutral)", result.Label)
	}

	return &result, nil
}

// parseGenerationResponse parses the journal generation response
func (c *Client) parseGenerationResponse(response string) (*models.GeneratedJournal, error) {
	var result models.GeneratedJournal

	// Clean the response (remove any potential markdown formatting)
	cleaned := cleanJSONResponse(response)

	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse generation JSON: %w", err)
	}

	// Validate the parsed result
	if result.Content == "" {
		return nil, fmt.Errorf("generated content cannot be empty")
	}

	if len(result.Metadata.Themes) == 0 {
		return nil, fmt.Errorf("generated metadata must include at least one theme")
	}

	return &result, nil
}

// HealthCheck performs a health check on the AI client using a simple prompt
func (c *Client) HealthCheck(ctx context.Context) error {
	c.logger.Info("Performing AI client health check",
		"model", c.modelName,
		"base_url", c.baseURL,
	)

	// Create a shorter timeout for health checks
	healthCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Simple health check with a basic prompt
	testPrompt := "Respond with 'OK' to confirm you are working."

	start := time.Now()
	response, err := llms.GenerateFromSinglePrompt(healthCtx, c.llm, testPrompt)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("Health check failed",
			"error", err,
			"duration", duration,
			"model", c.modelName,
		)
		return fmt.Errorf("health check failed: %w", err)
	}

	// Check if we got any response
	if len(response) == 0 {
		c.logger.Error("Health check failed: empty response",
			"duration", duration,
			"model", c.modelName,
		)
		return fmt.Errorf("health check failed: empty response from LLM")
	}

	c.logger.Info("AI client health check passed",
		"duration", duration,
		"response_length", len(response),
		"model", c.modelName,
	)

	return nil
}

// cleanJSONResponse removes markdown formatting and extracts JSON
func cleanJSONResponse(response string) string {
	// Remove common markdown formatting
	cleaned := response

	// Remove ```json and ``` markers
	if bytes.Contains([]byte(cleaned), []byte("```json")) {
		start := bytes.Index([]byte(cleaned), []byte("```json")) + 7
		end := bytes.LastIndex([]byte(cleaned), []byte("```"))
		if start < end {
			cleaned = string([]byte(cleaned)[start:end])
		}
	} else if bytes.Contains([]byte(cleaned), []byte("```")) {
		start := bytes.Index([]byte(cleaned), []byte("```")) + 3
		end := bytes.LastIndex([]byte(cleaned), []byte("```"))
		if start < end {
			cleaned = string([]byte(cleaned)[start:end])
		}
	}

	// Find JSON object boundaries
	start := bytes.Index([]byte(cleaned), []byte("{"))
	end := bytes.LastIndex([]byte(cleaned), []byte("}"))

	if start >= 0 && end > start {
		cleaned = string([]byte(cleaned)[start : end+1])
	}

	// Fix common JSON formatting issues from LLM responses
	cleaned = fixMalformedJSON(cleaned)

	return cleaned
}

// fixMalformedJSON fixes common JSON formatting issues from LLM responses
func fixMalformedJSON(jsonStr string) string {
	// First, handle the nested JSON string in content field
	// Pattern: "content": "{\"key\": \"value\"}" -> "content": "escaped content"

	// Handle trailing commas before closing braces/brackets
	re := regexp.MustCompile(`,(\s*[}\]])`)
	jsonStr = re.ReplaceAllString(jsonStr, "$1")

	// Handle missing quotes around string values
	re = regexp.MustCompile(`:\s*([^"{\[\]\s,}]+)(\s*[,}])`)
	jsonStr = re.ReplaceAllStringFunc(jsonStr, func(match string) string {
		// Don't quote numeric values or booleans
		parts := strings.SplitN(match, ":", 2)
		if len(parts) == 2 {
			value := strings.TrimSpace(parts[1])
			ending := ""
			if strings.HasSuffix(value, ",") {
				ending = ","
				value = strings.TrimSuffix(value, ",")
			} else if strings.HasSuffix(value, "}") {
				ending = "}"
				value = strings.TrimSuffix(value, "}")
			}
			value = strings.TrimSpace(value)

			// Check if it's a number, boolean, or null
			if isNumeric(value) || value == "true" || value == "false" || value == "null" {
				return parts[0] + ": " + value + ending
			}
			// Quote string values
			return parts[0] + ": \"" + strings.ReplaceAll(value, "\"", "\\\"") + "\"" + ending
		}
		return match
	})

	// Fix duplicate colons
	re = regexp.MustCompile(`::`)
	jsonStr = re.ReplaceAllString(jsonStr, ":")

	// Handle escaped quotes in metadata field
	if strings.Contains(jsonStr, `"metadata"`) {
		// Find the metadata value and escape it properly
		re = regexp.MustCompile(`"metadata":\s*"([^"]*(?:\\"[^"]*)*)"`)
		jsonStr = re.ReplaceAllStringFunc(jsonStr, func(match string) string {
			// Extract the content between quotes
			parts := strings.SplitN(match, `"metadata":`, 2)
			if len(parts) == 2 {
				content := strings.TrimSpace(parts[1])
				content = strings.Trim(content, `"`)
				// Escape internal quotes
				content = strings.ReplaceAll(content, `"`, `\"`)
				return `"metadata": "` + content + `"`
			}
			return match
		})
	}

	return jsonStr
}

// isNumeric checks if a string represents a numeric value
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
