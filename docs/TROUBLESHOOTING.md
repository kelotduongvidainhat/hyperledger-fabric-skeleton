# Troubleshooting Guide - Hyperledger Fabric Chaincode Deployment

This guide covers common issues encountered during chaincode deployment and their solutions.

---

## Table of Contents

1. [Endorsement Policy Failures](#endorsement-policy-failures)
2. [Version Compatibility & Timeouts](#version-compatibility--timeouts)
3. [Chaincode Connection Issues](#chaincode-connection-issues)
4. [Package ID Issues](#package-id-issues)
5. [Network Issues](#network-issues)
6. [CouchDB Issues](#couchdb-issues)
7. [CA Issues](#ca-issues)

---

## 1. Endorsement Policy Failures

### ❌ Error: ENDORSEMENT_POLICY_FAILURE

**Error Message:**
```
ENDORSEMENT_POLICY_FAILURE
validation of endorsement policy for chaincode asset-transfer in tx X:0 failed: 
implicit policy evaluation failed - 1 sub-policies were satisfied, 
but this policy requires 2 of the 'Endorsement' sub-policies to be satisfied
```

**Cause:**
The chaincode uses the default endorsement policy which requires **MAJORITY** of organizations to endorse transactions. For a 2-organization network (Org1 + Org2), this means **BOTH organizations must endorse**.

**What Happened:**
You invoked a transaction with only one peer:
```bash
# ❌ WRONG - Only Org1 peer
peer chaincode invoke ... \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /path/to/org1/tls/ca.crt
```

**Solution:**
Always include **BOTH peers** when invoking transactions:

```bash
# ✅ CORRECT - Both Org1 and Org2 peers
peer chaincode invoke \
  -o orderer.example.com:7050 \
  --tls --cafile $ORDERER_CA \
  -C mychannel \
  -n asset-transfer \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
  -c '{"function":"CreateAsset","Args":["id","color","size","owner","value"]}'
```

**Important Notes:**
- **Query operations** (read-only) only need ONE peer
- **Invoke operations** (write to ledger) need BOTH peers
- The order of `--peerAddresses` and `--tlsRootCertFiles` must match

**How to Check Endorsement Policy:**
```bash
docker exec cli peer lifecycle chaincode querycommitted --channelID mychannel --name asset-transfer
```

Look for the endorsement policy in the output.

---

## 2. Version Compatibility & Timeouts

### ❌ Issue: Connection Timeouts or Intermittent Failures

**Symptoms:**
- `context deadline exceeded` errors during peer or orderer commands.
- `peer chaincode invoke` hanging indefinitely or timing out.
- CLI commands hanging or timing out while communicating with the network.
- "Endorser client failed to connect" even when containers are up.

**Cause:**
This is often caused by a **version mismatch between the local binaries (cli/bin) and the Docker images**. 

**Key Learning:**
If the binary client version (e.g., `peer`, `osnadmin`, `fabric-ca-client`) does not match the Fabric images running in the network, protocol differences can lead to unexpected **timeout errors** instead of clear version mismatch errors. This is particularly common during `peer chaincode invoke` or `peer lifecycle chaincode` operations, where the client waits for a response that never comes or is misinterpreted due to protocol drift.

**Solution:**
Ensure all components are aligned to the same minor version (e.g., all 2.5.x).
- Check binary version: `peer version`
- Check image version: `docker inspect hyperledger/fabric-peer:latest | grep Version`

### 💻 Go Version Requirements

**Important:**
The Go version used to compile chaincode must be compatible with the Fabric version.
- For **Fabric 2.5.x**, it is recommended to use **Go 1.20 or newer** (we are using Go 1.25.2).
- Using an outdated Go version can cause compilation errors during `peer lifecycle chaincode install` or runtime crashes in the chaincode container.

---

## 3. Chaincode Connection Issues

### ❌ Error: Peer will not accept external chaincode connection

**Error Message:**
```
failed to start asset-transfer_1.0:asset-transfer_1.0:72532aa... 
peer will not accept external chaincode connection (except in dev mode)
```

**Cause:**
The `CHAINCODE_ID` environment variable in the chaincode container has the wrong format.

**What Happened:**
The CHAINCODE_ID was duplicated:
```
CHAINCODE_ID=asset-transfer_1.0:asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f
```

**Solution:**
The CHAINCODE_ID should be the **full package ID** (label:hash):

```yaml
# ✅ CORRECT format in docker-compose.yaml
environment:
  - CHAINCODE_ID=asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f
  - CORE_CHAINCODE_ID_NAME=asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f
```

**Verify the fix:**
```bash
# Check the environment variable
docker exec asset-transfer-org1 env | grep CHAINCODE_ID

# Should show:
# CHAINCODE_ID=asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f
```

**Restart containers after fixing:**
```bash
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer
docker-compose down
docker-compose up -d
```

---

## 4. Package ID Issues

### ❌ Error: Package ID not found

**Error Message:**
```
Error: chaincode install failed with status: 500
```

**Solution:**
Get the correct package ID:

```bash
docker exec cli bash -c '
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
peer lifecycle chaincode queryinstalled
'
```

Copy the Package ID from the output:
```
Package ID: asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f
```

---

## 5. Network Issues

### ❌ Error: Connection refused

**Error Message:**
```
failed to create new connection: connection error: desc = "transport: error while dialing: dial tcp: lookup peer0.org2.example.com on 127.0.0.11:53: no such host"
```

**Cause:**
Container is not running or network connectivity issue.

**Solution:**

1. **Check all containers are running:**
```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

2. **Check if peer is running:**
```bash
docker ps | grep peer0.org2
```

3. **Restart the peer if needed:**
```bash
cd /home/qwe/hyperledger-fabric-skeleton/network/docker
docker-compose restart peer0.org2.example.com
```

4. **Check peer logs:**
```bash
docker logs peer0.org2.example.com --tail 50
```

---

### ❌ Error: TLS handshake failed

**Error Message:**
```
Client TLS handshake failed: tls: failed to verify certificate: x509: certificate signed by unknown authority
```

**Cause:**
Incorrect TLS certificate path or certificate mismatch.

**Solution:**

1. **Verify certificate paths:**
```bash
# Check if certificate exists
docker exec cli ls -la /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
```

2. **Use correct environment variables:**
```bash
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
```

---

## 6. CouchDB Issues

### ❌ Error: Unable to connect to CouchDB

**Error Message:**
```
unable to connect to CouchDB, check the hostname and port
```

**Solution:**

1. **Check CouchDB is running:**
```bash
docker ps | grep couchdb
```

2. **Check CouchDB logs:**
```bash
docker logs couchdb0 --tail 30
docker logs couchdb1 --tail 30
```

3. **Restart CouchDB:**
```bash
cd /home/qwe/hyperledger-fabric-skeleton/network/docker
docker-compose restart couchdb0 couchdb1
```

4. **Wait for CouchDB to be ready, then restart peer:**
```bash
sleep 5
docker-compose restart peer0.org1.example.com peer0.org2.example.com
```

---

## 7. CA Issues

### ❌ Error: Configuration file version mismatch

**Error Message:**
```
Error: Configuration file version 'v1.5.16' is higher than server version 'v1.5.15'
```

**Cause:**
CA configuration file version doesn't match the CA server version.

**Solution:**

1. **Update the configuration file version:**
```bash
sed -i 's/^version: v1.5.16/version: v1.5.15/' \
  network/crypto-config/ordererOrganizations/example.com/ca/fabric-ca-server-config.yaml

sed -i 's/^version: v1.5.16/version: v1.5.15/' \
  network/crypto-config/peerOrganizations/org1.example.com/ca/fabric-ca-server-config.yaml

sed -i 's/^version: v1.5.16/version: v1.5.15/' \
  network/crypto-config/peerOrganizations/org2.example.com/ca/fabric-ca-server-config.yaml
```

2. **Restart CA containers:**
```bash
cd /home/qwe/hyperledger-fabric-skeleton/network/docker
docker-compose restart ca_orderer ca_org1 ca_org2
```

---

## Common Commands for Debugging

### Check Container Status
```bash
docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

### View Logs
```bash
# Peer logs
docker logs peer0.org1.example.com --tail 50
docker logs peer0.org2.example.com --tail 50

# Orderer logs
docker logs orderer.example.com --tail 50

# Chaincode logs
docker logs asset-transfer-org1 --tail 50
docker logs asset-transfer-org2 --tail 50

# CouchDB logs
docker logs couchdb0 --tail 30
docker logs couchdb1 --tail 30
```

### Follow Logs in Real-time
```bash
docker logs -f peer0.org1.example.com
```

### Check Chaincode Status
```bash
# Query installed chaincode
docker exec cli peer lifecycle chaincode queryinstalled

# Query committed chaincode
docker exec cli peer lifecycle chaincode querycommitted --channelID mychannel

# Check commit readiness
docker exec cli peer lifecycle chaincode checkcommitreadiness \
  --channelID mychannel \
  --name asset-transfer \
  --version 1.0 \
  --sequence 1
```

### Restart Everything
```bash
# Stop all containers
cd /home/qwe/hyperledger-fabric-skeleton/network/docker
docker-compose down

# Stop chaincode containers
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer
docker-compose down

# Start network
cd /home/qwe/hyperledger-fabric-skeleton/network/docker
docker-compose up -d

# Wait for network to be ready
sleep 10

# Start chaincode containers
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer
docker-compose up -d
```

---

## Quick Fixes

### Reset Chaincode
If chaincode is misbehaving, restart the containers:

```bash
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer
docker-compose restart
```

### Reset Entire Network
**⚠️ WARNING: This will delete all data!**

```bash
cd /home/qwe/hyperledger-fabric-skeleton/network/docker
docker-compose down -v
docker-compose up -d

# Rejoin channel
# Redeploy chaincode
```

---

## Getting Help

### Check System Status
```bash
# All containers
docker ps

# Disk space
df -h

# Memory usage
free -h

# Docker system info
docker system df
```

### Collect Logs for Support
```bash
# Save all logs
mkdir -p /tmp/fabric-logs
docker logs peer0.org1.example.com > /tmp/fabric-logs/peer0-org1.log 2>&1
docker logs peer0.org2.example.com > /tmp/fabric-logs/peer0-org2.log 2>&1
docker logs orderer.example.com > /tmp/fabric-logs/orderer.log 2>&1
docker logs asset-transfer-org1 > /tmp/fabric-logs/chaincode-org1.log 2>&1
docker logs asset-transfer-org2 > /tmp/fabric-logs/chaincode-org2.log 2>&1

# Create archive
tar -czf fabric-logs.tar.gz /tmp/fabric-logs/
```

---

## Summary of Key Points

✅ **Invoke transactions require BOTH peer endorsements**  
✅ **Query transactions only need ONE peer**  
✅ **CHAINCODE_ID must be the full package ID (label:hash)**  
✅ **Always check container logs when debugging**  
✅ **Restart containers after configuration changes**  
✅ **Wait 3-5 seconds after invoke before querying**  

---

**Last Updated**: 2025-12-29 11:32:00 +07:00
