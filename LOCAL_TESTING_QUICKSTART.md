# Quick Start: Local OIDC Testing

Get up and running with a local OIDC server in 5 minutes!

## One-Command Setup

```bash
./scripts/quick-test.sh
```

This script will:
1. ✅ Start Keycloak in Docker
2. ✅ Wait for it to be ready
3. ✅ Configure realm, client, and test user
4. ✅ Test authentication

## Manual Setup (Step by Step)

### 1. Start Keycloak

```bash
docker-compose up -d
```

Wait for Keycloak to be ready (check logs):
```bash
docker-compose logs -f keycloak
# Look for: "Keycloak is ready"
```

### 2. Setup Keycloak

```bash
./scripts/setup-keycloak.sh
```

This creates:
- Realm: `kubectl-login`
- Client: `kubectl-login-client`
- Test user: `testuser` / `testpassword`

**Save the client secret** from the output!

### 3. Test Authentication

```bash
# Build kubectl-login if needed
make build

# Test with browser
./scripts/test-with-keycloak.sh auth
```

Or manually:
```bash
./kubectl-login \
  --issuer-url http://localhost:8080/realms/kubectl-login \
  --client-id kubectl-login-client \
  --client-secret <your-client-secret>
```

## What You Get

- **Keycloak Admin Console**: http://localhost:8080
  - Username: `admin`
  - Password: `admin`

- **Test User**:
  - Username: `testuser`
  - Password: `testpassword`

- **OIDC Configuration**:
  - Issuer: `http://localhost:8080/realms/kubectl-login`
  - Client ID: `kubectl-login-client`
  - Client Secret: (from setup script output)

## Troubleshooting

### Keycloak won't start
```bash
# Check logs
docker-compose logs keycloak

# Restart
docker-compose restart
```

### Setup script fails
```bash
# Wait longer for Keycloak
sleep 30
./scripts/setup-keycloak.sh
```

### Port conflicts
If port 8080 is in use, edit `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # Use 8081 instead
```

## Cleanup

```bash
# Stop and remove containers
docker-compose down

# Remove everything including data
docker-compose down -v
```

## Next Steps

- Test exec credential plugin: `./scripts/test-with-keycloak.sh exec`
- Configure kubectl: See [LOCAL_TESTING.md](LOCAL_TESTING.md)
- Read full guide: [LOCAL_TESTING.md](LOCAL_TESTING.md)

