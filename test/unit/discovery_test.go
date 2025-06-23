package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lsendel/impl-zamaz/pkg/discovery"
)

func TestServiceRegistry(t *testing.T) {
	registry := discovery.NewServiceRegistry()
	assert.NotNil(t, registry)
}

func TestServiceRegistration(t *testing.T) {
	registry := discovery.NewServiceRegistry()

	service := &discovery.ServiceInfo{
		Name:       "test-service",
		URL:        "http://localhost:8080",
		TrustLevel: 25,
		Endpoints: []discovery.EndpointInfo{
			{
				Path:        "/api/test",
				Method:      "GET",
				Description: "Test endpoint",
				TrustLevel:  25,
			},
		},
		Metadata: map[string]string{
			"type":    "test",
			"version": "1.0.0",
		},
	}

	err := registry.RegisterService(service)
	assert.NoError(t, err)

	// Retrieve the service
	retrievedService, err := registry.GetService("test-service")
	assert.NoError(t, err)
	assert.Equal(t, service.Name, retrievedService.Name)
	assert.Equal(t, service.URL, retrievedService.URL)
	assert.Equal(t, service.TrustLevel, retrievedService.TrustLevel)
}

func TestServiceRegistrationValidation(t *testing.T) {
	registry := discovery.NewServiceRegistry()

	tests := []struct {
		name        string
		service     *discovery.ServiceInfo
		expectError bool
	}{
		{
			name: "Valid Service",
			service: &discovery.ServiceInfo{
				Name: "valid-service",
				URL:  "http://localhost:8080",
			},
			expectError: false,
		},
		{
			name: "Missing Name",
			service: &discovery.ServiceInfo{
				URL: "http://localhost:8080",
			},
			expectError: true,
		},
		{
			name: "Missing URL",
			service: &discovery.ServiceInfo{
				Name: "invalid-service",
			},
			expectError: true,
		},
		{
			name: "Empty Name",
			service: &discovery.ServiceInfo{
				Name: "",
				URL:  "http://localhost:8080",
			},
			expectError: true,
		},
		{
			name: "Empty URL",
			service: &discovery.ServiceInfo{
				Name: "invalid-service",
				URL:  "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterService(tt.service)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListServices(t *testing.T) {
	registry := discovery.NewServiceRegistry()

	// Register multiple services
	services := []*discovery.ServiceInfo{
		{Name: "service1", URL: "http://localhost:8081", TrustLevel: 25},
		{Name: "service2", URL: "http://localhost:8082", TrustLevel: 50},
		{Name: "service3", URL: "http://localhost:8083", TrustLevel: 75},
	}

	for _, service := range services {
		err := registry.RegisterService(service)
		require.NoError(t, err)
	}

	// List all services
	allServices := registry.ListServices()
	assert.Len(t, allServices, 3)

	// Test service names
	serviceNames := make(map[string]bool)
	for _, service := range allServices {
		serviceNames[service.Name] = true
	}
	assert.True(t, serviceNames["service1"])
	assert.True(t, serviceNames["service2"])
	assert.True(t, serviceNames["service3"])
}

func TestListServicesByTrustLevel(t *testing.T) {
	registry := discovery.NewServiceRegistry()

	// Register services with different trust levels
	services := []*discovery.ServiceInfo{
		{Name: "low-trust", URL: "http://localhost:8081", TrustLevel: 25},
		{Name: "medium-trust", URL: "http://localhost:8082", TrustLevel: 50},
		{Name: "high-trust", URL: "http://localhost:8083", TrustLevel: 75},
	}

	for _, service := range services {
		err := registry.RegisterService(service)
		require.NoError(t, err)
	}

	// Test different trust levels
	tests := []struct {
		trustLevel       int
		expectedCount    int
		expectedServices []string
	}{
		{
			trustLevel:       25,
			expectedCount:    1,
			expectedServices: []string{"low-trust"},
		},
		{
			trustLevel:       50,
			expectedCount:    2,
			expectedServices: []string{"low-trust", "medium-trust"},
		},
		{
			trustLevel:       75,
			expectedCount:    3,
			expectedServices: []string{"low-trust", "medium-trust", "high-trust"},
		},
		{
			trustLevel:       100,
			expectedCount:    3,
			expectedServices: []string{"low-trust", "medium-trust", "high-trust"},
		},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.trustLevel)), func(t *testing.T) {
			accessibleServices := registry.ListServicesByTrustLevel(tt.trustLevel)
			assert.Len(t, accessibleServices, tt.expectedCount)

			serviceNames := make(map[string]bool)
			for _, service := range accessibleServices {
				serviceNames[service.Name] = true
			}

			for _, expectedService := range tt.expectedServices {
				assert.True(t, serviceNames[expectedService], "Expected service %s not found", expectedService)
			}
		})
	}
}

