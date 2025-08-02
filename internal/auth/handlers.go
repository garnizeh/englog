package auth

import (
	"fmt"
	"net/http"

	"github.com/garnizeh/englog/internal/models"
	"github.com/garnizeh/englog/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// RegisterHandler registers a new user
// @Summary Register a new user
// @Description Create a new user account and return authentication tokens. The user will be automatically logged in after successful registration.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.UserRegistration true "User registration data" example({"email":"user@example.com","password":"securePassword123","first_name":"John","last_name":"Doe","timezone":"UTC"})
// @Success 201 {object} object{user=models.UserProfile,tokens=models.AuthTokens} "User created successfully"
// @Failure 400 {object} object{error=string,details=string} "Invalid request format"
// @Failure 409 {object} object{error=string} "User already exists"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /v1/auth/register [post]
func (a *AuthService) RegisterHandler(c *gin.Context) {
	var req models.UserRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Warn("Invalid registration request format", "error", err.Error(), "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	a.logger.WithContext(ctx).Info("User registration attempt", "email", req.Email, "ip", c.ClientIP())

	// Check if user already exists
	if err := a.db.Read(c, func(qtx *store.Queries) error {
		_, err := qtx.GetUserByEmail(ctx, req.Email)
		if err == nil {
			a.logger.WithContext(ctx).Warn("Registration failed - user already exists", "email", req.Email, "ip", c.ClientIP())
			c.JSON(http.StatusConflict, gin.H{
				"error": "User already exists",
			})
			return fmt.Errorf("User already exists with email: %s", req.Email)
		}

		return nil
	}); err != nil {
		return
	}

	// Hash password
	hashedPassword, err := a.HashPassword(ctx, req.Password)
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to hash password during registration", "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password",
		})
		return
	}

	// Create user
	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	var user store.User

	if err := a.db.Write(c, func(qtx *store.Queries) error {
		var preferences []byte // Initialize empty preferences
		user, err = qtx.CreateUser(ctx, store.CreateUserParams{
			Email:        req.Email,
			PasswordHash: hashedPassword,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Timezone:     pgtype.Text{String: timezone, Valid: true},
			Preferences:  preferences,
		})
		if err != nil {
			a.logger.LogError(ctx, err, "Failed to create user in database", "email", req.Email)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create user",
			})
			return fmt.Errorf("failed to create user: %w", err)
		}

		return nil
	}); err != nil {
		return
	}

	a.logger.WithContext(ctx).Info("User account created successfully", "email", req.Email, "user_id", user.ID, "ip", c.ClientIP())

	// Create tokens for immediate login
	accessToken, err := a.CreateAccessToken(ctx, user.ID.String())
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to create access token after registration", "email", req.Email, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create access token",
		})
		return
	}

	refreshToken, err := a.CreateRefreshToken(ctx, user.ID.String())
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to create refresh token after registration", "email", req.Email, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create refresh token",
		})
		return
	}

	a.logger.WithContext(ctx).Info("User registration completed successfully", "email", req.Email, "user_id", user.ID, "ip", c.ClientIP())

	// Create session for the user
	_, err = a.CreateUserSession(ctx, user.ID.String(), accessToken, refreshToken, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to create user session after registration", "email", req.Email, "user_id", user.ID)
		// Don't fail registration for session creation error, just log it
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": convertToUserProfile(user),
		"tokens": models.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int(a.accessTokenTTL.Seconds()),
			TokenType:    "Bearer",
		},
	})
}

