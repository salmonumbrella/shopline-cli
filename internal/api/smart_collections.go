package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SmartCollection represents a Shopline smart collection (auto-populated based on rules).
type SmartCollection struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Handle      string    `json:"handle"`
	BodyHTML    string    `json:"body_html"`
	SortOrder   string    `json:"sort_order"`  // alpha-asc, alpha-desc, best-selling, created, created-desc, manual, price-asc, price-desc
	Disjunctive bool      `json:"disjunctive"` // true = any rule matches, false = all rules match
	Rules       []Rule    `json:"rules"`
	Published   bool      `json:"published"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Rule defines a condition for smart collection product matching.
type Rule struct {
	Column    string `json:"column"`    // title, type, vendor, variant_title, etc.
	Relation  string `json:"relation"`  // equals, not_equals, contains, etc.
	Condition string `json:"condition"` // the value to match
}

// SmartCollectionsListOptions contains options for listing smart collections.
type SmartCollectionsListOptions struct {
	Page     int
	PageSize int
}

// SmartCollectionsListResponse is the paginated response for smart collections.
type SmartCollectionsListResponse = ListResponse[SmartCollection]

// SmartCollectionCreateRequest contains the request body for creating a smart collection.
type SmartCollectionCreateRequest struct {
	Title       string `json:"title"`
	Handle      string `json:"handle,omitempty"`
	BodyHTML    string `json:"body_html,omitempty"`
	SortOrder   string `json:"sort_order,omitempty"`
	Disjunctive bool   `json:"disjunctive"`
	Rules       []Rule `json:"rules"`
	Published   bool   `json:"published"`
}

// SmartCollectionUpdateRequest contains the request body for updating a smart collection.
type SmartCollectionUpdateRequest struct {
	Title       string `json:"title,omitempty"`
	Handle      string `json:"handle,omitempty"`
	BodyHTML    string `json:"body_html,omitempty"`
	SortOrder   string `json:"sort_order,omitempty"`
	Disjunctive *bool  `json:"disjunctive,omitempty"`
	Rules       []Rule `json:"rules,omitempty"`
	Published   *bool  `json:"published,omitempty"`
}

// ListSmartCollections retrieves a list of smart collections.
func (c *Client) ListSmartCollections(ctx context.Context, opts *SmartCollectionsListOptions) (*SmartCollectionsListResponse, error) {
	path := "/smart_collections"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp SmartCollectionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSmartCollection retrieves a single smart collection by ID.
func (c *Client) GetSmartCollection(ctx context.Context, id string) (*SmartCollection, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("smart collection id is required")
	}
	var collection SmartCollection
	if err := c.Get(ctx, fmt.Sprintf("/smart_collections/%s", id), &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

// CreateSmartCollection creates a new smart collection.
func (c *Client) CreateSmartCollection(ctx context.Context, req *SmartCollectionCreateRequest) (*SmartCollection, error) {
	var collection SmartCollection
	if err := c.Post(ctx, "/smart_collections", req, &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

// UpdateSmartCollection updates an existing smart collection.
func (c *Client) UpdateSmartCollection(ctx context.Context, id string, req *SmartCollectionUpdateRequest) (*SmartCollection, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("smart collection id is required")
	}
	var collection SmartCollection
	if err := c.Put(ctx, fmt.Sprintf("/smart_collections/%s", id), req, &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

// DeleteSmartCollection deletes a smart collection.
func (c *Client) DeleteSmartCollection(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("smart collection id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/smart_collections/%s", id))
}
