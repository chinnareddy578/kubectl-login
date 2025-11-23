package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/chinnareddy578/kubectl-login/pkg/types"
)

// TokenCache manages cached authentication tokens
type TokenCache struct {
	mu     sync.RWMutex
	tokens map[string]*types.TokenInfo
	path   string
}

// NewTokenCache creates a new token cache instance
func NewTokenCache() *TokenCache {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	cachePath := filepath.Join(cacheDir, "kubectl-login", "tokens.json")

	cache := &TokenCache{
		tokens: make(map[string]*types.TokenInfo),
		path:   cachePath,
	}

	// Load existing cache
	cache.load()

	return cache
}

// Get retrieves a cached token
func (c *TokenCache) Get(issuerURL, clientID string) *types.TokenInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.key(issuerURL, clientID)
	return c.tokens[key]
}

// Set stores a token in the cache
func (c *TokenCache) Set(issuerURL, clientID string, token *types.TokenInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.key(issuerURL, clientID)
	c.tokens[key] = token

	// Persist to disk
	c.save()
}

// Clear removes a token from the cache
func (c *TokenCache) Clear(issuerURL, clientID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.key(issuerURL, clientID)
	delete(c.tokens, key)

	// Persist to disk
	c.save()
}

// key generates a cache key from issuer URL and client ID
func (c *TokenCache) key(issuerURL, clientID string) string {
	return issuerURL + ":" + clientID
}

// load reads the cache from disk
func (c *TokenCache) load() {
	data, err := os.ReadFile(c.path)
	if err != nil {
		// Cache file doesn't exist yet, that's okay
		return
	}

	var cacheData map[string]*cacheEntry
	if err := json.Unmarshal(data, &cacheData); err != nil {
		// Invalid cache file, ignore it
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range cacheData {
		c.tokens[key] = &types.TokenInfo{
			AccessToken:  entry.AccessToken,
			RefreshToken: entry.RefreshToken,
			IDToken:      entry.IDToken,
			Expiry:       entry.Expiry,
		}
	}
}

// save writes the cache to disk
func (c *TokenCache) save() {
	cacheDir := filepath.Dir(c.path)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		// Failed to create cache directory, skip saving
		return
	}

	cacheData := make(map[string]*cacheEntry)
	for key, token := range c.tokens {
		cacheData[key] = &cacheEntry{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			IDToken:      token.IDToken,
			Expiry:       token.Expiry,
		}
	}

	data, err := json.MarshalIndent(cacheData, "", "  ")
	if err != nil {
		// Failed to marshal, skip saving
		return
	}

	// Write to temporary file first, then rename (atomic operation)
	tmpPath := c.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return
	}

	os.Rename(tmpPath, c.path)
}

// cacheEntry is used for JSON serialization
type cacheEntry struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	Expiry       time.Time `json:"expiry"`
}
