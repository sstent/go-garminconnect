package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/sstent/go-garminconnect/internal/auth"
	"github.com/sstent/go-garminconnect/internal/api"
)

func main() {
	// Get credentials from environment
	username := os.Getenv("GARMIN_USERNAME")
	password := os.Getenv("GARMIN_PASSWORD")
	if username == "" || password == "" {
		fmt.Println("GARMIN_USERNAME and GARMIN_PASSWORD must be set")
		os.Exit(1)
	}

	// Create authentication client with headless mode enabled
	authClient := auth.NewAuthClient()

	// Authenticate with credentials
	token, err := authClient.Authenticate(context.Background(), username, password, "")
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		os.Exit(1)
	}

	// API client not currently used in this simple server
	// It's created here for demonstration purposes only
	_, err = api.NewClient(token.AccessToken)
	if err != nil {
		fmt.Printf("Failed to create API client: %v\n", err)
		os.Exit(1)
	}

	// Create HTTP server
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)

	// For demonstration purposes, print API client status
	// This line was removed because baseURL is unexported
	// fmt.Printf("API client initialized for %s\n", apiClient.baseURL)

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
		<html>
			<body>
				<h1>Go GarminConnect Client</h1>
				<p>Authentication successful! API client ready.</p>
			</body>
		</html>
	`))
}

// Removed OAuth handlers since we're using credentials-based auth

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
