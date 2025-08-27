package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sstent/go-garminconnect/internal/auth"
)

func main() {
	// Get consumer key and secret from environment
	consumerKey := os.Getenv("GARMIN_CONSUMER_KEY")
	consumerSecret := os.Getenv("GARMIN_CONSUMER_SECRET")
	if consumerKey == "" || consumerSecret == "" {
		fmt.Println("GARMIN_CONSUMER_KEY and GARMIN_CONSUMER_SECRET must be set")
		os.Exit(1)
	}

	// Configure authentication
	oauthConfig := &auth.OAuthConfig{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
	}

	// Set up token storage
	tokenStorage := auth.NewFileStorage()

	// Create HTTP server
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler(oauthConfig, tokenStorage))
	http.HandleFunc("/callback", callbackHandler(oauthConfig, tokenStorage))
	http.HandleFunc("/mfa", auth.MFAHandler)
	http.HandleFunc("/health", healthHandler)

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
		<html>
			<body>
				<h1>Go GarminConnect Client</h1>
				<a href="/login">Login with Garmin</a>
			</body>
		</html>
	`))
}

func loginHandler(config *auth.OAuthConfig, storage auth.TokenStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth.Authenticate(w, r, config, storage)
	}
}

func callbackHandler(config *auth.OAuthConfig, storage auth.TokenStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real app, we'd retrieve the request secret from session storage
		// For now, we'll use a placeholder
		requestSecret := "placeholder-secret"
		
		auth.Callback(w, r, config, storage, requestSecret)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
