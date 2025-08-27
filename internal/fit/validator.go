package fit

// Validate performs basic validation of FIT file structure
func Validate(data []byte) bool {
	// Minimum FIT file size is 14 bytes (header)
	if len(data) < 14 {
		return false
	}
	
	// Check magic number: ".FIT"
	if string(data[8:12]) != ".FIT" {
		return false
	}
	
	return true
}

// MinFileSize returns the minimum size of a valid FIT file
func MinFileSize() int {
	return 14
}
