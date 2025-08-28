package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Enable debug logging based on environment variable
func debugEnabled() bool {
	return os.Getenv("DEBUG_AUTH") == "true"
}

// debugLog prints debug messages if debugging is enabled
func debugLog(format string, args ...interface{}) {
	if debugEnabled() {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// fetchLoginParams retrieves required tokens from Garmin login page
func (c *AuthClient) fetchLoginParams(ctx context.Context) (lt, execution string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://sso.garmin.com/sso/signin?service=https://connect.garmin.com", nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create login page request: %w", err)
	}

	req.Header = getBrowserHeaders()

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("login page request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read login page response: %w", err)
	}

	// For debugging: Log response status and headers
	debugLog("Login page response status: %s", resp.Status)
	debugLog("Login page response headers: %v", resp.Header)

	// Write body to debug log if it's not too large
	if len(body) < 5000 {
		debugLog("Login page body: %s", body)
	} else {
		debugLog("Login page body too large to log (%d bytes)", len(body))
	}

	lt, err = extractParam(`name="lt"\s+value="([^"]+)"`, string(body))
	if err != nil {
		return "", "", fmt.Errorf("lt param not found: %w", err)
	}

	execution, err = extractParam(`name="execution"\s+value="([^"]+)"`, string(body))
	if err != nil {
		return "", "", fmt.Errorf("execution param not found: %w", err)
	}

	return lt, execution, nil
}

// extractParam helper to extract regex pattern
func extractParam(pattern, body string) (string, error) {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("pattern not found")
	}
	return matches[1], nil
}

// getBrowserHeaders returns browser-like headers for requests
func getBrowserHeaders() http.Header {
	return http.Header{
		"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Connection":                {"keep-alive"},
		"Cache-Control":             {"max-age=0"},
		"Sec-Fetch-Site":            {"none"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Fetch-Dest":            {"document"},
		"DNT":                       {"1"},
		"Upgrade-Insecure-Requests": {"1"},
	}
}

// Authenticate handles Garmin Connect authentication with MFA support
func (c *AuthClient) Authenticate(ctx context.Context, username, password, mfaToken string) (*Token, error) {
	// Fetch required tokens from login page
	lt, execution, err := c.fetchLoginParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get login params: %w", err)
	}

	// Create login form data with required parameters
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("embed", "true")
	data.Set("rememberme", "on")
	data.Set("lt", lt)
	data.Set("execution", execution)
	data.Set("_eventId", "submit")
	data.Set("geolocation", "")
	data.Set("clientId", "GarminConnect")
	data.Set("service", "https://connect.garmin.com")
	data.Set("webhost", "https://connect.garmin.com")
	data.Set("fromPage", "oauth")
	data.Set("locale", "en_US")
	data.Set("id", "gauth-widget")
	data.Set("redirectAfterAccountLoginUrl", "https://connect.garmin.com/oauthConfirm")
	data.Set("redirectAfterAccountCreationUrl", "https://connect.garmin.com/oauthConfirm")

	// Create login request
	loginURL := "https://sso.garmin.com/sso/signin"
	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create SSO request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	// Key change: Request JSON response instead of HTML
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Log request details if debugging
	debugLog("Sending SSO request to: %s", loginURL)
	debugLog("Request headers: %v", req.Header)

	// Send SSO request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SSO request failed: %w", err)
	}
	defer resp.Body.Close()

	// Log response details
	debugLog("SSO response status: %s", resp.Status)
	debugLog("Response headers: %v", resp.Header)

	// Check for MFA requirement
	if resp.StatusCode == http.StatusPreconditionFailed {
		if mfaToken == "" {
			return nil, errors.New("MFA required but no token provided")
		}
		return c.handleMFA(ctx, username, password, mfaToken, "")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	// Parse JSON response to get ticket
	var authResponse struct {
		Ticket string `json:"ticket"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("failed to parse SSO response: %w", err)
	}

	if authResponse.Ticket == "" {
		return nil, errors.New("empty ticket in SSO response")
	}

	// Exchange ticket for tokens
	return c.exchangeTicketForTokens(ctx, authResponse.Ticket)
}

// extractSSOTicket finds the authentication ticket in the SSO response
func extractSSOTicket(body string) (string, error) {
	// The ticket is typically in a hidden input field
	ticketPattern := `name="ticket"\s+value="([^"]+)"`
	re := regexp.MustCompile(ticketPattern)
	matches := re.FindStringSubmatch(body)

	if len(matches) < 2 {
		if strings.Contains(body, "Cloudflare") {
			return "", errors.New("Cloudflare bot protection triggered")
		}
		return "", errors.New("ticket not found in SSO response")
	}
	return matches[1], nil
}

// handleMFA processes multi-factor authentication
func (c *AuthClient) handleMFA(ctx context.Context, username, password, mfaToken, responseBody string) (*Token, error) {
	// Extract required parameters from the initial response
	params, err := extractMFAParams(responseBody)
	if err != nil {
		return nil, err
	}

	// Prepare MFA request
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("embed", "false")
	data.Set("rememberme", "on")
	data.Set("_eventId", "submit")
	data.Set("mfaCode", mfaToken)

	// Add all parameters from the initial response
	for key, value := range params {
		data.Set(key, value)
	}

	// Create MFA request
	loginURL := "https://sso.garmin.com/sso/signin"
	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create MFA request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; go-garminconnect/1.0)")

	// Send MFA request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MFA request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MFA response: %w", err)
	}

	// Extract ticket from MFA response
	ticket, err := extractSSOTicket(string(body))
	if err != nil {
		return nil, fmt.Errorf("ticket not found in MFA response: %w", err)
	}

	// Exchange ticket for tokens
	return c.exchangeTicketForTokens(ctx, ticket)
}

// extractSessionCookie extracts session cookie from headers
func extractSessionCookie(cookieHeader string) string {
	sessionPattern := `SESSION=([^;]+)`
	re := regexp.MustCompile(sessionPattern)
	matches := re.FindStringSubmatch(cookieHeader)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// extractMFAParams extracts necessary parameters for MFA request
func extractMFAParams(body string) (map[string]string, error) {
	params := make(map[string]string)
	patterns := []string{
		`name="lt"\s+value="([^"]+)"`,
		`name="execution"\s+value="([^"]+)"`,
		`name="_eventId"\s+value="([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(body)
		if len(matches) < 2 {
			return nil, fmt.Errorf("required parameter not found: %s", pattern)
		}
		paramName := re.SubexpNames()[1]
		params[paramName] = matches[1]
	}

	return params, nil
}

// exchangeTicketForTokens exchanges an SSO ticket for access tokens
func (c *AuthClient) exchangeTicketForTokens(ctx context.Context, ticket string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", ticket)
	data.Set("redirect_uri", "https://connect.garmin.com")

	req, err := http.NewRequestWithContext(ctx, "POST", c.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; go-garminconnect/1.0)")

	// Add basic authentication
	req.SetBasicAuth("garmin-connect", "garmin-connect-secret")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %d %s", resp.StatusCode, body)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}
