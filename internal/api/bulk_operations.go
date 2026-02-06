package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// BulkOperation represents a Shopline bulk operation.
type BulkOperation struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	Query           string     `json:"query,omitempty"`
	URL             string     `json:"url,omitempty"`
	ErrorCode       string     `json:"error_code,omitempty"`
	ObjectCount     int        `json:"object_count"`
	FileSize        int64      `json:"file_size"`
	RootObjectCount int        `json:"root_object_count"`
	PartialDataURL  string     `json:"partial_data_url,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

// BulkOperationsListOptions contains options for listing bulk operations.
type BulkOperationsListOptions struct {
	Page     int
	PageSize int
	Status   string
	Type     string
}

// BulkOperationsListResponse contains the list response.
type BulkOperationsListResponse struct {
	Items      []BulkOperation `json:"items"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalCount int             `json:"total_count"`
	HasMore    bool            `json:"has_more"`
}

// BulkOperationCreateRequest contains the request body for creating a bulk operation.
type BulkOperationCreateRequest struct {
	Query string `json:"query"`
}

// BulkOperationMutationRequest contains the request body for a mutation bulk operation.
type BulkOperationMutationRequest struct {
	Mutation         string `json:"mutation"`
	StagedUploadPath string `json:"staged_upload_path"`
}

// ListBulkOperations retrieves a list of bulk operations.
func (c *Client) ListBulkOperations(ctx context.Context, opts *BulkOperationsListOptions) (*BulkOperationsListResponse, error) {
	path := "/bulk_operations"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			String("type", opts.Type).
			Build()
	}

	var resp BulkOperationsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBulkOperation retrieves a single bulk operation by ID.
func (c *Client) GetBulkOperation(ctx context.Context, id string) (*BulkOperation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("bulk operation id is required")
	}
	var op BulkOperation
	if err := c.Get(ctx, fmt.Sprintf("/bulk_operations/%s", id), &op); err != nil {
		return nil, err
	}
	return &op, nil
}

// CreateBulkQuery creates a bulk query operation.
func (c *Client) CreateBulkQuery(ctx context.Context, req *BulkOperationCreateRequest) (*BulkOperation, error) {
	var op BulkOperation
	if err := c.Post(ctx, "/bulk_operations/queries", req, &op); err != nil {
		return nil, err
	}
	return &op, nil
}

// CreateBulkMutation creates a bulk mutation operation.
func (c *Client) CreateBulkMutation(ctx context.Context, req *BulkOperationMutationRequest) (*BulkOperation, error) {
	var op BulkOperation
	if err := c.Post(ctx, "/bulk_operations/mutations", req, &op); err != nil {
		return nil, err
	}
	return &op, nil
}

// CancelBulkOperation cancels a running bulk operation.
func (c *Client) CancelBulkOperation(ctx context.Context, id string) (*BulkOperation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("bulk operation id is required")
	}
	var op BulkOperation
	if err := c.Post(ctx, fmt.Sprintf("/bulk_operations/%s/cancel", id), nil, &op); err != nil {
		return nil, err
	}
	return &op, nil
}

// GetCurrentBulkOperation retrieves the currently running bulk operation (if any).
func (c *Client) GetCurrentBulkOperation(ctx context.Context) (*BulkOperation, error) {
	var op BulkOperation
	if err := c.Get(ctx, "/bulk_operations/current", &op); err != nil {
		return nil, err
	}
	return &op, nil
}
