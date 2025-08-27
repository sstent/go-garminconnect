package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTokenRefresh tests the token refresh functionality
func TestTokenRefresh(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  interface{}
		mockStatus    int
		expectedToken *Token
		expectedError string
	}{
		{
			name: "successful token refresh",
			mockResponse: map[string]interface{}{
				"access_token":  "new-access-token",
				"refresh_token": "new-refresh-token",
				"expires_in":    3600,
				"token_type":    "Bearer",
			},
			mockStatus: http.StatusOK,
			expectedToken: &Token{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				ExpiresIn:    3600,
				TokenType:    "Bearer",
				Expiry:       time.Now().Add(3600 * time.Second),
			},
		},
		{
			name: "expired refresh token",
			mockResponse: map[string]interface{}{
				"error": "invalid_grant",
				"error_description": "Refresh token expired",
			},
			mockStatus:    http.StatusBadRequest,
			expectedError: "token refresh failed with status 400",
		},
		{
			name: "invalid token response",
			mockResponse: map[string]interface{}{
				"invalid": "data",
			},
			mockStatus:    http.StatusOK,
			expectedError: "token response missing required fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create auth client
			client := &AuthClient{
				Client:   &http.Client{},
				TokenURL: server.URL,
			}

			// Create token to refresh
			token := &Token{
				RefreshToken: "old-refresh-token",
			}

			// Execute test
			newToken, err := client.RefreshToken(context.Background(), token)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, newToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newToken)
				assert.Equal(t, tt.expectedToken.AccessToken, newToken.AccessToken)
				assert.Equal(t, tt.expectedToken.RefreshToken, newToken.RefreshToken)
				assert.Equal(t, tt.expectedToken.ExpiresIn, newToken.ExpiresIn)
				assert.WithinDuration(t, tt.expectedToken.Expiry, newToken.Expiry, 5*time.Second)
			}
		})
	}
}

// TestMFAAuthentication tests MFA authentication flow
func TestMFAAuthentication(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		mfaToken      string
		mockResponses []mockResponse // Multiple responses for MFA flow
		expectedToken *Token
		expectedError string
	}{
		{
			name:     "successful MFA authentication",
			username: "user@example.com",
			password: "password123",
			mfaToken: "123456",
			mockResponses: []mockResponse{
				{
					status: http.StatusUnauthorized,
					body: map[string]interface{}{
						"mfaToken": "mfa-challenge-token",
					},
				},
				{
					status: http.StatusOK,
					body: map[string]interface{}{},
					cookies: map[string]string{
						"access_token":  "access-token",
						"refresh_token": "refresh-token",
					},
				},
			},
			expectedToken: &Token{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    3600,
				TokenType:    "Bearer",
				Expiry:       time.Now().Add(3600 * time.Second),
			},
		},
		{
			name:     "invalid MFA code",
			username: "user@example.com",
			password: "password123",
			mfaToken: "wrong-code",
			mockResponses: []mockResponse{
				{
					status: http.StatusUnauthorized,
					body: map[string]interface{}{
						"mfaToken": "mfa-challenge-token",
					},
				},
				{
					status: http.StatusUnauthorized,
					body: map[string]interface{}{
						"error": "Invalid MFA token",
					},
				},
			},
			expectedError: "authentication failed: 401",
		},
		{
			name:     "MFA required but not provided",
			username: "user@example.com",
			password: "password123",
			mfaToken: "",
			mockResponses: []mockResponse{
				{
					status: http.StatusUnauthorized,
					body: map[string]interface{}{
						"mfaToken": "mfa-challenge-token",
					},
				},
			},
			expectedError: "MFA required but no token provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server with state
			currentResponse := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if currentResponse < len(tt.mockResponses) {
					response := tt.mockResponses[currentResponse]
					w.Header().Set("Content-Type", "application/json")
					// Set additional headers if specified
					for key, value := range response.headers {
						w.Header().Set(key, value)
					}
					// Set cookies if specified
					for name, value := range response.cookies {
						http.SetCookie(w, &http.Cookie{
							Name:  name,
							Value: value,
						})
					}
					w.WriteHeader(response.status)
					json.NewEncoder(w).Encode(response.body)
					currentResponse++
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			// Create auth client
			client := &AuthClient{
				Client:    &http.Client{},
				BaseURL:   server.URL,
				TokenURL:  fmt.Sprintf("%s/oauth/token", server.URL),
				LoginPath: "/sso/login",
			}

			// Execute test
			token, err := client.Authenticate(context.Background(), tt.username, tt.password, tt.mfaToken)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, tt.expectedToken.AccessToken, token.AccessToken)
				assert.Equal(t, tt.expectedToken.RefreshToken, token.RefreshToken)
				assert.Equal(t, tt.expectedToken.ExpiresIn, token.ExpiresIn)
				assert.WithinDuration(t, tt.expectedToken.Expiry, token.Expiry, 5*time.Second)
			}
		})
	}
}

// BenchmarkTokenRefresh measures the performance of token refresh
func BenchmarkTokenRefresh(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "benchmark-access-token",
			"refresh_token": "benchmark-refresh-token",
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	}))
	defer server.Close()

	// Create auth client
	client := &AuthClient{
		Client:   &http.Client{},
		TokenURL: server.URL,
	}

	// Create token to refresh
	token := &Token{
		RefreshToken: "benchmark-refresh-token",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.RefreshToken(context.Background(), token)
	}
}

// BenchmarkMFAAuthentication measures the performance of MFA authentication
func BenchmarkMFAAuthentication(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/sso/login" {
			// First request returns MFA challenge
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"mfaToken": "mfa-challenge-token",
			})
		} else if r.URL.Path == "/oauth/token" {
			// Second request returns tokens
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "benchmark-access-token",
				"refresh_token": "benchmark-refresh-token",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})
		}
	}))
	defer server.Close()

	// Create auth client
	client := &AuthClient{
		Client:    &http.Client{},
		BaseURL:   server.URL,
		TokenURL:  fmt.Sprintf("%s/oauth/token", server.URL),
		LoginPath: "/sso/login",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Authenticate(context.Background(), "benchmark@example.com", "benchmark-password", "123456")
	}
}

type mockResponse struct {
	status  int
	body    interface{}
	headers map[string]string
	cookies map[string]string
}
