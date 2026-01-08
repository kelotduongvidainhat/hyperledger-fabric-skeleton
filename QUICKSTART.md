# 🚀 Project Quickstart Guide

This guide will help you get the Hyperledger Fabric network and application up and running from scratch.

## 📋 Prerequisites
Ensure you have the following installed:
- **Docker** and **Docker Compose**
- **Go 1.25+**
- **Node.js** (for Frontend)
- **jq** (for script processing)

---

## 🛠️ Step 1: Network & Application "Fresh Start"
Reset and bootstrap the entire stack: Blockchain, Database, IPFS, **Backend**, and **Frontend**.

```bash
# Run with sudo to ensure volume cleanup
sudo ./scripts/fresh-start.sh
```
*This script regenerates crypto materials, bootstraps all 12+ containers, and builds the latest application code.*

---

## 💿 Step 2: Deploy Chaincode (CaaS)
Deploy the Smart Contract using **Chaincode-as-a-Service**.

1. **Wait for network stabilization** (about 10 seconds).
2. **Execute deployment script**:
```bash
# Copy package to CLI
docker cp chaincode/packaging/basic.tar.gz cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/

# Run deployment
docker exec cli ./scripts/deploy-caas.sh
```

---

## ⚓ Step 3: Update Anchor Peers
Enable cross-org discovery and satisfy endorsement policies.

```bash
docker exec cli ./scripts/update-anchor-peers.sh
```

---

## 🌐 Step 4: Access the Application
The entire stack is now running in Docker. You can access the interfaces immediately:

- **Frontend UI**: [http://localhost:5173](http://localhost:5173)
- **Backend API**: [http://localhost:3000](http://localhost:3000)
- **IPFS Gateway**: [http://localhost:8080](http://localhost:8080)

### Managing Services
If you make code changes, you can rebuild specific services without a full reset:
```bash
# Rebuild Backend
docker-compose -f network/docker-compose.yaml up -d --build backend

# Rebuild Frontend
docker-compose -f network/docker-compose.yaml up -d --build frontend
```

---

---

## 📊 Step 6: Initialize Test Data (Optional)
Populate the network with test users and assets for both organizations.

```bash
# Ensure backend is running in another terminal
# Initialize Org1 (10 users)
./scripts/init-data.sh

# Initialize Org2 (1 Admin + 5 users)
./scripts/init-org2.sh
```
*These scripts automate registration, admin approval, and asset creation.*

---

## 🧪 Testing the Setup
You can verify everything is working using `curl` or the provided test scripts:

**1. Register a User (Multi-Org Support)**:
```bash
# Register for Org1 (default) or specify "Org2MSP"
curl -X POST "http://localhost:3000/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username": "demo_user", "password": "password123", "org": "Org2MSP"}'
```

**2. Login**:
The system will automatically find your identity across all configured CAs.
```bash
curl -X POST "http://localhost:3000/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "demo_user", "password": "password123"}'
```

**3. Create an Asset**:
The asset will be record with your full identity: `Org2MSP::demo_user`.
```bash
# Get the token from login and use it:
curl -X POST "http://localhost:3000/assets" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"id": "qs_001", "name": "Quickstart Asset", "desc": "Born from guide"}'
```

---

## 🔍 Troubleshooting
- **Endorsement Errors?** Ensure Step 3 (Anchor Peers) was completed.
- **CA Connection Refused?** Check if `ca_org1` is healthy (`docker ps`) and that you are using `https` in `backend/main.go`.
- **Permission Denied on Cleanup?** Always use `sudo` for `fresh-start.sh` if volumes were created by root.
