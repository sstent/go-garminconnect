package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GearStats represents detailed statistics for a gear item
type GearStats struct {
	UUID             string  `json:"uuid"`             // Unique identifier for the gear item
	Name             string  `json:"name"`             // Display name of the gear item
	Distance         float64 `json:"distance"`         // in meters
	TotalActivities  int     `json:"totalActivities"`  // number of activities
	TotalTime        int     `json:"totalTime"`        // in seconds
	Calories         int     `json:"calories"`         // total calories
	ElevationGain    float64 `json:"elevationGain"`    // in meters
	ElevationLoss    float64 `json:"elevationLoss"`    // in meters
}

// GearActivity represents a simplified activity linked to a gear item
type GearActivity struct {
	ActivityID   int64     `json:"activityId"`          // Activity identifier
	ActivityName string    `json:"activityName"`        // Name of the activity
	StartTime    time.Time `json:"startTimeLocal"`      // Local start time of the activity
	Duration     int       `json:"duration"`            // Duration in seconds
	Distance     float64   `json:"distance"`            // Distance in meters
}

// GetGearStats retrieves statistics for a specific gear item by its UUID.
// Returns a GearStats struct containing gear usage metrics or an error.
func (c *Client) GetGearStats(gearUUID string) (GearStats, error) {
	endpoint := "gear-service/stats/" + gearUUID
	req, err := c.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return GearStats{}, err
	}

	var stats GearStats
	_, err = c.do(req, &stats)
	if err != nil {
		return GearStats{}, err
	}

	return stats, nil
}

// GetGearActivities retrieves paginated activities associated with a gear item.
// start: pagination start index
// limit: maximum number of results to return
// Returns a slice of GearActivity structs or an error.
func (c *Client) GetGearActivities(gearUUID string, start, limit int) ([]GearActivity, error) {
	endpoint := "gear-service/activities/" + gearUUID
	params := url.Values{}
	params.Add("start", strconv.Itoa(start))
	params.Add("limit", strconv.Itoa(limit))
	
	req, err := c.newRequest(http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var activities []GearActivity
	_, err = c.do(req, &activities)
	if err != nil {
		return nil, err
	}

	return activities, nil
}

// formatDuration converts total seconds to HH:MM:SS time format.
// Primarily used for displaying activity durations in a human-readable format.
func formatDuration(seconds int) string {
	d := time.Duration(seconds) * time.Second
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds = int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}
