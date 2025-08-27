package auth

import "time"

// Token represents OAuth2 tokens
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}
