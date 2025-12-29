# Identity Management Guide (Fabric CA)

This document details the advanced identity management features implemented in the skeleton using **Fabric CA**.

## 🏗️ Architecture
Static identity generation (`cryptogen`) has been replaced by **Dynamic Enrollment**. 
- **CAs**: `ca_org1`, `ca_org2`, `ca_orderer` are dedicated containers.
- **Enrollment Script**: `scripts/enroll-identities.sh` orchestrates the registration and enrollment of all network components.

## 🏷️ NodeOUs (Role Separation)
The network now uses **NodeOUs**, which allows it to distinguish between identity types based on the certificate itself, rather than folder location.
- **Client**: General users/applications.
- **Peer**: Network peer nodes.
- **Admin**: Organization administrators.
- **Orderer**: Orderer nodes.

## 🛡️ Advanced Features Explored

### 1. Attribute-Based Access Control (ABAC)
Attributes can be embedded in a user's certificate at registration time.
- **Use Case**: REST API or Chaincode can check for `role=manager` before allowing sensitive actions.
- **Example**:
  ```bash
  fabric-ca-client register --id.name user2 --id.attrs 'role=manager:ecert'
  ```

### 2. Affiliations (Departments)
Affiliations allow for organizational hierarchy within a single Fabric organization.
- **Hierarchical Path**: `org1.sales`, `org1.finance`.
- **Example**:
  ```bash
  fabric-ca-client affiliation add org1.sales
  fabric-ca-client register --id.name user3 --id.affiliation org1.sales
  ```

### 3. Revocation & CRL
Revocation provides a security "kill switch" for compromised identities.
- **Process**:
  1. Revoke the identity: `fabric-ca-client revoke --revoke.name user3`
  2. Generate Certificate Revocation List: `fabric-ca-client gencrl`
  3. The `crl.pem` is generated in the MSP directory, informing the network to reject transactions from that user.

## 🛠️ Debugging & Inspection
To inspect the attributes inside a generated certificate:
```bash
openssl x509 -in <path_to_cert>.pem -text -noout | grep "1.2.3.4.5.6.7.8.1" -A 1
```

## ⚠️ Important Note on Permissions
Since the CAs generate certificates as `root` inside containers (or scripts run with `sudo`), ensure permissions are corrected for backend development:
```bash
sudo chown -R $USER:$USER network/crypto-config/
```
