#!/bin/bash

# Setup script for Keycloak OIDC server
# This script configures Keycloak with a client for kubectl-login testing

set -e

KEYCLOAK_URL="http://localhost:8080"
ADMIN_USER="admin"
ADMIN_PASSWORD="admin"
REALM_NAME="kubectl-login"
CLIENT_ID="kubectl-login-client"
REDIRECT_URI="http://localhost:8000/callback"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
    print_info "Checking if Keycloak is running..."
    
    # Try multiple endpoints to check if Keycloak is ready
    if curl -s -f "${KEYCLOAK_URL}/realms/master/.well-known/openid-configuration" > /dev/null 2>&1; then
        print_success "Keycloak is running"
        return
    fi
    
    # Try health endpoint (older versions)
    if curl -s -f "${KEYCLOAK_URL}/health/ready" > /dev/null 2>&1; then
        print_success "Keycloak is running"
        return
    fi
    
    # Try root endpoint
    if curl -s -f "${KEYCLOAK_URL}" > /dev/null 2>&1; then
        print_success "Keycloak is running"
        return
    fi
    
    print_error "Keycloak is not running or not ready"
    print_info "Start it with: docker-compose up -d"
    print_info "Wait about 30-60 seconds for Keycloak to fully start"
    exit 1
}

# Get admin token
get_admin_token() {
    print_info "Getting admin token..." >&2
    TOKEN_RESPONSE=$(curl -s -X POST "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=${ADMIN_USER}" \
        -d "password=${ADMIN_PASSWORD}" \
        -d "grant_type=password" \
        -d "client_id=admin-cli")

    ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

    if [ -z "$ACCESS_TOKEN" ]; then
        print_error "Failed to get admin token" >&2
        echo "Response: $TOKEN_RESPONSE" >&2
        exit 1
    fi

    print_success "Admin token obtained" >&2
    echo "$ACCESS_TOKEN"
}

# Create realm
create_realm() {
    TOKEN=$1
    print_info "Creating realm: ${REALM_NAME}..."

    # Check if realm already exists first
    CHECK_HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}" \
        -H "Authorization: Bearer ${TOKEN}")
    
    if [ "$CHECK_HTTP_CODE" == "200" ]; then
        print_warning "Realm already exists, skipping creation"
        return
    fi

    REALM_CONFIG=$(cat <<EOF
{
  "realm": "${REALM_NAME}",
  "enabled": true
}
EOF
)

    # Try to create realm
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/admin/realms" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "$REALM_CONFIG" 2>&1)

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')

    if [ "$HTTP_CODE" == "201" ]; then
        print_success "Realm created"
    elif [ "$HTTP_CODE" == "409" ] || [ "$HTTP_CODE" == "400" ]; then
        # Double-check if it exists now
        CHECK_AGAIN=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}" \
            -H "Authorization: Bearer ${TOKEN}")
        if [ "$CHECK_AGAIN" == "200" ]; then
            print_warning "Realm already exists, skipping creation"
        else
            print_error "Failed to create realm (HTTP $HTTP_CODE)"
            if [ ! -z "$BODY" ] && [ "$BODY" != "null" ]; then
                echo "Response: $BODY"
            fi
            exit 1
        fi
    else
        print_error "Failed to create realm (HTTP $HTTP_CODE)"
        if [ ! -z "$BODY" ] && [ "$BODY" != "null" ]; then
            echo "Response: $BODY"
        fi
        exit 1
    fi
}

# Create client
create_client() {
    TOKEN=$1
    print_info "Creating client: ${CLIENT_ID}..."

    CLIENT_CONFIG=$(cat <<EOF
{
  "clientId": "${CLIENT_ID}",
  "enabled": true,
  "clientAuthenticatorType": "client-secret",
  "redirectUris": ["${REDIRECT_URI}"],
  "webOrigins": ["http://localhost:8000"],
  "standardFlowEnabled": true,
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": true,
  "publicClient": false,
  "protocol": "openid-connect"
}
EOF
)

    # Check if client already exists
    CHECK_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/clients?clientId=${CLIENT_ID}" \
        -H "Authorization: Bearer ${TOKEN}")
    CHECK_HTTP_CODE=$(echo "$CHECK_RESPONSE" | tail -n1)
    
    if [ "$CHECK_HTTP_CODE" == "200" ]; then
        CLIENT_LIST=$(echo "$CHECK_RESPONSE" | sed '$d')
        if echo "$CLIENT_LIST" | grep -q "\"clientId\":\"${CLIENT_ID}\""; then
            print_warning "Client already exists, skipping creation"
            return
        fi
    fi

    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/clients" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "$CLIENT_CONFIG")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')

    if [ "$HTTP_CODE" == "201" ]; then
        print_success "Client created"
    elif [ "$HTTP_CODE" == "409" ]; then
        print_warning "Client already exists, skipping creation"
    else
        print_error "Failed to create client (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
        exit 1
    fi
}

