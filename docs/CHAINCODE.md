# 📜 Chaincode Technical Documentation: Ownership Registry

This document provides a technical overview of the `ownership-registry` smart contract implemented in Go for Hyperledger Fabric.

## 🏛 Architecture Overview
The chaincode manages digital assets and their ownership provenance using a **Two-Factor Transfer Workflow** and **Role-Based Access Control (RBAC)**. It is designed to be deployed using the **Chaincode-as-a-Service (CaaS)** model.

---

## 🏗 Data Models

### 1. Asset Structure
The core entity stored in the World State.
| Field | Type | Description |
| :--- | :--- | :--- |
| `ID` | `string` | Unique identifier for the asset. |
| `Name` | `string` | Display name. |
| `Description` | `string` | Detailed information about the artifact. |
| `OwnerID` | `string` | Current owner in `MSPID::Username` format. |
| `ProposedOwnerID`| `string` | Target owner during a pending transfer. |
| `Status` | `string` | Current lifecycle state (see Constants). |
| `View` | `string` | Visibility tier (see Constants). |
| `ImageURL/Hash` | `string` | External reference and integrity check for media. |
| `LastUpdatedBy` | `string` | Identifier of the last actor who modified the state. |
| `LastUpdatedAt` | `string` | RFC3339 timestamp of the last mutation. |

### 2. History Record
Returned by the provenance query to show a time-series of changes.
- `TxId`: Fabric Transaction ID.
- `Timestamp`: Time of transaction.
- `ActorID`: Who performed the action.
- `ActionType`: Categorized action (e.g., CREATE, TRANSFER_PROPOSE).
- `Value`: Snapshot of the asset at that time.

---

## 📌 Constants & Enumerations

### Asset Status
- `ACTIVE`: Normal operating state.
- `FROZEN`: Administrative lock; no transfers allowed.
- `DELETED`: Soft-deleted or revoked.
- `PENDING_TRANSFER`: Locked in a transfer negotiation.

### Asset View
- `PUBLIC`: Visible to all authenticated users.
- `PRIVATE`: Visible only to the owner and administrators.

---

## ⚡ Core Functions

### Ledger Lifecycle
- `InitLedger()`: Initializes the world state with a "Genesis Asset".

### Asset Management
- `CreateAsset(id, name, desc, url, hash, view)`: Issues a new asset. The caller is automatically assigned as the `OwnerID`.
- `ReadAsset(id)`: Returns the current state of a specific asset.
- `UpdateAssetView(id, newView)`: **(Owner Only)** Toggles between `PUBLIC` and `PRIVATE`.

### Two-Step Transfer Workflow
To prevent accidental transfers, the process requires two distinct transactions:
1. `ProposeTransfer(id, targetOwner)`: **(Owner Only)** Sets the asset to `PENDING_TRANSFER` and names a recipient.
2. `AcceptTransfer(id)`: **(Proposed Owner Only)** Finalizes the change of ownership and returns state to `ACTIVE`.

### Administrative Operations
- `UpdateAssetStatus(id, newStatus)`: **(Admin Only)** Allows Orgs or users with the `admin=true` attribute to override asset status (Freeze/Revoke).

### Query & Provenance
- `GetAssetHistory(id)`: Returns the full audit trail of the asset from the ledger's history database.
- `GetAllAssets()`: Performs a range query to return all items. This is used by the Admin Dashboard for real-time monitoring of the chaincode state.

---

## 🔐 Identity & Security
The chaincode uses a custom helper `getClientFullIdentifier` to extract a unique, human-readable identity:
1. It tries to fetch the `hf.EnrollmentID` from the client certificate.
2. If missing, it falls back to the `CommonName (CN)`.
3. It prefixes this with the user's `MSPID` (e.g., `Org1MSP::alice`).

**Access Control Logic:**
- **Ownership**: Verified by comparing the transaction creator's identity against the stored `OwnerID`.
- **Administrative Rights**: Verified via the Client Identity (CID) library, checking for the `admin` attribute or membership in the Org's MSP (Org1MSP/Org2MSP defaults).
