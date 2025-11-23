# Next Steps Guide

This guide will help you set up and use the kubectl-login plugin.

## 1. Install the Plugin

### Option A: Install to a directory in your PATH

```bash
# Make sure kubectl-login is executable
chmod +x kubectl-login

# Install to ~/bin (create if it doesn't exist)
mkdir -p ~/bin
cp kubectl-login ~/bin/

# Add ~/bin to PATH if not already there (add to ~/.zshrc or ~/.bashrc)
export PATH="$HOME/bin:$PATH"
```

### Option B: Install to system directory

```bash
sudo cp kubectl-login /usr/local/bin/
```

### Verify Installation

```bash
kubectl plugin list
# Should show: kubectl-login

kubectl login --help
# Should show the help message
```

## 2. Get Your SSO Provider Details

You'll need the following information from your SSO provider:

- **Issuer URL**: The OIDC issuer endpoint (e.g., `https://accounts.google.com`, `https://your-org.okta.com/oauth2/default`)
- **Client ID**: Your OAuth2/OIDC client ID
- **Client Secret** (optional): Your client secret if required

### Common SSO Providers

#### Google Cloud Platform
- Issuer: `https://accounts.google.com`
- Get Client ID from: [Google Cloud Console](https://console.cloud.google.com/apis/credentials)

#### Okta
- Issuer: `https://your-org.okta.com/oauth2/default`
- Get Client ID from: Okta Admin Console → Applications

#### Azure AD
- Issuer: `https://login.microsoftonline.com/{TENANT_ID}/v2.0`
- Get Client ID from: Azure Portal → App Registrations

#### Generic OIDC Provider
- Check your provider's documentation for the issuer URL
- Usually found in `.well-known/openid-configuration` endpoint

## 3. Test Authentication

### Test with Browser Mode (Interactive)

```bash
kubectl login \
  --issuer-url https://your-oidc-provider.com \
  --client-id your-client-id
```

This should:
1. Open your browser
2. Prompt you to log in
3. Show a success message
4. Cache your token

### Test with Headless Mode

```bash
kubectl login \
  --issuer-url https://your-oidc-provider.com \
  --client-id your-client-id \
  --headless
```

## 4. Create a Configuration File (Optional but Recommended)

Create `~/.kubectl-login/config.json`:

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

## 5. Configure Kubernetes to Use the Plugin

### Option A: Manual kubeconfig Edit

Edit your `~/.kube/config` file and add/modify a user section:

```yaml
apiVersion: v1
kind: Config
users:
- name: my-sso-user
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
      # Or use config file:
      # - --config
      # - ~/.kubectl-login/config.json
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
    # Add certificate-authority-data if needed
contexts:
- name: my-context
  context:
    user: my-sso-user
    cluster: my-cluster
current-context: my-context
```

### Option B: Use kubectl config set-credentials

```bash
kubectl config set-credentials my-sso-user \
  --exec-command=kubectl-login \
  --exec-arg=--issuer-url \
  --exec-arg=https://your-oidc-provider.com \
  --exec-arg=--client-id \
  --exec-arg=your-client-id
```

## 6. Test kubectl Commands

After configuring, test that kubectl can authenticate:

```bash
# This should trigger authentication if needed
kubectl get pods

# Check your current user
kubectl config current-context
kubectl config view
```

## 7. Troubleshooting

### Check Token Cache

```bash
# View cached tokens (be careful, contains sensitive data)
cat ~/.cache/kubectl-login/tokens.json

# Clear cache if needed
rm ~/.cache/kubectl-login/tokens.json
```

### Debug Mode

Add verbose output by checking stderr:

```bash
kubectl login --issuer-url ... --client-id ... 2>&1
```

### Common Issues

1. **"Browser doesn't open"**
   - Copy the URL from stderr and open manually
   - Check that `open` command works on macOS

2. **"Port already in use"**
   - Use `--port` flag to specify a different port
   - Check what's using port 8000: `lsof -i :8000`

3. **"Authentication failed"**
   - Verify issuer URL is correct
   - Check client ID and secret
   - Ensure redirect URI is configured in your OIDC provider
   - For browser mode, redirect URI should be: `http://localhost:8000/callback`

4. **"Token refresh failed"**
   - Clear cache and re-authenticate
   - Check if refresh token is being returned

## 8. CI/CD Integration

For headless authentication in CI/CD:

```yaml
# Example GitHub Actions
- name: Authenticate with Kubernetes
  env:
    CLIENT_SECRET: ${{ secrets.OIDC_CLIENT_SECRET }}
  run: |
    kubectl login \
      --issuer-url https://your-oidc-provider.com \
      --client-id ${{ secrets.OIDC_CLIENT_ID }} \
      --headless
```

## 9. Security Best Practices

1. **Never commit secrets**: Use environment variables or secret management
2. **Restrict cache permissions**: Cache file should be 0600 (already handled)
3. **Use least privilege**: Configure OIDC client with minimal required scopes
4. **Rotate secrets**: Regularly rotate client secrets
5. **Monitor usage**: Check OIDC provider logs for suspicious activity

## 10. Advanced Configuration

### Multiple Clusters/Contexts

You can create different config files for different clusters:

```bash
# Production
kubectl login --config ~/.kubectl-login/prod.json

# Development  
kubectl login --config ~/.kubectl-login/dev.json
```

### Environment Variables

You can also use environment variables:

```bash
export OIDC_ISSUER_URL=https://your-oidc-provider.com
export OIDC_CLIENT_ID=your-client-id
export CLIENT_SECRET=your-client-secret

kubectl login --issuer-url $OIDC_ISSUER_URL --client-id $OIDC_CLIENT_ID
```

## Next: Customize for Your Use Case

- Add provider-specific logic if needed
- Customize scopes based on your requirements
- Add additional authentication methods
- Integrate with your organization's SSO solution

