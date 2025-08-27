package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"github.com/dghubble/oauth1"
)

// FileStorage implements TokenStorage using a JSON file
type FileStorage struct {
	Path string
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage() *FileStorage {
	// Default to storing token in user's home directory
	home, _ := os.UserHomeDir()
	return &FileStorage{
		Path: filepath.Join(home, ".garminconnect", "token.json"),
	}
}

// GetToken retrieves token from file
func (s *FileStorage) GetToken() (*oauth1.Token, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}

	var token oauth1.Token
	err = json.Unmarshal(data, &token)
	if err != nil {
		return nil, err
	}

	// Check if token is expired
	if token.Token == "" || token.TokenSecret == "" {
		return nil, os.ErrNotExist
	}

	return &token, nil
}

// SaveToken saves token to file
func (s *FileStorage) SaveToken(token *oauth1.Token) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.Path, data, 0600)
}
