# Testing Strategy

**Version:** 1.0
**Date:** August 4, 2025
**Author:** Senior Staff Software Architect, Garnizeh
**Status:** In Progress

---

## 📋 Overview

This document outlines the comprehensive testing strategy for the EngLog system, covering unit testing, integration testing, end-to-end testing, performance testing, and security testing. The strategy follows the testing pyramid principle with emphasis on fast, reliable feedback loops.

## 🧪 Testing Architecture

### Testing Pyramid

```mermaid
graph TD
    subgraph "Testing Pyramid"
        A[Manual Testing<br/>Exploratory Testing<br/>User Acceptance Testing]
        B[End-to-End Tests<br/>UI Tests<br/>System Tests]
        C[Integration Tests<br/>API Tests<br/>Database Tests]
        D[Unit Tests<br/>Component Tests<br/>Service Tests]
    end

    style A fill:#ff9999
    style B fill:#ffcc99
    style C fill:#99ccff
    style D fill:#99ff99

    A --> B
    B --> C
    C --> D
```

### Test Coverage Goals

| Layer             | Coverage Target | Focus Area                                |
| ----------------- | --------------- | ----------------------------------------- |
| Unit Tests        | 80%+            | Business logic, utilities, pure functions |
| Integration Tests | 70%+            | API endpoints, database interactions      |
| End-to-End Tests  | Critical paths  | User workflows, system integration        |
| Performance Tests | Key scenarios   | Load, stress, scalability                 |

## 🔬 Unit Testing

### Go Testing Framework

```go
// Test structure following table-driven tests pattern
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserRequest
        setup   func(*testing.T, *UserService)
        want    *User
        wantErr bool
        errType error
    }{
        {
            name: "valid user creation",
            input: CreateUserRequest{
                Email: "test@example.com",
                Name:  "Test User",
            },
            setup: func(t *testing.T, s *UserService) {
                // Setup mocks or test data
            },
            want: &User{
                Email: "test@example.com",
                Name:  "Test User",
                Status: "active",
            },
            wantErr: false,
        },
        {
            name: "duplicate email",
            input: CreateUserRequest{
                Email: "existing@example.com",
                Name:  "Another User",
            },
            setup: func(t *testing.T, s *UserService) {
                // Create existing user
                s.repository.CreateUser(context.Background(), User{
                    Email: "existing@example.com",
                    Name:  "Existing User",
                })
            },
            want:    nil,
            wantErr: true,
            errType: ErrUserAlreadyExists,
        },
        {
            name: "invalid email",
            input: CreateUserRequest{
                Email: "invalid-email",
                Name:  "Test User",
            },
            want:    nil,
            wantErr: true,
            errType: ErrInvalidEmail,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            service := setupUserService(t)
            if tt.setup != nil {
                tt.setup(t, service)
            }

            // Act
            got, err := service.CreateUser(context.Background(), tt.input)

            // Assert
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType != nil {
                    assert.ErrorIs(t, err, tt.errType)
                }
                assert.Nil(t, got)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want.Email, got.Email)
            assert.Equal(t, tt.want.Name, got.Name)
            assert.NotEmpty(t, got.ID)
            assert.NotZero(t, got.CreatedAt)
        })
    }
}

// Test helpers and fixtures
func setupUserService(t *testing.T) *UserService {
    db := setupTestDB(t)
    repo := NewUserRepository(db)
    return NewUserService(repo)
}

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", getTestDSN())
    require.NoError(t, err)

    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)

    // Cleanup after test
    t.Cleanup(func() {
        cleanupTestDB(db)
        db.Close()
    })

    return db
}
```

### Testify and Mockery Integration

