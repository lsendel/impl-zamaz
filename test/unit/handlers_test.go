package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lsendel/impl-zamaz/api"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestHandlersHealth(t *testing.T) {
	router := setupTestRouter()
	handlers := api.NewHandlers()

	router.GET("/health", handlers.Health)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response api.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "impl-zamaz", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
	assert.NotEmpty(t, response.Timestamp)
}

func TestHandlersLogin(t *testing.T) {
	router := setupTestRouter()
	handlers := api.NewHandlers()

	router.POST("/login", handlers.Login)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid Login Request",
			payload: api.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Missing Username",
			payload: api.LoginRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Missing Password",
			payload: api.LoginRequest{
				Username: "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Invalid JSON",
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Empty Request",
			payload:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.payload.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.payload)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response api.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Equal(t, 300, response.ExpiresIn)
				assert.Equal(t, 88, response.TrustScore)
				assert.NotEmpty(t, response.User.ID)
				assert.Equal(t, "testuser", response.User.Username)
			}

			if tt.expectError {
				var errorResponse api.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				assert.NotEmpty(t, errorResponse.Error)
			}
		})
	}
}

func TestHandlersGetTrustScore(t *testing.T) {
	router := setupTestRouter()
	handlers := api.NewHandlers()

	router.GET("/trust-score", handlers.GetTrustScore)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/trust-score", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response api.TrustScoreResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.UserID)
	assert.Equal(t, 88, response.Overall)
	assert.NotEmpty(t, response.Factors)
	assert.Contains(t, response.Factors, "identity")
	assert.Contains(t, response.Factors, "device")
	assert.Contains(t, response.Factors, "behavior")
	assert.Contains(t, response.Factors, "location")
	assert.Contains(t, response.Factors, "risk")
	assert.NotEmpty(t, response.Timestamp)
	assert.NotEmpty(t, response.NextCheck)
}

func TestHandlersGetProtectedResource(t *testing.T) {
	router := setupTestRouter()
	handlers := api.NewHandlers()

	router.GET("/protected", handlers.GetProtectedResource)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "This is protected data", response["data"])
	assert.Equal(t, float64(50), response["trust_level_required"])
	assert.Equal(t, float64(88), response["your_trust_level"])
	assert.NotEmpty(t, response["accessed_at"])
}

func TestAPIResponseHeaders(t *testing.T) {
	router := setupTestRouter()
	handlers := api.NewHandlers()

	router.GET("/test", handlers.Health)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestAPIErrorHandling(t *testing.T) {
	router := setupTestRouter()

	// Test 404 endpoint
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
