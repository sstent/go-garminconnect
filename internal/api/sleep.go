package api

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// SleepData represents sleep metrics from Garmin Connect
type SleepData struct {
	CalendarDate      time.Time `json:"calendarDate" validate:"required"`
	SleepTimeSeconds  int       `json:"sleepTimeSeconds" validate:"min=0"`
	DeepSleepSeconds  int       `json:"deepSleepSeconds" validate:"min=0"`
	LightSleepSeconds int       `json:"lightSleepSeconds" validate:"min=0"`
	RemSleepSeconds   int       `json:"remSleepSeconds" validate:"min=0"`
	AwakeSeconds      int       `json:"awakeSeconds" validate:"min=0"`
	SleepScore        int       `json:"sleepScore" validate:"min=0,max=100"`
	SleepScores       struct {
		Overall  int `json:"overall"`
		Duration int `json:"duration"`
		Deep     int `json:"deep"`
		Rem      int `json:"rem"`
		Light    int `json:"light"`
		Awake    int `json:"awake"`
	} `json:"sleepScores"`
}

// Validate ensures SleepData fields meet requirements
func (s *SleepData) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}
