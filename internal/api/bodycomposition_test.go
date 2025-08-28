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
	// Create mock session
	session := &garth.Session{OAuth2Token: "valid-token"}

	// Create test server for mocking API responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/body-composition?startDate=2023-01-01&endDate=2023-01-31", r.URL.String())

		// Return different responses based on test cases
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Query().Get("startDate") == "2023-02-01" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"boneMass": 2.8,
				"muscleMass": 55.2,
				"bodyFat": 15.3,
				"hydration": 58.7,
				"timestamp": "2023-01-15T08:00:00.000Z"
			}
		]`))
	}))
	defer server.Close()

	// Setup client with test server
	client, _ := NewClient(session, "")
	client.HTTPClient.SetBaseURL(server.URL)

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
			token:       "valid-token", // Test case doesn't actually change client token now
			start:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			end:         time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC),
			expectError: false,
			expectedLen: 1,
		},
		// Unauthorized test case is handled by the mock server's token check
		// We need to create a new client with invalid token
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
			// For unauthorized test, create a separate client
			if tc.token == "invalid-token" {
				invalidSession := &garth.Session{OAuth2Token: "invalid-token"}
				invalidClient, _ := NewClient(invalidSession, "")
				invalidClient.HTTPClient.SetBaseURL(server.URL)
				client = invalidClient
			} else {
				validSession := &garth.Session{OAuth2Token: "valid-token"}
				validClient, _ := NewClient(validSession, "")
				validClient.HTTPClient.SetBaseURL(server.URL)
				client = validClient
			}
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
