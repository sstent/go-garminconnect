package main

import (
	"context"
	"fmt"
	"os"
	"time"
	
	"github.com/joho/godotenv"
	"github.com/sstent/go-garminconnect/internal/api"
	"github.com/sstent/go-garminconnect/internal/auth"
)

func main() {
	// Try to load from .env if environment variables not set
	if os.Getenv("GARMIN_USERNAME") == "" || os.Getenv("GARMIN_PASSWORD") == "" {
		if err := godotenv.Load(); err != nil {
			fmt.Println("Failed to load .env file:", err)
		}
	}

	// Verify required credentials
	if os.Getenv("GARMIN_USERNAME") == "" || os.Getenv("GARMIN_PASSWORD") == "" {
		fmt.Println("GARMIN_USERNAME and GARMIN_PASSWORD must be set in environment or .env file")
		os.Exit(1)
	}

	// Set up authentication client with headless mode enabled
	client := auth.NewAuthClient()
	token, err := client.Authenticate(
		context.Background(),
		os.Getenv("GARMIN_USERNAME"),
		os.Getenv("GARMIN_PASSWORD"),
		"", // MFA token if needed
	)
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		os.Exit(1)
	}

	// Create API client
	apiClient, err := api.NewClient(token.AccessToken)
	if err != nil {
		fmt.Printf("Failed to create API client: %v\n", err)
		os.Exit(1)
	}

	// Parse date range (default: last 7 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	// Get body composition data
	composition, err := apiClient.GetBodyComposition(context.Background(), api.BodyCompositionRequest{
		StartDate: api.Time(startDate),
		EndDate:   api.Time(endDate),
	})
	if err != nil {
		fmt.Printf("Failed to get body composition: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println("Body Composition Data:")
	fmt.Println("Date\t\tBone Mass\tMuscle Mass\tBody Fat\tHydration")
	for _, entry := range composition {
		fmt.Printf("%s\t%.1fg\t\t%.1fg\t\t%.1f%%\t\t%.1f%%\n",
			time.Time(entry.Timestamp).Format("2006-01-02"),
			entry.BoneMass,
			entry.MuscleMass,
			entry.BodyFat,
			entry.Hydration,
		)
	}
}
