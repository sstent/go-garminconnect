package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sstent/go-garminconnect/internal/auth/garth"
	"github.com/stretchr/testify/assert"
)

func TestGetUserProfile(t *testing.T) {
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
				"displayName":          "Mock User",
				"fullName":             "Mock User Full",
				"emailAddress":         "mock@example.com",
				"username":             "mockuser",
				"profileId":            "mock-123",
				"profileImageUrlLarge": "https://example.com/mock.jpg",
				"location":             "Mock Location",
				"fitnessLevel":         "INTERMEDIATE",
				"height":               175.0,
				"weight":               70.0,
				"birthDate":            "1990-01-01",
			},
			mockStatus: http.StatusOK,
			expected: &UserProfile{
				DisplayName:  "Mock User",
				FullName:     "Mock User Full",
				EmailAddress: "mock@example.com",
				Username:     "mockuser",
				ProfileID:    "mock-123",
				ProfileImage: "https://example.com/mock.jpg",
				Location:     "Mock Location",
				FitnessLevel: "INTERMEDIATE",
				Height:       175.0,
				Weight:       70.0,
				Birthdate:    "1990-01-01",
			},
		},
		{
			name: "profile not found",
			mockResponse: map[string]interface{}{
				"error": "Profile not found",
			},
			mockStatus:    http.StatusNotFound,
			expectedError: "API error 404: Profile not found",
		},
		{
			name:          "invalid response format",
			mockResponse:  "not-a-valid-json-object",
			mockStatus:    http.StatusOK,
			expectedError: "failed to get user profile: json: cannot unmarshal string",
		},
		{
			name: "server error",
			mockResponse: map[string]interface{}{
				"error": "Internal server error",
			},
			mockStatus:    http.StatusInternalServerError,
			expectedError: "API error 500: Internal server error",
		},
	}

	mockServer := NewMockServer()
	defer mockServer.Close()
	// Create client with non-expired session
	session := &garth.Session{
		OAuth2Token: "test-token",
		ExpiresAt:   time.Now().Add(8 * time.Hour),
	}
	// Use mock authenticator
	mockAuth := NewMockAuthenticator()
	client, err := NewClient(mockAuth, session, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.HTTPClient.SetBaseURL(mockServer.URL())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer.Reset()

			// Set custom handler directly
			mockServer.SetUserHandler(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			})

			profile, err := client.GetUserProfile(context.Background())

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, profile)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile) // Add nil check
				assert.Equal(t, tt.expected, profile)
			}
		})
	}
}

func BenchmarkGetUserProfile(b *testing.B) {
	mockServer := NewMockServer()
	defer mockServer.Close()

	mockResponse := map[string]interface{}{
		"displayName":          "Benchmark User",
		"fullName":             "Benchmark User Full",
		"emailAddress":         "benchmark@example.com",
		"username":             "benchmark",
		"profileId":            "benchmark-123",
		"profileImageUrlLarge": "https://example.com/benchmark.jpg",
	}
	mockServer.SetResponse("/userprofile-service/socialProfile", http.StatusOK, mockResponse)

	client := NewClientWithBaseURL(mockServer.URL())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = client.GetUserProfile(context.Background())
	}
}

func TestGetUserStats(t *testing.T) {
	now := time.Now()
	testDate := now.Format("2006-01-02")

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
				"totalSteps":       10000,
				"totalDistance":    8500.5,
				"totalCalories":    2200,
				"activeMinutes":    45,
				"restingHeartRate": 55,
				"date":             testDate,
			},
			mockStatus: http.StatusOK,
			expected: &UserStats{
				TotalSteps:    10000,
				TotalDistance: 8500.5,
				TotalCalories: 2200,
				ActiveMinutes: 45,
				RestingHR:     55,
				Date:          testDate,
			},
		},
		{
			name: "stats not found for date",
			date: now,
			mockResponse: map[string]interface{}{
				"error": "No stats found",
			},
			mockStatus:    http.StatusNotFound,
			expectedError: "API error 404: No stats found",
		},
		{
			name:          "invalid stats response",
			date:          now,
			mockResponse:  "invalid-json-response",
			mockStatus:    http.StatusOK,
			expectedError: "failed to get user stats: json: cannot unmarshal string",
		},
	}

	mockServer := NewMockServer()
	defer mockServer.Close()
	// Create client with non-expired session
	session := &garth.Session{
		OAuth2Token: "test-token",
		ExpiresAt:   time.Now().Add(8 * time.Hour),
	}
	// Use mock authenticator
	mockAuth := NewMockAuthenticator()
	client, err := NewClient(mockAuth, session, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.HTTPClient.SetBaseURL(mockServer.URL())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer.Reset()

			// Set custom handler directly for stats
			mockServer.SetStatsHandler(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			})

			stats, err := client.GetUserStats(context.Background(), tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats) // Add nil check
				assert.Equal(t, tt.expected, stats)
			}
		})
	}
}

func BenchmarkGetUserStats(b *testing.B) {
	now := time.Now()
	mockServer := NewMockServer()
	defer mockServer.Close()

	path := fmt.Sprintf("/stats-service/stats/daily/%s", now.Format("2006-01-02"))
	mockResponse := map[string]interface{}{
		"totalSteps":    15000,
		"totalDistance": 12000.0,
		"totalCalories": 3000,
		"activeMinutes": 60,
	}
	mockServer.SetResponse(path, http.StatusOK, mockResponse)

	client := NewClientWithBaseURL(mockServer.URL())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = client.GetUserStats(context.Background(), now)
	}
}
