#!/bin/bash
set -e

# Configuration
CC_NAME="basic"
CC_LABEL="basic_1.0"
CC_SRC_PATH="./chaincode"
NETWORK_PATH="./network"

echo "📦 Packaging Chaincode '$CC_NAME'..."

# Create connection.json
cat <<JSON > connection.json
{
  "address": "chaincode-basic:9999",
  "transport_protocol": "grpc",
  "tls_required": false,
  "dial_timeout": "10s"
}
JSON

# Create metadata.json
cat <<JSON > metadata.json
{
  "type": "ccaas",
  "label": "$CC_LABEL"
}
JSON

# Package it
tar cfz code.tar.gz connection.json
tar cfz ${CC_NAME}.tar.gz metadata.json code.tar.gz

# Cleanup temp files
rm connection.json metadata.json code.tar.gz

echo "✅ Chaincode package '${CC_NAME}.tar.gz' created."

# Get Package ID and update .env
PACKAGE_ID=$(shasum -a 256 ${CC_NAME}.tar.gz | awk '{print $1}')
FULL_PACKAGE_ID="${CC_LABEL}:${PACKAGE_ID}"

echo "📝 Updating PACKAGE_ID in network/.env..."
echo "CHAINCODE_ID=${FULL_PACKAGE_ID}" > ${NETWORK_PATH}/.env

echo "🚀 New Package ID: ${FULL_PACKAGE_ID}"
