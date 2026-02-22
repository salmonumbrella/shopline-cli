package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductMetafieldsEndpoints(t *testing.T) {
	type tc struct {
		name       string
		method     string
		path       string
		call       func(c *Client) error
		wantStatus int
	}

	tests := []tc{
		{
			name:   "list product metafields",
			method: http.MethodGet,
			path:   "/products/prod_123/metafields",
			call: func(c *Client) error {
				_, err := c.ListProductMetafields(context.Background(), "prod_123")
				return err
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "get product metafield",
			method: http.MethodGet,
			path:   "/products/prod_123/metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.GetProductMetafield(context.Background(), "prod_123", "mf_1")
				return err
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "create product metafield",
			method: http.MethodPost,
			path:   "/products/prod_123/metafields",
			call: func(c *Client) error {
				_, err := c.CreateProductMetafield(context.Background(), "prod_123", map[string]any{"key": "k"})
				return err
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "update product metafield",
			method: http.MethodPut,
			path:   "/products/prod_123/metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.UpdateProductMetafield(context.Background(), "prod_123", "mf_1", map[string]any{"value": "v"})
				return err
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "delete product metafield",
			method: http.MethodDelete,
			path:   "/products/prod_123/metafields/mf_1",
			call: func(c *Client) error {
				return c.DeleteProductMetafield(context.Background(), "prod_123", "mf_1")
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "bulk create product metafields",
			method: http.MethodPost,
			path:   "/products/prod_123/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkCreateProductMetafields(context.Background(), "prod_123", map[string]any{"items": []any{}})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "bulk update product metafields",
			method: http.MethodPut,
			path:   "/products/prod_123/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkUpdateProductMetafields(context.Background(), "prod_123", map[string]any{"items": []any{}})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "bulk delete product metafields",
			method: http.MethodDelete,
			path:   "/products/prod_123/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkDeleteProductMetafields(context.Background(), "prod_123", map[string]any{"ids": []any{}})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "list product app metafields",
			method: http.MethodGet,
			path:   "/products/prod_123/app_metafields",
			call: func(c *Client) error {
				_, err := c.ListProductAppMetafields(context.Background(), "prod_123")
				return err
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "get product app metafield",
			method: http.MethodGet,
			path:   "/products/prod_123/app_metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.GetProductAppMetafield(context.Background(), "prod_123", "mf_1")
				return err
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "create product app metafield",
			method: http.MethodPost,
			path:   "/products/prod_123/app_metafields",
			call: func(c *Client) error {
				_, err := c.CreateProductAppMetafield(context.Background(), "prod_123", map[string]any{"key": "k"})
				return err
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "update product app metafield",
			method: http.MethodPut,
			path:   "/products/prod_123/app_metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.UpdateProductAppMetafield(context.Background(), "prod_123", "mf_1", map[string]any{"value": "v"})
				return err
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "delete product app metafield",
			method: http.MethodDelete,
			path:   "/products/prod_123/app_metafields/mf_1",
			call: func(c *Client) error {
				return c.DeleteProductAppMetafield(context.Background(), "prod_123", "mf_1")
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "bulk create product app metafields",
			method: http.MethodPost,
			path:   "/products/prod_123/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkCreateProductAppMetafields(context.Background(), "prod_123", map[string]any{"items": []any{}})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "bulk update product app metafields",
			method: http.MethodPut,
			path:   "/products/prod_123/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkUpdateProductAppMetafields(context.Background(), "prod_123", map[string]any{"items": []any{}})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "bulk delete product app metafields",
			method: http.MethodDelete,
			path:   "/products/prod_123/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkDeleteProductAppMetafields(context.Background(), "prod_123", map[string]any{"ids": []any{}})
			},
			wantStatus: http.StatusNoContent,
		},
	}

	var current tc
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != current.method {
			t.Errorf("%s: expected %s, got %s", current.name, current.method, r.Method)
		}
		if r.URL.Path != current.path {
			t.Errorf("%s: expected path %s, got %s", current.name, current.path, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(current.wantStatus)
		if current.wantStatus == http.StatusOK {
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	for _, tt := range tests {
		current = tt
		if err := tt.call(client); err != nil {
			t.Fatalf("%s: call failed: %v", tt.name, err)
		}
	}
}
