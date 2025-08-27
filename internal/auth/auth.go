package auth

import (
	"net/http"

	"github.com/dghubble/oauth1"
)

// OAuthConfig holds OAuth1 configuration for Garmin Connect
type OAuthConfig struct {
	ConsumerKey    string
	ConsumerSecret string
}

// TokenStorage defines the interface for storing and retrieving OAuth tokens
type TokenStorage interface {
	GetToken() (*oauth1.Token, error)
	SaveToken(*oauth1.Token) error
}

// Authenticate initiates the OAuth1 authentication flow
func Authenticate(w http.ResponseWriter, r *http.Request, config *OAuthConfig, storage TokenStorage) {
	// Create OAuth1 config
	oauthConfig := oauth1.Config{
		ConsumerKey:    config.ConsumerKey,
		ConsumerSecret: config.ConsumerSecret,
		CallbackURL:    "http://localhost:8080/callback",
		Endpoint: oauth1.Endpoint{
			RequestTokenURL: "https://connect.garmin.com/oauth-service/oauth/request_token",
			AuthorizeURL:    "https://connect.garmin.com/oauth-service/oauth/authorize",
			AccessTokenURL:  "https://connect.garmin.com/oauth-service/oauth/access_token",
		},
	}

	// Get request token
	requestToken, _, err := oauthConfig.RequestToken()
	if err != nil {
		http.Error(w, "Failed to get request token", http.StatusInternalServerError)
		return
	}

	// Save request token secret temporarily (for callback)
	// In a real application, you'd store this in a session

	// Redirect to authorization URL
	authURL, err := oauthConfig.AuthorizationURL(requestToken)
	if err != nil {
		http.Error(w, "Failed to get authorization URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL.String(), http.StatusTemporaryRedirect)
}

// Callback handles OAuth1 callback
func Callback(w http.ResponseWriter, r *http.Request, config *OAuthConfig, storage TokenStorage, requestSecret string) {
	// Get request token and verifier from query params
	requestToken := r.URL.Query().Get("oauth_token")
	verifier := r.URL.Query().Get("oauth_verifier")

	// Create OAuth1 config
	oauthConfig := oauth1.Config{
		ConsumerKey:    config.ConsumerKey,
		ConsumerSecret: config.ConsumerSecret,
		Endpoint: oauth1.Endpoint{
			RequestTokenURL: "https://connect.garmin.com/oauth-service/oauth/request_token",
			AccessTokenURL:  "https://connect.garmin.com/oauth-service/oauth/access_token",
		},
	}

	// Get access token
	accessToken, accessSecret, err := oauthConfig.AccessToken(requestToken, requestSecret, verifier)
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	// Create token and save
	token := &oauth1.Token{
		Token:       accessToken,
		TokenSecret: accessSecret,
	}
	
	err = storage.SaveToken(token)
	if err != nil {
		http.Error(w, "Failed to save token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authentication successful!"))
}
