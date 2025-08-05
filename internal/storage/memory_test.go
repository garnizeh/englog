package storage

import (
	"sync"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/models"
)

func TestNewMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	if store == nil {
		t.Fatal("NewMemoryStore returned nil")
	}

	if store.journals == nil {
		t.Error("journals map not initialized")
	}

	if store.Count() != 0 {
		t.Errorf("Expected empty store, got count: %d", store.Count())
	}
}

func TestMemoryStore_Store(t *testing.T) {
	store := NewMemoryStore()

	tests := []struct {
		name    string
		journal *models.Journal
		wantErr bool
	}{
		{
			name: "valid journal",
			journal: &models.Journal{
				ID:      "test-1",
				Content: "Test content",
			},
			wantErr: false,
		},
		{
			name: "journal with existing created time",
			journal: &models.Journal{
				ID:        "test-2",
				Content:   "Test content 2",
				CreatedAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "journal with empty ID",
			journal: &models.Journal{
				ID:      "",
				Content: "Test content",
			},
			wantErr: false, // MemoryStore doesn't validate IDs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCreatedAt := tt.journal.CreatedAt

			err := store.Store(tt.journal)

			if (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify journal was stored
				if store.Count() == 0 {
					t.Error("Journal was not stored")
				}

				// Check timestamps
				if originalCreatedAt.IsZero() && tt.journal.CreatedAt.IsZero() {
					t.Error("CreatedAt was not set for new journal")
				}

				if !originalCreatedAt.IsZero() && !tt.journal.CreatedAt.Equal(originalCreatedAt) {
					t.Error("CreatedAt was modified for existing journal")
				}

				if tt.journal.UpdatedAt.IsZero() {
					t.Error("UpdatedAt was not set")
				}
			}
		})
	}
}

func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore()

	// Store test journal
	journal := &models.Journal{
		ID:      "test-get",
		Content: "Test content for get",
	}
	store.Store(journal)

	tests := []struct {
		name    string
		id      string
		want    *models.Journal
		wantErr bool
	}{
		{
			name:    "existing journal",
			id:      "test-get",
			want:    journal,
			wantErr: false,
		},
		{
			name:    "non-existing journal",
			id:      "non-existing",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty ID",
			id:      "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Get(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Error("Get() returned nil for existing journal")
					return
				}

				if got.ID != tt.want.ID {
					t.Errorf("Get() ID = %v, want %v", got.ID, tt.want.ID)
				}

				if got.Content != tt.want.Content {
					t.Errorf("Get() Content = %v, want %v", got.Content, tt.want.Content)
				}
			}
		})
	}
}

func TestMemoryStore_GetAll(t *testing.T) {
	store := NewMemoryStore()

	t.Run("empty store", func(t *testing.T) {
		journals, err := store.GetAll()

		if err != nil {
			t.Errorf("GetAll() error = %v", err)
		}

		if len(journals) != 0 {
			t.Errorf("GetAll() returned %d journals, want 0", len(journals))
		}
	})

	t.Run("store with journals", func(t *testing.T) {
		// Add test journals
		testJournals := []*models.Journal{
			{ID: "1", Content: "Content 1"},
			{ID: "2", Content: "Content 2"},
			{ID: "3", Content: "Content 3"},
		}

		for _, journal := range testJournals {
			store.Store(journal)
		}

		journals, err := store.GetAll()

		if err != nil {
			t.Errorf("GetAll() error = %v", err)
		}

		if len(journals) != len(testJournals) {
			t.Errorf("GetAll() returned %d journals, want %d", len(journals), len(testJournals))
		}

		// Verify all journals are present (order may vary)
		foundIDs := make(map[string]bool)
		for _, journal := range journals {
			foundIDs[journal.ID] = true
		}

		for _, expected := range testJournals {
			if !foundIDs[expected.ID] {
				t.Errorf("Journal with ID %s not found in GetAll() results", expected.ID)
			}
		}
	})
}

