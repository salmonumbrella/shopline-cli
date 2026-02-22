package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ProductReviewCommentsListOptions contains options for listing product review comments.
type ProductReviewCommentsListOptions struct {
	Page     int
	PageSize int
}

// ListProductReviewComments lists product review comments (documented endpoint; raw JSON).
//
// Docs: GET /product_review_comments
func (c *Client) ListProductReviewComments(ctx context.Context, opts *ProductReviewCommentsListOptions) (json.RawMessage, error) {
	path := "/product_review_comments"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp json.RawMessage
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetProductReviewComment retrieves a product review comment by id (documented endpoint; raw JSON).
//
// Docs: GET /product_review_comments/{id}
func (c *Client) GetProductReviewComment(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product review comment id is required")
	}

	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/product_review_comments/%s", id), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateProductReviewComment creates a product review comment (documented endpoint; raw JSON body).
//
// Docs: POST /product_review_comments
func (c *Client) CreateProductReviewComment(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/product_review_comments", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateProductReviewComment updates a product review comment (documented endpoint; raw JSON body).
//
// Docs: PUT /product_review_comments/{id}
func (c *Client) UpdateProductReviewComment(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product review comment id is required")
	}

	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/product_review_comments/%s", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteProductReviewComment deletes a product review comment (documented endpoint; raw JSON).
//
// Docs: DELETE /product_review_comments/{id}
func (c *Client) DeleteProductReviewComment(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product review comment id is required")
	}

	var resp json.RawMessage
	if err := c.DeleteWithBody(ctx, fmt.Sprintf("/product_review_comments/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkCreateProductReviewComments bulk creates product review comments (documented endpoint; raw JSON body).
//
// Docs: POST /product_review_comments/bulk
func (c *Client) BulkCreateProductReviewComments(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/product_review_comments/bulk", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkUpdateProductReviewComments bulk updates product review comments (documented endpoint; raw JSON body).
//
// Docs: PUT /product_review_comments/bulk
func (c *Client) BulkUpdateProductReviewComments(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Put(ctx, "/product_review_comments/bulk", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkDeleteProductReviewComments bulk deletes product review comments (documented endpoint; raw JSON body).
//
// Docs: DELETE /product_review_comments/bulk
func (c *Client) BulkDeleteProductReviewComments(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.DeleteWithBody(ctx, "/product_review_comments/bulk", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
