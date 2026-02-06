package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ChannelProductListing represents a product listing in a sales channel.
type ChannelProductListing struct {
	ID               string                  `json:"id"`
	ProductID        string                  `json:"product_id"`
	ChannelID        string                  `json:"channel_id"`
	Title            string                  `json:"title"`
	Handle           string                  `json:"handle"`
	Status           string                  `json:"status"`
	Published        bool                    `json:"published"`
	PublishedAt      *time.Time              `json:"published_at,omitempty"`
	AvailableForSale bool                    `json:"available_for_sale"`
	Variants         []ChannelVariantListing `json:"variants,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

// ChannelVariantListing represents a variant listing in a channel.
type ChannelVariantListing struct {
	ID                string `json:"id"`
	VariantID         string `json:"variant_id"`
	Title             string `json:"title"`
	SKU               string `json:"sku,omitempty"`
	Price             string `json:"price"`
	CompareAtPrice    string `json:"compare_at_price,omitempty"`
	InventoryQuantity int    `json:"inventory_quantity"`
	Available         bool   `json:"available"`
}

// ChannelProductsListOptions contains options for listing channel products.
type ChannelProductsListOptions struct {
	Page             int
	PageSize         int
	Published        *bool
	AvailableForSale *bool
	Status           string
}

// ChannelProductsListResponse is the paginated response for channel products.
type ChannelProductsListResponse = ListResponse[ChannelProductListing]

// ChannelProductPublishRequest contains the request body for publishing a product to a channel.
type ChannelProductPublishRequest struct {
	ProductID  string   `json:"product_id"`
	VariantIDs []string `json:"variant_ids,omitempty"`
}

// ChannelProductUpdateRequest contains the request body for updating a channel product listing.
type ChannelProductUpdateRequest struct {
	Published        *bool    `json:"published,omitempty"`
	AvailableForSale *bool    `json:"available_for_sale,omitempty"`
	VariantIDs       []string `json:"variant_ids,omitempty"`
}

// ListChannelProductListings retrieves products published to a channel.
func (c *Client) ListChannelProductListings(ctx context.Context, channelID string, opts *ChannelProductsListOptions) (*ChannelProductsListResponse, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	path := fmt.Sprintf("/channels/%s/product_listings", channelID)
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			BoolPtr("published", opts.Published).
			BoolPtr("available_for_sale", opts.AvailableForSale).
			String("status", opts.Status).
			Build()
	}

	var resp ChannelProductsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetChannelProductListing retrieves a single product listing from a channel.
func (c *Client) GetChannelProductListing(ctx context.Context, channelID, productID string) (*ChannelProductListing, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var listing ChannelProductListing
	if err := c.Get(ctx, fmt.Sprintf("/channels/%s/product_listings/%s", channelID, productID), &listing); err != nil {
		return nil, err
	}
	return &listing, nil
}

// PublishProductToChannelListing publishes a product to a channel.
func (c *Client) PublishProductToChannelListing(ctx context.Context, channelID string, req *ChannelProductPublishRequest) (*ChannelProductListing, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	var listing ChannelProductListing
	if err := c.Post(ctx, fmt.Sprintf("/channels/%s/product_listings", channelID), req, &listing); err != nil {
		return nil, err
	}
	return &listing, nil
}

// UpdateChannelProductListing updates a product listing in a channel.
func (c *Client) UpdateChannelProductListing(ctx context.Context, channelID, productID string, req *ChannelProductUpdateRequest) (*ChannelProductListing, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var listing ChannelProductListing
	if err := c.Put(ctx, fmt.Sprintf("/channels/%s/product_listings/%s", channelID, productID), req, &listing); err != nil {
		return nil, err
	}
	return &listing, nil
}

// UnpublishProductFromChannelListing removes a product from a channel.
func (c *Client) UnpublishProductFromChannelListing(ctx context.Context, channelID, productID string) error {
	if strings.TrimSpace(channelID) == "" {
		return fmt.Errorf("channel id is required")
	}
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/channels/%s/product_listings/%s", channelID, productID))
}
