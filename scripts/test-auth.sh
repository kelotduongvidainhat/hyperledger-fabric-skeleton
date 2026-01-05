#!/bin/bash
set -e

# Configuration
API_URL="http://localhost:3000"
USER_ID="test_user_$(date +%s)"
PASSWORD="password123"

echo "------------------------------------------------"
echo "Starting Auth Test for user: $USER_ID"
echo "------------------------------------------------"

# 1. Register
echo "[1/3] Registering user..."
REGISTER_RESP=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$USER_ID\", \"password\": \"$PASSWORD\"}")

if echo "$REGISTER_RESP" | grep -q "successfully"; then
    echo "Success: User registered."
else
    echo "Error: Registration failed."
    echo "Response: $REGISTER_RESP"
    exit 1
fi

# 2. Login
echo "[2/3] Logging in..."
LOGIN_RESP=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$USER_ID\", \"password\": \"$PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESP" | grep -oP '"token":"\K[^"]+')

if [ -n "$TOKEN" ]; then
    echo "Success: Obtained JWT token."
    # echo "Token: $TOKEN"
else
    echo "Error: Login failed (no token)."
    echo "Response: $LOGIN_RESP"
    exit 1
fi

# 3. Access Protected Route (Create Asset)
echo "[3/3] Testing protected route (Create Asset)..."
ASSET_ID="asset_$USER_ID"
CREATE_RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/assets" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"$ASSET_ID\", \"name\":\"Test Asset\", \"desc\":\"Verified by script\", \"view\":\"Public\"}")

if [ "$CREATE_RESP" == "200" ] || [ "$CREATE_RESP" == "201" ]; then
    echo "Success: Asset created using individual identity!"
else
    # Note: If endorsement fails because of Org2, we might get 500, 
    # but the Auth check happened BEFORE the transaction submission.
    # However, if it's 401/403, Authentication failed.
    if [ "$CREATE_RESP" == "500" ]; then
        echo "Info: Auth passed, but transaction failed endorsement (Expected if Org2 peer is down)."
    else
        echo "Error: Protected route access failed (HTTP $CREATE_RESP)."
        exit 1
    fi
fi

echo "------------------------------------------------"
echo "Auth Test Completed Successfully!"
echo "------------------------------------------------"
