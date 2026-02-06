package api

import (
	"context"
	"fmt"
	"strings"
)

// Storefront OAuth applications (documented endpoints).
//
// Docs: /storefront/oauth_applications
// These appear similar to the existing /storefront_oauth/clients endpoints, but use
// a different path in the Open API reference.

type (
	StorefrontOAuthApplication              = StorefrontOAuthClient
	StorefrontOAuthApplicationsListOptions  = StorefrontOAuthClientsListOptions
	StorefrontOAuthApplicationsListResponse = StorefrontOAuthClientsListResponse
	StorefrontOAuthApplicationCreateRequest = StorefrontOAuthClientCreateRequest
)

// ListStorefrontOAuthApplications lists storefront OAuth applications.
//
// Docs: GET /storefront/oauth_applications
func (c *Client) ListStorefrontOAuthApplications(ctx context.Context, opts *StorefrontOAuthApplicationsListOptions) (*StorefrontOAuthApplicationsListResponse, error) {
	path := "/storefront/oauth_applications"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp StorefrontOAuthApplicationsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStorefrontOAuthApplication gets a single storefront OAuth application.
//
// Docs: GET /storefront/oauth_applications/{id}
func (c *Client) GetStorefrontOAuthApplication(ctx context.Context, id string) (*StorefrontOAuthApplication, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("oauth application id is required")
	}
	var app StorefrontOAuthApplication
	if err := c.Get(ctx, fmt.Sprintf("/storefront/oauth_applications/%s", id), &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// CreateStorefrontOAuthApplication creates a storefront OAuth application.
//
// Docs: POST /storefront/oauth_applications
func (c *Client) CreateStorefrontOAuthApplication(ctx context.Context, req *StorefrontOAuthApplicationCreateRequest) (*StorefrontOAuthApplication, error) {
	var app StorefrontOAuthApplication
	if err := c.Post(ctx, "/storefront/oauth_applications", req, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// DeleteStorefrontOAuthApplication deletes a storefront OAuth application.
//
// Docs: DELETE /storefront/oauth_applications/{id}
func (c *Client) DeleteStorefrontOAuthApplication(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("oauth application id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/storefront/oauth_applications/%s", id))
}
