package api

// MockAuthenticator implements the Authenticator interface for testing
type MockAuthenticator struct {
	// RefreshTokenFunc can be set for custom refresh behavior
	RefreshTokenFunc func(oauth1Token, oauth1Secret string) (string, error)

	// CallCount tracks how many times RefreshToken was called
	CallCount int
}

// RefreshToken implements the Authenticator interface
func (m *MockAuthenticator) RefreshToken(oauth1Token, oauth1Secret string) (string, error) {
	m.CallCount++

	// If custom function is provided, use it
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(oauth1Token, oauth1Secret)
	}

	// Default behavior: return a mock token
	return "refreshed-test-token", nil
}

// NewMockAuthenticator creates a new mock authenticator with default behavior
func NewMockAuthenticator() *MockAuthenticator {
	return &MockAuthenticator{}
}

// NewMockAuthenticatorWithFunc creates a mock authenticator with custom refresh behavior
func NewMockAuthenticatorWithFunc(refreshFunc func(string, string) (string, error)) *MockAuthenticator {
	return &MockAuthenticator{
		RefreshTokenFunc: refreshFunc,
	}
}
