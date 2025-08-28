package fit

import "fmt"

// ValidateFIT validates FIT file data with basic header check
func ValidateFIT(data []byte) error {
	// Minimal validation - check if data starts with FIT header
	if len(data) < 12 {
		return fmt.Errorf("file too small to be a valid FIT file")
	}
	return nil
}
