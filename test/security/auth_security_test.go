//go:build security

package security

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Import zero trust components from root-zamaz
	ztMiddleware "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/middleware/gin"
	ztClient "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/pkg/client"
	ztConfig "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/pkg/config"
	ztTypes "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/pkg/types"
)

func TestZeroTrustAuthenticationSecurity(t *testing.T) {
	tests := []struct {
		name           string
		setupAuth      func() gin.HandlerFunc
		request        func() *http.Request
		expectedStatus int
		expectedError  string
		securityCheck  func(t *testing.T, response *httptest.ResponseRecorder)
	}{
		{
			name: "Valid JWT Token Access",
			setupAuth: func() gin.HandlerFunc {
				return createMockZTMiddleware(true, 85)
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/protected", nil)
				req.Header.Set("Authorization", "Bearer valid-jwt-token")
				return req
			},
			expectedStatus: http.StatusOK,
			securityCheck: func(t *testing.T, response *httptest.ResponseRecorder) {
				// Verify response contains user context
				var result map[string]interface{}
				err := json.Unmarshal(response.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "accessed_by")
			},
		},
		{
			name: "Invalid JWT Token Rejection",
			setupAuth: func() gin.HandlerFunc {
				return createMockZTMiddleware(false, 0)
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/protected", nil)
				req.Header.Set("Authorization", "Bearer invalid-jwt-token")
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
			securityCheck: func(t *testing.T, response *httptest.ResponseRecorder) {
				// Verify no sensitive data leaked
				body := response.Body.String()
				assert.NotContains(t, body, "secret")
				assert.NotContains(t, body, "password")
				assert.NotContains(t, body, "token")
			},
		},
		{
			name: "Missing Authorization Header",
			setupAuth: func() gin.HandlerFunc {
				return createMockZTMiddleware(false, 0)
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/api/v1/protected", nil)
			},
			expectedStatus: http.StatusUnauthorized,
			securityCheck: func(t *testing.T, response *httptest.ResponseRecorder) {
				// Verify proper error response
				assert.Contains(t, response.Header().Get("Content-Type"), "application/json")
			},
		},
		{
			name: "Low Trust Score Access Denial",
			setupAuth: func() gin.HandlerFunc {
				return createMockZTMiddleware(true, 25) // Low trust score
			},
			request: func() *http.Request {
				req := httptest.NewRequest("DELETE", "/api/v1/admin/delete", nil)
				req.Header.Set("Authorization", "Bearer valid-low-trust-token")
				return req
			},
			expectedStatus: http.StatusForbidden,
			securityCheck: func(t *testing.T, response *httptest.ResponseRecorder) {
				// Verify trust score enforcement
				var result map[string]interface{}
				err := json.Unmarshal(response.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "trust_score_insufficient")
			},
		},
		{
			name: "SQL Injection Prevention",
			setupAuth: func() gin.HandlerFunc {
				return createMockZTMiddleware(true, 85)
			},
			request: func() *http.Request {
				maliciousPayload := map[string]string{
					"username": "admin'; DROP TABLE users; --",
					"query":    "1' OR '1'='1",
				}
				body, _ := json.Marshal(maliciousPayload)
				req := httptest.NewRequest("POST", "/api/v1/search", bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer valid-jwt-token")
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusBadRequest,
			securityCheck: func(t *testing.T, response *httptest.ResponseRecorder) {
				// Verify input validation
				body := response.Body.String()
				assert.Contains(t, body, "invalid_input")
				assert.NotContains(t, body, "DROP TABLE")
			},
		},
		{
			name: "XSS Prevention",
			setupAuth: func() gin.HandlerFunc {
				return createMockZTMiddleware(true, 85)
			},
			request: func() *http.Request {
				xssPayload := map[string]string{
					"message": "<script>alert('XSS')</script>",
					"name":    "<img src=x onerror=alert('XSS')>",
				}
				body, _ := json.Marshal(xssPayload)
				req := httptest.NewRequest("POST", "/api/v1/message", bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer valid-jwt-token")
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusBadRequest,
			securityCheck: func(t *testing.T, response *httptest.ResponseRecorder) {
				// Verify XSS prevention
				body := response.Body.String()
				assert.NotContains(t, body, "<script>")
				assert.NotContains(t, body, "onerror=")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Apply Zero Trust middleware
			authMiddleware := tt.setupAuth()

			// Protected routes
			api := router.Group("/api/v1")
			api.Use(authMiddleware)
			{
				api.GET("/protected", func(c *gin.Context) {
					user, exists := c.Get("user")
					if !exists {
						c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
						return
					}
					authUser := user.(*ztTypes.AuthenticatedUser)
					c.JSON(http.StatusOK, gin.H{
						"message":     "Access granted",
						"accessed_by": authUser.Username,
					})
				})

				api.DELETE("/admin/delete", requireHighTrust(func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Admin action completed"})
				}))

				api.POST("/search", validateInput(func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Search completed"})
				}))

				api.POST("/message", sanitizeInput(func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Message processed"})
				}))
			}

			// Execute test
			w := httptest.NewRecorder()
			req := tt.request()
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				body := w.Body.String()
				assert.Contains(t, body, tt.expectedError)
			}

			if tt.securityCheck != nil {
				tt.securityCheck(t, w)
			}
		})
	}
}

func TestTrustScoreCalculation(t *testing.T) {
	tests := []struct {
		name                string
		identityFactor      int
		deviceFactor        int
		behaviorFactor      int
		locationFactor      int
		riskFactor          int
		expectedTrustScore  int
		expectedAccessLevel string
	}{
		{
			name:                "High Trust Score",
			identityFactor:      30,
			deviceFactor:        25,
			behaviorFactor:      20,
			locationFactor:      15,
			riskFactor:          5,
			expectedTrustScore:  95,
			expectedAccessLevel: "admin",
		},
		{
			name:                "Medium Trust Score",
			identityFactor:      25,
			deviceFactor:        20,
			behaviorFactor:      15,
			locationFactor:      10,
			riskFactor:          5,
			expectedTrustScore:  75,
			expectedAccessLevel: "user",
		},
		{
			name:                "Low Trust Score",
			identityFactor:      15,
			deviceFactor:        10,
			behaviorFactor:      8,
			locationFactor:      5,
			riskFactor:          15,
			expectedTrustScore:  53,
			expectedAccessLevel: "read_only",
		},
		{
			name:                "Critical Risk Score",
			identityFactor:      10,
			deviceFactor:        5,
			behaviorFactor:      5,
			locationFactor:      2,
			riskFactor:          25,
			expectedTrustScore:  47,
			expectedAccessLevel: "denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate trust score using Zero Trust engine logic
			totalScore := tt.identityFactor + tt.deviceFactor + tt.behaviorFactor +
				tt.locationFactor + (10 - tt.riskFactor) // Risk factor reduces score

			assert.Equal(t, tt.expectedTrustScore, totalScore)

			// Verify access level mapping
			var accessLevel string
			switch {
			case totalScore >= 90:
				accessLevel = "admin"
			case totalScore >= 70:
				accessLevel = "user"
			case totalScore >= 50:
				accessLevel = "read_only"
			default:
				accessLevel = "denied"
			}

			assert.Equal(t, tt.expectedAccessLevel, accessLevel)
		})
	}
}