```go
//go:generate mockery --name=UserRepository --output=mocks

type UserRepository interface {
    CreateUser(ctx context.Context, user User) (*User, error)
    GetUserByID(ctx context.Context, id string) (*User, error)
    GetUserByEmail(ctx context.Context, email string) (*User, error)
    UpdateUser(ctx context.Context, user User) error
    DeleteUser(ctx context.Context, id string) error
}

// Using generated mocks in tests
func TestUserService_GetUser_WithMock(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewUserRepository(t)
    service := NewUserService(mockRepo)

    userID := "test-user-id"
    expectedUser := &User{
        ID:    userID,
        Email: "test@example.com",
        Name:  "Test User",
    }

    mockRepo.On("GetUserByID", mock.Anything, userID).
        Return(expectedUser, nil).
        Once()

    // Act
    user, err := service.GetUser(context.Background(), userID)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    mockRepo.AssertExpectations(t)
}

// Testing with race conditions
func TestConcurrentUserCreation(t *testing.T) {
    service := setupUserService(t)

    var wg sync.WaitGroup
    results := make(chan error, 10)

    // Try to create the same user concurrently
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, err := service.CreateUser(context.Background(), CreateUserRequest{
                Email: "concurrent@example.com",
                Name:  "Concurrent User",
            })
            results <- err
        }()
    }

    wg.Wait()
    close(results)

    // Count successful and failed creations
    var successes, failures int
    for err := range results {
        if err == nil {
            successes++
        } else {
            failures++
        }
    }

    // Only one should succeed due to unique constraint
    assert.Equal(t, 1, successes)
    assert.Equal(t, 9, failures)
}
```

### Domain Logic Testing

```go
// Testing journal analysis logic
func TestJournalAnalyzer_ExtractInsights(t *testing.T) {
    analyzer := NewJournalAnalyzer()

    tests := []struct {
        name     string
        content  string
        expected InsightResult
    }{
        {
            name:    "positive sentiment detection",
            content: "Today was an amazing day! I felt really happy and accomplished.",
            expected: InsightResult{
                Sentiment: SentimentResult{
                    Score:  0.8,
                    Label:  "positive",
                    Confidence: 0.9,
                },
                Emotions: []Emotion{
                    {Name: "joy", Intensity: 0.85},
                    {Name: "satisfaction", Intensity: 0.75},
                },
                Themes: []string{"achievement", "happiness"},
            },
        },
        {
            name:    "anxiety detection",
            content: "I'm really worried about the presentation tomorrow. What if I mess up?",
            expected: InsightResult{
                Sentiment: SentimentResult{
                    Score:  -0.6,
                    Label:  "negative",
                    Confidence: 0.85,
                },
                Emotions: []Emotion{
                    {Name: "anxiety", Intensity: 0.8},
                    {Name: "fear", Intensity: 0.6},
                },
                Themes: []string{"work", "fear", "performance"},
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := analyzer.ExtractInsights(tt.content)

            assert.NoError(t, err)
            assert.Equal(t, tt.expected.Sentiment.Label, result.Sentiment.Label)
            assert.InDelta(t, tt.expected.Sentiment.Score, result.Sentiment.Score, 0.1)

            // Check emotions are detected
            for _, expectedEmotion := range tt.expected.Emotions {
                found := false
                for _, actualEmotion := range result.Emotions {
                    if actualEmotion.Name == expectedEmotion.Name {
                        found = true
                        assert.InDelta(t, expectedEmotion.Intensity, actualEmotion.Intensity, 0.2)
                        break
                    }
                }
                assert.True(t, found, "Expected emotion %s not found", expectedEmotion.Name)
            }
        })
    }
}
```

## 🔗 Integration Testing

### Database Integration Tests

