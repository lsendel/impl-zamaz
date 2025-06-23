# 🛡️ Zero Trust Authentication Integration Guide

This guide helps you integrate the Go Keycloak Zero Trust library into your project.

**Generated**: Sun Jun 22 12:11:42 EDT 2025  
**Target Framework**: gin  
**Components**: core,middleware  
**Template**: microservice  
**Deployment**: docker

## 📋 Prerequisites

Before you begin, ensure you have:

- **Go 1.21+** installed ([Download Go](https://golang.org/dl/))
- **Docker** and **Docker Compose** installed ([Get Docker](https://docs.docker.com/get-docker/))
- **Git** for repository management
- **curl** for testing API endpoints

### Verify Prerequisites

```bash
# Check Go version
go version

# Check Docker
docker --version
docker-compose --version

# Check other tools
git --version
curl --version
```

## 🚀 Quick Integration

### Step 1: Install Components

Choose your preferred installation method:

#### Method 1: Go Modules (Recommended)
```bash
# Initialize Go module if not already done
go mod init your-project-name

# Install core component
go get github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/components/core@v1.0.0

# Install middleware component
go get github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/components/middleware@v1.0.0

# Install dependencies
go mod tidy
```

#### Method 2: Using Integration Script
```bash
# Download and run integration script
curl -sSL https://raw.githubusercontent.com/lsendel/root-zamaz/main/scripts/integrate.sh | bash -s -- \
    --framework=gin \
    --components=core,middleware \
    --template=microservice

# Or download manually
wget https://raw.githubusercontent.com/lsendel/root-zamaz/main/scripts/integrate.sh
chmod +x integrate.sh
./integrate.sh --framework=gin --components=core,middleware
```

### Step 2: Environment Configuration

Create environment configuration:

```bash
# Create .env file
cat > .env << 'ENVEOF'
# Keycloak Configuration
KEYCLOAK_BASE_URL=http://localhost:8080
KEYCLOAK_REALM=your-realm
KEYCLOAK_CLIENT_ID=your-client
KEYCLOAK_CLIENT_SECRET=your-secret

# Zero Trust Configuration
ZEROTRUST_TRUST_LEVEL_READ=25
ZEROTRUST_TRUST_LEVEL_WRITE=50
ZEROTRUST_TRUST_LEVEL_ADMIN=75
ZEROTRUST_TRUST_LEVEL_DELETE=90

# Cache Configuration
CACHE_TYPE=redis
CACHE_TTL=15m
REDIS_URL=redis://localhost:6379

# Database Configuration (for audit logging)
DATABASE_URL=postgres://user:password@localhost:5432/dbname

# Security Configuration
DEVICE_ATTESTATION_ENABLED=true
RISK_ASSESSMENT_ENABLED=true
CONTINUOUS_VERIFICATION=true
ENVEOF
```

### Step 3: Framework Integration

#### Gin Framework Integration

Create your main application file:

```go
// main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/components/core/zerotrust"
    "github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/components/middleware"
)

func main() {
    // Load configuration from environment
    config, err := zerotrust.LoadConfigFromEnv()
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // Create Zero Trust client
    client, err := zerotrust.NewKeycloakClient(config)
    if err != nil {
        log.Fatal("Failed to create Zero Trust client:", err)
    }
    defer client.Close()

    // Create Gin router
    r := gin.Default()

    // Add Zero Trust middleware
    middleware := zerotrust.NewGinMiddleware(client)

    // Health check endpoint (no authentication required)
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
            "timestamp": time.Now().UTC(),
        })
    })

    // Protected API routes
    api := r.Group("/api")
    api.Use(middleware.Authenticate())
    {
        // Public data (trust level 25+)
        api.GET("/public", middleware.RequireTrustLevel(25), func(c *gin.Context) {
            claims := middleware.GetClaims(c)
            c.JSON(http.StatusOK, gin.H{
                "message": "This is public data",
                "user_id": claims.UserID,
                "trust_level": claims.TrustLevel,
            })
        })

        // Sensitive data (trust level 50+)
        api.GET("/sensitive", middleware.RequireTrustLevel(50), func(c *gin.Context) {
            claims := middleware.GetClaims(c)
            c.JSON(http.StatusOK, gin.H{
                "message": "This is sensitive data",
                "user_id": claims.UserID,
                "trust_level": claims.TrustLevel,
                "device_verified": claims.DeviceVerified,
            })
        })

        // Admin operations (trust level 75+)
        api.POST("/admin", middleware.RequireTrustLevel(75), func(c *gin.Context) {
            claims := middleware.GetClaims(c)
            c.JSON(http.StatusOK, gin.H{
                "message": "Admin operation completed",
                "user_id": claims.UserID,
                "trust_level": claims.TrustLevel,
                "timestamp": time.Now().UTC(),
            })
        })

        // Critical operations (trust level 90+ and device verification)
        api.DELETE("/critical", 
            middleware.RequireTrustLevel(90),
            middleware.RequireDeviceVerification(),
            func(c *gin.Context) {
                claims := middleware.GetClaims(c)
                c.JSON(http.StatusOK, gin.H{
                    "message": "Critical operation completed",
                    "user_id": claims.UserID,
                    "trust_level": claims.TrustLevel,
                    "device_verified": claims.DeviceVerified,
                    "risk_score": claims.RiskScore,
                })
            })
    }

    // Start server with graceful shutdown
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    // Start server in a goroutine
    go func() {
        log.Println("🚀 Server starting on :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()

    // Wait for interrupt signal to gracefully shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("🛑 Server shutting down...")

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("✅ Server exited")
}
```


### Step 4: Testing Your Integration

#### Get a Test Token
```bash
# Get token from Keycloak
TOKEN=$(curl -s -X POST "http://localhost:8080/realms/your-realm/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=password" \
    -d "client_id=your-client" \
    -d "client_secret=your-secret" \
    -d "username=testuser" \
    -d "password=password" | jq -r '.access_token')

echo "Token: $TOKEN"
```

#### Test API Endpoints
```bash
# Test health endpoint (no auth required)
curl http://localhost:8080/health

# Test public endpoint (trust level 25+)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/public

# Test sensitive endpoint (trust level 50+)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/sensitive

# Test admin endpoint (trust level 75+)
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/admin
```

## 🐳 Docker Deployment

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/.env.template .

EXPOSE 8080

CMD ["./main"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - KEYCLOAK_BASE_URL=http://keycloak:8080
      - KEYCLOAK_REALM=your-realm
      - KEYCLOAK_CLIENT_ID=your-client
      - KEYCLOAK_CLIENT_SECRET=your-secret
      - REDIS_URL=redis://redis:6379
      - DATABASE_URL=postgres://postgres:password@postgres:5432/app
    depends_on:
      - keycloak
      - redis
      - postgres
    restart: unless-stopped

  keycloak:
    image: quay.io/keycloak/keycloak:22.0.5
    command: start-dev
    environment:
      - KEYCLOAK_ADMIN=admin
      - KEYCLOAK_ADMIN_PASSWORD=admin
      - KC_DB=postgres
      - KC_DB_URL=jdbc:postgresql://postgres:5432/keycloak
      - KC_DB_USERNAME=postgres
      - KC_DB_PASSWORD=password
    ports:
      - "8081:8080"
    depends_on:
      - postgres
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=postgres
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    restart: unless-stopped

volumes:
  postgres_data:
```

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f app

# Stop services
docker-compose down
```

## 🛠️ Development Workflow

### Makefile (Optional)

Create a Makefile for common tasks:

```makefile
# Makefile
.PHONY: build run test clean docker-build docker-run

# Go build
build:
	go build -o bin/app .

# Run locally
run:
	go run .

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Docker build
docker-build:
	docker build -t zerotrust-app .

# Docker run
docker-run: docker-build
	docker run --rm -p 8080:8080 --env-file .env zerotrust-app

# Development with hot reload (requires air)
dev:
	air

# Install development tools
tools:
	go install github.com/cosmtrek/air@latest
```

### Hot Reload Development

Install Air for hot reload during development:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Create .air.toml config
cat > .air.toml << 'AIREOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
AIREOF

# Start development server
make dev
```

## 🧪 Testing

### Unit Tests

```go
// main_test.go
package main

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    router := gin.Default()
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/health", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "healthy")
}
```

### Integration Tests

```bash
# Run integration tests
go test -tags=integration ./...

# Run with coverage
go test -cover ./...
```

## 📚 Additional Resources

### Documentation
- [Zero Trust Architecture Guide](https://github.com/lsendel/root-zamaz/blob/main/docs/architecture.md)
- [API Reference](https://github.com/lsendel/root-zamaz/blob/main/docs/api-reference.md)
- [Security Best Practices](https://github.com/lsendel/root-zamaz/blob/main/docs/security.md)

### Examples
- [Complete Examples Repository](https://github.com/lsendel/root-zamaz/tree/main/examples)
- [Framework-Specific Examples](https://github.com/lsendel/root-zamaz/tree/main/examples/frameworks)

### Support
- [GitHub Issues](https://github.com/lsendel/root-zamaz/issues)
- [GitHub Discussions](https://github.com/lsendel/root-zamaz/discussions)
- [Component Registry](https://github.com/lsendel/root-zamaz/tree/main/registry)

## 🎯 Next Steps

1. **Customize Configuration**: Adjust trust levels and security policies for your use case
2. **Add Monitoring**: Integrate with your monitoring and observability stack
3. **Scale Deployment**: Configure for production load and high availability
4. **Security Hardening**: Review and implement additional security measures
5. **Team Training**: Familiarize your team with Zero Trust concepts and implementation

---

**🛡️ Congratulations!** You now have Zero Trust authentication integrated into your project. Your application is protected with modern, adaptive security that continuously verifies trust levels and device integrity.

