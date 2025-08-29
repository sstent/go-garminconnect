package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sstent/go-garminconnect/internal/auth/garth"
)

// Authenticator defines the method required for token refresh
type Authenticator interface {
	RefreshToken(oauth1Token, oauth1Secret string) (string, error)
}

type Client struct {
	HTTPClient  *resty.Client
	sessionPath string
	session     *garth.Session
	auth        Authenticator // Use interface for token refresh
}

// NewClient creates a new API client with session management
func NewClient(auth Authenticator, session *garth.Session, sessionPath string) (*Client, error) {
	// Try to load session from file if not provided
	if session == nil && sessionPath != "" {
		if loadedSession, err := garth.LoadSession(sessionPath); err == nil {
			session = loadedSession
		}
	}

	if session == nil || auth == nil {
		return nil, errors.New("both authenticator and session are required")
	}

	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetHeader("Authorization", "Bearer "+session.OAuth2Token)
	client.SetHeader("User-Agent", "go-garminconnect/1.0")
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("Accept", "application/json")

	return &Client{
		HTTPClient:  client,
		sessionPath: sessionPath,
		session:     session,
		auth:        auth,
	}, nil
}

// Get performs a GET request with automatic token refresh
func (c *Client) Get(ctx context.Context, path string, v interface{}) error {
	// Refresh token if needed
	if err := c.refreshTokenIfNeeded(); err != nil {
		return err
	}

	resp, err := c.HTTPClient.R().
		SetContext(ctx).
		SetResult(v).
		Get(path)

	if err != nil {
		return err
	}

	// Handle unmarshaling errors for successful responses
	if resp.IsSuccess() && resp.Error() != nil {
		return handleAPIError(resp)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		// Force token refresh on next attempt
		c.session = nil
		return errors.New("token expired, please reauthenticate")
	}

	if resp.StatusCode() >= 400 {
		return handleAPIError(resp)
	}

	return nil
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}, v interface{}) error {
	resp, err := c.HTTPClient.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(v).
		Post(path)

	if err != nil {
		return err
	}

	// Handle unmarshaling errors for successful responses
	if resp.IsSuccess() && resp.Error() != nil {
		return handleAPIError(resp)
	}

	if resp.StatusCode() >= 400 {
		return handleAPIError(resp)
	}

	return nil
}

// refreshTokenIfNeeded refreshes the token if expired
func (c *Client) refreshTokenIfNeeded() error {
	if c.session == nil || !c.session.IsExpired() {
		return nil
	}

	if c.auth == nil {
		return errors.New("authenticator not configured for refresh")
	}

	// Refresh OAuth2 token using OAuth1 credentials
	newToken, err := c.auth.RefreshToken(c.session.OAuth1Token, c.session.OAuth1Secret)
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// Update session and extend expiration
	c.session.OAuth2Token = newToken
	c.session.ExpiresAt = time.Now().Add(8 * time.Hour)
	c.HTTPClient.SetHeader("Authorization", "Bearer "+newToken)

	// Persist updated session
	if c.sessionPath != "" {
		if err := c.session.Save(c.sessionPath); err != nil {
			return fmt.Errorf("failed to save refreshed session: %w", err)
		}
	}

	return nil
}

// handleAPIError processes API errors including JSON unmarshaling issues
func handleAPIError(resp *resty.Response) error {
	// First try to parse as standard Garmin error format
	standardError := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}
	if err := json.Unmarshal(resp.Body(), &standardError); err == nil && standardError.Code != 0 {
		return fmt.Errorf("API error %d: %s", standardError.Code, standardError.Message)
	}

	// Try to parse as alternative error format
	altError := struct {
		Error string `json:"error"`
	}{}
	if err := json.Unmarshal(resp.Body(), &altError); err == nil && altError.Error != "" {
		return fmt.Errorf("API error %d: %s", resp.StatusCode(), altError.Error)
	}

	// Check for unmarshaling errors in successful responses
	if resp.IsSuccess() {
		return fmt.Errorf("failed to parse successful response: %s", resp.String())
	}

	return fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode(), resp.String())
}
