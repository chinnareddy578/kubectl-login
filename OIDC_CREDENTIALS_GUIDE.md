# How to Get OIDC Credentials

This guide explains how to obtain OIDC credentials (Issuer URL, Client ID, and Client Secret) from various SSO providers.

## What You Need

For `kubectl-login` to work, you need three pieces of information:

1. **Issuer URL** - The OIDC provider's endpoint
2. **Client ID** - Your OAuth2 application's client identifier
3. **Client Secret** (optional) - Your OAuth2 application's secret (required for some flows)

## Quick Method: Discover OIDC Configuration

Most OIDC providers expose their configuration at a well-known endpoint. You can discover the issuer URL and other settings:

### Using the Discovery Script

```bash
# Use the provided discovery script
./scripts/discover-oidc.sh https://your-provider.com

# Examples:
./scripts/discover-oidc.sh https://accounts.google.com
./scripts/discover-oidc.sh https://your-org.okta.com
```

### Manual Discovery

```bash
# Replace with your provider's base URL
curl https://your-provider.com/.well-known/openid-configuration | jq

# Common endpoints:
# - https://accounts.google.com/.well-known/openid-configuration
# - https://your-org.okta.com/.well-known/openid-configuration
# - https://login.microsoftonline.com/{tenant}/v2.0/.well-known/openid-configuration
```

Look for:
- `issuer` - This is your Issuer URL
- `authorization_endpoint` - Used internally
- `token_endpoint` - Used internally

## Provider-Specific Guides

### 1. Google Cloud Platform / Google Workspace

#### Step 1: Create OAuth 2.0 Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select or create a project
3. Navigate to **APIs & Services** → **Credentials**
4. Click **+ CREATE CREDENTIALS** → **OAuth client ID**

#### Step 2: Configure OAuth Consent Screen

If this is your first time:
1. Go to **OAuth consent screen**
2. Choose **External** (or **Internal** for Google Workspace)
3. Fill in required fields (App name, User support email, etc.)
4. Add scopes: `openid`, `profile`, `email`
5. Add test users if needed
6. Save

#### Step 3: Create OAuth Client

1. Application type: **Web application**
2. Name: `kubectl-login` (or any name)
3. **Authorized redirect URIs**: Add:
   ```
   http://localhost:8000/callback
   ```
4. Click **CREATE**
5. Copy the **Client ID** and **Client Secret**

#### Step 4: Get Issuer URL

- **Issuer URL**: `https://accounts.google.com`

#### Example Configuration

```json
{
  "issuer_url": "https://accounts.google.com",
  "client_id": "123456789-abcdefghijklmnop.apps.googleusercontent.com",
  "client_secret": "GOCSPX-xxxxxxxxxxxxxxxxxxxxx",
  "headless": false,
  "port": 8000
}
```

---

### 2. Okta

#### Step 1: Create an Application

