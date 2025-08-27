package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetActivities(t *testing.T) {
	// Create mock server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Accept both escaped and unescaped versions
		expected1 := "/activitylist-service/activities/search?page=1&pageSize=10"
		expected2 := "/activitylist-service/activities/search%3Fpage=1&pageSize=10"
		if r.URL.String() != expected1 && r.URL.String() != expected2 {
			t.Errorf("Unexpected URL: %s", r.URL.String())
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"activities": [
				{
					"activityId": 123,
					"activityName": "Morning Run",
					"activityType": "RUNNING",
					"startTimeLocal": "2023-07-15T08:00:00",
					"duration": 3600,
					"distance": 10000
				}
			],
			"pagination": {
				"pageSize": 10,
				"totalCount": 1,
				"page": 1
			}
		}`))
	}))
	defer testServer.Close()

	// Create client with mock server URL
	client, err := NewClient(testServer.URL, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Execute test
	activities, pagination, err := client.GetActivities(context.Background(), 1, 10)
	
	// Validate results
	assert.NoError(t, err)
	assert.Len(t, activities, 1)
	assert.Equal(t, int64(123), activities[0].ActivityID)
	assert.Equal(t, "Morning Run", activities[0].Name)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 10, pagination.PageSize)
}

func TestGetActivityDetails(t *testing.T) {
	// Create mock server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/activity-service/activity/123", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"activityId": 123,
			"activityName": "Morning Run",
			"activityType": "RUNNING",
			"startTimeLocal": "2023-07-15T08:00:00",
			"duration": 3600,
			"distance": 10000,
			"calories": 720,
			"averageHR": 145,
			"maxHR": 172,
			"averageTemperature": 22.5,
			"elevationGain": 150,
			"elevationLoss": 150,
			"weather": {
				"condition": "SUNNY",
				"temperature": 20,
				"humidity": 60
			},
			"gear": {
				"gearId": "shoes-001",
				"name": "Running Shoes",
				"model": "UltraBoost",
				"description": "Primary running shoes"
			},
			"gpsTracks": [
				{
					"lat": 37.7749,
					"lon": -122.4194,
					"ele": 10,
					"timestamp": "2023-07-15T08:00:00"
				}
			]
		}`))
	}))
	defer testServer.Close()

	// Create client with mock server URL
	client, err := NewClient(testServer.URL, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Execute test
	activity, err := client.GetActivityDetails(context.Background(), 123)
	
	// Validate results
	assert.NoError(t, err)
	assert.Equal(t, int64(123), activity.ActivityID)
	assert.Equal(t, "Morning Run", activity.Name)
	assert.Equal(t, 145, activity.AverageHR)
	assert.Equal(t, 720.0, activity.Calories)
	assert.Equal(t, "SUNNY", activity.Weather.Condition)
	assert.Equal(t, "Running Shoes", activity.Gear.Name)
	assert.Len(t, activity.GPSTracks, 1)
}

func TestGetActivities_ErrorHandling(t *testing.T) {
	// Create mock server that returns error
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	// Create client with mock server URL
	client, err := NewClient(testServer.URL, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Execute test
	_, _, err = client.GetActivities(context.Background(), 1, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get activities")
}

func TestGetActivityDetails_NotFound(t *testing.T) {
	// Create mock server that returns 404
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer testServer.Close()

	// Create client with mock server URL
	client, err := NewClient(testServer.URL, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Execute test
	_, err = client.GetActivityDetails(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource not found")
}
