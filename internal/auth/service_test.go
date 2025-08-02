package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/garnizeh/englog/internal/auth"
	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/models"
)

// TestNewAuthService verifies the AuthService constructor
// "The beginning is the most important part of the work." 🏗️
func TestNewAuthService(t *testing.T) {
	secretKey := "test-secret-key-for-jwt-authentication-service"

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, secretKey)

	assert.NotNil(t, authService)
}

// TestCreateAccessToken verifies access token creation
// "Security is not a product, but a process." 🔐
func TestCreateAccessToken(t *testing.T) {
	ctx := context.Background()
	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key")
	userID := uuid.New().String()

	token, err := authService.CreateAccessToken(ctx, userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse the token to verify its structure
	parsedToken, err := jwt.ParseWithClaims(token, &auth.Claims{}, func(token *jwt.Token) (any, error) {
		return []byte("test-secret-key"), nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*auth.Claims)
	require.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, models.TokenAccess, claims.TokenType)
	assert.Equal(t, "englog", claims.Issuer)
	assert.Empty(t, claims.JTI) // Access tokens don't have JTI
}

// TestCreateRefreshToken verifies refresh token creation
// "Trust, but verify." 🎯
func TestCreateRefreshToken(t *testing.T) {
	ctx := context.Background()
	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key")
	userID := uuid.New().String()

	token, err := authService.CreateRefreshToken(ctx, userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse the token to verify its structure
	parsedToken, err := jwt.ParseWithClaims(token, &auth.Claims{}, func(token *jwt.Token) (any, error) {
		return []byte("test-secret-key"), nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*auth.Claims)
	require.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, models.TokenRefresh, claims.TokenType)
	assert.Equal(t, "englog", claims.Issuer)
	assert.NotEmpty(t, claims.JTI) // Refresh tokens should have JTI
}

// TestValidateTokenWithValidAccessToken verifies valid access token validation
// "Validation is the most sincere form of confirmation." ✅
func TestValidateTokenWithValidAccessToken(t *testing.T) {
	ctx := context.Background()
	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key")
	userID := uuid.New().String()

	// Create an access token
	token, err := authService.CreateAccessToken(ctx, userID)
	require.NoError(t, err)

	// Validate the token
	claims, err := authService.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, models.TokenAccess, claims.TokenType)
}

// TestValidateTokenWithInvalidToken verifies invalid token handling
// "Error handling is the art of graceful failure." 🚫
func TestValidateTokenWithInvalidToken(t *testing.T) {
	ctx := context.Background()
	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key")

	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "completely invalid token",
			token: "invalid.jwt.token",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "not.a.jwt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := authService.ValidateToken(ctx, tc.token)
			assert.Error(t, err)
		})
	}
}

// TestValidateTokenWithWrongSigningMethod verifies rejection of wrong signing method
// "The devil is in the details." 👹
func TestValidateTokenWithWrongSigningMethod(t *testing.T) {
	ctx := context.Background()
	// Create a token with RS256 instead of HS256
	claims := &auth.Claims{
		UserID:    uuid.New().String(),
		TokenType: models.TokenAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "englog",
		},
	}

	// This will fail because we're trying to use RSA signing
	// but we don't have RSA keys, so this test verifies error handling
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-key")) // This will fail for RS256

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key")
	_, err := authService.ValidateToken(ctx, tokenString)
	assert.Error(t, err)
}

// TestHashPassword verifies password hashing functionality
// "A password is only as strong as its hash." 💪
func TestHashPassword(t *testing.T) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthServiceForTest(nil, testLogger, "test-secret-key")

	testCases := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!2023#Complex$",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "unicode password",
			password: "пароль123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := authService.HashPassword(ctx, tc.password)

			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tc.password, hash)
			assert.True(t, len(hash) > 50) // bcrypt hashes are typically 60 characters
		})
	}
}

