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
  - **Asset Lifecycle:** Create, Read, Update, Delete, Transfer
  - **Security:** Asset Locking/Unlocking (Freezing)
  - **User Identity:** On-chain User profiles and role management
  - **Data Sync:** Real-time event listener + Backfill scripts for 100% data consistency

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

### Run the Project
The `fresh-start.sh` script is the magic button. It rebuilds the network, deploys chaincode, and starts the app.

```bash
./scripts/fresh-start.sh
```

Once complete:
- **Frontend:** [http://localhost:5173](http://localhost:5173)
- **Backend API:** [http://localhost:8080/health](http://localhost:8080/health)
- **Database:** Connect via `localhost:5432`

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

## 🔄 Recent Updates (User Management Refactor)
- **New Asset Fields:** `Name`, `Category`, `Status` (Replacing Color/Size)
- **User Management:** Create/Update users on-chain
- **TLS Fixes:** Improved volume cleanup to solve certificate errors

## License
MIT