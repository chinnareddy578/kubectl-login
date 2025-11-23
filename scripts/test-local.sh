#!/bin/bash

# Local test script for kubectl-login with Keycloak

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
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

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check if Keycloak is running
check_keycloak() {
    if ! curl -s -f http://localhost:8080/realms/master/.well-known/openid-configuration > /dev/null 2>&1; then
        print_error "Keycloak is not running"
        print_info "Start it with: docker-compose up -d"
        exit 1
    fi
}

# Get client secret
get_client_secret() {
    TOKEN=$(curl -s -X POST "http://localhost:8080/realms/master/protocol/openid-connect/token" \
        -d "username=admin" \
        -d "password=admin" \
        -d "grant_type=password" \
        -d "client_id=admin-cli" | jq -r '.access_token')
    
    CLIENTS=$(curl -s -X GET "http://localhost:8080/admin/realms/kubectl-login/clients?clientId=kubectl-login-client" \
        -H "Authorization: Bearer $TOKEN")
    
    CLIENT_UUID=$(echo "$CLIENTS" | jq -r '.[0].id')
    
    if [ "$CLIENT_UUID" == "null" ] || [ -z "$CLIENT_UUID" ]; then
        print_error "Client not found. Run setup-keycloak.sh first"
        exit 1
    fi
    
    SECRET=$(curl -s -X GET "http://localhost:8080/admin/realms/kubectl-login/clients/$CLIENT_UUID/client-secret" \
        -H "Authorization: Bearer $TOKEN" | jq -r '.value')
    
    echo "$SECRET"
}

# Create config file
create_config() {
    CLIENT_SECRET=$(get_client_secret)
    
    CONFIG_DIR="$HOME/.kubectl-login"
    CONFIG_FILE="$CONFIG_DIR/config.json"
    
    mkdir -p "$CONFIG_DIR"
    
    cat > "$CONFIG_FILE" <<EOF
{
  "issuer_url": "http://localhost:8080/realms/kubectl-login",
  "client_id": "kubectl-login-client",
  "client_secret": "$CLIENT_SECRET",
  "headless": false,
  "port": 8000
}
EOF
    
    chmod 600 "$CONFIG_FILE"
    print_success "Config file created: $CONFIG_FILE" >&2
    echo "$CONFIG_FILE"
}

# Test browser authentication
test_browser_auth() {
    print_info "Testing browser authentication..."
    echo ""
    
    if [ ! -f ./kubectl-login ]; then
        print_info "Building kubectl-login..."
        go build -o kubectl-login .
    fi
    
    CONFIG_FILE=$(create_config)
    
    print_info "Starting authentication..."
    print_info "A browser window should open. Use credentials: testuser / testpassword"
    echo ""
    
    ./kubectl-login --config "$CONFIG_FILE"
    
    if [ $? -eq 0 ]; then
        print_success "Authentication successful!"
        echo ""
        print_info "Token cached at: ~/.cache/kubectl-login/tokens.json"
    else
        print_error "Authentication failed"
        exit 1
    fi
}

# Test exec credential plugin
test_exec_credential() {
    print_info "Testing exec credential plugin mode..."
    echo ""
    
    if [ ! -f ./kubectl-login ]; then
        print_info "Building kubectl-login..."
        go build -o kubectl-login .
    fi
    
    CONFIG_FILE=$(create_config)
    CLIENT_SECRET=$(get_client_secret)
    
    REQUEST=$(cat <<EOF
{
  "apiVersion": "client.authentication.k8s.io/v1beta1",
  "kind": "ExecCredential",
  "spec": {}
}
EOF
)
    
    print_info "Sending exec credential request..."
    RESPONSE=$(echo "$REQUEST" | ./kubectl-login \
        --issuer-url "http://localhost:8080/realms/kubectl-login" \
        --client-id "kubectl-login-client" \
        --client-secret "$CLIENT_SECRET" 2>&1)
    
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
main() {
    echo ""
    echo "=========================================="
    echo "  Local Test: kubectl-login + Keycloak"
    echo "=========================================="
    echo ""
    
    check_keycloak
    
    case "${1:-browser}" in
        browser)
            test_browser_auth
            ;;
        exec)
            test_exec_credential
            ;;
        both)
            test_browser_auth
            echo ""
            echo "----------------------------------------"
            echo ""
            test_exec_credential
            ;;
        config)
            create_config
            print_info "Config file ready for manual testing" >&2
            ;;
        *)
            echo "Usage: $0 [browser|exec|both|config]"
            echo ""
            echo "  browser - Test browser authentication (default)"
            echo "  exec    - Test exec credential plugin"
            echo "  both    - Test both modes"
            echo "  config  - Just create config file"
            exit 1
            ;;
    esac
}

main "$@"

