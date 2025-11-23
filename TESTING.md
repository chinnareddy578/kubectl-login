# Testing Guide

This guide explains how to test the kubectl-login plugin locally using unit tests, integration tests, and a mock OIDC server.

## Quick Start

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...

# Run specific test package
go test -v ./pkg/config
go test -v ./pkg/cache
go test -v ./pkg/auth

# Run integration tests (requires more setup)
go test -v ./test -run Integration
```

## Test Structure

### Unit Tests

Unit tests are located alongside the source code:

- `pkg/config/config_test.go` - Configuration loading/saving tests
- `pkg/cache/cache_test.go` - Token cache tests
- `pkg/auth/authenticator_test.go` - Authentication logic tests

### Integration Tests

Integration tests are in the `test/` directory:

- `test/integration_test.go` - End-to-end integration tests

### Mock OIDC Provider

A mock OIDC provider is available for testing without a real SSO provider:

- `pkg/auth/mock_oidc.go` - Mock OIDC server implementation

## Running Tests

### All Tests

```bash
make test
```

### Specific Package

```bash
go test -v ./pkg/config
go test -v ./pkg/cache
go test -v ./pkg/auth
```

### With Coverage

```bash
go test -v -cover ./...
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Short Tests Only (Skip Integration)

```bash
go test -v -short ./...
```

## Using the Mock OIDC Provider

The mock OIDC provider allows you to test authentication without a real SSO provider:

```go
package main

import (
    "github.com/chinnareddy578/kubectl-login/pkg/auth"
    "github.com/chinnareddy578/kubectl-login/pkg/config"
)

func main() {
    // Create mock provider
    mockProvider := auth.NewMockOIDCProvider()
    defer mockProvider.Close()

    // Use it in your tests
    cfg := &config.Config{
        IssuerURL:    mockProvider.IssuerURL,
        ClientID:     "test-client-id",
        ClientSecret: "test-client-secret",
        Headless:     false,
        Port:         8000,
    }

    authenticator := auth.NewAuthenticator(cfg)
    // ... test authentication
}
```

## Manual Testing

### Test with Mock Provider

1. Start the mock provider in a test:

```go
mockProvider := auth.NewMockOIDCProvider()
defer mockProvider.Close()
```

2. Use the mock provider's URL:

```bash
kubectl login \
  --issuer-url http://localhost:<port> \
  --client-id test-client-id \
  --client-secret test-client-secret
```

### Test with Real Provider

For testing with a real OIDC provider:

```bash
# Set environment variables
export TEST_ISSUER_URL=https://your-oidc-provider.com
export TEST_CLIENT_ID=your-client-id
export TEST_CLIENT_SECRET=your-client-secret

# Run integration tests
go test -v ./pkg/auth -run TestAuthenticator_Integration
```

## Test Scenarios

### Configuration Tests

- ✅ Load config from file
- ✅ Save config to file
- ✅ Handle invalid JSON
- ✅ Handle missing file
- ✅ File permissions (0600)

### Cache Tests

- ✅ Get/Set tokens
- ✅ Clear tokens
- ✅ Persistence to disk
- ✅ Concurrent access
- ✅ Invalid cache file handling

### Authentication Tests

- ✅ Mock OIDC provider
- ✅ Token refresh
- ✅ Random string generation
- ⚠️ Browser flow (requires manual testing)
- ⚠️ Headless flow (requires device flow support)

## Writing New Tests

### Unit Test Example

```go
func TestMyFunction(t *testing.T) {
    // Arrange
    input := "test-input"
    expected := "test-output"

    // Act
    result := MyFunction(input)

    // Assert
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

### Integration Test Example

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Your integration test code
}
```

## Continuous Integration

Tests run automatically in CI (see `.github/workflows/test.yml`):

- Build verification
- Unit tests
- Integration tests (if not skipped)
- Binary verification

## Troubleshooting

### Tests Fail with "connection refused"

- Mock provider might not be starting correctly
- Check that `defer mockProvider.Close()` is called
- Verify port is not already in use

### Tests Timeout

- Integration tests might be waiting for user input
- Use `-short` flag to skip integration tests
- Check for hanging goroutines

### Coverage is Low

- Add tests for error paths
- Test edge cases
- Add integration tests for complex flows

## Best Practices

1. **Use table-driven tests** for multiple test cases
2. **Use `t.Helper()`** in test helpers
3. **Clean up resources** with `t.Cleanup()` or `defer`
4. **Skip integration tests** with `testing.Short()`
5. **Use subtests** with `t.Run()` for organization
6. **Test error cases** not just happy paths

## Example: Complete Test Suite

```go
package auth_test

import (
    "testing"
    "github.com/chinnareddy578/kubectl-login/pkg/auth"
)

func TestAuthenticator(t *testing.T) {
    t.Run("NewAuthenticator", func(t *testing.T) {
        // Test new authenticator creation
    })

    t.Run("Authenticate", func(t *testing.T) {
        if testing.Short() {
            t.Skip("Skipping in short mode")
        }
        // Test authentication
    })

    t.Run("RefreshToken", func(t *testing.T) {
        // Test token refresh
    })
}
```

## Next Steps

- Add more unit tests for edge cases
- Improve mock OIDC provider (add JWT signing)
- Add end-to-end tests with real kubectl
- Add performance/benchmark tests
- Add fuzzing tests for input validation