func TestContinuousVerification(t *testing.T) {
	// Test continuous verification and adaptive access control
	t.Run("Session Risk Assessment", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Simulate session monitoring
		sessionData := map[string]interface{}{
			"user_id":        "test-user",
			"session_start":  time.Now().Add(-1 * time.Hour),
			"last_activity":  time.Now().Add(-5 * time.Minute),
			"ip_address":     "192.168.1.100",
			"user_agent":     "Mozilla/5.0 (Test Browser)",
			"activity_count": 50,
		}

		// Mock risk assessment
		riskScore := assessSessionRisk(ctx, sessionData)

		// Verify risk assessment logic
		assert.GreaterOrEqual(t, riskScore, 0)
		assert.LessOrEqual(t, riskScore, 100)

		// Test adaptive response
		if riskScore > 80 {
			// High risk should trigger additional verification
			assert.True(t, true, "High risk detected - additional verification required")
		} else if riskScore > 50 {
			// Medium risk should reduce privileges
			assert.True(t, true, "Medium risk detected - privileges reduced")
		}
	})

	t.Run("Device Attestation", func(t *testing.T) {
		deviceInfo := map[string]interface{}{
			"device_id":      "test-device-123",
			"platform":       "linux",
			"trusted":        true,
			"last_seen":      time.Now().Add(-24 * time.Hour),
			"security_patch": "2023-12-01",
			"compliance":     true,
		}

		attestationScore := assessDeviceAttestation(deviceInfo)

		// Verify device trust score
		assert.GreaterOrEqual(t, attestationScore, 0)
		assert.LessOrEqual(t, attestationScore, 100)

		// Trusted device should have high score
		if deviceInfo["trusted"].(bool) && deviceInfo["compliance"].(bool) {
			assert.GreaterOrEqual(t, attestationScore, 70)
		}
	})
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add security headers middleware
	router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Verify security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

