package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sstent/go-garminconnect/internal/auth/garth"
)

// MockServer simulates the Garmin Connect API
type MockServer struct {
	server *httptest.Server
	mu     sync.Mutex
	
	// Endpoint handlers
	activitiesHandler      http.HandlerFunc
	activityDetailsHandler http.HandlerFunc
	uploadHandler          http.HandlerFunc
	userHandler            http.HandlerFunc
	healthHandler          http.HandlerFunc
	authHandler            http.HandlerFunc
	
	// Request counters
	requestCounters map[string]int
}

// NewMockServer creates a new mock Garmin Connect server
func NewMockServer() *MockServer {
	m := &MockServer{
		requestCounters: make(map[string]int),
	}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		// Track request count
		if m.requestCounters == nil {
			m.requestCounters = make(map[string]int)
		}
		endpointType := "unknown"
		path := r.URL.Path
		switch {
		case strings.HasPrefix(path, "/activitylist-service/activities/search") || path == "/activitylist-service/activities":
			endpointType = "activities"
			m.handleActivities(w, r)
		case strings.HasPrefix(path, "/activity-service/activities") || path == "/activity-service/activities":
			endpointType = "activities"
			m.handleActivities(w, r)
		case strings.HasPrefix(path, "/activity-service/activity/"):
			endpointType = "activityDetails"
			m.handleActivityDetails(w, r)
		case strings.HasPrefix(path, "/upload-service/upload") || path == "/upload-service/upload":
			endpointType = "upload"
			m.handleUpload(w, r)
		case strings.HasPrefix(path, "/user-service/user") || path == "/user-service/user":
			endpointType = "user"
			m.handleUserData(w, r)
		case strings.HasPrefix(path, "/health-service") || path == "/health-service":
			endpointType = "health"
			m.handleHealthData(w, r)
		case strings.HasPrefix(path, "/auth") || path == "/auth":
			endpointType = "auth"
			m.handleAuth(w, r)
		default:
			endpointType = "unknown"
			http.Error(w, "Not found", http.StatusNotFound)
		}
		m.requestCounters[endpointType]++
	}))
	return m
}

// URL returns the base URL of the mock server
func (m *MockServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server
func (m *MockServer) Close() {
	m.server.Close()
}

// SetActivitiesHandler sets a custom handler for activities endpoint
func (m *MockServer) SetActivitiesHandler(handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activitiesHandler = handler
}

// SetUploadHandler sets a custom handler for upload endpoint
func (m *MockServer) SetUploadHandler(handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploadHandler = handler
}

// SetActivityDetailsHandler sets a custom handler for activity details endpoint
func (m *MockServer) SetActivityDetailsHandler(handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activityDetailsHandler = handler
}

// SetUserHandler sets a custom handler for user endpoint
func (m *MockServer) SetUserHandler(handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userHandler = handler
}

// SetHealthHandler sets a custom handler for health endpoint
func (m *MockServer) SetHealthHandler(handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthHandler = handler
}

// SetAuthHandler sets a custom handler for auth endpoint
func (m *MockServer) SetAuthHandler(handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authHandler = handler
}

// Reset resets all handlers and counters to default state
func (m *MockServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activitiesHandler = nil
	m.activityDetailsHandler = nil
	m.uploadHandler = nil
	m.userHandler = nil
	m.healthHandler = nil
	m.authHandler = nil
	m.requestCounters = make(map[string]int)
}

// RequestCount returns the number of requests made to a specific endpoint
func (m *MockServer) RequestCount(endpoint string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requestCounters[endpoint]
}

// SetResponse sets a standardized response for a specific endpoint
func (m *MockServer) SetResponse(endpoint string, status int, body interface{}) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(body)
	}

	switch endpoint {
	case "activities":
		m.SetActivitiesHandler(handler)
	case "activityDetails":
		m.SetActivityDetailsHandler(handler)
	case "upload":
		m.SetUploadHandler(handler)
	case "user":
		m.SetUserHandler(handler)
	case "health":
		m.SetHealthHandler(handler)
	case "auth":
		m.SetAuthHandler(handler)
	}
}

