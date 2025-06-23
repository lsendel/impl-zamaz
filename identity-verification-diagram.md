# Identity Verification Workflow - Visual Summary

## 🔐 Complete Zero Trust Identity Verification Flow

```
┌──────────────────────────────────────────────────────────────────────┐
│                           USER REQUEST                                │
│                    "I want to access /api/data"                      │
└────────────────────────────┬─────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    1. TOKEN EXTRACTION                                │
│            Extract JWT from Authorization header                      │
│                 "Bearer eyJhbGciOiJSUzI1..."                         │
└────────────────────────────┬─────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    2. TOKEN VALIDATION                                │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────────┐        │
│  │  Check Cache │───▶│ Not in Cache │───▶│ Validate with   │        │
│  │  for Token  │    │              │    │   Keycloak      │        │
│  └─────────────┘    └──────────────┘    └─────────────────┘        │
│         │                                         │                   │
│         │ Found                                   ▼                   │
│         │                               ┌─────────────────┐          │
│         └──────────────────────────────▶│  Extract Claims │          │
│                                         │  (user, roles)  │          │
│                                         └─────────────────┘          │
└────────────────────────────┬─────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────┐
│                 3. ZERO TRUST VERIFICATION                            │
│                                                                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────────┐    │
│  │ Device Check    │  │ Location Check  │  │ Behavior Check   │    │
│  │ Score: 20/25    │  │ Score: 12/15    │  │ Score: 18/20     │    │
│  └─────────────────┘  └─────────────────┘  └──────────────────┘    │
│           │                    │                     │               │
│           └────────────────────┴─────────────────────┘               │
│                                │                                      │
│                                ▼                                      │
│                   ┌─────────────────────────┐                        │
│                   │  TRUST SCORE: 88/100    │                        │
│                   └─────────────────────────┘                        │
└────────────────────────────┬─────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    4. AUTHORIZATION DECISION                          │
│                                                                       │
│         Required Trust Level: 50 (for write operations)              │
│         Current Trust Score: 88                                       │
│                                                                       │
│                    ✅ ACCESS GRANTED                                  │
└────────────────────────────┬─────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    5. CONTINUOUS MONITORING                           │
│                                                                       │
│  • Log access attempt                                                 │
│  • Update behavioral profile                                          │
│  • Schedule next verification (5 min)                                 │
│  • Monitor for anomalies                                              │
└──────────────────────────────────────────────────────────────────────┘
```

## 📊 Trust Score Breakdown

```
Total Trust Score: 88/100
├── Identity (30/30) ────── Valid JWT token from Keycloak
├── Device (20/25) ──────── Known device, partial attestation  
├── Behavior (18/20) ────── Normal usage patterns detected
├── Location (12/15) ────── Trusted network location
└── Risk (8/10) ─────────── Low risk, no anomalies detected
```

## 🔒 Key Security Checks at Each Stage

### Stage 1: Token Extraction
- ✓ Proper Authorization header format
- ✓ Bearer token present
- ✓ Token not blacklisted

### Stage 2: Token Validation  
- ✓ Valid JWT structure (header.payload.signature)
- ✓ Signature verified with Keycloak public key
- ✓ Token not expired (exp claim)
- ✓ Correct issuer (iss: Keycloak URL)
- ✓ Correct audience (aud: client ID)

### Stage 3: Zero Trust Verification
- ✓ Device attestation (hardware integrity)
- ✓ Geolocation validation (trusted locations)
- ✓ Behavioral analysis (usage patterns)
- ✓ Risk assessment (threat detection)
- ✓ Trust decay calculation (time-based)

### Stage 4: Authorization
- ✓ Trust score meets threshold
- ✓ User has required roles
- ✓ Resource permissions validated
- ✓ Time-based access restrictions

### Stage 5: Continuous Monitoring
- ✓ Session activity tracking
- ✓ Anomaly detection
- ✓ Trust score updates
- ✓ Automatic session revocation if needed

## 🚀 Implementation Code Path

```
User Request
    ↓
GinMiddleware.Authenticate()           // middleware/gin/gin_middleware.go
    ↓
KeycloakClient.ValidateToken()         // pkg/client/keycloak_client.go
    ↓
TrustEngine.CalculateTrustScore()      // pkg/zerotrust/trust_engine.go
    ↓
Middleware.RequireTrustLevel()         // middleware/gin/gin_middleware.go
    ↓
Application Handler                     // Your business logic
```

This workflow ensures that **every request** goes through comprehensive verification, embodying the Zero Trust principle of "Never trust, always verify"!