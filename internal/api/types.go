package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// APIError represents an error returned by the API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API request failed with status %d: %s", e.StatusCode, e.Message)
}

// Error types for API responses
type ErrNotFound struct{}

func (e ErrNotFound) Error() string { return "resource not found" }

type ErrBadRequest struct{}

func (e ErrBadRequest) Error() string { return "bad request" }

// Time represents a Garmin Connect time value
type Time time.Time

// IsZero checks if the time is zero value
func (t Time) IsZero() bool {
	return time.Time(t).IsZero()
}

// After reports whether t is after u
func (t Time) After(u Time) bool {
	return time.Time(t).After(time.Time(u))
}

// Format formats the time using the provided layout
func (t Time) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (t *Time) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Try multiple time formats that Garmin might use
	formats := []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, s); err == nil {
			*t = Time(parsedTime)
			return nil
		}
	}

	// If none of the formats work, try parsing as RFC3339
	parsedTime, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = Time(parsedTime)
	return nil
}

// MarshalJSON implements json.Marshaler interface
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format(time.RFC3339))
}

// BodyComposition represents body composition metrics from Garmin Connect
type BodyComposition struct {
	BoneMass   float64 `json:"boneMass"`   // Grams
	MuscleMass float64 `json:"muscleMass"` // Grams
	BodyFat    float64 `json:"bodyFat"`    // Percentage
	Hydration  float64 `json:"hydration"`  // Percentage
	Timestamp  Time    `json:"timestamp"`  // Measurement time
}

// BodyCompositionRequest defines parameters for body composition API requests
type BodyCompositionRequest struct {
	StartDate Time `json:"startDate"`
	EndDate   Time `json:"endDate"`
}
