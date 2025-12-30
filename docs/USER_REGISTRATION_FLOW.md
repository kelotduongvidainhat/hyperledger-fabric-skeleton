# User Registration Flow in Hyperledger Fabric

## 📊 Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     HYPERLEDGER FABRIC USER FLOW                        │
└─────────────────────────────────────────────────────────────────────────┘

┌──────────────┐
│   Step 1:    │
│  CA Admin    │  Admin has credentials to Fabric CA
│  Connects    │  (admin:adminpw)
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────┐
│   Step 2: Register User                                      │
│   ─────────────────────────────────────────────────────────  │
│   fabric-ca-client register                                  │
│     --id.name user1                                          │
│     --id.secret user1pw                                      │
│     --id.type client                                         │
│                                                              │
│   Result: User registered in CA database                     │
└──────┬───────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────┐
│   Step 3: Enroll User (Get Certificates)                    │
│   ─────────────────────────────────────────────────────────  │
│   fabric-ca-client enroll                                    │
│     -u https://user1:user1pw@localhost:7054                  │
│     -M /path/to/msp                                          │
│                                                              │
│   CA generates:                                              │
│   ✓ X.509 Certificate (public key)                          │
│   ✓ Private Key                                              │
│   ✓ CA Certificate                                           │
└──────┬───────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────┐
│   Step 4: Store Credentials (MSP Structure)                 │
│   ─────────────────────────────────────────────────────────  │
│   network/crypto-config/peerOrganizations/org1/users/       │
│   └── user1@org1.example.com/                               │
│       └── msp/                                               │
│           ├── signcerts/cert.pem    ← User's certificate    │
│           ├── keystore/priv_sk      ← Private key           │
│           ├── cacerts/ca.pem        ← CA certificate         │
│           └── config.yaml           ← MSP config             │
└──────┬───────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────┐
│   Step 5: Backend Loads Identity                            │
│   ─────────────────────────────────────────────────────────  │
│   Backend Code (enroll.go):                                 │
│   1. Read cert.pem → Create X509Identity                    │
│   2. Read priv_sk  → Create Signer                          │
│   3. Combine → Gateway Identity                             │
└──────┬───────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────┐
│   Step 6: Submit Transaction                                │
│   ─────────────────────────────────────────────────────────  │
│   User Request:                                              │
│   POST /assets                                               │
│   Headers:                                                   │
│     X-User-ID: user1                                         │
│     X-Org-ID: org1                                           │
│                                                              │
│   Backend:                                                   │
│   1. Load user1@org1 identity from MSP                      │
│   2. Create Gateway connection                              │
│   3. Sign transaction with user's private key               │
│   4. Submit to blockchain                                   │
│                                                              │
│   Blockchain:                                                │
│   ✓ Verifies signature using user's certificate             │
│   ✓ Checks MSP membership                                   │
│   ✓ Executes chaincode                                      │
│   ✓ Returns result                                          │
└──────────────────────────────────────────────────────────────┘
```

---

## 🔑 Key Components

### 1. **Fabric CA (Certificate Authority)**
- **Location**: Docker container `ca_org1` (port 7054)
- **Purpose**: Issues and manages certificates
- **Database**: Stores registered users

### 2. **MSP (Membership Service Provider)**
- **Location**: `network/crypto-config/peerOrganizations/org1/users/`
- **Purpose**: Stores user credentials
- **Structure**:
  ```
  user1@org1.example.com/
  └── msp/
      ├── signcerts/cert.pem    # Public certificate
      ├── keystore/priv_sk      # Private key
      ├── cacerts/ca.pem        # CA certificate
      └── config.yaml           # MSP configuration
  ```

### 3. **Backend Identity Loader**
- **File**: `backend/internal/fabric/enroll.go`
- **Functions**:
  - `GetIdentity()` - Loads user's certificate
  - `GetSigner()` - Loads user's private key
  - Creates Gateway identity for transactions

---

## 🔄 Current System Flow

### **What Exists Now:**

1. **Registered Users** (from `enroll-identities.sh`):
   - `user1@org1.example.com` (registered but not enrolled)
   - `user1@org2.example.com` (registered but not enrolled)

2. **Enrolled Identities**:
   - `Admin@org1.example.com` ✅
   - `Admin@org2.example.com` ✅
   - `peer0` (for each org) ✅

### **What's Missing:**

❌ User identities are **registered** but not **enrolled**
❌ No MSP structure for regular users
❌ Backend can't load user identities

---

## 🛠️ How to Register & Enroll Users

### **Option 1: Manual Registration (Recommended)**

```bash
# 1. Set environment
export FABRIC_CA_CLIENT_HOME=${PWD}/network/crypto-config/peerOrganizations/org1.example.com/

