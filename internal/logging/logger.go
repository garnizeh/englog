package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request IDs
	RequestIDKey ContextKey = "request_id"
	// ProcessingIDKey is the context key for processing IDs
	ProcessingIDKey ContextKey = "processing_id"
)

// LogLevel represents the available log levels
type LogLevel string

const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
)

// Config holds logging configuration
type Config struct {
	Level  LogLevel
	Format string // "json" or "text"
}

// Logger wraps slog.Logger with additional context-aware functionality
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new logger with the specified configuration
func NewLogger(config Config) *Logger {
	var level slog.Level
	switch strings.ToUpper(string(config.Level)) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug, // Add source info only for debug level
	}

	var handler slog.Handler
	if config.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewLoggerFromEnv creates a logger using environment variables
func NewLoggerFromEnv() *Logger {
	config := Config{
		Level:  LogLevel(getEnvWithDefault("LOG_LEVEL", "INFO")),
		Format: getEnvWithDefault("LOG_FORMAT", "json"),
	}
	return NewLogger(config)
}

// WithRequestID returns a logger with request ID added to all log entries
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("request_id", requestID),
	}
}

// WithProcessingID returns a logger with processing ID added to all log entries
func (l *Logger) WithProcessingID(processingID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("processing_id", processingID),
	}
}

// WithContext returns a logger that includes context values in log entries
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		logger = logger.With("request_id", requestID)
	}

	if processingID := ctx.Value(ProcessingIDKey); processingID != nil {
		logger = logger.With("processing_id", processingID)
	}

	return &Logger{Logger: logger}
}

// LogHTTPRequest logs an HTTP request with structured information
func (l *Logger) LogHTTPRequest(method, path, remoteAddr, userAgent string, contentLength int64) {
	l.Info("HTTP request received",
		"method", method,
		"path", path,
		"remote_addr", remoteAddr,
		"user_agent", userAgent,
		"content_length", contentLength,
	)
}

// LogHTTPResponse logs an HTTP response with timing information
func (l *Logger) LogHTTPResponse(method, path string, statusCode int, durationMs int64) {
	l.Info("HTTP request processed",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", durationMs,
	)
}

// LogAIProcessingStart logs the start of AI processing
func (l *Logger) LogAIProcessingStart(journalID, contentPreview string, contentLength int) {
	// Limit content preview to avoid logging large amounts of data
	preview := contentPreview
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	l.Info("AI processing started",
		"journal_id", journalID,
		"content_length", contentLength,
		"content_preview", preview,
	)
}

// LogAIProcessingComplete logs the completion of AI processing
func (l *Logger) LogAIProcessingComplete(journalID string, durationMs int64, success bool, errorMsg string) {
	if success {
		l.Info("AI processing completed successfully",
			"journal_id", journalID,
			"duration_ms", durationMs,
		)
	} else {
		l.Error("AI processing failed",
			"journal_id", journalID,
			"duration_ms", durationMs,
			"error", errorMsg,
		)
	}
}

// LogValidationError logs validation errors with context
func (l *Logger) LogValidationError(operation string, errors any) {
	l.Warn("Validation failed",
		"operation", operation,
		"validation_errors", errors,
	)
}

// LogStorageOperation logs storage operations
func (l *Logger) LogStorageOperation(operation, entityType, entityID string, success bool, errorMsg string) {
	if success {
		l.Debug("Storage operation successful",
			"operation", operation,
			"entity_type", entityType,
			"entity_id", entityID,
		)
	} else {
		l.Error("Storage operation failed",
			"operation", operation,
			"entity_type", entityType,
			"entity_id", entityID,
			"error", errorMsg,
		)
	}
}

// LogPerformanceMetric logs performance-related metrics
func (l *Logger) LogPerformanceMetric(operation string, durationMs int64, metadata map[string]any) {
	args := []any{
		"operation", operation,
		"duration_ms", durationMs,
	}

	for k, v := range metadata {
		args = append(args, k, v)
	}

	l.Info("Performance metric", args...)
}

// LogSystemEvent logs significant system events
func (l *Logger) LogSystemEvent(event string, metadata map[string]any) {
	args := []any{"event", event}

	for k, v := range metadata {
		args = append(args, k, v)
	}

	l.Info("System event", args...)
}

// getEnvWithDefault gets an environment variable or returns a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
