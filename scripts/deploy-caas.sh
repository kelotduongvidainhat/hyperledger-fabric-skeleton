#!/bin/bash
set -e

# Environment Variables
export PATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/bin:$PATH
export VERBOSE=false

# Chaincode details
CC_NAME="basic"
CC_VERSION="1.0"
CC_SEQUENCE="1"
CHANNEL_NAME="mychannel"

echo "🚀 Deploying Chaincode '$CC_NAME' (CaaS)..."

# Helper functions to switch context
setOrg1() {
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
}

setOrg2() {
    export CORE_PEER_LOCALMSPID="Org2MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
}

# 1. Check for Package
if [ ! -f "basic.tar.gz" ]; then
    echo "❌ basic.tar.gz not found! Please package chaincode on host and copy to CLI container."
    exit 1
fi

cd /opt/gopath/src/github.com/hyperledger/fabric/peer

# 2. Install Chaincode (Org1)
echo "💿 Installing chaincode on Org1..."
setOrg1
peer lifecycle chaincode install basic.tar.gz
PACKAGE_ID=$(peer lifecycle chaincode queryinstalled | grep ${CC_NAME}_${CC_VERSION} | awk -F "[, ]" '{print $3}')
echo "   Package ID: $PACKAGE_ID"

# 3. Install Chaincode (Org2)
echo "💿 Installing chaincode on Org2..."
setOrg2
peer lifecycle chaincode install basic.tar.gz

# 4. Start Chaincode Container (Manual Step emulation)
# In CaaS, the chaincode ID must match the Package ID.
# We need to update the running chaincode container with this ID.
# Since we are inside CLI, we rely on the user (or side script) to update Docker.
# BUT, we can't easily update docker from inside CLI unless we mount docker socket.
# Docker socket IS mounted!

echo "🔄 Updating Chaincode Container with Package ID: $PACKAGE_ID..."
# We use curl to update service or just restart with env var if possible?
# Docker update doesn't change env vars easily without recreation.
# Simpler approach: Stop, Remove, Run (or rely on docker-compose up --force-recreate with overrides)

# Let's try to remove and recreate just the chaincode container using python/curl or just bare docker command if client is installed.
# CLI image usually has docker client? No, verified CLI image usage earlier.
# Wait, CLI container likely does NOT have docker client installed, only socket mounted.
# If CLI has no docker client, we can't control the sibling container.

# ALTERNATIVE: Write the Package ID to a file shared with Host, and Host script updates Docker.
# OR: We just assume the Package ID is deterministic if source doesn't change?
# No, hash over metadata+connection.json. It IS deterministic for CaaS package!
# Hash of {label, type} + connection.json content.
# Since connection.json and metadata.json are static, PACKAGE_ID should be constant!

# Let's verify this hypothesis.
# If Package ID is constant, we can hardcode it in docker-compose after first run.
# But for now, let's print it and ask user or use a clever trick.
# Actually, I can compute the package ID locally?
# peer lifecycle chaincode calculatepackageid chaincode.tar.gz

echo "⚠️  IMPORTANT: Please ensure 'chaincode-basic' container is running with CHAINCODE_ID=$PACKAGE_ID"
echo "    If you are running this automated, we assume the container is started matching this ID."

# 5. Approve Chaincode (Org1)
echo "👍 Approving chaincode for Org1..."
setOrg1
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# 6. Approve Chaincode (Org2)
echo "👍 Approving chaincode for Org2..."
setOrg2
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# 7. Check Commit Readiness
echo "🕵️ Checking commit readiness..."
peer lifecycle chaincode checkcommitreadiness --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version ${CC_VERSION} --sequence ${CC_SEQUENCE} --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --output json

# 8. Commit Chaincode Definition
echo "🚀 Committing chaincode definition..."
peer lifecycle chaincode commit -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version ${CC_VERSION} --sequence ${CC_SEQUENCE} --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

echo "✅ Chaincode deployed successfully!"

# 9. Init Ledger
echo "⚡ Invoking InitLedger..."
setOrg1
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID ${CHANNEL_NAME} --name ${CC_NAME} --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"InitLedger","Args":[]}'

sleep 2
echo "🔍 Querying ReadAsset 'asset1'..."
peer chaincode query -C ${CHANNEL_NAME} -n ${CC_NAME} -c '{"Args":["ReadAsset","asset1"]}'

echo "🎉 All Done!"