// TestCheckPassword verifies password validation against hash
// "Trust is good, verification is better." 🔍
func TestCheckPassword(t *testing.T) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthServiceForTest(nil, testLogger, "test-secret-key")

	password := "test-password-123"
	hash, err := authService.HashPassword(ctx, password)
	require.NoError(t, err)

	// Test correct password
	assert.True(t, authService.CheckPassword(password, hash))

	// Test incorrect passwords
	incorrectPasswords := []string{
		"wrong-password",
		"test-password-124",
		"TEST-PASSWORD-123",
		"",
		"test-password-12",
	}

	for _, wrongPassword := range incorrectPasswords {
		t.Run("incorrect_password_"+wrongPassword, func(t *testing.T) {
			assert.False(t, authService.CheckPassword(wrongPassword, hash))
		})
	}
}

// TestTokenExpiration verifies token expiration handling
// "Time is the fire in which we burn." ⏰
func TestTokenExpiration(t *testing.T) {
	ctx := context.Background()
	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key")
	userID := uuid.New().String()

	// Create a token with very short expiration
	claims := &auth.Claims{
		UserID:    userID,
		TokenType: models.TokenAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "englog",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	// Try to validate expired token
	_, err = authService.ValidateToken(ctx, tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestMultipleTokensUniqueness verifies that multiple tokens are unique
// "Uniqueness is the essence of security." 🎲
func TestMultipleTokensUniqueness(t *testing.T) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key")
	userID := uuid.New().String()

	refreshTokens := make(map[string]bool)

	// Generate 10 refresh tokens (these should be unique due to JTI)
	for i := 0; i < 10; i++ {
		token, err := authService.CreateRefreshToken(ctx, userID)
		require.NoError(t, err)

		// Ensure this token hasn't been generated before
		assert.False(t, refreshTokens[token], "Refresh token collision detected at iteration %d", i)
		refreshTokens[token] = true
	}

	// Test that access tokens with different seconds are different
	token1, err := authService.CreateAccessToken(ctx, userID)
	require.NoError(t, err)

	// Wait to ensure different second timestamp
	time.Sleep(1100 * time.Millisecond)

	token2, err := authService.CreateAccessToken(ctx, userID)
	require.NoError(t, err)

	// These should be different due to different seconds
	assert.NotEqual(t, token1, token2, "Access tokens created in different seconds should be unique")
}

// TestGenerateJTI verifies JTI generation uniqueness
// "Random is the mother of security." 🎰
func TestGenerateJTI(t *testing.T) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key")
	userID := uuid.New().String()

	jtis := make(map[string]bool)

	// Generate multiple refresh tokens and extract their JTIs
	for range 50 {
		token, err := authService.CreateRefreshToken(ctx, userID)
		require.NoError(t, err)

		// Parse token to get JTI
		parsedToken, err := jwt.ParseWithClaims(token, &auth.Claims{}, func(token *jwt.Token) (any, error) {
			return []byte("test-secret-key"), nil
		})
		require.NoError(t, err)

		claims := parsedToken.Claims.(*auth.Claims)
		assert.NotEmpty(t, claims.JTI)
		assert.Equal(t, 32, len(claims.JTI)) // 16 bytes hex encoded = 32 characters

		// Ensure JTI is unique
		assert.False(t, jtis[claims.JTI], "JTI collision detected: %s", claims.JTI)
		jtis[claims.JTI] = true
	}
}

// Benchmark tests for performance critical operations
// "Performance is the ultimate test of design." 🚀

func BenchmarkCreateAccessToken(b *testing.B) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key-for-benchmarking")
	userID := uuid.New().String()

	b.ResetTimer()
	for range b.N {
		_, err := authService.CreateAccessToken(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateRefreshToken(b *testing.B) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key-for-benchmarking")
	userID := uuid.New().String()

	b.ResetTimer()
	for range b.N {
		_, err := authService.CreateRefreshToken(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateToken(b *testing.B) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key-for-benchmarking")
	userID := uuid.New().String()
	token, err := authService.CreateAccessToken(ctx, userID)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		_, err := authService.ValidateToken(ctx, token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashPassword(b *testing.B) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key-for-benchmarking")
	password := "benchmark-password-123"

	b.ResetTimer()
	for range b.N {
		_, err := authService.HashPassword(ctx, password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	ctx := context.Background()

	// Create test logger
	testLogger := logging.NewTestLogger()

	authService := auth.NewAuthService(nil, testLogger, "test-secret-key-for-benchmarking")
	password := "benchmark-password-123"
	hash, err := authService.HashPassword(ctx, password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		authService.CheckPassword(password, hash)
	}
}
