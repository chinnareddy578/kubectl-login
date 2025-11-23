package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chinnareddy578/kubectl-login/pkg/auth"
	"github.com/chinnareddy578/kubectl-login/pkg/cache"
	"github.com/chinnareddy578/kubectl-login/pkg/config"
	"github.com/chinnareddy578/kubectl-login/pkg/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
)

var (
	issuerURL    string
	clientID     string
	clientSecret string
	headless     bool
	port         int
	configFile   string
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-login",
	Short: "kubectl login plugin for SSO authentication",
	Long: `kubectl-login is a kubectl plugin that provides SSO authentication
using OIDC. It supports both browser-based and headless authentication modes.`,
	RunE: runLogin,
}

func init() {
	rootCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "OIDC issuer URL (required if --config not used)")
	rootCmd.Flags().StringVar(&clientID, "client-id", "", "OIDC client ID (required if --config not used)")
	rootCmd.Flags().StringVar(&clientSecret, "client-secret", "", "OIDC client secret (optional, can be set via CLIENT_SECRET env var)")
	rootCmd.Flags().BoolVar(&headless, "headless", false, "Use headless authentication (for CI/CD)")
	rootCmd.Flags().IntVar(&port, "port", 8000, "Local port for OAuth callback")
	rootCmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")
}

func Execute() error {
	return rootCmd.Execute()
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Check if we're being called as an exec credential plugin
	if isExecCredentialMode() {
		return handleExecCredential()
	}

	// Validate that either config file or required flags are provided
	if configFile == "" {
		if issuerURL == "" {
			return fmt.Errorf("required flag(s) \"issuer-url\" not set (or use --config)")
		}
		if clientID == "" {
			return fmt.Errorf("required flag(s) \"client-id\" not set (or use --config)")
		}
	}

	// Otherwise, run as a regular login command
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check cache first
	cache := cache.NewTokenCache()
	if cached := cache.Get(cfg.IssuerURL, cfg.ClientID); cached != nil {
		if time.Until(cached.Expiry) > 5*time.Minute {
			fmt.Printf("Using cached token (expires in %v)\n", time.Until(cached.Expiry))
			return nil
		}
		// Try to refresh if token is expiring soon
		if cached.RefreshToken != "" {
			authenticator := auth.NewAuthenticator(cfg)
			if refreshed, err := authenticator.RefreshToken(cached.RefreshToken); err == nil {
				cache.Set(cfg.IssuerURL, cfg.ClientID, refreshed)
				fmt.Printf("Token refreshed! Expires in %v\n", time.Until(refreshed.Expiry))
				return nil
			}
		}
	}

	authenticator := auth.NewAuthenticator(cfg)

	token, err := authenticator.Authenticate()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Cache the token
	cache.Set(cfg.IssuerURL, cfg.ClientID, token)

	fmt.Printf("Successfully authenticated! Token expires in %v\n", time.Until(token.Expiry))
	fmt.Println("You can now use kubectl commands.")

	return nil
}

func isExecCredentialMode() bool {
	// kubectl exec credential plugins receive input via stdin
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func handleExecCredential() error {
	// Read the exec credential request from stdin
	var request clientauthv1beta1.ExecCredential
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&request); err != nil {
		return fmt.Errorf("failed to decode exec credential request: %w", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check cache first
	var token *types.TokenInfo
	tokenCache := cache.NewTokenCache()
	if cached := tokenCache.Get(cfg.IssuerURL, cfg.ClientID); cached != nil {
		if time.Until(cached.Expiry) > 5*time.Minute {
			// Use cached token
			token = cached
		} else if cached.RefreshToken != "" {
			// Try to refresh
			authenticator := auth.NewAuthenticator(cfg)
			if refreshed, err := authenticator.RefreshToken(cached.RefreshToken); err == nil {
				tokenCache.Set(cfg.IssuerURL, cfg.ClientID, refreshed)
				token = refreshed
			}
		}
	}

	// If no valid cached token, authenticate
	if token == nil {
		authenticator := auth.NewAuthenticator(cfg)
		var err error
		token, err = authenticator.Authenticate()
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		tokenCache.Set(cfg.IssuerURL, cfg.ClientID, token)
	}

	// Create the exec credential response
	expiryTime := metav1.NewTime(token.Expiry)
	response := clientauthv1beta1.ExecCredential{
		TypeMeta: request.TypeMeta,
		Status: &clientauthv1beta1.ExecCredentialStatus{
			Token:               token.AccessToken,
			ExpirationTimestamp: &expiryTime,
		},
	}

	// Write the response to stdout
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("failed to encode exec credential response: %w", err)
	}

	return nil
}

func loadConfig() (*config.Config, error) {
	cfg := &config.Config{
		IssuerURL:    issuerURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Headless:     headless,
		Port:         port,
	}

	// Override with environment variables if set
	if secret := os.Getenv("CLIENT_SECRET"); secret != "" {
		cfg.ClientSecret = secret
	}

	// Load from config file if provided
	if configFile != "" {
		fileCfg, err := config.LoadFromFile(configFile)
		if err != nil {
			return nil, err
		}
		// Merge file config with flags
		if fileCfg.IssuerURL != "" {
			cfg.IssuerURL = fileCfg.IssuerURL
		}
		if fileCfg.ClientID != "" {
			cfg.ClientID = fileCfg.ClientID
		}
		if fileCfg.ClientSecret != "" {
			cfg.ClientSecret = fileCfg.ClientSecret
		}
		if fileCfg.Headless {
			cfg.Headless = fileCfg.Headless
		}
		if fileCfg.Port != 0 {
			cfg.Port = fileCfg.Port
		}
	}

	return cfg, nil
}
