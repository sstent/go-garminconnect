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

	// Build URL with query parameters
	u := c.baseURL.ResolveReference(&url.URL{
		Path: "/body-composition",
		RawQuery: fmt.Sprintf("startDate=%s&endDate=%s",
			req.StartDate.Format("2006-01-02"),
			req.EndDate.Format("2006-01-02"),
		),
	})

	// Execute GET request
	var results []BodyComposition
	err := c.Get(ctx, u.String(), &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