func TestMemoryStore_Update(t *testing.T) {
	store := NewMemoryStore()

	// Store initial journal
	originalJournal := &models.Journal{
		ID:        "test-update",
		Content:   "Original content",
		CreatedAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	store.Store(originalJournal)

	tests := []struct {
		name       string
		id         string
		newJournal *models.Journal
		wantErr    bool
	}{
		{
			name: "valid update",
			id:   "test-update",
			newJournal: &models.Journal{
				Content: "Updated content",
			},
			wantErr: false,
		},
		{
			name: "update non-existing journal",
			id:   "non-existing",
			newJournal: &models.Journal{
				Content: "New content",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Update(tt.id, tt.newJournal)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify update was applied
				updated, err := store.Get(tt.id)
				if err != nil {
					t.Fatalf("Failed to get updated journal: %v", err)
				}

				if updated.Content != tt.newJournal.Content {
					t.Errorf("Content not updated: got %v, want %v", updated.Content, tt.newJournal.Content)
				}

				// Verify CreatedAt was preserved
				if !updated.CreatedAt.Equal(originalJournal.CreatedAt) {
					t.Error("CreatedAt was not preserved during update")
				}

				// Verify UpdatedAt was set
				if updated.UpdatedAt.IsZero() {
					t.Error("UpdatedAt was not set during update")
				}

				// Verify ID was set correctly
				if updated.ID != tt.id {
					t.Errorf("ID not set correctly: got %v, want %v", updated.ID, tt.id)
				}
			}
		})
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()

	// Store test journal
	journal := &models.Journal{
		ID:      "test-delete",
		Content: "Content to delete",
	}
	store.Store(journal)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "delete existing journal",
			id:      "test-delete",
			wantErr: false,
		},
		{
			name:    "delete non-existing journal",
			id:      "non-existing",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := store.Count()

			err := store.Delete(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify journal was deleted
				if store.Count() != initialCount-1 {
					t.Errorf("Count after delete: got %d, want %d", store.Count(), initialCount-1)
				}

				// Verify journal cannot be retrieved
				_, err := store.Get(tt.id)
				if err == nil {
					t.Error("Deleted journal can still be retrieved")
				}
			} else {
				// Verify count didn't change for failed delete
				if store.Count() != initialCount {
					t.Errorf("Count changed after failed delete: got %d, want %d", store.Count(), initialCount)
				}
			}
		})
	}
}

func TestMemoryStore_Count(t *testing.T) {
	store := NewMemoryStore()

	// Test empty store
	if count := store.Count(); count != 0 {
		t.Errorf("Empty store count: got %d, want 0", count)
	}

	// Add journals and test count
	for i := 0; i < 5; i++ {
		journal := &models.Journal{
			ID:      string(rune('a' + i)),
			Content: "Test content",
		}
		store.Store(journal)

		expectedCount := i + 1
		if count := store.Count(); count != expectedCount {
			t.Errorf("Count after adding %d journals: got %d, want %d", expectedCount, count, expectedCount)
		}
	}

	// Delete journals and test count
	for i := 4; i >= 0; i-- {
		id := string(rune('a' + i))
		store.Delete(id)

		expectedCount := i
		if count := store.Count(); count != expectedCount {
			t.Errorf("Count after deleting journal %s: got %d, want %d", id, count, expectedCount)
		}
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()

	// Test concurrent writes
	t.Run("concurrent writes", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10
		journalsPerGoroutine := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < journalsPerGoroutine; j++ {
					journal := &models.Journal{
						ID:      string(rune('a'+goroutineID)) + string(rune('0'+j)),
						Content: "Concurrent content",
					}
					err := store.Store(journal)
					if err != nil {
						t.Errorf("Concurrent store failed: %v", err)
					}
				}
			}(i)
		}

		wg.Wait()

		expectedCount := numGoroutines * journalsPerGoroutine
		if count := store.Count(); count != expectedCount {
			t.Errorf("Concurrent writes count: got %d, want %d", count, expectedCount)
		}
	})

	// Test concurrent reads
	t.Run("concurrent reads", func(t *testing.T) {
		// Add a test journal
		testJournal := &models.Journal{
			ID:      "concurrent-read-test",
			Content: "Content for concurrent reading",
		}
		store.Store(testJournal)

		var wg sync.WaitGroup
		numReaders := 20

		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				journal, err := store.Get("concurrent-read-test")
				if err != nil {
					t.Errorf("Concurrent read failed: %v", err)
					return
				}
				if journal == nil {
					t.Error("Concurrent read returned nil journal")
					return
				}
				if journal.Content != testJournal.Content {
					t.Errorf("Concurrent read got wrong content: %v", journal.Content)
				}
			}()
		}

		wg.Wait()
	})

	// Test mixed concurrent operations
	t.Run("mixed concurrent operations", func(t *testing.T) {
		var wg sync.WaitGroup

		// Concurrent stores
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				journal := &models.Journal{
					ID:      "mixed-" + string(rune('0'+id)),
					Content: "Mixed operation content",
				}
				store.Store(journal)
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				store.GetAll()
			}()
		}

		// Concurrent count checks
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				store.Count()
			}()
		}

		wg.Wait()
	})
}

