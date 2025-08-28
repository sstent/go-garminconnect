package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sstent/go-garminconnect/internal/fit"
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

// ActivityDetail represents comprehensive activity data
type ActivityDetail struct {
	Activity
	Calories      float64         `json:"calories"`
	AverageHR     int             `json:"averageHR"`
	MaxHR         int             `json:"maxHR"`
	AverageTemp   float64         `json:"averageTemperature"`
	ElevationGain float64         `json:"elevationGain"`
	ElevationLoss float64         `json:"elevationLoss"`
	Weather       Weather         `json:"weather"`
	Gear          Gear            `json:"gear"`
	GPSTracks     []GPSTrackPoint `json:"gpsTracks"`
}

// garminTime implements custom JSON unmarshaling for Garmin's time format
type garminTime struct {
	time.Time
}

const garminTimeLayout = "2006-01-02T15:04:05"

func (gt *garminTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(garminTimeLayout, s)
	if err != nil {
		return err
	}
	gt.Time = t
	return nil
}

// ActivityResponse is used for JSON unmarshaling with custom time handling
type ActivityResponse struct {
	ActivityID int64      `json:"activityId"`
	Name       string     `json:"activityName"`
	Type       string     `json:"activityType"`
	StartTime  garminTime `json:"startTimeLocal"`
	Duration   float64    `json:"duration"`
	Distance   float64    `json:"distance"`
}

// ActivityDetailResponse is used for JSON unmarshaling with custom time handling
type ActivityDetailResponse struct {
	ActivityResponse
	Calories      float64         `json:"calories"`
	AverageHR     int             `json:"averageHR"`
	MaxHR         int             `json:"maxHR"`
	AverageTemp   float64         `json:"averageTemperature"`
	ElevationGain float64         `json:"elevationGain"`
	ElevationLoss float64         `json:"elevationLoss"`
	Weather       Weather         `json:"weather"`
	Gear          Gear            `json:"gear"`
	GPSTracks     []GPSTrackPoint `json:"gpsTracks"`
}

// Convert to ActivityDetail
func (adr *ActivityDetailResponse) ToActivityDetail() ActivityDetail {
	return ActivityDetail{
		Activity: Activity{
			ActivityID: adr.ActivityID,
			Name:       adr.Name,
			Type:       adr.Type,
			StartTime:  adr.StartTime.Time,
			Duration:   adr.Duration,
			Distance:   adr.Distance,
		},
		Calories:      adr.Calories,
		AverageHR:     adr.AverageHR,
		MaxHR:         adr.MaxHR,
		AverageTemp:   adr.AverageTemp,
		ElevationGain: adr.ElevationGain,
		ElevationLoss: adr.ElevationLoss,
		Weather:       adr.Weather,
		Gear:          adr.Gear,
		GPSTracks:     adr.GPSTracks,
	}
}

// Convert to Activity
func (ar *ActivityResponse) ToActivity() Activity {
	return Activity{
		ActivityID: ar.ActivityID,
		Name:       ar.Name,
		Type:       ar.Type,
		StartTime:  ar.StartTime.Time,
		Duration:   ar.Duration,
		Distance:   ar.Distance,
	}
}

// Weather contains weather conditions during activity
type Weather struct {
	Condition   string  `json:"condition"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

// Gear represents equipment used in activity
type Gear struct {
	ID          string `json:"gearId"`
	Name        string `json:"name"`
	Model       string `json:"model"`
	Description string `json:"description"`
}

// GPSTrackPoint contains geo coordinates
type GPSTrackPoint struct {
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Ele       float64   `json:"ele"`
	Timestamp time.Time `json:"timestamp"`
}

func (gtp *GPSTrackPoint) UnmarshalJSON(data []byte) error {
	type Alias GPSTrackPoint
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(gtp),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Timestamp != "" {
		t, err := time.Parse(garminTimeLayout, aux.Timestamp)
		if err != nil {
			return err
		}
		gtp.Timestamp = t
	}
	return nil
}

// ActivitiesResponse represents the response from the activities endpoint
type ActivitiesResponse struct {
	Activities []ActivityResponse `json:"activities"`
	Pagination Pagination         `json:"pagination"`
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
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("pageSize", strconv.Itoa(pageSize))

	var response ActivitiesResponse
	err := c.Get(ctx, fmt.Sprintf("%s?%s", path, params.Encode()), &response)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get activities: %w", err)
	}

	// Convert response to Activity slice
	activities := make([]Activity, len(response.Activities))
	for i, ar := range response.Activities {
		activities[i] = ar.ToActivity()
	}

	// Validate we received some activities
	if len(activities) == 0 {
		return nil, nil, fmt.Errorf("no activities found")
	}

	return activities, &response.Pagination, nil
}

// GetActivityDetails retrieves comprehensive data for a specific activity
func (c *Client) GetActivityDetails(ctx context.Context, activityID int64) (*ActivityDetail, error) {
	path := fmt.Sprintf("/activity-service/activity/%d", activityID)

	var response ActivityDetailResponse
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity details: %w", err)
	}

	activityDetail := response.ToActivityDetail()

	// Validate we received activity data
	if activityDetail.ActivityID == 0 {
		return nil, fmt.Errorf("no activity found for ID %d", activityID)
	}

	return &activityDetail, nil
}

// UploadActivity handles FIT file uploads
func (c *Client) UploadActivity(ctx context.Context, fitFile []byte) (int64, error) {
	// Validate FIT file
	if err := fit.ValidateFIT(fitFile); err != nil {
		return 0, fmt.Errorf("invalid FIT file: %w", err)
	}

	// Refresh token if needed
	if err := c.refreshTokenIfNeeded(); err != nil {
		return 0, err
	}

	path := "/upload-service/upload/.fit"

	resp, err := c.HTTPClient.R().
		SetContext(ctx).
		SetFileReader("file", "activity.fit", bytes.NewReader(fitFile)).
		SetHeader("Content-Type", "multipart/form-data").
		Post(path)

	if err != nil {
		return 0, err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return 0, errors.New("token expired, please reauthenticate")
	}

	if resp.StatusCode() >= 400 {
		return 0, handleAPIError(resp)
	}

	// Parse response to get activity ID
	var result struct {
		ActivityID int64 `json:"activityId"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, err
	}

	return result.ActivityID, nil
}

// DownloadActivity retrieves a FIT file for an activity
func (c *Client) DownloadActivity(ctx context.Context, activityID int64) ([]byte, error) {
	// Refresh token if needed
	if err := c.refreshTokenIfNeeded(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/download-service/export/activity/%d", activityID)

	resp, err := c.HTTPClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/fit").
		Get(path)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, errors.New("token expired, please reauthenticate")
	}

	if resp.StatusCode() >= 400 {
		return nil, handleAPIError(resp)
	}

	return resp.Body(), nil
}
