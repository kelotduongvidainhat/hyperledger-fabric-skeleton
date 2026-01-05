#!/bin/bash
set -e

API_URL="http://localhost:3000"
ASSET_ID="api_test_$(date +%s)"

echo "🚀 Starting Backend API Integration Test..."

# 1. Create Asset via API
echo "----------------------------------------------------------------"
echo "STEP 1: Creating Asset '$ASSET_ID' via API..."
curl -X POST "$API_URL/assets" \
     -H "Content-Type: application/json" \
     -d '{"id":"'"$ASSET_ID"'", "name":"API Asset", "desc":"Created via Fiber", "image_url":"http://api.io/img", "image_hash":"QmApi", "view":"Public"}'

echo ""
echo "✅ Asset Creation Request Sent"
sleep 3 # Wait for commit

# 2. Read Asset via API
echo "----------------------------------------------------------------"
echo "STEP 2: Reading Asset '$ASSET_ID' via API..."
RESPONSE=$(curl -s "$API_URL/assets/$ASSET_ID")
echo "Response: $RESPONSE"

if [[ $RESPONSE == *"$ASSET_ID"* ]]; then
    echo "✅ API Read Success"
else
    echo "❌ API Read Failed"
    exit 1
fi

# 3. Read Asset History
echo "----------------------------------------------------------------"
echo "STEP 3: Checking History via API..."
HISTORY=$(curl -s "$API_URL/assets/$ASSET_ID/history")
echo "History: $HISTORY"

if [[ $HISTORY == *"CREATE"* ]] || [[ $HISTORY == *"TRANSFER_ACCEPT_OR_CREATE"* ]]; then
     echo "✅ History Verified"
else
     echo "❌ History Check Failed"
     exit 1
fi

echo "----------------------------------------------------------------"
echo "🎉 API Integration Test Passed!"