```go
// Test database operations with real database
func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    db := setupIntegrationDB(t)
    repo := NewUserRepository(db)

    t.Run("create and retrieve user", func(t *testing.T) {
        // Create user
        user := User{
            Email: "integration@example.com",
            Name:  "Integration Test User",
        }

        created, err := repo.CreateUser(context.Background(), user)
        require.NoError(t, err)
        require.NotEmpty(t, created.ID)

        // Retrieve user
        retrieved, err := repo.GetUserByID(context.Background(), created.ID)
        require.NoError(t, err)

        assert.Equal(t, created.Email, retrieved.Email)
        assert.Equal(t, created.Name, retrieved.Name)
        assert.Equal(t, created.ID, retrieved.ID)
    })

    t.Run("test row level security", func(t *testing.T) {
        // Create two users
        user1, _ := repo.CreateUser(context.Background(), User{
            Email: "user1@example.com",
            Name:  "User 1",
        })

        user2, _ := repo.CreateUser(context.Background(), User{
            Email: "user2@example.com",
            Name:  "User 2",
        })

        // Set user context for user1
        err := repo.SetUserContext(user1.ID)
        require.NoError(t, err)

        // User1 should be able to see their own data
        retrieved, err := repo.GetUserByID(context.Background(), user1.ID)
        assert.NoError(t, err)
        assert.Equal(t, user1.ID, retrieved.ID)

        // User1 should NOT be able to see user2's data
        _, err = repo.GetUserByID(context.Background(), user2.ID)
        assert.Error(t, err)
    })
}

// Testing with transactions
func TestJournalService_CreateJournalWithProcessing(t *testing.T) {
    db := setupIntegrationDB(t)
    journalRepo := NewJournalRepository(db)
    processedRepo := NewProcessedContentRepository(db)
    aiService := &mockAIService{}

    service := NewJournalService(journalRepo, processedRepo, aiService)

    t.Run("successful creation with processing", func(t *testing.T) {
        userID := uuid.New().String()

        request := CreateJournalRequest{
            Content: "This is a test journal entry about my day.",
            Tags:    []string{"daily", "reflection"},
        }

        // Mock AI service response
        aiService.On("ProcessContent", mock.Anything).Return(&ProcessingResult{
            Sentiment: "positive",
            Insights:  []string{"Daily reflection", "Personal growth"},
        }, nil)

        entry, err := service.CreateJournal(context.Background(), userID, request)

        assert.NoError(t, err)
        assert.NotEmpty(t, entry.ID)

        // Verify processed content was created
        processed, err := processedRepo.GetByJournalID(context.Background(), entry.ID)
        assert.NoError(t, err)
        assert.Equal(t, "positive", processed.Sentiment)
    })

    t.Run("rollback on processing failure", func(t *testing.T) {
        userID := uuid.New().String()

        request := CreateJournalRequest{
            Content: "This should fail processing.",
        }

        // Mock AI service to return error
        aiService.On("ProcessContent", mock.Anything).Return(nil, errors.New("AI service error"))

        _, err := service.CreateJournal(context.Background(), userID, request)

        assert.Error(t, err)

        // Verify no journal entry was created
        journals, err := journalRepo.GetByUserID(context.Background(), userID, 10, 0)
        assert.NoError(t, err)
        assert.Empty(t, journals)
    })
}
```

### API Integration Tests

