package garth

import (
	"context"
)

// MockMFAPrompter is a mock implementation of MFAPrompter for testing
type MockMFAPrompter struct {
	Code string
	Err  error
}

func (m *MockMFAPrompter) GetMFACode(ctx context.Context) (string, error) {
	return m.Code, m.Err
}
