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
*Note: The script installs, approves, commits, and initializes the ledger. If you see 'Chaincode already installed', it means deployment was already attempted.*

---

## ⚓ Step 3: Update Anchor Peers
Enable cross-org discovery and satisfy endorsement policies.

```bash
docker exec cli ./scripts/update-anchor-peers.sh
```

---

## 🌐 Step 4: Access the Application
The entire stack is now running in Docker. You can access the interfaces immediately:

- **Frontend UI**: [http://localhost:5173](http://localhost:5173) (Asset Explorer)
- **Backend API**: [http://localhost:3000](http://localhost:3000) (REST Entrypoint)
- **IPFS Gateway**: [http://localhost:8080](http://localhost:8080) (Decentralized Images)

### Managing Services
If you make code changes, you can rebuild specific services without a full reset:
```bash
# Rebuild Backend
docker-compose -f network/docker-compose.yaml up -d --build backend

# Rebuild Frontend
docker-compose -f network/docker-compose.yaml up -d --build frontend
```

---

## 🏗️ How it Works: Identity & Policy
This project uses a **Hybrid CA + Policy-as-Code** model:
- **Registration**: The Backend uses `docker exec` to trigger `fabric-ca-client` directly inside the CA containers. This avoids complex TLS certificate mismatches.
- **Persistence**: User identities are stored as standard Fabric Wallet files in `backend/wallet/`. 
- **Authorization (OPA)**: A dedicated **Open Policy Agent** sidecar service manages all authorization logic. The backend delegates "Who can do what" decisions to OPA via **Rego** policies, allowing for dynamic rule updates without code changes.
- **Volumes**: Mounting `/var/run/docker.sock` to the Backend container is required for CA orchestration.

---

## 📊 Step 6: Initialize Test Data (Optional)
Populate the network with test users and assets for both organizations.

```bash
# Ensure backend is running and healthy
./scripts/init-data.sh    # Org1 Setup
./scripts/init-org2.sh    # Org2 Setup
```

---

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
- **CA Connection Refused?** If `docker logs backend` shows "connection refused" to `ca_org1`, wait 5 seconds and `docker-compose -f network/docker-compose.yaml restart backend`.
- **Docker Permission Errors?** Ensure your user is in the `docker` group or run commands with `sudo`.
- **Permission Denied on Cleanup?** The Blockchain volumes are created by `root` inside containers. `fresh-start.sh` requires `sudo` to wipe them.

## 🚀 Next Steps
1. **Explore the Gallery**: Check how assets are displayed with their IPFS CIDs.
2. **Transfer Assets**: Try transferring an asset between Org1 and Org2 users.
3. **Database Sync**: Use the "Sync" button in the Admin Dashboard to reconcile ledger data with PostgreSQL.
