package types

import "time"

// TokenInfo holds the authentication token information
type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	Expiry       time.Time
}
