package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DomainStatus represents the status of a domain.
type DomainStatus string

const (
	DomainStatusActive    DomainStatus = "active"
	DomainStatusPending   DomainStatus = "pending"
	DomainStatusVerifying DomainStatus = "verifying"
	DomainStatusFailed    DomainStatus = "failed"
	DomainStatusExpired   DomainStatus = "expired"
)

// Domain represents a domain configuration.
type Domain struct {
	ID                string       `json:"id"`
	Host              string       `json:"host"`
	Primary           bool         `json:"primary"`
	SSL               bool         `json:"ssl"`
	SSLStatus         string       `json:"ssl_status"`
	Status            DomainStatus `json:"status"`
	VerificationDNS   string       `json:"verification_dns"`
	VerificationToken string       `json:"verification_token"`
	Verified          bool         `json:"verified"`
	VerifiedAt        *time.Time   `json:"verified_at"`
	ExpiresAt         *time.Time   `json:"expires_at"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

// DomainsListOptions contains options for listing domains.
type DomainsListOptions struct {
	Page     int
	PageSize int
	Status   DomainStatus
	Primary  *bool
}

// DomainsListResponse is the paginated response for domains.
type DomainsListResponse = ListResponse[Domain]

// DomainCreateRequest contains the request body for creating a domain.
type DomainCreateRequest struct {
	Host    string `json:"host"`
	Primary bool   `json:"primary,omitempty"`
}

// DomainUpdateRequest contains the request body for updating a domain.
type DomainUpdateRequest struct {
	Primary *bool `json:"primary,omitempty"`
}

// ListDomains retrieves a list of domains.
func (c *Client) ListDomains(ctx context.Context, opts *DomainsListOptions) (*DomainsListResponse, error) {
	path := "/domains"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", string(opts.Status)).
			BoolPtr("primary", opts.Primary).
			Build()
	}

	var resp DomainsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDomain retrieves a single domain by ID.
func (c *Client) GetDomain(ctx context.Context, id string) (*Domain, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("domain id is required")
	}
	var domain Domain
	if err := c.Get(ctx, fmt.Sprintf("/domains/%s", id), &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}

// CreateDomain creates a new domain.
func (c *Client) CreateDomain(ctx context.Context, req *DomainCreateRequest) (*Domain, error) {
	var domain Domain
	if err := c.Post(ctx, "/domains", req, &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}

// UpdateDomain updates an existing domain.
func (c *Client) UpdateDomain(ctx context.Context, id string, req *DomainUpdateRequest) (*Domain, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("domain id is required")
	}
	var domain Domain
	if err := c.Put(ctx, fmt.Sprintf("/domains/%s", id), req, &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}

// DeleteDomain deletes a domain.
func (c *Client) DeleteDomain(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("domain id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/domains/%s", id))
}

// VerifyDomain triggers domain verification.
func (c *Client) VerifyDomain(ctx context.Context, id string) (*Domain, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("domain id is required")
	}
	var domain Domain
	if err := c.Post(ctx, fmt.Sprintf("/domains/%s/verify", id), nil, &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}
