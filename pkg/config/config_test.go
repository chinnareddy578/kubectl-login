package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
  "issuer_url": "https://test-issuer.com",
  "client_id": "test-client-id",
  "client_secret": "test-secret",
  "headless": true,
  "port": 9000
}`

	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if cfg.IssuerURL != "https://test-issuer.com" {
		t.Errorf("Expected issuer_url 'https://test-issuer.com', got '%s'", cfg.IssuerURL)
	}
	if cfg.ClientID != "test-client-id" {
		t.Errorf("Expected client_id 'test-client-id', got '%s'", cfg.ClientID)
	}
	if cfg.ClientSecret != "test-secret" {
		t.Errorf("Expected client_secret 'test-secret', got '%s'", cfg.ClientSecret)
	}
	if !cfg.Headless {
		t.Error("Expected headless to be true")
	}
	if cfg.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", cfg.Port)
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(configPath, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadFromFile(configPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestSaveToFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		IssuerURL:    "https://test-issuer.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		Headless:     true,
		Port:         9000,
	}

	if err := SaveToFile(cfg, configPath); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify file permissions (should be 0600)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
	}

	// Load and verify
	loaded, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.IssuerURL != cfg.IssuerURL {
		t.Errorf("IssuerURL mismatch: expected %s, got %s", cfg.IssuerURL, loaded.IssuerURL)
	}
	if loaded.ClientID != cfg.ClientID {
		t.Errorf("ClientID mismatch: expected %s, got %s", cfg.ClientID, loaded.ClientID)
	}
}

func TestSaveToFile_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.json")

	cfg := &Config{}

	if err := SaveToFile(cfg, configPath); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	loaded, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.IssuerURL != "" {
		t.Errorf("Expected empty IssuerURL, got %s", loaded.IssuerURL)
	}
}

