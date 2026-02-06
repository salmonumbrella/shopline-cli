package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// WishList represents a customer's wish list.
type WishList struct {
	ID          string         `json:"id"`
	CustomerID  string         `json:"customer_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	IsDefault   bool           `json:"is_default"`
	IsPublic    bool           `json:"is_public"`
	ShareToken  string         `json:"share_token"`
	ShareURL    string         `json:"share_url"`
	ItemCount   int            `json:"item_count"`
	Items       []WishListItem `json:"items"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// WishListItem represents an item in a wish list.
type WishListItem struct {
	ID             string    `json:"id"`
	ProductID      string    `json:"product_id"`
	VariantID      string    `json:"variant_id"`
	Title          string    `json:"title"`
	VariantTitle   string    `json:"variant_title"`
	Handle         string    `json:"handle"`
	Price          string    `json:"price"`
	CompareAtPrice string    `json:"compare_at_price"`
	Currency       string    `json:"currency"`
	Available      bool      `json:"available"`
	ImageURL       string    `json:"image_url"`
	Quantity       int       `json:"quantity"`
	Priority       int       `json:"priority"`
	Notes          string    `json:"notes"`
	AddedAt        time.Time `json:"added_at"`
}

// WishListsListOptions contains options for listing wish lists.
type WishListsListOptions struct {
	Page       int
	PageSize   int
	CustomerID string
	IsPublic   *bool
	SortBy     string
	SortOrder  string
}

// WishListsListResponse is the paginated response for wish lists.
type WishListsListResponse = ListResponse[WishList]

// WishListCreateRequest contains the data for creating a wish list.
type WishListCreateRequest struct {
	CustomerID  string `json:"customer_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
	IsPublic    bool   `json:"is_public,omitempty"`
}

// WishListItemCreateRequest contains the data for adding an item to a wish list.
type WishListItemCreateRequest struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id,omitempty"`
	Quantity  int    `json:"quantity,omitempty"`
	Priority  int    `json:"priority,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// ListWishLists retrieves a list of wish lists.
func (c *Client) ListWishLists(ctx context.Context, opts *WishListsListOptions) (*WishListsListResponse, error) {
	path := "/wish_lists"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("customer_id", opts.CustomerID).
			BoolPtr("is_public", opts.IsPublic).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp WishListsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetWishList retrieves a single wish list by ID.
func (c *Client) GetWishList(ctx context.Context, id string) (*WishList, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("wish list id is required")
	}
	var wishList WishList
	if err := c.Get(ctx, fmt.Sprintf("/wish_lists/%s", id), &wishList); err != nil {
		return nil, err
	}
	return &wishList, nil
}

// CreateWishList creates a new wish list.
func (c *Client) CreateWishList(ctx context.Context, req *WishListCreateRequest) (*WishList, error) {
	var wishList WishList
	if err := c.Post(ctx, "/wish_lists", req, &wishList); err != nil {
		return nil, err
	}
	return &wishList, nil
}

// DeleteWishList deletes a wish list.
func (c *Client) DeleteWishList(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("wish list id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/wish_lists/%s", id))
}

// AddWishListItem adds an item to a wish list.
func (c *Client) AddWishListItem(ctx context.Context, wishListID string, req *WishListItemCreateRequest) (*WishListItem, error) {
	if strings.TrimSpace(wishListID) == "" {
		return nil, fmt.Errorf("wish list id is required")
	}
	var item WishListItem
	if err := c.Post(ctx, fmt.Sprintf("/wish_lists/%s/items", wishListID), req, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// RemoveWishListItem removes an item from a wish list.
func (c *Client) RemoveWishListItem(ctx context.Context, wishListID, itemID string) error {
	if strings.TrimSpace(wishListID) == "" {
		return fmt.Errorf("wish list id is required")
	}
	if strings.TrimSpace(itemID) == "" {
		return fmt.Errorf("item id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/wish_lists/%s/items/%s", wishListID, itemID))
}
