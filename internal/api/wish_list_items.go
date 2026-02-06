package api

import (
	"context"
	"encoding/json"
)

// Wish list items (documented endpoints).
//
// Docs: GET/POST/DELETE /wish_list_items

type WishListItemsListOptions struct {
	Page     int
	PageSize int
}

func (c *Client) ListWishListItems(ctx context.Context, opts *WishListItemsListOptions) (json.RawMessage, error) {
	path := "/wish_list_items"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}
	var resp json.RawMessage
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateWishListItem(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/wish_list_items", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteWishListItems(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.DeleteWithBody(ctx, "/wish_list_items", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
