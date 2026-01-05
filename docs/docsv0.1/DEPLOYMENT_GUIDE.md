# Chaincode Deployment Guide - Chaincode-as-a-Service (CaaS)

This guide provides step-by-step instructions for deploying the Asset Transfer chaincode using Chaincode-as-a-Service (CaaS).

---

## Prerequisites

- ✅ Network is running (`docker ps` shows all containers up)
- ✅ Channel `mychannel` is created
- ✅ Both peers joined the channel
- ✅ Go 1.25.2 installed
- ✅ Docker and Docker Compose installed

---

## Deployment Steps

### Step 1: Build Chaincode Docker Image

```bash
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer

# Build the Docker image
docker build -t asset-transfer:1.0 .

# Verify image was created
docker images | grep asset-transfer
```

**Expected Output:**
```
asset-transfer   1.0   <image-id>   <time>   <size>
```

---

### Step 2: Create CaaS Package Structure

```bash
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer

# Create package directory
mkdir -p caas-pkg

# Create connection.json
cat > caas-pkg/connection.json <<'EOF'
{
  "address": "asset-transfer-org1:9999",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF

# Create metadata.json
cat > caas-pkg/metadata.json <<'EOF'
{
  "type": "ccaas",
  "label": "asset-transfer_1.0"
}
EOF

# Package the chaincode
cd caas-pkg
tar cfz code.tar.gz connection.json
cd ..
tar cfz asset-transfer.tar.gz -C caas-pkg metadata.json code.tar.gz

# Verify package was created
ls -lh asset-transfer.tar.gz
```

---

### Step 3: Install Chaincode on Org1 Peer

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode install /opt/gopath/src/github.com/chaincode/asset-transfer/asset-transfer.tar.gz
'
```

**Expected Output:**
```
Chaincode code package identifier: asset-transfer_1.0:<package-id>
```

**Save the Package ID!** You'll need it in later steps.

---

### Step 4: Install Chaincode on Org2 Peer

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org2MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

peer lifecycle chaincode install /opt/gopath/src/github.com/chaincode/asset-transfer/asset-transfer.tar.gz
'
```

**Expected Output:**
```
Chaincode code package identifier: asset-transfer_1.0:<same-package-id>
```

---

### Step 5: Query Installed Chaincode (Get Package ID)

```bash
# Query on Org1
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode queryinstalled
'
```

**Copy the Package ID** from the output. It will look like:
```
Package ID: asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f
```

**Set it as a variable:**
```bash
export PACKAGE_ID="asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f"
```
*(Replace with your actual package ID)*

---

### Step 6: Start Chaincode Containers

```bash
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer

# Set environment variables
export CHAINCODE_ID_ORG1=$PACKAGE_ID
export CHAINCODE_ID_ORG2=$PACKAGE_ID

# Start the chaincode containers
docker-compose up -d

# Verify containers are running
docker ps | grep asset-transfer
```

**Expected Output:**
```
asset-transfer-org1   Up   0.0.0.0:9991->9999/tcp
asset-transfer-org2   Up   0.0.0.0:9992->9999/tcp
```

**Wait 5-10 seconds** for the chaincode containers to be ready.

---

### Step 7: Approve Chaincode for Org1

```bash
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode approveformyorg \
  -o orderer.example.com:7050 \
  --channelID mychannel \
  --name asset-transfer \
  --version 1.0 \
  --package-id $PACKAGE_ID \
  --sequence 1 \
  --tls \
  --cafile \$ORDERER_CA
"
```

**Expected Output:**
```
txid [...] committed with status (VALID)
```

---

### Step 8: Approve Chaincode for Org2

```bash
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org2MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode approveformyorg \
  -o orderer.example.com:7050 \
  --channelID mychannel \
  --name asset-transfer \
  --version 1.0 \
  --package-id $PACKAGE_ID \
  --sequence 1 \
  --tls \
  --cafile \$ORDERER_CA
"
```

---

### Step 9: Check Commit Readiness

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode checkcommitreadiness \
  --channelID mychannel \
  --name asset-transfer \
  --version 1.0 \
  --sequence 1 \
  --tls \
  --cafile $ORDERER_CA \
  --output json
'
```

**Expected Output:**
```json
{
  "approvals": {
    "Org1MSP": true,
    "Org2MSP": true
  }
}
```

---

### Step 10: Commit Chaincode Definition

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode commit \
  -o orderer.example.com:7050 \
  --channelID mychannel \
  --name asset-transfer \
  --version 1.0 \
  --sequence 1 \
  --tls \
  --cafile $ORDERER_CA \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
'
```

