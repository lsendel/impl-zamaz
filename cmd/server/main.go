package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/gin-gonic/gin"
	swaggerFil
	swaggerFiles "github.com/swaggo/files"

	
	ztMiddleware "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/middleware/gin"
	ztConfig "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/pkg/config"
	ztConfig "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/pkg/config"
	ztMiddleware "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/middleware/gin"
	ztDiscovery "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/pkg/discovery"
	ztTypes "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/pkg/types"
)

// Config holds application configuration
type Config struct {
	Port               int    `env:"APP_PORT" envDefault:"8080"`
	Host               string `env:"HOST" envDefault:"0.0.0.0"`
	LogLevel           string `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout    int    `env:"SHUTDOWN_TIMEOUT" envDefault:"30"`
	HealthCheckTimeout int    `env:"HEALTH_CHECK_TIMEOUT" envDefault:"30"`
	SwaggerEnabled     bool   `env:"SWAGGER_ENABLED" envDefault:"true"`
	APIBaseURL         string `env:"API_BASE_URL" envDefault:"http://localhost:8080"`
	APIVersion         string `env:"API_VERSION" envDefault:"v1"`
	CORSOrigins        string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:8080"`
	ServerReadTimeout  int    `env:"SERVER_READ_TIMEOUT" envDefault:"15"`
	ServerWriteTimeout int    `env:"SERVER_WRITE_TIMEOUT" envDefault:"15"`

	
	// Zero Trust configuration will be loaded from ztConfig
}

// @title impl-zamaz Zero Trust API
// @version 1.0.0
// @description Zero Trust authentication implementation using root-zamaz libraries
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

func main() {
	// Load configuration
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal("Failed to parse config:", err)
	}

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))
	slog.SetDefault(logger)

	// Load Zero Trust configuration from root-zamaz
	ztCfg, err := ztConfig.LoadConfig()
	if err != nil {
		logger.Error("Failed to load Zero Trust config", "error", err)
		// Continue with default config for demo
		ztCfg = ztConfig.DefaultConfig()
	}

	// Initialize Keycloak client using root-zamaz library
	keycloakClient, err := ztClient.NewKeycloakClient(ztCfg)
	if err != nil {
		logger.Warn("Failed to initialize Keycloak client", "error", err)
		// Continue without Keycloak for demo
	}

	// Initialize service discovery
	serviceRegistry := ztDiscovery.NewServiceRegistry()

	
	// Start health checks in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go serviceRegistry.StartHealthChecks(ctx, time.Duration(cfg.HealthCheckTimeout)*time.Second)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware with configurable origins
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cfg.CORSOrigins)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize Zero Trust middleware (if Keycloak is available)
	var authMiddleware gin.HandlerFunc
	if keycloakClient != nil {
		middleware := ztMiddleware.NewMiddleware(keycloakClient, &ztTypes.MiddlewareConfig{
			TokenHeader:    "Authorization",
			ContextUserKey: "user",
			SkipPaths:      []string{"/health", "/info", "/", "/swagger", "/api-docs"},
		})
		authMiddleware = middleware.Authenticate()
	} else {
		// Mock middleware for demo
		authMiddleware = func(c *gin.Context) {
			c.Set("user", &ztTypes.AuthenticatedUser{
				ID:       "demo-user",
				Username: "demo",
				Email:    "demo@example.com",
				Roles:    []string{"user"},
			})
			c.Next()
		}
	}

	// Root endpoint with service information

	
	// System endpoints
	r.GET("/health", handleHealth)

	
	// API documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/api-docs", handleAPIDocs)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handleLogin(keycloakClient))
			auth.POST("/logout", authMiddleware, handleLogout(keycloakClient))
		}

		// Service discovery (public for demo)
		discovery := v1.Group("/discovery")
		{
			discoveryHandler := ztDiscovery.NewServiceDiscoveryHandler(serviceRegistry)
			discovery.GET("/services", gin.WrapF(discoveryHandler.HandleListServices))
			discovery.GET("/services/:name", gin.WrapF(discoveryHandler.HandleGetService))
			discovery.POST("/services", gin.WrapF(discoveryHandler.HandleRegisterService))
		}

		// Protected endpoints
		protected := v1.Group("/")
		protected.Use(authMiddleware)
		{
			protected.GET("/trust-score", handleTrustScore)
			protected.GET("/user/profile", handleUserProfile)
			protected.GET("/protected", handleProtectedResource)
		}
	}

	// Serve static files (for React frontend if built)
	r.Static("/static", "./frontend/build/static")
	r.StaticFile("/favicon.ico", "./frontend/build/favicon.ico")

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.ServerIdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	logger.Info("Shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err, "timeout", cfg.ShutdownTimeout)
		return
	}
	logger.Info("Server shutdown completed successfully")

	if keycloakClient != nil {
		if err := keycloakClient.Close(); err != nil {
			logger.Error("Failed to close Keycloak client", "error", err)
		} else {
			logger.Info("Keycloak client closed successfully")
		}
	}

	logger.Info("Server stopped")
}

