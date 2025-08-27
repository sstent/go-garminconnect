package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUserProfile(t *testing.T) {
	// Define test cases
	tests := []struct {
		name          string
		mockResponse  interface{}
		mockStatus    int
		expected      *UserProfile
		expectedError string
	}{
		{
			name: "successful profile retrieval",
			mockResponse: map[string]interface{}{
				"displayName":      "John Doe",
				"fullName":         "John Michael Doe",
				"emailAddress":     "john.doe@example.com",
				"username":         "johndoe",
				"profileId":        "123456",
				"profileImageUrlLarge": "https://example.com/profile.jpg",
				"location":         "San Francisco, CA",
				"fitnessLevel":     "INTERMEDIATE",
				"height":           180.0,
				"weight":           75.0,
				"birthDate":        "1985-01-01",
			},
			mockStatus: http.StatusOK,
			expected: &UserProfile{
				DisplayName:   "John Doe",
				FullName:      "John Michael Doe",
				EmailAddress:  "john.doe@example.com",
				Username:      "johndoe",
				ProfileID:     "123456",
				ProfileImage:  "https://example.com/profile.jpg",
				Location:      "San Francisco, CA",
				FitnessLevel:  "INTERMEDIATE",
				Height:        180.0,
				Weight:        75.0,
				Birthdate:     "1985-01-01",
			},
		},
		{
			name: "profile not found",
			mockResponse: map[string]interface{}{
				"error": "Profile not found",
			},
			mockStatus:    http.StatusNotFound,
			expectedError: "user profile not found",
		},
		{
			name: "invalid response format",
			mockResponse: map[string]interface{}{
				"invalid": "data",
			},
			mockStatus:    http.StatusOK,
			expectedError: "failed to parse user profile",
		},
		{
			name: "server error",
			mockResponse: map[string]interface{}{
				"error": "Internal server error",
			},
			mockStatus:    http.StatusInternalServerError,
			expectedError: "API request failed with status 500",
		},
	}

	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Create client
	client := NewClientWithBaseURL(mockServer.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure mock server
			mockServer.Reset()
			mockServer.SetResponse("/userprofile-service/socialProfile", tt.mockStatus, tt.mockResponse)

			// Execute test
			profile, err := client.GetUserProfile(context.Background())

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, profile)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, profile)
			}
		})
	}
}

// BenchmarkGetUserProfile measures performance of GetUserProfile method
func BenchmarkGetUserProfile(b *testing.B) {
	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()
	
	// Setup successful response
	mockResponse := map[string]interface{}{
		"displayName":      "Benchmark User",
		"fullName":         "Benchmark User Full",
		"emailAddress":     "benchmark@example.com",
		"username":         "benchmark",
		"profileId":        "benchmark-123",
		"profileImageUrlLarge": "https://example.com/benchmark.jpg",
		"location":         "Benchmark City",
		"fitnessLevel":     "ADVANCED",
		"height":           185.0,
		"weight":           80.0,
		"birthDate":        "1990-01-01",
	}
	mockServer.SetResponse("/userprofile-service/socialProfile", http.StatusOK, mockResponse)

	// Create client
	client := NewClientWithBaseURL(mockServer.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetUserProfile(context.Background())
	}
}

// BenchmarkGetUserStats measures performance of GetUserStats method
func BenchmarkGetUserStats(b *testing.B) {
	now := time.Now()
	testDate := now.Format("2006-01-02")
	
	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()
	
	// Setup successful response
	mockResponse := map[string]interface{}{
		"totalSteps":     15000,
		"totalDistance":  12000.0,
		"totalCalories":  3000,
		"activeMinutes":  60,
		"restingHeartRate": 50,
		"date":           testDate,
	}
	path := fmt.Sprintf("/stats-service/stats/daily/%s", now.Format("2006-01-02"))
	mockServer.SetResponse(path, http.StatusOK, mockResponse)

	// Create client
	client := NewClientWithBaseURL(mockServer.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetUserStats(context.Background(), now)
	}
}

func TestGetUserStats(t *testing.T) {
	now := time.Now()
	testDate := now.Format("2006-01-02")
	
	// Define test cases
	tests := []struct {
		name          string
		date          time.Time
		mockResponse  interface{}
		mockStatus    int
		expected      *UserStats
		expectedError string
	}{
		{
			name: "successful stats retrieval",
			date: now,
			mockResponse: map[string]interface{}{
				"totalSteps":     10000,
				"totalDistance":  8500.5,
				"totalCalories":  2200,
				"activeMinutes":  45,
				"restingHeartRate": 55,
				"date":           testDate,
			},
			mockStatus: http.StatusOK,
			expected: &UserStats{
				TotalSteps:    10000,
				TotalDistance: 8500.5,
				TotalCalories: 2200,
				ActiveMinutes: 45,
				RestingHR:     55,
				Date:          now.Truncate(24 * time.Hour), // Date without time component
			},
		},
		{
			name: "stats not found for date",
			date: now,
			mockResponse: map[string]interface{}{
				"error": "No stats found",
			},
			mockStatus:    http.StatusNotFound,
			expectedError: "failed to get user stats",
		},
		{
			name: "future date error",
			date: now.AddDate(0, 0, 1),
			mockResponse: map[string]interface{}{
				"error": "Date cannot be in the future",
			},
			mockStatus:    http.StatusBadRequest,
			expectedError: "API request failed with status 400",
		},
		{
			name: "invalid stats response",
			date: now,
			mockResponse: map[string]interface{}{
				"invalid": "data",
			},
			mockStatus:    http.StatusOK,
			expectedError: "failed to parse user stats",
		},
	}

	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Create client
	client := NewClientWithBaseURL(mockServer.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure mock server
			mockServer.Reset()
			path := fmt.Sprintf("/stats-service/stats/daily/%s", tt.date.Format("2006-01-02"))
			mockServer.SetResponse(path, tt.mockStatus, tt.mockResponse)

			// Execute test
			stats, err := client.GetUserStats(context.Background(), tt.date)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, stats)
			}
		})
	}
}
