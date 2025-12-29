# Frontend & Backend Documentation

## 🌐 Full Stack Architecture

This project implements a complete DApp architecture:

`[React Frontend]` <---> `[Go Backend API]` <---> `[Fabric Gateway]` <---> `[Fabric Network]`

### 1. Frontend (Port 5173)
- **Tech**: React, Vite, TailwindCSS
- **Key Features**:
  - **Identity Switcher**: Users can toggle between `Admin` and `User` roles.
  - **Asset Dashboard**: Real-time view of assets on the ledger.
  - **Interactivity**: Direct Create/Transfer actions.

### 2. Backend (Port 8080)
- **Tech**: Go, Gin, Fabric Gateway SDK
- **Key Features**:
  - **Dynamic Identity**: Uses the `X-User-ID` header to determine which wallet identity to sign transactions with.
  - **CORS**: Configured to accept requests from localhost:5173.
  - **Wallet Management**: Intelligent loading of standard MSP structures (`msp/signcerts`, `msp/keystore`).

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
