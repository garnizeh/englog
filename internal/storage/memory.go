package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/garnizeh/englog/internal/models"
)

// MemoryStore provides in-memory storage for journal entries
type MemoryStore struct {
	journals map[string]*models.Journal
	mu       sync.RWMutex
}

// NewMemoryStore creates a new in-memory storage instance
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		journals: make(map[string]*models.Journal),
	}
}

// Store saves a journal entry to memory
func (ms *MemoryStore) Store(journal *models.Journal) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now()
	if journal.CreatedAt.IsZero() {
		journal.CreatedAt = now
	}
	journal.UpdatedAt = now

	ms.journals[journal.ID] = journal
	return nil
}

// Get retrieves a journal entry by ID
func (ms *MemoryStore) Get(id string) (*models.Journal, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	journal, exists := ms.journals[id]
	if !exists {
		return nil, fmt.Errorf("journal with ID %s not found", id)
	}

	return journal, nil
}

// GetAll returns all journal entries
func (ms *MemoryStore) GetAll() ([]*models.Journal, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	journals := make([]*models.Journal, 0, len(ms.journals))
	for _, journal := range ms.journals {
		journals = append(journals, journal)
	}

	return journals, nil
}

// Update modifies an existing journal entry
func (ms *MemoryStore) Update(id string, journal *models.Journal) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	existing, exists := ms.journals[id]
	if !exists {
		return fmt.Errorf("journal with ID %s not found", id)
	}

	// Preserve original creation time
	journal.CreatedAt = existing.CreatedAt
	journal.UpdatedAt = time.Now()
	journal.ID = id

	ms.journals[id] = journal
	return nil
}

// Delete removes a journal entry
func (ms *MemoryStore) Delete(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.journals[id]; !exists {
		return fmt.Errorf("journal with ID %s not found", id)
	}

	delete(ms.journals, id)
	return nil
}

// Count returns the total number of journal entries
func (ms *MemoryStore) Count() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return len(ms.journals)
}

// StorageStats represents statistics about stored journals
type StorageStats struct {
	TotalJournals       int     `json:"total_journals"`
	ProcessedJournals   int     `json:"processed_journals"`
	AvgProcessingTimeMS float64 `json:"avg_processing_time_ms"`
	OldestJournalAge    string  `json:"oldest_journal_age,omitempty"`
	NewestJournalAge    string  `json:"newest_journal_age,omitempty"`
}

// GetStats returns statistics about stored journals
func (ms *MemoryStore) GetStats() StorageStats {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	stats := StorageStats{
		TotalJournals: len(ms.journals),
	}

	if len(ms.journals) == 0 {
		return stats
	}

	var processedCount int
	var totalProcessingTime float64
	var oldestTime, newestTime time.Time

	for _, journal := range ms.journals {
		// Check if journal has been processed
		if journal.ProcessingResult != nil && journal.ProcessingResult.Status == models.ProcessingStatusCompleted {
			processedCount++
			if journal.ProcessingResult.ProcessingTime != nil {
				totalProcessingTime += journal.ProcessingResult.ProcessingTime.Seconds() * 1000 // Convert to milliseconds
			}
		}

		// Track oldest and newest journal times
		if oldestTime.IsZero() || journal.CreatedAt.Before(oldestTime) {
			oldestTime = journal.CreatedAt
		}
		if newestTime.IsZero() || journal.CreatedAt.After(newestTime) {
			newestTime = journal.CreatedAt
		}
	}

	stats.ProcessedJournals = processedCount
	if processedCount > 0 {
		stats.AvgProcessingTimeMS = totalProcessingTime / float64(processedCount)
	}

	now := time.Now()
	if !oldestTime.IsZero() {
		stats.OldestJournalAge = now.Sub(oldestTime).String()
	}
	if !newestTime.IsZero() {
		stats.NewestJournalAge = now.Sub(newestTime).String()
	}

	return stats
}
