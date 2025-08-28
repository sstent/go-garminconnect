package api

import (
	"time"
)

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
