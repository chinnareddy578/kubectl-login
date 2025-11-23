#!/bin/bash

# Test runner script for kubectl-login
# Provides easy commands to run different test suites

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_header() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Parse command line arguments
case "${1:-all}" in
    all)
        print_header "Running All Tests"
        go test -v ./...
        ;;
    unit)
        print_header "Running Unit Tests"
        go test -v -short ./...
        ;;
    config)
        print_header "Running Config Tests"
        go test -v ./pkg/config
        ;;
    cache)
        print_header "Running Cache Tests"
        go test -v ./pkg/cache
        ;;
    auth)
        print_header "Running Auth Tests"
        go test -v ./pkg/auth
        ;;
    integration)
        print_header "Running Integration Tests"
        go test -v ./test
        ;;
    coverage)
        print_header "Running Tests with Coverage"
        go test -v -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
        echo ""
        print_info "Coverage report saved to coverage.out"
        print_info "View HTML report: go tool cover -html=coverage.out"
        ;;
    mock)
        print_header "Testing with Mock OIDC Provider"
        go test -v ./pkg/auth -run Mock
        ;;
    watch)
        print_info "Watching for changes and running tests..."
        if command -v entr &> /dev/null; then
            find . -name "*.go" -not -path "./vendor/*" | entr -c go test -v ./...
        else
            print_error "entr not found. Install it with: brew install entr (macOS) or apt-get install entr (Linux)"
            exit 1
        fi
        ;;
    help|--help|-h)
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  all          - Run all tests (default)"
        echo "  unit         - Run unit tests only (skip integration)"
        echo "  config       - Run config package tests"
        echo "  cache        - Run cache package tests"
        echo "  auth         - Run auth package tests"
        echo "  integration  - Run integration tests"
        echo "  coverage     - Run tests with coverage report"
        echo "  mock         - Test with mock OIDC provider"
        echo "  watch        - Watch for changes and run tests"
        echo "  help         - Show this help message"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0 help' for usage information"
        exit 1
        ;;
esac

print_success "Tests completed!"

