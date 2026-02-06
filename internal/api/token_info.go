package api

import (
	"context"
	"encoding/json"
)

// GetTokenInfo retrieves information about the current access token.
//
// Docs: GET /token/info
func (c *Client) GetTokenInfo(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/token/info", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
