package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// InventoryLevel represents inventory quantity at a location.
type InventoryLevel struct {
	ID              string    `json:"id"`
	InventoryItemID string    `json:"inventory_item_id"`
	LocationID      string    `json:"location_id"`
	Available       int       `json:"available"`
	Reserved        int       `json:"reserved"`
	Incoming        int       `json:"incoming"`
	OnHand          int       `json:"on_hand"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// InventoryListOptions contains options for listing inventory levels.
type InventoryListOptions struct {
	LocationID       string
	InventoryItemIDs []string
	Page             int
	PageSize         int
}

// InventoryListResponse is the paginated response for inventory levels.
type InventoryListResponse = ListResponse[InventoryLevel]

// InventoryAdjustRequest represents a request to adjust inventory.
type InventoryAdjustRequest struct {
	Delta int `json:"delta"`
}

// InventoryLevelsListOptions contains options for listing inventory levels.
type InventoryLevelsListOptions struct {
	Page            int
	PageSize        int
	LocationID      string
	InventoryItemID string
}

// InventoryLevelsListResponse is the paginated response for inventory levels.
type InventoryLevelsListResponse = ListResponse[InventoryLevel]

// InventoryLevelAdjustRequest contains the request body for adjusting an inventory level.
type InventoryLevelAdjustRequest struct {
	InventoryItemID     string `json:"inventory_item_id"`
	LocationID          string `json:"location_id"`
	AvailableAdjustment int    `json:"available_adjustment"`
}

// InventoryLevelSetRequest contains the request body for setting an inventory level.
type InventoryLevelSetRequest struct {
	InventoryItemID string `json:"inventory_item_id"`
	LocationID      string `json:"location_id"`
	Available       int    `json:"available"`
}

// ListInventoryLevels retrieves a list of inventory levels.
func (c *Client) ListInventoryLevels(ctx context.Context, opts *InventoryListOptions) (*InventoryListResponse, error) {
	path := "/inventory_levels"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.LocationID != "" {
			params.Set("location_id", opts.LocationID)
		}
		if len(opts.InventoryItemIDs) > 0 {
			params.Set("inventory_item_ids", strings.Join(opts.InventoryItemIDs, ","))
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp InventoryListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetInventoryLevel retrieves a single inventory level by ID.
func (c *Client) GetInventoryLevel(ctx context.Context, id string) (*InventoryLevel, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("inventory level id is required")
	}
	var level InventoryLevel
	if err := c.Get(ctx, fmt.Sprintf("/inventory_levels/%s", id), &level); err != nil {
		return nil, err
	}
	return &level, nil
}

// AdjustInventory adjusts the available quantity at an inventory level.
func (c *Client) AdjustInventory(ctx context.Context, id string, delta int) (*InventoryLevel, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("inventory level id is required")
	}
	req := InventoryAdjustRequest{Delta: delta}
	var level InventoryLevel
	if err := c.Post(ctx, fmt.Sprintf("/inventory_levels/%s/adjust", id), req, &level); err != nil {
		return nil, err
	}
	return &level, nil
}

// AdjustInventoryLevel adjusts the available quantity at a location.
func (c *Client) AdjustInventoryLevel(ctx context.Context, req *InventoryLevelAdjustRequest) (*InventoryLevel, error) {
	var level InventoryLevel
	if err := c.Post(ctx, "/inventory_levels/adjust", req, &level); err != nil {
		return nil, err
	}
	return &level, nil
}

// SetInventoryLevel sets the available quantity at a location.
func (c *Client) SetInventoryLevel(ctx context.Context, req *InventoryLevelSetRequest) (*InventoryLevel, error) {
	var level InventoryLevel
	if err := c.Post(ctx, "/inventory_levels/set", req, &level); err != nil {
		return nil, err
	}
	return &level, nil
}
