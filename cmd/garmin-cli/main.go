package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/sstent/go-garminconnect/internal/api"
	"github.com/sstent/go-garminconnect/internal/auth/garth"
)

var rootCmd = &cobra.Command{
	Use:   "garmin-cli",
	Short: "CLI for interacting with Garmin Connect API",
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Garmin Connect",
	Run:   loginHandler,
}

func loginHandler(cmd *cobra.Command, args []string) {
	// Try to load from .env if environment variables not set
	if os.Getenv("GARMIN_USERNAME") == "" || os.Getenv("GARMIN_PASSWORD") == "" {
		if err := godotenv.Load(); err != nil {
			fmt.Println("Failed to load .env file:", err)
		}

		// Re-check after loading .env
		if os.Getenv("GARMIN_USERNAME") == "" || os.Getenv("GARMIN_PASSWORD") == "" {
			fmt.Println("GARMIN_USERNAME and GARMIN_PASSWORD must be set in environment or .env file")
			os.Exit(1)
		}
	}

	// Configure session persistence
	sessionPath := filepath.Join(os.Getenv("HOME"), ".garmin", "session.json")
	authClient := garth.NewAuthenticator("https://connect.garmin.com", sessionPath)

	// Implement CLI prompter
	authClient.MFAPrompter = ConsolePrompter{}

	// Try to load existing session
	var session *garth.Session
	var err error
	if _, err = os.Stat(sessionPath); err == nil {
		session, err = garth.LoadSession(sessionPath)
		if err != nil {
			fmt.Printf("Session loading failed: %v\n", err)
		}
	}

	// Perform authentication if no valid session
	if session == nil {
		username := os.Getenv("GARMIN_USERNAME")
		password := os.Getenv("GARMIN_PASSWORD")
		session, err = authClient.Login(username, password)
		if err != nil {
			fmt.Printf("Authentication failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Create API client with session management
	apiClient, err := api.NewClient(authClient, session, sessionPath)
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

func main() {
	// Setup command structure
	authCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(authCmd)

	// Execute CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// getCredentials prompts for username and password
func getCredentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Garmin username: ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Enter Garmin password: ")
	password, _ := reader.ReadString('\n')
	return strings.TrimSpace(username), strings.TrimSpace(password)
}

// ConsolePrompter implements MFAPrompter for CLI
type ConsolePrompter struct{}

func (c ConsolePrompter) GetMFACode(ctx context.Context) (string, error) {
	fmt.Print("Enter Garmin MFA code: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}
