#!/bin/bash
set -e

# Environment Variables
export PATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/bin:$PATH
export FABRIC_CFG_PATH=/etc/hyperledger/fabric
CHANNEL_NAME="mychannel"

echo "⚓ Updating Anchor Peers..."

# 1. Fetch Channel Config
echo "Fetching channel config..."
peer channel fetch config config_block.pb -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com -c ${CHANNEL_NAME} --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# 2. Decode Config
echo "Decoding config..."
configtxlator proto_decode --input config_block.pb --type common.Block | jq .data.data[0].payload.data.config > config.json

# 3. Modify Config for Org1
echo "Modifying config for Org1..."
jq '.channel_group.groups.Application.groups.Org1MSP.values += {"AnchorPeers":{"mod_policy": "Admins","value":{"anchor_peers": [{"host": "peer0.org1.example.com","port": 7051}]}}}' config.json > modified_config.json

# 4. Compute Update
configtxlator proto_encode --input config.json --type common.Config > original_config.pb
configtxlator proto_encode --input modified_config.json --type common.Config > modified_config.pb
configtxlator compute_update --channel_id ${CHANNEL_NAME} --original original_config.pb --updated modified_config.pb > config_update.pb
configtxlator proto_decode --input config_update.pb --type common.ConfigUpdate > config_update.json
echo '{"payload":{"header":{"channel_header":{"channel_id":"'"${CHANNEL_NAME}"'", "type":2}},"data":{"config_update":'$(cat config_update.json)'}}}' | jq . > config_update_in_envelope.json
configtxlator proto_encode --input config_update_in_envelope.json --type common.Envelope > anchor_update_org1.pb

# 5. Sign and Submit for Org1
echo "Signing and submitting for Org1..."
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer channel update -f anchor_update_org1.pb -c ${CHANNEL_NAME} -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# 6. Modify Config for Org2 (Fetch fresh config first to avoid version conflict)
echo "Fetching fresh config for Org2 update..."
peer channel fetch config config_block_org2.pb -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com -c ${CHANNEL_NAME} --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

configtxlator proto_decode --input config_block_org2.pb --type common.Block | jq .data.data[0].payload.data.config > config_org2.json

echo "Modifying config for Org2..."
jq '.channel_group.groups.Application.groups.Org2MSP.values += {"AnchorPeers":{"mod_policy": "Admins","value":{"anchor_peers": [{"host": "peer0.org2.example.com","port": 9051}]}}}' config_org2.json > modified_config_org2.json

configtxlator proto_encode --input config_org2.json --type common.Config > original_config_org2.pb
configtxlator proto_encode --input modified_config_org2.json --type common.Config > modified_config_org2.pb
configtxlator compute_update --channel_id ${CHANNEL_NAME} --original original_config_org2.pb --updated modified_config_org2.pb > config_update_org2.pb
configtxlator proto_decode --input config_update_org2.pb --type common.ConfigUpdate > config_update_org2.json
echo '{"payload":{"header":{"channel_header":{"channel_id":"'"${CHANNEL_NAME}"'", "type":2}},"data":{"config_update":'$(cat config_update_org2.json)'}}}' | jq . > config_update_in_envelope_org2.json
configtxlator proto_encode --input config_update_in_envelope_org2.json --type common.Envelope > anchor_update_org2.pb

# 7. Sign and Submit for Org2
echo "Signing and submitting for Org2..."
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

peer channel update -f anchor_update_org2.pb -c ${CHANNEL_NAME} -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

echo "✅ Anchor Peers Updated Successfully!"
