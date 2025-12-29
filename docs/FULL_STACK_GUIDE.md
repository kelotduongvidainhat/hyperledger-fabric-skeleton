# Frontend & Backend Documentation

## 🌐 Full Stack Architecture

This project implements a complete DApp architecture:

`[React Frontend]` <---> `[Go Backend API]` <---> `[Fabric Gateway]` <---> `[Fabric Network]`
                                      |
                                      v
                             `[PostgreSQL Database]` (Off-chain Replica)

### 1. Frontend (Port 5173)
- **Tech**: React, Vite, TailwindCSS
- **Key Features**:
  - **Identity Switcher**: Users can toggle between `Admin` and `User` roles.
  - **Asset Dashboard**: Real-time view of assets on the ledger.
  - **Interactivity**: Direct Create/Transfer actions.
  - **Data Source Toggle**: Switch between live `Blockchain` data and off-chain `Database` replica.

### 2. Backend (Port 8080)
- **Tech**: Go, Gin, Fabric Gateway SDK
- **Key Features**:
  - **Dynamic Identity**: Uses the `X-User-ID` header to determine which wallet identity to sign transactions with.
  - **CORS**: Configured to accept requests from localhost:5173.
  - **Wallet Management**: Intelligent loading of standard MSP structures (`msp/signcerts`, `msp/keystore`).
  - **Event Listener**: Automatically syncs `AssetCreated` and `AssetTransferred` events to PostgreSQL.

### 3. Database (Port 5432)
- **Tech**: PostgreSQL 15, Docker
- **Schema**: Simple `assets` table mirroring chaincode state.
- **Pattern**: CQRS (Command Query Responsibility Segregation) - Writes go to Blockchain, Complex Reads can go to DB.

## 🚀 Running the Application

### Start the Backend
```bash
cd backend
go run main.go
# Server listens on :8080
```

### Start the Frontend
```bash
cd frontend
npm install
npm run dev
# App accessible at http://localhost:5173
```