# 2. Register user (if not already done)
fabric-ca-client register \
  --caname ca-org1 \
  --id.name user1 \
  --id.secret user1pw \
  --id.type client \
  --tls.certfiles network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

# 3. Enroll user
mkdir -p network/crypto-config/peerOrganizations/org1.example.com/users/user1@org1.example.com
fabric-ca-client enroll \
  -u https://user1:user1pw@localhost:7054 \
  --caname ca-org1 \
  -M network/crypto-config/peerOrganizations/org1.example.com/users/user1@org1.example.com/msp \
  --tls.certfiles network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

# 4. Copy MSP config
cp network/crypto-config/peerOrganizations/org1.example.com/msp/config.yaml \
   network/crypto-config/peerOrganizations/org1.example.com/users/user1@org1.example.com/msp/config.yaml
```

### **Option 2: Use Script**

Create a simple enrollment script for existing registered users:

```bash
#!/bin/bash
# enroll-user.sh <username> <password> <org>

USERNAME=$1
PASSWORD=$2
ORG=$3  # org1 or org2

if [ "$ORG" = "org1" ]; then
    CA_PORT=7054
else
    CA_PORT=8054
fi

export FABRIC_CA_CLIENT_HOME=${PWD}/network/crypto-config/peerOrganizations/${ORG}.example.com/

fabric-ca-client enroll \
  -u https://${USERNAME}:${PASSWORD}@localhost:${CA_PORT} \
  --caname ca-${ORG} \
  -M network/crypto-config/peerOrganizations/${ORG}.example.com/users/${USERNAME}@${ORG}.example.com/msp \
  --tls.certfiles network/crypto-config/peerOrganizations/${ORG}.example.com/ca/ca.${ORG}.example.com-cert.pem

cp network/crypto-config/peerOrganizations/${ORG}.example.com/msp/config.yaml \
   network/crypto-config/peerOrganizations/${ORG}.example.com/users/${USERNAME}@${ORG}.example.com/msp/config.yaml

echo "✓ ${USERNAME}@${ORG}.example.com enrolled"
```

---

## 📝 Backend Usage

Once users are enrolled, use them in API calls:

```bash
# Create asset as user1
curl -X POST http://localhost:8080/assets \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user1" \
  -H "X-Org-ID: org1" \
  -d '{
    "id": "asset200",
    "color": "red",
    "size": 10,
    "owner": "User1",
    "appraisedValue": 500
  }'
```

**Backend Process:**
1. Receives headers: `X-User-ID: user1`, `X-Org-ID: org1`
2. Looks for: `network/crypto-config/peerOrganizations/org1.example.com/users/user1@org1.example.com/msp/`
3. Loads certificate and private key
4. Creates Gateway identity
5. Signs and submits transaction

---

## ✅ Summary

**Registration Flow:**
1. **Register** → User added to CA database
2. **Enroll** → User gets certificates (MSP structure created)
3. **Backend** → Loads MSP credentials
4. **Transaction** → Signs with user's private key
5. **Blockchain** → Verifies signature and executes

**Current Status:**
- ✅ CA running
- ✅ Users registered (user1)
- ❌ Users not enrolled (need to run enrollment)
- ❌ Backend can't use users yet

**Next Step:**
Enroll user1 to create the MSP structure so backend can use it!
