package handlers

import (
	"github.com/garnizeh/englog/internal/auth"
	"github.com/garnizeh/englog/internal/config"
	"github.com/garnizeh/englog/internal/grpc"
	"github.com/garnizeh/englog/internal/logging"
	"github.com/garnizeh/englog/internal/middleware"
	"github.com/garnizeh/englog/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures and returns the main router with all endpoints
func SetupRoutes(
	cfg *config.Config,
	logger *logging.Logger,
	redisClient *redis.Client,
	authService *auth.AuthService,
	logEntryService *services.LogEntryService,
	projectService *services.ProjectService,
	analyticsService *services.AnalyticsService,
	tagService *services.TagService,
	userService *services.UserService,
	grpcManager *grpc.Manager,
) *gin.Engine {
	r := gin.New() // Use gin.New() instead of gin.Default() for custom middleware

	// Disable automatic trailing slash redirects
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// Add request timeout middleware (first, to wrap all requests)
	r.Use(middleware.RequestTimeout(cfg.Server, logger))

	// Add structured logging middleware with request context
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.ErrorLogger(logger))
	r.Use(middleware.RecoveryLogger(logger))

	// Add security middleware
	r.Use(middleware.CORS(cfg.Security))
	r.Use(middleware.SecurityHeaders(cfg.Security))

	// Add rate limiting middleware
	rateLimitConfig := cfg.RateLimit
	if redisClient == nil {
		// Disable Redis-based rate limiting when Redis is unavailable
		rateLimitConfig.RedisEnabled = false
		logger.Warn("Rate limiting will use fallback mode (no-op) due to Redis unavailability")
	}
	rateLimiter := middleware.NewRateLimiter(redisClient, rateLimitConfig, logger)
	r.Use(rateLimiter.Middleware())

	// Create validation middleware instance
	validator := middleware.NewValidationMiddleware()

	// Health endpoints (no auth required)
	healthHandler := NewHealthHandler()
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)

	// API version 1
	v1 := r.Group("/v1")

	// Authentication endpoints (no auth required)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authService.RegisterHandler)
		auth.POST("/login", authService.LoginHandler)
		auth.POST("/refresh", authService.RefreshHandler)
		auth.POST("/logout", authService.LogoutHandler)
		auth.GET("/me", authService.RequireAuth(), authService.MeHandler)
		auth.GET("/sessions", authService.RequireAuth(), authService.GetActiveSessionsHandler)
		auth.POST("/logout-all", authService.RequireAuth(), authService.LogoutFromAllDevicesHandler)
	}

	// Protected endpoints with session tracking
	protected := v1.Group("/")
	protected.Use(authService.RequireAuthWithSession())

	// Log entries
	logEntryHandler := NewLogEntryHandler(logEntryService)
	logs := protected.Group("/logs")
	{
		logs.POST("", logEntryHandler.CreateLogEntry)
		logs.GET("", logEntryHandler.GetLogEntries)
		logs.GET("/:id", validator.ValidateUUIDParam("id"), logEntryHandler.GetLogEntry)
		logs.PUT("/:id", validator.ValidateUUIDParam("id"), logEntryHandler.UpdateLogEntry)
		logs.DELETE("/:id", validator.ValidateUUIDParam("id"), logEntryHandler.DeleteLogEntry)
		logs.POST("/bulk", logEntryHandler.BulkCreateLogEntries)
	}

	// Projects
	projectHandler := NewProjectHandler(projectService)
	projects := protected.Group("/projects")
	{
		projects.POST("", projectHandler.CreateProject)
		projects.GET("", projectHandler.GetProjects)
		projects.GET("/:id", validator.ValidateUUIDParam("id"), projectHandler.GetProject)
		projects.PUT("/:id", validator.ValidateUUIDParam("id"), projectHandler.UpdateProject)
		projects.DELETE("/:id", validator.ValidateUUIDParam("id"), projectHandler.DeleteProject)
	}

	// Analytics
	analyticsHandler := NewAnalyticsHandler(analyticsService)
	analytics := protected.Group("/analytics")
	{
		analytics.GET("/productivity", analyticsHandler.GetProductivityMetrics)
		analytics.GET("/summary", analyticsHandler.GetActivitySummary)
	}

	// Tags
	tagHandler := NewTagHandler(tagService)
	tags := protected.Group("/tags")
	{
		tags.POST("", tagHandler.CreateTag)
		tags.GET("", tagHandler.GetTags)
		tags.GET("/popular", tagHandler.GetPopularTags)
		tags.GET("/recent", tagHandler.GetRecentlyUsedTags)
		tags.GET("/search", tagHandler.SearchTags)
		tags.GET("/usage", tagHandler.GetUserTagUsage)
		tags.GET("/:id", tagHandler.GetTag)
		tags.PUT("/:id", validator.ValidateUUIDParam("id"), tagHandler.UpdateTag)
		tags.DELETE("/:id", tagHandler.DeleteTag)
	}

	// Users (Profile Management)
	userHandler := NewUserHandler(userService)
	users := protected.Group("/users")
	{
		users.GET("/profile", userHandler.GetProfile)
		users.PUT("/profile", userHandler.UpdateProfile)
		users.POST("/change-password", userHandler.ChangePassword)
		users.DELETE("/account", userHandler.DeleteAccount)
	}

	// Worker and task management routes (protected)
	if grpcManager != nil {
		SetupWorkerRoutes(protected, grpcManager)
	} else {
		logger.Warn("gRPC manager is nil, skipping worker routes setup")
	}

	// Swagger documentation endpoint (no auth required)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