# Get client secret
get_client_secret() {
    TOKEN=$1
    print_info "Getting client secret..." >&2

    # Get client ID (UUID)
    CLIENTS_RESPONSE=$(curl -s -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/clients?clientId=${CLIENT_ID}" \
        -H "Authorization: Bearer ${TOKEN}")

    CLIENT_UUID=$(echo "$CLIENTS_RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

    if [ -z "$CLIENT_UUID" ]; then
        print_error "Failed to find client UUID" >&2
        exit 1
    fi

    # Get client secret
    SECRET_RESPONSE=$(curl -s -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/clients/${CLIENT_UUID}/client-secret" \
        -H "Authorization: Bearer ${TOKEN}")

    CLIENT_SECRET=$(echo "$SECRET_RESPONSE" | grep -o '"value":"[^"]*' | cut -d'"' -f4)

    if [ -z "$CLIENT_SECRET" ]; then
        print_error "Failed to get client secret" >&2
        echo "Response: $SECRET_RESPONSE" >&2
        exit 1
    fi

    print_success "Client secret obtained" >&2
    echo "$CLIENT_SECRET"
}

# Create test user
create_test_user() {
    TOKEN=$1
    print_info "Creating test user..."

    USER_CONFIG=$(cat <<EOF
{
  "username": "testuser",
  "enabled": true,
  "email": "testuser@example.com",
  "firstName": "Test",
  "lastName": "User",
  "credentials": [{
    "type": "password",
    "value": "testpassword",
    "temporary": false
  }]
}
EOF
)

    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "$USER_CONFIG")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')

    if [ "$HTTP_CODE" == "201" ] || [ "$HTTP_CODE" == "409" ]; then
        if [ "$HTTP_CODE" == "409" ]; then
            print_warning "User already exists, skipping creation"
        else
            print_success "Test user created (username: testuser, password: testpassword)"
        fi
    else
        print_error "Failed to create user (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
        exit 1
    fi
}

# Main execution
main() {
    echo ""
    echo "=========================================="
    echo "  Keycloak Setup for kubectl-login"
    echo "=========================================="
    echo ""

    check_keycloak
    TOKEN=$(get_admin_token)
    create_realm "$TOKEN"
    create_client "$TOKEN"
    CLIENT_SECRET=$(get_client_secret "$TOKEN")
    create_test_user "$TOKEN"

    echo ""
    echo "=========================================="
    echo "  Setup Complete!"
    echo "=========================================="
    echo ""
    echo "Configuration for kubectl-login:"
    echo ""
    echo "  Issuer URL:    ${KEYCLOAK_URL}/realms/${REALM_NAME}"
    echo "  Client ID:     ${CLIENT_ID}"
    echo "  Client Secret: ${CLIENT_SECRET}"
    echo ""
    echo "Test user credentials:"
    echo "  Username:      testuser"
    echo "  Password:      testpassword"
    echo ""
    echo "Save this configuration to ~/.kubectl-login/config.json:"
    echo ""
    cat <<EOF
{
  "issuer_url": "${KEYCLOAK_URL}/realms/${REALM_NAME}",
  "client_id": "${CLIENT_ID}",
  "client_secret": "${CLIENT_SECRET}",
  "headless": false,
  "port": 8000
}
EOF
    echo ""
    print_info "Keycloak Admin Console: http://localhost:8080"
    print_info "Admin username: ${ADMIN_USER}"
    print_info "Admin password: ${ADMIN_PASSWORD}"
    echo ""
}

main

