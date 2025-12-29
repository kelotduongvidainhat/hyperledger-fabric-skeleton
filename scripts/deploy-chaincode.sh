#!/bin/bash

# Asset Transfer Chaincode Deployment Script
# This script packages, installs, approves, and commits the chaincode

set -e

# Colors for output
GREEN='\033[0.32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Asset Transfer Chaincode Deployment${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Configuration
CHAINCODE_NAME="asset-transfer"
CHAINCODE_VERSION="1.0"
CHAINCODE_SEQUENCE="1"
CHANNEL_NAME="mychannel"
CHAINCODE_PATH="../chaincode/asset-transfer"
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Step 1: Package the chaincode
echo -e "${YELLOW}Step 1: Packaging chaincode...${NC}"
cd /opt/gopath/src/github.com/chaincode
peer lifecycle chaincode package ${CHAINCODE_NAME}.tar.gz \
    --path ${CHAINCODE_PATH} \
    --lang golang \
    --label ${CHAINCODE_NAME}_${CHAINCODE_VERSION}

echo -e "${GREEN}✓ Chaincode packaged successfully${NC}"
echo ""

# Step 2: Install on Org1 Peer
echo -e "${YELLOW}Step 2: Installing chaincode on Org1 peer...${NC}"
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode install ${CHAINCODE_NAME}.tar.gz

echo -e "${GREEN}✓ Chaincode installed on Org1 peer${NC}"
echo ""

# Step 3: Install on Org2 Peer
echo -e "${YELLOW}Step 3: Installing chaincode on Org2 peer...${NC}"
export CORE_PEER_LOCALMSPID=Org2MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

peer lifecycle chaincode install ${CHAINCODE_NAME}.tar.gz

echo -e "${GREEN}✓ Chaincode installed on Org2 peer${NC}"
echo ""

# Step 4: Query installed chaincode to get package ID
echo -e "${YELLOW}Step 4: Querying installed chaincode...${NC}"
peer lifecycle chaincode queryinstalled > installed.txt
PACKAGE_ID=$(sed -n "/${CHAINCODE_NAME}_${CHAINCODE_VERSION}/{s/^Package ID: //; s/, Label:.*$//; p;}" installed.txt)
echo "Package ID: ${PACKAGE_ID}"
echo -e "${GREEN}✓ Package ID retrieved${NC}"
echo ""

# Step 5: Approve for Org2
echo -e "${YELLOW}Step 5: Approving chaincode for Org2...${NC}"
peer lifecycle chaincode approveformyorg \
    -o orderer.example.com:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --channelID ${CHANNEL_NAME} \
    --name ${CHAINCODE_NAME} \
    --version ${CHAINCODE_VERSION} \
    --package-id ${PACKAGE_ID} \
    --sequence ${CHAINCODE_SEQUENCE} \
    --tls \
    --cafile $ORDERER_CA

echo -e "${GREEN}✓ Chaincode approved for Org2${NC}"
echo ""

# Step 6: Approve for Org1
echo -e "${YELLOW}Step 6: Approving chaincode for Org1...${NC}"
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode approveformyorg \
    -o orderer.example.com:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --channelID ${CHANNEL_NAME} \
    --name ${CHAINCODE_NAME} \
    --version ${CHAINCODE_VERSION} \
    --package-id ${PACKAGE_ID} \
    --sequence ${CHAINCODE_SEQUENCE} \
    --tls \
    --cafile $ORDERER_CA

echo -e "${GREEN}✓ Chaincode approved for Org1${NC}"
echo ""

# Step 7: Check commit readiness
echo -e "${YELLOW}Step 7: Checking commit readiness...${NC}"
peer lifecycle chaincode checkcommitreadiness \
    --channelID ${CHANNEL_NAME} \
    --name ${CHAINCODE_NAME} \
    --version ${CHAINCODE_VERSION} \
    --sequence ${CHAINCODE_SEQUENCE} \
    --tls \
    --cafile $ORDERER_CA \
    --output json

echo -e "${GREEN}✓ Commit readiness checked${NC}"
echo ""

# Step 8: Commit chaincode definition
echo -e "${YELLOW}Step 8: Committing chaincode definition...${NC}"
peer lifecycle chaincode commit \
    -o orderer.example.com:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --channelID ${CHANNEL_NAME} \
    --name ${CHAINCODE_NAME} \
    --version ${CHAINCODE_VERSION} \
    --sequence ${CHAINCODE_SEQUENCE} \
    --tls \
    --cafile $ORDERER_CA \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

echo -e "${GREEN}✓ Chaincode committed successfully${NC}"
echo ""

# Step 9: Query committed chaincode
echo -e "${YELLOW}Step 9: Querying committed chaincode...${NC}"
peer lifecycle chaincode querycommitted --channelID ${CHANNEL_NAME} --name ${CHAINCODE_NAME}

echo -e "${GREEN}✓ Chaincode query completed${NC}"
echo ""

# Step 10: Initialize the ledger
echo -e "${YELLOW}Step 10: Initializing ledger...${NC}"
peer chaincode invoke \
    -o orderer.example.com:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --tls \
    --cafile $ORDERER_CA \
    -C ${CHANNEL_NAME} \
    -n ${CHAINCODE_NAME} \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    -c '{"function":"InitLedger","Args":[]}'

echo -e "${GREEN}✓ Ledger initialized${NC}"
echo ""

# Step 11: Query all assets
echo -e "${YELLOW}Step 11: Querying all assets...${NC}"
sleep 3
peer chaincode query -C ${CHANNEL_NAME} -n ${CHAINCODE_NAME} -c '{"Args":["GetAllAssets"]}'

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ Chaincode deployment completed!${NC}"
echo -e "${GREEN}========================================${NC}"
