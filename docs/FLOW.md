# Application Workflows & Sequence Diagrams

This document outlines the core operational flows between the Frontend, Backend, PostgreSQL Database, and Hyperledger Fabric Ledger.

## 1. Admin Oversight & Real-time Monitoring
This flow occurs when an administrator loads the **Admin Dashboard**. It performs a hybrid lookup of both cached database stats and live blockchain records.

```mermaid
sequenceDiagram
    participant Admin as Admin Browser
    participant API as Backend (Fiber)
    participant DB as PostgreSQL
    participant Fabric as Chaincode (Go)

    Admin->>API: GET /admin/stats
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
    API->>API: Establish Admin Gateway
    API->>Fabric: EvaluateTransaction("GetAllAssets")
    Fabric-->>API: Full Ledger Array (Asset + Audit)
    
    loop For each Asset in Array
        API->>DB: UPSERT (Save) Asset Record
        DB-->>API: OK
    end
    
    API-->>Admin: Sync Complete (Count: N)
```

## 3. Asset Lifecycle (Creation & Sync)
When a user creates an asset, it is written to the blockchain first. The database is updated via the asynchronous Event Listener.

```mermaid
sequenceDiagram
    participant User as User Browser
    participant API as Backend (Fiber)
    participant Fabric as Chaincode (Go)
    participant Listener as Event Listener (Go Routine)
    participant DB as PostgreSQL

    User->>API: POST /assets (CreateAsset)
    API->>Fabric: SubmitTransaction("CreateAsset", id, name...)
    Fabric->>Fabric: Commit to Ledger & Emit EVENT
    Fabric-->>API: Success
    API-->>User: "Asset Created"
    
    Note over Fabric, Listener: Asynchronous Event Notification
    
    Fabric-))Listener: ChaincodeEvent (com.own.registry:AssetCreated)
    Listener->>API: Process Event Payload
    Listener->>DB: Save to assets table
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
    API->>Fabric: SubmitTransaction("UpdateAssetStatus", id, "FROZEN")
    Fabric->>Fabric: Update World State
    Fabric-->>API: Success
    API->>DB: UPDATE assets SET status = 'FROZEN' WHERE id = :id
    DB-->>API: Updated
    API-->>Admin: "Asset status updated: FROZEN"
```
