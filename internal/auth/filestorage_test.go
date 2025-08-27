package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dghubble/oauth1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storage := NewFileStorage()
	storage.Path = filepath.Join(tempDir, "token.json")

	t.Run("SaveToken and GetToken", func(t *testing.T) {
		token := &oauth1.Token{
			Token:       "test_token",
			TokenSecret: "test_secret",
		}

		// Save token
		err := storage.SaveToken(token)
		require.NoError(t, err)

		// Get token
		retrievedToken, err := storage.GetToken()
		require.NoError(t, err)

		// Verify
		assert.Equal(t, token.Token, retrievedToken.Token)
		assert.Equal(t, token.TokenSecret, retrievedToken.TokenSecret)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		token := &oauth1.Token{
			Token:       "",
			TokenSecret: "",
		}

		err := storage.SaveToken(token)
		require.NoError(t, err)

		_, err = storage.GetToken()
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		storage.Path = filepath.Join(tempDir, "nonexistent.json")
		_, err := storage.GetToken()
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}
