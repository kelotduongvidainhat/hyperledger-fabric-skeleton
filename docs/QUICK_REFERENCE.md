# Quick Reference - Chaincode Deployment

## Environment Setup

```bash
# Set Package ID (replace with your actual ID from step 5)
export PACKAGE_ID="asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f"
```

## One-Command Deployment (After Building Image and Package)

```bash
# Start chaincode containers
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer
export CHAINCODE_ID_ORG1=$PACKAGE_ID
export CHAINCODE_ID_ORG2=$PACKAGE_ID
docker-compose up -d

# Approve for Org1
docker exec cli peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --channelID mychannel --name asset-transfer --version 1.0 --package-id $PACKAGE_ID --sequence 1 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Switch to Org2 and approve
docker exec cli bash -c 'export CORE_PEER_LOCALMSPID=Org2MSP && export CORE_PEER_ADDRESS=peer0.org2.example.com:9051 && export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt && export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp && peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --channelID mychannel --name asset-transfer --version 1.0 --package-id '"$PACKAGE_ID"' --sequence 1 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem'

# Commit
docker exec cli peer lifecycle chaincode commit -o orderer.example.com:7050 --channelID mychannel --name asset-transfer --version 1.0 --sequence 1 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

# Initialize
docker exec cli peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n asset-transfer --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"InitLedger","Args":[]}'

# Query
sleep 3
docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{"Args":["GetAllAssets"]}'
```

## Common Commands

### Query Installed Chaincode
```bash
docker exec cli peer lifecycle chaincode queryinstalled
```

### Query Committed Chaincode
```bash
docker exec cli peer lifecycle chaincode querycommitted --channelID mychannel
```

### Check Container Status
```bash
docker ps | grep -E "asset-transfer|peer|orderer"
```

### View Chaincode Logs
```bash
docker logs asset-transfer-org1 -f
docker logs asset-transfer-org2 -f
```

### Restart Chaincode
```bash
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer
docker-compose restart
```

## Test Transactions

### Create Asset
```bash
docker exec cli peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n asset-transfer --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt -c '{"function":"CreateAsset","Args":["asset99","gold","25","TestUser","5000"]}'
```

### Read Asset
```bash
docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{"Args":["ReadAsset","asset99"]}'
```

### Transfer Asset
```bash
docker exec cli peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n asset-transfer --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt -c '{"function":"TransferAsset","Args":["asset99","NewOwner"]}'
```

### Get All Assets
```bash
docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{"Args":["GetAllAssets"]}'
```
