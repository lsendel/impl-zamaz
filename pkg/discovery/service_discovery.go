package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ServiceInfo represents a discovered service
type ServiceInfo struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Status      string            `json:"status"` // healthy, unhealthy, unknown
	TrustLevel  int               `json:"trust_level_required"`
	Endpoints   []EndpointInfo    `json:"endpoints"`
	LastChecked time.Time         `json:"last_checked"`
	Metadata    map[string]string `json:"metadata"`
}

// EndpointInfo represents a service endpoint
type EndpointInfo struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Description string   `json:"description"`
	TrustLevel  int      `json:"trust_level_required"`
	Scopes      []string `json:"scopes,omitempty"`
}

// ServiceRegistry manages service discovery
type ServiceRegistry struct {
	services map[string]*ServiceInfo
	mu       sync.RWMutex
	checker  *HealthChecker
}

// HealthChecker performs health checks on services
type HealthChecker struct {
	client  *http.Client
	timeout time.Duration
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*ServiceInfo),
		checker: &HealthChecker{
			client: &http.Client{
				Timeout: 5 * time.Second,
			},
			timeout: 5 * time.Second,
		},
	}
}

// RegisterService registers a new service
func (sr *ServiceRegistry) RegisterService(service *ServiceInfo) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if service.Name == "" || service.URL == "" {
		return fmt.Errorf("service name and URL are required")
	}

	sr.services[service.Name] = service

	// Perform initial health check
	go sr.checkServiceHealth(service.Name)

	return nil
}

// GetService retrieves a service by name
func (sr *ServiceRegistry) GetService(name string) (*ServiceInfo, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	service, exists := sr.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// ListServices returns all registered services
func (sr *ServiceRegistry) ListServices() []*ServiceInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(sr.services))
	for _, service := range sr.services {
		services = append(services, service)
	}

	return services
}

// ListHealthyServices returns only healthy services
func (sr *ServiceRegistry) ListHealthyServices() []*ServiceInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	services := make([]*ServiceInfo, 0)
	for _, service := range sr.services {
		if service.Status == "healthy" {
			services = append(services, service)
		}
	}

	return services
}

// ListServicesByTrustLevel returns services accessible at given trust level
func (sr *ServiceRegistry) ListServicesByTrustLevel(trustLevel int) []*ServiceInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	services := make([]*ServiceInfo, 0)
	for _, service := range sr.services {
		if trustLevel >= service.TrustLevel {
			services = append(services, service)
		}
	}

	return services
}

// StartHealthChecks starts periodic health checks
func (sr *ServiceRegistry) StartHealthChecks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sr.checkAllServices()
		case <-ctx.Done():
			return
		}
	}
}

// checkAllServices checks health of all registered services
func (sr *ServiceRegistry) checkAllServices() {
	sr.mu.RLock()
	serviceNames := make([]string, 0, len(sr.services))
	for name := range sr.services {
		serviceNames = append(serviceNames, name)
	}
	sr.mu.RUnlock()

	var wg sync.WaitGroup
	for _, name := range serviceNames {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()
			sr.checkServiceHealth(serviceName)
		}(name)
	}
	wg.Wait()
}

// checkServiceHealth checks the health of a specific service
func (sr *ServiceRegistry) checkServiceHealth(name string) {
	sr.mu.RLock()
	service, exists := sr.services[name]
	sr.mu.RUnlock()

	if !exists {
		return
	}

	healthURL := fmt.Sprintf("%s/health", service.URL)
	resp, err := sr.checker.client.Get(healthURL)

	sr.mu.Lock()
	defer sr.mu.Unlock()

	service.LastChecked = time.Now()

	if err != nil {
		service.Status = "unhealthy"
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		service.Status = "healthy"
	} else {
		service.Status = "unhealthy"
	}
}

// ServiceDiscoveryHandler provides HTTP endpoints for service discovery
type ServiceDiscoveryHandler struct {
	registry *ServiceRegistry
}

