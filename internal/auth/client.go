package auth

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

// AuthClient struct handles authentication
type AuthClient struct {
	Client   *http.Client
	TokenURL string
}

// NewAuthClient creates a new authentication client with cookie persistence
func NewAuthClient() *AuthClient {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
	return &AuthClient{
		Client:   client,
		TokenURL: "https://connectapi.garmin.com/oauth-service/oauth/exchange/user/2.0",
	}
}
