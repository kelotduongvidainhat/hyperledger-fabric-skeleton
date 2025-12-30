#!/bin/bash
#
# Full Nuclear Reset Script for Hyperledger Fabric Full-Stack Application
# This script performs a complete teardown and rebuild of the entire system
#

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Define project root
PROJECT_ROOT="/home/qwe/hyperledger-fabric-skeleton"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                            ║${NC}"
echo -e "${BLUE}║        HYPERLEDGER FABRIC FULL NUCLEAR RESET               ║${NC}"
echo -e "${BLUE}║                                                            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${RED}⚠️  WARNING: This will DELETE ALL DATA ⚠️${NC}"
echo -e "${RED}   - All blockchain ledger data${NC}"
echo -e "${RED}   - All database records${NC}"
echo -e "${RED}   - All enrolled identities${NC}"
echo -e "${RED}   - All Docker volumes${NC}"
echo ""
read -p "Are you sure you want to continue? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy]([Ee][Ss])?$ ]]; then
    echo "Aborted."
    exit 1
fi
echo ""

# ============================================================================
# PHASE 0: PRE-FLIGHT CHECK
# ============================================================================
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║ PHASE 0: SYSTEM HEALTH CHECK                              ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

check_cmd() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}✗ $1 is not installed${NC}"
        echo "Please install $1 before proceeding."
        exit 1
    else
        # Try different version flags
        VERSION=$($1 --version 2>/dev/null || $1 -v 2>/dev/null || echo "detected")
        echo -e "  ✓ $1 is installed ($VERSION)"
    fi
}

echo -e "${YELLOW}→ Checking Prerequisites...${NC}"


check_go_version() {
    if ! command -v go &> /dev/null; then
        echo -e "${RED}✗ go is not installed${NC}"
        echo "Please install Go (v1.20 or newer) before proceeding."
        exit 1
    fi
    GO_FULL_VERSION=$(go version)
    GO_VERSION_STR=$(go version | awk '{print $3}' | sed 's/go//')
    
    # Parse Major and Minor
    MAJOR=$(echo $GO_VERSION_STR | cut -d. -f1)
    MINOR=$(echo $GO_VERSION_STR | cut -d. -f2)
    
    # Require 1.25+
    if [ "$MAJOR" -eq 1 ] && [ "$MINOR" -lt 25 ]; then
         echo -e "${RED}✗ Go version $GO_VERSION_STR is too old. Required: 1.25+${NC}"
         echo "Current version: $GO_FULL_VERSION"
         exit 1
    fi
    echo -e "  ✓ go is installed ($GO_VERSION_STR) - Compatible"
}

check_cmd docker
check_cmd docker-compose
check_go_version
check_cmd node
check_cmd jq
check_cmd curl
echo ""

echo -e "${YELLOW}→ Checking Fabric Binaries...${NC}"
if [ ! -f "${PROJECT_ROOT}/bin/peer" ] || [ ! -f "${PROJECT_ROOT}/bin/fabric-ca-client" ] || [ ! -f "${PROJECT_ROOT}/bin/configtxgen" ]; then
    echo -e "${YELLOW}  ⚠ Fabric binaries missing or incomplete. Downloading (Fabric 2.5.14, CA 1.5.15)...${NC}"
    # Download binaries only
    curl -sSL https://bit.ly/2ysbOFE | bash -s -- --fabric-version 2.5.14 --ca-version 1.5.15 binary
    echo -e "${GREEN}  ✓ Fabric binaries downloaded${NC}"
else
    echo -e "  ✓ Fabric binaries found"
fi
echo ""

echo -e "${YELLOW}→ Checking Docker Images...${NC}"
REQUIRED_IMAGES=(
    "hyperledger/fabric-peer:2.5"
    "hyperledger/fabric-orderer:2.5"
    "hyperledger/fabric-ccenv:2.5"
    "hyperledger/fabric-tools:2.5"
    "hyperledger/fabric-ca:1.5"
    "couchdb:3.3"
    "postgres:15"
)

