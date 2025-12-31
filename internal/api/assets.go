package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Asset represents a Shopline theme asset.
type Asset struct {
	Key         string    `json:"key"`
	ThemeID     string    `json:"theme_id"`
	ContentType string    `json:"content_type"`
	Size        int       `json:"size"`
	Checksum    string    `json:"checksum"`
	Value       string    `json:"value,omitempty"`
	PublicURL   string    `json:"public_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AssetsListResponse contains the list response.
type AssetsListResponse struct {
	Items []Asset `json:"items"`
}

// AssetUpdateRequest contains the request body for updating an asset.
type AssetUpdateRequest struct {
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	Attachment string `json:"attachment,omitempty"`
	SourceURL  string `json:"src,omitempty"`
}

// ListAssets retrieves a list of assets for a theme.
func (c *Client) ListAssets(ctx context.Context, themeID string) (*AssetsListResponse, error) {
	if strings.TrimSpace(themeID) == "" {
		return nil, fmt.Errorf("theme id is required")
	}
	var resp AssetsListResponse
	if err := c.Get(ctx, fmt.Sprintf("/themes/%s/assets", themeID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAsset retrieves a single asset by key.
func (c *Client) GetAsset(ctx context.Context, themeID, key string) (*Asset, error) {
	if strings.TrimSpace(themeID) == "" {
		return nil, fmt.Errorf("theme id is required")
	}
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("asset key is required")
	}
	params := url.Values{}
	params.Set("key", key)
	var asset Asset
	if err := c.Get(ctx, fmt.Sprintf("/themes/%s/assets?%s", themeID, params.Encode()), &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

// UpdateAsset creates or updates an asset.
func (c *Client) UpdateAsset(ctx context.Context, themeID string, req *AssetUpdateRequest) (*Asset, error) {
	if strings.TrimSpace(themeID) == "" {
		return nil, fmt.Errorf("theme id is required")
	}
	var asset Asset
	if err := c.Put(ctx, fmt.Sprintf("/themes/%s/assets", themeID), req, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

// DeleteAsset deletes an asset.
func (c *Client) DeleteAsset(ctx context.Context, themeID, key string) error {
	if strings.TrimSpace(themeID) == "" {
		return fmt.Errorf("theme id is required")
	}
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("asset key is required")
	}
	params := url.Values{}
	params.Set("key", key)
	return c.Delete(ctx, fmt.Sprintf("/themes/%s/assets?%s", themeID, params.Encode()))
}
