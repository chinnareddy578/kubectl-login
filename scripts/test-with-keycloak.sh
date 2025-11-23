#!/bin/bash

# Test script for kubectl-login with local Keycloak

set -e

KEYCLOAK_URL="http://localhost:8080"
REALM_NAME="kubectl-login"
CLIENT_ID="kubectl-login-client"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check if Keycloak is running
check_keycloak() {
    if ! curl -s -f "${KEYCLOAK_URL}/realms/master/.well-known/openid-configuration" > /dev/null 2>&1; then
        print_error "Keycloak is not running"
        print_info "Start it with: docker-compose up -d"
        exit 1
    fi
}

# Get client secret from Keycloak
get_client_secret() {
    # Try to get from config file first
    if [ -f ~/.kubectl-login/config.json ]; then
        CLIENT_SECRET=$(grep -o '"client_secret":"[^"]*' ~/.kubectl-login/config.json | cut -d'"' -f4)
        if [ ! -z "$CLIENT_SECRET" ]; then
            echo "$CLIENT_SECRET"
            return
        fi
    fi

    print_error "Client secret not found. Run setup-keycloak.sh first"
    exit 1
}

# Test authentication
test_auth() {
    print_info "Testing kubectl-login with Keycloak..."
    echo ""

    CLIENT_SECRET=$(get_client_secret)
    ISSUER_URL="${KEYCLOAK_URL}/realms/${REALM_NAME}"

    print_info "Issuer URL: ${ISSUER_URL}"
    print_info "Client ID: ${CLIENT_ID}"
    echo ""

    # Build if needed
    if [ ! -f ./kubectl-login ]; then
        print_info "Building kubectl-login..."
        go build -o kubectl-login .
    fi

    # Run authentication
    print_info "Starting authentication..."
    echo ""
    ./kubectl-login \
        --issuer-url "${ISSUER_URL}" \
        --client-id "${CLIENT_ID}" \
        --client-secret "${CLIENT_SECRET}"

    if [ $? -eq 0 ]; then
        print_success "Authentication successful!"
    else
        print_error "Authentication failed"
        exit 1
    fi
}

# Test exec credential plugin mode
test_exec_credential() {
    print_info "Testing exec credential plugin mode..."
    echo ""

    CLIENT_SECRET=$(get_client_secret)
    ISSUER_URL="${KEYCLOAK_URL}/realms/${REALM_NAME}"

    # Create exec credential request
    REQUEST=$(cat <<EOF
{
  "apiVersion": "client.authentication.k8s.io/v1beta1",
  "kind": "ExecCredential",
  "spec": {}
}
EOF
)

    # Test exec credential
    RESPONSE=$(echo "$REQUEST" | ./kubectl-login \
        --issuer-url "${ISSUER_URL}" \
        --client-id "${CLIENT_ID}" \
        --client-secret "${CLIENT_SECRET}" 2>&1)

    if echo "$RESPONSE" | grep -q "token"; then
        print_success "Exec credential plugin working!"
        echo ""
        echo "Response:"
        echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    else
        print_error "Exec credential plugin test failed"
        echo "Response: $RESPONSE"
        exit 1
    fi
}

# Main
case "${1:-auth}" in
    auth)
        check_keycloak
        test_auth
        ;;
    exec)
        check_keycloak
        test_exec_credential
        ;;
    both)
        check_keycloak
        test_auth
        echo ""
        test_exec_credential
        ;;
    *)
        echo "Usage: $0 [auth|exec|both]"
        echo "  auth - Test browser authentication (default)"
        echo "  exec - Test exec credential plugin"
        echo "  both - Test both modes"
        exit 1
        ;;
esac

