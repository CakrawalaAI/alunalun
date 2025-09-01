#!/bin/bash

# Test OAuth Flow Script
# This demonstrates the full server-side OAuth flow

BASE_URL="http://localhost:8080"
REDIRECT_URI="http://localhost:3000"

echo "=== OAuth Flow Test ==="
echo ""

# Step 1: Check available providers
echo "1. Checking available OAuth providers..."
curl -s "$BASE_URL/auth/oauth/" | jq .
echo ""

# Step 2: Initiate OAuth flow (this would normally redirect the browser)
echo "2. Initiating Google OAuth flow..."
echo "In a real scenario, the browser would be redirected to:"
echo "$BASE_URL/auth/oauth/google?redirect_uri=$REDIRECT_URI"
echo ""

# Step 3: Test anonymous session creation
echo "3. Creating anonymous session..."
ANON_RESPONSE=$(curl -s -X POST "$BASE_URL/api.v1.service.auth.AuthService/InitAnonymous" \
  -H "Content-Type: application/json" \
  -d '{"username": "test_user_123"}')
echo "$ANON_RESPONSE" | jq .

# Extract token and session_id
ANON_TOKEN=$(echo "$ANON_RESPONSE" | jq -r '.token')
SESSION_ID=$(echo "$ANON_RESPONSE" | jq -r '.session_id')

echo ""
echo "Anonymous token (never expires): ${ANON_TOKEN:0:50}..."
echo "Session ID: $SESSION_ID"
echo ""

# Step 4: Test email/password authentication
echo "4. Testing email/password authentication..."
# First, you'd need to register a user (not shown here)
# Then authenticate:
# curl -s -X POST "$BASE_URL/api.v1.service.auth.AuthService/Authenticate" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "provider": "email",
#     "credential": "{\"email\":\"test@example.com\",\"password\":\"Test123!\"}"
#   }' | jq .

# Step 5: Test token refresh
echo "5. Token refresh endpoint available at:"
echo "POST $BASE_URL/auth/refresh"
echo "Body: {\"expired_token\": \"<your_expired_jwt>\"}"
echo ""

# Step 6: Get public key for JWT verification
echo "6. Getting public key for JWT verification..."
curl -s "$BASE_URL/auth/public-key" | jq .
echo ""

echo "=== OAuth Flow URLs ==="
echo "Web flow (redirects):"
echo "  1. User visits: $BASE_URL/auth/oauth/google?redirect_uri=$REDIRECT_URI"
echo "  2. User authorizes on Google"
echo "  3. Google redirects to: $BASE_URL/auth/oauth/google/callback?code=XXX&state=YYY"
echo "  4. Server redirects to: $REDIRECT_URI#token=JWT_TOKEN"
echo ""
echo "Mobile flow (JSON response):"
echo "  Same flow but with 'X-Client-Type: mobile' header"
echo "  Returns JSON instead of redirect"
echo ""

echo "=== Session Migration ==="
echo "To migrate anonymous session to authenticated:"
echo "  Include session_id=$SESSION_ID in OAuth initiation URL"
echo "  Example: $BASE_URL/auth/oauth/google?redirect_uri=$REDIRECT_URI&session_id=$SESSION_ID"
echo ""

echo "=== Test Complete ==="