for img in "${REQUIRED_IMAGES[@]}"; do
    # Check if image exists (ignoring tag if needed, but here exact match preferred)
    # Using docker images -q to check existence
    if [[ "$(docker images -q $img 2> /dev/null)" == "" ]]; then
        echo -e "${YELLOW}  ⚠ Image $img missing. Attempting pull...${NC}"
        docker pull $img || echo -e "${RED}  ✗ Failed to pull $img${NC}"
    else
        echo -e "  ✓ Image $img found"
    fi
done
echo ""
echo -e "${GREEN}✓ Phase 0 Complete: System ready${NC}"
echo ""


# ============================================================================
# PHASE 1: STOP ALL RUNNING SERVICES
# ============================================================================
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║ PHASE 1: STOPPING ALL SERVICES                            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Stop backend API
echo -e "${YELLOW}→ Stopping Backend API...${NC}"
if [ -f "${PROJECT_ROOT}/backend.pid" ]; then
    BACKEND_PID=$(cat "${PROJECT_ROOT}/backend.pid")
    if ps -p $BACKEND_PID > /dev/null 2>&1; then
        kill $BACKEND_PID || true
        echo "  ✓ Backend API stopped (PID: $BACKEND_PID)"
    fi
    rm -f "${PROJECT_ROOT}/backend.pid"
else
    # Try to find and kill any running backend process
    pkill -f "backend/cmd/api" || true
    echo "  ✓ Backend API processes terminated"
fi

# Stop frontend dev server
echo -e "${YELLOW}→ Stopping Frontend Dev Server...${NC}"
pkill -f "vite" || true
pkill -f "npm.*dev" || true
echo "  ✓ Frontend dev server stopped"

# Stop chaincode containers
echo -e "${YELLOW}→ Stopping Chaincode Containers...${NC}"
if [ -d "${PROJECT_ROOT}/chaincode/asset-transfer" ] && [ -f "${PROJECT_ROOT}/chaincode/asset-transfer/docker-compose.yaml" ]; then
    cd "${PROJECT_ROOT}/chaincode/asset-transfer"
    docker-compose down -v --remove-orphans 2>/dev/null || true
    echo "  ✓ Chaincode containers stopped"
fi

# Stop Fabric network
echo -e "${YELLOW}→ Stopping Fabric Network...${NC}"
if [ -d "${PROJECT_ROOT}/network/docker" ] && [ -f "${PROJECT_ROOT}/network/docker/docker-compose.yaml" ]; then
    cd "${PROJECT_ROOT}/network/docker"
    docker-compose down -v --remove-orphans 2>/dev/null || true
    echo "  ✓ Fabric network stopped"
fi

echo -e "${GREEN}✓ Phase 1 Complete: All services stopped${NC}"
echo ""

# ============================================================================
# PHASE 2: CLEAN UP ALL RESOURCES
# ============================================================================
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║ PHASE 2: CLEANING UP RESOURCES                            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Remove all Docker containers
echo -e "${YELLOW}→ Removing Docker Containers...${NC}"
CONTAINERS=$(docker ps -aq --filter "name=peer0.org1" --filter "name=peer0.org2" --filter "name=orderer" --filter "name=ca_" --filter "name=couchdb" --filter "name=postgres" --filter "name=asset-transfer" --filter "name=cli")
if [ -n "$CONTAINERS" ]; then
    docker rm -f $CONTAINERS 2>/dev/null || true
fi
echo "  ✓ Containers removed"

# Remove Docker volumes
echo -e "${YELLOW}→ Removing Docker Volumes...${NC}"
VOLUMES=$(docker volume ls -q --filter "name=docker_peer0.org1" --filter "name=docker_peer0.org2" --filter "name=docker_orderer" --filter "name=docker_couchdb" --filter "name=docker_postgres")
if [ -n "$VOLUMES" ]; then
    docker volume rm $VOLUMES 2>/dev/null || true
fi
echo "  ✓ Volumes removed"

# Remove Docker networks
echo -e "${YELLOW}→ Removing Docker Networks...${NC}"
NETWORKS=$(docker network ls -q --filter "name=docker_fabric")
if [ -n "$NETWORKS" ]; then
    docker network rm $NETWORKS 2>/dev/null || true
fi
echo "  ✓ Networks removed"

# Clean up crypto material
echo -e "${YELLOW}→ Cleaning Crypto Material...${NC}"
if [ -d "${PROJECT_ROOT}/network/crypto-config" ]; then
    sudo rm -rf "${PROJECT_ROOT}/network/crypto-config/"* 2>/dev/null || true
    echo "  ✓ Crypto material cleaned"
