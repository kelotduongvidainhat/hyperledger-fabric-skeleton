#!/bin/bash
#
# Enroll a single user that's already registered in Fabric CA
# Usage: ./enroll-user.sh <username> <password> <org>
# Example: ./enroll-user.sh user1 user1pw org1
#

set -e

USERNAME=$1
PASSWORD=$2
ORG=$3

if [ -z "$USERNAME" ] || [ -z "$PASSWORD" ] || [ -z "$ORG" ]; then
    echo "Usage: $0 <username> <password> <org>"
    echo "Example: $0 user1 user1pw org1"
    exit 1
fi

# Set CA port based on org
if [ "$ORG" = "org1" ]; then
    CA_PORT=7054
elif [ "$ORG" = "org2" ]; then
    CA_PORT=8054
else
    echo "Error: org must be 'org1' or 'org2'"
    exit 1
fi

export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/network/config
export FABRIC_CA_CLIENT_HOME=${PWD}/network/crypto-config/peerOrganizations/${ORG}.example.com/

echo "Enrolling ${USERNAME}@${ORG}.example.com..."

# Create user directory
mkdir -p network/crypto-config/peerOrganizations/${ORG}.example.com/users/${USERNAME}@${ORG}.example.com

# Enroll user
fabric-ca-client enroll \
  -u https://${USERNAME}:${PASSWORD}@localhost:${CA_PORT} \
  --caname ca-${ORG} \
  -M ${PWD}/network/crypto-config/peerOrganizations/${ORG}.example.com/users/${USERNAME}@${ORG}.example.com/msp \
  --tls.certfiles ${PWD}/network/crypto-config/peerOrganizations/${ORG}.example.com/ca/ca.${ORG}.example.com-cert.pem

# Copy MSP config
cp ${PWD}/network/crypto-config/peerOrganizations/${ORG}.example.com/msp/config.yaml \
   ${PWD}/network/crypto-config/peerOrganizations/${ORG}.example.com/users/${USERNAME}@${ORG}.example.com/msp/config.yaml

echo "✓ ${USERNAME}@${ORG}.example.com enrolled successfully!"
echo ""
echo "Credentials location:"
echo "  ${PWD}/network/crypto-config/peerOrganizations/${ORG}.example.com/users/${USERNAME}@${ORG}.example.com/msp/"
echo ""
echo "Use with backend API:"
echo "  -H \"X-User-ID: ${USERNAME}\" -H \"X-Org-ID: ${ORG}\""
