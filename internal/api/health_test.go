package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// BenchmarkGetSleepData measures performance of GetSleepData method
func BenchmarkGetSleepData(b *testing.B) {
	now := time.Now()
	testDate := now.Format("2006-01-02")

	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Setup successful response
	mockResponse := map[string]interface{}{
		"date":     testDate,
		"duration": 480.0,
		"quality":  85.0,
		"sleepStages": map[string]interface{}{
			"deep":  120.0,
			"light": 240.0,
			"rem":   90.0,
			"awake": 30.0,
		},
	}
	path := fmt.Sprintf("/wellness-service/sleep/daily/%s", now.Format("2006-01-02"))
	mockServer.SetResponse(path, http.StatusOK, mockResponse)

	// Create client
	client := NewClientWithBaseURL(mockServer.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetSleepData(context.Background(), now)
	}
}

// BenchmarkGetHRVData measures performance of GetHRVData method
func BenchmarkGetHRVData(b *testing.B) {
	now := time.Now()
	testDate := now.Format("2006-01-02")

	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Setup successful response
	mockResponse := map[string]interface{}{
		"date":         testDate,
		"restingHrv":   65.0,
		"weeklyAvg":    62.0,
		"lastNightAvg": 68.0,
	}
	path := fmt.Sprintf("/hrv-service/hrv/%s", now.Format("2006-01-02"))
	mockServer.SetResponse(path, http.StatusOK, mockResponse)

	// Create client
	client := NewClientWithBaseURL(mockServer.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetHRVData(context.Background(), now)
	}
}

// BenchmarkGetBodyBatteryData measures performance of GetBodyBatteryData method
func BenchmarkGetBodyBatteryData(b *testing.B) {
	now := time.Now()
	testDate := now.Format("2006-01-02")

	// Create test server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Setup successful response
	mockResponse := map[string]interface{}{
		"date":    testDate,
		"charged": 85,
		"drained": 45,
		"highest": 95,
		"lowest":  30,
	}
	path := fmt.Sprintf("/bodybattery-service/bodybattery/%s", now.Format("2006-01-02"))
	mockServer.SetResponse(path, http.StatusOK, mockResponse)

	// Create client
	client := NewClientWithBaseURL(mockServer.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetBodyBatteryData(context.Background(), now)
	}
}

func TestGetSleepData(t *testing.T) {
	now := time.Now()
	testDate := now.Format("2006-01-02")

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
				Date:     now.Truncate(24 * time.Hour),
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
		{
			name: "invalid sleep response",
			date: now,
			mockResponse: map[string]interface{}{
				"invalid": "data",
			},
			mockStatus:    http.StatusOK,
			expectedError: "failed to parse sleep data",
		},
	}

	mockServer := NewMockServer()
	defer mockServer.Close()
	client := NewClientWithBaseURL(mockServer.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer.Reset()
			path := fmt.Sprintf("/wellness-service/sleep/daily/%s", tt.date.Format("2006-01-02"))
			mockServer.SetResponse(path, tt.mockStatus, tt.mockResponse)

			data, err := client.GetSleepData(context.Background(), tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, data)
			}
		})
	}
}

func TestGetHRVData(t *testing.T) {
	now := time.Now()
	testDate := now.Format("2006-01-02")

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
				Date:         now.Truncate(24 * time.Hour),
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
	client := NewClientWithBaseURL(mockServer.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer.Reset()
			path := fmt.Sprintf("/hrv-service/hrv/%s", tt.date.Format("2006-01-02"))
			mockServer.SetResponse(path, tt.mockStatus, tt.mockResponse)

			data, err := client.GetHRVData(context.Background(), tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, data)
			}
		})
	}
}

func TestGetBodyBatteryData(t *testing.T) {
	now := time.Now()
	testDate := now.Format("2006-01-02")

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
				Date:    now.Truncate(24 * time.Hour),
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
	client := NewClientWithBaseURL(mockServer.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer.Reset()
			path := fmt.Sprintf("/bodybattery-service/bodybattery/%s", tt.date.Format("2006-01-02"))
			mockServer.SetResponse(path, tt.mockStatus, tt.mockResponse)

			data, err := client.GetBodyBatteryData(context.Background(), tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, data)
			}
		})
	}
}
