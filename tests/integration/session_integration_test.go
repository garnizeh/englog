package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/englog/internal/auth"
	"github.com/garnizeh/englog/internal/handlers"
	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/services"
	"github.com/garnizeh/englog/internal/testutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionManagementIntegration(t *testing.T) {
	// Setup
	db := testutils.DB(t)
	logger := logging.NewTestLogger()

	// Create services
	authService := auth.NewAuthServiceForTest(db, logger, "test-secret-key")
	userService := services.NewUserService(db, logger)

	// Setup router with session middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/v1")

	// Auth routes
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authService.RegisterHandler)
		authGroup.POST("/login", authService.LoginHandler)
		authGroup.GET("/sessions", authService.RequireAuth(), authService.GetActiveSessionsHandler)
		authGroup.POST("/logout-all", authService.RequireAuth(), authService.LogoutFromAllDevicesHandler)
	}

	// Protected routes with session tracking
	protected := v1.Group("/")
	protected.Use(authService.RequireAuthWithSession())

	userHandler := handlers.NewUserHandler(userService)
	protected.GET("/profile", userHandler.GetProfile)

	// Test complete flow
	t.Run("CompleteSessionFlow", func(t *testing.T) {
		// 1. Register user
		regReq := models.UserRegistration{
			Email:     "session-integration@example.com",
			Password:  "password123",
			FirstName: "Session",
			LastName:  "Integration",
			Timezone:  "UTC",
		}
		regBody, _ := json.Marshal(regReq)

		req := httptest.NewRequest("POST", "/v1/auth/register", bytes.NewBuffer(regBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "IntegrationTest/1.0")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var regResp struct {
			User   models.UserProfile `json:"user"`
			Tokens models.AuthTokens  `json:"tokens"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &regResp)
		require.NoError(t, err)

		userID := regResp.User.ID.String()
		accessToken := regResp.Tokens.AccessToken

		t.Logf("User registered: %s", userID)
		t.Logf("Access token: %s", accessToken[:20]+"...")

		// 2. Check that session was created during registration
		req = httptest.NewRequest("GET", "/v1/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var sessionsResp struct {
			Sessions []interface{} `json:"sessions"`
			Count    int           `json:"count"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &sessionsResp)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, sessionsResp.Count, 1, "Should have at least one active session")
		t.Logf("Active sessions count: %d", sessionsResp.Count)

		// 3. Use protected endpoint with session tracking
		req = httptest.NewRequest("GET", "/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("User-Agent", "IntegrationTest/1.0")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Profile request should succeed with session tracking")

		// 4. Logout from all devices
		req = httptest.NewRequest("POST", "/v1/auth/logout-all", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// 5. Check that sessions are deactivated
		req = httptest.NewRequest("GET", "/v1/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &sessionsResp)
		require.NoError(t, err)

		assert.Equal(t, 0, sessionsResp.Count, "Should have no active sessions after logout-all")
		t.Logf("Active sessions after logout-all: %d", sessionsResp.Count)
	})

	t.Run("MultipleSessionsFlow", func(t *testing.T) {
		// Create user
		_, err := userService.CreateUser(context.Background(), &models.UserRegistration{
			Email:     "multi-session@example.com",
			Password:  "password123",
			FirstName: "Multi",
			LastName:  "Session",
			Timezone:  "UTC",
		})
		require.NoError(t, err)

		// Login from different devices
		var tokens []string
		for i := 0; i < 3; i++ {
			loginReq := models.UserLogin{
				Email:    "multi-session@example.com",
				Password: "password123",
			}
			loginBody, _ := json.Marshal(loginReq)

			req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(loginBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "Device"+string(rune('A'+i)))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)

			var loginResp struct {
				Tokens models.AuthTokens `json:"tokens"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &loginResp)
			require.NoError(t, err)

			tokens = append(tokens, loginResp.Tokens.AccessToken)
		}

		// Check we have 3 active sessions
		req := httptest.NewRequest("GET", "/v1/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+tokens[0])
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var sessionsResp struct {
			Sessions []interface{} `json:"sessions"`
			Count    int           `json:"count"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &sessionsResp)
		require.NoError(t, err)

		assert.Equal(t, 3, sessionsResp.Count, "Should have 3 active sessions")
		t.Logf("Active sessions count: %d", sessionsResp.Count)

		// Logout from all devices
		req = httptest.NewRequest("POST", "/v1/auth/logout-all", nil)
		req.Header.Set("Authorization", "Bearer "+tokens[0])
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Check all sessions are deactivated
		req = httptest.NewRequest("GET", "/v1/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+tokens[0])
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &sessionsResp)
		require.NoError(t, err)

		assert.Equal(t, 0, sessionsResp.Count, "Should have no active sessions after logout-all")
		t.Logf("Active sessions after logout-all: %d", sessionsResp.Count)
	})
}
