# Backend API - Asset Transfer

This is a REST API built with Go and the Fabric Gateway SDK to interact with the Hyperledger Fabric network.

## 🚀 Getting Started

### Prerequisites
- Go 1.25+
- Hyperledger Fabric network running (use `./scripts/network-setup.sh`)

### Running the API
```bash
go run main.go
```
The server will start on `http://localhost:8080`.

## 📌 API Endpoints

### 1. Health Check
`GET /health`
Returns the status of the API.

### 2. Get All Assets
`GET /assets`
Returns a list of all assets on the ledger.

### 3. Get Asset by ID
`GET /assets/:id`
Returns the details of a specific asset.

### 4. Create Asset
`POST /assets`
Creates a new asset.
**Body:**
```json
{
  "id": "asset100",
  "color": "blue",
  "size": 5,
  "owner": "John",
  "appraisedValue": 500
}
```

### 5. Transfer Asset
`PUT /assets/:id/transfer`
Updates the owner of an asset.
**Body:**
```json
{
  "newOwner": "Alice"
}
```

## 🛠️ Implementation Details
- **Framework**: [Gin](https://github.com/gin-gonic/gin)
- **SDK**: [Fabric Gateway Go SDK](https://github.com/hyperledger/fabric-gateway)
- **Identity**: Uses the `Admin@org1.example.com` identity (configured in `internal/fabric/client.go`).
- **Connection**: Connects via gRPC to `peer0.org1.example.com` at `localhost:7051`.
