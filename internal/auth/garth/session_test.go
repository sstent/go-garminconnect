package garth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionPersistence(t *testing.T) {
	// Setup temporary file
	tmpDir := os.TempDir()
	sessionFile := filepath.Join(tmpDir, "test_session.json")
	defer os.Remove(sessionFile)

	// Create test session
	testSession := &Session{
		OAuth1Token:  "test_oauth1_token",
		OAuth1Secret: "test_oauth1_secret",
		OAuth2Token:  "test_oauth2_token",
	}

	// Test saving
	err := testSession.Save(sessionFile)
	assert.NoError(t, err, "Saving session should not produce error")

	// Test loading
	loadedSession, err := LoadSession(sessionFile)
	assert.NoError(t, err, "Loading session should not produce error")
	assert.Equal(t, testSession, loadedSession, "Loaded session should match saved session")

	// Test loading non-existent file
	_, err = LoadSession("non_existent_file.json")
	assert.Error(t, err, "Loading non-existent file should return error")
}

func TestSessionContextHandling(t *testing.T) {
	// Create authenticator with session path
	tmpDir := os.TempDir()
	sessionFile := filepath.Join(tmpDir, "context_session.json")
	defer os.Remove(sessionFile)

	auth := NewAuthenticator("https://example.com", sessionFile)

	// Verify empty session returns error
	_, err := auth.Login("user", "pass")
	assert.Error(t, err, "Should return error when no active session")
}

func TestMFAPrompterInterface(t *testing.T) {
	// Test console prompter implements interface
	var prompter MFAPrompter = DefaultConsolePrompter{}
	_, err := prompter.GetMFACode(context.Background())
	assert.NoError(t, err, "Default prompter should not produce errors")

	// Test mock prompter
	mock := &MockMFAPrompter{Code: "123456", Err: nil}
	code, err := mock.GetMFACode(context.Background())
	assert.Equal(t, "123456", code, "Mock prompter should return provided code")
	assert.NoError(t, err, "Mock prompter should not return error when Err is nil")

	// Test error case
	errorMock := &MockMFAPrompter{Err: errors.New("prompt error")}
	_, err = errorMock.GetMFACode(context.Background())
	assert.Error(t, err, "Mock prompter should return error when set")
}
