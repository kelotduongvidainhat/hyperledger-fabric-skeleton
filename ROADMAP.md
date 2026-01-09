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
- [x] **UI/UX Enhancements & Privacy**:
    - [x] Dynamic navigation ("Back to Assets" vs "Back to Registry").
    - [x] Loading/Syncing indicators and visual status badges.
    - [x] Split Discovery Gallery from Personal Management Dashboard.
    - [x] Implementation of Backend Privacy Filtering (PUBLIC/PRIVATE tiers).
- [x] **Dev Experience**:
    - [x] CaaS External Builder configuration for stable production-like connections.
    - [x] Automated Data Initialization (10 Users/Assets) via `init-data.sh`.

## Phase 3: Advanced Features (Short Term)
- [X] **IPFS Integration**: 
    - [x] Transition to decentralized IPFS for robust, censorship-resistant storage.
    - [X] Backend: Implement file upload endpoint to push images to IPFS.
    - [X] Store the IPFS Content ID (CID) on-chain.
    - [ ] **Long Term**: Integrate with Pinata/managed IPFS services for persistent pinning and better availability.
- [x] **Role Management**:
    - [x] **Admin**: User Management (Approvals/Ban), Asset Status Override (Freeze/Revoke).
    - [x] **Auditor**: Profile promotion and dedicated role assignment.
    - [x] **User**: Standard Asset Management (Create, Transfer, View Own Assets).
- [x] **Event Monitoring**:
    - [x] Integrated real-time blockchain feed into the Admin Dashboard overview.
- [x] **Advanced Authentication**:
    - [x] Implement persistent sessions and secure token refresh via HttpOnly cookies.
- [ ] **Auditor View**: Advanced filtering and report generation.

## Phase 3: Advanced Features (Medium Term)
- [x] **Policy-as-Code (Open Policy Agent)**:
    - [x] Decouple authorization logic from the Backend into a dedicated OPA service.
    - [x] Implement `rego` policies for asset lifecycle (e.g., specific transfer window rules).
    - [x] Centralize "Who can do what" management across the entire network.
- [x] **Object Storage (MinIO)**:
    - [x] Secondary storage for supporting documentation and high-performance image delivery.
    - [x] Backend: Integrated MinIO Go SDK for S3-compatible operations.
    - [x] Frontend: Implementation of Pre-signed URLs for secure, authenticated media access.
    - [x] UI: Secure "View" vs "Download" options for asset attachments.
- [ ] **Private Data Collections (PDC)**:
    - Update Chaincode to support private pricing or metadata that is shared only between Buyer and Seller, not the whole network.
- [ ] **Complex Transfer Logic**:
    - Implement "Offer/Bid" mechanism with price negotiation on-chain.
    - Add "Escrow" functionality.

## Phase 4: Production Readiness (Long Term)
- [ ] **API Gateway (Nginx)**:
    - [ ] Implement a unified entry point (Port 80) for Frontend and Backend.
    - [ ] Routing: `/api/*` to Backend, `/*` to Frontend.
    - [ ] Configure Rate Limiting and SSL (Let's Encrypt placeholder).
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
