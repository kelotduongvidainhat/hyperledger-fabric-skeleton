#!/bin/bash
#
# Script to generate crypto material, channel artifacts, and start the network.
#

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_ROOT="/home/qwe/hyperledger-fabric-skeleton"

# 0. Clean previous state
echo -e "${YELLOW}Step 0: Cleaning previous state...${NC}"
cd "${PROJECT_ROOT}/network/docker"
docker-compose down -v --remove-orphans || true
rm -rf "${PROJECT_ROOT}/network/crypto-config"
rm -rf "${PROJECT_ROOT}/network/channel-artifacts"
mkdir -p "${PROJECT_ROOT}/network/crypto-config"
mkdir -p "${PROJECT_ROOT}/network/channel-artifacts"
echo -e "${GREEN}✓ Cleaned${NC}"

# 1. Generate Crypto Material using Fabric CA
echo -e "${YELLOW}Step 1: Generating crypto material via Fabric CA...${NC}"

# Create CA directories
# Create CA directories
mkdir -p "${PROJECT_ROOT}/network/crypto-config/peerOrganizations/org1.example.com/ca"
mkdir -p "${PROJECT_ROOT}/network/crypto-config/ordererOrganizations/example.com/ca"

# Start only the CA containers
cd "${PROJECT_ROOT}/network/docker"
docker-compose up -d ca_org1 ca_orderer

echo "Waiting for CAs to start (5s)..."
sleep 5

# Run the enrollment script
cd "${PROJECT_ROOT}"
./scripts/enroll-identities.sh

# sudo chown -R $USER:$USER "${PROJECT_ROOT}/network/crypto-config"

# 2. Generate Channel Artifacts
echo -e "${YELLOW}Step 2: Generating channel artifacts...${NC}"
export FABRIC_CFG_PATH="${PROJECT_ROOT}/network/config"

echo "Generating genesis block..."
"${PROJECT_ROOT}/bin/configtxgen" -profile OneOrgOrdererGenesis -channelID system-channel -outputBlock "${PROJECT_ROOT}/network/channel-artifacts/genesis.block"

echo "Generating channel creation transaction..."
"${PROJECT_ROOT}/bin/configtxgen" -profile OneOrgChannel -outputCreateChannelTx "${PROJECT_ROOT}/network/channel-artifacts/mychannel.tx" -channelID mychannel

echo "Generating anchor peer updates..."
"${PROJECT_ROOT}/bin/configtxgen" -profile OneOrgChannel -outputAnchorPeersUpdate "${PROJECT_ROOT}/network/channel-artifacts/Org1MSPanchors.tx" -channelID mychannel -asOrg Org1MSP

echo -e "${GREEN}✓ Artifacts generated${NC}"

# 3. Start Network
echo -e "${YELLOW}Step 3: Starting Hyperledger Fabric network...${NC}"
cd "${PROJECT_ROOT}/network/docker"
docker-compose up -d

echo -e "${YELLOW}Waiting for network to be ready (15s)...${NC}"
sleep 15

# 4. Create and Join Channel
echo -e "${YELLOW}Step 4: Creating and joining channel...${NC}"

# Create Channel
docker exec cli bash -c "
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt
peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/mychannel.tx --outputBlock ./channel-artifacts/mychannel.block --tls --cafile \$ORDERER_CA
"

# Join Org1 Peer0
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
peer channel join -b ./channel-artifacts/mychannel.block
"

# Join Org1 Peer1
docker exec cli bash -c "
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer1.org1.example.com:8051
peer channel join -b ./channel-artifacts/mychannel.block
"

# Update Anchor Peers
echo -e "${YELLOW}Updating anchor peers...${NC}"
docker exec cli bash -c "
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
peer channel update -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/Org1MSPanchors.tx --tls --cafile \$ORDERER_CA
"

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}✓ Network Setup Completed Successfully!${NC}"
echo -e "${GREEN}============================================${NC}"
echo "You can now run: ./scripts/deploy-caas-chaincode.sh"
