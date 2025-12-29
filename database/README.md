# Database Layer (Off-Chain)

This directory manages off-chain PostgreSQL storage for the Hyperledger Fabric network.

## Purpose

While the blockchain acts as the immutable ledger, a traditional database is used for:
- **CQRS Pattern**: Fast, complex queries without burdening the blockchain.
- **Event Synchronization**: Automatic replication of chaincode events to SQL tables.
- **Analytics and Reporting**: Rich SQL queries for business intelligence.

## Implementation

### PostgreSQL Service
- **Version**: PostgreSQL 15
- **Port**: 5432
- **Database**: `fabric_assets`
- **Deployment**: Docker container (`docker-postgres-1`)

### Schema
See [`schema.sql`](schema.sql) for the complete database schema.

**Assets Table**:
```sql
CREATE TABLE assets (
    id VARCHAR(64) PRIMARY KEY,
    color VARCHAR(32),
    size INT,
    owner VARCHAR(64),
    appraised_value INT,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Event Synchronization
The backend service (`backend/internal/listener/service.go`) listens for chaincode events:
- `AssetCreated` → INSERT/UPSERT into `assets` table
- `AssetTransferred` → UPDATE `owner` field

### API Endpoints
- `GET /api/query/assets` - Query assets from PostgreSQL (fast, off-chain)
- `GET /assets` - Query assets from Blockchain (authoritative, on-chain)

### Usage

**Apply Schema**:
```bash
cat database/schema.sql | docker exec -i docker-postgres-1 psql -U postgres -d fabric_assets
```

**Query Database Directly**:
```bash
docker exec -it docker-postgres-1 psql -U postgres -d fabric_assets -c "SELECT * FROM assets;"
```

**Frontend Toggle**:
The React dashboard includes a "Source" toggle to switch between Blockchain and Database views.

## Architecture

```
Chaincode (CreateAsset/TransferAsset)
    ↓ (emits event)
Backend Listener
    ↓ (writes to)
PostgreSQL
    ↑ (reads from)
Frontend (via /api/query/assets)
```
