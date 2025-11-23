# Quick Start Guide

Get up and running with kubectl-login in 5 minutes!

## Step 1: Build and Install

```bash
# Build the plugin
make build

# Install to your local bin (recommended)
mkdir -p ~/bin
cp kubectl-login ~/bin/

# Add ~/bin to PATH (add to ~/.zshrc or ~/.bashrc)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Or install system-wide:
```bash
make install
# or
sudo cp kubectl-login /usr/local/bin/
```

## Step 2: Verify Installation

```bash
kubectl plugin list
# Should show: kubectl-login
```

## Step 3: Get Your SSO Credentials

You need:
- **Issuer URL**: Your OIDC provider's issuer endpoint
- **Client ID**: Your OAuth2 client ID
- **Client Secret**: Your OAuth2 client secret (optional but recommended)

### Find Your Issuer URL

Most OIDC providers expose their configuration at:
```
https://your-provider.com/.well-known/openid-configuration
```

Look for the `issuer` field in the JSON response.

**For Local Testing with Keycloak**: 
- Issuer URL: `http://localhost:8080/realms/kubectl-login`
- Run `./scripts/setup-keycloak.sh` to set up a local Keycloak instance
- See [README.md](README.md#local-oidc-testing-with-keycloak) for details

## Step 4: Test Authentication

```bash
kubectl login \
  --issuer-url https://your-oidc-provider.com \
  --client-id your-client-id
```

This will:
1. Open your browser
2. Prompt you to log in
3. Cache your token
4. Show success message

## Step 5: Configure kubectl

Edit your `~/.kube/config` and add the plugin configuration for automatic authentication:

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
contexts:
- name: my-context
  context:
    user: sso-user
    cluster: my-cluster
current-context: my-context
```

Then test:

```bash
kubectl get pods
```

The plugin will automatically authenticate on the first command and reuse/refresh tokens as needed.

## Troubleshooting

### "unknown command 'login'" Error

The plugin isn't in your PATH. Verify it's installed:
```bash
# Check if plugin is discoverable
kubectl plugin list | grep login

# If not found, ensure ~/bin is in PATH
echo $PATH | grep -o "$HOME/bin"

# If empty, add to PATH and reload shell
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Browser Doesn't Open

Copy the URL from the terminal output and paste it into your browser manually.

### "Connection refused" Error

Make sure your OIDC provider is running:
```bash
# For Keycloak
docker-compose ps | grep keycloak

# For other providers, verify endpoint is accessible
curl https://your-oidc-provider.com/.well-known/openid-configuration
```

### Config File Path Issue

Use absolute paths, not tilde (~):
```bash
# ❌ Wrong
--config ~/.kubectl-login/config.json

# ✅ Correct
--config $HOME/.kubectl-login/config.json
```

## That's It! 🎉

You're now authenticated with SSO. For local testing, see [README.md](README.md#local-oidc-testing-with-keycloak).

