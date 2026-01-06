# 🚀 Project Quickstart Guide

This guide will help you get the Hyperledger Fabric network and application up and running from scratch.

## 📋 Prerequisites
Ensure you have the following installed:
- **Docker** and **Docker Compose**
- **Go 1.25+**
- **Node.js** (for Frontend)
- **jq** (for script processing)

---

## 🛠️ Step 1: Network "Fresh Start"
Reset and bootstrap the base Hyperledger Fabric network (Orgs, Peers, Orderer) and the **PostgreSQL** off-chain database.

```bash
# Run with sudo to ensure volume cleanup
sudo ./scripts/fresh-start.sh
```
*This script bootstraps core network containers and initializes the PostgreSQL storage at `localhost:5432`.*

---

## 💿 Step 2: Deploy Chaincode (CaaS)
Deploy the Smart Contract using **Chaincode-as-a-Service**.

1. **Wait for network stabilization** (about 10 seconds).
2. **Execute deployment inside CLI**:
```bash
# Copy the package to CLI
docker cp basic.tar.gz cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/

# Run deployment script
docker exec cli ./scripts/deploy-caas.sh
```

> **Note:** If the `CHAINCODE_ID` changes (due to code edits), update the `network/.env` file with the new `Package ID` and restart the chaincode container:
> ```bash
> docker-compose -f network/docker-compose.yaml up -d --no-deps chaincode-basic
> ```

---

## ⚓ Step 3: Update Anchor Peers
Critical for cross-org discovery and satisfying endorsement policies.

```bash
docker exec cli ./scripts/update-anchor-peers.sh
```

---

## 💻 Step 4: Run the Backend
The backend provides the REST API and interacts with the Fabric CA (now with TLS enabled).

```bash
cd backend
# Install dependencies
go mod download
# Run the server
go run main.go
```
*The server will run on `http://localhost:3000`.*

---

## 🌐 Step 5: Run the Frontend (Optional)
```bash
cd frontend
npm install
npm run dev
```
*The app will be available at `http://localhost:5173`.*

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
