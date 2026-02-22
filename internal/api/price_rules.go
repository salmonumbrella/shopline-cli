package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PriceRule represents a Shopline price rule.
type PriceRule struct {
	ID                      string    `json:"id"`
	Title                   string    `json:"title"`
	TargetType              string    `json:"target_type"`
	TargetSelection         string    `json:"target_selection"`
	AllocationMethod        string    `json:"allocation_method"`
	ValueType               string    `json:"value_type"`
	Value                   string    `json:"value"`
	CustomerSelection       string    `json:"customer_selection"`
	OncePerCustomer         bool      `json:"once_per_customer"`
	UsageLimit              int       `json:"usage_limit"`
	StartsAt                time.Time `json:"starts_at"`
	EndsAt                  time.Time `json:"ends_at"`
	PrerequisiteSubtotalMin string    `json:"prerequisite_subtotal_min"`
	PrerequisiteQuantityMin int       `json:"prerequisite_quantity_min"`
	EntitledProductIDs      []string  `json:"entitled_product_ids"`
	EntitledCollectionIDs   []string  `json:"entitled_collection_ids"`
	PrerequisiteProductIDs  []string  `json:"prerequisite_product_ids"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// PriceRulesListOptions contains options for listing price rules.
type PriceRulesListOptions struct {
	Page     int
	PageSize int
}

// PriceRulesListResponse contains the list response.
type PriceRulesListResponse struct {
	Items      []PriceRule `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	HasMore    bool        `json:"has_more"`
}

// PriceRuleCreateRequest contains the request body for creating a price rule.
type PriceRuleCreateRequest struct {
	Title                   string     `json:"title"`
	TargetType              string     `json:"target_type"`
	TargetSelection         string     `json:"target_selection"`
	AllocationMethod        string     `json:"allocation_method"`
	ValueType               string     `json:"value_type"`
	Value                   string     `json:"value"`
	CustomerSelection       string     `json:"customer_selection,omitempty"`
	OncePerCustomer         bool       `json:"once_per_customer,omitempty"`
	UsageLimit              int        `json:"usage_limit,omitempty"`
	StartsAt                *time.Time `json:"starts_at,omitempty"`
	EndsAt                  *time.Time `json:"ends_at,omitempty"`
	PrerequisiteSubtotalMin string     `json:"prerequisite_subtotal_min,omitempty"`
	PrerequisiteQuantityMin int        `json:"prerequisite_quantity_min,omitempty"`
	EntitledProductIDs      []string   `json:"entitled_product_ids,omitempty"`
	EntitledCollectionIDs   []string   `json:"entitled_collection_ids,omitempty"`
}

// PriceRuleUpdateRequest contains the request body for updating a price rule.
type PriceRuleUpdateRequest struct {
	Title                   string     `json:"title,omitempty"`
	ValueType               string     `json:"value_type,omitempty"`
	Value                   string     `json:"value,omitempty"`
	OncePerCustomer         *bool      `json:"once_per_customer,omitempty"`
	UsageLimit              *int       `json:"usage_limit,omitempty"`
	StartsAt                *time.Time `json:"starts_at,omitempty"`
	EndsAt                  *time.Time `json:"ends_at,omitempty"`
	PrerequisiteSubtotalMin string     `json:"prerequisite_subtotal_min,omitempty"`
}

// ListPriceRules retrieves a list of price rules.
func (c *Client) ListPriceRules(ctx context.Context, opts *PriceRulesListOptions) (*PriceRulesListResponse, error) {
	path := "/price_rules"
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

	var resp PriceRulesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPriceRule retrieves a single price rule by ID.
func (c *Client) GetPriceRule(ctx context.Context, id string) (*PriceRule, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("price rule id is required")
	}
	var rule PriceRule
	if err := c.Get(ctx, fmt.Sprintf("/price_rules/%s", id), &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreatePriceRule creates a new price rule.
func (c *Client) CreatePriceRule(ctx context.Context, req *PriceRuleCreateRequest) (*PriceRule, error) {
	var rule PriceRule
	if err := c.Post(ctx, "/price_rules", req, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdatePriceRule updates an existing price rule.
func (c *Client) UpdatePriceRule(ctx context.Context, id string, req *PriceRuleUpdateRequest) (*PriceRule, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("price rule id is required")
	}
	var rule PriceRule
	if err := c.Put(ctx, fmt.Sprintf("/price_rules/%s", id), req, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// DeletePriceRule deletes a price rule.
func (c *Client) DeletePriceRule(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("price rule id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/price_rules/%s", id))
}
