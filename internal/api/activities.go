package api

import (
	"context"
	"fmt"
	"time"
)

// Activity represents a Garmin Connect activity
type Activity struct {
	ActivityID int64     `json:"activityId"`
	Name       string    `json:"activityName"`
	Type       string    `json:"activityType"`
	StartTime  time.Time `json:"startTimeLocal"`
	Duration   float64   `json:"duration"`
	Distance   float64   `json:"distance"`
}

// ActivitiesResponse represents the response from the activities endpoint
type ActivitiesResponse struct {
	Activities []Activity `json:"activities"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination information in API responses
type Pagination struct {
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
	Page       int `json:"page"`
}

// GetActivities retrieves a list of activities with pagination
func (c *Client) GetActivities(ctx context.Context, page int, pageSize int) ([]Activity, *Pagination, error) {
	path := "/activitylist-service/activities/search"
	query := fmt.Sprintf("?page=%d&pageSize=%d", page, pageSize)

	var response ActivitiesResponse
	err := c.Get(ctx, path+query, &response)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get activities: %w", err)
	}

	// Validate we received some activities
	if len(response.Activities) == 0 {
		return nil, nil, fmt.Errorf("no activities found")
	}

	return response.Activities, &response.Pagination, nil
}