```go
// HTTP API integration tests
func TestAPI_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping API integration test in short mode")
    }

    // Setup test server
    router := setupTestRouter(t)
    server := httptest.NewServer(router)
    defer server.Close()

    client := &http.Client{Timeout: 10 * time.Second}

    t.Run("user authentication flow", func(t *testing.T) {
        // Register user
        registerReq := map[string]string{
            "email": "test@example.com",
            "name":  "Test User",
        }

        resp := makeRequest(t, client, "POST", server.URL+"/api/auth/register", registerReq)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)

        var registerResp map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&registerResp)
        require.NoError(t, err)

        // Login user
        loginReq := map[string]string{
            "email": "test@example.com",
        }

        resp = makeRequest(t, client, "POST", server.URL+"/api/auth/login", loginReq)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        var loginResp map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&loginResp)
        require.NoError(t, err)

        token := loginResp["access_token"].(string)
        assert.NotEmpty(t, token)

        // Test authenticated request
        req, _ := http.NewRequest("GET", server.URL+"/api/user/profile", nil)
        req.Header.Set("Authorization", "Bearer "+token)

        resp, err = client.Do(req)
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
    })

    t.Run("journal CRUD operations", func(t *testing.T) {
        token := authenticateTestUser(t, client, server.URL)

        // Create journal
        journalReq := map[string]interface{}{
            "content": "This is my test journal entry.",
            "tags":    []string{"test", "api"},
        }

        resp := makeAuthenticatedRequest(t, client, "POST",
            server.URL+"/api/journals", journalReq, token)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)

        var createResp map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&createResp)
        require.NoError(t, err)

        journalID := createResp["id"].(string)

        // Get journal
        resp = makeAuthenticatedRequest(t, client, "GET",
            server.URL+"/api/journals/"+journalID, nil, token)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        // Update journal
        updateReq := map[string]interface{}{
            "content": "Updated journal content.",
        }

        resp = makeAuthenticatedRequest(t, client, "PUT",
            server.URL+"/api/journals/"+journalID, updateReq, token)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        // Delete journal
        resp = makeAuthenticatedRequest(t, client, "DELETE",
            server.URL+"/api/journals/"+journalID, nil, token)
        assert.Equal(t, http.StatusNoContent, resp.StatusCode)
    })
}

func makeAuthenticatedRequest(t *testing.T, client *http.Client, method, url string,
    body interface{}, token string) *http.Response {

    var reqBody io.Reader
    if body != nil {
        jsonData, _ := json.Marshal(body)
        reqBody = bytes.NewBuffer(jsonData)
    }

    req, err := http.NewRequest(method, url, reqBody)
    require.NoError(t, err)

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := client.Do(req)
    require.NoError(t, err)

    return resp
}
```

## 🌐 End-to-End Testing

### Playwright Test Framework

```javascript
// e2e/tests/user-flow.spec.js
import { test, expect } from "@playwright/test";

test.describe("User Authentication Flow", () => {
  test("user can register, login, and access dashboard", async ({ page }) => {
    // Go to landing page
    await page.goto("/");

    // Click register button
    await page.click('[data-testid="register-button"]');

    // Fill registration form
    await page.fill('[data-testid="email-input"]', "test@example.com");
    await page.fill('[data-testid="name-input"]', "Test User");
    await page.click('[data-testid="submit-registration"]');

    // Should redirect to OTP verification
    await expect(page).toHaveURL("/auth/verify-otp");

    // Mock OTP (in real test, this would be retrieved from email)
    await page.fill('[data-testid="otp-input"]', "123456");
    await page.click('[data-testid="verify-otp"]');

    // Should redirect to dashboard
    await expect(page).toHaveURL("/dashboard");
    await expect(page.locator('[data-testid="welcome-message"]')).toContainText(
      "Welcome, Test User"
    );
  });

  test("invalid login shows error message", async ({ page }) => {
    await page.goto("/auth/login");

    await page.fill('[data-testid="email-input"]', "nonexistent@example.com");
    await page.click('[data-testid="login-button"]');

    await expect(page.locator('[data-testid="error-message"]')).toContainText(
      "User not found"
    );
  });
});

test.describe("Journal Management", () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await authenticateUser(page, "test@example.com");
  });

  test("user can create and view journal entry", async ({ page }) => {
    // Navigate to journals
    await page.click('[data-testid="nav-journals"]');

    // Create new journal
    await page.click('[data-testid="new-journal-button"]');

    // Fill journal form
    await page.fill('[data-testid="journal-title"]', "My Test Journal");
    await page.fill(
      '[data-testid="journal-content"]',
      "This is my test journal entry content."
    );
    await page.fill('[data-testid="journal-tags"]', "test, automation");

    // Submit journal
    await page.click('[data-testid="save-journal"]');

    // Should redirect to journal view
    await expect(page.locator('[data-testid="journal-title"]')).toContainText(
      "My Test Journal"
    );
    await expect(page.locator('[data-testid="journal-content"]')).toContainText(
      "This is my test journal entry content."
    );

    // Check if AI insights are generated
    await expect(page.locator('[data-testid="ai-insights"]')).toBeVisible();
  });

  test("user can search journals", async ({ page }) => {
    await page.goto("/journals");

    // Use search
    await page.fill('[data-testid="journal-search"]', "test");
    await page.press('[data-testid="journal-search"]', "Enter");

    // Should show filtered results
    await expect(page.locator('[data-testid="journal-item"]')).toBeVisible();
    await expect(
      page.locator('[data-testid="search-results-count"]')
    ).toContainText("1 result");
  });
});

// Test helper functions
async function authenticateUser(page, email) {
  await page.goto("/auth/login");
  await page.fill('[data-testid="email-input"]', email);
  await page.click('[data-testid="login-button"]');

  // Mock OTP verification
  await page.fill('[data-testid="otp-input"]', "123456");
  await page.click('[data-testid="verify-otp"]');

  await page.waitForURL("/dashboard");
}
```

