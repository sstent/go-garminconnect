
# Garth Go Port Plan - Test-Driven Development Implementation

## Project Overview

Port the Python Garth library (Garmin SSO auth + Connect API client) to Go with comprehensive test coverage and modern Go practices.

## Core Architecture Analysis

Based on the original Garth library, the main components are:
- **Authentication**: OAuth1/OAuth2 token management with auto-refresh
- **API Client**: HTTP client for Garmin Connect API requests
- **Data Models**: Structured data types for health/fitness metrics
- **Session Management**: Token persistence and restoration

## 1. Project Structure

```
garth-go/
├── cmd/
│   └── garth/                    # CLI tool (like Python's uvx garth login)
│       └── main.go
├── pkg/
│   ├── auth/                     # Authentication module
│   │   ├── oauth.go
│   │   ├── oauth_test.go
│   │   ├── session.go
│   │   └── session_test.go
│   ├── client/                   # HTTP client module
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── endpoints.go
│   │   └── endpoints_test.go
│   ├── models/                   # Data structures
│   │   ├── sleep.go
│   │   ├── sleep_test.go
│   │   ├── stress.go
│   │   ├── stress_test.go
│   │   ├── steps.go
│   │   ├── weight.go
│   │   ├── hrv.go
│   │   └── user.go
│   └── garth/                    # Main package interface
│       ├── garth.go
│       └── garth_test.go
├── internal/
│   ├── testutil/                 # Test utilities
│   │   ├── fixtures.go
│   │   └── mock_server.go
│   └── config/                   # Internal configuration
│       └── constants.go
├── examples/                     # Usage examples
│   ├── basic/
│   ├── sleep_analysis/
│   └── stress_tracking/
├── go.mod
├── go.sum
├── README.md
├── Makefile
└── .github/
    └── workflows/
        └── ci.yml
```

## 2. Data Flow Architecture

### Authentication Flow
```
User Credentials → OAuth1 Token → OAuth2 Token → API Requests
                      ↓              ↓
                   Persisted      Auto-refresh
```

### API Request Flow
```
Client Request → Token Validation → HTTP Request → JSON Response → Struct Unmarshaling
                      ↓
                 Auto-refresh if expired
```

### Data Processing Flow
```
Raw API Response → JSON Unmarshaling → Data Validation → Business Logic → Client Response
```

## 3. Recommended Go Modules

### Core Dependencies
```go
// HTTP client and utilities
"net/http"
"context"
"time"

// JSON handling
"encoding/json"

// OAuth implementation
"golang.org/x/oauth2" // For OAuth2 flows

// HTTP client with advanced features
"github.com/go-resty/resty/v2" // Alternative to net/http with better ergonomics

// Configuration and environment
"github.com/spf13/viper" // Configuration management
"github.com/spf13/cobra" // CLI framework

// Validation
"github.com/go-playground/validator/v10" // Struct validation

// Logging
"go.uber.org/zap" // Structured logging

// Testing
"github.com/stretchr/testify" // Testing utilities
"github.com/jarcoal/httpmock" // HTTP mocking
```

### Development Dependencies
```go
// Code generation
"github.com/golang/mock/gomock" // Mock generation

// Linting and quality
"github.com/golangci/golangci-lint"
```

## 4. TDD Implementation Plan

### Phase 1: Authentication Module (Week 1-2)

#### Test Cases to Implement First:

**OAuth Session Tests:**
```go
func TestSessionSave(t *testing.T)
func TestSessionLoad(t *testing.T)
func TestSessionValidation(t *testing.T)
func TestSessionExpiry(t *testing.T)
```

**OAuth Flow Tests:**
```go
func TestOAuth1Login(t *testing.T)
func TestOAuth2TokenRefresh(t *testing.T)
func TestMFAHandling(t *testing.T)
func TestLoginFailure(t *testing.T)
```

#### Implementation Order:
1. **Write failing tests** for session management
2. **Implement** basic session struct and methods
3. **Write failing tests** for OAuth1 authentication
4. **Implement** OAuth1 flow
5. **Write failing tests** for OAuth2 token refresh
6. **Implement** OAuth2 auto-refresh mechanism
7. **Write failing tests** for MFA handling
8. **Implement** MFA prompt system

### Phase 2: HTTP Client Module (Week 3)

#### Test Cases:
```go
func TestClientCreation(t *testing.T)
func TestAPIRequest(t *testing.T)
func TestAuthenticationHeaders(t *testing.T)
func TestErrorHandling(t *testing.T)
func TestRetryLogic(t *testing.T)
```

#### Mock Server Setup:
```go
// Create mock Garmin Connect API responses
func setupMockGarminServer() *httptest.Server
func mockSuccessResponse() string
func mockErrorResponse() string
```

### Phase 3: Data Models (Week 4-5)

#### Core Models Implementation Order:

**1. User Profile:**
```go
type UserProfile struct {
    ID           int    `json:"id" validate:"required"`
    ProfileID    int    `json:"profileId" validate:"required"`
    DisplayName  string `json:"displayName"`
    FullName     string `json:"fullName"`
    // ... other fields
}
```

**2. Sleep Data:**
```go
type SleepData struct {
    CalendarDate    time.Time `json:"calendarDate"`
    SleepTimeSeconds int      `json:"sleepTimeSeconds"`
    DeepSleep       int      `json:"deepSleepSeconds"`
    LightSleep      int      `json:"lightSleepSeconds"`
    // ... other fields
}
```

