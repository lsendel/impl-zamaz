# impl-zamaz - Zero Trust Integration Test Makefile
# This Makefile provides commands to install, run, and test Zero Trust integration

# ================================
# Configuration
# ================================

# Go configuration
GO_VERSION := 1.21
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GORUN := $(GOCMD) run

# Project configuration
PROJECT_NAME := impl-zamaz
MAIN_PATH := ./cmd/server
BUILD_DIR := ./build
TEST_TIMEOUT := 5m

# Docker configuration
DOCKER_COMPOSE := docker-compose
APP_CONTAINER := impl-zamaz-app
KEYCLOAK_CONTAINER := impl-zamaz-keycloak
POSTGRES_CONTAINER := impl-zamaz-postgres
REDIS_CONTAINER := impl-zamaz-redis

# Test configuration
COVERAGE_OUT := coverage.out
COVERAGE_HTML := coverage.html

# Security analysis configuration
GOLANGCI_LINT_VERSION := v1.55.2
GOSEC_VERSION := v2.18.2
NANCY_VERSION := v1.0.42
SECURITY_REPORT := security-report.json

# Environment files
ENV_FILE := .env
ENV_TEMPLATE := .env.template

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
NC := \033[0m # No Color
BOLD := \033[1m

# ================================
# Help Documentation
# ================================

.PHONY: help
help: ## 📚 Show this help message
	@echo "$(CYAN)$(BOLD)impl-zamaz Zero Trust Integration$(NC)"
	@echo "$(CYAN)================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(CYAN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Quick Start:$(NC)"
	@echo "  make setup     # Complete setup and installation"
	@echo "  make start     # Start all services"
	@echo "  make test-e2e  # Run end-to-end tests"
	@echo "  make demo      # Run interactive demo"

# ================================
# Installation & Setup
# ================================

.PHONY: check-deps
check-deps: ## 🔍 Check system dependencies
	@echo "$(BLUE)Checking system dependencies...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)❌ Go not installed$(NC)"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)❌ Docker not installed$(NC)"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "$(RED)❌ Docker Compose not installed$(NC)"; exit 1; }
	@command -v curl >/dev/null 2>&1 || { echo "$(RED)❌ curl not installed$(NC)"; exit 1; }
	@echo "$(GREEN)✅ All dependencies available$(NC)"

.PHONY: install-zerotrust
install-zerotrust: ## 📦 Install Zero Trust components
	@echo "$(BLUE)Installing Zero Trust components...$(NC)"
	@if [ -f "./integrate-zerotrust.sh" ]; then \
		./integrate-zerotrust.sh || { echo "$(RED)❌ Integration script failed$(NC)"; exit 1; }; \
	else \
		echo "$(YELLOW)Integration script not found, running manual installation...$(NC)"; \
		$(MAKE) manual-install; \
	fi
	@echo "$(GREEN)✅ Zero Trust components installed$(NC)"

.PHONY: manual-install
manual-install: ## 📦 Manual installation of components
	@echo "$(BLUE)Manual installation of Zero Trust components...$(NC)"
	@if [ ! -f "go.mod" ]; then \
		echo "$(YELLOW)Initializing Go module...$(NC)"; \
		go mod init github.com/lsendel/impl-zamaz; \
	fi
	@echo "$(YELLOW)Installing core component...$(NC)"
	@$(GOGET) github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/components/core@latest || \
		$(GOGET) github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust@latest
	@echo "$(YELLOW)Installing middleware component...$(NC)"
	@$(GOGET) github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/components/middleware@latest || \
		$(GOGET) github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/middleware@latest
	@echo "$(YELLOW)Installing framework dependencies...$(NC)"
	@$(GOGET) github.com/gin-gonic/gin@latest
	@$(GOGET) github.com/golang-jwt/jwt/v5@latest
	@$(GOGET) golang.org/x/crypto@latest
	@$(GOMOD) tidy
	@echo "$(GREEN)✅ Manual installation completed$(NC)"

.PHONY: setup-env
setup-env: ## ⚙️ Setup environment configuration
	@echo "$(BLUE)Setting up environment configuration...$(NC)"
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "$(YELLOW)Creating .env file...$(NC)"; \
		echo "# Keycloak Configuration" > $(ENV_FILE); \
		echo "KEYCLOAK_BASE_URL=http://localhost:8082" >> $(ENV_FILE); \
		echo "KEYCLOAK_REALM=zerotrust-test" >> $(ENV_FILE); \
		echo "KEYCLOAK_CLIENT_ID=zerotrust-client" >> $(ENV_FILE); \
		echo "KEYCLOAK_CLIENT_SECRET=zerotrust-secret-12345" >> $(ENV_FILE); \
		echo "" >> $(ENV_FILE); \
		echo "# Zero Trust Configuration" >> $(ENV_FILE); \
		echo "ZEROTRUST_TRUST_LEVEL_READ=25" >> $(ENV_FILE); \
		echo "ZEROTRUST_TRUST_LEVEL_WRITE=50" >> $(ENV_FILE); \
		echo "ZEROTRUST_TRUST_LEVEL_ADMIN=75" >> $(ENV_FILE); \
		echo "ZEROTRUST_TRUST_LEVEL_DELETE=90" >> $(ENV_FILE); \
		echo "" >> $(ENV_FILE); \
		echo "# Cache Configuration" >> $(ENV_FILE); \
		echo "CACHE_TYPE=redis" >> $(ENV_FILE); \
		echo "CACHE_TTL=15m" >> $(ENV_FILE); \
		echo "REDIS_URL=redis://localhost:6380" >> $(ENV_FILE); \
		echo "" >> $(ENV_FILE); \
		echo "# Security Configuration" >> $(ENV_FILE); \
		echo "DEVICE_ATTESTATION_ENABLED=true" >> $(ENV_FILE); \
		echo "RISK_ASSESSMENT_ENABLED=true" >> $(ENV_FILE); \
		echo "CONTINUOUS_VERIFICATION=true" >> $(ENV_FILE); \
		echo "$(GREEN)✅ Environment file created$(NC)"; \
	else \
		echo "$(YELLOW)Using existing .env file$(NC)"; \
	fi

.PHONY: create-main
create-main: ## 🔧 Create main application file
	@echo "$(BLUE)Creating main application...$(NC)"
	@mkdir -p cmd/server
	@cp ../projects/root-zamaz/libraries/go-keycloak-zerotrust/scripts/create-main.go cmd/server/main.go 2>/dev/null || \
		echo 'package main\n\nimport (\n\t"fmt"\n\t"log"\n\t"net/http"\n\t"github.com/gin-gonic/gin"\n)\n\nfunc main() {\n\tr := gin.Default()\n\tr.GET("/health", func(c *gin.Context) {\n\t\tc.JSON(200, gin.H{"status": "healthy", "service": "impl-zamaz"})\n\t})\n\tlog.Println("🚀 Server starting on :8080")\n\tlog.Fatal(http.ListenAndServe(":8080", r))\n}' > cmd/server/main.go
	@echo "$(GREEN)✅ Main application created$(NC)"

.PHONY: setup
setup: check-deps install-zerotrust setup-env create-main ## 🛠️ Complete setup process
	@echo "$(GREEN)✅ Setup completed successfully!$(NC)"
	@echo ""
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. make start      # Start all services"
	@echo "  2. make test-e2e   # Run end-to-end tests"
	@echo "  3. make demo       # Run interactive demo"

# ================================
# Service Management
# ================================

.PHONY: create-docker-compose
create-docker-compose: ## 🐳 Create Docker Compose configuration
	@echo "$(BLUE)Creating Docker Compose configuration...$(NC)"
	@echo 'version: "3.8"' > docker-compose.yml
	@echo '' >> docker-compose.yml
	@echo 'services:' >> docker-compose.yml
	@echo '  app:' >> docker-compose.yml
	@echo '    build: .' >> docker-compose.yml
	@echo '    container_name: impl-zamaz-app' >> docker-compose.yml
	@echo '    ports:' >> docker-compose.yml
	@echo '      - "8080:8080"' >> docker-compose.yml
	@echo '    environment:' >> docker-compose.yml
	@echo '      - KEYCLOAK_BASE_URL=http://keycloak:8080' >> docker-compose.yml
	@echo '      - KEYCLOAK_REALM=zerotrust-test' >> docker-compose.yml
	@echo '      - KEYCLOAK_CLIENT_ID=zerotrust-client' >> docker-compose.yml
	@echo '      - KEYCLOAK_CLIENT_SECRET=zerotrust-secret-12345' >> docker-compose.yml
	@echo '    depends_on:' >> docker-compose.yml
	@echo '      - keycloak' >> docker-compose.yml
	@echo '      - redis' >> docker-compose.yml
	@echo '    restart: unless-stopped' >> docker-compose.yml
	@echo '    networks:' >> docker-compose.yml
	@echo '      - zerotrust-network' >> docker-compose.yml
	@echo '' >> docker-compose.yml
	@echo '  keycloak:' >> docker-compose.yml
	@echo '    image: quay.io/keycloak/keycloak:22.0.5' >> docker-compose.yml
	@echo '    container_name: impl-zamaz-keycloak' >> docker-compose.yml
	@echo '    command: start-dev --import-realm' >> docker-compose.yml
	@echo '    environment:' >> docker-compose.yml
	@echo '      - KEYCLOAK_ADMIN=admin' >> docker-compose.yml
	@echo '      - KEYCLOAK_ADMIN_PASSWORD=admin' >> docker-compose.yml
	@echo '      - KC_DB=postgres' >> docker-compose.yml
	@echo '      - KC_DB_URL=jdbc:postgresql://postgres:5432/keycloak' >> docker-compose.yml
	@echo '      - KC_DB_USERNAME=keycloak' >> docker-compose.yml
	@echo '      - KC_DB_PASSWORD=keycloak_password' >> docker-compose.yml
	@echo '    ports:' >> docker-compose.yml
	@echo '      - "8082:8080"' >> docker-compose.yml
	@echo '    volumes:' >> docker-compose.yml
	@echo '      - ./keycloak/imports:/opt/keycloak/data/import' >> docker-compose.yml
	@echo '    depends_on:' >> docker-compose.yml
	@echo '      - postgres' >> docker-compose.yml
	@echo '    restart: unless-stopped' >> docker-compose.yml
	@echo '    networks:' >> docker-compose.yml
	@echo '      - zerotrust-network' >> docker-compose.yml
	@echo '' >> docker-compose.yml
	@echo '  postgres:' >> docker-compose.yml
	@echo '    image: postgres:15-alpine' >> docker-compose.yml
	@echo '    container_name: impl-zamaz-postgres' >> docker-compose.yml
	@echo '    environment:' >> docker-compose.yml
	@echo '      - POSTGRES_DB=postgres' >> docker-compose.yml
	@echo '      - POSTGRES_USER=postgres' >> docker-compose.yml
	@echo '      - POSTGRES_PASSWORD=postgres_password' >> docker-compose.yml
	@echo '    volumes:' >> docker-compose.yml
	@echo '      - postgres_data:/var/lib/postgresql/data' >> docker-compose.yml
	@echo '    ports:' >> docker-compose.yml
	@echo '      - "5433:5432"' >> docker-compose.yml
	@echo '    restart: unless-stopped' >> docker-compose.yml
	@echo '    networks:' >> docker-compose.yml
	@echo '      - zerotrust-network' >> docker-compose.yml
	@echo '' >> docker-compose.yml
	@echo '  redis:' >> docker-compose.yml
	@echo '    image: redis:7-alpine' >> docker-compose.yml
	@echo '    container_name: impl-zamaz-redis' >> docker-compose.yml
	@echo '    ports:' >> docker-compose.yml
	@echo '      - "6380:6379"' >> docker-compose.yml
	@echo '    restart: unless-stopped' >> docker-compose.yml
	@echo '    networks:' >> docker-compose.yml
	@echo '      - zerotrust-network' >> docker-compose.yml
	@echo '' >> docker-compose.yml
	@echo 'volumes:' >> docker-compose.yml
	@echo '  postgres_data:' >> docker-compose.yml
	@echo '' >> docker-compose.yml
	@echo 'networks:' >> docker-compose.yml
	@echo '  zerotrust-network:' >> docker-compose.yml
	@echo '    driver: bridge' >> docker-compose.yml
	@echo "$(GREEN)✅ Docker Compose configuration created$(NC)"

.PHONY: create-dockerfile
create-dockerfile: ## 🐳 Create Dockerfile
	@echo "$(BLUE)Creating Dockerfile...$(NC)"
	@echo '# Build stage' > Dockerfile
	@echo 'FROM golang:1.21-alpine AS builder' >> Dockerfile
	@echo '' >> Dockerfile
	@echo 'WORKDIR /app' >> Dockerfile
	@echo '' >> Dockerfile
	@echo '# Copy go mod files' >> Dockerfile
	@echo 'COPY go.mod go.sum ./' >> Dockerfile
	@echo 'RUN go mod download' >> Dockerfile
	@echo '' >> Dockerfile
	@echo '# Copy source code' >> Dockerfile
	@echo 'COPY . .' >> Dockerfile
	@echo '' >> Dockerfile
	@echo '# Build binary' >> Dockerfile
	@echo 'RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server' >> Dockerfile
	@echo '' >> Dockerfile
	@echo '# Final stage' >> Dockerfile
	@echo 'FROM alpine:latest' >> Dockerfile
	@echo '' >> Dockerfile
	@echo 'RUN apk --no-cache add ca-certificates tzdata curl' >> Dockerfile
	@echo 'WORKDIR /root/' >> Dockerfile
	@echo '' >> Dockerfile
	@echo '# Copy binary from builder stage' >> Dockerfile
	@echo 'COPY --from=builder /app/main .' >> Dockerfile
	@echo '' >> Dockerfile
	@echo 'EXPOSE 8080' >> Dockerfile
	@echo '' >> Dockerfile
	@echo 'CMD ["./main"]' >> Dockerfile
	@echo "$(GREEN)✅ Dockerfile created$(NC)"

.PHONY: copy-keycloak-config
copy-keycloak-config: ## 🔐 Copy Keycloak configuration from main project
	@echo "$(BLUE)Copying Keycloak configuration...$(NC)"
	@mkdir -p keycloak/imports scripts
	@cp -r ../projects/root-zamaz/libraries/go-keycloak-zerotrust/keycloak/imports/* keycloak/imports/ 2>/dev/null || \
		echo "$(YELLOW)Keycloak imports not found, will use default configuration$(NC)"
	@echo "$(GREEN)✅ Keycloak configuration setup$(NC)"

.PHONY: start
start: create-docker-compose create-dockerfile copy-keycloak-config ## 🚀 Start all services
	@echo "$(BLUE)Starting all services...$(NC)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(YELLOW)⏳ Waiting for services to be ready...$(NC)"
	@sleep 30
	@$(MAKE) --no-print-directory wait-for-services
	@$(MAKE) --no-print-directory services-status

.PHONY: wait-for-services
wait-for-services: ## ⏳ Wait for all services to be ready
	@echo "$(BLUE)Checking service readiness...$(NC)"
	@echo -n "$(YELLOW)⏳ PostgreSQL: $(NC)"
	@timeout 60 bash -c 'until docker exec $(POSTGRES_CONTAINER) pg_isready -U postgres >/dev/null 2>&1; do sleep 2; done' 2>/dev/null && echo "$(GREEN)✅ Ready$(NC)" || echo "$(RED)❌ Timeout$(NC)"
	@echo -n "$(YELLOW)⏳ Redis: $(NC)"
	@timeout 60 bash -c 'until docker exec $(REDIS_CONTAINER) redis-cli ping >/dev/null 2>&1; do sleep 2; done' 2>/dev/null && echo "$(GREEN)✅ Ready$(NC)" || echo "$(RED)❌ Timeout$(NC)"
	@echo -n "$(YELLOW)⏳ Keycloak Admin Console: $(NC)"
	@timeout 120 bash -c 'until curl -sf http://localhost:8082/admin/ >/dev/null 2>&1; do sleep 5; done' 2>/dev/null && echo "$(GREEN)✅ Ready$(NC)" || echo "$(RED)❌ Timeout$(NC)"
	@echo -n "$(YELLOW)⏳ Application: $(NC)"
	@timeout 60 bash -c 'until curl -sf http://localhost:8080/health >/dev/null 2>&1; do sleep 3; done' 2>/dev/null && echo "$(GREEN)✅ Ready$(NC)" || echo "$(RED)❌ Timeout$(NC)"

.PHONY: services-status
services-status: ## 📊 Show service status
	@echo ""
	@echo "$(CYAN)$(BOLD)Service Status:$(NC)"
	@echo "$(CYAN)==============$(NC)"
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "$(CYAN)$(BOLD)Service URLs:$(NC)"
	@echo "$(CYAN)==============$(NC)"
	@echo "🌐 Application:      http://localhost:8080"
	@echo "📊 Application Info: http://localhost:8080/info"
	@echo "🔐 Keycloak Admin:   http://localhost:8082/admin"
	@echo "📈 Keycloak Metrics: http://localhost:8082/metrics"
	@echo "🗄️  PostgreSQL:      localhost:5433"
	@echo "🔴 Redis:           localhost:6380"
	@echo ""
	@echo "$(CYAN)$(BOLD)Default Credentials:$(NC)"
	@echo "$(CYAN)====================$(NC)"
	@echo "👤 Keycloak Admin: admin / admin"
	@echo "🗄️  PostgreSQL: postgres / postgres_password"

.PHONY: stop
stop: ## 🛑 Stop all services
	@echo "$(BLUE)Stopping all services...$(NC)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✅ All services stopped$(NC)"

.PHONY: restart
restart: stop start ## 🔄 Restart all services
	@echo "$(GREEN)✅ All services restarted$(NC)"

.PHONY: rebuild
rebuild: ## 🔨 Rebuild and restart with new configuration
	@echo "$(BLUE)Rebuilding services with updated configuration...$(NC)"
	@$(DOCKER_COMPOSE) down -v
	@$(DOCKER_COMPOSE) build --no-cache
	@$(MAKE) start
	@echo "$(GREEN)✅ Services rebuilt and restarted$(NC)"

.PHONY: fix-ports
fix-ports: ## 🔧 Fix port configuration and restart
	@echo "$(BLUE)Fixing port configuration...$(NC)"
	@echo "$(YELLOW)Current .env configuration:$(NC)"
	@grep -E "(PORT|URL)" .env || echo "No port configuration found"
	@echo ""
	@$(MAKE) rebuild

.PHONY: logs
logs: ## 📜 Show logs from all services
	@$(DOCKER_COMPOSE) logs -f

.PHONY: logs-app
logs-app: ## 📜 Show application logs
	@$(DOCKER_COMPOSE) logs -f app

.PHONY: logs-keycloak
logs-keycloak: ## 📜 Show Keycloak logs
	@$(DOCKER_COMPOSE) logs -f keycloak

# ================================
# Testing
# ================================

.PHONY: test-deps
test-deps: ## 📦 Install test dependencies
	@echo "$(BLUE)Installing test dependencies...$(NC)"
	@$(GOGET) github.com/stretchr/testify/assert@latest
	@$(GOGET) github.com/stretchr/testify/require@latest
	@$(GOMOD) tidy
	@echo "$(GREEN)✅ Test dependencies installed$(NC)"

.PHONY: create-simple-e2e
create-simple-e2e: ## 🧪 Create simple e2e test
	@echo "$(BLUE)Creating simple end-to-end test...$(NC)"
	@mkdir -p test/e2e
	@echo 'package main' > test/e2e/simple_test.go
	@echo '' >> test/e2e/simple_test.go
	@echo 'import (' >> test/e2e/simple_test.go
	@echo '	"io"' >> test/e2e/simple_test.go
	@echo '	"net/http"' >> test/e2e/simple_test.go
	@echo '	"testing"' >> test/e2e/simple_test.go
	@echo '	"time"' >> test/e2e/simple_test.go
	@echo ')' >> test/e2e/simple_test.go
	@echo '' >> test/e2e/simple_test.go
	@echo 'func TestHealthEndpoint(t *testing.T) {' >> test/e2e/simple_test.go
	@echo '	// Wait a bit for services' >> test/e2e/simple_test.go
	@echo '	time.Sleep(5 * time.Second)' >> test/e2e/simple_test.go
	@echo '	' >> test/e2e/simple_test.go
	@echo '	resp, err := http.Get("http://localhost:8080/health")' >> test/e2e/simple_test.go
	@echo '	if err != nil {' >> test/e2e/simple_test.go
	@echo '		t.Fatalf("Failed to get health endpoint: %v", err)' >> test/e2e/simple_test.go
	@echo '	}' >> test/e2e/simple_test.go
	@echo '	defer resp.Body.Close()' >> test/e2e/simple_test.go
	@echo '	' >> test/e2e/simple_test.go
	@echo '	if resp.StatusCode != http.StatusOK {' >> test/e2e/simple_test.go
	@echo '		t.Fatalf("Expected status 200, got %d", resp.StatusCode)' >> test/e2e/simple_test.go
	@echo '	}' >> test/e2e/simple_test.go
	@echo '	' >> test/e2e/simple_test.go
	@echo '	body, err := io.ReadAll(resp.Body)' >> test/e2e/simple_test.go
	@echo '	if err != nil {' >> test/e2e/simple_test.go
	@echo '		t.Fatalf("Failed to read response body: %v", err)' >> test/e2e/simple_test.go
	@echo '	}' >> test/e2e/simple_test.go
	@echo '	' >> test/e2e/simple_test.go
	@echo '	t.Logf("Health endpoint response: %s", string(body))' >> test/e2e/simple_test.go
	@echo '	t.Log("✅ Health endpoint test passed!")' >> test/e2e/simple_test.go
	@echo '}' >> test/e2e/simple_test.go
	@echo "$(GREEN)✅ Simple end-to-end test created$(NC)"

.PHONY: test-unit
test-unit: ## 🧪 Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	@$(GOTEST) -v ./...
	@echo "$(GREEN)✅ Unit tests completed$(NC)"

.PHONY: test-e2e
test-e2e: create-simple-e2e test-deps ## 🎯 Run end-to-end tests
	@echo "$(BLUE)Running end-to-end tests...$(NC)"
	@echo "$(YELLOW)Ensuring services are running...$(NC)"
	@$(MAKE) --no-print-directory services-status
	@echo ""
	@echo "$(CYAN)$(BOLD)🧪 Starting E2E Test Suite$(NC)"
	@echo "$(CYAN)=========================$(NC)"
	@$(GOTEST) -v -timeout $(TEST_TIMEOUT) ./test/e2e/...
	@echo ""
	@echo "$(GREEN)$(BOLD)✅ End-to-end tests completed successfully!$(NC)"

# ================================
# Demo & Manual Testing
# ================================

.PHONY: demo
demo: ## 🎬 Run interactive demo
	@echo "$(CYAN)$(BOLD)🎬 impl-zamaz Zero Trust Demo$(NC)"
	@echo "$(CYAN)==============================$(NC)"
	@echo ""
	@echo "$(YELLOW)Available endpoints:$(NC)"
	@echo "  🌐 Health:     GET  http://localhost:8080/health"
	@echo ""
	@$(MAKE) --no-print-directory demo-interactive

.PHONY: demo-interactive
demo-interactive: ## 🎮 Interactive demo session
	@echo "$(BLUE)Testing endpoints...$(NC)"
	@echo ""
	@echo "$(CYAN)1. Health Check:$(NC)"
	@curl -s http://localhost:8080/health || echo "Service not available"
	@echo ""
	@echo "$(GREEN)$(BOLD)🎉 Demo completed! Basic service is running.$(NC)"

.PHONY: manual-test
manual-test: ## 🧪 Manual testing commands
	@echo "$(CYAN)Manual testing commands:$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Test health endpoint:$(NC)"
	@echo "curl http://localhost:8080/health"

# ================================
# Security Analysis & Code Quality
# ================================

.PHONY: security-install
security-install: ## 🛡️ Install security analysis tools
	@echo "$(BLUE)Installing security analysis tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	@go install github.com/sonatypecommunity/nancy@$(NANCY_VERSION)
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "$(GREEN)✅ Security tools installed$(NC)"

.PHONY: lint
lint: ## 🔍 Run golangci-lint static analysis
	@echo "$(BLUE)Running golangci-lint analysis...$(NC)"
	@which golangci-lint >/dev/null || $(MAKE) security-install
	@golangci-lint run --config .golangci.yml --timeout 5m
	@echo "$(GREEN)✅ Linting completed$(NC)"

.PHONY: security-scan
security-scan: ## 🛡️ Run gosec security scan
	@echo "$(BLUE)Running gosec security scan...$(NC)"
	@which gosec >/dev/null || $(MAKE) security-install
	@gosec -fmt json -out $(SECURITY_REPORT) ./...
	@gosec -fmt text ./...
	@echo "$(GREEN)✅ Security scan completed$(NC)"

.PHONY: vuln-check
vuln-check: ## 🔍 Check for known vulnerabilities
	@echo "$(BLUE)Checking for known vulnerabilities...$(NC)"
	@which govulncheck >/dev/null || $(MAKE) security-install
	@govulncheck ./...
	@echo "$(GREEN)✅ Vulnerability check completed$(NC)"

.PHONY: deps-audit
deps-audit: ## 🔍 Audit dependencies with nancy
	@echo "$(BLUE)Auditing dependencies...$(NC)"
	@which nancy >/dev/null || $(MAKE) security-install
	@go list -json -deps ./... | nancy sleuth
	@echo "$(GREEN)✅ Dependency audit completed$(NC)"

.PHONY: staticcheck
staticcheck: ## 🔍 Run staticcheck analysis
	@echo "$(BLUE)Running staticcheck analysis...$(NC)"
	@which staticcheck >/dev/null || $(MAKE) security-install
	@staticcheck ./...
	@echo "$(GREEN)✅ Staticcheck completed$(NC)"

.PHONY: test-security
test-security: ## 🧪 Run security-focused tests
	@echo "$(BLUE)Running security tests...$(NC)"
	@$(GOTEST) -v -tags=security ./test/security/...
	@echo "$(GREEN)✅ Security tests completed$(NC)"

.PHONY: security-full
security-full: lint security-scan vuln-check deps-audit staticcheck test-security ## 🛡️ Run complete security analysis
	@echo "$(GREEN)$(BOLD)🛡️ Complete security analysis finished!$(NC)"
	@echo ""
	@echo "$(CYAN)Security Analysis Summary:$(NC)"
	@echo "  ✅ Static code analysis (golangci-lint)"
	@echo "  ✅ Security vulnerability scan (gosec)"
	@echo "  ✅ Known vulnerability check (govulncheck)"
	@echo "  ✅ Dependency security audit (nancy)"
	@echo "  ✅ Additional static analysis (staticcheck)"
	@echo "  ✅ Security-focused unit tests"
	@echo ""
	@echo "$(YELLOW)Reports generated:$(NC)"
	@echo "  📄 Security report: $(SECURITY_REPORT)"
	@echo ""
	@echo "$(BLUE)💡 To view detailed security report:$(NC)"
	@echo "  cat $(SECURITY_REPORT) | jq ."

.PHONY: security-ci
security-ci: ## 🤖 Run security checks for CI/CD
	@echo "$(BLUE)Running CI security checks...$(NC)"
	@$(MAKE) lint
	@$(MAKE) security-scan
	@$(MAKE) vuln-check
	@$(MAKE) test-security
	@echo "$(GREEN)✅ CI security checks completed$(NC)"

.PHONY: format
format: ## 🎨 Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@which goimports >/dev/null || go install golang.org/x/tools/cmd/goimports@latest
	@goimports -w .
	@echo "$(GREEN)✅ Code formatting completed$(NC)"

.PHONY: quality-check
quality-check: format lint staticcheck ## 📊 Run code quality checks
	@echo "$(GREEN)✅ Code quality checks completed$(NC)"

# ================================
# Cleanup
# ================================

.PHONY: clean
clean: ## 🧹 Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
	@$(GOCMD) clean -cache -testcache
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

.PHONY: clean-docker
clean-docker: stop ## 🧹 Clean Docker resources
	@echo "$(BLUE)Cleaning Docker resources...$(NC)"
	@$(DOCKER_COMPOSE) down -v --remove-orphans
	@docker system prune -f
	@echo "$(GREEN)✅ Docker cleanup completed$(NC)"

.PHONY: reset
reset: clean-docker clean setup ## 🔄 Complete reset and setup
	@echo "$(GREEN)✅ Complete reset finished$(NC)"

# ================================
# Quick Access Commands
# ================================

.PHONY: dev
dev: start ## 🚀 Start development environment
	@echo "$(GREEN)✅ Development environment ready!$(NC)"
	@echo ""
	@echo "$(YELLOW)Quick commands:$(NC)"
	@echo "  make test-e2e  # Run tests"
	@echo "  make demo      # Interactive demo"
	@echo "  make logs-app  # View app logs"

.PHONY: status
status: services-status ## 📊 Show current status

.PHONY: full-test
full-test: test-unit test-e2e ## 🎯 Run all tests
	@echo "$(GREEN)$(BOLD)✅ All tests completed successfully!$(NC)"

# Default target
.DEFAULT_GOAL := help