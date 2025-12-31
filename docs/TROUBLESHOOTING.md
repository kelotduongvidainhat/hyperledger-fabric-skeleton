# Troubleshooting Guide - Hyperledger Fabric Chaincode Deployment

This guide covers common issues encountered during chaincode deployment and their solutions.

---

## Table of Contents

1. [Backend Access Denied](#backend-access-denied)
2. [TLS Certificate Verification Failed](#tls-certificate-verification-failed)
3. [Endorsement Policy Failures](#endorsement-policy-failures)
4. [Version Compatibility & Timeouts](#version-compatibility--timeouts)
5. [Chaincode Connection Issues](#chaincode-connection-issues)
6. [Package ID Issues](#package-id-issues)
7. [Network Issues](#network-issues)
8. [CouchDB Issues](#couchdb-issues)
9. [CA Issues](#ca-issues)

---

## 1. Backend Access Denied

### ❌ Error: access denied: channel [mychannel] creator org [Org1MSP]

**Error Message:**
```
failed to evaluate transaction: rpc error: code = FailedPrecondition desc = evaluate call to endorser returned error: error validating proposal: access denied: channel [mychannel] creator org [Org1MSP]
```

**Cause:**
The identity attempting the transaction does not have the required permissions (Role/OU) in the channel policy. This can be caused by one of the following 5 issues:

1.  **NodeOU Configuration Mismatch (Missing `OrdererOUIdentifier`)**: The channel requires specific OUs for validation. If `config.yaml` enables NodeOUs but the **Orderer MSP** is missing the `OrdererOUIdentifier` block, the Orderer nodes cannot validate their role, causing `permission denied` during System Channel writes (e.g., creating a channel). **Note:** The **Org MSP** should *not* have `OrdererOUIdentifier`. 
2.  **Identity Type Mismatch**: The backend is using an identity registered as `admin` but the policy strictly requires `client` (or vice versa).
3.  **Channel Policy Definition Config**: The `configtx.yaml` might incorrectly restrict Writers to only one role type that doesn't match your identity.
4.  **Signature/SDK Issue**: The Fabric SDK might be calculating the proposal signature differently than the Peer expects (version mismatch).
5.  **Certificate Chain Split-Brain**: The Peer trusts Root CA "A" but the user's cert is signed by Root CA "B" (due to persistent volume issues during reset).

**Solution Checklist:**
1. **Verify User Enrollment:** Ensure the user was enrolled with the correct type (`client` or `admin`).
   ```bash
   fabric-ca-client register --id.type client ...
   fabric-ca-client enroll ...
   ```
2. **Check Certificate OUs:**
   ```bash
   openssl x509 -in path/to/cert.pem -text -noout | grep "OU="
   ```
   Must show `OU=client` or `OU=admin`.
3. **Verify MSP Config:** Ensure `config.yaml` exists in the user's MSP folder enabling NodeOUs.
4. **Resync Keys:** If Cause #5, perform a full volume prune and fresh start.

---

## 2. TLS Certificate Verification Failed

### ❌ Error: x509: certificate signed by unknown authority

**Error Message:**
```
rpc error: code = Unavailable desc = connection error: desc = "transport: authentication handshake failed: tls: failed to verify certificate: x509: certificate signed by unknown authority"
```

**Cause:**
1.  **Split-Brain CA Issue**: The Fabric CA container re-used an old anonymous volume (database) while the disk files were regenerated, leading to key mismatch.
2.  **CA Name Collision**: Multiple CAs (e.g., `ca_org1` and `ca_orderer`) defaulted to the same Common Name (`fabric-ca-server`) for their Root Cert. This confuses the Go client's TLS verification (key mismatch against wrong Root with same name).
3.  **Zombie Backend Process**: An old instance of the backend process is still running (holding port 8080) with STALE CA certificates loaded in memory. New requests hit the old process, which rejects the new Peer certificates.

**Solution:**
1.  **Set Distinct CA Names**: Ensure `docker-compose.yaml` sets `FABRIC_CA_SERVER_CSR_CN` to unique values (e.g., `ca-org1`, `ca-orderer`).
2.  **Kill Zombie Processes**:
    ```bash
    fuser -k 8080/tcp  # Kills process on port 8080
    ```
3.  **Full Reset**: Use `fresh-start.sh` which performs `docker volume prune -af` to ensure no stale CA volumes persist.

**Solution:**
You must remove the anonymous CA volumes.

**Manual Fix:**
```bash
# 1. Stop CA containers
docker stop ca_org1 ca_orderer
docker rm -f ca_org1 ca_orderer

# 2. Prune all volumes (including anonymous ones)
docker volume prune -af
```

**Automated Fix:**
Use the improved `./scripts/fresh-start.sh` which handles this automatically.

---

## 3. Endorsement Policy Failures

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
  --tlsRootCertFiles ... \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles ... \
  -c '{"function":"CreateAsset","Args":...}'
```

---

## 4. Version Compatibility & Timeouts

### ❌ Issue: Connection Timeouts or Intermittent Failures

**Symptoms:**
- `context deadline exceeded` errors during peer or orderer commands.
- `peer chaincode invoke` hanging indefinitely or timing out.

**Cause:**
Version mismatch between local binaries (cli/bin) and Docker images.

**Solution:**
Ensure all components are aligned to the same minor version (e.g., all 2.5.x).
- Check binary version: `peer version`
- Check image version: `docker inspect hyperledger/fabric-peer:latest | grep Version`

---

## 5. Chaincode Connection Issues

### ❌ Error: Peer will not accept external chaincode connection

**Error Message:**
```
peer will not accept external chaincode connection (except in dev mode)
```

**Cause:**
The `CHAINCODE_ID` environment variable in the chaincode container has the wrong format.

**Solution:**
The CHAINCODE_ID should be the **full package ID** (label:hash):
```yaml
environment:
  - CHAINCODE_ID=asset-transfer_1.0:72532aa0858d4c2d11fe7355936842ce76118efb89ad23f52200c9938d4e8d6f
```

---

## 6. Package ID Issues

### ❌ Error: Package ID not found

**Solution:**
Get the correct package ID:
```bash
docker exec cli peer lifecycle chaincode queryinstalled
```

---

## 7. Network Issues

### ❌ Error: Connection refused

**Solution:**
1. Check containers: `docker ps`
2. Restart peer: `docker-compose restart peer0.org2.example.com`

---

## 8. CouchDB Issues

### ❌ Error: Unable to connect to CouchDB

**Solution:**
1. Check running: `docker ps | grep couchdb`
2. Check logs: `docker logs couchdb0`

---

## 9. CA Issues

### ❌ Error: Configuration file version mismatch

**Cause:**
CA configuration file version doesn't match the CA server version.

**Solution:**
Update the configuration file version in `fabric-ca-server-config.yaml` to match the image version (e.g., `v1.5.15`).

---

## Getting Help

### Check System Status
```bash
docker ps
df -h
docker system df
```

### Collect Logs
```bash
docker logs peer0.org1.example.com > peer.log 2>&1
docker logs asset-transfer-org1 > web-chaincode.log 2>&1
```

---

**Last Updated**: 2025-12-31 14:00:00 +07:00
