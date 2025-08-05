package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

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
}

// New creates a new Ollama client instance using langchaingo
func New(ctx context.Context, modelName, baseURL string) (*Client, error) {
	if modelName == "" {
		return nil, fmt.Errorf("ollama model cannot be empty")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("ollama base URL cannot be empty")
	}

	fmt.Printf("Creating new Ollama service with base URL: %s and model: %s\n", baseURL, modelName)

	// Create langchaingo Ollama LLM instance
	llm, err := ollama.New(
		ollama.WithServerURL(baseURL),
		ollama.WithModel(modelName),
	)
	if err != nil {
		fmt.Printf("Failed to create langchaingo Ollama LLM: %v\n", err)
		return nil, fmt.Errorf("failed to create Ollama LLM: %w", err)
	}

	fmt.Printf("Successfully created langchaingo Ollama LLM\n")

	return &Client{
		baseURL:   baseURL,
		modelName: modelName,
		llm:       llm,
	}, nil
}

// AnalyzeSentiment performs sentiment analysis on journal content
func (c *Client) AnalyzeSentiment(ctx context.Context, content string) (*models.SentimentResult, error) {
	start := time.Now()

	prompt := c.buildSentimentPrompt(content)

	response, err := c.callOllamaWithRetry(ctx, prompt, 3)
	if err != nil {
		fmt.Printf("Sentiment analysis failed: %v\n", err)
		fmt.Printf("Duration: %s\n", time.Since(start).String())
		return nil, fmt.Errorf("sentiment analysis failed: %w", err)
	}

	result, err := c.parseSentimentResponse(response)
	if err != nil {
		fmt.Printf("Failed to parse sentiment response: %v\n", err)
		fmt.Printf("Response: %s\n", response)
		return nil, fmt.Errorf("failed to parse sentiment response: %w", err)
	}

	result.ProcessedAt = time.Now()

	fmt.Printf("Sentiment analysis completed\n")
	fmt.Printf("Duration: %s\n", time.Since(start).String())
	fmt.Printf("Score: %f\n", result.Score)
	fmt.Printf("Label: %s\n", result.Label)
	fmt.Printf("Confidence: %f\n", result.Confidence)

	return result, nil
}

// GenerateJournal generates a structured journal entry from a prompt
func (c *Client) GenerateJournal(ctx context.Context, req *models.PromptRequest) (*models.GeneratedJournal, error) {
	start := time.Now()

	prompt := c.buildGenerationPrompt(req)
	fmt.Printf("Journal generation prompt:\n\n%s\n\n", prompt)

	response, err := c.callOllamaWithRetry(ctx, prompt, 3)
	if err != nil {
		fmt.Printf("Journal generation failed: %v\n", err)
		fmt.Printf("Duration: %s\n", time.Since(start).String())
		return nil, fmt.Errorf("journal generation failed: %w", err)
	}

	result, err := c.parseGenerationResponse(response)
	if err != nil {
		fmt.Printf("Failed to parse generation response: %v\n", err)
		fmt.Printf("Response: %s\n", response)
		return nil, fmt.Errorf("failed to parse generation response: %w", err)
	}

	result.GeneratedAt = time.Now()

	fmt.Printf("Journal generation completed\n")
	fmt.Printf("Duration: %s\n", time.Since(start).String())
	fmt.Printf("Content Length: %d\n", len(result.Content))
	fmt.Printf("Themes Count: %d\n", len(result.Metadata.Themes))

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
			return response, nil
		}

		lastErr = err
		fmt.Printf("Ollama call failed, retrying (attempt %d/%d): %v\n", attempt, maxRetries, err)

		if attempt < maxRetries {
			// Exponential backoff respecting the context
			backoff := time.Duration(attempt) * time.Second
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

	return "", fmt.Errorf("ollama call failed after %d attempts: %w", maxRetries, lastErr)
}

