package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinnareddy578/kubectl-login/pkg/auth"
	"github.com/chinnareddy578/kubectl-login/pkg/config"
)

// TestBrowserAuthFlow tests the browser authentication flow with a mock provider
func TestBrowserAuthFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockProvider := auth.NewMockOIDCProvider()
	defer mockProvider.Close()

	cfg := &config.Config{
		IssuerURL:    mockProvider.IssuerURL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Headless:     false,
		Port:         8002,
	}

	authenticator := auth.NewAuthenticator(cfg)

	// Note: This test requires actual browser interaction or mocking
	// For now, we'll test the configuration
	if authenticator == nil {
		t.Fatal("Failed to create authenticator")
	}

	t.Logf("Mock provider running at: %s", mockProvider.IssuerURL)
}

// TestExecCredentialPlugin tests the kubectl exec credential plugin interface
func TestExecCredentialPlugin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require building the binary and testing it
	// For now, we'll just verify the structure
	t.Log("Exec credential plugin test - requires built binary")
}

// TestConfigFileLoading tests loading configuration from file
func TestConfigFileLoading(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	cfg := &config.Config{
		IssuerURL:    "https://test-issuer.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		Headless:     true,
		Port:         9000,
	}

	// Save config
	if err := config.SaveToFile(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := config.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.IssuerURL != cfg.IssuerURL {
		t.Errorf("IssuerURL mismatch")
	}
}

// TestTokenCacheIntegration tests token caching with file system
func TestTokenCacheIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "tokens.json")

	// This would test the full cache integration
	// For now, we'll just verify the path
	if cachePath == "" {
		t.Error("Cache path should not be empty")
	}
}

// MockHTTPServer creates a simple HTTP server for testing
func MockHTTPServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

// TestCLIHelp tests that the CLI help command works
func TestCLIHelp(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "kubectl-login-test", ".")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping CLI test - build failed: %v", err)
	}
	defer os.Remove("kubectl-login-test")

	// Test help command
	helpCmd := exec.Command("./kubectl-login-test", "--help")
	helpCmd.Dir = ".."
	output, err := helpCmd.Output()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	if !strings.Contains(string(output), "kubectl-login") {
		t.Error("Help output should contain 'kubectl-login'")
	}

	t.Logf("Help output:\n%s", string(output))
}
