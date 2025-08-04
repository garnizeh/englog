package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/models"
)

const baseURL = "http://localhost:8080"

func TestJournalEndpoints(t *testing.T) {
	// Test 1: Create a journal entry
	t.Run("CreateJournal", func(t *testing.T) {
		createReq := models.CreateJournalRequest{
			Content: "This is my first journal entry for testing.",
			Metadata: map[string]interface{}{
				"mood":     8,
				"location": "home",
				"tags":     []string{"test", "first-entry"},
			},
		}

		jsonData, err := json.Marshal(createReq)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		resp, err := http.Post(baseURL+"/journals", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create journal: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
		}

		var journal models.Journal
		if err := json.NewDecoder(resp.Body).Decode(&journal); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if journal.ID == "" {
			t.Error("Expected journal ID to be generated")
		}

		if journal.Content != createReq.Content {
			t.Errorf("Expected content '%s', got '%s'", createReq.Content, journal.Content)
		}

		fmt.Printf("✅ Created journal with ID: %s\n", journal.ID)
	})

	// Test 2: Get all journals
	t.Run("GetAllJournals", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/journals")
		if err != nil {
			t.Fatalf("Failed to get journals: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		journals, ok := response["journals"].([]interface{})
		if !ok {
			t.Error("Expected 'journals' field in response")
		}

		count, ok := response["count"].(float64)
		if !ok {
			t.Error("Expected 'count' field in response")
		}

		if len(journals) != int(count) {
			t.Errorf("Journal count mismatch: expected %d, got %d", int(count), len(journals))
		}

		fmt.Printf("✅ Retrieved %d journals\n", len(journals))
	})

	// Test 3: Test invalid request (empty content)
	t.Run("CreateJournalWithEmptyContent", func(t *testing.T) {
		createReq := models.CreateJournalRequest{
			Content: "",
		}

		jsonData, err := json.Marshal(createReq)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		resp, err := http.Post(baseURL+"/journals", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d for empty content, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		fmt.Printf("✅ Correctly rejected empty content with status %d\n", resp.StatusCode)
	})

	// Test 4: Test invalid JSON
	t.Run("CreateJournalWithInvalidJSON", func(t *testing.T) {
		invalidJSON := `{"content": "test", "invalid": json}`

		resp, err := http.Post(baseURL+"/journals", "application/json", bytes.NewBufferString(invalidJSON))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d for invalid JSON, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		fmt.Printf("✅ Correctly rejected invalid JSON with status %d\n", resp.StatusCode)
	})
}

// Manual test function for development
func manualTest() {
	// Wait a moment for server to start
	time.Sleep(1 * time.Second)

	fmt.Println("🚀 Testing Journal API Endpoints")
	fmt.Println("=================================")

	// Test 1: Create a few journal entries
	journals := []models.CreateJournalRequest{
		{
			Content: "Today was a great day! I learned a lot about Go programming.",
			Metadata: map[string]interface{}{
				"mood":     9,
				"category": "learning",
				"tags":     []string{"go", "programming", "positive"},
			},
		},
		{
			Content: "Feeling a bit overwhelmed with all the tasks I need to complete.",
			Metadata: map[string]interface{}{
				"mood":     5,
				"category": "work",
				"tags":     []string{"stress", "work", "tasks"},
			},
		},
		{
			Content: "Had a wonderful dinner with family. Grateful for these moments.",
			Metadata: map[string]interface{}{
				"mood":     10,
				"category": "family",
				"tags":     []string{"family", "gratitude", "dinner"},
			},
		},
	}

	var createdJournalIDs []string

	for i, journal := range journals {
		jsonData, _ := json.Marshal(journal)
		resp, err := http.Post(baseURL+"/journals", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ Failed to create journal %d: %v\n", i+1, err)
			continue
		}

		var created models.Journal
		json.NewDecoder(resp.Body).Decode(&created)
		resp.Body.Close()

		fmt.Printf("✅ Created journal %d: %s (ID: %s)\n", i+1, created.Content[:30]+"...", created.ID)
		createdJournalIDs = append(createdJournalIDs, created.ID)
	}

	// Test 2: Get all journals
	fmt.Println("\n📚 Retrieving all journals:")
	resp, err := http.Get(baseURL + "/journals")
	if err != nil {
		fmt.Printf("❌ Failed to get journals: %v\n", err)
		return
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	resp.Body.Close()

	count := response["count"].(float64)
	fmt.Printf("✅ Total journals in memory: %.0f\n", count)

	// Test 3: Get specific journals by ID
	fmt.Println("\n🔍 Testing individual journal retrieval:")
	for i, id := range createdJournalIDs {
		resp, err := http.Get(baseURL + "/journals/" + id)
		if err != nil {
			fmt.Printf("❌ Failed to get journal %s: %v\n", id, err)
			continue
		}

		var journal models.Journal
		json.NewDecoder(resp.Body).Decode(&journal)
		resp.Body.Close()

		fmt.Printf("✅ Retrieved journal %d: %s\n", i+1, journal.Content[:50]+"...")
	}

	// Test 4: Test error cases
	fmt.Println("\n⚠️  Testing error cases:")

	// Non-existent journal
	resp, _ = http.Get(baseURL + "/journals/non-existent-id")
	fmt.Printf("✅ Non-existent journal returns status: %d\n", resp.StatusCode)
	resp.Body.Close()

	// Invalid method
	req, _ := http.NewRequest("DELETE", baseURL+"/journals", nil)
	client := &http.Client{}
	resp, _ = client.Do(req)
	fmt.Printf("✅ Invalid method (DELETE) returns status: %d\n", resp.StatusCode)
	resp.Body.Close()

	fmt.Println("\n🎉 Manual testing completed!")
}
