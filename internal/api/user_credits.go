package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type UserCreditsListOptions struct {
	Page     int
	PageSize int
}

// ListUserCredits retrieves store credits records.
//
// Docs: GET /user_credits
func (c *Client) ListUserCredits(ctx context.Context, opts *UserCreditsListOptions) (json.RawMessage, error) {
	path := "/user_credits"
	if opts != nil {
		q := NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize)
		path += q.Build()
	}
	var resp json.RawMessage
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkUpdateUserCredits bulk updates multiple user credits records.
//
// Docs: POST /user_credits/bulk_update
func (c *Client) BulkUpdateUserCredits(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/user_credits/bulk_update", body, &resp); err != nil {
		return nil, fmt.Errorf("bulk update user credits failed: %w", err)
	}
	return resp, nil
}
