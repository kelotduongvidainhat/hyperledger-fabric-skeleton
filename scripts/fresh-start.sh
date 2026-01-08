#!/bin/bash
# fresh-start.sh - Automated Hyperledger Fabric Network Setup
# Recreates the network from scratch.

set -e

# export PATH to include local binaries
export PATH=${PWD}/bin:$PATH


# 1. Check/Download Binaries
if [ ! -d "bin" ]; then
    echo "⚠️ Binaries missing. Running download-binaries.sh..."
    ./scripts/download-binaries.sh
fi

echo "🚀 Starting Fresh Network Setup..."

# 2. Cleanup
echo "🧹 Cleaning up previous artifacts..."
# Remove containers
if [ -f "network/docker-compose.yaml" ]; then
    docker-compose -f network/docker-compose.yaml down --volumes --remove-orphans || true
fi
# Remove crypto-config, channel-artifacts, and local wallet
rm -rf network/crypto-config network/channel-artifacts backend/wallet

# 3. Enter Network Directory
cd network
mkdir -p channel-artifacts

# 4. Generate Crypto Materials
echo "🔐 Generating crypto materials..."
cryptogen generate --config=crypto-config.yaml --output="crypto-config"

# 5. Generate Channel Artifacts
echo "📝 Generating channel artifacts..."
export FABRIC_CFG_PATH=${PWD}

# Generate Genesis Block for mychannel (includes Orderer config now)
configtxgen -profile TwoOrgsChannel -outputBlock ./channel-artifacts/mychannel.block -channelID mychannel

# 6. Start Network
echo "🐳 Starting Docker containers (including Backend/Frontend)..."
docker-compose up -d --build

# Return to root
cd ..

# Wait for containers to be ready
echo "⏳ Waiting for network to stabilize (10s)..."
sleep 10

# 7. Join Orderer to Channel (osnadmin)
echo "📢 Joining Orderer to 'mychannel'..."
docker exec cli osnadmin channel join \
    --channelID mychannel \
    --config-block /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/mychannel.block \
    -o orderer.example.com:7053 \
    --ca-file /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    --client-cert /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt \
    --client-key /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key

# 8. Join Peers to Channel
echo "🔗 Joining peers to channel..."

# Join Org1 Peer
echo "  > Joining Peer0 Org1..."
docker exec cli bash -c '
    export CORE_PEER_LOCALMSPID=Org1MSP
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
    
    peer channel join -b ./channel-artifacts/mychannel.block
'

# Join Org2 Peer
echo "  > Joining Peer0 Org2..."
docker exec cli bash -c '
    export CORE_PEER_LOCALMSPID=Org2MSP
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
    
    peer channel join -b ./channel-artifacts/mychannel.block
'

echo "✅ Network Setup Complete! Channel 'mychannel' created and peers joined via osnadmin."
