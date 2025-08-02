package auth

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/store"
	"github.com/garnizeh/englog/internal/testutils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionManagement(t *testing.T) {
	// Setup test database
	db := testutils.DB(t)
	logger := logging.NewTestLogger()

	authService := NewAuthServiceForTest(db, logger, "test-secret-key")
	ctx := context.Background()

	// Create a test user first
	var userID string
	err := db.Write(ctx, func(qtx *store.Queries) error {
		user, err := qtx.CreateUser(ctx, store.CreateUserParams{
			Email:        "test@example.com",
			PasswordHash: "hashedpassword",
			FirstName:    "Test",
			LastName:     "User",
			Timezone:     pgtype.Text{String: "UTC", Valid: true},
			Preferences:  []byte("{}"),
		})
		if err != nil {
			return err
		}
		userID = user.ID.String()
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, userID)

	t.Run("CreateUserSession", func(t *testing.T) {
		// Create access and refresh tokens
		accessToken, err := authService.CreateAccessToken(ctx, userID)
		require.NoError(t, err)

		refreshToken, err := authService.CreateRefreshToken(ctx, userID)
		require.NoError(t, err)

		// Create session
		session, err := authService.CreateUserSession(ctx, userID, accessToken, refreshToken, "127.0.0.1", "Test-Agent/1.0")
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, userID, session.UserID.String())
		assert.NotEmpty(t, session.SessionTokenHash)
		assert.NotEmpty(t, session.RefreshTokenHash)
		assert.True(t, session.IsActive.Bool)
		assert.NotZero(t, session.CreatedAt)
	})

	t.Run("GetUserSessionByToken", func(t *testing.T) {
		// Create unique tokens for this test
		accessToken, err := authService.CreateAccessToken(ctx, userID)
		require.NoError(t, err)

		refreshToken, err := authService.CreateRefreshToken(ctx, userID)
		require.NoError(t, err)

		createdSession, err := authService.CreateUserSession(ctx, userID, accessToken, refreshToken, "127.0.0.1", "Test-Agent/1.0")
		require.NoError(t, err)

		// Retrieve session by the SAME token used to create it
		retrievedSession, err := authService.GetUserSessionByToken(ctx, accessToken)
		require.NoError(t, err)

		assert.Equal(t, createdSession.ID, retrievedSession.ID)
		assert.Equal(t, createdSession.UserID, retrievedSession.UserID)
	})

	t.Run("UpdateSessionActivity", func(t *testing.T) {
		// Create unique tokens for this test
		accessToken, err := authService.CreateAccessToken(ctx, userID)
		require.NoError(t, err)

		refreshToken, err := authService.CreateRefreshToken(ctx, userID)
		require.NoError(t, err)

		session, err := authService.CreateUserSession(ctx, userID, accessToken, refreshToken, "127.0.0.1", "Test-Agent/1.0")
		require.NoError(t, err)

		// Wait a sufficient amount to ensure time difference
		time.Sleep(100 * time.Millisecond)

		// Update session activity
		err = authService.UpdateSessionActivity(ctx, session.ID)
		require.NoError(t, err)

		// Verify the update - retrieve fresh session
		updatedSession, err := authService.GetUserSessionByToken(ctx, accessToken)
		require.NoError(t, err)
		require.NotNil(t, updatedSession)

		// LastActivity should be more recent than CreatedAt
		if updatedSession.LastActivity.Valid {
			assert.True(t, updatedSession.LastActivity.Time.After(updatedSession.CreatedAt.Time),
				"LastActivity (%v) should be after CreatedAt (%v)",
				updatedSession.LastActivity.Time, updatedSession.CreatedAt.Time)
		} else {
			t.Error("LastActivity is not valid after update")
		}
	})

	t.Run("DeactivateSession", func(t *testing.T) {
		// Create unique tokens for this test with specific suffix
		accessToken, err := authService.CreateAccessToken(ctx, userID)
		require.NoError(t, err)

		// Add unique suffix to ensure test isolation
		accessToken = accessToken + "_deactivate_test"

		refreshToken, err := authService.CreateRefreshToken(ctx, userID)
		require.NoError(t, err)

		session, err := authService.CreateUserSession(ctx, userID, accessToken, refreshToken, "127.0.0.1", "Test-Agent/1.0")
		require.NoError(t, err)
		require.NotNil(t, session)

		// Verify session is active initially
		activeSession, err := authService.GetUserSessionByToken(ctx, accessToken)
		require.NoError(t, err)
		require.True(t, activeSession.IsActive.Bool, "Session should be active initially")

		// Deactivate session
		err = authService.DeactivateSession(ctx, session.ID)
		require.NoError(t, err)

		// Try to retrieve session - should fail since it's inactive
		_, err = authService.GetUserSessionByToken(ctx, accessToken)
		require.Error(t, err, "Should get error when trying to retrieve inactive session")
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("GetActiveSessionsByUser", func(t *testing.T) {
		// Create multiple sessions for the user
		sessions := make([]*struct{ accessToken, refreshToken string }, 3)

		for i := 0; i < 3; i++ {
			accessToken, err := authService.CreateAccessToken(ctx, userID)
			require.NoError(t, err)

			refreshToken, err := authService.CreateRefreshToken(ctx, userID)
			require.NoError(t, err)

			_, err = authService.CreateUserSession(ctx, userID, accessToken, refreshToken, "127.0.0.1", "Test-Agent/1.0")
			require.NoError(t, err)

			sessions[i] = &struct{ accessToken, refreshToken string }{accessToken, refreshToken}
		}

		// Get all active sessions
		activeSessions, err := authService.GetActiveSessionsByUser(ctx, userID)
		require.NoError(t, err)

		// Should have at least 3 sessions (might have more from other tests)
		assert.GreaterOrEqual(t, len(activeSessions), 3)

		// All should be active
		for _, session := range activeSessions {
			assert.True(t, session.IsActive.Bool)
			assert.Equal(t, userID, session.UserID.String())
		}
	})

	t.Run("DeactivateUserSessions", func(t *testing.T) {
		// Create a few sessions
		for i := 0; i < 2; i++ {
			accessToken, err := authService.CreateAccessToken(ctx, userID)
			require.NoError(t, err)

			refreshToken, err := authService.CreateRefreshToken(ctx, userID)
			require.NoError(t, err)

			_, err = authService.CreateUserSession(ctx, userID, accessToken, refreshToken, "127.0.0.1", "Test-Agent/1.0")
			require.NoError(t, err)
		}

		// Deactivate all user sessions
		err := authService.DeactivateUserSessions(ctx, userID)
		require.NoError(t, err)

		// Check that no active sessions remain
		activeSessions, err := authService.GetActiveSessionsByUser(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, activeSessions)
	})
}