### Visual Regression Testing

```javascript
// e2e/tests/visual.spec.js
import { test, expect } from "@playwright/test";

test.describe("Visual Regression Tests", () => {
  test("dashboard layout is consistent", async ({ page }) => {
    await authenticateUser(page, "visual-test@example.com");
    await page.goto("/dashboard");

    // Wait for content to load
    await page.waitForSelector('[data-testid="dashboard-stats"]');

    // Take screenshot and compare
    await expect(page).toHaveScreenshot("dashboard-layout.png");
  });

  test("journal editor appearance", async ({ page }) => {
    await authenticateUser(page, "visual-test@example.com");
    await page.goto("/journals/new");

    // Fill some content to show the editor state
    await page.fill('[data-testid="journal-title"]', "Sample Title");
    await page.fill(
      '[data-testid="journal-content"]',
      "Sample content for visual test"
    );

    await expect(
      page.locator('[data-testid="journal-editor"]')
    ).toHaveScreenshot("journal-editor.png");
  });

  test("responsive design on mobile", async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
    await authenticateUser(page, "mobile-test@example.com");

    await page.goto("/dashboard");
    await expect(page).toHaveScreenshot("mobile-dashboard.png");

    await page.goto("/journals");
    await expect(page).toHaveScreenshot("mobile-journals.png");
  });
});
```

## ⚡ Performance Testing

### Load Testing with Artillery

```yaml
# performance/load-test.yml
config:
  target: "http://localhost:8080"
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 120
      arrivalRate: 50
      name: "Ramp up load"
    - duration: 300
      arrivalRate: 100
      name: "Sustained load"
    - duration: 60
      arrivalRate: 200
      name: "Peak load"
  variables:
    testEmail: "loadtest-{{ $uuid }}@example.com"
    testName: "Load Test User {{ $uuid }}"

scenarios:
  - name: "User Registration and Journal Creation"
    weight: 30
    flow:
      - post:
          url: "/api/auth/register"
          json:
            email: "{{ testEmail }}"
            name: "{{ testName }}"
          capture:
            - json: "$.id"
              as: "userId"
            - json: "$.access_token"
              as: "accessToken"

      - think: 2

      - post:
          url: "/api/journals"
          headers:
            Authorization: "Bearer {{ accessToken }}"
          json:
            content: "This is a load test journal entry with some content to analyze."
            tags: ["loadtest", "performance"]
          capture:
            - json: "$.id"
              as: "journalId"

      - think: 1

      - get:
          url: "/api/journals/{{ journalId }}"
          headers:
            Authorization: "Bearer {{ accessToken }}"

  - name: "Journal Reading and Search"
    weight: 50
    flow:
      - post:
          url: "/api/auth/login"
          json:
            email: "existing-user@example.com"
          capture:
            - json: "$.access_token"
              as: "accessToken"

      - get:
          url: "/api/journals"
          headers:
            Authorization: "Bearer {{ accessToken }}"
          qs:
            limit: 20
            offset: 0

      - think: 2

      - get:
          url: "/api/journals/search"
          headers:
            Authorization: "Bearer {{ accessToken }}"
          qs:
            q: "test"
            limit: 10

  - name: "AI Processing Load"
    weight: 20
    flow:
      - post:
          url: "/api/auth/login"
          json:
            email: "ai-test-user@example.com"
          capture:
            - json: "$.access_token"
              as: "accessToken"

      - post:
          url: "/api/journals"
          headers:
            Authorization: "Bearer {{ accessToken }}"
          json:
            content: "Today I felt really anxious about the upcoming presentation. I kept thinking about all the things that could go wrong. However, I realized that preparation is key, so I spent extra time rehearsing. By the end of the day, I felt more confident and ready to face the challenge."
            process_immediately: true

      - think: 5

      - get:
          url: "/api/insights/daily"
          headers:
            Authorization: "Bearer {{ accessToken }}"
```

