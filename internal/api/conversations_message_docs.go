package api

import (
	"context"
	"encoding/json"
)

// CreateConversationShopMessage creates a shop message (documented endpoint; raw JSON body).
//
// Docs: POST /conversations/message
func (c *Client) CreateConversationShopMessage(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/conversations/message", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
