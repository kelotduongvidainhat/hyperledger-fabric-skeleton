#!/bin/bash

# This script enrolls and registers CA identities for a Hyperledger Fabric network.

export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/network/config

# Color coding for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function enrollOrg1() {
  echo -e "${YELLOW}Enrolling Org1 identities...${NC}"
  mkdir -p network/crypto-config/peerOrganizations/org1.example.com/

  export FABRIC_CA_CLIENT_HOME=${PWD}/network/crypto-config/peerOrganizations/org1.example.com/

  # Enroll the CA admin
  echo "Enrolling CA Admin..."
  fabric-ca-client enroll -u https://admin:adminpw@localhost:7054 --caname ca-org1 --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

  # Find the CA cert name
  CA_CERT=$(ls ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/msp/cacerts/*.pem | head -n 1 | xargs basename)

  # Create Org1 MSP config.yaml (needed for NodeOU support)
  echo "NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: orderer" > ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/msp/config.yaml

  # Copy CA cert to tlscacerts for the organization msp
  mkdir -p ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/msp/tlscacerts
  cp ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/msp/tlscacerts/ca.crt

  # Register peer0
  echo "Registering peer0..."
  fabric-ca-client register --caname ca-org1 --id.name peer0 --id.secret peer0pw --id.type peer --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

  # Register user
  echo "Registering user..."
  fabric-ca-client register --caname ca-org1 --id.name user1 --id.secret user1pw --id.type client --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

  # Register admin
  echo "Registering admin..."
  fabric-ca-client register --caname ca-org1 --id.name org1admin --id.secret org1adminpw --id.type admin --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

  # Enroll peer0
  echo "Enrolling peer0..."
  mkdir -p network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com
  fabric-ca-client enroll -u https://peer0:peer0pw@localhost:7054 --caname ca-org1 -M ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp --csr.hosts peer0.org1.example.com --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem
  cp ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/msp/config.yaml ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/config.yaml

  # Enroll peer0 TLS
  echo "Enrolling peer0 TLS..."
  fabric-ca-client enroll -u https://peer0:peer0pw@localhost:7054 --caname ca-org1 -M ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls --enrollment.profile tls --csr.hosts peer0.org1.example.com --csr.hosts localhost --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

  # Fix identity file names for TLS (Fabric expects server.crt, server.key, ca.crt)
  cp ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/signcerts/* ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt
  cp ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/keystore/* ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
  cat ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/tlscacerts/* > ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt

  # Enroll admin
  echo "Enrolling admin..."
  mkdir -p network/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com
  fabric-ca-client enroll -u https://org1admin:org1adminpw@localhost:7054 --caname ca-org1 -M ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem
  cp ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/msp/config.yaml ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/config.yaml
}

function enrollOrg2() {
  echo -e "${YELLOW}Enrolling Org2 identities...${NC}"
  mkdir -p network/crypto-config/peerOrganizations/org2.example.com/

  export FABRIC_CA_CLIENT_HOME=${PWD}/network/crypto-config/peerOrganizations/org2.example.com/

  # Enroll the CA admin
  echo "Enrolling CA Admin..."
  fabric-ca-client enroll -u https://admin:adminpw@localhost:8054 --caname ca-org2 --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem

  # Find the CA cert name
  CA_CERT=$(ls ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/msp/cacerts/*.pem | head -n 1 | xargs basename)

  # Create Org2 MSP config.yaml
  echo "NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: orderer" > ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/msp/config.yaml

  # Copy CA cert to tlscacerts
  mkdir -p ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/msp/tlscacerts
  cp ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/msp/tlscacerts/ca.crt

  # Register peer0
  echo "Registering peer0..."
  fabric-ca-client register --caname ca-org2 --id.name peer0 --id.secret peer0pw --id.type peer --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem

  # Register user
  echo "Registering user..."
  fabric-ca-client register --caname ca-org2 --id.name user1 --id.secret user1pw --id.type client --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem

  # Register admin
  echo "Registering admin..."
  fabric-ca-client register --caname ca-org2 --id.name org2admin --id.secret org2adminpw --id.type admin --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem

  # Enroll peer0
  echo "Enrolling peer0..."
  mkdir -p network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com
  fabric-ca-client enroll -u https://peer0:peer0pw@localhost:8054 --caname ca-org2 -M ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp --csr.hosts peer0.org2.example.com --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem
  cp ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/msp/config.yaml ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp/config.yaml

  # Enroll peer0 TLS
  echo "Enrolling peer0 TLS..."
  fabric-ca-client enroll -u https://peer0:peer0pw@localhost:8054 --caname ca-org2 -M ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls --enrollment.profile tls --csr.hosts peer0.org2.example.com --csr.hosts localhost --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem

  # Fix identity file names for TLS
  cp ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/signcerts/* ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/server.crt
  cp ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/keystore/* ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/server.key
  cat ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/tlscacerts/* > ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

  # Enroll admin
  echo "Enrolling admin..."
  mkdir -p network/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com
  fabric-ca-client enroll -u https://org2admin:org2adminpw@localhost:8054 --caname ca-org2 -M ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem
  cp ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/msp/config.yaml ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/config.yaml
}

function enrollOrderer() {
  echo -e "${YELLOW}Enrolling Orderer identities...${NC}"
  mkdir -p network/crypto-config/ordererOrganizations/example.com/

  export FABRIC_CA_CLIENT_HOME=${PWD}/network/crypto-config/ordererOrganizations/example.com/

  # Enroll the CA admin
  echo "Enrolling CA Admin..."
  fabric-ca-client enroll -u https://admin:adminpw@localhost:9054 --caname ca-orderer --tls.certfiles ${PWD}/network/crypto-config/ordererOrganizations/example.com/ca/ca.example.com-cert.pem

  # Find the CA cert name
  CA_CERT=$(ls ${PWD}/network/crypto-config/ordererOrganizations/example.com/msp/cacerts/*.pem | head -n 1 | xargs basename)

  # Create Orderer MSP config.yaml
  echo "NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/${CA_CERT}
    OrganizationalUnitIdentifier: orderer" > ${PWD}/network/crypto-config/ordererOrganizations/example.com/msp/config.yaml

  # Copy CA cert to tlscacerts
  mkdir -p ${PWD}/network/crypto-config/ordererOrganizations/example.com/msp/tlscacerts
  cp ${PWD}/network/crypto-config/ordererOrganizations/example.com/ca/ca.example.com-cert.pem ${PWD}/network/crypto-config/ordererOrganizations/example.com/msp/tlscacerts/ca.crt

  # Register orderer
  echo "Registering orderer..."
  fabric-ca-client register --caname ca-orderer --id.name orderer --id.secret ordererpw --id.type orderer --tls.certfiles ${PWD}/network/crypto-config/ordererOrganizations/example.com/ca/ca.example.com-cert.pem

  # Register orderer admin
  echo "Registering orderer admin..."
  fabric-ca-client register --caname ca-orderer --id.name ordereradmin --id.secret ordereradminpw --id.type admin --tls.certfiles ${PWD}/network/crypto-config/ordererOrganizations/example.com/ca/ca.example.com-cert.pem

  # Enroll orderer
  echo "Enrolling orderer..."
  mkdir -p network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com
  fabric-ca-client enroll -u https://orderer:ordererpw@localhost:9054 --caname ca-orderer -M ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp --csr.hosts orderer.example.com --tls.certfiles ${PWD}/network/crypto-config/ordererOrganizations/example.com/ca/ca.example.com-cert.pem
  cp ${PWD}/network/crypto-config/ordererOrganizations/example.com/msp/config.yaml ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/config.yaml

  # Enroll orderer TLS
  echo "Enrolling orderer TLS..."
  fabric-ca-client enroll -u https://orderer:ordererpw@localhost:9054 --caname ca-orderer -M ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls --enrollment.profile tls --csr.hosts orderer.example.com --csr.hosts localhost --tls.certfiles ${PWD}/network/crypto-config/ordererOrganizations/example.com/ca/ca.example.com-cert.pem

  # Fix identity file names for TLS
  cp ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/signcerts/* ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
  cp ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/keystore/* ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key
  cat ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/tlscacerts/* > ${PWD}/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt

  # Enroll orderer admin
  echo "Enrolling orderer admin..."
  mkdir -p network/crypto-config/ordererOrganizations/example.com/users/Admin@example.com
  fabric-ca-client enroll -u https://ordereradmin:ordereradminpw@localhost:9054 --caname ca-orderer -M ${PWD}/network/crypto-config/ordererOrganizations/example.com/users/Admin@example.com/msp --tls.certfiles ${PWD}/network/crypto-config/ordererOrganizations/example.com/ca/ca.example.com-cert.pem
  cp ${PWD}/network/crypto-config/ordererOrganizations/example.com/msp/config.yaml ${PWD}/network/crypto-config/ordererOrganizations/example.com/users/Admin@example.com/msp/config.yaml
}

enrollOrg1
enrollOrg2
enrollOrderer

echo -e "${GREEN}✓ All identities enrolled successfully${NC}"