1. Log in to [Okta Admin Console](https://admin.okta.com/)
2. Go to **Applications** → **Applications**
3. Click **Create App Integration**
4. Choose **OIDC - OpenID Connect**
5. Choose **Web Application**
6. Click **Next**

#### Step 2: Configure Application

1. **App integration name**: `kubectl-login`
2. **Sign-in redirect URIs**: Add:
   ```
   http://localhost:8000/callback
   ```
3. **Sign-out redirect URIs**: (optional)
4. **Controlled access**: Choose appropriate option
5. Click **Save**

#### Step 3: Get Credentials

1. Go to **General Settings** tab
2. Copy:
   - **Client ID**
   - **Client secret** (click "Show client secret")

#### Step 4: Get Issuer URL

1. Go to **Security** → **API** → **Authorization Servers**
2. Click on your authorization server (usually "default")
3. Copy the **Issuer URI**

Or use the format:
```
https://{your-org}.okta.com/oauth2/default
```

#### Example Configuration

```json
{
  "issuer_url": "https://dev-123456.okta.com/oauth2/default",
  "client_id": "0oa1b2c3d4e5f6g7h8i9",
  "client_secret": "abc123xyz789",
  "headless": false,
  "port": 8000
}
```

---

### 3. Microsoft Azure AD / Entra ID

#### Step 1: Register an Application

1. Go to [Azure Portal](https://portal.azure.com/)
2. Navigate to **Azure Active Directory** → **App registrations**
3. Click **+ New registration**

#### Step 2: Configure Application

1. **Name**: `kubectl-login`
2. **Supported account types**: Choose appropriate option
3. **Redirect URI**: 
   - Platform: **Web**
   - URI: `http://localhost:8000/callback`
4. Click **Register**

#### Step 3: Get Credentials

1. On the application overview page, copy:
   - **Application (client) ID** - This is your Client ID
2. Go to **Certificates & secrets**
3. Click **+ New client secret**
4. Add description and expiration
5. Copy the **Value** (this is your Client Secret - save it immediately!)

#### Step 4: Get Issuer URL

Format:
```
https://login.microsoftonline.com/{TENANT_ID}/v2.0
```

To find your Tenant ID:
1. Go to **Azure Active Directory** → **Overview**
2. Copy the **Tenant ID**

Or use:
```
https://login.microsoftonline.com/common/v2.0
```
(works for multi-tenant, but less secure)

#### Step 5: Configure API Permissions

1. Go to **API permissions**
2. Click **+ Add a permission**
3. Select **Microsoft Graph** → **Delegated permissions**
4. Add: `openid`, `profile`, `email`
5. Click **Add permissions**
6. Click **Grant admin consent** (if you have permission)

#### Example Configuration

```json
{
  "issuer_url": "https://login.microsoftonline.com/12345678-1234-1234-1234-123456789abc/v2.0",
  "client_id": "87654321-4321-4321-4321-cba987654321",
  "client_secret": "abc~123xyz789",
  "headless": false,
  "port": 8000
}
```

---

### 4. Generic OIDC Provider

For any OIDC-compliant provider:

#### Step 1: Find the Issuer URL

The issuer URL is typically:
- The base URL of your identity provider
- Or found in the `.well-known/openid-configuration` endpoint

Try:
```bash
curl https://your-provider.com/.well-known/openid-configuration
```

Look for the `issuer` field.

#### Step 2: Register an OAuth2 Client

The process varies by provider, but generally:

1. Log in to your identity provider's admin console
2. Navigate to **Applications**, **Clients**, or **OAuth2 Clients**
3. Create a new client/application
4. Set the redirect URI to: `http://localhost:8000/callback`
5. Copy the Client ID and Client Secret

#### Step 3: Configure Scopes

Ensure these scopes are available:
- `openid` (required)
- `profile` (recommended)
- `email` (recommended)
- `offline_access` (for refresh tokens)

---

### 5. Keycloak (Self-Hosted)

#### Step 1: Create a Client

1. Log in to Keycloak Admin Console
2. Select your **Realm**
3. Go to **Clients** → **Create**
4. **Client ID**: `kubectl-login`
5. **Client Protocol**: `openid-connect`
6. Click **Save**

#### Step 2: Configure Client

1. **Access Type**: `public` or `confidential`
2. **Valid Redirect URIs**: `http://localhost:8000/callback`
3. **Web Origins**: `http://localhost:8000`
4. **Standard Flow Enabled**: ON
5. Click **Save**

#### Step 3: Get Credentials

1. Go to **Credentials** tab
2. Copy:
   - **Client ID**
   - **Client Secret** (if confidential client)

#### Step 4: Get Issuer URL

Format:
```
https://your-keycloak.com/realms/{realm-name}
```

Example:
```
https://auth.example.com/realms/myrealm
```

---

## Verify Your Credentials

### Test the Issuer URL

```bash
# Test if issuer URL is correct
curl https://your-issuer-url/.well-known/openid-configuration | jq '.issuer'
```

Should return your issuer URL.

### Test with kubectl-login

```bash
# Test authentication (will open browser)
kubectl login \
  --issuer-url https://your-issuer-url \
  --client-id your-client-id \
  --client-secret your-client-secret
```

### Common Issues

#### "Invalid redirect URI"
- Ensure `http://localhost:8000/callback` is added to your OAuth2 client
- Check for typos (http vs https, port number)

#### "Invalid client"
- Verify Client ID is correct
- Check if client is enabled/active

#### "Invalid client secret"
- Verify Client Secret is correct (no extra spaces)
- Regenerate if needed

#### "Issuer URL not found"
- Verify issuer URL is correct
- Test the `.well-known/openid-configuration` endpoint
- Check if provider supports OIDC

---

## Security Best Practices

1. **Never commit secrets** to version control
2. **Use environment variables** for secrets:
   ```bash
   export CLIENT_SECRET=your-secret
   ```
3. **Rotate secrets regularly**
4. **Use least privilege** - only request necessary scopes
5. **Restrict redirect URIs** - only allow `http://localhost:8000/callback`
6. **Use separate clients** for development and production

---

## Quick Reference Table

| Provider | Issuer URL Format | Where to Get Credentials |
|----------|------------------|--------------------------|
| Google | `https://accounts.google.com` | [Google Cloud Console](https://console.cloud.google.com/apis/credentials) |
| Okta | `https://{org}.okta.com/oauth2/default` | Okta Admin Console → Applications |
| Azure AD | `https://login.microsoftonline.com/{tenant}/v2.0` | Azure Portal → App Registrations |
| Keycloak | `https://{host}/realms/{realm}` | Keycloak Admin Console → Clients |
| Generic | Check `.well-known/openid-configuration` | Provider's admin console |

---

## Need Help?

If you're having trouble finding credentials:

1. **Check provider documentation** - Most providers have detailed OAuth2/OIDC setup guides
2. **Contact your IT/DevOps team** - They may already have OAuth2 clients configured
3. **Test with a simple provider first** - Try Google Cloud to validate the flow
4. **Use the discovery endpoint** - Most providers expose configuration at `/.well-known/openid-configuration`

---

## Example: Complete Setup Flow

```bash
# 1. Get credentials from your provider (see above)

# 2. Create config file
cat > ~/.kubectl-login/config.json <<EOF
{
  "issuer_url": "https://your-issuer-url",
  "client_id": "your-client-id",
  "client_secret": "your-client-secret",
  "headless": false,
  "port": 8000
}
EOF

# 3. Test authentication
kubectl login --config ~/.kubectl-login/config.json

# 4. If successful, configure kubectl (see NEXT_STEPS.md)
```