// LoginHandler authenticates a user
// @Summary Login user
// @Description Authenticate user with email and password, return authentication tokens. Requires valid email and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.UserLogin true "User login credentials" example({"email":"user@example.com","password":"securePassword123"})
// @Success 200 {object} object{user=models.UserProfile,tokens=models.AuthTokens} "Login successful"
// @Failure 400 {object} object{error=string,details=string} "Invalid request format"
// @Failure 401 {object} object{error=string} "Invalid credentials"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /v1/auth/login [post]
func (a *AuthService) LoginHandler(c *gin.Context) {
	var req models.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Warn("Invalid login request format", "error", err.Error(), "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	a.logger.WithContext(ctx).Info("User login attempt", "email", req.Email, "ip", c.ClientIP())

	var user store.User

	// Get user by email
	if err := a.db.Read(c, func(qtx *store.Queries) error {
		var err error
		user, err = qtx.GetUserByEmail(ctx, req.Email)
		if err != nil {
			a.logger.WithContext(ctx).Warn("Login failed - user not found", "email", req.Email, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
			})
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		return nil
	}); err != nil {
		return
	}

	// Check password
	if !a.CheckPassword(req.Password, user.PasswordHash) {
		a.logger.Warn("Login failed - invalid password", "email", req.Email, "user_id", user.ID, "ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	a.logger.WithContext(ctx).Info("Password verification successful", "email", req.Email, "user_id", user.ID, "ip", c.ClientIP())

	// Update last login
	if err := a.db.Write(c, func(qtx *store.Queries) error {
		return qtx.UpdateUserLastLogin(ctx, user.ID)
	}); err != nil {
		a.logger.Warn("Failed to update last login timestamp", "error", err.Error(), "email", req.Email, "user_id", user.ID)
		// Don't fail the login for this non-critical error
	}

	// Create tokens
	accessToken, err := a.CreateAccessToken(ctx, user.ID.String())
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to create access token during login", "email", req.Email, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create access token",
		})
		return
	}

	refreshToken, err := a.CreateRefreshToken(ctx, user.ID.String())
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to create refresh token during login", "email", req.Email, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create refresh token",
		})
		return
	}

	a.logger.WithContext(ctx).Info("User login successful", "email", req.Email, "user_id", user.ID, "ip", c.ClientIP())

	// Create session for the user
	_, err = a.CreateUserSession(ctx, user.ID.String(), accessToken, refreshToken, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to create user session during login", "email", req.Email, "user_id", user.ID)
		// Don't fail login for session creation error, just log it
	}

	c.JSON(http.StatusOK, gin.H{
		"user": convertToUserProfile(user),
		"tokens": models.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int(a.accessTokenTTL.Seconds()),
			TokenType:    "Bearer",
		},
	})
}

// RefreshHandler refreshes authentication tokens
// @Summary Refresh access token
// @Description Generate new access and refresh tokens using a valid refresh token. This allows extending the user session without requiring re-authentication.
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body object{refresh_token=string} true "Refresh token data" example({"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."})
// @Success 200 {object} object{tokens=models.AuthTokens} "Tokens refreshed successfully"
// @Failure 400 {object} object{error=string,details=string} "Invalid request format"
// @Failure 401 {object} object{error=string,details=string} "Invalid refresh token"
// @Router /v1/auth/refresh [post]
func (a *AuthService) RefreshHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Warn("Invalid refresh token request format", "error", err.Error(), "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// Extract user info from token for logging
	userID := "unknown"
	if claims, err := a.ValidateToken(ctx, req.RefreshToken); err == nil {
		userID = claims.Subject
	}

	a.logger.Info("Token refresh attempt", "user_id", userID, "ip", c.ClientIP())

	newAccessToken, newRefreshToken, err := a.RotateRefreshTokenWithSession(ctx, req.RefreshToken, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		a.logger.Warn("Token refresh failed", "error", err.Error(), "user_id", userID, "ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid refresh token",
			"details": err.Error(),
		})
		return
	}

	a.logger.Info("Token refresh successful", "user_id", userID, "ip", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"tokens": models.AuthTokens{
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    int(a.accessTokenTTL.Seconds()),
			TokenType:    "Bearer",
		},
	})
}

