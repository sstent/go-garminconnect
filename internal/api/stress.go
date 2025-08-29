package api

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// DailyStress represents daily stress data from Garmin Connect
type DailyStress struct {
	CalendarDate         time.Time `json:"calendarDate" validate:"required"`
	OverallStressLevel   int       `json:"overallStressLevel" validate:"min=0,max=100"`
	RestStressDuration   int       `json:"restStressDuration" validate:"min=0"`
	LowStressDuration    int       `json:"lowStressDuration" validate:"min=0"`
	MediumStressDuration int       `json:"mediumStressDuration" validate:"min=0"`
	HighStressDuration   int       `json:"highStressDuration" validate:"min=0"`
	StressQualifier      string    `json:"stressQualifier"`
}

// Validate ensures DailyStress fields meet requirements
func (s *DailyStress) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}
