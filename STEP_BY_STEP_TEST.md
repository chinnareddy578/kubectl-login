# Step-by-Step: Testing kubectl-login with Keycloak

## Prerequisites Check

✅ Keycloak is running (you confirmed this)

## Step 1: Ensure Keycloak is Configured

Run the setup script to create realm, client, and test user:

```bash
./scripts/setup-keycloak.sh
```

**Expected output:**
- ✓ Keycloak is running
- ✓ Admin token obtained
- ⚠ Realm already exists (or ✓ Realm created)
- ✓ Client created
- ✓ Test user created
- Configuration details with Client Secret

**Save the Client Secret** from the output - you'll need it!

## Step 2: Build kubectl-login (if not already built)

```bash
go build -o kubectl-login .
```

Or use make:
```bash
make build
```

## Step 3: Create Config File

Create the configuration file:

```bash
mkdir -p ~/.kubectl-login
```

Then create `~/.kubectl-login/config.json` with your client secret:

```bash
cat > ~/.kubectl-login/config.json <<'EOF'
{
  "issuer_url": "http://localhost:8080/realms/kubectl-login",
  "client_id": "kubectl-login-client",
  "client_secret": "YOUR_CLIENT_SECRET_HERE",
  "headless": false,
  "port": 8000
}
EOF
```

**Replace `YOUR_CLIENT_SECRET_HERE`** with the client secret from Step 1.

Or use the automated script:
```bash
./scripts/test-local.sh config
```

## Step 4: Test Browser Authentication

Run kubectl-login:

```bash
./kubectl-login --config ~/.kubectl-login/config.json
```

**What happens:**
1. Opens your browser automatically
2. Shows Keycloak login page at `http://localhost:8080`
3. **Login with:**
   - Username: `testuser`
   - Password: `testpassword`
4. Browser redirects back
5. Terminal shows success message

**Expected output:**
```
Opening browser for authentication...
If the browser doesn't open, visit: http://localhost:8080/realms/kubectl-login/protocol/openid-connect/auth?...
Successfully authenticated as: testuser@example.com
Successfully authenticated! Token expires in 59m59.123s
You can now use kubectl commands.
```

## Step 5: Verify Token Cache

Check that the token was cached:

```bash
cat ~/.cache/kubectl-login/tokens.json | jq '.'
```

You should see your cached token with expiry information.

## Step 6: Test Again (Uses Cached Token)

Run the same command again immediately:

```bash
./kubectl-login --config ~/.kubectl-login/config.json
```

**Expected output:**
```
Using cached token (expires in 59m30s)
```

No browser should open this time - it uses the cached token!

## Step 7: Test Exec Credential Plugin

Test the kubectl exec credential interface:

```bash
echo '{"apiVersion":"client.authentication.k8s.io/v1beta1","kind":"ExecCredential"}' | \
  ./kubectl-login \
    --issuer-url http://localhost:8080/realms/kubectl-login \
    --client-id kubectl-login-client \
    --client-secret YOUR_CLIENT_SECRET
```

**Expected output (JSON):**
```json
{
  "apiVersion": "client.authentication.k8s.io/v1beta1",
  "kind": "ExecCredential",
  "status": {
    "token": "eyJhbGciOiJSUzI1NiIs...",
    "expirationTimestamp": "2025-11-22T20:30:00Z"
  }
}
```

## Alternative: Use Command-Line Flags

Instead of config file, you can use flags directly:

```bash
./kubectl-login \
  --issuer-url http://localhost:8080/realms/kubectl-login \
  --client-id kubectl-login-client \
  --client-secret YOUR_CLIENT_SECRET
```

Or with environment variable:

```bash
export CLIENT_SECRET=YOUR_CLIENT_SECRET
./kubectl-login \
  --issuer-url http://localhost:8080/realms/kubectl-login \
  --client-id kubectl-login-client
```

## Quick Test Script (Easiest)

Use the automated test script:

```bash
# Test browser authentication
./scripts/test-local.sh browser

# Test exec credential plugin
./scripts/test-local.sh exec

# Test both
./scripts/test-local.sh both
```

## Troubleshooting

### Browser doesn't open
- Copy the URL from the output and paste in browser
- Or manually visit: http://localhost:8080

### Port 8000 in use
```bash
./kubectl-login --port 8001 --config ~/.kubectl-login/config.json
```

### Authentication fails
1. Check client secret is correct
2. Verify Keycloak is running: `curl http://localhost:8080/realms/kubectl-login/.well-known/openid-configuration`
3. Clear cache: `rm ~/.cache/kubectl-login/tokens.json`
4. Try again

### "Invalid redirect URI"
- Check Keycloak client settings
- Redirect URI should be: `http://localhost:8000/callback`
- Access: http://localhost:8080 → Clients → kubectl-login-client

## Summary

**Quickest way to test:**
```bash
# 1. Setup (if not done)
./scripts/setup-keycloak.sh

# 2. Test
./scripts/test-local.sh browser
```

**Manual way:**
```bash
# 1. Get client secret
./scripts/setup-keycloak.sh | grep "Client Secret:"

# 2. Create config
./scripts/test-local.sh config

# 3. Test
./kubectl-login --config ~/.kubectl-login/config.json
```

## Next: Configure kubectl

Once authentication works, configure kubectl to use the plugin:

```yaml
# ~/.kube/config
users:
- name: keycloak-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: kubectl-login
      args:
      - --issuer-url
      - http://localhost:8080/realms/kubectl-login
      - --client-id
      - kubectl-login-client
      - --client-secret
      - YOUR_CLIENT_SECRET
```

Then test:
```bash
kubectl get pods
```

