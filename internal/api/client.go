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

type Client struct {
	HTTPClient  *resty.Client
	sessionPath string
	session     *garth.Session
}

// NewClient creates a new API client with session management
func NewClient(session *garth.Session, sessionPath string) (*Client, error) {
	if session == nil {
		return nil, errors.New("session is required")
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

	if c.sessionPath == "" {
		return errors.New("session path not configured for refresh")
	}

	session, err := garth.LoadSession(c.sessionPath)
	if err != nil {
		return fmt.Errorf("failed to load session for refresh: %w", err)
	}

	if session.IsExpired() {
		return errors.New("session expired, please reauthenticate")
	}

	c.session = session
	c.HTTPClient.SetHeader("Authorization", "Bearer "+session.OAuth2Token)
	return nil
}

// handleAPIError processes non-200 responses
func handleAPIError(resp *resty.Response) error {
	errorResponse := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}

	if err := json.Unmarshal(resp.Body(), &errorResponse); err == nil {
		return fmt.Errorf("API error %d: %s", errorResponse.Code, errorResponse.Message)
	}

	return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
}
