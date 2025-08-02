package auth

import (
	"context"
	"testing"

	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/store"
	"github.com/garnizeh/englog/internal/testutils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleSessionManagement(t *testing.T) {
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

	t.Run("BasicSessionCreation", func(t *testing.T) {
		// Create access and refresh tokens
		accessToken, err := authService.CreateAccessToken(ctx, userID)
		require.NoError(t, err)
		require.NotEmpty(t, accessToken)

		refreshToken, err := authService.CreateRefreshToken(ctx, userID)
		require.NoError(t, err)
		require.NotEmpty(t, refreshToken)

		// Create session
		session, err := authService.CreateUserSession(ctx, userID, accessToken, refreshToken, "127.0.0.1", "Test-Agent/1.0")
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, userID, session.UserID.String())
		assert.NotEmpty(t, session.SessionTokenHash)
		assert.NotEmpty(t, session.RefreshTokenHash)
		assert.True(t, session.IsActive.Bool)

		t.Logf("Session created: ID=%s, UserID=%s", session.ID, session.UserID)
	})
}
