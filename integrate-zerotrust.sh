#!/bin/bash

# ==================================================
# Zero Trust Integration Script
# ==================================================
# Automatically integrates Zero Trust components into your project

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_step() {
    echo -e "${GREEN}🔸 $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo -e "\n${BLUE}🛡️ Zero Trust Integration Script${NC}\n"

# Check prerequisites
print_step "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21+"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker"
    exit 1
fi

print_success "Prerequisites check passed"

# Initialize Go module if needed
print_step "Initializing Go module..."
if [ ! -f "go.mod" ]; then
    print_info "No go.mod found. Initializing..."
    read -p "Enter module name (e.g., github.com/yourorg/yourproject): " MODULE_NAME
    go mod init "$MODULE_NAME"
    print_success "Go module initialized"
else
    print_info "Using existing go.mod"
fi

# Install Zero Trust components
print_step "Installing Zero Trust components..."

print_info "Installing core component..."
go get github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/components/core@v1.0.0

print_info "Installing middleware component..."
go get github.com/lsendel/root-zamaz/libraries/go-keycloak-zerotrust/components/middleware@v1.0.0

print_info "Installing Gin framework..."
go get github.com/gin-gonic/gin@latest

# Install common dependencies
go get github.com/golang-jwt/jwt/v5@latest
go get golang.org/x/crypto@latest

print_success "Components installed successfully"

# Tidy dependencies
print_step "Tidying Go modules..."
go mod tidy
print_success "Dependencies resolved"

# Create .env template if it doesn't exist
print_step "Creating environment configuration..."
if [ ! -f ".env" ]; then
    cat > .env.template << 'ENVEOF'
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

# Database Configuration
DATABASE_URL=postgres://user:password@localhost:5432/dbname

# Security Configuration
DEVICE_ATTESTATION_ENABLED=true
RISK_ASSESSMENT_ENABLED=true
CONTINUOUS_VERIFICATION=true
ENVEOF

    cp .env.template .env
    print_success "Environment configuration created (.env.template and .env)"
    print_info "Please update .env with your actual configuration values"
else
    print_info "Using existing .env file"
fi

# Create .gitignore if it doesn't exist
print_step "Creating .gitignore..."
if [ ! -f ".gitignore" ]; then
    cat > .gitignore << 'GITIGNOREEOF'
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# Environment files
.env
.env.local
.env.*.local

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Build artifacts
/bin/
/dist/
/tmp/
GITIGNOREEOF

    print_success ".gitignore created"
else
    print_info "Using existing .gitignore"
fi

print_success "Zero Trust integration completed!"

echo -e "\n${YELLOW}Next Steps:${NC}"
echo "1. Update .env with your Keycloak configuration"
echo "2. Review the generated ZEROTRUST_INTEGRATION.md guide"
echo "3. Run your application: go run ."
echo "4. Test the integration with the provided examples"

echo -e "\n${GREEN}🎉 Your project is now protected with Zero Trust authentication!${NC}"
