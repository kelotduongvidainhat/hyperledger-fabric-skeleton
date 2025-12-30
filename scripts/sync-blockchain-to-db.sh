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
    COLOR=$(echo "$asset" | jq -r '.Color')
    SIZE=$(echo "$asset" | jq -r '.Size')
    OWNER=$(echo "$asset" | jq -r '.Owner')
    VALUE=$(echo "$asset" | jq -r '.AppraisedValue')
    
    # Insert or update in database
    docker exec docker-postgres-1 psql -U postgres -d fabricdb -c \
        "INSERT INTO assets (id, color, size, owner, appraised_value, last_updated)
         VALUES ('$ID', '$COLOR', $SIZE, '$OWNER', $VALUE, NOW())
         ON CONFLICT (id) DO UPDATE SET
            color = EXCLUDED.color,
            size = EXCLUDED.size,
            owner = EXCLUDED.owner,
            appraised_value = EXCLUDED.appraised_value,
            last_updated = NOW();" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} Synced: $ID"
    else
        echo -e "  ${RED}✗${NC} Failed: $ID"
    fi
done

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

