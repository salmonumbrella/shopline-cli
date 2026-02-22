package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CustomerGroup represents a Shopline customer group.
type CustomerGroup struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CustomerCount int       `json:"customer_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CustomerGroupsListOptions contains options for listing customer groups.
type CustomerGroupsListOptions struct {
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// CustomerGroupsListResponse is the paginated response for customer groups.
type CustomerGroupsListResponse = ListResponse[CustomerGroup]

// CustomerGroupCreateRequest contains the request body for creating a customer group.
type CustomerGroupCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CustomerGroupUpdateRequest contains the request body for updating a customer group.
type CustomerGroupUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListCustomerGroups retrieves a list of customer groups.
func (c *Client) ListCustomerGroups(ctx context.Context, opts *CustomerGroupsListOptions) (*CustomerGroupsListResponse, error) {
	path := "/customer_groups"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			params.Set("sort_by", opts.SortBy)
		}
		if opts.SortOrder != "" {
			params.Set("sort_order", opts.SortOrder)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp CustomerGroupsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomerGroup retrieves a single customer group by ID.
func (c *Client) GetCustomerGroup(ctx context.Context, id string) (*CustomerGroup, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer group id is required")
	}
	var group CustomerGroup
	if err := c.Get(ctx, fmt.Sprintf("/customer_groups/%s", id), &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// CreateCustomerGroup creates a new customer group.
func (c *Client) CreateCustomerGroup(ctx context.Context, req *CustomerGroupCreateRequest) (*CustomerGroup, error) {
	var group CustomerGroup
	if err := c.Post(ctx, "/customer_groups", req, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateCustomerGroup updates an existing customer group.
func (c *Client) UpdateCustomerGroup(ctx context.Context, id string, req *CustomerGroupUpdateRequest) (*CustomerGroup, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer group id is required")
	}
	var group CustomerGroup
	if err := c.Put(ctx, fmt.Sprintf("/customer_groups/%s", id), req, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// DeleteCustomerGroup deletes a customer group.
func (c *Client) DeleteCustomerGroup(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("customer group id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/customer_groups/%s", id))
}

// CustomerGroupSearchOptions contains options for searching customer groups.
type CustomerGroupSearchOptions struct {
	Query    string
	Page     int
	PageSize int
}

// CustomerGroupIDsResponse represents customer IDs in a group.
type CustomerGroupIDsResponse struct {
	CustomerIDs []string `json:"customer_ids"`
	TotalCount  int      `json:"total_count"`
}

// SearchCustomerGroups searches for customer groups.
func (c *Client) SearchCustomerGroups(ctx context.Context, opts *CustomerGroupSearchOptions) (*CustomerGroupsListResponse, error) {
	path := "/customer_groups/search" + NewQuery().
		String("query", opts.Query).
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp CustomerGroupsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomerGroupIDs retrieves customer IDs in a customer group.
func (c *Client) GetCustomerGroupIDs(ctx context.Context, groupID string) (*CustomerGroupIDsResponse, error) {
	if strings.TrimSpace(groupID) == "" {
		return nil, fmt.Errorf("customer group id is required")
	}
	var resp CustomerGroupIDsResponse
	if err := c.Get(ctx, fmt.Sprintf("/customer_groups/%s/customer_ids", groupID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomerGroupChildren retrieves child customer groups of a customer group.
//
// Docs: GET /customer_groups/{parentCustomerGroupId}/customer_group_children
func (c *Client) GetCustomerGroupChildren(ctx context.Context, parentGroupID string) (json.RawMessage, error) {
	if strings.TrimSpace(parentGroupID) == "" {
		return nil, fmt.Errorf("customer group id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customer_groups/%s/customer_group_children", parentGroupID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetCustomerGroupChildCustomerIDs retrieves customer IDs in a child customer group of a parent group.
//
// Docs: GET /customer_groups/{parentCustomerGroupId}/customer_group_children/{id}/customer_ids
func (c *Client) GetCustomerGroupChildCustomerIDs(ctx context.Context, parentGroupID, childGroupID string) (*CustomerGroupIDsResponse, error) {
	if strings.TrimSpace(parentGroupID) == "" || strings.TrimSpace(childGroupID) == "" {
		return nil, fmt.Errorf("customer group id is required")
	}
	var resp CustomerGroupIDsResponse
	if err := c.Get(ctx, fmt.Sprintf("/customer_groups/%s/customer_group_children/%s/customer_ids", parentGroupID, childGroupID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
