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
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode install /opt/gopath/src/github.com/chaincode/asset-transfer/${CHAINCODE_NAME}.tar.gz
"

echo -e "${GREEN}✓ Installed on Org1${NC}"
echo ""

# Step 6: Install on Org2
echo -e "${YELLOW}Step 6: Installing chaincode on Org2 peer...${NC}"
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org2MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

peer lifecycle chaincode install /opt/gopath/src/github.com/chaincode/asset-transfer/${CHAINCODE_NAME}.tar.gz
"

echo -e "${GREEN}✓ Installed on Org2${NC}"
echo ""

# Step 7: Get package IDs
echo -e "${YELLOW}Step 7: Getting package IDs...${NC}"

PACKAGE_ID_ORG1=$(docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
peer lifecycle chaincode queryinstalled | grep ${CHAINCODE_LABEL} | sed 's/^Package ID: //' | sed 's/, Label:.*$//'
")

PACKAGE_ID_ORG2=$(docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org2MSP
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
peer lifecycle chaincode queryinstalled | grep ${CHAINCODE_LABEL} | sed 's/^Package ID: //' | sed 's/, Label:.*$//'
")

echo "Org1 Package ID: $PACKAGE_ID_ORG1"
echo "Org2 Package ID: $PACKAGE_ID_ORG2"
echo -e "${GREEN}✓ Package IDs retrieved${NC}"
echo ""

# Step 8: Start chaincode containers
echo -e "${YELLOW}Step 8: Starting chaincode containers...${NC}"

# Start chaincode containers
export CHAINCODE_ID_ORG1=$PACKAGE_ID_ORG1
export CHAINCODE_ID_ORG2=$PACKAGE_ID_ORG2

docker-compose up -d

echo -e "${GREEN}✓ Chaincode containers started${NC}"
echo ""

# Wait for containers to be ready
echo -e "${YELLOW}Waiting for chaincode containers to be ready...${NC}"
sleep 5
echo ""

# Step 9: Approve for Org1
echo -e "${YELLOW}Step 9: Approving chaincode for Org1...${NC}"
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode approveformyorg \
  -o orderer.example.com:7050 \
  --channelID ${CHANNEL_NAME} \
  --name ${CHAINCODE_NAME} \
  --version ${CHAINCODE_VERSION} \
  --package-id ${PACKAGE_ID_ORG1} \
  --sequence 1 \
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
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode approveformyorg \
  -o orderer.example.com:7050 \
  --channelID ${CHANNEL_NAME} \
  --name ${CHAINCODE_NAME} \
  --version ${CHAINCODE_VERSION} \
  --package-id ${PACKAGE_ID_ORG2} \
  --sequence 1 \
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
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode commit \
  -o orderer.example.com:7050 \
  --channelID ${CHANNEL_NAME} \
  --name ${CHAINCODE_NAME} \
  --version ${CHAINCODE_VERSION} \
  --sequence 1 \
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
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

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
