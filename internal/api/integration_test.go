package api

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sstent/go-garminconnect/internal/auth/garth"
	"github.com/stretchr/testify/assert"
)

// TestIntegrationHealthMetrics tests end-to-end retrieval of all health metrics
func TestIntegrationHealthMetrics(t *testing.T) {
	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Setup mock responses
	mockServer.SetHealthHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "sleep/daily"):
			w.Write([]byte(`{
				"calendarDate": "2025-08-28T00:00:00Z",
				"sleepTimeSeconds": 28800,
				"deepSleepSeconds": 7200,
				"lightSleepSeconds": 14400,
				"remSleepSeconds": 7200,
				"awakeSeconds": 1800,
				"sleepScore": 85,
				"sleepScores": {
					"overall": 85,
					"duration": 90,
					"deep": 80,
					"rem": 75,
					"light": 70,
					"awake": 95
				}
			}`))
		case strings.Contains(r.URL.Path, "stress/daily"):
			w.Write([]byte(`{
				"calendarDate": "2025-08-28T00:00:00Z",
				"overallStressLevel": 42,
				"restStressDuration": 18000,
				"lowStressDuration": 14400,
				"mediumStressDuration": 7200,
				"highStressDuration": 3600,
				"stressQualifier": "Balanced"
			}`))
		case strings.Contains(r.URL.Path, "steps/daily"):
			w.Write([]byte(`{
				"calendarDate": "2025-08-28T00:00:00Z",
				"totalSteps": 12500,
				"goal": 10000,
				"activeMinutes": 90,
				"distanceMeters": 8500.5,
				"caloriesBurned": 450,
				"stepsToGoal": 0,
				"stepGoalAchieved": true
			}`))
		case strings.Contains(r.URL.Path, "hrv-service/hrv/"):
			w.Write([]byte(`{
				"date": "2025-08-28T00:00:00Z",
				"restingHrv": 65,
				"weeklyAvg": 62,
				"lastNightAvg": 68,
				"hrvStatus": "Balanced",
				"hrvStatusMessage": "Normal variation",
				"baselineHrv": 64,
				"changeFromBaseline": 1
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Create authenticated client
	session := &garth.Session{
		OAuth2Token: "test-token",
		ExpiresAt:   time.Now().Add(8 * time.Hour),
	}
	client, err := NewClient(session, "")
	assert.NoError(t, err)
	client.HTTPClient.SetBaseURL(mockServer.URL())

	// Test context
	ctx := context.Background()
	date := time.Date(2025, 8, 28, 0, 0, 0, 0, time.UTC)

	t.Run("RetrieveSleepData", func(t *testing.T) {
		data, err := client.GetSleepData(ctx, date)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, 28800, data.SleepTimeSeconds)
		assert.Equal(t, 85, data.SleepScore)
	})

	t.Run("RetrieveStressData", func(t *testing.T) {
		data, err := client.GetStressData(ctx, date)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, 42, data.OverallStressLevel)
		assert.Equal(t, "Balanced", data.StressQualifier)
	})

	t.Run("RetrieveStepsData", func(t *testing.T) {
		data, err := client.GetStepsData(ctx, date)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, 12500, data.TotalSteps)
		assert.True(t, data.StepGoalAchieved)
	})

	t.Run("RetrieveHRVData", func(t *testing.T) {
		data, err := client.GetHRVData(ctx, date)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, 65.0, data.RestingHrv)
		assert.Equal(t, "Balanced", data.HrvStatus)
	})
}
