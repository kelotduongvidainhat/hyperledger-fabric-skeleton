#!/bin/bash
#
# Script to shut down the Hyperledger Fabric network and clean up all generated artifacts.
#

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}Hyperledger Fabric Network Cleanup${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""

# Define project root
PROJECT_ROOT="/home/qwe/hyperledger-fabric-skeleton"

# 1. Stop and remove containers and volumes
echo -e "${YELLOW}Step 1: Stopping and removing containers and volumes...${NC}"

# Chaincode containers
if [ -d "${PROJECT_ROOT}/chaincode/asset-transfer" ] && [ -f "${PROJECT_ROOT}/chaincode/asset-transfer/docker-compose.yaml" ]; then
    echo "Shutting down chaincode containers..."
    cd "${PROJECT_ROOT}/chaincode/asset-transfer"
    docker-compose down -v --remove-orphans || true
fi

# Network containers
if [ -d "${PROJECT_ROOT}/network/docker" ] && [ -f "${PROJECT_ROOT}/network/docker/docker-compose.yaml" ]; then
    echo "Shutting down network containers..."
    cd "${PROJECT_ROOT}/network/docker"
    docker-compose down -v --remove-orphans || true
fi

echo -e "${GREEN}✓ Containers and volumes removed${NC}"
echo ""

# 2. Clean up crypto material
echo -e "${YELLOW}Step 2: Cleaning up crypto material...${NC}"
if [ -d "${PROJECT_ROOT}/network/crypto-config" ]; then
    sudo rm -rf "${PROJECT_ROOT}/network/crypto-config/"*
    echo "✓ network/crypto-config cleaned"
fi

# 3. Clean up channel artifacts
echo -e "${YELLOW}Step 3: Cleaning up channel artifacts...${NC}"
if [ -d "${PROJECT_ROOT}/network/channel-artifacts" ]; then
    rm -rf "${PROJECT_ROOT}/network/channel-artifacts/"*
    echo "✓ network/channel-artifacts cleaned"
fi

# 4. Clean up chaincode packages and build artifacts
echo -e "${YELLOW}Step 4: Cleaning up chaincode artifacts...${NC}"
if [ -d "${PROJECT_ROOT}/chaincode/asset-transfer" ]; then
    rm -f "${PROJECT_ROOT}/chaincode/asset-transfer/"*.tar.gz
    rm -rf "${PROJECT_ROOT}/chaincode/asset-transfer/caas-pkg"
    echo "✓ chaincode artifacts cleaned"
fi

# 5. Final docker cleanup (optional but recommended for orphan networks)
echo -e "${YELLOW}Step 5: Final Docker cleanup...${NC}"
# Remove any leftover peer chaincode images if any (non-CaaS ones)
PEER_IMAGES=$(docker images --filter=reference='dev-peer*' -q)
if [ -n "$PEER_IMAGES" ]; then
    docker rmi $PEER_IMAGES || true
fi

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}✓ Cleanup Completed Successfully!${NC}"
echo -e "${GREEN}============================================${NC}"
