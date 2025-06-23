package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}{
	Version:     "1.0.0",
	Host:        "localhost:8080",
	BasePath:    "/api/v1",
	Schemes:     []string{"http", "https"},
	Title:       "impl-zamaz Zero Trust API",
	Description: "Zero Trust Authentication Service with Swagger documentation",
}

// @title impl-zamaz Zero Trust API
// @version 1.0.0
// @description Zero Trust Authentication Service with device attestation, risk assessment, and continuous verification
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/lsendel/root-zamaz
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// SetupSwagger configures Swagger UI with authentication
func SetupSwagger(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	// Public swagger docs (no auth required for viewing API docs)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API documentation endpoint
	r.GET("/api-docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"swagger": "/swagger/index.html",
			"info":    SwaggerInfo,
			"authentication": gin.H{
				"type":   "Bearer JWT",
				"header": "Authorization",
				"format": "Bearer <token>",
				"obtain_token": gin.H{
					"endpoint": "/api/v1/auth/login",
					"method":   "POST",
					"body": gin.H{
						"username": "string",
						"password": "string",
					},
				},
			},
			"trust_levels": gin.H{
				"read":   25,
				"write":  50,
				"admin":  75,
				"delete": 90,
			},
		})
	})
}

// API Models for Swagger documentation

// LoginRequest represents login credentials
// @Description User login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"testuser"`
	Password string `json:"password" binding:"required" example:"password123"`
} // @name LoginRequest

// LoginResponse represents successful login response
// @Description JWT tokens and user information
type LoginResponse struct {
	AccessToken  string   `json:"access_token" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string   `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int      `json:"expires_in" example:"300"`
	TokenType    string   `json:"token_type" example:"Bearer"`
	User         UserInfo `json:"user"`
	TrustScore   int      `json:"trust_score" example:"88"`
} // @name LoginResponse

// UserInfo represents authenticated user information
// @Description User profile and permissions
type UserInfo struct {
	ID       string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username string   `json:"username" example:"testuser"`
	Email    string   `json:"email" example:"testuser@example.com"`
	Roles    []string `json:"roles" example:"user,admin"`
} // @name UserInfo

// TrustScoreResponse represents current trust score
// @Description Detailed trust score breakdown
type TrustScoreResponse struct {
	UserID    string         `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Overall   int            `json:"overall" example:"88"`
	Factors   map[string]int `json:"factors"`
	Timestamp string         `json:"timestamp" example:"2025-06-22T12:00:00Z"`
	NextCheck string         `json:"next_check" example:"2025-06-22T12:05:00Z"`
} // @name TrustScoreResponse

// ErrorResponse represents API error
// @Description Standard error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Unauthorized"`
	Code    string `json:"code" example:"AUTH_001"`
	Message string `json:"message" example:"Invalid or expired token"`
} // @name ErrorResponse

// HealthResponse represents service health
// @Description Service health status
type HealthResponse struct {
	Status    string `json:"status" example:"healthy"`
	Service   string `json:"service" example:"impl-zamaz"`
	Version   string `json:"version" example:"1.0.0"`
	Timestamp string `json:"timestamp" example:"2025-06-22T12:00:00Z"`
} // @name HealthResponse
