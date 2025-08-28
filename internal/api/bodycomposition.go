package api

import (
	"context"
	"fmt"
	"net/url"
)

// GetBodyComposition retrieves body composition data within a date range
func (c *Client) GetBodyComposition(ctx context.Context, req BodyCompositionRequest) ([]BodyComposition, error) {
	// Validate date range
	if req.StartDate.IsZero() || req.EndDate.IsZero() || req.StartDate.After(req.EndDate) {
		return nil, fmt.Errorf("invalid date range: start %s to end %s",
			req.StartDate.Format("2006-01-02"),
			req.EndDate.Format("2006-01-02"))
	}

	// Build query parameters
	params := url.Values{}
	params.Add("startDate", req.StartDate.Format("2006-01-02"))
	params.Add("endDate", req.EndDate.Format("2006-01-02"))
	path := fmt.Sprintf("/body-composition?%s", params.Encode())

	// Execute GET request
	var results []BodyComposition
	err := c.Get(ctx, path, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get body composition: %w", err)
	}

	return results, nil
}