func TestMemoryStore_StoreOverwrite(t *testing.T) {
	store := NewMemoryStore()

	// Store initial journal
	journal1 := &models.Journal{
		ID:      "overwrite-test",
		Content: "Original content",
	}
	store.Store(journal1)

	initialCount := store.Count()

	// Store journal with same ID
	journal2 := &models.Journal{
		ID:      "overwrite-test",
		Content: "Overwritten content",
	}
	store.Store(journal2)

	// Count should remain the same (overwrite, not add)
	if count := store.Count(); count != initialCount {
		t.Errorf("Count after overwrite: got %d, want %d", count, initialCount)
	}

	// Content should be updated
	retrieved, err := store.Get("overwrite-test")
	if err != nil {
		t.Fatalf("Failed to get overwritten journal: %v", err)
	}

	if retrieved.Content != journal2.Content {
		t.Errorf("Content after overwrite: got %v, want %v", retrieved.Content, journal2.Content)
	}
}

// Benchmark tests
func BenchmarkMemoryStore_Store(b *testing.B) {
	store := NewMemoryStore()

	b.ResetTimer()

	for b.Loop() {
		journal := &models.Journal{
			ID:      string(rune('a'+b.N%26)) + string(rune('0'+(b.N/26)%10)),
			Content: "Benchmark content",
		}
		store.Store(journal)
	}
}

func BenchmarkMemoryStore_Get(b *testing.B) {
	store := NewMemoryStore()

	// Pre-populate store
	for i := 0; i < 1000; i++ {
		journal := &models.Journal{
			ID:      string(rune('a'+i%26)) + string(rune('0'+(i/26)%10)) + string(rune('0'+(i/260)%10)),
			Content: "Benchmark content",
		}
		store.Store(journal)
	}

	ids := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		ids[i] = string(rune('a'+i%26)) + string(rune('0'+(i/26)%10)) + string(rune('0'+(i/260)%10))
	}

	b.ResetTimer()

	for b.Loop() {
		id := ids[b.N%len(ids)]
		store.Get(id)
	}
}

func BenchmarkMemoryStore_GetAll(b *testing.B) {
	store := NewMemoryStore()

	// Pre-populate store
	for i := 0; i < 100; i++ {
		journal := &models.Journal{
			ID:      string(rune('a'+i%26)) + string(rune('0'+(i/26)%10)),
			Content: "Benchmark content",
		}
		store.Store(journal)
	}

	b.ResetTimer()

	for b.Loop() {
		store.GetAll()
	}
}

func BenchmarkMemoryStore_Count(b *testing.B) {
	store := NewMemoryStore()

	// Pre-populate store
	for i := 0; i < 100; i++ {
		journal := &models.Journal{
			ID:      string(rune('a'+i%26)) + string(rune('0'+(i/26)%10)),
			Content: "Benchmark content",
		}
		store.Store(journal)
	}

	b.ResetTimer()

	for b.Loop() {
		store.Count()
	}
}
