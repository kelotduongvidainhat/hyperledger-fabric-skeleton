#!/bin/bash
set -e

# Environment Variables
export PATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/bin:$PATH
export VERBOSE=false
CHANNEL_NAME="mychannel"
CC_NAME="basic"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 Starting Comprehensive Network & Chaincode Test...${NC}"

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

invoke() {
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID ${CHANNEL_NAME} --name ${CC_NAME} --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c "$1" >&log.txt
    res=$?
    cat log.txt
    if [ $res -ne 0 ]; then
        echo -e "${RED}❌ Invoke failed${NC}"
        exit 1
    fi
    echo "   (Waiting for block commit...)"
    sleep 3
}

query() {
    peer chaincode query -C ${CHANNEL_NAME} -n ${CC_NAME} -c "$1"
}

# Generate Unique Asset ID
ASSET_ID="asset_test_$(date +%s)"

# 1. Create Asset (as Org1)
echo "----------------------------------------------------------------"
echo "STEP 1: Creating Asset '$ASSET_ID' as Org1..."
setOrg1
invoke '{"function":"CreateAsset","Args":["'"$ASSET_ID"'", "Test Asset", "Description", "http://ipfs.io/image", "QmHash", "Public"]}'
echo -e "${GREEN}✅ Asset Created${NC}"

# 2. Read Asset
echo "----------------------------------------------------------------"
echo "STEP 2: Reading Asset '$ASSET_ID'..."
query '{"Args":["ReadAsset","'"$ASSET_ID"'"]}'
echo -e "${GREEN}✅ Read Success${NC}"

# 3. Propose Transfer (Org1 -> Org2)
echo "----------------------------------------------------------------"
echo "STEP 3: Proposing Transfer to Org2MSP..."
# Access Control Check: Only Owner (Org1) can propose
invoke '{"function":"ProposeTransfer","Args":["'"$ASSET_ID"'", "Org2MSP"]}'
echo -e "${GREEN}✅ Transfer Proposed${NC}"

# 4. Verify Status is PENDING_TRANSFER
echo "----------------------------------------------------------------"
echo "STEP 4: Verifying Status is PENDING_TRANSFER..."
RESULT=$(query '{"Args":["ReadAsset","'"$ASSET_ID"'"]}')
if [[ $RESULT == *"PENDING_TRANSFER"* ]]; then
    echo -e "${GREEN}✅ Status Correct: PENDING_TRANSFER${NC}"
else
    echo -e "${RED}❌ Status Incorrect: $RESULT${NC}"
    exit 1
fi

# 5. Accept Transfer (as Org2)
echo "----------------------------------------------------------------"
echo "STEP 5: Accepting Transfer as Org2..."
setOrg2
invoke '{"function":"AcceptTransfer","Args":["'"$ASSET_ID"'"]}'
echo -e "${GREEN}✅ Transfer Accepted${NC}"

# 6. Verify New Owner
echo "----------------------------------------------------------------"
echo "STEP 6: Verifying New Owner is Org2MSP..."
RESULT=$(query '{"Args":["ReadAsset","'"$ASSET_ID"'"]}')
if [[ $RESULT == *"Org2MSP"* ]]; then
    echo -e "${GREEN}✅ Owner Correct: Org2MSP${NC}"
else
    echo -e "${RED}❌ Owner Incorrect: $RESULT${NC}"
    exit 1
fi

# 7. Check History
echo "----------------------------------------------------------------"
echo "STEP 7: Checking Asset History..."
query '{"Args":["GetAssetHistory","'"$ASSET_ID"'"]}'
echo -e "${GREEN}✅ History Retrieved${NC}"

echo "----------------------------------------------------------------"
echo -e "${GREEN}🎉 ALL TESTS PASSED! Network and Chaincode are Healthy.${NC}"
