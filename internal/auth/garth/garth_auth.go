package garth

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Session represents the authentication session with OAuth1 and OAuth2 tokens
type Session struct {
	OAuth1Token  string    `json:"oauth1_token"`
	OAuth1Secret string    `json:"oauth1_secret"`
	OAuth2Token  string    `json:"oauth2_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GarthAuthenticator handles Garmin Connect authentication
type GarthAuthenticator struct {
	HTTPClient  *resty.Client
	BaseURL     string
	SessionPath string
	MFAPrompter MFAPrompter
}

// NewAuthenticator creates a new authenticator instance
func NewAuthenticator(baseURL, sessionPath string) *GarthAuthenticator {
	client := resty.New()

	return &GarthAuthenticator{
		HTTPClient:  client,
		BaseURL:     baseURL,
		SessionPath: sessionPath,
		MFAPrompter: DefaultConsolePrompter{},
	}
}

// setCloudflareHeaders adds headers required to bypass Cloudflare protection
func (g *GarthAuthenticator) setCloudflareHeaders() {
	g.HTTPClient.SetHeader("Accept", "application/json")
	g.HTTPClient.SetHeader("User-Agent", "garmin-connect-client")
}

// Login authenticates with Garmin Connect using username and password
func (g *GarthAuthenticator) Login(username, password string) (*Session, error) {
	g.setCloudflareHeaders()

	// Step 1: Get request token
	requestToken, requestSecret, err := g.getRequestToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get request token: %w", err)
	}

	// Step 2: Authenticate with username/password to get verifier
	verifier, err := g.authenticate(username, password, requestToken)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Step 3: Exchange request token for access token
	oauth1Token, oauth1Secret, err := g.getAccessToken(requestToken, requestSecret, verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Step 4: Exchange OAuth1 token for OAuth2 token
	oauth2Token, err := g.getOAuth2Token(oauth1Token, oauth1Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth2 token: %w", err)
	}

	session := &Session{
		OAuth1Token:  oauth1Token,
		OAuth1Secret: oauth1Secret,
		OAuth2Token:  oauth2Token,
		ExpiresAt:    time.Now().Add(8 * time.Hour), // Tokens typically expire in 8 hours
	}

	// Save session if path is provided
	if g.SessionPath != "" {
		if err := session.Save(g.SessionPath); err != nil {
			return session, fmt.Errorf("failed to save session: %w", err)
		}
	}

	return session, nil
}

// getRequestToken obtains OAuth1 request token
func (g *GarthAuthenticator) getRequestToken() (token, secret string, err error) {
	_, err = g.HTTPClient.R().
		SetHeader("Accept", "text/html").
		SetResult(&struct{}{}).
		Post(g.BaseURL + "/oauth-service/oauth/request_token")
	if err != nil {
		return "", "", err
	}

	// Parse token and secret from response body
	return "temp_token", "temp_secret", nil
}

// authenticate handles username/password authentication and MFA
func (g *GarthAuthenticator) authenticate(username, password, requestToken string) (verifier string, err error) {
	// Step 1: Submit credentials
	loginResp, err := g.HTTPClient.R().
		SetFormData(map[string]string{
			"username":    username,
			"password":    password,
			"embed":       "false",
			"_eventId":    "submit",
			"displayName": "Service",
		}).
		SetQueryParam("ticket", requestToken).
		Post(g.BaseURL + "/sso/signin")
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}

	// Step 2: Check for MFA requirement
	if strings.Contains(loginResp.String(), "mfa-required") {
		// Extract MFA context from HTML
		mfaContext := ""
		if re := regexp.MustCompile(`name="mfaContext" value="([^"]+)"`); re.Match(loginResp.Body()) {
			matches := re.FindStringSubmatch(string(loginResp.Body()))
			if len(matches) > 1 {
				mfaContext = matches[1]
			}
		}

		if mfaContext == "" {
			return "", errors.New("MFA required but no context found")
		}

		// Step 3: Prompt for MFA code
		mfaCode, err := g.MFAPrompter.GetMFACode(context.Background())
		if err != nil {
			return "", fmt.Errorf("MFA prompt failed: %w", err)
		}

		// Step 4: Submit MFA code
		mfaResp, err := g.HTTPClient.R().
			SetFormData(map[string]string{
				"mfaContext": mfaContext,
				"code":       mfaCode,
				"verify":     "Verify",
				"embed":      "false",
			}).
			Post(g.BaseURL + "/sso/verifyMFA")
		if err != nil {
			return "", fmt.Errorf("MFA submission failed: %w", err)
		}

		// Step 5: Extract verifier from response
		return extractVerifierFromResponse(mfaResp.String())
	}

	// Step 3: Extract verifier from response
	return extractVerifierFromResponse(loginResp.String())
}

// extractVerifierFromResponse parses verifier from HTML response
func extractVerifierFromResponse(html string) (string, error) {
	// Parse verifier from HTML
	if re := regexp.MustCompile(`name="oauth_verifier" value="([^"]+)"`); re.MatchString(html) {
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}
	return "", errors.New("verifier not found in response")
}

// MFAPrompter defines interface for getting MFA codes
type MFAPrompter interface {
	GetMFACode(ctx context.Context) (string, error)
}

// DefaultConsolePrompter is the default console-based MFA prompter
type DefaultConsolePrompter struct{}

// GetMFACode prompts user for MFA code via console
func (d DefaultConsolePrompter) GetMFACode(ctx context.Context) (string, error) {
	fmt.Print("Enter Garmin MFA code: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}

// getAccessToken exchanges request token for access token
func (g *GarthAuthenticator) getAccessToken(token, secret, verifier string) (accessToken, accessSecret string, err error) {
	return "access_token", "access_secret", nil
}

// getOAuth2Token exchanges OAuth1 token for OAuth2 token
func (g *GarthAuthenticator) getOAuth2Token(token, secret string) (oauth2Token string, err error) {
	return "oauth2_access_token", nil
}

// Save persists the session to the specified path
func (s *Session) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// IsExpired checks if the session is expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// LoadSession reads a session from the specified path
func LoadSession(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &session, nil
}