### Database Performance Testing

```go
// performance/db_test.go
func BenchmarkUserRepository_GetUserByID(b *testing.B) {
    db := setupBenchmarkDB(b)
    repo := NewUserRepository(db)

    // Pre-populate with test data
    userIDs := make([]string, 1000)
    for i := 0; i < 1000; i++ {
        user, _ := repo.CreateUser(context.Background(), User{
            Email: fmt.Sprintf("bench%d@example.com", i),
            Name:  fmt.Sprintf("Benchmark User %d", i),
        })
        userIDs[i] = user.ID
    }

    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            userID := userIDs[rand.Intn(len(userIDs))]
            _, err := repo.GetUserByID(context.Background(), userID)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}

func BenchmarkJournalSearch_FullText(b *testing.B) {
    db := setupBenchmarkDB(b)
    repo := NewJournalRepository(db)

    // Pre-populate with realistic journal data
    userID := uuid.New().String()
    contents := []string{
        "Today was a wonderful day filled with joy and happiness...",
        "Feeling anxious about the upcoming presentation tomorrow...",
        "Grateful for my family and friends who support me...",
        "Struggling with work-life balance lately...",
        // ... more realistic content
    }

    for i := 0; i < 10000; i++ {
        content := contents[rand.Intn(len(contents))]
        repo.CreateJournal(context.Background(), Journal{
            UserID:  userID,
            Content: content + fmt.Sprintf(" Entry %d", i),
            Tags:    []string{"benchmark", "test"},
        })
    }

    searchTerms := []string{"joy", "anxious", "grateful", "work", "family"}

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        term := searchTerms[rand.Intn(len(searchTerms))]
        _, err := repo.SearchJournals(context.Background(), userID, term, 20, 0)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// Memory usage benchmarks
func BenchmarkAIProcessor_ProcessLargeEntry(b *testing.B) {
    processor := NewAIProcessor()

    // Create large journal entry (5000 words)
    largeContent := strings.Repeat("This is a test sentence about daily life and emotions. ", 500)

    var memBefore, memAfter runtime.MemStats

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        runtime.GC()
        runtime.ReadMemStats(&memBefore)

        _, err := processor.ProcessContent(context.Background(), largeContent)
        if err != nil {
            b.Fatal(err)
        }

        runtime.ReadMemStats(&memAfter)

        // Track memory allocation
        b.ReportMetric(float64(memAfter.Alloc-memBefore.Alloc), "bytes/op")
    }
}
```

## 🔒 Security Testing

### Authentication Security Tests

