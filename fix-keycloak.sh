#!/bin/bash

# Fix Keycloak Setup Script
# This script fixes common Keycloak database and configuration issues

echo "🔧 Fixing Keycloak Configuration..."

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

cd /Users/lsendel/IdeaProjects/impl-zamaz

echo -e "${BLUE}1. Stopping services...${NC}"
docker-compose down

echo -e "${BLUE}2. Cleaning up volumes...${NC}"
docker volume rm impl-zamaz_postgres_data 2>/dev/null || true

echo -e "${BLUE}3. Updating Docker Compose for simplified Keycloak...${NC}"

# Create simplified docker-compose override
cat > docker-compose.override.yml << 'EOF'
version: "3.8"

services:
  keycloak:
    image: quay.io/keycloak/keycloak:22.0.5
    container_name: impl-zamaz-keycloak
    command: start-dev
    environment:
      - KEYCLOAK_ADMIN=admin
      - KEYCLOAK_ADMIN_PASSWORD=admin
      - KC_HTTP_PORT=8080
    ports:
      - "8082:8080"
    restart: unless-stopped
    networks:
      - zerotrust-network
    # Remove database dependency for dev mode

  postgres:
    image: postgres:15-alpine
    container_name: impl-zamaz-postgres
    environment:
      - POSTGRES_DB=postgres
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5433:5432"
    restart: unless-stopped
    networks:
      - zerotrust-network
EOF

echo -e "${BLUE}4. Starting services with new configuration...${NC}"
docker-compose up -d

echo -e "${YELLOW}⏳ Waiting for Keycloak to start (this may take 2-3 minutes)...${NC}"
sleep 30

# Wait for Keycloak to be ready
MAX_ATTEMPTS=30
ATTEMPT=1

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    echo -e "${YELLOW}Attempt $ATTEMPT/$MAX_ATTEMPTS: Checking Keycloak...${NC}"
    
    if curl -s http://localhost:8082/admin/ >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Keycloak is ready!${NC}"
        break
    fi
    
    if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
        echo -e "${RED}❌ Keycloak failed to start after $MAX_ATTEMPTS attempts${NC}"
        echo -e "${YELLOW}Checking logs...${NC}"
        docker logs impl-zamaz-keycloak --tail 20
        exit 1
    fi
    
    sleep 10
    ((ATTEMPT++))
done

echo -e "\n${BLUE}5. Testing Keycloak admin access...${NC}"
ADMIN_CHECK=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/admin/)

if [ "$ADMIN_CHECK" = "200" ]; then
    echo -e "${GREEN}✅ Keycloak admin console accessible${NC}"
else
    echo -e "${YELLOW}⚠️  Admin console returned status: $ADMIN_CHECK${NC}"
fi

echo -e "\n${BLUE}6. Creating Zero Trust realm (manual step required)...${NC}"
echo -e "${GREEN}Keycloak is now running! Please follow these steps:${NC}"
echo ""
echo "1. 🌐 Open: http://localhost:8082/admin"
echo "2. 🔑 Login with: admin / admin"
echo "3. 🏛️  Create a new realm called 'zerotrust-test'"
echo "4. 👥 Create a client called 'zerotrust-client'"
echo "5. 🔐 Set client secret to 'zerotrust-secret-12345'"
echo "6. 👤 Create test users"
echo ""
echo -e "${YELLOW}Or use this quick setup script after manual realm creation:${NC}"

# Create realm setup script
cat > setup-realm.sh << 'EOF'
#!/bin/bash
# Quick realm setup script (run after creating realm manually)

echo "Creating client and users via Keycloak API..."

# Get admin token
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8082/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" \
  -d "username=admin" \
  -d "password=admin" | jq -r '.access_token')

echo "Admin token obtained: ${ADMIN_TOKEN:0:20}..."

# Create realm (if it doesn't exist)
curl -s -X POST http://localhost:8082/admin/realms \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "realm": "zerotrust-test",
    "enabled": true,
    "displayName": "Zero Trust Test Realm"
  }' || echo "Realm might already exist"

echo "Realm setup attempted. Please verify in admin console."
EOF

chmod +x setup-realm.sh

echo -e "\n${GREEN}✅ Keycloak fix completed!${NC}"
echo -e "\n${BLUE}Service Status:${NC}"
docker-compose ps

echo -e "\n${CYAN}Next Steps:${NC}"
echo "1. Access Keycloak: http://localhost:8082/admin (admin/admin)"
echo "2. Create the zerotrust-test realm manually"
echo "3. Run: ./setup-realm.sh (optional, for API setup)"
echo "4. Test with: ./test-keycloak.sh"