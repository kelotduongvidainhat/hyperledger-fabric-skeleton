# Project Roadmap: Hyperledger Fabric Asset Management

## 📌 Phase 1: Foundation (COMPLETED) ✅
*Focus: Stable Infrastructure & Core Functionality*

- [x] **Network Setup**: One Organization, 3 Raft Orderers, 2 Peers, 2 CAs.
- [x] **Smart Contract**: Asset Transfer (Go) with Lock/Unlock, History, and Timestamping.
- [x] **Backend API**: Go (Gin) using Fabric Gateway SDK.
- [x] **Database Integration**: PostgreSQL off-chain sync via Chaincode Events.
- [x] **Identity Management**: Strict NodeOU support (Admin, Client, Peer, Orderer).
- [x] **Frontend**: Basic React Dashboard (View/Add Assets).
- [x] **DevOps**: `fresh-start.sh` automation for rapid reset and deployment.

---

## 🚀 Phase 2: Enhanced Application Logic (Current Focus)
*Focus: User Experience & Business Rules*

- [ ] **Dynamic User Enrollment**
    - [ ] Backend API to register/enroll new users via Fabric CA.
    - [ ] Frontend "Sign Up" / "Login" page.
    - [ ] Wallet management (store credentials securely in browser/backend).

- [ ] **Access Control (ABAC/RBAC)**
    - [ ] Enforce `"Owner"` check in Chaincode for `UpdateAsset`/`TransferAsset`.
    - [ ] Add `GetAssetsByOwner` query.
    - [ ] Frontend: "My Assets" vs "Marketplace" view.

- [ ] **Asset History Visualization**
    - [ ] Enhance Frontend to display the timeline of asset modifications (provenance).
    - [ ] Show "Locked" status history.

---

## 🌐 Phase 3: Network Expansion
*Focus: Real-world Simulation*

- [ ] **Add Organization 2**
    - [ ] Update `configtx.yaml` & `crypto-config`.
    - [ ] Join Org2 Peers to `mychannel`.
    - [ ] Update Chaincode Endorsement Policy (`AND('Org1.member', 'Org2.member')`).

- [ ] **Private Data Collections (PDC)**
    - [ ] Implement implicit private data for sensitive fields (e.g., Appraisal Value).
    - [ ] Update Chaincode to handle transient data.

---

## 🛠 Phase 4: Production Hardening
*Focus: Security & Scalability*

- [ ] **Kubernetes Migration**
    - [ ] Port Docker Compose setup to K8s manifests (or Fabric Operator).
    - [ ] Setup Ingress for Backend/Fabric interaction.

- [ ] **Monitoring & Logging**
    - [ ] Integrate Prometheus/Grafana for Peer metrics.
    - [ ] Centralized logging (ELK/Loki) for Chaincode/Backend logs.

- [ ] **CI/CD Pipeline**
    - [ ] Automated Chaincode Tests (GitHub Actions).
    - [ ] Linting & Security Scans (`gosec`).