```go
func TestSecurityVulnerabilities(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()

    t.Run("SQL injection protection", func(t *testing.T) {
        // Test SQL injection in search endpoint
        maliciousQueries := []string{
            "'; DROP TABLE users; --",
            "' OR '1'='1",
            "admin'--",
            "' UNION SELECT * FROM users --",
        }

        for _, query := range maliciousQueries {
            resp, err := http.Get(server.URL + "/api/journals/search?q=" + url.QueryEscape(query))
            require.NoError(t, err)

            // Should not return 500 (internal server error)
            assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)

            // Should return appropriate error or empty results
            assert.True(t, resp.StatusCode == http.StatusBadRequest ||
                       resp.StatusCode == http.StatusOK)
        }
    })

    t.Run("XSS protection", func(t *testing.T) {
        token := authenticateTestUser(t, server)

        xssPayloads := []string{
            "<script>alert('xss')</script>",
            "javascript:alert('xss')",
            "<img src=x onerror=alert('xss')>",
            "&#60;script&#62;alert('xss')&#60;/script&#62;",
        }

        for _, payload := range xssPayloads {
            req := CreateJournalRequest{
                Content: "Normal content " + payload,
                Title:   strPtr("Title " + payload),
            }

            resp := makeAuthenticatedRequest(t, http.DefaultClient, "POST",
                server.URL+"/api/journals", req, token)

            // Should accept the request
            assert.Equal(t, http.StatusCreated, resp.StatusCode)

            var journal map[string]interface{}
            json.NewDecoder(resp.Body).Decode(&journal)

            // But content should be escaped
            content := journal["content"].(string)
            assert.NotContains(t, content, "<script>")
            assert.NotContains(t, content, "javascript:")
        }
    })

    t.Run("authentication bypass attempts", func(t *testing.T) {
        // Test accessing protected endpoints without token
        protectedEndpoints := []string{
            "/api/user/profile",
            "/api/journals",
            "/api/insights",
        }

        for _, endpoint := range protectedEndpoints {
            resp, err := http.Get(server.URL + endpoint)
            require.NoError(t, err)
            assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
        }

        // Test with invalid token
        req, _ := http.NewRequest("GET", server.URL+"/api/user/profile", nil)
        req.Header.Set("Authorization", "Bearer invalid-token")

        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })

    t.Run("rate limiting", func(t *testing.T) {
        // Test rate limiting by making many requests quickly
        var statusCodes []int

        for i := 0; i < 100; i++ {
            resp, err := http.Get(server.URL + "/api/health")
            require.NoError(t, err)
            statusCodes = append(statusCodes, resp.StatusCode)
            resp.Body.Close()
        }

        // Should see some 429 (Too Many Requests) responses
        rateLimitHit := false
        for _, code := range statusCodes {
            if code == http.StatusTooManyRequests {
                rateLimitHit = true
                break
            }
        }

        assert.True(t, rateLimitHit, "Rate limiting should be triggered")
    })
}
```

## 📊 Test Data Management

### Test Fixtures and Factories

```go
// Test data factory pattern
type UserFactory struct {
    db *sql.DB
}

func NewUserFactory(db *sql.DB) *UserFactory {
    return &UserFactory{db: db}
}

func (f *UserFactory) Create(overrides ...func(*User)) *User {
    user := &User{
        ID:    uuid.New().String(),
        Email: fmt.Sprintf("user%d@example.com", time.Now().UnixNano()),
        Name:  "Test User",
        Status: "active",
        CreatedAt: time.Now(),
    }

    // Apply overrides
    for _, override := range overrides {
        override(user)
    }

    // Save to database
    _, err := f.db.Exec(`
        INSERT INTO users (id, email, name, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, user.ID, user.Email, user.Name, user.Status, user.CreatedAt)

    if err != nil {
        panic(err)
    }

    return user
}

func (f *UserFactory) CreateWithJournals(journalCount int) (*User, []*Journal) {
    user := f.Create()

    journalFactory := NewJournalFactory(f.db)
    var journals []*Journal

    for i := 0; i < journalCount; i++ {
        journal := journalFactory.Create(func(j *Journal) {
            j.UserID = user.ID
            j.Content = fmt.Sprintf("Journal entry %d content", i+1)
        })
        journals = append(journals, journal)
    }

    return user, journals
}