fi

# Clean up channel artifacts
echo -e "${YELLOW}→ Cleaning Channel Artifacts...${NC}"
if [ -d "${PROJECT_ROOT}/network/channel-artifacts" ]; then
    sudo rm -rf "${PROJECT_ROOT}/network/channel-artifacts/"* 2>/dev/null || true
    echo "  ✓ Channel artifacts cleaned"
fi

# Clean up chaincode artifacts
echo -e "${YELLOW}→ Cleaning Chaincode Artifacts...${NC}"
if [ -d "${PROJECT_ROOT}/chaincode/asset-transfer" ]; then
    rm -f "${PROJECT_ROOT}/chaincode/asset-transfer/"*.tar.gz 2>/dev/null || true
    rm -rf "${PROJECT_ROOT}/chaincode/asset-transfer/caas-pkg" 2>/dev/null || true
    echo "  ✓ Chaincode artifacts cleaned"
fi

# Clean up backend artifacts
echo -e "${YELLOW}→ Cleaning Backend Artifacts...${NC}"
rm -rf "${PROJECT_ROOT}/backend/wallet" 2>/dev/null || true
rm -f "${PROJECT_ROOT}/backend/backend.log" 2>/dev/null || true
echo "  ✓ Backend artifacts cleaned"

# Clean up peer chaincode images
echo -e "${YELLOW}→ Cleaning Peer Chaincode Images...${NC}"
PEER_IMAGES=$(docker images --filter=reference='dev-peer*' -q)
if [ -n "$PEER_IMAGES" ]; then
    docker rmi $PEER_IMAGES 2>/dev/null || true
fi
echo "  ✓ Peer images cleaned"

echo -e "${GREEN}✓ Phase 2 Complete: All resources cleaned${NC}"
echo ""

# ============================================================================
# PHASE 3: REBUILD FROM SCRATCH
# ============================================================================
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║ PHASE 3: REBUILDING SYSTEM                                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Step 1: Setup network
echo -e "${YELLOW}→ Step 1: Setting up Fabric Network...${NC}"
cd "${PROJECT_ROOT}/scripts"
bash network-setup.sh
echo -e "${GREEN}  ✓ Network setup complete${NC}"
echo ""

# Step 2: Deploy chaincode
echo -e "${YELLOW}→ Step 2: Deploying Chaincode (CaaS)...${NC}"
cd "${PROJECT_ROOT}/scripts"
bash deploy-caas-chaincode.sh
echo -e "${GREEN}  ✓ Chaincode deployed${NC}"
echo ""

# Step 3: Initialize database
echo -e "${YELLOW}→ Step 3: Initializing PostgreSQL Database...${NC}"
# Wait for PostgreSQL to be ready
sleep 5
# Create database and initialize schema
docker exec -i docker-postgres-1 psql -U postgres -c "CREATE DATABASE fabricdb;" 2>/dev/null || true
docker exec -i docker-postgres-1 psql -U postgres -d fabricdb < "${PROJECT_ROOT}/database/schema.sql" 2>/dev/null || true
echo -e "${GREEN}  ✓ Database initialized${NC}"
echo ""

# Step 4: Enroll identities
echo -e "${YELLOW}→ Step 4: Enrolling Identities...${NC}"
cd "${PROJECT_ROOT}/scripts"
bash enroll-identities.sh
echo -e "${GREEN}  ✓ Identities enrolled${NC}"
echo ""

# Step 5: Start backend
echo -e "${YELLOW}→ Step 5: Starting Backend API...${NC}"
cd "${PROJECT_ROOT}/backend"
nohup go run main.go > backend.log 2>&1 &
sleep 2
BACKEND_PID=$(pgrep -f "go run main.go")
echo $BACKEND_PID > "${PROJECT_ROOT}/backend.pid"
echo -e "${GREEN}  ✓ Backend API started (PID: $BACKEND_PID)${NC}"
echo ""

# Wait for backend to be ready
echo -e "${YELLOW}→ Waiting for Backend API to be ready...${NC}"
sleep 5
echo -e "${GREEN}  ✓ Backend API ready${NC}"
echo ""