**3. Stress Data:**
```go
type DailyStress struct {
    CalendarDate         time.Time `json:"calendarDate"`
    OverallStressLevel   int       `json:"overallStressLevel"`
    RestStressDuration   int       `json:"restStressDuration"`
    // ... other fields
}
```

#### Test Implementation Strategy:
1. **JSON Unmarshaling Tests** - Test API response parsing
2. **Validation Tests** - Test struct validation
3. **Business Logic Tests** - Test derived properties and methods

### Phase 4: Main Interface (Week 6)

#### High-level API Tests:
```go
func TestGarthLogin(t *testing.T)
func TestGarthConnectAPI(t *testing.T)
func TestGarthSave(t *testing.T)
func TestGarthResume(t *testing.T)
```

#### Integration Tests:
```go
func TestEndToEndSleepDataRetrieval(t *testing.T)
func TestEndToEndStressDataRetrieval(t *testing.T)
```

## 5. TDD Development Workflow

### Red-Green-Refactor Cycle:

#### For Each Feature:
1. **RED**: Write failing test that describes desired behavior
2. **GREEN**: Write minimal code to make test pass
3. **REFACTOR**: Clean up code while keeping tests green
4. **REPEAT**: Add more test cases and iterate

#### Example TDD Session - Session Management:

**Step 1 - RED**: Write failing test
```go
func TestSessionSave(t *testing.T) {
    session := &Session{
        OAuth1Token: "token1",
        OAuth2Token: "token2",
    }
    
    err := session.Save("/tmp/test_session")
    require.NoError(t, err)
    
    // Should create file
    _, err = os.Stat("/tmp/test_session")
    assert.NoError(t, err)
}
```

**Step 2 - GREEN**: Make test pass
```go
type Session struct {
    OAuth1Token string `json:"oauth1_token"`
    OAuth2Token string `json:"oauth2_token"`
}

func (s *Session) Save(path string) error {
    data, err := json.Marshal(s)
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}
```

**Step 3 - REFACTOR**: Improve implementation
```go
func (s *Session) Save(path string) error {
    // Add validation
    if s.OAuth1Token == "" {
        return errors.New("oauth1 token required")
    }
    
    data, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal session: %w", err)
    }
    
    return os.WriteFile(path, data, 0600) // More secure permissions
}
```

### Testing Strategy:

#### Unit Tests (80% coverage target):
- All public methods tested
- Error conditions covered
- Edge cases handled

#### Integration Tests:
- Full authentication flow
- API request/response cycles
- File I/O operations

#### Mock Usage:
```go
type MockHTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type MockGarminAPI struct {
    responses map[string]*http.Response
}

func (m *MockGarminAPI) Do(req *http.Request) (*http.Response, error) {
    response, exists := m.responses[req.URL.Path]
    if !exists {
        return nil, errors.New("unexpected request")
    }
    return response, nil
}
```

## 6. Implementation Timeline

### Week 1: Project Setup + Authentication Tests
- [x] Initialize Go module and project structure
- [x] Write authentication test cases
- [ ] Set up CI/CD pipeline
- [x] Implement basic session management

### Week 2: Complete Authentication Module
- [ ] Implement OAuth1 flow (in progress)
- [ ] Implement OAuth2 token refresh
- [x] Add MFA support (core implementation)
- [ ] Comprehensive authentication testing

### Week 3: HTTP Client Module
- [ ] Write HTTP client tests
- [ ] Implement client with retry logic
- [ ] Add request/response logging
- [ ] Mock server for testing

### Week 4: Data Models - Core Types
- [ ] User profile models
- [ ] Sleep data models
- [ ] JSON marshaling/unmarshaling tests

### Week 5: Data Models - Health Metrics
- [x] Stress data models (implemented)
- [x] Steps, HRV, weight models (implemented)
- [x] Validation and business logic

### Week 6: Main Interface + Integration
- [ ] High-level API implementation
- [ ] Integration tests
- [ ] Documentation and examples
- [ ] Performance optimization

### Week 7: CLI Tool + Polish
- [ ] Command-line interface
- [ ] Error handling improvements
- [ ] Final testing and bug fixes

## 7. Quality Gates

### Before Each Phase Completion:
- [ ] All tests passing
- [ ] Code coverage > 80%
- [ ] Linting passes
- [ ] Documentation updated

### Before Release:
- [ ] Integration tests with real Garmin API (optional)
- [ ] Performance benchmarks
- [ ] Security review
- [ ] Cross-platform testing

## 8. Success Metrics

### Functional Requirements:
- [ ] Authentication flow matches Python library (in progress)
- [x] All data models supported
- [ ] API requests work identically
- [x] Session persistence compatible

### Quality Requirements:
- [ ] >90% test coverage
- [ ] Zero critical security issues
- [ ] Memory usage < 50MB for typical operations
- [ ] API response time < 2s for standard requests

### Developer Experience:
- [ ] Clear documentation with examples
- [ ] Easy installation (`go install`)
- [ ] Intuitive API design
- [ ] Comprehensive error messages

This TDD approach ensures that the Go port will be robust, well-tested, and maintain feature parity with the original Python library while leveraging Go's strengths in performance and concurrency.
