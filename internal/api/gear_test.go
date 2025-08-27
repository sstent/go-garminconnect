package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGearService(t *testing.T) {
	// Create test server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gear-service/stats/valid-uuid":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GearStats{
				UUID:            "valid-uuid",
				Name:            "Test Gear",
				Distance:        1500.5,
				TotalActivities: 10,
				TotalTime:       3600,
			})
		case "/gear-service/stats/invalid-uuid":
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, `{"message": "gear not found"}`)
		case "/gear-service/activities/valid-uuid":
			startStr := r.URL.Query().Get("start")
			limitStr := r.URL.Query().Get("limit")
			start, _ := strconv.Atoi(startStr)
			limit, _ := strconv.Atoi(limitStr)

			activities := []GearActivity{
				{ActivityID: 1, ActivityName: "Run 1", StartTime: time.Now(), Duration: 1800, Distance: 5000},
				{ActivityID: 2, ActivityName: "Run 2", StartTime: time.Now().Add(-24*time.Hour), Duration: 3600, Distance: 10000},
			}

			// Simulate pagination
			if start < 0 {
				start = 0
			}
			end := start + limit
			if end > len(activities) {
				end = len(activities)
			}
			if start > len(activities) {
				start = len(activities)
				end = len(activities)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(activities[start:end])
		case "/gear-service/activities/invalid-uuid":
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, `{"message": "gear activities not found"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Create client
	client, _ := NewClient(srv.URL, http.DefaultClient)
	client.SetLogger(NewTestLogger(t))

	t.Run("GetGearStats success", func(t *testing.T) {
		stats, err := client.GetGearStats("valid-uuid")
		assert.NoError(t, err)
		assert.Equal(t, "Test Gear", stats.Name)
		assert.Equal(t, 1500.5, stats.Distance)
	})

	t.Run("GetGearStats not found", func(t *testing.T) {
		_, err := client.GetGearStats("invalid-uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status code: 404")
	})

	t.Run("GetGearActivities pagination", func(t *testing.T) {
		activities, err := client.GetGearActivities("valid-uuid", 0, 1)
		assert.NoError(t, err)
		assert.Len(t, activities, 1)
		assert.Equal(t, "Run 1", activities[0].ActivityName)

		activities, err = client.GetGearActivities("valid-uuid", 1, 1)
		assert.NoError(t, err)
		assert.Len(t, activities, 1)
		assert.Equal(t, "Run 2", activities[0].ActivityName)
		
		_, err = client.GetGearActivities("invalid-uuid", 0, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status code: 404")
	})
}