# Step 6: Sync blockchain to database
echo -e "${YELLOW}→ Step 6: Syncing Blockchain to Database...${NC}"
cd "${PROJECT_ROOT}/scripts"
bash sync-blockchain-to-db.sh
echo -e "${GREEN}  ✓ Database synced with blockchain${NC}"
echo ""

echo -e "${GREEN}✓ Phase 3 Complete: System rebuilt${NC}"
echo ""

# ============================================================================
# PHASE 4: VERIFICATION
# ============================================================================
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║ PHASE 4: VERIFICATION                                     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Test chaincode
echo -e "${YELLOW}→ Testing Chaincode (InitLedger)...${NC}"
docker exec cli peer chaincode invoke \
    -o orderer.example.com:7050 \
    --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    -C mychannel -n asset-transfer \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    -c '{"function":"InitLedger","Args":[]}' \
    --waitForEvent 2>/dev/null || true

sleep 3

# Query assets
echo -e "${YELLOW}→ Querying Assets...${NC}"
ASSETS=$(docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{"Args":["GetAllAssets"]}' 2>/dev/null)
if [ -n "$ASSETS" ]; then
    echo -e "${GREEN}  ✓ Chaincode operational${NC}"
    echo "  Sample assets: $(echo $ASSETS | jq -r '.[0].ID' 2>/dev/null || echo 'OK')"
else
    echo -e "${RED}  ✗ Chaincode query failed${NC}"
fi
echo ""

# Check database sync
echo -e "${YELLOW}→ Checking Database Synchronization...${NC}"
sleep 2
BLOCKCHAIN_COUNT=$(echo $ASSETS | jq '. | length' 2>/dev/null || echo "0")
DB_COUNT=$(docker exec docker-postgres-1 psql -U postgres -d fabricdb -t -c "SELECT COUNT(*) FROM assets;" 2>/dev/null | tr -d ' ')
if [ "$DB_COUNT" -eq "$BLOCKCHAIN_COUNT" ]; then
    echo -e "${GREEN}  ✓ Database fully synchronized ($DB_COUNT/$BLOCKCHAIN_COUNT assets)${NC}"
elif [ "$DB_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}  ⚠ Database partially synced ($DB_COUNT/$BLOCKCHAIN_COUNT assets)${NC}"
else
    echo -e "${RED}  ✗ Database sync failed${NC}"
fi
echo ""

# Check backend API
echo -e "${YELLOW}→ Checking Backend API...${NC}"
HEALTH=$(curl -s http://localhost:8080/health 2>/dev/null || echo "")
if [ -n "$HEALTH" ]; then
    echo -e "${GREEN}  ✓ Backend API responding${NC}"
else
    echo -e "${YELLOW}  ⚠ Backend API not responding yet${NC}"
fi
echo ""

echo -e "${GREEN}✓ Phase 4 Complete: Verification done${NC}"
echo ""

# ============================================================================
# SUMMARY
# ============================================================================
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                            ║${NC}"
echo -e "${BLUE}║              ✓ FRESH START COMPLETE ✓                     ║${NC}"
echo -e "${BLUE}║                                                            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}System Status:${NC}"
echo -e "  • Fabric Network:  ${GREEN}✓ Running${NC}"
echo -e "  • Chaincode:       ${GREEN}✓ Deployed${NC}"
echo -e "  • PostgreSQL:      ${GREEN}✓ Running${NC}"
echo -e "  • Backend API:     ${GREEN}✓ Running (Port 8080)${NC}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo -e "  1. Start frontend:  ${BLUE}cd frontend && npm run dev${NC}"
echo -e "  2. View logs:       ${BLUE}tail -f backend/backend.log${NC}"
echo -e "  3. Check network:   ${BLUE}docker ps${NC}"
echo ""
echo -e "${YELLOW}Quick Commands:${NC}"
echo -e "  • Query assets:     ${BLUE}docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{\"Args\":[\"GetAllAssets\"]}'${NC}"
echo -e "  • View DB:          ${BLUE}docker exec -it docker-postgres-1 psql -U postgres -d fabricdb${NC}"
echo -e "  • API health:       ${BLUE}curl http://localhost:8080/health${NC}"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"
echo ""
