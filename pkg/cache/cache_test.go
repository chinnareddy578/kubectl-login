package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chinnareddy578/kubectl-login/pkg/types"
)

func TestTokenCache_GetSet(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "tokens.json")

	cache := &TokenCache{
		tokens: make(map[string]*types.TokenInfo),
		path:   cachePath,
	}

	issuerURL := "https://test-issuer.com"
	clientID := "test-client-id"

	token := &types.TokenInfo{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		IDToken:      "test-id-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Test Set
	cache.Set(issuerURL, clientID, token)

	// Test Get
	retrieved := cache.Get(issuerURL, clientID)
	if retrieved == nil {
		t.Fatal("Expected token to be cached")
	}

	if retrieved.AccessToken != token.AccessToken {
		t.Errorf("AccessToken mismatch: expected %s, got %s", token.AccessToken, retrieved.AccessToken)
	}
	if retrieved.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken mismatch: expected %s, got %s", token.RefreshToken, retrieved.RefreshToken)
	}
	if retrieved.IDToken != token.IDToken {
		t.Errorf("IDToken mismatch: expected %s, got %s", token.IDToken, retrieved.IDToken)
	}
}

func TestTokenCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "tokens.json")

	cache := &TokenCache{
		tokens: make(map[string]*types.TokenInfo),
		path:   cachePath,
	}

	issuerURL := "https://test-issuer.com"
	clientID := "test-client-id"

	token := &types.TokenInfo{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	cache.Set(issuerURL, clientID, token)

	// Verify it's cached
	if cache.Get(issuerURL, clientID) == nil {
		t.Fatal("Token should be cached")
	}

	// Clear it
	cache.Clear(issuerURL, clientID)

	// Verify it's gone
	if cache.Get(issuerURL, clientID) != nil {
		t.Error("Token should be cleared")
	}
}

func TestTokenCache_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "tokens.json")

	issuerURL := "https://test-issuer.com"
	clientID := "test-client-id"

	token := &types.TokenInfo{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		IDToken:      "test-id-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Create first cache instance and save
	cache1 := &TokenCache{
		tokens: make(map[string]*types.TokenInfo),
		path:   cachePath,
	}
	cache1.Set(issuerURL, clientID, token)

	// Create second cache instance and load
	cache2 := &TokenCache{
		tokens: make(map[string]*types.TokenInfo),
		path:   cachePath,
	}
	cache2.load()

	// Verify token was persisted
	retrieved := cache2.Get(issuerURL, clientID)
	if retrieved == nil {
		t.Fatal("Expected token to be persisted")
	}

	if retrieved.AccessToken != token.AccessToken {
		t.Errorf("AccessToken mismatch: expected %s, got %s", token.AccessToken, retrieved.AccessToken)
	}
}

func TestTokenCache_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(cachePath, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("Failed to write invalid cache: %v", err)
	}

	cache := &TokenCache{
		tokens: make(map[string]*types.TokenInfo),
		path:   cachePath,
	}

	// Should not panic, just ignore invalid file
	cache.load()

	// Cache should be empty
	if len(cache.tokens) != 0 {
		t.Error("Expected empty cache after loading invalid file")
	}
}

func TestTokenCache_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "tokens.json")

	cache := &TokenCache{
		tokens: make(map[string]*types.TokenInfo),
		path:   cachePath,
	}

	issuerURL := "https://test-issuer.com"
	clientID := "test-client-id"

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			token := &types.TokenInfo{
				AccessToken: "token-" + string(rune(id)),
				Expiry:      time.Now().Add(1 * time.Hour),
			}
			cache.Set(issuerURL, clientID, token)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should have a token
	if cache.Get(issuerURL, clientID) == nil {
		t.Error("Expected token after concurrent writes")
	}
}

func TestTokenCache_KeyGeneration(t *testing.T) {
	cache := &TokenCache{
		tokens: make(map[string]*types.TokenInfo),
		path:   "/tmp/test",
	}

	key1 := cache.key("https://issuer1.com", "client1")
	key2 := cache.key("https://issuer2.com", "client1")
	key3 := cache.key("https://issuer1.com", "client2")

	if key1 == key2 {
		t.Error("Different issuers should generate different keys")
	}
	if key1 == key3 {
		t.Error("Different clients should generate different keys")
	}
	if key2 == key3 {
		t.Error("Different issuer/client combinations should generate different keys")
	}
}

