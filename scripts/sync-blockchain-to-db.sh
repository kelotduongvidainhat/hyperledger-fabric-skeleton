#!/bin/bash
#
# Sync all blockchain assets to PostgreSQL database
# This script backfills the database with existing blockchain data
#

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Blockchain to Database Sync Utility                   ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Get all assets from blockchain
echo -e "${YELLOW}→ Fetching all assets from blockchain...${NC}"
ASSETS_JSON=$(docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{"Args":["GetAllAssets"]}' 2>/dev/null)

if [ -z "$ASSETS_JSON" ] || [ "$ASSETS_JSON" == "null" ]; then
    echo -e "${YELLOW}No assets found on blockchain${NC}"
    exit 0
fi

# Count assets
ASSET_COUNT=$(echo "$ASSETS_JSON" | jq '. | length')
echo -e "${GREEN}  ✓ Found $ASSET_COUNT assets on blockchain${NC}"
echo ""

# Sync each asset to database
echo -e "${YELLOW}→ Syncing assets to database...${NC}"

echo "$ASSETS_JSON" | jq -c '.[]' | while read -r asset; do
    ID=$(echo "$asset" | jq -r '.ID')
    NAME=$(echo "$asset" | jq -r '.Name')
    CATEGORY=$(echo "$asset" | jq -r '.Category')
    OWNER=$(echo "$asset" | jq -r '.Owner')
    STATUS=$(echo "$asset" | jq -r '.Status')
    UPDATED=$(echo "$asset" | jq -r '.Updated')
    UPDATED_BY=$(echo "$asset" | jq -r '.UpdatedBy')
    
    # Insert or update in database
    docker exec docker-postgres-1 psql -U postgres -d fabricdb -c \
        "INSERT INTO assets (id, name, category, owner, status, updated, updated_by, last_updated)
         VALUES ('$ID', '$NAME', '$CATEGORY', '$OWNER', '$STATUS', '$UPDATED', '$UPDATED_BY', NOW())
         ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            category = EXCLUDED.category,
            owner = EXCLUDED.owner,
            status = EXCLUDED.status,
            updated = EXCLUDED.updated,
            updated_by = EXCLUDED.updated_by,
            last_updated = NOW();" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} Synced Asset: $ID"
    else
        echo -e "  ${RED}✗${NC} Failed Asset: $ID"
    fi
done

# Sync users
echo -e "${YELLOW}→ Fetching all users from blockchain...${NC}"
USERS_JSON=$(docker exec cli peer chaincode query -C mychannel -n asset-transfer -c '{"Args":["GetAllUsers"]}' 2>/dev/null)

if [ -n "$USERS_JSON" ] && [ "$USERS_JSON" != "null" ]; then
    USER_COUNT=$(echo "$USERS_JSON" | jq '. | length')
    echo -e "${GREEN}  ✓ Found $USER_COUNT users on blockchain${NC}"
    echo -e "${YELLOW}→ Syncing users to database...${NC}"

    echo "$USERS_JSON" | jq -c '.[]' | while read -r user; do
        ID=$(echo "$user" | jq -r '.ID')
        NAME=$(echo "$user" | jq -r '.Name')
        ROLE=$(echo "$user" | jq -r '.Role')
        STATUS=$(echo "$user" | jq -r '.Status')
        UPDATED=$(echo "$user" | jq -r '.Updated')
        UPDATED_BY=$(echo "$user" | jq -r '.UpdatedBy')

        # Insert or update in database
        docker exec docker-postgres-1 psql -U postgres -d fabricdb -c \
            "INSERT INTO users (id, name, role, status, updated, updated_by, last_updated)
             VALUES ('$ID', '$NAME', '$ROLE', '$STATUS', '$UPDATED', '$UPDATED_BY', NOW())
             ON CONFLICT (id) DO UPDATE SET
                name = EXCLUDED.name,
                role = EXCLUDED.role,
                status = EXCLUDED.status,
                updated = EXCLUDED.updated,
                updated_by = EXCLUDED.updated_by,
                last_updated = NOW();" > /dev/null 2>&1
        
        if [ $? -eq 0 ]; then
            echo -e "  ${GREEN}✓${NC} Synced User: $ID"
        else
            echo -e "  ${RED}✗${NC} Failed User: $ID"
        fi
    done
else
    echo -e "${YELLOW}No users found on blockchain${NC}"
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║ Sync Complete!                                            ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Verify database count
DB_COUNT=$(docker exec docker-postgres-1 psql -U postgres -d fabricdb -t -c "SELECT COUNT(*) FROM assets;" 2>/dev/null | tr -d ' ')
echo -e "${YELLOW}→ Verification:${NC}"
echo -e "  • Blockchain assets: ${BLUE}$ASSET_COUNT${NC}"
echo -e "  • Database assets:   ${BLUE}$DB_COUNT${NC}"

if [ "$ASSET_COUNT" -eq "$DB_COUNT" ]; then
    echo -e "  • Status: ${GREEN}✓ 100% synced${NC}"
else
    echo -e "  • Status: ${YELLOW}⚠ Partial sync${NC}"
fi
echo ""