// Root endpoint
// @Summary Welcome page
// @Description Service information and available endpoints
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func handleRoot(c *gin.Context) {
		"service": "impl-zamaz",
		"status":  "running",
		"message": "Zero Trust authentication implementation using root-zamaz libraries",
		"version": "1.0.0",
		"version":     "1.0.0",
			"health":    "/health - Service health check",
			"info":      "/info - Service information",
			"swagger":   "/swagger/index.html - API documentation",
			"discovery": "/api/v1/discovery/services - Service discovery",
			"auth":      "/api/v1/auth/login - Authentication",
			"auth":       "/api/v1/auth/login - Authentication",
		},
		"features": []string{
			"Zero Trust Authentication",
			"Service Discovery",
			"JWT Token Validation",
			"Trust Level Authorization",
			"Swagger Documentation",
			"Real-time Health Monitoring",
		},
	})
}

// Health check endpoint
// @Summary Health check
// @Description Check service health status
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "impl-zamaz",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// Info endpoint
// @Summary Service information
// @Description Get detailed service information
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /info [get]
func handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "impl-zamaz Zero Trust Demo",
		"description": "Zero Trust authentication implementation using root-zamaz libraries",
		"version":     "1.0.0",
		"libraries": gin.H{
			"zero_trust": "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust",
			"client":     "pkg/client",
			"client":     "pkg/client", 
			"middleware": "middleware/gin",
			"discovery":  "pkg/discovery",
		},
		"features": []string{
			"JWT token validation",
			"Trust level authorization",
			"Device verification ready",
			"Risk assessment ready",
			"Continuous verification ready",
			"Service discovery",
			"Health monitoring",
		},
	})
}

// API documentation endpoint
func handleAPIDocs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"swagger":     "/swagger/index.html",
		"title":       "impl-zamaz Zero Trust API",
		"version":     "1.0.0",
		"description": "Zero Trust authentication using root-zamaz libraries",
	})
}

// Login endpoint
// @Summary User login
// @Description Authenticate user with Keycloak
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body object{username=string,password=string} true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func handleLogin(client ztTypes.KeycloakClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			slog.Warn("Invalid login request format", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request format",
				"code":  "REQ_INVALID_JSON",
			})
			return
		}

		if client != nil {
			// Use real Keycloak authentication
			// TODO: Implement actual login with client.Login()
				"message":  "Keycloak authentication would be performed here",
				"message": "Keycloak authentication would be performed here",
				"note":     "Using mock response for demo",
				"note": "Using mock response for demo",
			})
		} else {
			// Mock response for demo
			c.JSON(http.StatusOK, gin.H{
				"access_token":  "demo-jwt-token-" + req.Username,
				"refresh_token": "demo-refresh-token",
				"expires_in":    300,
				"token_type":    "Bearer",
				"user": gin.H{
					"id":       "demo-" + req.Username,
					"username": req.Username,
					"email":    req.Username + "@example.com",
					"roles":    []string{"user"},
				},
				"trust_score": 88,
			})
		}
	}
}

// Logout endpoint
// @Summary User logout
// @Description Logout user and invalidate token
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func handleLogout(client ztTypes.KeycloakClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement actual logout with token invalidation
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully logged out",
		})
	}
}

// Trust score endpoint
// @Summary Get trust score
// @Description Get current user trust score
// @Tags trust
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /trust-score [get]
func handleTrustScore(c *gin.Context) {
	user, _ := c.Get("user")
	authUser := user.(*ztTypes.AuthenticatedUser)

	// TODO: Calculate real trust score using Zero Trust engine
		"user_id": authUser.ID,
		"overall": 88,
		"overall":  88,
		"factors": gin.H{
			"identity": 30,
			"device":   20,
			"behavior": 18,
			"location": 12,
			"risk":     8,
		"timestamp":  time.Now().UTC(),
		"timestamp": time.Now().UTC(),
		"next_check": time.Now().Add(5 * time.Minute).UTC(),
	})
}

// User profile endpoint
// @Summary Get user profile
// @Description Get authenticated user profile
// @Tags user
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /user/profile [get]
func handleUserProfile(c *gin.Context) {
	user, _ := c.Get("user")
	authUser := user.(*ztTypes.AuthenticatedUser)

		"id":          authUser.ID,
		"username":    authUser.Username,
		"email":       authUser.Email,
		"roles":       authUser.Roles,
		"last_login":  time.Now().Add(-2 * time.Hour).UTC(),
		"last_login": time.Now().Add(-2 * time.Hour).UTC(),
		"trust_level": 88,
	})
}

// Protected resource endpoint
// @Summary Access protected resource
// @Description Access a protected resource requiring authentication
// @Tags resources
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /protected [get]
func handleProtectedResource(c *gin.Context) {
	user, _ := c.Get("user")
	authUser := user.(*ztTypes.AuthenticatedUser)

	c.JSON(http.StatusOK, gin.H{
		"message":     "This is protected data accessible to authenticated users",
		"accessed_by": authUser.Username,
		"endpoint":    "/api/v1/protected",
		"timestamp":   time.Now().UTC(),
		"requirement": "Valid JWT token and sufficient trust level",
	})
}

// parseLogLevel converts string to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
o
}