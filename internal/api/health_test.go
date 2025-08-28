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
			"date":     testDate,
			"duration": 480.0,
			"quality":  85.0,
			"sleepStages": map[string]interface{}{
				"deep":  120.0,
				"light": 240.0,
				"rem":   90.0,
				"awake": 30.0,
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
				"date":     testDate,
				"duration": 480.0,
				"quality":  85.0,
				"sleepStages": map[string]interface{}{
					"deep":  120.0,
					"light": 240.0,
					"rem":   90.0,
					"awake": 30.0,
				},
			},
			mockStatus: http.StatusOK,
			expected: &SleepData{
				Date:     now.Truncate(time.Second), // Truncate to avoid precision issues
				Duration: 480.0,
				Quality:  85.0,
				SleepStages: struct {
					Deep  float64 `json:"deep"`
					Light float64 `json:"light"`
					REM   float64 `json:"rem"`
					Awake float64 `json:"awake"`
				}{
					Deep:  120.0,
					Light: 240.0,
					REM:   90.0,
					Awake: 30.0,
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
					assert.Equal(t, tt.expected.Duration, data.Duration)
					assert.Equal(t, tt.expected.Quality, data.Quality)
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