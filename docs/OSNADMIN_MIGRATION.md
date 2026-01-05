# Migration to OSNAdmin (System Channel-less)

## Overview
This document describes the migration of the network architecture from using a bootstrap **System Channel** to the modern **System Channel-less** approach managed by **osnadmin**.

### Key Changes

1.  **Removal of System Channel**:
    - The Orderer no longer starts with a `genesis.block` derived from a System Channel profile.
    - `ORDERER_GENERAL_BOOTSTRAPMETHOD` is set to `none`.

2.  **Self-Contained Application Channels**:
    - Channel genesis blocks are now generated including the **Orderer Configuration** (Consenters, BatchSize, etc.).
    - This allows each application channel to be independent.

3.  **OSNAdmin for Management**:
    - The `osnadmin` CLI is used to "join" the Orderer to channels via its new Admin API (port 7053).
    - This replaces the implicit creation via `peer channel create` against a system channel.

## New Workflow

### 1. Generate Genesis Block
Instead of creating a channel transaction (`.tx`), we create a full block (`.block`) that includes the Orderer group.
```bash
configtxgen -profile TwoOrgsChannel -outputBlock ./channel-artifacts/mychannel.block -channelID mychannel
```

### 2. Join Orderer
The Orderer is joined to the channel using its Admin endpoint.
```bash
osnadmin channel join \
    --channelID mychannel \
    --config-block mychannel.block \
    -o orderer.example.com:7053 \
    --ca-file ... \
    --client-cert ... \
    --client-key ...
```

### 3. Join Peers
Peers join using the same block.
```bash
peer channel join -b mychannel.block
```

## Benefits
- **Privacy**: No shared system channel leaking information about other channels.
- **Modularity**: Easier to remove/add orderers or channels without global system updates.
- **Standardization**: Aligns with Hyperledger Fabric v2.5+ best practices.
