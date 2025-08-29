package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sstent/go-garminconnect/internal/auth/garth"
	"github.com/stretchr/testify/assert"
)

// BenchmarkGetSleepData measures performance of GetSleepData method
func BenchmarkGetSleepData(b *testing.B) {
	now := time.Now()
	testDate := now.Format(time.RFC3339)

	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Setup handler for health endpoint
	mockServer.SetHealthHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"calendarDate":      testDate,
			"sleepTimeSeconds":  28800, // 8 hours in seconds
			"deepSleepSeconds":  7200,  // 2 hours
			"lightSleepSeconds": 14400, // 4 hours
			"remSleepSeconds":   7200,  // 2 hours
			"awakeSeconds":      1800,  // 30 minutes
			"sleepScore":        85,
			"sleepScores": map[string]interface{}{
				"overall":  85,
				"duration": 90,
				"deep":     80,
				"rem":      75,
				"light":    70,
				"awake":    95,
			},
		})
	})

	// Create client
	client := NewClientWithBaseURL(mockServer.URL())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetSleepData(context.Background(), now)
	}
}

// BenchmarkGetHRVData measures performance of GetHRVData method
func BenchmarkGetHRVData(b *testing.B) {
	now := time.Now()
	testDate := now.Format(time.RFC3339)

	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Setup handler for health endpoint
	mockServer.SetHealthHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"date":         testDate,
			"restingHrv":   65.0,
			"weeklyAvg":    62.0,
			"lastNightAvg": 68.0,
		})
	})

	// Create client
	client := NewClientWithBaseURL(mockServer.URL())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetHRVData(context.Background(), now)
	}
}

// BenchmarkGetBodyBatteryData measures performance of GetBodyBatteryData method
func BenchmarkGetBodyBatteryData(b *testing.B) {
	now := time.Now()
	testDate := now.Format(time.RFC3339)

	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Setup handler for health endpoint
	mockServer.SetHealthHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"date":    testDate,
			"charged": 85,
			"drained": 45,
			"highest": 95,
			"lowest":  30,
		})
	})

	// Create client
	client := NewClientWithBaseURL(mockServer.URL())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetBodyBatteryData(context.Background(), now)
	}
}

func TestGetSleepData(t *testing.T) {
	now := time.Now()
	testDate := now.Format(time.RFC3339) // Use RFC3339 format for proper time parsing

	tests := []struct {
		name          string
		date          time.Time
		mockResponse  interface{}
		mockStatus    int
		expected      *SleepData
		expectedError string
	}{
		{
			name: "successful sleep data retrieval",
			date: now,
			mockResponse: map[string]interface{}{
				"calendarDate":      testDate,
				"sleepTimeSeconds":  28800,
				"deepSleepSeconds":  7200,
				"lightSleepSeconds": 14400,
				"remSleepSeconds":   7200,
				"awakeSeconds":      1800,
				"sleepScore":        85,
				"sleepScores": map[string]interface{}{
					"overall":  85,
					"duration": 90,
					"deep":     80,
					"rem":      75,
					"light":    70,
					"awake":    95,
				},
			},
			mockStatus: http.StatusOK,
			expected: &SleepData{
				CalendarDate:      now.Truncate(time.Second), // Truncate to avoid precision issues
				SleepTimeSeconds:  28800,
				DeepSleepSeconds:  7200,
				LightSleepSeconds: 14400,
				RemSleepSeconds:   7200,
				AwakeSeconds:      1800,
				SleepScore:        85,
				SleepScores: struct {
					Overall  int `json:"overall"`
					Duration int `json:"duration"`
					Deep     int `json:"deep"`
					Rem      int `json:"rem"`
					Light    int `json:"light"`
					Awake    int `json:"awake"`
				}{
					Overall:  85,
					Duration: 90,
					Deep:     80,
					Rem:      75,
					Light:    70,
					Awake:    95,
				},
			},
		},
		{
			name: "sleep data not found",
			date: now,
			mockResponse: map[string]interface{}{
				"error": "No sleep data found",
			},
			mockStatus:    http.StatusNotFound,
			expectedError: "failed to get sleep data",
		},
	}

	mockServer := NewMockServer()
	defer mockServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with non-expired session
			session := &garth.Session{
				OAuth2Token: "test-token",
				ExpiresAt:   time.Now().Add(8 * time.Hour),
			}
			client, err := NewClient(session, "")
			assert.NoError(t, err)
			client.HTTPClient.SetBaseURL(mockServer.URL())

			mockServer.Reset()
			mockServer.SetHealthHandler(func(w http.ResponseWriter, r *http.Request) {
				// Only handle sleep data requests
				if strings.Contains(r.URL.Path, "sleep/daily") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					json.NewEncoder(w).Encode(tt.mockResponse)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			})

			data, err := client.GetSleepData(context.Background(), tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, data)
				return // Early return to prevent nil pointer access
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				// Only check fields if data is not nil
				if data != nil {
					assert.Equal(t, tt.expected.CalendarDate, data.CalendarDate)
					assert.Equal(t, tt.expected.SleepTimeSeconds, data.SleepTimeSeconds)
					assert.Equal(t, tt.expected.DeepSleepSeconds, data.DeepSleepSeconds)
					assert.Equal(t, tt.expected.LightSleepSeconds, data.LightSleepSeconds)
					assert.Equal(t, tt.expected.RemSleepSeconds, data.RemSleepSeconds)
					assert.Equal(t, tt.expected.AwakeSeconds, data.AwakeSeconds)
					assert.Equal(t, tt.expected.SleepScore, data.SleepScore)
					assert.Equal(t, tt.expected.SleepScores.Overall, data.SleepScores.Overall)
					assert.Equal(t, tt.expected.SleepScores.Duration, data.SleepScores.Duration)
					assert.Equal(t, tt.expected.SleepScores.Deep, data.SleepScores.Deep)
					assert.Equal(t, tt.expected.SleepScores.Rem, data.SleepScores.Rem)
					assert.Equal(t, tt.expected.SleepScores.Light, data.SleepScores.Light)
					assert.Equal(t, tt.expected.SleepScores.Awake, data.SleepScores.Awake)
				}
			}
		})
	}
}

