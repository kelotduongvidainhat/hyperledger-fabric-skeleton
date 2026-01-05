#!/bin/bash
# Download Hyperledger Fabric binaries
# Usage: ./download-binaries.sh

set -e

# Versions
FABRIC_VERSION=2.5.14
CA_VERSION=1.5.15

echo "Downloading Hyperledger Fabric v${FABRIC_VERSION} and Fabric CA v${CA_VERSION}..."

# Create bin and config directories if they don't exist
mkdir -p bin config

# Download script
curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh

# Run install script to download binaries and config
# -b: download binaries
# -d: download docker images (skipping as we use docker-compose)
# -s: download samples (skipping)
./install-fabric.sh --fabric-version ${FABRIC_VERSION} --ca-version ${CA_VERSION} binary

# Clean up
rm install-fabric.sh

echo "Binaries downloaded successfully to bin/"
echo "Configuration files downloaded successfully to config/"
