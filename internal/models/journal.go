package models

import (
	"time"
)

// Journal represents a journal entry in the system
type Journal struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	Timestamp time.Time      `json:"timestamp"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// CreateJournalRequest represents the request body for creating a journal
type CreateJournalRequest struct {
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
