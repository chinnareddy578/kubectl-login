# kubectl-login

A kubectl plugin for SSO authentication using OIDC (OpenID Connect). This plugin provides seamless authentication similar to `gcloud auth login` and AWS SSO, supporting both browser-based and headless authentication modes.

## Features

- 🔐 **OIDC Authentication**: Supports any OIDC-compliant SSO provider
- 🌐 **Browser-based Login**: Interactive authentication with automatic browser opening
- 🤖 **Headless Mode**: Device flow and client credentials for CI/CD environments
- 💾 **Token Caching**: Secure token storage with automatic refresh
- 🔄 **Token Refresh**: Automatic token refresh before expiration
- ⚙️ **Flexible Configuration**: Command-line flags, environment variables, or config files

## Quick Start

```bash
# Build and install
make build && make install

# Test authentication
kubectl login --issuer-url https://your-oidc-provider.com --client-id your-client-id

# Configure kubectl (edit ~/.kube/config)
# See examples/kubeconfig.example.yaml for reference
```

For detailed setup instructions, see [QUICKSTART.md](QUICKSTART.md).

For local testing with Keycloak, see the [Local OIDC Testing](#local-oidc-testing) section below.

## Requirements

- **Go**: 1.22 or higher
- **kubectl**: 1.19 or higher
- **OpenID Connect Provider**: Any OIDC-compliant provider (Google, Okta, Azure AD, Keycloak, etc.)

## Installation

### Build from Source

```bash
git clone https://github.com/chinnareddy578/kubectl-login.git
cd kubectl-login
go build -o kubectl-login
```

### Install as kubectl Plugin

kubectl automatically discovers plugins named `kubectl-*` in your PATH.

**Option 1: Install to User Directory (Recommended)**

```bash
# Create ~/bin if it doesn't exist
mkdir -p ~/bin

# Copy the binary
cp kubectl-login ~/bin/

# Add ~/bin to your PATH (add to ~/.zshrc, ~/.bashrc, or ~/.fish/config.fish)
export PATH="$HOME/bin:$PATH"

# Reload shell
source ~/.zshrc  # or source ~/.bashrc
```

**Option 2: Install System-wide**

```bash
sudo cp kubectl-login /usr/local/bin/
```

**Verify Installation**

```bash
kubectl plugin list | grep login
# Should show: /path/to/kubectl-login

kubectl login --help
# Should display the plugin help
```

## Usage

### Basic Authentication

```bash
kubectl login \
  --issuer-url https://your-oidc-provider.com \
  --client-id your-client-id
```

### With Client Secret

```bash
kubectl login \
  --issuer-url https://your-oidc-provider.com \
  --client-id your-client-id \
  --client-secret your-client-secret
```

Or use environment variable:

```bash
export CLIENT_SECRET=your-client-secret
kubectl login --issuer-url https://your-oidc-provider.com --client-id your-client-id
```

### Headless Mode (for CI/CD)

```bash
kubectl login \
  --issuer-url https://your-oidc-provider.com \
  --client-id your-client-id \
  --headless
```

### Using Configuration File

Create a config file `~/.kubectl-login/config.json`:

```json
{
  "issuer_url": "https://your-oidc-provider.com",
  "client_id": "your-client-id",
  "client_secret": "your-client-secret",
  "headless": false,
  "port": 8000
}
```

Then use it:

```bash
kubectl login --config ~/.kubectl-login/config.json
```

## Kubernetes Integration

### Configure kubeconfig for Automatic Authentication

Edit your `~/.kube/config` to use the plugin for automatic authentication:

```yaml
apiVersion: v1
kind: Config
users:
- name: sso-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: kubectl-login
      args:
      - --config
      - ~/.kubectl-login/config.json
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
    certificate-authority-data: LS0tLS1CRUdJTi... # base64 encoded cert
contexts:
- name: my-context
  context:
    user: sso-user
    cluster: my-cluster
current-context: my-context
```

### Test kubectl Integration

Once configured, test with a kubectl command:

```bash
kubectl get pods
```

The plugin will:
1. Check for a cached token (reuse if valid)
2. Refresh the token if expiring soon
3. Authenticate with OIDC provider if needed
4. Return the token to kubectl for API calls
      - your-client-secret
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    user: my-user
    cluster: my-cluster
current-context: my-context
```

## Command-line Options

```
Flags:
  --issuer-url string      OIDC issuer URL (required)
  --client-id string        OIDC client ID (required)
  --client-secret string    OIDC client secret (optional, can be set via CLIENT_SECRET env var)
  --headless               Use headless authentication (for CI/CD)
  --port int               Local port for OAuth callback (default 8000)
  --config string          Path to configuration file
  -h, --help               Help for kubectl-login
```

## How It Works

### Browser Mode (Default)

1. Opens your default browser to the OIDC provider's login page
2. After successful authentication, receives the authorization code via callback
3. Exchanges the code for access token, refresh token, and ID token
4. Caches tokens securely for future use
5. Automatically refreshes tokens before expiration

### Headless Mode

1. Initiates device flow or client credentials flow
2. For device flow: displays a URL and code for manual authentication
3. Polls for token until authentication is complete
4. Caches tokens for subsequent use

### Exec Credential Plugin Mode

When called by kubectl as an exec credential plugin:
1. Reads the exec credential request from stdin
2. Authenticates (using cache if available)
3. Returns the token in the exec credential response format
4. kubectl uses this token for API requests

## Token Caching

Tokens are cached securely in:
- **macOS/Linux**: `~/.cache/kubectl-login/tokens.json`
- **Windows**: `%LOCALAPPDATA%\kubectl-login\tokens.json`

The cache file has restricted permissions (0600) and contains encrypted tokens.

## Examples

### Google Cloud Platform

```bash
kubectl login \
  --issuer-url https://accounts.google.com \
  --client-id YOUR_GCP_CLIENT_ID
```

### Okta

```bash
kubectl login \
  --issuer-url https://your-org.okta.com/oauth2/default \
  --client-id YOUR_OKTA_CLIENT_ID
```

### Azure AD

```bash
kubectl login \
  --issuer-url https://login.microsoftonline.com/YOUR_TENANT_ID/v2.0 \
  --client-id YOUR_AZURE_CLIENT_ID
```

## Troubleshooting

### Plugin Not Found

**Problem**: `error: unknown command "login" for "kubectl"`

**Solution**:
```bash
# 1. Verify plugin is in PATH
kubectl plugin list | grep login

# If empty, add ~/bin to PATH:
export PATH="$HOME/bin:$PATH"

# 2. Reload shell
source ~/.zshrc  # or source ~/.bashrc

# 3. Verify again
kubectl plugin list | grep login
```

### Browser doesn't open

If the browser doesn't open automatically, the plugin will display the URL. Copy and paste it into your browser manually.

### Port already in use

If the default port (8000) is in use, specify a different port:

```bash
kubectl login --port 8080 --issuer-url ... --client-id ...
```

### Token refresh fails

If token refresh fails, the plugin will attempt a new authentication. Clear the cache if you continue to have issues:

```bash
rm ~/.cache/kubectl-login/tokens.json
```

### Config file not found

**Problem**: `open ~/.kubectl-login/config.json: no such file or directory`

**Solution**: Use absolute path in kubeconfig (tilde ~ doesn't expand):
```bash
# ❌ Wrong (tilde not expanded)
--exec-arg=~/.kubectl-login/config.json

# ✅ Correct (absolute path)
--exec-arg=$HOME/.kubectl-login/config.json
```

### Keycloak connection refused

**Problem**: `connection refused` when testing with Keycloak

**Solution**:
```bash
# Check Keycloak is running
docker-compose ps | grep keycloak

# Start if not running
docker-compose up -d

# Wait for it to be ready
docker-compose logs -f keycloak | grep "Keycloak is ready"
```

### Invalid redirect URI

**Problem**: `invalid_redirect_uri` error during authentication

**Solution**: Ensure redirect URI is configured in your OIDC provider:
```
http://localhost:8000/callback
```

For Keycloak:
```bash
# Access admin console
open http://localhost:8080

# Login as admin/admin
# Go to Clients → kubectl-login-client
# Verify "Valid Redirect URIs" includes: http://localhost:8000/callback
```

## Development

### Building

```bash
go build -o kubectl-login
# or
make build
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only (skip integration)
make test-short

# Run with coverage
make test-coverage

# Run specific test suite
go test -v ./pkg/config
go test -v ./pkg/cache
go test -v ./pkg/auth

# Use the test runner script
./scripts/run-tests.sh help
```

See [TESTING.md](TESTING.md) for detailed testing documentation.

### Local OIDC Testing with Keycloak

Test with a self-hosted OIDC server (Keycloak) using Docker:

```bash
# Start Keycloak and PostgreSQL
docker-compose up -d

# Setup Keycloak realm, client, and test user
./scripts/setup-keycloak.sh

# Get a token for testing
curl -X POST "http://localhost:8080/realms/kubectl-login/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=kubectl-login-client" \
  -d "client_secret=<client_secret_from_setup>" \
  -d "username=testuser" \
  -d "password=testpassword"
```

Test user credentials created by the setup script:
- **Username**: `testuser`
- **Password**: `testpassword`

Configuration saved to `~/.kubectl-login/config.json` for local testing.

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
