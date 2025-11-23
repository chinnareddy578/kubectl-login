#!/bin/bash

# OIDC Discovery Script
# Helps discover OIDC configuration from a provider

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <provider-base-url>"
    echo ""
    echo "Examples:"
    echo "  $0 https://accounts.google.com"
    echo "  $0 https://your-org.okta.com"
    echo "  $0 https://login.microsoftonline.com/your-tenant-id/v2.0"
    exit 1
fi

PROVIDER_URL="$1"
WELL_KNOWN="${PROVIDER_URL}/.well-known/openid-configuration"

echo "Discovering OIDC configuration for: $PROVIDER_URL"
echo "=========================================="
echo ""

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "⚠️  jq is not installed. Installing it will provide better output."
    echo "   On macOS: brew install jq"
    echo "   On Ubuntu: sudo apt-get install jq"
    echo ""
    echo "Fetching configuration without jq..."
    echo ""
    curl -s "$WELL_KNOWN" | python3 -m json.tool 2>/dev/null || curl -s "$WELL_KNOWN"
else
    echo "✅ Found OIDC configuration:"
    echo ""
    
    # Fetch and display key information
    CONFIG=$(curl -s "$WELL_KNOWN")
    
    if [ $? -ne 0 ]; then
        echo "❌ Failed to fetch configuration from: $WELL_KNOWN"
        echo ""
        echo "Common issues:"
        echo "  - URL might be incorrect"
        echo "  - Provider might not support OIDC discovery"
        echo "  - Network connectivity issues"
        exit 1
    fi
    
    echo "📋 Key Information:"
    echo "-------------------"
    echo ""
    echo "Issuer URL:"
    echo "$CONFIG" | jq -r '.issuer // "Not found"'
    echo ""
    echo "Authorization Endpoint:"
    echo "$CONFIG" | jq -r '.authorization_endpoint // "Not found"'
    echo ""
    echo "Token Endpoint:"
    echo "$CONFIG" | jq -r '.token_endpoint // "Not found"'
    echo ""
    echo "Supported Scopes:"
    echo "$CONFIG" | jq -r '.scopes_supported // "Not found"'
    echo ""
    echo "Response Types:"
    echo "$CONFIG" | jq -r '.response_types_supported // "Not found"'
    echo ""
    echo "📄 Full Configuration:"
    echo "----------------------"
    echo "$CONFIG" | jq '.'
    echo ""
    echo "💡 Next Steps:"
    echo "  1. Use the 'Issuer URL' above as your --issuer-url"
    echo "  2. Register an OAuth2 client with your provider"
    echo "  3. Set redirect URI to: http://localhost:8000/callback"
    echo "  4. Get your Client ID and Client Secret"
    echo ""
    echo "See OIDC_CREDENTIALS_GUIDE.md for detailed instructions."
fi

