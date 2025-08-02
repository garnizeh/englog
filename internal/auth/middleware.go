package auth

import (
	"net/http"
	"strings"

	"github.com/garnizeh/englog/internal/models"
	"github.com/gin-gonic/gin"
)

func (a *AuthService) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			a.logger.Warn("Authentication required - missing authorization header", "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			a.logger.Warn("Authentication failed - invalid authorization header format", "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		token := bearerToken[1]
		claims, err := a.ValidateToken(c.Request.Context(), token)
		if err != nil {
			a.logger.Warn("Authentication failed - invalid token", "error", err.Error(), "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		if claims.TokenType != models.TokenAccess {
			a.logger.Warn("Authentication failed - invalid token type", "token_type", claims.TokenType, "user_id", claims.UserID, "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		a.logger.Debug("Authentication successful", "user_id", claims.UserID, "path", c.Request.URL.Path, "ip", c.ClientIP())

		// Set user ID in context for use in handlers
		c.Set("user_id", claims.UserID)
		c.Set("access_token", token)
		c.Next()
	}
}

// RequireAuthWithSession provides authentication with session activity tracking
func (a *AuthService) RequireAuthWithSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			a.logger.Warn("Authentication required - missing authorization header", "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			a.logger.Warn("Authentication failed - invalid authorization header format", "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		token := bearerToken[1]
		claims, err := a.ValidateToken(c.Request.Context(), token)
		if err != nil {
			a.logger.Warn("Authentication failed - invalid token", "error", err.Error(), "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		if claims.TokenType != models.TokenAccess {
			a.logger.Warn("Authentication failed - invalid token type", "token_type", claims.TokenType, "user_id", claims.UserID, "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		// Track session activity
		session, err := a.GetUserSessionByToken(c.Request.Context(), token)
		if err != nil {
			a.logger.Debug("Session not found for token, continuing with basic auth", "user_id", claims.UserID, "error", err.Error())
		} else {
			// Update session activity asynchronously to avoid blocking the request
			go func() {
				if updateErr := a.UpdateSessionActivity(c.Request.Context(), session.ID); updateErr != nil {
					a.logger.Debug("Failed to update session activity", "session_id", session.ID, "error", updateErr.Error())
				}
			}()
			c.Set("session_id", session.ID.String())
		}

		a.logger.Debug("Authentication successful", "user_id", claims.UserID, "path", c.Request.URL.Path, "ip", c.ClientIP())

		// Set user ID and token in context for use in handlers
		c.Set("user_id", claims.UserID)
		c.Set("access_token", token)
		c.Next()
	}
}

func (a *AuthService) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			a.logger.Debug("Optional auth - no authorization header provided", "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.Next()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			a.logger.Debug("Optional auth - invalid authorization header format, continuing without auth", "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.Next()
			return
		}

		token := bearerToken[1]
		claims, err := a.ValidateToken(c.Request.Context(), token)
		if err != nil {
			a.logger.Debug("Optional auth - invalid token, continuing without auth", "error", err.Error(), "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.Next()
			return
		}

		if claims.TokenType == models.TokenAccess {
			a.logger.Debug("Optional auth - valid token found", "user_id", claims.UserID, "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.Set("user_id", claims.UserID)
		} else {
			a.logger.Debug("Optional auth - non-access token type, continuing without auth", "token_type", claims.TokenType, "path", c.Request.URL.Path, "ip", c.ClientIP())
		}

		c.Next()
	}
}
