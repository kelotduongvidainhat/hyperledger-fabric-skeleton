#!/bin/bash

# Chaincode-as-a-Service (CaaS) Deployment Script
# This script deploys chaincode as external services

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}Chaincode-as-a-Service (CaaS) Deployment${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""

# Configuration
CHAINCODE_NAME="asset-transfer"
CHAINCODE_VERSION="1.0"
CHAINCODE_LABEL="${CHAINCODE_NAME}_${CHAINCODE_VERSION}"
CHANNEL_NAME="mychannel"
CHAINCODE_DIR="/home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer"

# Detect current sequence
echo -e "${YELLOW}Detecting current chaincode sequence...${NC}"
CURRENT_SEQUENCE=$(docker exec cli peer lifecycle chaincode querycommitted --channelID ${CHANNEL_NAME} --name ${CHAINCODE_NAME} 2>/dev/null | grep "Sequence:" | sed 's/.*Sequence: //;s/,.*//')
if [ -z "$CURRENT_SEQUENCE" ]; then
  SEQUENCE=1
  echo "No existing definition found. Starting with sequence 1."
else
  SEQUENCE=$((CURRENT_SEQUENCE + 1))
  echo "Current sequence is $CURRENT_SEQUENCE. Using sequence $SEQUENCE."
fi

cd $CHAINCODE_DIR

# Step 1: Build chaincode Docker image
echo -e "${YELLOW}Step 1: Building chaincode Docker image...${NC}"
docker build -t ${CHAINCODE_NAME}:${CHAINCODE_VERSION} .
echo -e "${GREEN}✓ Docker image built${NC}"
echo ""

# Step 2: Create connection.json files for each peer
echo -e "${YELLOW}Step 2: Creating connection.json files...${NC}"

mkdir -p caas-pkg

cat > caas-pkg/connection.json <<EOF
{
  "address": "asset-transfer-org1:9999",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF

echo -e "${GREEN}✓ Connection files created${NC}"
echo ""

# Step 3: Create metadata.json
echo -e "${YELLOW}Step 3: Creating metadata.json...${NC}"

cat > caas-pkg/metadata.json <<EOF
{
  "type": "ccaas",
  "label": "${CHAINCODE_LABEL}"
}
EOF

echo -e "${GREEN}✓ Metadata created${NC}"
echo ""

# Step 4: Package the chaincode
echo -e "${YELLOW}Step 4: Packaging chaincode for CaaS...${NC}"
cd caas-pkg
tar cfz code.tar.gz connection.json
cd ..
tar cfz ${CHAINCODE_NAME}.tar.gz -C caas-pkg metadata.json code.tar.gz

echo -e "${GREEN}✓ Chaincode packaged${NC}"
echo ""

# Step 5: Install on Org1
echo -e "${YELLOW}Step 5: Installing chaincode on Org1 peer...${NC}"
INSTALL_OUTPUT_ORG1=$(docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode install /opt/gopath/src/github.com/chaincode/asset-transfer/${CHAINCODE_NAME}.tar.gz 2>&1
")

PACKAGE_ID=$(echo "$INSTALL_OUTPUT_ORG1" | grep "Chaincode code package identifier:" | sed 's/.*identifier: //')
if [ -z "$PACKAGE_ID" ]; then
# Fallback: find the package ID with the latest hash (since they are alphabetical, this might not be perfect, but better than tail -n 1 which is non-deterministic)
    # Actually, let's just use queryinstalled and find the one that matches our label
    PACKAGE_ID=$(docker exec cli bash -c "
    export CORE_PEER_LOCALMSPID=Org1MSP
    export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
    peer lifecycle chaincode queryinstalled" | grep "Label: ${CHAINCODE_LABEL}" | grep "Package ID:" | tail -n 1 | sed 's/^Package ID: //' | sed 's/, Label:.*$//' | tr -d '\r\n ')
fi

echo "Package ID: $PACKAGE_ID"
echo -e "${GREEN}✓ Installed and captured ID${NC}"
echo ""

# Step 6: Install on Org2
echo -e "${YELLOW}Step 6: Installing chaincode on Org2 peer...${NC}"
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org2MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

peer lifecycle chaincode install /opt/gopath/src/github.com/chaincode/asset-transfer/${CHAINCODE_NAME}.tar.gz 2>&1 || true
"

echo -e "${GREEN}✓ Installed on Org2${NC}"
echo ""

# Use the same package ID for both since we used the same package file
PACKAGE_ID_ORG1=$PACKAGE_ID
PACKAGE_ID_ORG2=$PACKAGE_ID
echo ""

# Step 8: Start chaincode containers
echo -e "${YELLOW}Step 8: Starting chaincode containers...${NC}"

# Start chaincode containers
# Force recreate containers to ensure env update
export CHAINCODE_ID_ORG1=$PACKAGE_ID_ORG1
export CHAINCODE_ID_ORG2=$PACKAGE_ID_ORG2
docker-compose down
docker-compose up -d

echo -e "${GREEN}✓ Chaincode containers started${NC}"
echo ""

# Wait for containers to be ready
echo -e "${YELLOW}Waiting for chaincode containers to be ready...${NC}"
sleep 10
echo ""

# Step 9: Approve for Org1
echo -e "${YELLOW}Step 9: Approving chaincode for Org1...${NC}"
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt

peer lifecycle chaincode approveformyorg \
  -o orderer.example.com:7050 \
  --channelID ${CHANNEL_NAME} \
  --name ${CHAINCODE_NAME} \
  --version ${CHAINCODE_VERSION} \
  --package-id ${PACKAGE_ID_ORG1} \
  --sequence ${SEQUENCE} \
  --tls \
  --cafile \$ORDERER_CA
"

echo -e "${GREEN}✓ Approved for Org1${NC}"
echo ""

# Step 10: Approve for Org2
echo -e "${YELLOW}Step 10: Approving chaincode for Org2...${NC}"
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org2MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt

peer lifecycle chaincode approveformyorg \
  -o orderer.example.com:7050 \
  --channelID ${CHANNEL_NAME} \
  --name ${CHAINCODE_NAME} \
  --version ${CHAINCODE_VERSION} \
  --package-id ${PACKAGE_ID_ORG2} \
  --sequence ${SEQUENCE} \
  --tls \
  --cafile \$ORDERER_CA
"

echo -e "${GREEN}✓ Approved for Org2${NC}"
echo ""

# Step 11: Commit chaincode
echo -e "${YELLOW}Step 11: Committing chaincode definition...${NC}"
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt

peer lifecycle chaincode commit \
  -o orderer.example.com:7050 \
  --channelID ${CHANNEL_NAME} \
  --name ${CHAINCODE_NAME} \
  --version ${CHAINCODE_VERSION} \
  --sequence ${SEQUENCE} \
  --tls \
  --cafile \$ORDERER_CA \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
"

echo -e "${GREEN}✓ Chaincode committed${NC}"
echo ""

# Step 12: Initialize ledger
echo -e "${YELLOW}Step 12: Initializing ledger...${NC}"
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt

peer chaincode invoke \
  -o orderer.example.com:7050 \
  --tls \
  --cafile \$ORDERER_CA \
  -C ${CHANNEL_NAME} \
  -n ${CHAINCODE_NAME} \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
  -c '{\"function\":\"InitLedger\",\"Args\":[]}'
"

echo -e "${GREEN}✓ Ledger initialized${NC}"
echo ""

# Step 13: Query assets
echo -e "${YELLOW}Step 13: Querying all assets...${NC}"
sleep 3
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer chaincode query -C ${CHANNEL_NAME} -n ${CHAINCODE_NAME} -c '{\"Args\":[\"GetAllAssets\"]}'
"

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}✓ CaaS Deployment Completed Successfully!${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "${YELLOW}Chaincode containers running:${NC}"
docker ps --filter "name=asset-transfer" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