// Usage in tests
func TestJournalService_GetUserJournals(t *testing.T) {
    db := setupTestDB(t)
    userFactory := NewUserFactory(db)

    // Create user with 5 journals
    user, journals := userFactory.CreateWithJournals(5)

    service := NewJournalService(NewJournalRepository(db))

    retrieved, err := service.GetUserJournals(context.Background(), user.ID, 10, 0)

    assert.NoError(t, err)
    assert.Len(t, retrieved, 5)
}

// Seed data for different test scenarios
func SeedTestData(db *sql.DB) TestDataSet {
    userFactory := NewUserFactory(db)
    journalFactory := NewJournalFactory(db)

    // Create users with different patterns
    activeUser := userFactory.Create(func(u *User) {
        u.Email = "active@example.com"
        u.Status = "active"
    })

    premiumUser := userFactory.Create(func(u *User) {
        u.Email = "premium@example.com"
        u.Status = "premium"
    })

    // Create journals with different sentiments
    positiveJournal := journalFactory.Create(func(j *Journal) {
        j.UserID = activeUser.ID
        j.Content = "Today was an amazing day! I felt so happy and accomplished."
    })

    negativeJournal := journalFactory.Create(func(j *Journal) {
        j.UserID = activeUser.ID
        j.Content = "I'm feeling really down today. Everything seems difficult."
    })

    return TestDataSet{
        Users: map[string]*User{
            "active":  activeUser,
            "premium": premiumUser,
        },
        Journals: map[string]*Journal{
            "positive": positiveJournal,
            "negative": negativeJournal,
        },
    }
}
```

## 🎯 Test Automation

### Continuous Integration Pipeline

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpass
          POSTGRES_USER: testuser
          POSTGRES_DB: englog_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.24"

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: Install dependencies
        run: go mod download

      - name: Run database migrations
        run: make migrate-test
        env:
          DB_URL: postgres://testuser:testpass@localhost:5432/englog_test?sslmode=disable

      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...
        env:
          DB_URL: postgres://testuser:testpass@localhost:5432/englog_test?sslmode=disable
          REDIS_URL: redis://localhost:6379

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests

    steps:
      - uses: actions/checkout@v4

      - name: Start services with Docker Compose
        run: docker-compose -f docker-compose.test.yml up -d

      - name: Wait for services
        run: |
          docker-compose -f docker-compose.test.yml exec -T api /wait-for-it.sh db:5432 --timeout=60
          docker-compose -f docker-compose.test.yml exec -T api /wait-for-it.sh redis:6379 --timeout=60

      - name: Run integration tests
        run: docker-compose -f docker-compose.test.yml exec -T api go test -tags=integration ./test/integration/...

      - name: Stop services
        run: docker-compose -f docker-compose.test.yml down

  e2e-tests:
    runs-on: ubuntu-latest
    needs: integration-tests

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install Playwright
        run: |
          npm ci
          npx playwright install --with-deps

      - name: Start application
        run: |
          docker-compose up -d
          ./scripts/wait-for-app.sh

      - name: Run E2E tests
        run: npx playwright test

      - name: Upload Playwright report
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: playwright-report/

  performance-tests:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'

    steps:
      - uses: actions/checkout@v4

      - name: Start application
        run: docker-compose -f docker-compose.prod.yml up -d

      - name: Install Artillery
        run: npm install -g artillery

      - name: Run load tests
        run: artillery run performance/load-test.yml

      - name: Run database benchmarks
        run: go test -bench=. -benchmem ./test/performance/...
```

---

## 🔗 Related Documents

- **[API Service Design](../components/API_SERVICE.md)** - API testing considerations
- **[Database Design](../components/DATABASE.md)** - Database testing patterns
- **[Security](./SECURITY.md)** - Security testing requirements
- **[Deployment](./DEPLOYMENT.md)** - Test environment setup
- **[Monitoring](./MONITORING.md)** - Testing metrics and monitoring

---

**Document Status:** 🚧 In Progress
**Next Review:** 2025-09-04
**Last Updated:** 2025-08-04

---

_This document outlines the comprehensive testing strategy for the EngLog system. It serves as the technical specification for implementing thorough testing across all system components and layers._
