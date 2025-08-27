package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/stretchr/testify/assert"
)

func TestFileStorage(t *testing.T) {
	// Create temp directory for tests
	tempDir, err := os.MkdirTemp("", "garmin-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage := &FileStorage{
		Path: filepath.Join(tempDir, "token.json"),
	}

	// Test saving and loading token
	t.Run("SaveAndLoadToken", func(t *testing.T) {
		testToken := &oauth1.Token{
			Token:       "access-token",
			TokenSecret: "access-secret",
		}

		// Save token
		err := storage.SaveToken(testToken)
		assert.NoError(t, err)

		// Load token
		loadedToken, err := storage.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, testToken.Token, loadedToken.Token)
		assert.Equal(t, testToken.TokenSecret, loadedToken.TokenSecret)
	})

	// Test missing token file
	t.Run("TokenMissing", func(t *testing.T) {
		_, err := storage.GetToken()
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	// Test token expiration
	t.Run("TokenExpiration", func(t *testing.T) {
		testCases := []struct {
			name     string
			token    *oauth1.Token
			expected bool
		}{
			{
				name:     "EmptyToken",
				token:    &oauth1.Token{},
				expected: true,
			},
			{
				name: "ValidToken",
				token: &oauth1.Token{
					Token:       "valid",
					TokenSecret: "valid",
				},
				expected: false,
			},
			{
				name: "ExpiredToken",
				token: &oauth1.Token{
					Token:       "expired",
					TokenSecret: "expired",
					CreatedAt:   time.Now().Add(-200 * 24 * time.Hour), // 200 days ago
				},
				expected: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				expired := storage.TokenExpired(tc.token)
				assert.Equal(t, tc.expected, expired)
			})
		}
	})
}