**Expected Output:**
```
txid [...] committed with status (VALID)
```

---

### Step 11: Verify Chaincode is Committed

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode querycommitted --channelID mychannel --name asset-transfer
'
```

**Expected Output:**
```
Committed chaincode definition for chaincode 'asset-transfer' on channel 'mychannel':
Version: 1.0, Sequence: 1, Endorsement Plugin: escc, Validation Plugin: vscc
Approvals: [Org1MSP: true, Org2MSP: true]
```

---

### Step 12: Initialize the Ledger

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer chaincode invoke \
  -o orderer.example.com:7050 \
  --tls \
  --cafile $ORDERER_CA \
  -C mychannel \
  -n asset-transfer \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
  -c '"'"'{"function":"InitLedger","Args":[]}'"'"'
'
```

**Expected Output:**
```
txid [...] committed with status (VALID)
```

**Wait 3-5 seconds** for the transaction to be committed.

---

### Step 13: Query All Assets

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer chaincode query -C mychannel -n asset-transfer -c '"'"'{"Args":["GetAllAssets"]}'"'"'
'
```

**Expected Output:**
```json
[
  {"ID":"asset1","Color":"blue","Size":5,"Owner":"Tomoko","AppraisedValue":300},
  {"ID":"asset2","Color":"red","Size":5,"Owner":"Brad","AppraisedValue":400},
  {"ID":"asset3","Color":"green","Size":10,"Owner":"Jin Soo","AppraisedValue":500},
  {"ID":"asset4","Color":"yellow","Size":10,"Owner":"Max","AppraisedValue":600},
  {"ID":"asset5","Color":"black","Size":15,"Owner":"Adriana","AppraisedValue":700},
  {"ID":"asset6","Color":"white","Size":15,"Owner":"Michel","AppraisedValue":800}
]
```

---

## Testing the Chaincode

### Create a New Asset

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer chaincode invoke \
  -o orderer.example.com:7050 \
  --tls \
  --cafile $ORDERER_CA \
  -C mychannel \
  -n asset-transfer \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  -c '"'"'{"function":"CreateAsset","Args":["asset7","purple","20","Alice","1000"]}'"'"'
'
```

### Read a Specific Asset

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer chaincode query -C mychannel -n asset-transfer -c '"'"'{"Args":["ReadAsset","asset7"]}'"'"'
'
```

### Transfer an Asset

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer chaincode invoke \
  -o orderer.example.com:7050 \
  --tls \
  --cafile $ORDERER_CA \
  -C mychannel \
  -n asset-transfer \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  -c '"'"'{"function":"TransferAsset","Args":["asset7","Bob"]}'"'"'
'
```

---

## Troubleshooting

### Check Chaincode Container Logs

```bash
# Org1 chaincode logs
docker logs asset-transfer-org1

# Org2 chaincode logs
docker logs asset-transfer-org2
```

### Check Peer Logs

```bash
# Org1 peer logs
docker logs peer0.org1.example.com --tail 50

# Org2 peer logs
docker logs peer0.org2.example.com --tail 50
```

### Restart Chaincode Containers

```bash
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer

# Stop containers
docker-compose down

# Start containers
export CHAINCODE_ID_ORG1=$PACKAGE_ID
export CHAINCODE_ID_ORG2=$PACKAGE_ID
docker-compose up -d
```

### Verify All Containers are Running

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

**Expected containers:**
- orderer.example.com
- peer0.org1.example.com
- peer0.org2.example.com
- couchdb0
- couchdb1
- ca_orderer
- ca_org1
- ca_org2
- cli
- asset-transfer-org1
- asset-transfer-org2

---

## Summary

✅ **Chaincode Type**: Chaincode-as-a-Service (CaaS)  
✅ **Chaincode Name**: asset-transfer  
✅ **Version**: 1.0  
✅ **Channel**: mychannel  
✅ **Organizations**: Org1MSP, Org2MSP  
✅ **External Builder**: ccaas_builder  

**Package ID Format:**
```
asset-transfer_1.0:<sha256-hash>
```

**Chaincode Endpoints:**
- Org1: `asset-transfer-org1:9999` (exposed on host: 9991)
- Org2: `asset-transfer-org2:9999` (exposed on host: 9992)

---

## Next Steps

1. **Deploy Backend Listener** - Connect to the blockchain events
2. **Deploy Frontend** - User interface for asset management
3. **Add More Chaincode Functions** - Extend functionality
4. **Implement Access Control** - Add permission management
5. **Set up Monitoring** - Track chaincode performance

---

**Last Updated**: 2025-12-29  
**Fabric Version**: 2.5.14  
**Go Version**: 1.25.2
