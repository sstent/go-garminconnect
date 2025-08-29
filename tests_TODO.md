# Mock Authenticator Implementation Tasks

## Overview
Implement a shared MockAuthenticator to fix test failures caused by:
- Improper interface implementation in tests
- Duplicate mock implementations across files
- Inconsistent test client creation patterns

## Tasks

### Phase 1: Create Shared Test Helper (test_helpers.go)
```go
package api

type MockAuthenticator struct {
    RefreshTokenFunc func(oauth1Token, oauth1Secret string) (string, error)
    CallCount        int
}

func (m *MockAuthenticator) RefreshToken(oauth1Token, oauth1Secret string) (string, error) {
    m.CallCount++
    if m.RefreshTokenFunc != nil {
        return m.RefreshTokenFunc(oauth1Token, oauth1Secret)
    }
    return "refreshed-test-token", nil
}

func NewMockAuthenticator() *MockAuthenticator {
    return &MockAuthenticator{}
}

func NewMockAuthenticatorWithFunc(refreshFunc func(string, string) (string, error)) *MockAuthenticator {
    return &MockAuthenticator{
        RefreshTokenFunc: refreshFunc,
    }
}
```

### Phase 2: Update Test Files
1. **bodycomposition_test.go**  
   Replace existing mock with:
   ```go
   mockAuth := NewMockAuthenticator()
   ```

2. **gear_test.go**  
   Remove `mockAuthImpl` definition and use:
   ```go
   mockAuth := NewMockAuthenticator()
   ```

3. **health_test.go**  
   Update client creation:
   ```go
   mockAuth := NewMockAuthenticator()
   client, err := NewClient(mockAuth, session, "")
   ```

4. **user_test.go**  
   Update client creation:
   ```go
   mockAuth := NewMockAuthenticator()
   client, err := NewClient(mockAuth, session, "")
   ```

5. **mock_server_test.go**  
   Update `NewClientWithBaseURL`:
   ```go
   auth := NewMockAuthenticator()
   ```

### Phase 3: Verification
- [ ] Run tests: `go test ./internal/api/...`
- [ ] Fix any remaining compilation errors
- [ ] Verify all tests pass
- [ ] Check for consistent mock usage across all test files

## Progress Tracking
- [x] test_helpers.go created
- [x] bodycomposition_test.go updated
- [x] gear_test.go updated
- [x] health_test.go updated
- [x] user_test.go updated
- [x] mock_server_test.go updated
- [x] integration_test.go updated
- [x] Run tests: `go test ./internal/api/...`
- [x] Verify all tests pass
