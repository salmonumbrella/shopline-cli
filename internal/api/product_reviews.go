package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ProductReview represents a Shopline product review.
type ProductReview struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Rating       int       `json:"rating"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	Verified     bool      `json:"verified"`
	HelpfulCount int       `json:"helpful_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProductReviewsListOptions contains options for listing product reviews.
type ProductReviewsListOptions struct {
	Page      int
	PageSize  int
	ProductID string
	Status    string
	Rating    int
}

// ProductReviewsListResponse is the paginated response for product reviews.
type ProductReviewsListResponse = ListResponse[ProductReview]

// ProductReviewCreateRequest contains the data for creating a product review.
type ProductReviewCreateRequest struct {
	ProductID    string `json:"product_id"`
	CustomerID   string `json:"customer_id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
	Rating       int    `json:"rating"`
	Title        string `json:"title,omitempty"`
	Content      string `json:"content"`
}

// ListProductReviews retrieves a list of product reviews.
func (c *Client) ListProductReviews(ctx context.Context, opts *ProductReviewsListOptions) (*ProductReviewsListResponse, error) {
	path := "/product_reviews"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("product_id", opts.ProductID).
			String("status", opts.Status).
			Int("rating", opts.Rating).
			Build()
	}

	var resp ProductReviewsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetProductReview retrieves a single product review by ID.
func (c *Client) GetProductReview(ctx context.Context, id string) (*ProductReview, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product review id is required")
	}
	var review ProductReview
	if err := c.Get(ctx, fmt.Sprintf("/product_reviews/%s", id), &review); err != nil {
		return nil, err
	}
	return &review, nil
}

// CreateProductReview creates a new product review.
func (c *Client) CreateProductReview(ctx context.Context, req *ProductReviewCreateRequest) (*ProductReview, error) {
	var review ProductReview
	if err := c.Post(ctx, "/product_reviews", req, &review); err != nil {
		return nil, err
	}
	return &review, nil
}

// DeleteProductReview deletes a product review.
func (c *Client) DeleteProductReview(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product review id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/product_reviews/%s", id))
}
