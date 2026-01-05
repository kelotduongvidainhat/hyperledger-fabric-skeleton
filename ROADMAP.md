# Project Roadmap: Hyperledger Fabric Ownership Registry

This roadmap outlines the future development and milestones for the Ownership Registry application, building upon the current MVP foundation.

## Phase 1: MVP Foundation (Completed ✅)
- [x] **Network Infrastructure**: 2 Organizations, 1 Orderer (OSNAdmin managed), CaaS Chaincode.
- [x] **Smart Contract**: Go Chaincode with `CreateAsset`, `ReadAsset`, `ProposeTransfer`, `AcceptTransfer`, `GetAssetHistory`, `GetAllAssets`.
- [x] **Backend API**: Go/Fiber REST API integrated with Fabric Gateway SDK.
- [x] **Frontend UI**: React + Vite + Tailwind CSS ("Parchment & Ink" design) with Dashboard and Transfer workflows.

---

## Phase 2: Role-Based Access & UI (Short Term)
- [ ] **Role Management**:
    - **Admin**: Full dashboard stats, User Management (Approvals/Ban), System Configuration.
    - **Auditor**: Read-only access to full history logs, compliance reporting view.
    - **User**: Standard Asset Management (Create, Transfer, View Own Assets).
- [ ] **Real Authentication**: Replace dummy "Logged in via Gateway" with real JWT-based authentication.
    - Integrate `fabric-ca-client` to register/enroll users via API.
    - Implement Login endpoint returning JWT with embedded roles.
- [ ] **Object Storage (MinIO)**:
    - Deploy MinIO (S3-compatible) container.
    - Backend: Implement file upload endpoint (`POST /upload`) to push images to MinIO.
- [ ] **UI Enhancements**:
    - **Admin Panel**: Dedicated route `/admin` for user oversight.
    - **Auditor View**: Advanced filtering and report generation.
    - **User Wallet**: UI for managing personal keys/identities.

## Phase 3: Advanced Features (Medium Term)
- [ ] **IPFS Integration**:
    - Transition from centralized MinIO to decentralized IPFS for robust, censorship-resistant storage.
    - Store the IPFS Content ID (CID) on-chain.
- [ ] **Private Data Collections (PDC)**:
    - Update Chaincode to support private pricing or metadata that is shared only between Buyer and Seller, not the whole network.
- [ ] **Complex Transfer Logic**:
    - Implement "Offer/Bid" mechanism with price negotiation on-chain.
    - Add "Escrow" functionality.

## Phase 4: Production Readiness (Long Term)
- [ ] **Kubernetes Deployment**:
    - Migrate from `docker-compose` to K8s (using Fabric Operator).
- [ ] **Monitoring & Logging**:
    - Integrate Prometheus & Grafana for network metrics.
    - ELK Stack for chaincode and peer logs.
- [ ] **High Availability**:
    - Add multiple Orderers (Raft consensus verification).
    - Add multiple Peers per Organization with Gossip properly configured.

## Phase 5: Expansion
- [ ] **Multi-Channel Support**: create separate channels for different asset classes.
- [ ] **Cross-Chain Interoperability**: Explore Hyperledger Cactus/Centi for bridging with Ethereum.
