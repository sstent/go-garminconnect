package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// GearStats represents detailed statistics for a gear item
type GearStats struct {
	UUID            string  `json:"uuid"`            // Unique identifier for the gear item
	Name            string  `json:"name"`            // Display name of the gear item
	Distance        float64 `json:"distance"`        // in meters
	TotalActivities int     `json:"totalActivities"` // number of activities
	TotalTime       int     `json:"totalTime"`       // in seconds
	Calories        int     `json:"calories"`        // total calories
	ElevationGain   float64 `json:"elevationGain"`   // in meters
	ElevationLoss   float64 `json:"elevationLoss"`   // in meters
}

// GearActivity represents a simplified activity linked to a gear item
type GearActivity struct {
	ActivityID   int64     `json:"activityId"`     // Activity identifier
	ActivityName string    `json:"activityName"`   // Name of the activity
	StartTime    time.Time `json:"startTimeLocal"` // Local start time of the activity
	Duration     int       `json:"duration"`       // Duration in seconds
	Distance     float64   `json:"distance"`       // Distance in meters
}

// GetGearStats retrieves statistics for a specific gear item by its UUID
func (c *Client) GetGearStats(ctx context.Context, gearUUID string) (GearStats, error) {
	endpoint := fmt.Sprintf("/gear-service/stats/%s", gearUUID)

	var stats GearStats
	err := c.Get(ctx, endpoint, &stats)
	if err != nil {
		return GearStats{}, err
	}

	return stats, nil
}

// GetGearActivities retrieves paginated activities associated with a gear item
func (c *Client) GetGearActivities(ctx context.Context, gearUUID string, start, limit int) ([]GearActivity, error) {
	path := fmt.Sprintf("/gear-service/activities/%s", gearUUID)
	params := url.Values{}
	params.Add("start", strconv.Itoa(start))
	params.Add("limit", strconv.Itoa(limit))

	var activities []GearActivity
	err := c.Get(ctx, fmt.Sprintf("%s?%s", path, params.Encode()), &activities)
	if err != nil {
		return nil, fmt.Errorf("failed to get gear activities: %w", err)
	}

	return activities, nil
}
