#!/bin/bash

echo "🔐 Keycloak Zero Trust Demo"
echo "=========================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 1. Check Keycloak health
echo -e "\n${BLUE}1. Checking Keycloak health...${NC}"
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/health/ready)

if [ "$HEALTH_CHECK" = "200" ]; then
  echo -e "${GREEN}✅ Keycloak is healthy${NC}"
else
  echo -e "${RED}❌ Keycloak health check failed (Status: $HEALTH_CHECK)${NC}"
  echo -e "${YELLOW}Make sure Keycloak is running on port 8082${NC}"
  exit 1
fi

# 2. Get access token
echo -e "\n${BLUE}2. Getting access token...${NC}"
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=zerotrust-client" \
  -d "client_secret=zerotrust-secret-12345" \
  -d "username=admin" \
  -d "password=admin" 2>/dev/null)

# Check if we got a token
if echo "$TOKEN_RESPONSE" | grep -q "access_token"; then
  ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token' 2>/dev/null || echo "")
  REFRESH_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.refresh_token' 2>/dev/null || echo "")
  
  if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
    echo -e "${GREEN}✅ Successfully obtained access token!${NC}"
    echo "Token (first 50 chars): ${ACCESS_TOKEN:0:50}..."
    
    # 3. Decode token to show claims
    echo -e "\n${BLUE}3. Decoding JWT token claims...${NC}"
    # Extract payload from JWT
    PAYLOAD=$(echo $ACCESS_TOKEN | cut -d'.' -f2)
    # Add padding if needed
    PAYLOAD_LENGTH=$(echo -n $PAYLOAD | wc -c)
    PADDING=$((4 - PAYLOAD_LENGTH % 4))
    if [ $PADDING -ne 4 ]; then
      PAYLOAD="${PAYLOAD}$(printf '=%.0s' $(seq 1 $PADDING))"
    fi
    
    # Decode base64url
    DECODED=$(echo $PAYLOAD | tr '_-' '/+' | base64 -d 2>/dev/null)
    
    if [ -n "$DECODED" ]; then
      echo "Token Claims:"
      echo "$DECODED" | jq . 2>/dev/null || echo "$DECODED"
    fi
    
    # 4. Get user info
    echo -e "\n${BLUE}4. Getting user info from userinfo endpoint...${NC}"
    USER_INFO=$(curl -s http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/userinfo \
      -H "Authorization: Bearer $ACCESS_TOKEN")
    
    if echo "$USER_INFO" | grep -q "sub"; then
      echo -e "${GREEN}✅ User info retrieved:${NC}"
      echo "$USER_INFO" | jq . 2>/dev/null || echo "$USER_INFO"
    else
      echo -e "${RED}❌ Failed to get user info${NC}"
    fi
    
    # 5. Introspect token
    echo -e "\n${BLUE}5. Introspecting token...${NC}"
    INTROSPECT=$(curl -s -X POST http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/token/introspect \
      -H "Content-Type: application/x-www-form-urlencoded" \
      -d "token=$ACCESS_TOKEN" \
      -d "client_id=zerotrust-client" \
      -d "client_secret=zerotrust-secret-12345")
    
    if echo "$INTROSPECT" | grep -q "active"; then
      echo -e "${GREEN}✅ Token introspection result:${NC}"
      echo "$INTROSPECT" | jq . 2>/dev/null || echo "$INTROSPECT"
    fi
    
  else
    echo -e "${RED}❌ Failed to extract access token${NC}"
  fi
else
  echo -e "${RED}❌ Failed to get access token${NC}"
  echo "Response: $TOKEN_RESPONSE"
  echo -e "\n${YELLOW}Possible issues:${NC}"
  echo "1. Keycloak might not be fully started yet"
  echo "2. The realm 'zerotrust-test' might not be imported"
  echo "3. The client credentials might be incorrect"
fi

echo -e "\n${BLUE}===============================================${NC}"
echo -e "${GREEN}Keycloak Admin Console:${NC} http://localhost:8082/admin"
echo -e "${GREEN}Username:${NC} admin"
echo -e "${GREEN}Password:${NC} admin"
echo -e "${BLUE}===============================================${NC}"