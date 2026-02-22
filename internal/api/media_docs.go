package api

import (
	"context"
	"encoding/json"
)

// CreateMediaImage creates an image (documented endpoint).
//
// Docs: POST /media
func (c *Client) CreateMediaImage(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/media", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
