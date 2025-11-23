#!/bin/bash

# Quick test script - starts everything and tests kubectl-login

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Quick Test: kubectl-login + Keycloak${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Error: Docker Compose is not installed"
    exit 1
fi

# Start Keycloak
echo -e "${YELLOW}Step 1: Starting Keycloak...${NC}"
docker-compose up -d

echo -e "${YELLOW}Step 2: Waiting for Keycloak to be ready...${NC}"
echo "This may take 30-60 seconds..."
for i in {1..60}; do
    if curl -s -f http://localhost:8080/health/ready > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Keycloak is ready!${NC}"
        break
    fi
    if [ $i -eq 60 ]; then
        echo "Error: Keycloak did not become ready in time"
        docker-compose logs keycloak | tail -20
        exit 1
    fi
    sleep 1
    echo -n "."
done
echo ""

# Setup Keycloak
echo -e "${YELLOW}Step 3: Setting up Keycloak...${NC}"
./scripts/setup-keycloak.sh

# Build kubectl-login if needed
if [ ! -f ./kubectl-login ]; then
    echo -e "${YELLOW}Step 4: Building kubectl-login...${NC}"
    go build -o kubectl-login .
fi

# Test
echo ""
echo -e "${YELLOW}Step 5: Testing authentication...${NC}"
echo ""
./scripts/test-with-keycloak.sh auth

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  All tests completed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Keycloak Admin Console: http://localhost:8080"
echo "Admin: admin / admin"
echo ""
echo "Test user: testuser / testpassword"
echo ""
echo "To stop Keycloak: docker-compose down"

