package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sstent/go-garminconnect/internal/api"
	"github.com/sstent/go-garminconnect/internal/auth/garth"
)

func main() {
	// Initialize authentication session
	session := &garth.Session{
		OAuth2Token: "your_oauth2_token_here",
		ExpiresAt:   time.Now().Add(8 * time.Hour),
	}

	// Create API client
	client, err := api.NewClient(session, "")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get user profile
	profile, err := client.GetUserProfile(context.Background())
	if err != nil {
		log.Fatalf("Failed to get user profile: %v", err)
	}
	fmt.Printf("User: %s (%s)\n", profile.FullName, profile.DisplayName)

	// Get sleep data for today
	today := time.Now()
	sleepData, err := client.GetSleepData(context.Background(), today)
	if err != nil {
		log.Fatalf("Failed to get sleep data: %v", err)
	}
	fmt.Printf("Sleep duration: %s\n", time.Duration(sleepData.SleepTimeSeconds)*time.Second)

	// Get stress data
	stressData, err := client.GetStressData(context.Background(), today)
	if err != nil {
		log.Fatalf("Failed to get stress data: %v", err)
	}
	fmt.Printf("Daily stress level: %d\n", stressData.OverallStressLevel)

	// Get steps data
	stepsData, err := client.GetStepsData(context.Background(), today)
	if err != nil {
		log.Fatalf("Failed to get steps data: %v", err)
	}
	fmt.Printf("Steps today: %d\n", stepsData.TotalSteps)
}
