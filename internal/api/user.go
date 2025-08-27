package api

import (
	"context"
	"fmt"
)

// UserProfile represents a Garmin Connect user profile
type UserProfile struct {
	DisplayName   string `json:"displayName"`
	FullName      string `json:"fullName"`
	EmailAddress  string `json:"emailAddress"`
	Username      string `json:"username"`
	ProfileID     string `json:"profileId"`
	ProfileImage  string `json:"profileImageUrlLarge"`
	Location      string `json:"location"`
	FitnessLevel  string `json:"fitnessLevel"`
	Height        float64 `json:"height"`
	Weight        float64 `json:"weight"`
	Birthdate     string `json:"birthDate"`
}

// GetUserProfile retrieves the user's profile information
func (c *Client) GetUserProfile(ctx context.Context) (*UserProfile, error) {
	var profile UserProfile
	path := "/userprofile-service/socialProfile"
	
	if err := c.Get(ctx, path, &profile); err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	
	// Handle empty profile response
	if profile.ProfileID == "" {
		return nil, fmt.Errorf("user profile not found")
	}
	
	return &profile, nil
}
