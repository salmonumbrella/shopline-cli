package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Staff represents a Shopline staff account.
type Staff struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        string    `json:"phone,omitempty"`
	AccountOwner bool      `json:"account_owner"`
	Locale       string    `json:"locale"`
	Permissions  []string  `json:"permissions"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StaffsListOptions contains options for listing staff accounts.
type StaffsListOptions struct {
	Page     int
	PageSize int
}

// StaffsListResponse is the paginated response for staffs.
type StaffsListResponse = ListResponse[Staff]

// StaffUpdateRequest contains the data for updating a staff member.
type StaffUpdateRequest struct {
	FirstName   string   `json:"first_name,omitempty"`
	LastName    string   `json:"last_name,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	Locale      string   `json:"locale,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// ListStaffs retrieves a list of staff accounts.
func (c *Client) ListStaffs(ctx context.Context, opts *StaffsListOptions) (*StaffsListResponse, error) {
	path := "/staffs"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp StaffsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStaff retrieves a single staff account by ID.
func (c *Client) GetStaff(ctx context.Context, id string) (*Staff, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("staff id is required")
	}
	var staff Staff
	if err := c.Get(ctx, fmt.Sprintf("/staffs/%s", id), &staff); err != nil {
		return nil, err
	}
	return &staff, nil
}

// UpdateStaff updates an existing staff account.
func (c *Client) UpdateStaff(ctx context.Context, id string, req *StaffUpdateRequest) (*Staff, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("staff id is required")
	}
	var staff Staff
	if err := c.Put(ctx, fmt.Sprintf("/staffs/%s", id), req, &staff); err != nil {
		return nil, err
	}
	return &staff, nil
}

// DeleteStaff removes a staff account.
func (c *Client) DeleteStaff(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("staff id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/staffs/%s", id))
}
