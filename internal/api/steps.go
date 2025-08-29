package api

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// DailySteps represents daily step count data from Garmin Connect
type DailySteps struct {
	CalendarDate     time.Time `json:"calendarDate" validate:"required"`
	TotalSteps       int       `json:"totalSteps" validate:"min=0"`
	Goal             int       `json:"goal" validate:"min=0"`
	ActiveMinutes    int       `json:"activeMinutes" validate:"min=0"`
	DistanceMeters   float64   `json:"distanceMeters" validate:"min=0"`
	CaloriesBurned   int       `json:"caloriesBurned" validate:"min=0"`
	StepsToGoal      int       `json:"stepsToGoal"`
	StepGoalAchieved bool      `json:"stepGoalAchieved"`
}

// Validate ensures DailySteps fields meet requirements
func (s *DailySteps) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}
