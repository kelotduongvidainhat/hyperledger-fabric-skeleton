# Hyperledger Fabric Skeleton Project (Refactored)

A comprehensive starter template for building production-grade Hyperledger Fabric applications. This project has been refactored to include advanced **User Management**, **Asset Locking**, and **Blockchain-to-Database Synchronization**.

## 🌟 Key Features

- **Full Stack Architecture:** 
  - **Blockchain:** Hyperledger Fabric 2.5 (2 Orgs, 1 Orderer)
  - **Chaincode:** Go (Asset Transfer + User Management)
  - **Backend:** Go/Gin REST API with Fabric SDK
  - **Frontend:** React/Vite with Tailwind CSS
  - **Database:** PostgreSQL for off-chain query optimization

- **Advanced Functionality:**
  - **Asset Lifecycle:** Create, Read, Update, Delete, Transfer with full history auditing.
  - **Admin Dashboard:** Integrated oversight of all network identities and assets.
  - **Data Sync:** Manual and automated "Dual Source" synchronization (Ledger ↔ PostgreSQL).
  - **Security:** RBAC (Role Based Access Control) for standard users and administrators.

- **Developer Experience:**
  - **One-Click Setup:** `./scripts/fresh-start.sh` handles EVERYTHING
  - **Dockerized:** All components run in containers
  - **Documentation:** Comprehensive guides in `docs/`

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.25+
- Node.js & NPM
- JQ & Curl

### Quick Start Guide
For a step-by-step walkthrough of setting up the environment, deploying chaincode, and starting the API, see:
👉 **[QUICKSTART.md](QUICKSTART.md)**

### Basic Run
The `fresh-start.sh` script bootstraps the core network infrastructure.
```bash
sudo ./scripts/fresh-start.sh
```

Once complete:
- **Frontend:** [http://localhost:5173](http://localhost:5173)
- **Backend API:** [http://localhost:3000](http://localhost:3000)
- **Database:** PostgreSQL on `localhost:5432`

## 📂 Project Structure

```
├── backend/            # Go REST API & Event Listener
├── chaincode/          # Go Chaincode (Smart Contracts)
├── database/           # PostgreSQL Schema & Init Scripts
├── docs/               # Detailed Documentation
├── frontend/           # React Web Application
├── network/            # Fabric Network Config (Docker, Crypto)
└── scripts/            # Automation Scripts (Setup, Sync, Test)
```

## 📖 Documentation

- **[Setup Guide](docs/SETUP_COMPLETE.md)** - Detailed system status and usage
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common fixes
- **[Deployment Guide](docs/DEPLOYMENT_GUIDE.md)** - Moving to production

## 🔄 Recent Updates (Privacy & UI Specialization)
- **Privacy Filtering:** Backend now enforces strict visibility rules. Non-admins can only see `PUBLIC` assets or those they own.
- **Gallery vs. Collection:** Frontend split into a global **Public Gallery** (database-driven discovery) and a personal **My Collection** (management workspace).
- **Read-Only Detail Pages:** Specialized `/gallery/:id` route for browsing public artifacts without management overhead.
- **Enhanced JWT Identity:** Token now includes user organization for robust cross-layer identity verification.
- **Standardized Visibility:** Unified all asset visibility to uppercase `PUBLIC` and `PRIVATE` constants across Chaincode, Backend, and Frontend.

## License
MIT