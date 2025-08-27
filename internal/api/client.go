package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

// Client handles communication with the Garmin Connect API
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	limiter    *rate.Limiter
	logger     Logger
}

// NewClient creates a new API client
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    u,
		httpClient: httpClient,
		limiter:    rate.NewLimiter(rate.Every(time.Second/10), 10), // 10 requests per second
		logger:     &stdLogger{},
	}, nil
}

// SetLogger sets the client's logger
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
}

// SetRateLimit configures the rate limiter
func (c *Client) SetRateLimit(interval time.Duration, burst int) {
	c.limiter = rate.NewLimiter(rate.Every(interval), burst)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, v interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, v)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body io.Reader, v interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, v)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader, v interface{}) error {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Create request
	u := c.baseURL.ResolveReference(&url.URL{Path: path})
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.logger.Debugf("Request: %s %s", method, u.String())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Debugf("Response status: %s", resp.Status)

	// Handle specific status codes
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("resource not found")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	if v == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	return nil
}

// Logger defines the logging interface for the client
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// stdLogger is the default logger that uses the standard log package
type stdLogger struct{}

func (l *stdLogger) Debugf(format string, args ...interface{}) {}
func (l *stdLogger) Infof(format string, args ...interface{})  {}
func (l *stdLogger) Errorf(format string, args ...interface{}) {}
