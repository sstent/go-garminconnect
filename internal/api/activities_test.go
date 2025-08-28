package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/sstent/go-garminconnect/internal/auth/garth"
	"github.com/stretchr/testify/assert"
)

// TEST PROGRESS:
// - [ ] Move ValidateFIT to internal/fit package
// - [ ] Create unified mock server implementation
// - [ ] Extend mock server for upload handler
// - [ ] Remove ValidateFIT from this file
// - [ ] Create shared test helper package

// TestGetActivities is now part of table-driven tests below

func TestActivitiesEndpoints(t *testing.T) {
	// Create mock server
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Create a mock session
	session := &garth.Session{OAuth2Token: "test-token"}

	// Create client with mock server URL and session
	client, err := NewClient(session, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.HTTPClient.SetBaseURL(mockServer.URL())

	testCases := []struct {
		name        string
		testFunc    func(t *testing.T, client *Client)
		description string
	}{
		{
			name:        "GetActivitiesSuccess",
			description: "Test successful activity list retrieval",
			testFunc: func(t *testing.T, client *Client) {
				activities, pagination, err := client.GetActivities(context.Background(), 1, 10)
				assert.NoError(t, err)
				assert.Len(t, activities, 1)
				assert.Equal(t, int64(1), activities[0].ActivityID)
				assert.Equal(t, "Morning Run", activities[0].Name)
				assert.Equal(t, 1, pagination.Page)
				assert.Equal(t, 10, pagination.PageSize)
			},
		},
		{
			name:        "GetActivityDetailsSuccess",
			description: "Test successful activity details retrieval",
			testFunc: func(t *testing.T, client *Client) {
				activity, err := client.GetActivityDetails(context.Background(), 1)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), activity.ActivityID)
				assert.Equal(t, "Mock Activity", activity.Name)
				assert.Equal(t, 150, activity.AverageHR)
				assert.Equal(t, "RUNNING", activity.Type)
			},
		},
		{
			name:        "GetActivitiesServerError",
			description: "Test server error handling for activity list",
			testFunc: func(t *testing.T, client *Client) {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				})
				_, _, err := client.GetActivities(context.Background(), 1, 10)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get activities")
			},
		},
		{
			name:        "GetActivityDetailsNotFound",
			description: "Test not found error for activity details",
			testFunc: func(t *testing.T, client *Client) {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				})
				_, err := client.GetActivityDetails(context.Background(), 999)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "resource not found")
			},
		},
		{
			name:        "GetActivitiesInvalidPagination",
			description: "Test invalid pagination parameters",
			testFunc: func(t *testing.T, client *Client) {
				_, _, err := client.GetActivities(context.Background(), 0, 0)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid pagination parameters")
			},
		},
		{
			name:        "GetActivitiesTimeout",
			description: "Test request timeout handling",
			testFunc: func(t *testing.T, client *Client) {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(2 * time.Second) // Simulate delay
				})
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				_, _, err := client.GetActivities(ctx, 1, 10)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context deadline exceeded")
			},
		},
		{
			name:        "GetActivitiesInvalidResponse",
			description: "Test invalid response handling",
			testFunc: func(t *testing.T, client *Client) {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("invalid json"))
				})
				_, _, err := client.GetActivities(context.Background(), 1, 10)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to parse response")
			},
		},
		{
			name:        "GetActivitiesLargeDataset",
			description: "Test handling of large activity datasets",
			testFunc: func(t *testing.T, client *Client) {
				// Create large dataset
				var activities []ActivityResponse
				for i := 0; i < 500; i++ {
					activities = append(activities, ActivityResponse{
						ActivityID: int64(i + 1),
						Name:       fmt.Sprintf("Activity %d", i+1),
					})
				}

				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(ActivitiesResponse{
						Activities: activities,
						Pagination: Pagination{
							Page:       1,
							PageSize:   500,
							TotalCount: 500,
						},
					})
				})

				result, pagination, err := client.GetActivities(context.Background(), 1, 500)
				assert.NoError(t, err)
				assert.Len(t, result, 500)
				assert.Equal(t, 500, pagination.TotalCount)
			},
		},
		{
			name:        "GetActivityDetailsInvalidResponse",
			description: "Test invalid activity details response",
			testFunc: func(t *testing.T, client *Client) {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("invalid json"))
				})
				_, err := client.GetActivityDetails(context.Background(), 1)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to parse response")
			},
		},
		{
			name:        "GetActivityDetailsMalformedID",
			description: "Test handling of malformed activity ID in server response",
			testFunc: func(t *testing.T, client *Client) {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"activityId": "invalid"}`)) // Should be number
				})
				_, err := client.GetActivityDetails(context.Background(), 1)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to parse response")
			},
		},
		{
			name:        "UploadActivitySuccess",
			description: "Test successful activity upload",
			testFunc: func(t *testing.T, client *Client) {
				id, err := client.UploadActivity(context.Background(), []byte("test fit data"))
				assert.NoError(t, err)
				assert.Equal(t, int64(12345), id)
			},
		},
		{
			name:        "UploadActivityInvalidData",
			description: "Test uploading invalid FIT data",
			testFunc: func(t *testing.T, client *Client) {
				mockServer.SetUploadHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "Invalid FIT file"}`))
				})
				_, err := client.UploadActivity(context.Background(), []byte("invalid"))
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "upload failed with status 400")
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.description)
			tc.testFunc(t, client)
		})
	}
}
