package garth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth1LoginFlow(t *testing.T) {
	// Setup mock server to simulate Garmin SSO flow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The request token step uses text/html Accept header
		if r.URL.Path == "/oauth-service/oauth/request_token" {
			assert.Equal(t, "text/html", r.Header.Get("Accept"))
		} else {
			// Other requests use application/json
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
		}
		assert.Equal(t, "garmin-connect-client", r.Header.Get("User-Agent"))

		// Simulate successful SSO response
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<input type="hidden" name="oauth_verifier" value="test_verifier" />`))
	}))
	defer server.Close()

	// Initialize authenticator with test configuration
	auth := NewAuthenticator(server.URL, "")
	auth.MFAPrompter = &MockMFAPrompter{Code: "123456", Err: nil}

	// Test login with mock credentials
	session, err := auth.Login("test_user", "test_pass")
	assert.NoError(t, err, "Login should succeed")
	assert.NotNil(t, session, "Session should be created")
}

func TestMFAFlow(t *testing.T) {
	mfaTriggered := false
	// Setup mock server to simulate MFA requirement
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !mfaTriggered {
			// First response requires MFA
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div class="mfa-required"><input type="hidden" name="mfaContext" value="context123" /></div>`))
			mfaTriggered = true
		} else {
			// Second response after MFA
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<input type="hidden" name="oauth_verifier" value="mfa_verifier" />`))
		}
	}))
	defer server.Close()

	// Initialize authenticator with mock MFA prompter
	auth := NewAuthenticator(server.URL, "")
	auth.MFAPrompter = &MockMFAPrompter{Code: "654321", Err: nil}

	// Test login with MFA
	session, err := auth.Login("mfa_user", "mfa_pass")
	assert.NoError(t, err, "MFA login should succeed")
	assert.NotNil(t, session, "Session should be created")
}

func TestLoginFailure(t *testing.T) {
	// Setup mock server that returns failure responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	auth := NewAuthenticator(server.URL, "")
	auth.MFAPrompter = &MockMFAPrompter{Err: nil}

	session, err := auth.Login("bad_user", "bad_pass")
	assert.Error(t, err, "Should return error for failed login")
	assert.Nil(t, session, "No session should be created on failure")
}

func TestMFAFailure(t *testing.T) {
	mfaTriggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !mfaTriggered {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div class="mfa-required"><input type="hidden" name="mfaContext" value="context123" /></div>`))
			mfaTriggered = true
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer server.Close()

	auth := NewAuthenticator(server.URL, "")
	auth.MFAPrompter = &MockMFAPrompter{Code: "wrong", Err: nil}

	session, err := auth.Login("mfa_user", "mfa_pass")
	assert.Error(t, err, "Should return error for MFA failure")
	assert.Nil(t, session, "No session should be created on MFA failure")
}
