package garth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth1LoginFlow(t *testing.T) {
	// Setup mock server to simulate complete Garmin SSO flow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth-service/oauth/request_token":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("oauth_token=test_token&oauth_token_secret=test_secret"))

		case "/sso/signin":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<input type="hidden" name="oauth_verifier" value="test_verifier" />`))

		case "/oauth-service/oauth/access_token":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("oauth_token=access_token&oauth_token_secret=access_secret"))

		case "/oauth-service/oauth/exchange/user/2.0":
			w.Write([]byte("oauth2_token"))

		default:
			t.Errorf("Unexpected request to path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	// Initialize authenticator with test configuration
	auth := NewAuthenticator(server.URL, "")
	auth.MFAPrompter = &MockMFAPrompter{Code: "123456", Err: nil}

	// Test login with mock credentials
	session, err := auth.Login("test_user", "test_pass")
	assert.NoError(t, err, "Login should succeed")
	assert.NotNil(t, session, "Session should be created")

	// Verify session values
	assert.Equal(t, "access_token", session.OAuth1Token)
	assert.Equal(t, "access_secret", session.OAuth1Secret)
	assert.Equal(t, "oauth2_token", session.OAuth2Token)
	assert.False(t, session.IsExpired(), "Session should not be expired")
}

func TestMFAFlow(t *testing.T) {
	mfaTriggered := false
	// Setup mock server to simulate MFA requirement and complete flow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth-service/oauth/request_token":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("oauth_token=test_token&oauth_token_secret=test_secret"))

		case r.URL.Path == "/sso/signin" && !mfaTriggered:
			// First response requires MFA
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div class="mfa-required"><input type="hidden" name="mfaContext" value="context123" /></div>`))
			mfaTriggered = true

		case r.URL.Path == "/sso/verifyMFA":
			// MFA verification
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<input type="hidden" name="oauth_verifier" value="mfa_verifier" />`))

		case r.URL.Path == "/oauth-service/oauth/access_token":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("oauth_token=access_token&oauth_token_secret=access_secret"))

		case r.URL.Path == "/oauth-service/oauth/exchange/user/2.0":
			w.Write([]byte("oauth2_token"))

		default:
			t.Errorf("Unexpected request to path: %s", r.URL.Path)
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

	// Verify session values
	assert.Equal(t, "access_token", session.OAuth1Token)
	assert.Equal(t, "access_secret", session.OAuth1Secret)
	assert.Equal(t, "oauth2_token", session.OAuth2Token)
	assert.False(t, session.IsExpired(), "Session should not be expired")
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