// LogoutHandler logs out a user
// @Summary Logout user
// @Description Invalidate refresh token and log out the user. This adds the refresh token to a denylist to prevent further use.
// @Tags auth
// @Accept json
// @Produce json
// @Param logout body object{refresh_token=string} true "Refresh token to invalidate" example({"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."})
// @Success 200 {object} object{message=string} "Successfully logged out"
// @Failure 400 {object} object{error=string} "Invalid request format or token"
// @Failure 500 {object} object{error=string} "Failed to logout"
// @Router /v1/auth/logout [post]
func (a *AuthService) LogoutHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Warn("Invalid logout request format", "error", err.Error(), "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	claims, err := a.ValidateToken(ctx, req.RefreshToken)
	if err != nil {
		a.logger.Warn("Logout failed - invalid refresh token", "error", err.Error(), "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid refresh token",
		})
		return
	}

	userID := claims.Subject
	a.logger.Info("User logout attempt", "user_id", userID, "ip", c.ClientIP())

	if claims.TokenType != models.TokenRefresh {
		a.logger.Warn("Logout failed - invalid token type", "token_type", claims.TokenType, "user_id", userID, "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token type",
		})
		return
	}

	// Denylist the refresh token
	if claims.JTI != "" {
		err = a.DenylistRefreshToken(ctx, claims.JTI, claims.UserID)
		if err != nil {
			a.logger.LogError(ctx, err, "Failed to denylist refresh token during logout", "user_id", userID, "jti", claims.JTI)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to logout",
			})
			return
		}
	}

	// Deactivate sessions for this refresh token
	err = a.deactivateSessionsByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to deactivate sessions during logout", "user_id", userID)
		// Don't fail logout for session deactivation error, just log it
	}

	a.logger.Info("User logout successful", "user_id", userID, "ip", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// MeHandler gets current user profile
// @Summary Get current user profile
// @Description Retrieve the profile information of the currently authenticated user. Requires a valid Bearer token in the Authorization header.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{user=models.UserProfile} "User profile retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid user ID"
// @Failure 401 {object} object{error=string} "Unauthorized - Invalid or missing token"
// @Failure 404 {object} object{error=string} "User not found"
// @Router /v1/auth/me [get]
func (a *AuthService) MeHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		a.logger.Warn("Unauthorized access to user profile - missing user_id in context", "ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		a.logger.Warn("Invalid user ID format in profile request", "user_id", userID, "error", err.Error(), "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	a.logger.Info("User profile request", "user_id", userUUID, "ip", c.ClientIP())

	if err := a.db.Read(c, func(qtx *store.Queries) error {
		ctx := c.Request.Context()
		user, err := qtx.GetUserByID(ctx, userUUID)
		if err != nil {
			a.logger.Warn("User not found in profile request", "user_id", userUUID, "error", err.Error())
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return fmt.Errorf("failed to get user by ID: %w", err)
		}

		a.logger.Info("User profile retrieved successfully", "user_id", userUUID, "email", user.Email)

		c.JSON(http.StatusOK, gin.H{
			"user": convertToUserProfile(user),
		})

		return nil
	}); err != nil {
		return
	}
}

// Helper function to convert store.User to models.UserProfile
func convertToUserProfile(user store.User) models.UserProfile {
	profile := models.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	if user.Timezone.Valid {
		profile.Timezone = user.Timezone.String
	}

	// For preferences, since it's []byte, we'll leave it as nil for now
	// In a real implementation, you'd want to marshal/unmarshal JSON
	profile.Preferences = make(map[string]any)

	if user.LastLoginAt.Valid {
		profile.LastLoginAt = &user.LastLoginAt.Time
	}

	profile.CreatedAt = user.CreatedAt.Time

	return profile
}

// GetActiveSessionsHandler returns active sessions for the current user
// @Summary Get active sessions
// @Description Retrieve all active sessions for the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} object{sessions=[]store.UserSession} "Active sessions retrieved"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Failed to retrieve sessions"
// @Router /v1/auth/sessions [get]
func (a *AuthService) GetActiveSessionsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	sessions, err := a.GetActiveSessionsByUser(c.Request.Context(), userID.(string))
	if err != nil {
		a.logger.LogError(c.Request.Context(), err, "Failed to get active sessions", "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve sessions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// LogoutFromAllDevicesHandler deactivates all sessions for the current user
// @Summary Logout from all devices
// @Description Deactivate all sessions for the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} object{message=string} "Logged out from all devices"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Failed to logout from all devices"
// @Router /v1/auth/logout-all [post]
func (a *AuthService) LogoutFromAllDevicesHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	ctx := c.Request.Context()
	userIDStr := userID.(string)

	a.logger.Info("Logout from all devices request", "user_id", userIDStr, "ip", c.ClientIP())

	err := a.DeactivateUserSessions(ctx, userIDStr)
	if err != nil {
		a.logger.LogError(ctx, err, "Failed to deactivate all user sessions", "user_id", userIDStr)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout from all devices",
		})
		return
	}

	a.logger.Info("User logged out from all devices", "user_id", userIDStr, "ip", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out from all devices",
	})
}
