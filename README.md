# Hyperledger Fabric Full-Stack Application

A comprehensive full-stack project structure for Hyperledger Fabric, separated into 5 distinct layers.

## Project Architecture

```
├── frontend/           # (UI Layer) Web application for end-users
├── backend/            # (API Layer) REST/gRPC server handling business logic & Gateway SDK
├── chaincode/          # (Ledger Layer) Smart Contracts defining on-chain logic
│   └── smart-contract/ 
├── network/            # (Infra Layer) Hyperledger Fabric network configuration
│   ├── config/         
│   └── docker/         
├── database/           # (Off-Chain Layer) Auxiliary storage (e.g., Postres/MongoDB) for caching/metadata
└── scripts/            # Automation for deployment and testing
```

## Network Architecture

The project runs a 2-Organization Hyperledger Fabric network (v2.5) with Raft consensus.

| Service | Container Name | Port (Host) | Description |
|---------|----------------|-------------|-------------|
| **Orderer** | `orderer.example.com` | 7050 | Raft Consensus Service (Term 2 Leader) |
| **Org1 Peer** | `peer0.org1.example.com` | 7051 | Anchor Peer for Org1 |
| **Org2 Peer** | `peer0.org2.example.com` | 9051 | Anchor Peer for Org2 |
| **Org1 CA** | `ca_org1` | 7054 | Certificate Authority for Org1 |
| **Org2 CA** | `ca_org2` | 8054 | Certificate Authority for Org2 |
| **Orderer CA** | `ca_orderer` | 9054 | Certificate Authority for Orderer |
| **CouchDB** | `couchdb0`, `couchdb1` | 5984, 7984 | World State Databases |
| **PostgreSQL** | `postgres` | 5432 | Off-chain Database (Event Sync) |
| **Backend API** | `backend-api` | 8080 | Go REST API (Gin + Fabric Gateway) |
| **Frontend** | `vite-react` | 5173 | React DApp Interface |

## Current Status

- [x] **Network Up**: All containers running (Fabric 2.5.14)
- [x] **Channel Created**: `mychannel` initialized and joined by both peers
- [x] **Anchor Peers**: Configured for cross-org discovery
- [x] **Chaincode Deployed**: `asset-transfer` v1.0 (CaaS) - ✅ **OPERATIONAL**
- [x] **Fabric CA**: 3 CAs running (v1.5.15)
- [x] **CouchDB**: 2 instances running (v3.4.2)
- [x] **Backend API**: REST API with Dynamic Identity Support (Port 8080)
- [x] **Frontend**: React Dashboard for Asset Management (Port 5173)
- [x] **PostgreSQL**: Off-chain database with automatic event synchronization

## 📚 Documentation

All documentation is located in the **[`docs/`](docs/)** folder:

- **[docs/README.md](docs/README.md)** - Documentation index
- **[docs/SETUP_COMPLETE.md](docs/SETUP_COMPLETE.md)** - Complete system overview
- **[docs/DEPLOYMENT_GUIDE.md](docs/DEPLOYMENT_GUIDE.md)** - Chaincode deployment guide
- **[docs/QUICK_REFERENCE.md](docs/QUICK_REFERENCE.md)** - Quick command reference
- [docs/IDENTITY_MANAGEMENT.md](docs/IDENTITY_MANAGEMENT.md) - Fabric CA, ABAC, and CRL guide
- [docs/FULL_STACK_GUIDE.md](docs/FULL_STACK_GUIDE.md) - Frontend & Backend Architecture
- [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) - Common issues and solutions

## Getting Started

### 1. Verify Network
```bash
cd network/docker
docker-compose ps
```

### 2. Test Chaincode
```bash
# Query all assets
docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{"Args":["GetAllAssets"]}'
```

### 3. Create a New Asset
```bash
docker exec cli bash -c '
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt

peer chaincode invoke \
  -o orderer.example.com:7050 \
  --tls --cafile $ORDERER_CA \
  -C mychannel -n asset-transfer \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
  -c '"'"'{"function":"CreateAsset","Args":["myasset","blue","10","John","500"]}'"'"'
'
```

## Quick Links

- **Network Config**: [`network/config/configtx.yaml`](network/config/configtx.yaml)
- **Chaincode**: [`chaincode/asset-transfer/`](chaincode/asset-transfer/)
- **Scripts**: [`scripts/`](scripts/)
- **Documentation**: [`docs/`](docs/)

---

**Status**: ✅ **Fully Operational** | **Last Updated**: 2025-12-29