func TestGetHRVData(t *testing.T) {
	now := time.Now()
	testDate := now.Format(time.RFC3339)

	tests := []struct {
		name          string
		date          time.Time
		mockResponse  interface{}
		mockStatus    int
		expected      *HRVData
		expectedError string
	}{
		{
			name: "successful HRV data retrieval",
			date: now,
			mockResponse: map[string]interface{}{
				"date":         testDate,
				"restingHrv":   65.0,
				"weeklyAvg":    62.0,
				"lastNightAvg": 68.0,
			},
			mockStatus: http.StatusOK,
			expected: &HRVData{
				Date:         now.Truncate(time.Second),
				RestingHrv:   65.0,
				WeeklyAvg:    62.0,
				LastNightAvg: 68.0,
			},
		},
		{
			name: "HRV data not available",
			date: now,
			mockResponse: map[string]interface{}{
				"error": "No HRV data",
			},
			mockStatus:    http.StatusNotFound,
			expectedError: "failed to get HRV data",
		},
	}

	mockServer := NewMockServer()
	defer mockServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with non-expired session
			session := &garth.Session{
				OAuth2Token: "test-token",
				ExpiresAt:   time.Now().Add(8 * time.Hour),
			}
			client, err := NewClient(session, "")
			assert.NoError(t, err)
			client.HTTPClient.SetBaseURL(mockServer.URL())

			mockServer.Reset()
			mockServer.SetHealthHandler(func(w http.ResponseWriter, r *http.Request) {
				// Only handle HRV data requests
				if strings.Contains(r.URL.Path, "hrv/") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					json.NewEncoder(w).Encode(tt.mockResponse)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			})

			data, err := client.GetHRVData(context.Background(), tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, data)
				return // Early return to prevent nil pointer access
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				if data != nil {
					assert.Equal(t, tt.expected.RestingHrv, data.RestingHrv)
					assert.Equal(t, tt.expected.WeeklyAvg, data.WeeklyAvg)
					assert.Equal(t, tt.expected.LastNightAvg, data.LastNightAvg)
				}
			}
		})
	}
}

func TestGetBodyBatteryData(t *testing.T) {
	now := time.Now()
	testDate := now.Format(time.RFC3339)

	tests := []struct {
		name          string
		date          time.Time
		mockResponse  interface{}
		mockStatus    int
		expected      *BodyBatteryData
		expectedError string
	}{
		{
			name: "successful body battery retrieval",
			date: now,
			mockResponse: map[string]interface{}{
				"date":    testDate,
				"charged": 85,
				"drained": 45,
				"highest": 95,
				"lowest":  30,
			},
			mockStatus: http.StatusOK,
			expected: &BodyBatteryData{
				Date:    now.Truncate(time.Second),
				Charged: 85,
				Drained: 45,
				Highest: 95,
				Lowest:  30,
			},
		},
		{
			name: "body battery data missing",
			date: now,
			mockResponse: map[string]interface{}{
				"error": "Body battery data unavailable",
			},
			mockStatus:    http.StatusNotFound,
			expectedError: "failed to get Body Battery data",
		},
	}

	mockServer := NewMockServer()
	defer mockServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with non-expired session
			session := &garth.Session{
				OAuth2Token: "test-token",
				ExpiresAt:   time.Now().Add(8 * time.Hour),
			}
			client, err := NewClient(session, "")
			assert.NoError(t, err)
			client.HTTPClient.SetBaseURL(mockServer.URL())

			mockServer.Reset()
			mockServer.SetHealthHandler(func(w http.ResponseWriter, r *http.Request) {
				// Only handle body battery requests
				if strings.Contains(r.URL.Path, "bodybattery/") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					json.NewEncoder(w).Encode(tt.mockResponse)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			})

			data, err := client.GetBodyBatteryData(context.Background(), tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, data)
				return // Early return to prevent nil pointer access
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				if data != nil {
					assert.Equal(t, tt.expected.Charged, data.Charged)
					assert.Equal(t, tt.expected.Drained, data.Drained)
					assert.Equal(t, tt.expected.Highest, data.Highest)
					assert.Equal(t, tt.expected.Lowest, data.Lowest)
				}
			}
		})
	}
}
