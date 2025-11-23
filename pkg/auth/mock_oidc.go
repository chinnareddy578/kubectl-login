package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// MockOIDCProvider is a mock OIDC provider for testing
type MockOIDCProvider struct {
	server           *httptest.Server
	IssuerURL        string
	ClientID         string
	ClientSecret     string
	AuthorizationURL string
	TokenURL         string
	UserInfo         map[string]interface{}
	Tokens           map[string]*MockToken
}

// MockToken represents a mock token response
type MockToken struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	TokenType    string
}

// NewMockOIDCProvider creates a new mock OIDC provider
func NewMockOIDCProvider() *MockOIDCProvider {
	mock := &MockOIDCProvider{
		UserInfo: map[string]interface{}{
			"sub":   "test-user-123",
			"email": "test@example.com",
			"name":  "Test User",
		},
		Tokens: make(map[string]*MockToken),
	}

	mux := http.NewServeMux()

	// Well-known configuration endpoint
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"issuer":                 mock.server.URL,
			"authorization_endpoint": mock.server.URL + "/authorize",
			"token_endpoint":         mock.server.URL + "/token",
			"userinfo_endpoint":       mock.server.URL + "/userinfo",
			"jwks_uri":               mock.server.URL + "/.well-known/jwks.json",
			"response_types_supported": []string{"code"},
			"subject_types_supported": []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	})

	// Authorization endpoint
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		redirectURI := r.URL.Query().Get("redirect_uri")
		code := fmt.Sprintf("mock-auth-code-%d", time.Now().UnixNano())

		// Store code for token exchange
		mock.Tokens[code] = &MockToken{
			AccessToken:  "mock-access-token-" + code,
			RefreshToken: "mock-refresh-token-" + code,
			IDToken:      mock.generateIDToken(),
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		// Redirect back with code
		redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	// Token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		code := r.FormValue("code")
		grantType := r.FormValue("grant_type")

		var token *MockToken
		var ok bool

		if grantType == "authorization_code" {
			token, ok = mock.Tokens[code]
			if !ok {
				http.Error(w, "Invalid authorization code", http.StatusBadRequest)
				return
			}
		} else if grantType == "refresh_token" {
			refreshToken := r.FormValue("refresh_token")
			// Find token by refresh token
			for _, t := range mock.Tokens {
				if t.RefreshToken == refreshToken {
					// Generate new tokens
					token = &MockToken{
						AccessToken:  "refreshed-access-token",
						RefreshToken: refreshToken,
						IDToken:      mock.generateIDToken(),
						ExpiresIn:    3600,
						TokenType:    "Bearer",
					}
					break
				}
			}
			if token == nil {
				http.Error(w, "Invalid refresh token", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Unsupported grant type", http.StatusBadRequest)
			return
		}

		response := map[string]interface{}{
			"access_token":  token.AccessToken,
			"refresh_token": token.RefreshToken,
			"id_token":      token.IDToken,
			"token_type":    token.TokenType,
			"expires_in":    token.ExpiresIn,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// UserInfo endpoint
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mock.UserInfo)
	})

	// JWKS endpoint (simplified)
	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		jwks := map[string]interface{}{
			"keys": []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	mock.server = httptest.NewServer(mux)
	mock.IssuerURL = mock.server.URL
	mock.AuthorizationURL = mock.server.URL + "/authorize"
	mock.TokenURL = mock.server.URL + "/token"

	return mock
}

// Close shuts down the mock server
func (m *MockOIDCProvider) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// generateIDToken generates a simple mock ID token
func (m *MockOIDCProvider) generateIDToken() string {
	// In a real implementation, this would be a properly signed JWT
	// For testing, we'll return a simple string that can be verified
	claims := map[string]interface{}{
		"iss":   m.IssuerURL,
		"sub":   m.UserInfo["sub"],
		"aud":   "test-client-id",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"email": m.UserInfo["email"],
		"name":  m.UserInfo["name"],
	}

	data, _ := json.Marshal(claims)
	return fmt.Sprintf("mock.id.token.%s", string(data))
}

// GetOAuth2Config returns an OAuth2 config for the mock provider
func (m *MockOIDCProvider) GetOAuth2Config(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  m.AuthorizationURL,
			TokenURL: m.TokenURL,
		},
		RedirectURL: redirectURL,
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
	}
}

