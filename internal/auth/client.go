package auth

import (
	"net/http"
	"time"
)

// AuthClient handles authentication with Garmin Connect
type AuthClient struct {
	BaseURL   string
	LoginPath string
	TokenURL  string
	Client    *http.Client
}

// NewAuthClient creates a new authentication client
func NewAuthClient() *AuthClient {
	return &AuthClient{
		BaseURL:   "https://connect.garmin.com",
		LoginPath: "/signin",
		TokenURL:  "https://connect.garmin.com/oauth/token",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
