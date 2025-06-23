# 🔐 Keycloak Access Guide

## 📍 Keycloak Access URLs

### 1. **Admin Console**
```
URL: http://localhost:8082/admin
Username: admin
Password: admin
```

### 2. **Account Console** (for users)
```
URL: http://localhost:8082/realms/zerotrust-test/account
```

### 3. **REST API Endpoints**
```
Token Endpoint: http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/token
User Info: http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/userinfo
JWKS: http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/certs
```

## 🚀 Quick Access Steps

### Step 1: Access Keycloak Admin Console

1. Open your browser and go to: **http://localhost:8082/admin**
2. Login with:
   - Username: `admin`
   - Password: `admin`

### Step 2: Navigate to Zero Trust Realm

Once logged in:
1. Click on the realm dropdown (top-left)
2. Select **zerotrust-test** realm
3. You'll see the realm dashboard

## 🔧 Common Keycloak Tasks

### 1. **Create a Test User**

```bash
# Via Admin Console:
1. Go to Users → Add User
2. Fill in:
   - Username: testuser
   - Email: testuser@example.com
   - Email Verified: ON
3. Click Save
4. Go to Credentials tab
5. Set Password: testpass123
6. Temporary: OFF
7. Click Save
```

### 2. **Get Access Token via API**

```bash
# Get token using curl
curl -X POST http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=zerotrust-client" \
  -d "client_secret=zerotrust-secret-12345" \
  -d "username=testuser" \
  -d "password=testpass123"
```

### 3. **Verify Token**

```bash
# Introspect token
curl -X POST http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/token/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=<your-access-token>" \
  -d "client_id=zerotrust-client" \
  -d "client_secret=zerotrust-secret-12345"
```

## 🔍 Keycloak Integration with Zero Trust

### 1. **Client Configuration**

The Zero Trust client is pre-configured with:
- **Client ID**: `zerotrust-client`
- **Client Secret**: `zerotrust-secret-12345`
- **Access Type**: Confidential
- **Valid Redirect URIs**: `http://localhost:*`

### 2. **Realm Roles**

Pre-configured roles:
- `user` - Basic access (Trust Level 25+)
- `admin` - Administrative access (Trust Level 75+)
- `super-admin` - Full access (Trust Level 90+)

### 3. **User Attributes for Zero Trust**

Custom attributes used for trust scoring:
- `device_id` - Registered device identifier
- `trust_level` - Base trust level
- `location_whitelist` - Allowed locations
- `risk_profile` - User risk assessment

## 📝 Testing Keycloak Integration

### 1. **Test Login Flow**

```javascript
// Using JavaScript/React
const login = async (username, password) => {
  const response = await fetch('http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/token', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      grant_type: 'password',
      client_id: 'zerotrust-client',
      client_secret: 'zerotrust-secret-12345',
      username: username,
      password: password,
    }),
  });

  const data = await response.json();
  console.log('Access Token:', data.access_token);
  console.log('Refresh Token:', data.refresh_token);
  
  return data;
};
```

### 2. **Decode JWT Token**

```javascript
// Decode JWT to see claims
function parseJwt(token) {
  const base64Url = token.split('.')[1];
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
  }).join(''));

  return JSON.parse(jsonPayload);
}

// Example usage
const token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...";
const claims = parseJwt(token);
console.log('User ID:', claims.sub);
console.log('Username:', claims.preferred_username);
console.log('Roles:', claims.realm_access.roles);
```

## 🛠️ Troubleshooting Keycloak Access

### Issue: Cannot access Keycloak at localhost:8082

1. **Check if Keycloak is running:**
   ```bash
   docker ps | grep keycloak
   ```

2. **Check Keycloak logs:**
   ```bash
   docker logs impl-zamaz-keycloak
   ```

3. **Restart Keycloak:**
   ```bash
   docker-compose restart keycloak
   ```

### Issue: Login fails with 401

1. **Verify credentials are correct**
2. **Check client configuration:**
   - Go to Clients → zerotrust-client
   - Verify Client Protocol: openid-connect
   - Check Valid Redirect URIs includes your app URL

3. **Check realm is correct:**
   - Ensure you're using `zerotrust-test` realm
   - Not the `master` realm

### Issue: Token validation fails

1. **Check token expiration:**
   ```bash
   # Tokens expire after 5 minutes by default
   ```

2. **Verify client secret:**
   - Clients → zerotrust-client → Credentials
   - Regenerate secret if needed

## 🔐 Security Best Practices

1. **Change Default Passwords**
   - Change admin password in production
   - Use strong client secrets

2. **Enable HTTPS**
   - Use SSL certificates in production
   - Update URLs to https://

3. **Configure CORS**
   - Set allowed origins properly
   - Don't use * in production

4. **Token Lifetime**
   - Adjust token expiration times
   - Use refresh tokens appropriately

## 📊 Keycloak Metrics & Monitoring

Access Keycloak metrics at:
- **Metrics**: http://localhost:8082/metrics
- **Health**: http://localhost:8082/health

Monitor important metrics:
- Active sessions
- Failed login attempts
- Token issuance rate
- Response times

## 🚀 Quick Demo Script

```bash
#!/bin/bash

echo "🔐 Keycloak Zero Trust Demo"
echo "=========================="

# 1. Check Keycloak health
echo "1. Checking Keycloak health..."
curl -s http://localhost:8082/health/ready || echo "Keycloak not ready"

# 2. Get access token
echo -e "\n2. Getting access token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=zerotrust-client" \
  -d "client_secret=zerotrust-secret-12345" \
  -d "username=admin" \
  -d "password=admin")

ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')

if [ "$ACCESS_TOKEN" != "null" ]; then
  echo "✅ Successfully obtained access token!"
  echo "Token (first 50 chars): ${ACCESS_TOKEN:0:50}..."
  
  # 3. Get user info
  echo -e "\n3. Getting user info..."
  curl -s http://localhost:8082/realms/zerotrust-test/protocol/openid-connect/userinfo \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
else
  echo "❌ Failed to get access token"
  echo "Response: $TOKEN_RESPONSE"
fi
```

Run this script to verify Keycloak is working properly! Alternatively, use `make test-ports` to test all service connectivity.