# Quick Start Guide

Get up and running with kubectl-login in 5 minutes!

## Step 1: Build and Install

```bash
# Build the plugin
make build

# Install to your local bin (recommended)
make install

# OR install system-wide
make install-system
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

**📖 Detailed Instructions**: See [OIDC_CREDENTIALS_GUIDE.md](OIDC_CREDENTIALS_GUIDE.md) for step-by-step instructions for:
- Google Cloud Platform
- Okta
- Microsoft Azure AD
- Keycloak
- Generic OIDC providers

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

Edit your `~/.kube/config` and add:

```yaml
users:
- name: sso-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: kubectl-login
      args:
      - --issuer-url
      - https://your-oidc-provider.com
      - --client-id
      - your-client-id
```

Then test:

```bash
kubectl get pods
```

## That's It! 🎉

You're now authenticated with SSO. For more details, see [NEXT_STEPS.md](NEXT_STEPS.md).

