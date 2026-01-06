#!/bin/bash
BASE_URL="http://localhost:3000"

# 1. Login Admin
echo "🔑 Logging in as Admin..."
# Try adminpw (default CA password)
ADMIN_RES=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "adminpw"}')
ADMIN_TOKEN=$(echo $ADMIN_RES | jq -r '.token')

if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Admin login failed with 'adminpw'. Response: $ADMIN_RES"
    exit 1
fi
echo "   Admin Token: ${ADMIN_TOKEN:0:15}..."

# 2. Register & Login Users
TOKENS=()
echo "👥 Registering 10 Users..."
for i in {1..10}; do
    USER="user$i"
    PASS="password123"
    
    # Register (ignore error if exists)
   curl -s -X POST "$BASE_URL/auth/register" \
      -H "Content-Type: application/json" \
      -d "{\"username\": \"$USER\", \"password\": \"$PASS\", \"email\": \"$USER@example.com\"}" > /dev/null

    # Login
    TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"username\": \"$USER\", \"password\": \"$PASS\"}" | jq -r '.token')
    
    if [ "$TOKEN" == "null" ]; then
        echo "   ❌ Failed to login $USER"
    else
        TOKENS+=($TOKEN)
    fi
done
echo "   ✅ ${#TOKENS[@]} users logged in."

# 3. Create Assets
echo "💎 Creating 15 Assets..."
for i in {1..15}; do
    USER_IDX=$(( (i - 1) % 10 ))
    TOKEN=${TOKENS[$USER_IDX]}
    ASSET_ID="asset_seed_$i"
    
    # Create
    curl -s -X POST "$BASE_URL/assets" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"id\": \"$ASSET_ID\", 
        \"name\": \"Seeded Asset $i\", 
        \"desc\": \"Automatically generated asset #$i\", 
        \"view\": \"Public\"
      }" > /dev/null
      
    echo -n "."
done
echo ""

# 4. Sync
echo "🔄 Syncing to Database..."
curl -s -X POST "$BASE_URL/admin/sync" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

echo "✅ Seeding Complete!"
