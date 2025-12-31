package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MembershipTier represents a membership tier.
type MembershipTier struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	Description string    `json:"description"`
	MinPoints   int       `json:"min_points"`
	MaxPoints   int       `json:"max_points"`
	Discount    float64   `json:"discount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MembershipTiersListOptions contains options for listing tiers.
type MembershipTiersListOptions struct {
	Page     int
	PageSize int
}

// MembershipTiersListResponse is the paginated response for membership tiers.
type MembershipTiersListResponse = ListResponse[MembershipTier]

// MembershipTierCreateRequest contains the request body for creating a tier.
type MembershipTierCreateRequest struct {
	Name        string  `json:"name"`
	Level       int     `json:"level"`
	Description string  `json:"description,omitempty"`
	MinPoints   int     `json:"min_points"`
	MaxPoints   int     `json:"max_points,omitempty"`
	Discount    float64 `json:"discount,omitempty"`
}

// ListMembershipTiers retrieves a list of membership tiers.
// Note: This endpoint returns an array directly, not paginated.
func (c *Client) ListMembershipTiers(ctx context.Context, opts *MembershipTiersListOptions) (*MembershipTiersListResponse, error) {
	path := "/membership_tiers"
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

	// API returns array directly, wrap in ListResponse
	var tiers []MembershipTier
	if err := c.Get(ctx, path, &tiers); err != nil {
		return nil, err
	}
	return &MembershipTiersListResponse{Items: tiers}, nil
}

// GetMembershipTier retrieves a single membership tier by ID.
func (c *Client) GetMembershipTier(ctx context.Context, id string) (*MembershipTier, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("tier id is required")
	}
	var tier MembershipTier
	if err := c.Get(ctx, fmt.Sprintf("/membership_tiers/%s", id), &tier); err != nil {
		return nil, err
	}
	return &tier, nil
}

// CreateMembershipTier creates a new membership tier.
func (c *Client) CreateMembershipTier(ctx context.Context, req *MembershipTierCreateRequest) (*MembershipTier, error) {
	var tier MembershipTier
	if err := c.Post(ctx, "/membership_tiers", req, &tier); err != nil {
		return nil, err
	}
	return &tier, nil
}

// DeleteMembershipTier deletes a membership tier.
func (c *Client) DeleteMembershipTier(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("tier id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/membership_tiers/%s", id))
}
