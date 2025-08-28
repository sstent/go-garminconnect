package api

import (
	"context"
	"fmt"
	"time"
)

// UserProfile represents a Garmin Connect user profile
type UserProfile struct {
	DisplayName  string  `json:"displayName"`
	FullName     string  `json:"fullName"`
	EmailAddress string  `json:"emailAddress"`
	Username     string  `json:"username"`
	ProfileID    string  `json:"profileId"`
	ProfileImage string  `json:"profileImageUrlLarge"`
	Location     string  `json:"location"`
	FitnessLevel string  `json:"fitnessLevel"`
	Height       float64 `json:"height"`
	Weight       float64 `json:"weight"`
	Birthdate    string  `json:"birthDate"`
}

// UserStats represents fitness statistics for a user
type UserStats struct {
	TotalSteps    int     `json:"totalSteps"`
	TotalDistance float64 `json:"totalDistance"` // in meters
	TotalCalories int     `json:"totalCalories"`
	ActiveMinutes int     `json:"activeMinutes"`
	RestingHR     int     `json:"restingHeartRate"`
	Date          string  `json:"date"` // Store as string in "YYYY-MM-DD" format
}

// GetUserProfile retrieves the user's profile information
func (c *Client) GetUserProfile(ctx context.Context) (*UserProfile, error) {
	var profile UserProfile
	if err := c.Get(ctx, "/userprofile-service/socialProfile", &profile); err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return &profile, nil
}

// GetUserStats retrieves fitness statistics for a user for a specific date
func (c *Client) GetUserStats(ctx context.Context, date time.Time) (*UserStats, error) {
	var stats UserStats
	path := fmt.Sprintf("/stats-service/stats/daily/%s", date.Format("2006-01-02"))
	if err := c.Get(ctx, path, &stats); err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	return &stats, nil
}