func TestServiceDiscoveryHTTPHandlers(t *testing.T) {
	registry := discovery.NewServiceRegistry()
	handler := discovery.NewServiceDiscoveryHandler(registry)

	// Register a test service
	testService := &discovery.ServiceInfo{
		Name:       "test-api",
		URL:        "http://localhost:8080",
		TrustLevel: 25,
		Endpoints: []discovery.EndpointInfo{
			{Path: "/api/test", Method: "GET", Description: "Test endpoint", TrustLevel: 25},
		},
		Metadata: map[string]string{"type": "api", "version": "1.0.0"},
	}

	err := registry.RegisterService(testService)
	require.NoError(t, err)

	t.Run("List Services", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/services", nil)
		w := httptest.NewRecorder()

		handler.HandleListServices(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		services, ok := response["services"].([]interface{})
		require.True(t, ok)
		assert.Len(t, services, 1)

		count, ok := response["count"].(float64)
		require.True(t, ok)
		assert.Equal(t, float64(1), count)
	})

	t.Run("Get Specific Service", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/services?name=test-api", nil)
		w := httptest.NewRecorder()

		handler.HandleGetService(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var service discovery.ServiceInfo
		err := json.Unmarshal(w.Body.Bytes(), &service)
		require.NoError(t, err)

		assert.Equal(t, "test-api", service.Name)
		assert.Equal(t, "http://localhost:8080", service.URL)
		assert.Equal(t, 25, service.TrustLevel)
	})

	t.Run("Get Non-existent Service", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/services?name=non-existent", nil)
		w := httptest.NewRecorder()

		handler.HandleGetService(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Register New Service", func(t *testing.T) {
		newService := discovery.ServiceInfo{
			Name:       "new-service",
			URL:        "http://localhost:9090",
			TrustLevel: 50,
		}

		body, err := json.Marshal(newService)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/services", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegisterService(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "registered", response["status"])
		assert.Equal(t, "new-service", response["name"])
	})
}

func TestHealthChecks(t *testing.T) {
	registry := discovery.NewServiceRegistry()

	// Create a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "healthy"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer testServer.Close()

	// Register service with test server URL
	service := &discovery.ServiceInfo{
		Name:       "test-service",
		URL:        testServer.URL,
		TrustLevel: 25,
	}

	err := registry.RegisterService(service)
	require.NoError(t, err)

	// Start health checks with short interval for testing
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go registry.StartHealthChecks(ctx, 1*time.Second)

	// Wait for health check to complete
	time.Sleep(2 * time.Second)

	// Verify service status
	healthyServices := registry.ListHealthyServices()
	assert.Len(t, healthyServices, 1)
	assert.Equal(t, "healthy", healthyServices[0].Status)
}

func TestInitializeDefaultServices(t *testing.T) {
	registry := discovery.NewServiceRegistry()

	discovery.InitializeDefaultServices(registry)

	services := registry.ListServices()
	assert.GreaterOrEqual(t, len(services), 4) // At least 4 default services

	// Check for expected default services
	serviceNames := make(map[string]bool)
	for _, service := range services {
		serviceNames[service.Name] = true
	}

	expectedServices := []string{"api-gateway", "keycloak", "user-service", "admin-service", "audit-service"}
	for _, expectedService := range expectedServices {
		assert.True(t, serviceNames[expectedService], "Expected default service %s not found", expectedService)
	}
}
