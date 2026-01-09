# Application Workflows & Sequence Diagrams

This document outlines the core operational flows between the Frontend, Backend, PostgreSQL Database, and Hyperledger Fabric Ledger.

## 1. Admin Oversight & Real-time Monitoring
This flow occurs when an administrator loads the **Admin Dashboard**. It performs a hybrid lookup of both cached database stats and live blockchain records.

```mermaid
sequenceDiagram
    participant Admin as Admin Browser
    participant API as Backend (Fiber)
    participant OPA as Policy Agent (OPA)
    participant DB as PostgreSQL
    participant Fabric as Chaincode (Go)

    Admin->>API: GET /admin/stats
    API->>OPA: CheckAuthorization(user, role...)
    OPA-->>API: Result: ALLOW
    API->>DB: Count Assets & Owners
    DB-->>API: Row Counts
    API-->>Admin: NetworkStats JSON

    Admin->>API: GET /admin/assets?source=blockchain
    API->>API: Load Admin Identity from Wallet
    API->>Fabric: EvaluateTransaction("GetAllAssets")
    Fabric-->>API: Raw Ledger State
    API->>API: Flatten Audit Metadata
    API-->>Admin: Recent Ledger Activity JSON
```

## 2. Ledger-to-Database Synchronization
This flow ensures the off-chain search engine (PostgreSQL) is consistent with the immutable blockchain ledger.

```mermaid
sequenceDiagram
    participant Admin as Admin Browser
    participant API as Backend (Fiber)
    participant Fabric as Chaincode (Go)
    participant DB as PostgreSQL

    Admin->>API: POST /admin/sync
    API->>OPA: CheckAuthorization(user, role...)
    OPA-->>API: Result: ALLOW
    API->>API: Establish Admin Gateway
    API->>Fabric: EvaluateTransaction("GetAllAssets")
    Fabric-->>API: Full Ledger Array (Asset + Audit)
    
    loop For each Asset in Array
        API->>DB: UPSERT (Save) Asset Record
        DB-->>API: OK
    end
    
    API-->>Admin: Sync Complete (Count: N)
```

## 3. Asset Lifecycle (Dual Storage & On-Chain)
When a user creates an asset, the backend orchestrates a dual-upload to both IPFS (for permanence) and MinIO (for fast delivery) before committing the registry record to the blockchain.

```mermaid
sequenceDiagram
    participant User as User Browser
    participant API as Backend (Fiber)
    participant Storage as Storage (MinIO + IPFS)
    participant Fabric as Chaincode (Go)
    participant Listener as Event Listener (Go Routine)
    participant DB as PostgreSQL

    User->>API: POST /assets (CreateAsset)
    API->>API: Process Form Data
    
    par Dual Storage Upload
        API->>Storage: PUSH to MinIO (Performance)
        API->>Storage: PUSH to IPFS (Provenance)
    end
    
    Storage-->>API: returns {StoragePath, IpfsCID}
    
    API->>Fabric: SubmitTransaction("CreateAsset", {..., StoragePath, IpfsCID})
    Fabric->>Fabric: Commit to Ledger & Emit EVENT
    Fabric-->>API: Success
    API-->>User: "Asset Registered"
    
    Note over Fabric, Listener: Asynchronous Event Notification
    
    Fabric-))Listener: ChaincodeEvent (com.own.registry:AssetCreated)
    Listener->>API: Process Event Payload
    Listener->>DB: UPSERT Registry Cache (assets table)
    DB-->>Listener: Persisted
```

## 4. Administrative Control (Freeze/Revoke)
How an administrator overrides asset state for governance purposes.

```mermaid
sequenceDiagram
    participant Admin as Admin Browser
    participant API as Backend (Fiber)
    participant Fabric as Chaincode (Go)
    participant DB as PostgreSQL

    Admin->>API: POST /admin/assets/:id/status (e.g., FROZEN)
    API->>OPA: CheckAuthorization(user, role...)
    OPA-->>API: Result: ALLOW
    API->>Fabric: SubmitTransaction("UpdateAssetStatus", id, "FROZEN")
    Fabric->>Fabric: Update World State
    Fabric-->>API: Success
    API->>DB: UPDATE assets SET status = 'FROZEN' WHERE id = :id
    DB-->>API: Updated
    API-->>Admin: "Asset status updated: FROZEN"
```
