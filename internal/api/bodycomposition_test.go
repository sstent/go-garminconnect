package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sstent/go-garminconnect/internal/auth/garth"
	"github.com/stretchr/testify/assert"
)

func TestGetBodyComposition(t *testing.T) {
	// Create test server for mocking API responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for required parameters without enforcing order
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")
		
		assert.Equal(t, "2023-01-01", startDate, "startDate should match")
		assert.Equal(t, "2023-01-31", endDate, "endDate should match")

		// Return different responses based on test cases
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if startDate == "2023-02-01" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Successful response with proper timestamp format
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"boneMass": 2.8,
				"muscleMass": 55.2,
				"bodyFat": 15.3,
				"hydration": 58.7,
				"timestamp": "2023-01-15T08:00:00Z"
			}
		]`))
	}))
	defer server.Close()

	// Test cases
	testCases := []struct {
		name        string
		token       string
		start       time.Time
		end         time.Time
		expectError bool
		expectedLen int
	}{
		{
			name:        "Successful request",
			token:       "valid-token",
			start:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			end:         time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC),
			expectError: false,
			expectedLen: 1,
		},
		{
			name:        "Unauthorized access",
			token:       "invalid-token",
			start:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			end:         time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC),
			expectError: true,
		},
		{
			name:        "Invalid date range",
			token:       "valid-token",
			start:       time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			end:         time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create session with appropriate token
			session := &garth.Session{
				OAuth2Token: tc.token,
				ExpiresAt:   time.Now().Add(8 * time.Hour), // Not expired
			}

			// Setup client with test server
			client, err := NewClient(session, "")
			assert.NoError(t, err)
			client.HTTPClient.SetBaseURL(server.URL)

			results, err := client.GetBodyComposition(context.Background(), BodyCompositionRequest{
				StartDate: Time(tc.start),
				EndDate:   Time(tc.end),
			})

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, results, tc.expectedLen)

			if tc.expectedLen > 0 {
				result := results[0]
				assert.Equal(t, 2.8, result.BoneMass)
				assert.Equal(t, 55.2, result.MuscleMass)
				assert.Equal(t, 15.3, result.BodyFat)
				assert.Equal(t, 58.7, result.Hydration)
			}
		})
	}
}