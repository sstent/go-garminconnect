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

const BaseURL = "https://connect.garmin.com/modern/proxy"

// Client handles communication with the Garmin Connect API
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	limiter    *rate.Limiter
	logger     Logger
	token      string
}

// NewClient creates a new API client
func NewClient(token string) (*Client, error) {
	u, err := url.Parse(BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		baseURL:    u,
		httpClient: httpClient,
		limiter:    rate.NewLimiter(rate.Every(time.Second/10), 10), // 10 requests per second
		logger:     &stdLogger{},
		token:      token,
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

// setAuthHeaders adds authorization headers to requests
func (c *Client) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "go-garminconnect/1.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

// doRequest executes API requests with rate limiting and authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader, v interface{}) error {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Build full URL
	fullURL := c.baseURL.ResolveReference(&url.URL{Path: path}).String()

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// Add authentication headers
	c.setAuthHeaders(req)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return handleAPIError(resp)
	}

	// Parse successful response
	if v == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// handleAPIError processes non-200 responses
func handleAPIError(resp *http.Response) error {
	errorResponse := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
		return fmt.Errorf("API error %d: %s", errorResponse.Code, errorResponse.Message)
	}

	return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, v interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, v)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body io.Reader, v interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, v)
}

// Logger defines the logging interface
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// stdLogger is the default logger
type stdLogger struct{}

func (l *stdLogger) Debugf(format string, args ...interface{}) {}
func (l *stdLogger) Infof(format string, args ...interface{})  {}
func (l *stdLogger) Errorf(format string, args ...interface{}) {}
