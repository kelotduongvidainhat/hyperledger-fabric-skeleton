# Chaincode Deployment Checklist

Use this checklist to track your progress through the deployment process.

---

## Pre-Deployment Checks

- [ ] All Docker containers are running
  ```bash
  docker ps | wc -l  # Should show 10 containers (including header)
  ```

- [ ] Channel exists and peers are joined
  ```bash
  docker exec cli peer channel list  # Should show: mychannel
  ```

- [ ] Chaincode Docker image is built
  ```bash
  docker images | grep asset-transfer  # Should show: asset-transfer:1.0
  ```

---

## Deployment Steps

### Step 1: Build & Package
- [ ] Navigate to chaincode directory
  ```bash
  cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer
  ```

- [ ] Docker image built
  ```bash
  docker build -t asset-transfer:1.0 .
  ```

- [ ] CaaS package created
  ```bash
  # Create directories and files
  mkdir -p caas-pkg
  # Create connection.json and metadata.json
  # Package with tar
  ```

- [ ] Package file exists
  ```bash
  ls -lh asset-transfer.tar.gz
  ```

### Step 2: Install Chaincode
- [ ] Installed on Org1 peer
  ```bash
  docker exec cli peer lifecycle chaincode install ...
  ```

- [ ] Installed on Org2 peer
  ```bash
  docker exec cli peer lifecycle chaincode install ...
  ```

- [ ] Package ID retrieved and saved
  ```bash
  export PACKAGE_ID="asset-transfer_1.0:..."
  echo $PACKAGE_ID  # Verify it's set
  ```

### Step 3: Start Chaincode Containers
- [ ] Environment variables set
  ```bash
  export CHAINCODE_ID_ORG1=$PACKAGE_ID
  export CHAINCODE_ID_ORG2=$PACKAGE_ID
  ```

- [ ] Containers started
  ```bash
  docker-compose up -d
  ```

- [ ] Containers are running
  ```bash
  docker ps | grep asset-transfer  # Should show 2 containers
  ```

- [ ] Waited 5-10 seconds for containers to be ready

### Step 4: Approve Chaincode
- [ ] Approved for Org1
  ```bash
  docker exec cli peer lifecycle chaincode approveformyorg ...
  ```

- [ ] Approved for Org2
  ```bash
  docker exec cli peer lifecycle chaincode approveformyorg ...
  ```

- [ ] Commit readiness checked
  ```bash
  docker exec cli peer lifecycle chaincode checkcommitreadiness ...
  # Should show: Org1MSP: true, Org2MSP: true
  ```

### Step 5: Commit Chaincode
- [ ] Chaincode definition committed
  ```bash
  docker exec cli peer lifecycle chaincode commit ...
  ```

- [ ] Commitment verified
  ```bash
  docker exec cli peer lifecycle chaincode querycommitted --channelID mychannel --name asset-transfer
  ```

### Step 6: Initialize & Test
- [ ] Ledger initialized
  ```bash
  docker exec cli peer chaincode invoke ... -c '{"function":"InitLedger","Args":[]}'
  ```

- [ ] Waited 3-5 seconds for transaction to commit

- [ ] Assets queried successfully
  ```bash
  docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{"Args":["GetAllAssets"]}'
  # Should return 6 assets
  ```

---

## Post-Deployment Verification

- [ ] Create new asset works
  ```bash
  # CreateAsset command
  ```

- [ ] Read asset works
  ```bash
  # ReadAsset command
  ```

- [ ] Transfer asset works
  ```bash
  # TransferAsset command
  ```

- [ ] Query returns updated data
  ```bash
  # GetAllAssets shows new asset
  ```

---

## Troubleshooting (If Needed)

- [ ] Checked chaincode container logs
  ```bash
  docker logs asset-transfer-org1
  docker logs asset-transfer-org2
  ```

- [ ] Checked peer logs
  ```bash
  docker logs peer0.org1.example.com --tail 50
  docker logs peer0.org2.example.com --tail 50
  ```

- [ ] Restarted chaincode containers if needed
  ```bash
  docker-compose restart
  ```

---

## Success Indicators

✅ **Deployment Successful** when:
1. All checkboxes above are checked
2. `GetAllAssets` returns 6 assets
3. Can create, read, and transfer assets
4. No errors in container logs

---

## Notes & Package ID

**My Package ID:**
```
_____________________________________________________________
```

**Deployment Date:** __________

**Issues Encountered:**
```
_____________________________________________________________
_____________________________________________________________
_____________________________________________________________
```

**Resolution:**
```
_____________________________________________________________
_____________________________________________________________
_____________________________________________________________
```

---

## Quick Commands Reference

**Get Package ID:**
```bash
docker exec cli peer lifecycle chaincode queryinstalled | grep asset-transfer
```

**Check Containers:**
```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

**View Logs:**
```bash
docker logs asset-transfer-org1 --tail 20
```

**Restart Chaincode:**
```bash
cd /home/qwe/hyperledger-fabric-skeleton/chaincode/asset-transfer
docker-compose restart
```

---

**For detailed instructions, see:**
- `DEPLOYMENT_GUIDE.md` - Full step-by-step guide
- `QUICK_REFERENCE.md` - Quick command reference
