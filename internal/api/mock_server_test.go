package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// MockServer simulates the Garmin Connect API
type MockServer struct {
	server *httptest.Server
	mu     sync.Mutex

	// Endpoint handlers
	activitiesHandler   http.HandlerFunc
	activityDetailsHandler http.HandlerFunc
	uploadHandler       http.HandlerFunc
	userHandler         http.HandlerFunc
	healthHandler       http.HandlerFunc
	authHandler         http.HandlerFunc
}

// NewMockServer creates a new mock Garmin Connect server
func NewMockServer() *MockServer {
	m := &MockServer{}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		switch {
		case strings.HasPrefix(r.URL.Path, "/activity-service/activities"):
			m.handleActivities(w, r)
		case strings.HasPrefix(r.URL.Path, "/activity-service/activity/"):
			m.handleActivityDetails(w, r)
		case strings.HasPrefix(r.URL.Path, "/upload-service/upload"):
			m.handleUpload(w, r)
		case strings.HasPrefix(r.URL.Path, "/user-service/user"):
			m.handleUserData(w, r)
		case strings.HasPrefix(r.URL.Path, "/health-service"):
			m.handleHealthData(w, r)
		case strings.HasPrefix(r.URL.Path, "/auth"):
			m.handleAuth(w, r)
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
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
	// Simulate successful authentication
	response := map[string]interface{}{
		"token": "mock-token-123",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
