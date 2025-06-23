#!/bin/bash

# Fix configuration and restart impl-zamaz services
# This script fixes port mismatches and rebuilds services

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "\n${BLUE}🔧 impl-zamaz Configuration Fix${NC}\n"

# Navigate to project directory
cd /Users/lsendel/IdeaProjects/impl-zamaz

echo -e "${YELLOW}Current .env configuration:${NC}"
grep -E "(PORT|URL)" .env || echo "No port configuration found"
echo ""

echo -e "${BLUE}Stopping existing services...${NC}"
docker-compose down -v

echo -e "${BLUE}Removing old containers and volumes...${NC}"
docker-compose rm -f || true
docker volume prune -f || true

echo -e "${BLUE}Building fresh containers...${NC}"
docker-compose build --no-cache

echo -e "${BLUE}Starting services with corrected configuration...${NC}"
docker-compose up -d

echo -e "${YELLOW}Waiting for services to start...${NC}"
sleep 45

echo -e "${BLUE}Checking service health...${NC}"

# Check PostgreSQL
echo -n "PostgreSQL: "
if docker exec impl-zamaz-postgres pg_isready -U postgres >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Ready${NC}"
else
    echo -e "${RED}❌ Not Ready${NC}"
fi

# Check Redis
echo -n "Redis: "
if docker exec impl-zamaz-redis redis-cli ping >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Ready${NC}"
else
    echo -e "${RED}❌ Not Ready${NC}"
fi

# Check Keycloak
echo -n "Keycloak: "
if curl -sf http://localhost:8082/admin/ >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Ready${NC}"
else
    echo -e "${RED}❌ Not Ready${NC}"
fi

# Check Application
echo -n "Application: "
if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Ready${NC}"
    echo ""
    echo -e "${GREEN}Application health check:${NC}"
    curl -s http://localhost:8080/health | jq . || curl -s http://localhost:8080/health
else
    echo -e "${RED}❌ Not Ready${NC}"
fi

echo ""
echo -e "${BLUE}Service URLs:${NC}"
echo "🌐 Application:      http://localhost:8080"
echo "📊 Application Info: http://localhost:8080/info"
echo "🔐 Keycloak Admin:   http://localhost:8082/admin"
echo "📈 Keycloak Health:  http://localhost:8082/health"
echo "🗄️  PostgreSQL:      localhost:5433"
echo "🔴 Redis:           localhost:6380"

echo ""
echo -e "${GREEN}✅ Configuration fix completed!${NC}"