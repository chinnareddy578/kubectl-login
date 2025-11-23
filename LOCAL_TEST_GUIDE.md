# Local Testing Guide

Quick guide to test kubectl-login with your local Keycloak instance.

## Prerequisites

1. Keycloak running: `docker-compose up -d`
2. Keycloak configured: `./scripts/setup-keycloak.sh`
3. kubectl-login built: `make build` or `go build -o kubectl-login .`

## Quick Test (Recommended)

Use the automated test script:

```bash
# Test browser authentication
./scripts/test-local.sh browser

# Test exec credential plugin
./scripts/test-local.sh exec

# Test both
./scripts/test-local.sh both

# Just create config file
./scripts/test-local.sh config
```

## Manual Testing

### Step 1: Create Config File

The test script creates it automatically, or create manually:

```bash
mkdir -p ~/.kubectl-login

cat > ~/.kubectl-login/config.json <<EOF
{
  "issuer_url": "http://localhost:8080/realms/kubectl-login",
  "client_id": "kubectl-login-client",
  "client_secret": "<get-from-setup-output>",
  "headless": false,
  "port": 8000
}
EOF

chmod 600 ~/.kubectl-login/config.json
```

Get the client secret by running:
```bash
./scripts/setup-keycloak.sh | grep "Client Secret:"
```

### Step 2: Test Browser Authentication

```bash
./kubectl-login --config ~/.kubectl-login/config.json
```

This will:
1. Open your browser to Keycloak login page
2. Log in with: `testuser` / `testpassword`
3. Redirect back and authenticate
4. Cache the token

**Expected output:**
```
Opening browser for authentication...
If the browser doesn't open, visit: http://localhost:8080/...
Successfully authenticated as: testuser@example.com
Successfully authenticated! Token expires in ...
You can now use kubectl commands.
```

### Step 3: Test Exec Credential Plugin

Test the kubectl exec credential interface:

```bash
echo '{"apiVersion":"client.authentication.k8s.io/v1beta1","kind":"ExecCredential"}' | \
  ./kubectl-login \
    --issuer-url http://localhost:8080/realms/kubectl-login \
    --client-id kubectl-login-client \
    --client-secret <your-client-secret>
```

**Expected output:**
```json
{
  "apiVersion": "client.authentication.k8s.io/v1beta1",
  "kind": "ExecCredential",
  "status": {
    "token": "eyJhbGciOiJSUzI1NiIs...",
    "expirationTimestamp": "2025-11-22T..."
  }
}
```

### Step 4: Verify Token Cache

Check that token was cached:

```bash
cat ~/.cache/kubectl-login/tokens.json | jq '.'
```

You should see the cached token with expiry information.

## Test Scenarios

### Scenario 1: First Time Authentication

1. Clear cache: `rm ~/.cache/kubectl-login/tokens.json`
2. Run: `./kubectl-login --config ~/.kubectl-login/config.json`
3. Should open browser and prompt for login

### Scenario 2: Using Cached Token

1. Run authentication once (creates cache)
2. Run again immediately: `./kubectl-login --config ~/.kubectl-login/config.json`
3. Should use cached token without browser

### Scenario 3: Token Refresh

1. Wait for token to expire (or manually expire it in cache)
2. Run: `./kubectl-login --config ~/.kubectl-login/config.json`
3. Should automatically refresh using refresh token

### Scenario 4: Exec Credential Mode

1. Test with kubectl exec credential request
2. Should return token in proper format
3. Can be used in kubeconfig

## Troubleshooting

### Browser Doesn't Open

```bash
# Manually visit the URL shown in output
# Or use headless mode (requires device flow setup)
./kubectl-login --headless --issuer-url ... --client-id ...
```

### Port 8000 Already in Use

```bash
# Use different port
./kubectl-login --port 8001 --config ~/.kubectl-login/config.json

# Update redirect URI in Keycloak client settings
```

### Authentication Fails

1. **Check Keycloak is running:**
   ```bash
   curl http://localhost:8080/realms/kubectl-login/.well-known/openid-configuration
   ```

2. **Check client secret:**
   ```bash
   ./scripts/setup-keycloak.sh | grep "Client Secret:"
   ```

3. **Check redirect URI in Keycloak:**
   - Should be: `http://localhost:8000/callback`
   - Access: http://localhost:8080 → Clients → kubectl-login-client

4. **Clear cache and retry:**
   ```bash
   rm ~/.cache/kubectl-login/tokens.json
   ```

### Token Not Cached

Check cache directory permissions:
```bash
ls -la ~/.cache/kubectl-login/
# Should be readable/writable by you
```

## Integration with kubectl

Once authentication works, configure kubectl:

```yaml
# ~/.kube/config
apiVersion: v1
kind: Config
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
      - <your-client-secret>
```

Then test:
```bash
kubectl get pods
# Should trigger authentication if needed
```

## What to Verify

✅ Browser opens to Keycloak login  
✅ Can log in with testuser/testpassword  
✅ Redirects back successfully  
✅ Token is cached  
✅ Exec credential plugin returns token  
✅ Token refresh works  
✅ kubectl can use the plugin  

## Next Steps

- Test with different users
- Test token expiration and refresh
- Test with actual Kubernetes cluster
- Test headless mode (if needed)
- Customize scopes and claims

## Quick Reference

```bash
# Start everything
docker-compose up -d
./scripts/setup-keycloak.sh
./scripts/test-local.sh browser

# Check status
docker-compose ps
curl http://localhost:8080/realms/kubectl-login/.well-known/openid-configuration

# Clean up
docker-compose down
rm -rf ~/.cache/kubectl-login
rm ~/.kubectl-login/config.json
```

