package auth

import "time"

// Token represents both OAuth1 and OAuth2 tokens
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`

	// OAuth1 tokens for compatibility with legacy systems
	OAuthToken  string `json:"oauth_token"`
	OAuthSecret string `json:"oauth_secret"`
}
