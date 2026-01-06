# Project Roadmap: Hyperledger Fabric Ownership Registry

This roadmap outlines the future development and milestones for the Ownership Registry application, building upon the current MVP foundation.

## Phase 1: MVP Foundation (Completed ✅)
- [x] **Network Infrastructure**: 2 Organizations, 1 Orderer (OSNAdmin managed), CaaS Chaincode.
- [x] **Smart Contract**: Go Chaincode with `CreateAsset`, `ReadAsset`, `ProposeTransfer`, `AcceptTransfer`, `GetAssetHistory`, `GetAllAssets`.
- [x] **Backend API**: Go/Fiber REST API integrated with Fabric Gateway SDK.
- [x] **Frontend UI**: React + Vite + Tailwind CSS ("Parchment & Ink" design) with Dashboard and Transfer workflows.

---

## Phase 2: Decentralization & Security (Completed ✅)
- [x] **Admin Data Sync**: 
    - [x] Implement off-chain database (PostgreSQL) storage for asset caching.
    - [x] Create synchronization logic/endpoint for Ledger-to-Database migration.
    - [x] Support "Dual Source" (Blockchain vs Database) viewing in Admin UI.
- [x] **Admin Panel**: 
    - [x] Dedicated route `/admin` for system oversight.
    - [x] Identity Audit page (CA identities + Mock DB profiles).
    - [x] Global Inventory page with Source switcher and Manual Sync.
- [x] **UI/UX Enhancements**:
    - [x] Dynamic navigation ("Back to Assets" vs "Back to Registry").
    - [x] Loading/Syncing indicators and visual status badges.

## Phase 3: Advanced Features (Short Term)
- [ ] **IPFS Integration**: 
    - Transition to decentralized IPFS for robust, censorship-resistant storage.
    - Backend: Implement file upload endpoint to push images to IPFS.
    - Store the IPFS Content ID (CID) on-chain.
- [ ] **Role Management**:
    - **Admin**: User Management (Approvals/Ban), System Configuration.
    - **Auditor**: Read-only access to full history logs, compliance reporting view.
    - **User**: Standard Asset Management (Create, Transfer, View Own Assets).
- [ ] **Advanced Authentication**:
    - Implement persistent sessions and secure token refresh.
- [ ] **Auditor View**: Advanced filtering and report generation.

## Phase 3: Advanced Features (Medium Term)
- [ ] **Object Storage (MinIO)**:
    - Secondary storage for larger documents or metadata that doesn't fit on IPFS.
    - S3-compatible integration for legacy support.
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

## License
MIT
