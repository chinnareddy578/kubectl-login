# Troubleshooting: "No authorization code received"

## Common Causes

### 1. Browser Not Redirecting

**Symptom:** Browser opens but doesn't redirect back after login.

**Solution:**
- Make sure you complete the login in the browser
- After clicking "Sign In", the browser should automatically redirect
- If it doesn't, check the browser console for errors
- Try a different browser

### 2. Port 8000 Blocked or In Use

**Check:**
```bash
lsof -i :8000
```

**Solution:**
- Use a different port: `--port 8001`
- Update Keycloak client redirect URI to match

### 3. Redirect URI Mismatch

**Check Keycloak client settings:**
- Go to: http://localhost:8080
- Login as admin/admin
- Navigate to: Clients → kubectl-login-client
- Check "Valid redirect URIs" includes: `http://localhost:8000/callback`

**Fix:**
```bash
# Get admin token and update client
TOKEN=$(curl -s -X POST "http://localhost:8080/realms/master/protocol/openid-connect/token" \
  -d "username=admin" -d "password=admin" -d "grant_type=password" -d "client_id=admin-cli" | jq -r '.access_token')

# Get client UUID
CLIENTS=$(curl -s -X GET "http://localhost:8080/admin/realms/kubectl-login/clients?clientId=kubectl-login-client" \
  -H "Authorization: Bearer $TOKEN")
CLIENT_UUID=$(echo "$CLIENTS" | jq -r '.[0].id')

# Update redirect URIs
curl -X PUT "http://localhost:8080/admin/realms/kubectl-login/clients/$CLIENT_UUID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "redirectUris": ["http://localhost:8000/callback", "http://localhost:8001/callback"],
    "webOrigins": ["http://localhost:8000", "http://localhost:8001"]
  }'
```

### 4. PKCE Code Challenge Issue

The code uses PKCE with "plain" method. Some OIDC providers require "S256".

**Check Keycloak supports plain:**
```bash
curl -s http://localhost:8080/realms/kubectl-login/.well-known/openid-configuration | jq '.code_challenge_methods_supported'
```

Should include: `["plain", "S256"]`

### 5. Firewall or Network Issues

**Check:**
- Localhost connections are allowed
- No firewall blocking port 8000
- Try: `curl http://localhost:8000` (should fail, but confirms port is accessible)

### 6. Timeout

The callback waits up to 5 minutes. If you take too long:
- Run the command again
- Complete login quickly

## Debug Steps

### Step 1: Check Callback Server

Run with verbose output:
```bash
./kubectl-login --config ~/.kubectl-login/config.json 2>&1 | tee auth.log
```

Look for:
- "Callback received" messages
- Any error messages
- Query parameters in callback

### Step 2: Test Redirect Manually

1. Get the auth URL from the output
2. Open it in browser
3. Complete login
4. Check the final redirect URL - it should have `?code=...&state=...`

### Step 3: Check Keycloak Logs

```bash
docker-compose logs keycloak | tail -50
```

Look for:
- Authorization requests
- Token exchanges
- Errors

### Step 4: Test with curl

Test the OAuth flow manually:

```bash
# 1. Get authorization URL (from kubectl-login output)
AUTH_URL="http://localhost:8080/realms/kubectl-login/protocol/openid-connect/auth?..."
echo "Visit: $AUTH_URL"

# 2. After login, get the code from redirect URL
# 3. Exchange code for token
CODE="<code-from-redirect>"
curl -X POST "http://localhost:8080/realms/kubectl-login/protocol/openid-connect/token" \
  -d "grant_type=authorization_code" \
  -d "code=$CODE" \
  -d "client_id=kubectl-login-client" \
  -d "client_secret=<your-secret>" \
  -d "redirect_uri=http://localhost:8000/callback"
```

## Quick Fixes

### Fix 1: Re-run Setup

```bash
./scripts/setup-keycloak.sh
```

This ensures client is configured correctly.

### Fix 2: Use Different Port

```bash
# Update config
cat > ~/.kubectl-login/config.json <<EOF
{
  "issuer_url": "http://localhost:8080/realms/kubectl-login",
  "client_id": "kubectl-login-client",
  "client_secret": "<your-secret>",
  "headless": false,
  "port": 8001
}
EOF

# Update Keycloak client redirect URI (see above)
# Then test
./kubectl-login --config ~/.kubectl-login/config.json
```

### Fix 3: Clear Cache and Retry

```bash
rm -rf ~/.cache/kubectl-login
./kubectl-login --config ~/.kubectl-login/config.json
```

## Still Not Working?

1. Check browser console for JavaScript errors
2. Verify Keycloak is accessible: `curl http://localhost:8080/realms/kubectl-login/.well-known/openid-configuration`
3. Check if callback server starts: Look for "Opening browser" message
4. Try headless mode (if device flow is configured)
5. Check network connectivity between browser and localhost:8000

## Expected Flow

1. ✅ kubectl-login starts callback server on port 8000
2. ✅ Opens browser to Keycloak
3. ✅ User logs in
4. ✅ Keycloak redirects to `http://localhost:8000/callback?code=...&state=...`
5. ✅ Callback server receives request
6. ✅ Extracts code and exchanges for token
7. ✅ Returns token

If step 4-5 fails, that's where the issue is.

