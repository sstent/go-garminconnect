package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Authenticate handles Garmin Connect authentication with MFA support
func (c *AuthClient) Authenticate(ctx context.Context, username, password, mfaToken string) (*Token, error) {
	// Create login form data
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("embed", "false")
	data.Set("rememberme", "on")

	// Create login request
	loginURL := fmt.Sprintf("%s%s", c.BaseURL, c.LoginPath)
	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; go-garminconnect/1.0)")

	// Send login request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if MFA is required
	if resp.StatusCode == http.StatusUnauthorized {
		// Parse MFA response
		var mfaResponse struct {
			MFAToken string `json:"mfaToken"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&mfaResponse); err != nil {
			return nil, fmt.Errorf("failed to parse MFA response: %w", err)
		}

		// Validate MFA token
		if mfaToken == "" {
			return nil, errors.New("MFA required but no token provided")
		}

		// Create MFA verification request
		mfaData := url.Values{}
		mfaData.Set("token", mfaResponse.MFAToken)
		mfaData.Set("rememberme", "on")
		mfaData.Set("mfaCode", mfaToken)

		req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(mfaData.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create MFA request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; go-garminconnect/1.0)")

		// Send MFA request
		resp, err = c.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("MFA request failed: %w", err)
		}
		defer resp.Body.Close()
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed: %d\n%s", resp.StatusCode, body)
	}

	// Parse response cookies to get tokens
	var token Token
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			token.AccessToken = cookie.Value
		} else if cookie.Name == "refresh_token" {
			token.RefreshToken = cookie.Value
		}
	}

	// Validate tokens
	if token.AccessToken == "" || token.RefreshToken == "" {
		return nil, errors.New("tokens not found in authentication response")
	}

	// Set expiration time
	token.Expiry = time.Now().Add(time.Duration(3600) * time.Second)
	token.ExpiresIn = 3600
	token.TokenType = "Bearer"

	return &token, nil
}

// RefreshToken exchanges a refresh token for a new access token
func (c *AuthClient) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	// Create token refresh data
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token.RefreshToken)

	// Create refresh token request
	req, err := http.NewRequestWithContext(ctx, "POST", c.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; go-garminconnect/1.0)")

	// Send refresh token request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d", resp.StatusCode)
	}

	// Parse token response
	var newToken Token
	if err := json.NewDecoder(resp.Body).Decode(&newToken); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Validate token
	if newToken.AccessToken == "" || newToken.RefreshToken == "" {
		return nil, errors.New("token response missing required fields")
	}

	// Set expiration time
	newToken.Expiry = time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)

	return &newToken, nil
}