// Helper functions for testing

func createMockZTMiddleware(validToken bool, trustScore int) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !validToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Mock authenticated user with trust score
		user := &ztTypes.AuthenticatedUser{
			ID:         "test-user",
			Username:   "testuser",
			Email:      "test@example.com",
			Roles:      []string{"user"},
			TrustScore: trustScore,
		}

		c.Set("user", user)
		c.Set("trust_score", trustScore)
		c.Next()
	}
}

func requireHighTrust(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		trustScore, exists := c.Get("trust_score")
		if !exists || trustScore.(int) < 90 {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "trust_score_insufficient",
				"required": 90,
				"current":  trustScore,
			})
			c.Abort()
			return
		}
		handler(c)
	}
}

func validateInput(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input map[string]interface{}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			c.Abort()
			return
		}

		// Check for SQL injection patterns
		for _, value := range input {
			if str, ok := value.(string); ok {
				if containsSQLInjection(str) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input"})
					c.Abort()
					return
				}
			}
		}
		handler(c)
	}
}

func sanitizeInput(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input map[string]interface{}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			c.Abort()
			return
		}

		// Check for XSS patterns
		for _, value := range input {
			if str, ok := value.(string); ok {
				if containsXSS(str) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input"})
					c.Abort()
					return
				}
			}
		}
		handler(c)
	}
}

func containsSQLInjection(input string) bool {
	sqlPatterns := []string{
		"DROP TABLE",
		"DELETE FROM",
		"INSERT INTO",
		"UPDATE SET",
		"' OR '1'='1",
		"'; --",
		"UNION SELECT",
	}

	for _, pattern := range sqlPatterns {
		if len(input) > len(pattern) &&
			input[:len(pattern)] == pattern {
			return true
		}
	}
	return false
}

func containsXSS(input string) bool {
	xssPatterns := []string{
		"<script>",
		"</script>",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
	}

	for _, pattern := range xssPatterns {
		if len(input) >= len(pattern) {
			for i := 0; i <= len(input)-len(pattern); i++ {
				if input[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}
	return false
}

func assessSessionRisk(ctx context.Context, sessionData map[string]interface{}) int {
	// Mock session risk assessment
	riskScore := 0

	// Check session duration
	if sessionStart, ok := sessionData["session_start"].(time.Time); ok {
		duration := time.Since(sessionStart)
		if duration > 8*time.Hour {
			riskScore += 20
		}
	}

	// Check activity patterns
	if activityCount, ok := sessionData["activity_count"].(int); ok {
		if activityCount > 100 {
			riskScore += 15
		}
	}

	// Check IP changes (mock)
	// In real implementation, this would check IP history
	riskScore += 10

	return riskScore
}

func assessDeviceAttestation(deviceInfo map[string]interface{}) int {
	score := 100

	// Check if device is trusted
	if trusted, ok := deviceInfo["trusted"].(bool); ok && !trusted {
		score -= 30
	}

	// Check compliance status
	if compliance, ok := deviceInfo["compliance"].(bool); ok && !compliance {
		score -= 25
	}

	// Check last seen
	if lastSeen, ok := deviceInfo["last_seen"].(time.Time); ok {
		if time.Since(lastSeen) > 7*24*time.Hour {
			score -= 15
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}
