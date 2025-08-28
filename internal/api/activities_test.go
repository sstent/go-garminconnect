package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestActivitiesEndpoints(t *testing.T) {
	mockServer := NewMockServer()
	defer mockServer.Close()
	client := NewClientWithBaseURL(mockServer.URL)

	tests := []struct {
		name        string
		setup       func()
		testFunc    func(t *testing.T)
	}{
		{
			name: "GetActivitiesSuccess",
			setup: func() {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					activities := []ActivityResponse{{
						ActivityID: 1,
						Name:       "Morning Run",
						Type:       "RUNNING",
						StartTime:  garminTime{time.Now().Add(-24 * time.Hour)},
						Duration:   3600,
						Distance:   10.0,
					}}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ActivitiesResponse{
						Activities: activities,
						Pagination: Pagination{Page: 1, PageSize: 10, TotalCount: 1},
					})
				})
			},
			testFunc: func(t *testing.T) {
				activities, pagination, err := client.GetActivities(context.Background(), 1, 10)
				assert.NoError(t, err)
				assert.Len(t, activities, 1)
				assert.Equal(t, int64(1), activities[0].ActivityID)
				assert.Equal(t, "Morning Run", activities[0].Name)
				assert.NotNil(t, pagination)
				assert.Equal(t, 1, pagination.TotalCount)
			},
		},
		{
			name: "GetActivityDetailsSuccess",
			setup: func() {
				mockServer.SetActivityDetailsHandler(func(w http.ResponseWriter, r *http.Request) {
					pathParts := strings.Split(r.URL.Path, "/")
					if len(pathParts) < 2 {
						w.WriteHeader(http.StatusNotFound)
						return
					}

					activityID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ActivityDetailResponse{
						ActivityResponse: ActivityResponse{
							ActivityID: activityID,
							Name:       "Mock Activity",
							Type:       "RUNNING",
							StartTime:  garminTime{time.Now().Add(-24 * time.Hour)},
							Duration:   3600,
							Distance:   10.0,
						},
						Calories:      500,
						AverageHR:     150,
						MaxHR:         170,
						ElevationGain: 100,
					})
				})
			},
			testFunc: func(t *testing.T) {
				activity, err := client.GetActivityDetails(context.Background(), 1)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), activity.ActivityID)
				assert.Equal(t, "Mock Activity", activity.Name)
				assert.Equal(t, float64(500), activity.Calories)
			},
		},
		{
			name: "GetActivitiesServerError",
			setup: func() {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "Internal server error"}`))
				})
			},
			testFunc: func(t *testing.T) {
				_, _, err := client.GetActivities(context.Background(), 1, 10)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get activities")
			},
		},
		{
			name: "GetActivitiesEmptyResponse",
			setup: func() {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ActivitiesResponse{
						Activities: []ActivityResponse{}, // Empty activities
						Pagination: Pagination{Page: 1, PageSize: 10, TotalCount: 0},
					})
				})
			},
			testFunc: func(t *testing.T) {
				activities, pagination, err := client.GetActivities(context.Background(), 1, 10)
				assert.NoError(t, err) // Should not error when no activities exist
				assert.Len(t, activities, 0)
				assert.NotNil(t, pagination)
				assert.Equal(t, 0, pagination.TotalCount)
			},
		},
		{
			name: "UploadActivitySuccess",
			setup: func() {
				mockServer.SetUploadHandler(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(map[string]interface{}{"activityId": 12345})
				})
			},
			testFunc: func(t *testing.T) {
				// Create a minimal valid FIT file data for testing
				fitData := make([]byte, 20)
				fitData[0] = 14 // header size
				copy(fitData[8:12], []byte(".FIT"))
				
				id, err := client.UploadActivity(context.Background(), fitData)
				assert.NoError(t, err)
				assert.Equal(t, int64(12345), id)
			},
		},
		{
			name: "GetActivityDetailsNotFound",
			setup: func() {
				mockServer.SetActivityDetailsHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"error": "Activity not found"}`))
				})
			},
			testFunc: func(t *testing.T) {
				_, err := client.GetActivityDetails(context.Background(), 999)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get activity details")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer.Reset()
			if tt.setup != nil {
				tt.setup()
			}
			tt.testFunc(t)
		})
	}
}