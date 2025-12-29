# Network Infrastructure

The network is built using Hyperledger Fabric v2.5 and orchestrated via Docker Compose.

## 1. Topology
The network consists of **Two Peer Organizations** and a **Raft Orderer Service**.

### Organizations
- **OrdererOrg**: Manages the ordering service.
    - MSP ID: `OrdererMSP`
    - CA: `ca_orderer` (Port 9054)
    - Orderer: `orderer.example.com` (Port 7050)
- **Org1**:
    - MSP ID: `Org1MSP`
    - CA: `ca_org1` (Port 7054)
    - Peer: `peer0.org1.example.com` (Port 7051) + CouchDB (Port 5984)
- **Org2**:
    - MSP ID: `Org2MSP`
    - CA: `ca_org2` (Port 8054)
    - Peer: `peer0.org2.example.com` (Port 9051) + CouchDB (Port 7984)

## 2. Directory Structure
```
network/
├── config/              # Configuration files
│   ├── configtx.yaml    # Channel configuration definitions
│   └── crypto-config.yaml # Crypto material generation specs
├── crypto-config/       # Generated certificates and keys (DO NOT COMMIT)
├── channel-artifacts/   # Genesis block and channel transactions (DO NOT COMMIT)
└── docker/              # Docker Compose files
    └── docker-compose.yaml
```

## 3. Administration
Use the `cli` container to perform network operations. It is pre-mounted with all necessary crypto material.

### Connecting to CLI
```bash
docker exec -it cli bash
```

### Common Commands (Inside CLI)

**List Channels (Org1):**
```bash
CORE_PEER_LOCALMSPID=Org1MSP \
CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp \
CORE_PEER_ADDRESS=peer0.org1.example.com:7051 \
peer channel list
```

## 4. Credentials
- **Root CAs**: Located in `network/crypto-config/peerOrganizations/<org>/ca/`
- **User Keys**: Located in `network/crypto-config/peerOrganizations/<org>/users/User1@<org>/msp/keystore/`

## 5. Setup Instructions

Follow these steps to generate artifacts and start the network.

**Prerequisites:** Ensure `bin/` contains Fabric binaries (`cryptogen`, `configtxgen`, etc.).

### Step 1: Generate Crypto Material
```bash
./bin/cryptogen generate --config=network/config/crypto-config.yaml --output=network/crypto-config
```

### Step 2: Generate Channel Artifacts
```bash
export FABRIC_CFG_PATH=$PWD/network/config
mkdir -p network/channel-artifacts

# Genesis Block
./bin/configtxgen -profile TwoOrgsOrdererGenesis -channelID system-channel -outputBlock network/channel-artifacts/genesis.block

# Channel Creation Transaction
./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx network/channel-artifacts/mychannel.tx -channelID mychannel

# Anchor Peer Updates
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate network/channel-artifacts/Org1MSPanchors.tx -channelID mychannel -asOrg Org1MSP
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate network/channel-artifacts/Org2MSPanchors.tx -channelID mychannel -asOrg Org2MSP
```

### Step 3: Start Network
```bash
cd network/docker
docker-compose up -d
```

### Step 4: Create & Join Channel
Enter the CLI container:
```bash
docker exec -it cli bash
```

Run the following inside the CLI:
```bash
# 1. Create Channel
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/mychannel.tx --outputBlock ./channel-artifacts/mychannel.block --tls --cafile $ORDERER_CA

# 2. Join Org1
CORE_PEER_LOCALMSPID=Org1MSP \
CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp \
CORE_PEER_ADDRESS=peer0.org1.example.com:7051 \
peer channel join -b ./channel-artifacts/mychannel.block

# 3. Join Org2
CORE_PEER_LOCALMSPID=Org2MSP \
CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
peer channel join -b ./channel-artifacts/mychannel.block
```
