package api

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// HRVSummary represents Heart Rate Variability summary data from Garmin Connect
type HRVSummary struct {
	Date               time.Time `json:"date" validate:"required"`
	RestingHrv         float64   `json:"restingHrv" validate:"min=0"`
	WeeklyAvg          float64   `json:"weeklyAvg" validate:"min=0"`
	LastNightAvg       float64   `json:"lastNightAvg" validate:"min=0"`
	HrvStatus          string    `json:"hrvStatus"`
	HrvStatusMessage   string    `json:"hrvStatusMessage"`
	BaselineHrv        int       `json:"baselineHrv" validate:"min=0"`
	ChangeFromBaseline int       `json:"changeFromBaseline"`
}

// Validate ensures HRVSummary fields meet requirements
func (h *HRVSummary) Validate() error {
	validate := validator.New()
	return validate.Struct(h)
}