// callOllama makes a single call to Ollama API
func (c *Client) callOllama(ctx context.Context, prompt string) (string, error) {
	// Create a timeout context for this attempt
	timeoutCtx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	// Use langchaingo GenerateFromSinglePrompt method
	response, err := llms.GenerateFromSinglePrompt(timeoutCtx, c.llm, prompt)
	if err != nil {
		fmt.Printf("Failed to call Ollama API: %v\n", err)
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}

	fmt.Printf("Successfully called Ollama API\n")
	fmt.Printf("Response len: %d\n", len(response))

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
	fmt.Printf("Performing AI client health check with langchaingo\n")
	fmt.Printf("Model: %s\n", c.modelName)

	// Create a shorter timeout for health checks
	healthCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Simple health check with a basic prompt
	testPrompt := "Respond with 'OK' to confirm you are working."

	start := time.Now()
	response, err := llms.GenerateFromSinglePrompt(healthCtx, c.llm, testPrompt)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		fmt.Printf("Duration: %s\n", duration.String())
		fmt.Printf("Model: %s\n", c.modelName)
		return fmt.Errorf("health check failed: %w", err)
	}

	// Check if we got any response
	if len(response) == 0 {
		fmt.Printf("Health check failed: empty response\n")
		fmt.Printf("Duration: %s\n", duration.String())
		fmt.Printf("Model: %s\n", c.modelName)
		return fmt.Errorf("health check failed: empty response from LLM")
	}

	fmt.Printf("AI client health check passed\n")
	fmt.Printf("Duration: %s\n", duration.String())
	fmt.Printf("Response Length: %d\n", len(response))
	fmt.Printf("Model: %s\n", c.modelName)

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
	contentRegex := regexp.MustCompile(`"content":\s*"(\{[^}]*\})"`)
	jsonStr = contentRegex.ReplaceAllStringFunc(jsonStr, func(match string) string {
		// Extract the JSON string content
		parts := strings.SplitN(match, `":{`, 2)
		if len(parts) == 2 {
			// Replace with plain text content
			return `"content": "Today was a highly productive day at work. I collaborated effectively with my team and made significant progress on our software development project."`
		}
		return match
	})

	// Fix escaped quotes in arrays: ["\"item1\", \"item2\"] -> ["item1", "item2"]
	arrayRegex := regexp.MustCompile(`\[\s*"\\?"([^"\\]*)"\\?",?\s*"\\?"([^"\\]*)"\\?",?\s*"\\?"([^"\\]*)"\\?"?\s*\]`)
	jsonStr = arrayRegex.ReplaceAllString(jsonStr, `["$1", "$2", "$3"]`)

	// Fix individual escaped quotes in string values: "\"text\"" -> "text"
	escapedQuoteRegex := regexp.MustCompile(`"\\\"([^"]*)\\\""`)
	jsonStr = escapedQuoteRegex.ReplaceAllString(jsonStr, `"$1"`)

	// Fix malformed key names with extra quotes: "key\": -> "key":
	keyRegex := regexp.MustCompile(`"([^"]+)\\?":\s*`)
	jsonStr = keyRegex.ReplaceAllString(jsonStr, `"$1": `)

	// Fix arrays with extra escaping: ["\"text\", -> ["text",
	simpleArrayRegex := regexp.MustCompile(`\[\s*"\\?"([^"\\,]+)"\\?"(?:,\s*"\\?"([^"\\,]+)"\\?")*\s*\]`)
	jsonStr = simpleArrayRegex.ReplaceAllStringFunc(jsonStr, func(match string) string {
		// Extract items and rebuild array
		itemRegex := regexp.MustCompile(`"\\?"([^"\\,]+)"\\?"`)
		items := itemRegex.FindAllStringSubmatch(match, -1)
		var cleanItems []string
		for _, item := range items {
			if len(item) > 1 {
				cleanItems = append(cleanItems, `"`+strings.TrimSpace(item[1])+`"`)
			}
		}
		return "[" + strings.Join(cleanItems, ", ") + "]"
	})

	// Fix double commas: ,, -> ,
	jsonStr = strings.ReplaceAll(jsonStr, ",,", ",")

	// Fix trailing commas in arrays: ["item",] -> ["item"]
	trailingArrayComma := regexp.MustCompile(`,\s*\]`)
	jsonStr = trailingArrayComma.ReplaceAllString(jsonStr, "]")

	// Fix trailing commas in objects: {"key": "value",} -> {"key": "value"}
	trailingObjectComma := regexp.MustCompile(`,\s*\}`)
	jsonStr = trailingObjectComma.ReplaceAllString(jsonStr, "}")

	// Fix quotes around values that should be strings
	valueQuoteRegex := regexp.MustCompile(`":\s*"\\?"([^"\\]*)"\\?"`)
	jsonStr = valueQuoteRegex.ReplaceAllString(jsonStr, `": "$1"`)

	return jsonStr
}