// NewServiceDiscoveryHandler creates a new handler
func NewServiceDiscoveryHandler(registry *ServiceRegistry) *ServiceDiscoveryHandler {
	return &ServiceDiscoveryHandler{
		registry: registry,
	}
}

// HandleListServices returns all services
func (h *ServiceDiscoveryHandler) HandleListServices(w http.ResponseWriter, r *http.Request) {
	services := h.registry.ListServices()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services":  services,
		"count":     len(services),
		"timestamp": time.Now(),
	})
}

// HandleGetService returns a specific service
func (h *ServiceDiscoveryHandler) HandleGetService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("name")
	if serviceName == "" {
		http.Error(w, "service name required", http.StatusBadRequest)
		return
	}

	service, err := h.registry.GetService(serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// HandleRegisterService registers a new service
func (h *ServiceDiscoveryHandler) HandleRegisterService(w http.ResponseWriter, r *http.Request) {
	var service ServiceInfo
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.registry.RegisterService(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "registered",
		"name":   service.Name,
	})
}

// InitializeDefaultServices registers default services
func InitializeDefaultServices(registry *ServiceRegistry) {
	defaultServices := []*ServiceInfo{
		{
			Name:       "api-gateway",
			URL:        "http://localhost:8080",
			TrustLevel: 0,
			Endpoints: []EndpointInfo{
				{Path: "/health", Method: "GET", Description: "Health check", TrustLevel: 0},
				{Path: "/api/v1/auth/login", Method: "POST", Description: "User login", TrustLevel: 0},
				{Path: "/api/v1/trust-score", Method: "GET", Description: "Get trust score", TrustLevel: 25},
				{Path: "/api/v1/protected", Method: "GET", Description: "Protected resource", TrustLevel: 50},
			},
			Metadata: map[string]string{
				"type":    "gateway",
				"version": "1.0.0",
			},
		},
		{
			Name:       "keycloak",
			URL:        "http://localhost:8082",
			TrustLevel: 0,
			Endpoints: []EndpointInfo{
				{Path: "/admin", Method: "GET", Description: "Admin console", TrustLevel: 75},
				{Path: "/realms/zerotrust-test", Method: "GET", Description: "Realm info", TrustLevel: 0},
			},
			Metadata: map[string]string{
				"type":    "identity-provider",
				"version": "22.0.5",
			},
		},
		{
			Name:       "user-service",
			URL:        "http://localhost:8081",
			TrustLevel: 25,
			Endpoints: []EndpointInfo{
				{Path: "/api/users", Method: "GET", Description: "List users", TrustLevel: 25},
				{Path: "/api/users/{id}", Method: "GET", Description: "Get user", TrustLevel: 25},
				{Path: "/api/users/{id}", Method: "PUT", Description: "Update user", TrustLevel: 50},
			},
			Metadata: map[string]string{
				"type":   "microservice",
				"domain": "user-management",
			},
		},
		{
			Name:       "admin-service",
			URL:        "http://localhost:8083",
			TrustLevel: 75,
			Endpoints: []EndpointInfo{
				{Path: "/api/admin/users", Method: "DELETE", Description: "Delete user", TrustLevel: 90},
				{Path: "/api/admin/audit", Method: "GET", Description: "View audit logs", TrustLevel: 75},
				{Path: "/api/admin/config", Method: "PUT", Description: "Update config", TrustLevel: 90},
			},
			Metadata: map[string]string{
				"type":   "microservice",
				"domain": "administration",
			},
		},
		{
			Name:       "audit-service",
			URL:        "http://localhost:8084",
			TrustLevel: 50,
			Endpoints: []EndpointInfo{
				{Path: "/api/audit/logs", Method: "GET", Description: "Get audit logs", TrustLevel: 50},
				{Path: "/api/audit/export", Method: "POST", Description: "Export audit logs", TrustLevel: 75},
			},
			Metadata: map[string]string{
				"type":   "microservice",
				"domain": "compliance",
			},
		},
	}

	for _, service := range defaultServices {
		registry.RegisterService(service)
	}
}
