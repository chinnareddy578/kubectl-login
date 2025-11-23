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

For detailed setup instructions, see [QUICKSTART.md](QUICKSTART.md) or [NEXT_STEPS.md](NEXT_STEPS.md).

**Need OIDC credentials?** See [OIDC_CREDENTIALS_GUIDE.md](OIDC_CREDENTIALS_GUIDE.md) for step-by-step instructions for Google, Okta, Azure AD, and other providers.

## Installation

### Build from Source

```bash
git clone https://github.com/chinnareddy578/kubectl-login.git
cd kubectl-login
go build -o kubectl-login
```

### Install as kubectl Plugin

```bash
# Copy the binary to a directory in your PATH
cp kubectl-login ~/bin/

# Or install to system directory
sudo cp kubectl-login /usr/local/bin/
```

kubectl will automatically discover plugins named `kubectl-*` in your PATH.

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

## Kubernetes Configuration

To use this plugin with kubectl, configure your kubeconfig to use the exec credential plugin:

```yaml
apiVersion: v1
kind: Config
users:
- name: my-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: kubectl-login
      args:
      - --issuer-url
      - https://your-oidc-provider.com
      - --client-id
      - your-client-id
      - --client-secret
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

### Local OIDC Testing

Test with a self-hosted OIDC server (Keycloak) using Docker:

```bash
# Start Keycloak
docker-compose up -d

# Setup Keycloak
./scripts/setup-keycloak.sh

# Test authentication
./scripts/test-with-keycloak.sh
```

See [LOCAL_TESTING.md](LOCAL_TESTING.md) for detailed instructions or [LOCAL_TESTING_QUICKSTART.md](LOCAL_TESTING_QUICKSTART.md) for a quick start.

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
