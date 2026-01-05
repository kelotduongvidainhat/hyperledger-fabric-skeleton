# Chaincode-as-a-Service (CaaS) Troubleshooting

## Common Deployment Errors

### 1. `Chaincode Registration Failed: Timeout Expired`

**Error Log:**
```
Error: endorsement failure during invoke. response: status:500 message:"error in simulation: ... could not launch chaincode basic_1.0:<hash>: chaincode registration failed: connection to basic_1.0:<hash> failed: ... context deadline exceeded"
```

**Cause:**
The Peer is unable to connect to the External Chaincode container. This usually happens because:
1.  **Mismatched IDs**: The `CHAINCODE_ID` environment variable in the chaincode container does NOT match the `Package ID` installed on the Peer.
2.  **Network Issue**: The chaincode container is not on the same Docker network (`fabric_test`) as the Peer.
3.  **Container Down**: The chaincode container is not running.

**Solution:**
1.  **Get the Package ID**:
    Run `peer lifecycle chaincode queryinstalled` (inside CLI) to get the correct ID (e.g., `basic_1.0:abc123...`).
2.  **Update Configuration**:
    Update `network/.env` (or `docker-compose.yaml`) with this ID:
    ```bash
    CHAINCODE_ID=basic_1.0:abc123...
    ```
3.  **Restart Container**:
    Recreate the chaincode container to pick up the new ID:
    ```bash
    docker-compose -f network/docker-compose.yaml up -d --no-deps chaincode-basic
    ```
4.  **Verify**:
    Check logs: `docker logs chaincode-basic` should show the correct ID.
    Retry the invocation.

### 2. `Chaincode Definition for 'basic' exists, but chaincode is not installed`

**Error Log:**
```
Error: endorsement failure during invoke ... chaincode definition for 'basic' exists, but chaincode is not installed
```

**Cause:**
The Chaincode Definition committed to the channel (e.g., Sequence 1) points to a Package ID that is NOT currently installed on the Peer. This happens if you re-install the chaincode (generating a new hash) but fail to Approve/Commit the new definition (Sequence 2).

**Solution:**
1.  **Check Installed Packages**: `peer lifecycle chaincode queryinstalled`.
2.  **Check Committed Definition**: `peer lifecycle chaincode querycommitted --channelID mychannel --name basic`.
3.  **Sync**:
    *   If hashes differ, you must **Approve** and **Commit** the new Package ID with the **Next Sequence** number.

### 3. Packaging "No such file or directory" inside Scripts

**Cause:**
Scripts running inside the `cli` container cannot access files created on the host unless mapped via volumes.

**Solution:**
*   Package the chaincode (`basic.tar.gz`) on the **Host** first.
*   Copy it to the container: `docker cp basic.tar.gz cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/`.
*   Run the script inside the container.

## Deployment Workflow (Safe Manual Method)

1.  **Package** on Host -> `basic.tar.gz`.
2.  **Copy** to CLI container.
3.  **Install** on Peers (get Package ID).
4.  **Update** `CHAINCODE_ID` in `network/.env`.
5.  **Restart** Chaincode Container.
6.  **Approve & Commit** Definition.
7.  **Invoke** `InitLedger`.
