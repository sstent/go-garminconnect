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
	client := NewClientWithBaseURL(mockServer.URL())

	// Setup standard mock handlers
	mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
		activities := []ActivityResponse{{
			ActivityID: 1,
			Name:       "Morning Run",
		}}
		json.NewEncoder(w).Encode(ActivitiesResponse{
			Activities: activities,
			Pagination: Pagination{Page: 1, PageSize: 10, TotalCount: 1},
		})
	})

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

		json.NewEncoder(w).Encode(ActivityDetailResponse{
			ActivityResponse: ActivityResponse{
				ActivityID: activityID,
				Name:       "Mock Activity",
				Type:       "RUNNING",
				StartTime:  garminTime{time.Now().Add(-24 * time.Hour)},
			},
		})
	})

	mockServer.SetUploadHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"activityId": 12345})
	})

	tests := []struct {
		name        string
		setup       func()
		testFunc    func(t *testing.T)
	}{
		{
			name: "GetActivitiesSuccess",
			testFunc: func(t *testing.T) {
				activities, _, err := client.GetActivities(context.Background(), 1, 10)
				assert.NoError(t, err)
				assert.Len(t, activities, 1)
				assert.Equal(t, int64(1), activities[0].ActivityID)
			},
		},
		{
			name: "GetActivityDetailsSuccess",
			testFunc: func(t *testing.T) {
				activity, err := client.GetActivityDetails(context.Background(), 1)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), activity.ActivityID)
			},
		},
		{
			name: "GetActivitiesServerError",
			setup: func() {
				mockServer.SetActivitiesHandler(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				})
			},
			testFunc: func(t *testing.T) {
				_, _, err := client.GetActivities(context.Background(), 1, 10)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get activities")
			},
		},
		{
			name: "UploadActivitySuccess",
			testFunc: func(t *testing.T) {
				id, err := client.UploadActivity(context.Background(), []byte("test fit data"))
				assert.NoError(t, err)
				assert.Equal(t, int64(12345), id)
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
