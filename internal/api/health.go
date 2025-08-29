package api

import (
	"context"
	"fmt"
	"time"
)

// HRVData represents Heart Rate Variability data
type HRVData struct {
	Date               time.Time `json:"date"`
	RestingHrv         float64   `json:"restingHrv"`
	WeeklyAvg          float64   `json:"weeklyAvg"`
	LastNightAvg       float64   `json:"lastNightAvg"`
	HrvStatus          string    `json:"hrvStatus"`
	HrvStatusMessage   string    `json:"hrvStatusMessage"`
	BaselineHrv        int       `json:"baselineHrv"`
	ChangeFromBaseline int       `json:"changeFromBaseline"`
}

// BodyBatteryData represents Garmin's Body Battery energy metric
type BodyBatteryData struct {
	Date    time.Time `json:"date"`
	Charged int       `json:"charged"` // 0-100 scale
	Drained int       `json:"drained"` // 0-100 scale
	Highest int       `json:"highest"` // highest value of the day
	Lowest  int       `json:"lowest"`  // lowest value of the day
}

// GetSleepData retrieves sleep data for a specific date
func (c *Client) GetSleepData(ctx context.Context, date time.Time) (*SleepData, error) {
	var data SleepData
	path := fmt.Sprintf("/wellness-service/sleep/daily/%s", date.Format("2006-01-02"))

	if err := c.Get(ctx, path, &data); err != nil {
		return nil, fmt.Errorf("failed to get sleep data: %w", err)
	}
	return &data, nil
}

// GetHRVData retrieves Heart Rate Variability data for a specific date
func (c *Client) GetHRVData(ctx context.Context, date time.Time) (*HRVData, error) {
	var data HRVData
	path := fmt.Sprintf("/hrv-service/hrv/%s", date.Format("2006-01-02"))

	if err := c.Get(ctx, path, &data); err != nil {
		return nil, fmt.Errorf("failed to get HRV data: %w", err)
	}
	return &data, nil
}

// GetStressData retrieves stress data for a specific date
func (c *Client) GetStressData(ctx context.Context, date time.Time) (*DailyStress, error) {
	var data DailyStress
	path := fmt.Sprintf("/wellness-service/stress/daily/%s", date.Format("2006-01-02"))

	if err := c.Get(ctx, path, &data); err != nil {
		return nil, fmt.Errorf("failed to get stress data: %w", err)
	}
	return &data, nil
}

// GetStepsData retrieves step count data for a specific date
func (c *Client) GetStepsData(ctx context.Context, date time.Time) (*DailySteps, error) {
	var data DailySteps
	path := fmt.Sprintf("/wellness-service/steps/daily/%s", date.Format("2006-01-02"))

	if err := c.Get(ctx, path, &data); err != nil {
		return nil, fmt.Errorf("failed to get steps data: %w", err)
	}
	return &data, nil
}

// GetBodyBatteryData retrieves Body Battery data for a specific date
func (c *Client) GetBodyBatteryData(ctx context.Context, date time.Time) (*BodyBatteryData, error) {
	var data BodyBatteryData
	path := fmt.Sprintf("/bodybattery-service/bodybattery/%s", date.Format("2006-01-02"))

	if err := c.Get(ctx, path, &data); err != nil {
		return nil, fmt.Errorf("failed to get Body Battery data: %w", err)
	}
	return &data, nil
}
