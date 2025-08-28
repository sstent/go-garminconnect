package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MFAState represents the state of an MFA verification session
type MFAState struct {
	VerificationURL string    `json:"verification_url"`
	SessionToken    string    `json:"session_token"`
	MFACode         string    `json:"mfa_code"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// MFAStorage handles persistence of MFA state
type MFAStorage interface {
	Store(state MFAState) error
	Get() (MFAState, error)
	Clear() error
}

// FileMFAStorage implements MFAStorage using a JSON file
type FileMFAStorage struct {
	filePath string
	mutex    sync.RWMutex
}

// NewFileMFAStorage creates a new file-based MFA storage
func NewFileMFAStorage() *FileMFAStorage {
	home, _ := os.UserHomeDir()
	return &FileMFAStorage{
		filePath: filepath.Join(home, ".garminconnect", "mfa_state.json"),
	}
}

// Store saves MFA state to file
func (s *FileMFAStorage) Store(state MFAState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create directory if needed
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600)
}

// Get retrieves MFA state from file
func (s *FileMFAStorage) Get() (MFAState, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return MFAState{}, nil
		}
		return MFAState{}, err
	}

	var state MFAState
	err = json.Unmarshal(data, &state)
	return state, err
}

// Clear removes the MFA state file
func (s *FileMFAStorage) Clear() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return os.Remove(s.filePath)
}
