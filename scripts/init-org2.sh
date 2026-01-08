#!/bin/bash

BASE_URL="http://localhost:3000"

echo "🎯 Starting Initialization for Org2: 1 Admin + 5 Users + 1 Asset per User"

# 0. Login as existing Admin (Org1) to get Admin Token for approvals
echo "🔑 Logging in as Org1 Admin..."
ADMIN_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "adminpw", "org": "Org1MSP"}' | jq -r '.token')

if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Admin login failed. Make sure backend is running."
    exit 1
fi

# 1. Create and Approve Admin for Org2
USERNAME="admin-org2"
PASSWORD="user"
EMAIL="admin-org2@org2.example.com"

echo "------------------------------------------"
echo "👑 Creating Admin for Org2: $USERNAME"

echo "  📝 Registering..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\", \"email\": \"$EMAIL\", \"org\": \"Org2MSP\"}" > /dev/null

echo "  ✅ Approving..."
curl -s -X POST "$BASE_URL/admin/users/$USERNAME/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "ACTIVE"}' > /dev/null

# 2. Login as Org2 Admin to gain its token (optional, but good for verification)
echo "  🔑 Verifying $USERNAME login..."
ORG2_ADMIN_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\", \"org\": \"Org2MSP\"}" | jq -r '.token')

if [ "$ORG2_ADMIN_TOKEN" != "null" ]; then
    echo "    ✅ Org2 Admin activated"
fi

# 3. Create 5 Users for Org2
for i in $(seq -f "%02g" 1 5); do
    USER="user_org2_$i"
    PASS="user"
    MAIL="$USER@org2.example.com"
    ASSET="asset_org2_$i"
    
    echo "------------------------------------------"
    echo "👤 Processing $USER..."

    # Register
    echo "  📝 Registering..."
    curl -s -X POST "$BASE_URL/auth/register" \
      -H "Content-Type: application/json" \
      -d "{\"username\": \"$USER\", \"password\": \"$PASS\", \"email\": \"$MAIL\", \"org\": \"Org2MSP\"}" > /dev/null
    
    # Approve (Act as Org1 Admin)
    echo "  ✅ Approving..."
    curl -s -X POST "$BASE_URL/admin/users/$USER/status" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"status": "ACTIVE"}' > /dev/null

    # Login to get User Token
    echo "  🔑 Logging in..."
    U_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"username\": \"$USER\", \"password\": \"$PASS\", \"org\": \"Org2MSP\"}" | jq -r '.token')

    if [ "$U_TOKEN" == "null" ] || [ -z "$U_TOKEN" ]; then
        echo "    ❌ Login failed for $USER"
        continue
    fi

    # Create Asset
    echo "  📦 Creating $ASSET..."
    curl -s -o /dev/null -X POST "$BASE_URL/assets" \
      -H "Authorization: Bearer $U_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"id\": \"$ASSET\", \"name\": \"Asset from Org2 ($USER)\", \"desc\": \"Cross-org asset test\", \"view\": \"PUBLIC\"}"
    
    echo "    ✅ Done"
done

echo "------------------------------------------"
echo "🎉 Org2 Initialization Complete!"
