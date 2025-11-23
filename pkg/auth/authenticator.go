package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/chinnareddy578/kubectl-login/pkg/config"
	"github.com/chinnareddy578/kubectl-login/pkg/types"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Authenticator handles OIDC authentication
type Authenticator struct {
	config *config.Config
	ctx    context.Context
}

// NewAuthenticator creates a new authenticator instance
func NewAuthenticator(cfg *config.Config) *Authenticator {
	return &Authenticator{
		config: cfg,
		ctx:    context.Background(),
	}
}

// Authenticate performs the authentication flow
func (a *Authenticator) Authenticate() (*types.TokenInfo, error) {
	// Perform new authentication
	var token *types.TokenInfo
	var err error

	if a.config.Headless {
		token, err = a.authenticateHeadless()
	} else {
		token, err = a.authenticateBrowser()
	}

	if err != nil {
		return nil, err
	}

	return token, nil
}

// authenticateBrowser performs browser-based authentication
func (a *Authenticator) authenticateBrowser() (*types.TokenInfo, error) {
	provider, err := oidc.NewProvider(a.ctx, a.config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oidcConfig := &oidc.Config{
		ClientID: a.config.ClientID,
	}

	verifier := provider.Verifier(oidcConfig)

	redirectURL := fmt.Sprintf("http://localhost:%d/callback", a.config.Port)
	oauth2Config := &oauth2.Config{
		ClientID:     a.config.ClientID,
		ClientSecret: a.config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
	}

	// Generate state and PKCE code verifier
	state, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Generate PKCE code verifier (43-128 characters, URL-safe)
	codeVerifierBytes := make([]byte, 32)
	if _, err := rand.Read(codeVerifierBytes); err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeVerifier := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(codeVerifierBytes)

	// Generate code challenge using S256 (SHA256)
	codeChallengeBytes := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(codeChallengeBytes[:])

	// Start local server for callback
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.Port),
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Log the callback for debugging
		fmt.Fprintf(os.Stderr, "Callback received: %s %s\n", r.Method, r.URL.String())
		
		// Check for errors from OAuth provider
		if errorParam := r.URL.Query().Get("error"); errorParam != "" {
			errorDesc := r.URL.Query().Get("error_description")
			errMsg := fmt.Sprintf("OAuth error: %s", errorParam)
			if errorDesc != "" {
				errMsg += fmt.Sprintf(" - %s", errorDesc)
			}
			errChan <- fmt.Errorf(errMsg)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Authentication failed: %s", errMsg)))
			return
		}

		// Check state parameter
		receivedState := r.URL.Query().Get("state")
		if receivedState != state {
			errChan <- fmt.Errorf("invalid state parameter: expected %s, got %s", state, receivedState)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid state parameter"))
			return
		}

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			// Log all query parameters for debugging
			fmt.Fprintf(os.Stderr, "No code in callback. Query params: %v\n", r.URL.Query())
			errChan <- fmt.Errorf("no authorization code received in callback")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No authorization code received. Please try again."))
			return
		}

		codeChan <- code
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authentication successful! You can close this window."))
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Build authorization URL with PKCE (S256 method)
	authURL := oauth2Config.AuthCodeURL(state, 
		oauth2.AccessTypeOffline, 
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))

	// Open browser
	if err := openBrowser(authURL); err != nil {
		server.Close()
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Opening browser for authentication...\n")
	fmt.Fprintf(os.Stderr, "If the browser doesn't open, visit: %s\n", authURL)

	// Wait for callback
	select {
	case code := <-codeChan:
		server.Close()
		// Exchange code for token
		token, err := oauth2Config.Exchange(a.ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
		if err != nil {
			return nil, fmt.Errorf("failed to exchange code for token: %w", err)
		}

		// Extract ID token
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			return nil, fmt.Errorf("no id_token in token response")
		}

		// Verify ID token
		idToken, err := verifier.Verify(a.ctx, rawIDToken)
		if err != nil {
			return nil, fmt.Errorf("failed to verify ID token: %w", err)
		}

		// Extract user info
		var claims struct {
			Email string `json:"email"`
		}
		if err := idToken.Claims(&claims); err != nil {
			return nil, fmt.Errorf("failed to extract claims: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Successfully authenticated as: %s\n", claims.Email)

		return &types.TokenInfo{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			IDToken:      rawIDToken,
			Expiry:       token.Expiry,
		}, nil

	case err := <-errChan:
		server.Close()
		return nil, err

	case <-time.After(5 * time.Minute):
		server.Close()
		return nil, fmt.Errorf("authentication timeout")
	}
}

// authenticateHeadless performs headless authentication (for CI/CD)
func (a *Authenticator) authenticateHeadless() (*types.TokenInfo, error) {
	provider, err := oidc.NewProvider(a.ctx, a.config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oidcConfig := &oidc.Config{
		ClientID: a.config.ClientID,
	}

	verifier := provider.Verifier(oidcConfig)

	oauth2Config := &oauth2.Config{
		ClientID:     a.config.ClientID,
		ClientSecret: a.config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
	}

	// For headless mode, try client credentials first if secret is available
	// Otherwise, attempt device flow
	if a.config.ClientSecret != "" {
		// Try client credentials flow
		if token, err := a.clientCredentialsFlow(oauth2Config); err == nil {
			return token, nil
		}
		fmt.Fprintf(os.Stderr, "Client credentials flow failed, trying device flow...\n")
	}

	// Try device flow (endpoints may vary by provider)
	// Common device flow endpoints
	deviceEndpoints := []string{
		a.config.IssuerURL + "/device",
		a.config.IssuerURL + "/oauth2/device",
		a.config.IssuerURL + "/v1/device",
	}

	for _, deviceAuthURL := range deviceEndpoints {
		if token, err := a.deviceFlow(deviceAuthURL, oauth2Config, verifier); err == nil {
			return token, nil
		}
	}

	return nil, fmt.Errorf("headless authentication failed: device flow not supported or client credentials invalid. Please use browser mode or configure device flow endpoints")
}

// deviceFlow implements OAuth2 device flow for headless authentication
func (a *Authenticator) deviceFlow(deviceAuthURL string, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier) (*types.TokenInfo, error) {
	// Request device code
	resp, err := http.PostForm(deviceAuthURL, url.Values{
		"client_id": {a.config.ClientID},
		"scope":     {"openid profile email offline_access"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed with status %d", resp.StatusCode)
	}

	var deviceResp struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		Interval        int    `json:"interval"`
		ExpiresIn       int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "Please visit: %s\n", deviceResp.VerificationURI)
	fmt.Fprintf(os.Stderr, "Enter code: %s\n", deviceResp.UserCode)

	// Poll for token
	tokenURL := a.config.IssuerURL + "/token"
	interval := time.Duration(deviceResp.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	expiresAt := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)

	for time.Now().Before(expiresAt) {
		time.Sleep(interval)

		resp, err := http.PostForm(tokenURL, url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code": {deviceResp.DeviceCode},
			"client_id":   {a.config.ClientID},
		})
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var tokenResp struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				IDToken      string `json:"id_token"`
				ExpiresIn    int    `json:"expires_in"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			// Verify ID token
			idToken, err := verifier.Verify(a.ctx, tokenResp.IDToken)
			if err != nil {
				return nil, fmt.Errorf("failed to verify ID token: %w", err)
			}

			var claims struct {
				Email string `json:"email"`
			}
			if err := idToken.Claims(&claims); err != nil {
				return nil, fmt.Errorf("failed to extract claims: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Successfully authenticated as: %s\n", claims.Email)

			return &types.TokenInfo{
				AccessToken:  tokenResp.AccessToken,
				RefreshToken: tokenResp.RefreshToken,
				IDToken:      tokenResp.IDToken,
				Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
			}, nil
		}

		resp.Body.Close()
	}

	return nil, fmt.Errorf("device flow authentication timeout")
}

// clientCredentialsFlow implements OAuth2 client credentials flow
func (a *Authenticator) clientCredentialsFlow(oauth2Config *oauth2.Config) (*types.TokenInfo, error) {
	// Client credentials flow requires a custom token endpoint request
	// since oauth2.Config doesn't directly support client credentials grant
	tokenURL := oauth2Config.Endpoint.TokenURL
	if tokenURL == "" {
		return nil, fmt.Errorf("token URL not available from OIDC provider")
	}

	// Make client credentials request
	resp, err := http.PostForm(tokenURL, url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {a.config.ClientID},
		"client_secret": {a.config.ClientSecret},
		"scope":         {"openid profile email"},
	})
	if err != nil {
		return nil, fmt.Errorf("client credentials request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client credentials flow failed with status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &types.TokenInfo{
		AccessToken: tokenResp.AccessToken,
		Expiry:      expiry,
	}, nil
}

// RefreshToken refreshes an expired token
func (a *Authenticator) RefreshToken(refreshToken string) (*types.TokenInfo, error) {
	provider, err := oidc.NewProvider(a.ctx, a.config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     a.config.ClientID,
		ClientSecret: a.config.ClientSecret,
		Endpoint:     provider.Endpoint(),
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	newToken, err := oauth2Config.TokenSource(a.ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &types.TokenInfo{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		Expiry:       newToken.Expiry,
	}, nil
}

// generateRandomString generates a random string for state and PKCE
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}
