#!/bin/bash

BASE_URL="http://localhost:3000"

echo "🎯 Starting Initialization: 10 Users + 1 Asset per User"

# 0. Login as Admin to get Admin Token (needed for approval)
echo "🔑 Logging in as Admin..."
ADMIN_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "adminpw", "org": "Org1MSP"}' | jq -r '.token')

if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Admin login failed. Make sure backend is running and admin is enrolled."
    exit 1
fi
echo "   ✅ Admin token obtained"

for i in $(seq -f "%02g" 1 10); do
    USERNAME="user_$i"
    PASSWORD="password123"
    EMAIL="user_$i@example.org"
    ASSET_ID="asset_$i"
    
    echo "------------------------------------------"
    echo "👤 Processing $USERNAME..."

    # 1. Register User (with Email)
    echo "  📝 Registering..."
    REG_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/auth/register" \
      -H "Content-Type: application/json" \
      -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\", \"email\": \"$EMAIL\", \"org\": \"Org1MSP\"}")
    
    echo "    ✅ Registration Status: $REG_STATUS"

    # 2. Approve User (Activate)
    echo "  ✅ Approving..."
    APP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/admin/users/$USERNAME/status" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"status": "ACTIVE"}')
    
    echo "    ✅ Approval Status: $APP_STATUS"

    # 3. Login to get User Token
    echo "  🔑 Logging in..."
    TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\", \"org\": \"Org1MSP\"}" | jq -r '.token')

    if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
        echo "    ❌ Login failed for $USERNAME"
        continue
    fi

    # 4. Create Asset
    echo "  📦 Creating $ASSET_ID..."
    ASSET_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/assets" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"id\": \"$ASSET_ID\", \"name\": \"Asset of $USERNAME\", \"desc\": \"Automatically generated for test\", \"view\": \"PUBLIC\"}")

    echo "    ✅ Asset creation status: $ASSET_STATUS"
done

echo "------------------------------------------"
echo "🎉 Initialization Complete!"
