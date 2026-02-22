package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SizeChart represents a Shopline size chart.
type SizeChart struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Unit        string         `json:"unit"`
	Headers     []string       `json:"headers"`
	Rows        []SizeChartRow `json:"rows"`
	ProductIDs  []string       `json:"product_ids"`
	Active      bool           `json:"active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// SizeChartRow represents a row in a size chart.
type SizeChartRow struct {
	Size   string   `json:"size"`
	Values []string `json:"values"`
}

// SizeChartsListOptions contains options for listing size charts.
type SizeChartsListOptions struct {
	Page     int
	PageSize int
	Active   *bool
}

// SizeChartsListResponse is the paginated response for size charts.
type SizeChartsListResponse = ListResponse[SizeChart]

// SizeChartCreateRequest contains the data for creating a size chart.
type SizeChartCreateRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Unit        string         `json:"unit,omitempty"`
	Headers     []string       `json:"headers,omitempty"`
	Rows        []SizeChartRow `json:"rows,omitempty"`
	ProductIDs  []string       `json:"product_ids,omitempty"`
	Active      bool           `json:"active"`
}

// SizeChartUpdateRequest contains the data for updating a size chart.
type SizeChartUpdateRequest struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Unit        string         `json:"unit,omitempty"`
	Headers     []string       `json:"headers,omitempty"`
	Rows        []SizeChartRow `json:"rows,omitempty"`
	ProductIDs  []string       `json:"product_ids,omitempty"`
	Active      *bool          `json:"active,omitempty"`
}

// ListSizeCharts retrieves a list of size charts.
func (c *Client) ListSizeCharts(ctx context.Context, opts *SizeChartsListOptions) (*SizeChartsListResponse, error) {
	path := "/size_charts"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			BoolPtr("active", opts.Active).
			Build()
	}

	var resp SizeChartsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSizeChart retrieves a single size chart by ID.
func (c *Client) GetSizeChart(ctx context.Context, id string) (*SizeChart, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("size chart id is required")
	}
	var sizeChart SizeChart
	if err := c.Get(ctx, fmt.Sprintf("/size_charts/%s", id), &sizeChart); err != nil {
		return nil, err
	}
	return &sizeChart, nil
}

// CreateSizeChart creates a new size chart.
func (c *Client) CreateSizeChart(ctx context.Context, req *SizeChartCreateRequest) (*SizeChart, error) {
	var sizeChart SizeChart
	if err := c.Post(ctx, "/size_charts", req, &sizeChart); err != nil {
		return nil, err
	}
	return &sizeChart, nil
}

// UpdateSizeChart updates an existing size chart.
func (c *Client) UpdateSizeChart(ctx context.Context, id string, req *SizeChartUpdateRequest) (*SizeChart, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("size chart id is required")
	}
	var sizeChart SizeChart
	if err := c.Put(ctx, fmt.Sprintf("/size_charts/%s", id), req, &sizeChart); err != nil {
		return nil, err
	}
	return &sizeChart, nil
}

// DeleteSizeChart deletes a size chart.
func (c *Client) DeleteSizeChart(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("size chart id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/size_charts/%s", id))
}
