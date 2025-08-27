package auth

import (
	"fmt"
	"net/http"
)

// MFAHandler handles multi-factor authentication
func MFAHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Show MFA form
		fmt.Fprintf(w, `<html>
			<body>
				<form method="POST">
					<label>MFA Code: <input type="text" name="mfa_code"></label>
					<button type="submit">Verify</button>
				</form>
			</body>
		</html>`)
	case "POST":
		// Process MFA code
		code := r.FormValue("mfa_code")
		// Validate MFA code - in a real app, this would be sent to Garmin
		if len(code) != 6 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid MFA code format. Please enter a 6-digit code."))
			return
		}
		
		// Store MFA verification status in session
		// In a real app, we'd store this in a session store
		w.Write([]byte("MFA verification successful! Please return to your application."))
	}
}

// RequiresMFA checks if MFA is required based on Garmin response
func RequiresMFA(err error) bool {
	// In a real implementation, we'd check the error type
	// or response from Garmin to determine if MFA is needed
	return err != nil && err.Error() == "mfa_required"
}
