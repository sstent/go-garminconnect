package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// BodyCompositionHandler handles mock responses for body composition endpoint
func BodyCompositionHandler(w http.ResponseWriter, r *http.Request) {
	// Validate parameters
	start := r.URL.Query().Get("startDate")
	end := r.URL.Query().Get("endDate")
	if start == "" || end == "" || start > end {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Return different responses based on test cases
	if r.Header.Get("Authorization") == "" || strings.Contains(r.Header.Get("Authorization"), "invalid") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Successful response
	data := []BodyComposition{
		{
			BoneMass:   2.8,
			MuscleMass: 55.2,
			BodyFat:    15.3,
			Hydration:  58.7,
			Timestamp:  Time(parseTime("2023-01-15T08:00:00Z")),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// parseTime helper for creating time values in mock handlers
func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
