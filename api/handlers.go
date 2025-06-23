package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handlers contains all API handlers
type Handlers struct {
	// Add your dependencies here (e.g., Keycloak client, DB, etc.)
}

// NewHandlers creates a new handlers instance
func NewHandlers() *Handlers {
	return &Handlers{}
}

// Login godoc
// @Summary User login
// @Description Authenticate user and receive JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid login request format", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Code:    "REQ_001",
			Message: "Invalid request format",
		})
		return
	}

	// TODO: Implement actual Keycloak authentication
	// For demo purposes, returning mock response
	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		ExpiresIn:    300,
		TokenType:    "Bearer",
		User: UserInfo{
			ID:       "550e8400-e29b-41d4-a716-446655440000",
			Username: req.Username,
			Email:    req.Username + "@example.com",
			Roles:    []string{"user"},
		},
		TrustScore: 88,
	})
}

// GetTrustScore godoc
// @Summary Get current trust score
// @Description Get detailed trust score breakdown for authenticated user
// @Tags trust
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} TrustScoreResponse
// @Failure 401 {object} ErrorResponse
// @Router /trust-score [get]
func (h *Handlers) GetTrustScore(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID := "550e8400-e29b-41d4-a716-446655440000" // TODO: Get from context

	c.JSON(http.StatusOK, TrustScoreResponse{
		UserID:  userID,
		Overall: 88,
		Factors: map[string]int{
			"identity": 30,
			"device":   20,
			"behavior": 18,
			"location": 12,
			"risk":     8,
		},
		Timestamp: time.Now().Format(time.RFC3339),
		NextCheck: time.Now().Add(5 * time.Minute).Format(time.RFC3339),
	})
}

// GetProtectedResource godoc
// @Summary Access protected resource
// @Description Access a resource that requires specific trust level
// @Tags resources
// @Accept json
// @Produce json
// @Security Bearer
// @Param trust query int false "Required trust level" default(50)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /protected [get]
func (h *Handlers) GetProtectedResource(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data":                 "This is protected data",
		"accessed_at":          time.Now().Format(time.RFC3339),
		"trust_level_required": 50,
		"your_trust_level":     88,
	})
}

// Health godoc
// @Summary Health check
// @Description Check service health status
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Service:   "impl-zamaz",
		Version:   "1.0.0",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
