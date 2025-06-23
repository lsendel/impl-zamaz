#!/bin/bash

# Demo: Zero Trust Identity Verification Workflow
# This script demonstrates the complete identity verification process

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

print_header() {
    echo -e "\n${CYAN}${BOLD}=== $1 ===${NC}\n"
}

print_step() {
    echo -e "${GREEN}[Step $1]${NC} $2"
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

# Configuration
KEYCLOAK_URL="http://localhost:8082"
KEYCLOAK_REALM="zerotrust-test"
CLIENT_ID="zerotrust-client"
CLIENT_SECRET="zerotrust-secret-12345"
APP_URL="http://localhost:8080"

clear
print_header "Zero Trust Identity Verification Workflow Demo"

# Step 1: Show initial request without authentication
print_step "1" "Attempting to access protected resource without authentication"
echo -e "${BLUE}Request:${NC} GET $APP_URL/api/public"
echo -e "${BLUE}Headers:${NC} None\n"

RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" $APP_URL/api/public)
HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed -n '1,/HTTP_STATUS/p' | sed '$d')

if [ "$HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}Response (200 OK):${NC}"
    echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
    print_info "This endpoint is currently public (no auth required)"
else
    echo -e "${RED}Response ($HTTP_STATUS):${NC}"
    echo "$BODY"
fi

# Step 2: Authenticate with Keycloak
print_step "2" "Authenticating with Keycloak to obtain JWT token"
echo -e "${BLUE}Request:${NC} POST $KEYCLOAK_URL/realms/$KEYCLOAK_REALM/protocol/openid-connect/token"
echo -e "${BLUE}Payload:${NC}"
echo "  - grant_type: password"
echo "  - client_id: $CLIENT_ID"
echo "  - username: admin"
echo "  - password: admin"
echo ""

# Note: In a real scenario, we would authenticate with Keycloak
# For this demo, we'll simulate the process
print_info "In a production environment, this would return:"
cat << 'EOF'
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJGSjg...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI...",
  "token_type": "Bearer",
  "not-before-policy": 0,
  "session_state": "5f7c4f88-0c94-4252-a52d-26f1b6a15b01",
  "scope": "email profile"
}
EOF

# Step 3: Show JWT token structure
print_step "3" "JWT Token Structure (Decoded)"
echo -e "${BLUE}Header:${NC}"
cat << 'EOF'
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "FJ86GcF3jTbNLOco4NvZkUCIUmfYCqoqtOQeMfbhNlE"
}
EOF

echo -e "\n${BLUE}Payload:${NC}"
cat << 'EOF'
{
  "exp": 1624300800,
  "iat": 1624300500,
  "jti": "5c5f8f89-df3e-4c3b-9a91-4b5fd7f8e3a9",
  "iss": "http://localhost:8082/realms/zerotrust-test",
  "aud": "zerotrust-client",
  "sub": "f:550e8400-e29b-41d4-a716-446655440000:testuser",
  "typ": "Bearer",
  "azp": "zerotrust-client",
  "session_state": "5f7c4f88-0c94-4252-a52d-26f1b6a15b01",
  "preferred_username": "testuser",
  "email": "testuser@example.com",
  "email_verified": true,
  "realm_access": {
    "roles": ["user", "admin"]
  }
}
EOF

echo -e "\n${BLUE}Signature:${NC} [Verified with Keycloak's public key]"

# Step 4: Show Zero Trust verification process
print_step "4" "Zero Trust Verification Process"
echo -e "${PURPLE}The middleware performs these checks:${NC}\n"

echo "1️⃣  ${BOLD}Token Validation${NC}"
echo "   ✓ Verify JWT signature with Keycloak's public key"
echo "   ✓ Check token expiration (exp claim)"
echo "   ✓ Validate issuer (iss) matches Keycloak URL"
echo "   ✓ Verify audience (aud) matches client ID"
echo ""

echo "2️⃣  ${BOLD}Trust Score Calculation${NC}"
echo "   📊 Identity Factor: 30/30 (Valid authentication)"
echo "   📱 Device Factor: 20/25 (Known device, not attested)"
echo "   🔍 Behavior Factor: 18/20 (Normal usage pattern)"
echo "   🌍 Location Factor: 12/15 (Trusted location)"
echo "   ⚠️  Risk Factor: 8/10 (Low risk)"
echo "   ${BOLD}Total Trust Score: 88/100${NC}"
echo ""

echo "3️⃣  ${BOLD}Authorization Decision${NC}"
echo "   Required trust level for endpoint: 25"
echo "   Current trust score: 88"
echo "   ${GREEN}✅ Access GRANTED${NC}"

# Step 5: Show continuous verification
print_step "5" "Continuous Verification (Background Process)"
echo -e "${PURPLE}Every 5 minutes, the system:${NC}\n"
echo "• Re-evaluates trust score"
echo "• Checks for anomalous behavior"
echo "• Monitors session activity"
echo "• Updates device attestation status"
echo "• Applies trust decay over time"
echo ""
print_info "If trust score drops below threshold, session is revoked"

# Step 6: Show trust level requirements
print_step "6" "Trust Level Requirements for Different Operations"
echo ""
echo "┌─────────────────────┬──────────────┬────────────────────────┐"
echo "│ Operation           │ Trust Level  │ Description            │"
echo "├─────────────────────┼──────────────┼────────────────────────┤"
echo "│ Read Public Data    │ 25+          │ Basic authentication   │"
echo "│ Write User Data     │ 50+          │ Verified device        │"
echo "│ Admin Operations    │ 75+          │ Full attestation       │"
echo "│ Delete Operations   │ 90+          │ Multi-factor + recent  │"
echo "└─────────────────────┴──────────────┴────────────────────────┘"

# Step 7: Show audit trail
print_step "7" "Audit Trail Example"
echo -e "${PURPLE}Every access attempt is logged:${NC}\n"
cat << 'EOF'
{
  "timestamp": "2025-06-22T23:15:32.123Z",
  "user_id": "f:550e8400-e29b-41d4-a716-446655440000:testuser",
  "action": "ACCESS_GRANTED",
  "endpoint": "/api/public",
  "trust_score": 88,
  "factors": {
    "identity": 30,
    "device": 20,
    "behavior": 18,
    "location": 12,
    "risk": 8
  },
  "client_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "session_id": "5f7c4f88-0c94-4252-a52d-26f1b6a15b01"
}
EOF

# Summary
print_header "Summary: Zero Trust Identity Verification"

echo "The Zero Trust model ensures:"
echo ""
echo "🔐 ${BOLD}No Implicit Trust${NC}"
echo "   Every request is verified, regardless of source"
echo ""
echo "🔄 ${BOLD}Continuous Verification${NC}"
echo "   Trust is constantly re-evaluated, not just at login"
echo ""
echo "📊 ${BOLD}Dynamic Trust Levels${NC}"
echo "   Access rights adjust based on real-time risk assessment"
echo ""
echo "🛡️ ${BOLD}Defense in Depth${NC}"
echo "   Multiple verification factors prevent single point of failure"
echo ""
echo "📝 ${BOLD}Complete Audit Trail${NC}"
echo "   Every action is logged for compliance and security analysis"

print_success "Identity verification workflow demonstration complete!"

echo -e "\n${YELLOW}Try it yourself:${NC}"
echo "1. Set up Keycloak with a test user"
echo "2. Obtain a real JWT token"
echo "3. Make authenticated requests to see trust scores"
echo "4. Observe how trust levels affect access permissions"