// SetErrorResponse configures an error response for a specific endpoint
func (m *MockServer) SetErrorResponse(endpoint string, status int, message string) {
	m.SetResponse(endpoint, status, map[string]string{"error": message})
}

// Default handler implementations would follow for each endpoint
// ...

// handleActivities is the default activities endpoint handler
func (m *MockServer) handleActivities(w http.ResponseWriter, r *http.Request) {
	if m.activitiesHandler != nil {
		m.activitiesHandler(w, r)
		return
	}
	// Default implementation
	activities := []ActivityResponse{
		{
			ActivityID: 1,
			Name:       "Morning Run",
			StartTime:  garminTime{time.Now().Add(-24 * time.Hour)},
			Duration:   3600,
			Distance:   10.0,
		},
	}
	json.NewEncoder(w).Encode(ActivitiesResponse{
		Activities: activities,
		Pagination: Pagination{
			Page:       1,
			PageSize:   10,
			TotalCount: 1,
		},
	})
}

// handleActivityDetails is the default activity details endpoint handler
func (m *MockServer) handleActivityDetails(w http.ResponseWriter, r *http.Request) {
	if m.activityDetailsHandler != nil {
		m.activityDetailsHandler(w, r)
		return
	}
	// Extract activity ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	activityID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	activity := ActivityDetailResponse{
		ActivityResponse: ActivityResponse{
			ActivityID: activityID,
			Name:       "Mock Activity",
			Type:       "RUNNING",
			StartTime:  garminTime{time.Now().Add(-24 * time.Hour)},
			Duration:   3600,
			Distance:   10.0,
		},
		Calories:      500,
		AverageHR:     150,
		MaxHR:         170,
		ElevationGain: 100,
	}

	json.NewEncoder(w).Encode(activity)
}

// handleUpload is the default activity upload handler
func (m *MockServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if m.uploadHandler != nil {
		m.uploadHandler(w, r)
		return
	}
	// Simulate successful upload
	response := map[string]interface{}{
		"activityId": 12345,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleUserData is the default user data handler
func (m *MockServer) handleUserData(w http.ResponseWriter, r *http.Request) {
	if m.userHandler != nil {
		m.userHandler(w, r)
		return
	}
	// Return mock user data
	user := map[string]interface{}{
		"displayName": "Mock User",
		"email":       "mock@example.com",
	}
	json.NewEncoder(w).Encode(user)
}

// handleHealthData is the default health data handler
func (m *MockServer) handleHealthData(w http.ResponseWriter, r *http.Request) {
	if m.healthHandler != nil {
		m.healthHandler(w, r)
		return
	}
	// Return mock health data
	data := map[string]interface{}{
		"bodyBattery": 90,
		"stress":      35,
		"sleep": map[string]interface{}{
			"duration": 480,
			"quality":  85,
		},
	}
	json.NewEncoder(w).Encode(data)
}

// handleAuth is the default authentication handler
func (m *MockServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	if m.authHandler != nil {
		m.authHandler(w, r)
		return
	}

	// Handle session refresh requests
	if strings.Contains(r.URL.Path, "/refresh") {
		// Validate refresh token and return new access token
		response := map[string]interface{}{
			"oauth2_token": "new-mock-token",
			"expires_in":   3600,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Simulate successful authentication
	response := map[string]interface{}{
		"oauth2_token": "mock-access-token",
		"refresh_token": "mock-refresh-token",
		"expires_in":   3600,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// NewClientWithBaseURL creates a test client that uses the mock server's URL
func NewClientWithBaseURL(baseURL string) *Client {
	session := &garth.Session{
		OAuth2Token: "mock-token",
		ExpiresAt:   time.Now().Add(8 * time.Hour),
	}
	client, err := NewClient(session, "")
	if err != nil {
		panic("failed to create test client: " + err.Error())
	}
	client.HTTPClient.SetBaseURL(baseURL)
	return client
}
