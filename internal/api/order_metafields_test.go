package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderMetafieldsEndpoints(t *testing.T) {
	type tc struct {
		name   string
		method string
		path   string
		call   func(c *Client) error
	}

	raw := json.RawMessage(`{"foo":"bar"}`)

	tests := []tc{
		{
			name:   "ListOrderMetafields",
			method: http.MethodGet,
			path:   "/orders/ord_1/metafields",
			call: func(c *Client) error {
				_, err := c.ListOrderMetafields(context.Background(), "ord_1")
				return err
			},
		},
		{
			name:   "GetOrderMetafield",
			method: http.MethodGet,
			path:   "/orders/ord_1/metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.GetOrderMetafield(context.Background(), "ord_1", "mf_1")
				return err
			},
		},
		{
			name:   "CreateOrderMetafield",
			method: http.MethodPost,
			path:   "/orders/ord_1/metafields",
			call: func(c *Client) error {
				_, err := c.CreateOrderMetafield(context.Background(), "ord_1", raw)
				return err
			},
		},
		{
			name:   "UpdateOrderMetafield",
			method: http.MethodPut,
			path:   "/orders/ord_1/metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.UpdateOrderMetafield(context.Background(), "ord_1", "mf_1", raw)
				return err
			},
		},
		{
			name:   "DeleteOrderMetafield",
			method: http.MethodDelete,
			path:   "/orders/ord_1/metafields/mf_1",
			call: func(c *Client) error {
				return c.DeleteOrderMetafield(context.Background(), "ord_1", "mf_1")
			},
		},
		{
			name:   "BulkCreateOrderMetafields",
			method: http.MethodPost,
			path:   "/orders/ord_1/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkCreateOrderMetafields(context.Background(), "ord_1", raw)
			},
		},
		{
			name:   "BulkUpdateOrderMetafields",
			method: http.MethodPut,
			path:   "/orders/ord_1/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkUpdateOrderMetafields(context.Background(), "ord_1", raw)
			},
		},
		{
			name:   "BulkDeleteOrderMetafields",
			method: http.MethodDelete,
			path:   "/orders/ord_1/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkDeleteOrderMetafields(context.Background(), "ord_1", raw)
			},
		},

		{
			name:   "ListOrderAppMetafields",
			method: http.MethodGet,
			path:   "/orders/ord_1/app_metafields",
			call: func(c *Client) error {
				_, err := c.ListOrderAppMetafields(context.Background(), "ord_1")
				return err
			},
		},
		{
			name:   "GetOrderAppMetafield",
			method: http.MethodGet,
			path:   "/orders/ord_1/app_metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.GetOrderAppMetafield(context.Background(), "ord_1", "mf_1")
				return err
			},
		},
		{
			name:   "CreateOrderAppMetafield",
			method: http.MethodPost,
			path:   "/orders/ord_1/app_metafields",
			call: func(c *Client) error {
				_, err := c.CreateOrderAppMetafield(context.Background(), "ord_1", raw)
				return err
			},
		},
		{
			name:   "UpdateOrderAppMetafield",
			method: http.MethodPut,
			path:   "/orders/ord_1/app_metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.UpdateOrderAppMetafield(context.Background(), "ord_1", "mf_1", raw)
				return err
			},
		},
		{
			name:   "DeleteOrderAppMetafield",
			method: http.MethodDelete,
			path:   "/orders/ord_1/app_metafields/mf_1",
			call: func(c *Client) error {
				return c.DeleteOrderAppMetafield(context.Background(), "ord_1", "mf_1")
			},
		},
		{
			name:   "BulkCreateOrderAppMetafields",
			method: http.MethodPost,
			path:   "/orders/ord_1/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkCreateOrderAppMetafields(context.Background(), "ord_1", raw)
			},
		},
		{
			name:   "BulkUpdateOrderAppMetafields",
			method: http.MethodPut,
			path:   "/orders/ord_1/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkUpdateOrderAppMetafields(context.Background(), "ord_1", raw)
			},
		},
		{
			name:   "BulkDeleteOrderAppMetafields",
			method: http.MethodDelete,
			path:   "/orders/ord_1/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkDeleteOrderAppMetafields(context.Background(), "ord_1", raw)
			},
		},

		{
			name:   "ListOrderItemMetafields",
			method: http.MethodGet,
			path:   "/orders/ord_1/items/metafields",
			call: func(c *Client) error {
				_, err := c.ListOrderItemMetafields(context.Background(), "ord_1")
				return err
			},
		},
		{
			name:   "BulkCreateOrderItemMetafields",
			method: http.MethodPost,
			path:   "/orders/ord_1/items/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkCreateOrderItemMetafields(context.Background(), "ord_1", raw)
			},
		},
		{
			name:   "BulkUpdateOrderItemMetafields",
			method: http.MethodPut,
			path:   "/orders/ord_1/items/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkUpdateOrderItemMetafields(context.Background(), "ord_1", raw)
			},
		},
		{
			name:   "BulkDeleteOrderItemMetafields",
			method: http.MethodDelete,
			path:   "/orders/ord_1/items/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkDeleteOrderItemMetafields(context.Background(), "ord_1", raw)
			},
		},

		{
			name:   "ListOrderItemAppMetafields",
			method: http.MethodGet,
			path:   "/orders/ord_1/items/app_metafields",
			call: func(c *Client) error {
				_, err := c.ListOrderItemAppMetafields(context.Background(), "ord_1")
				return err
			},
		},
		{
			name:   "BulkCreateOrderItemAppMetafields",
			method: http.MethodPost,
			path:   "/orders/ord_1/items/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkCreateOrderItemAppMetafields(context.Background(), "ord_1", raw)
			},
		},
		{
			name:   "BulkUpdateOrderItemAppMetafields",
			method: http.MethodPut,
			path:   "/orders/ord_1/items/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkUpdateOrderItemAppMetafields(context.Background(), "ord_1", raw)
			},
		},
		{
			name:   "BulkDeleteOrderItemAppMetafields",
			method: http.MethodDelete,
			path:   "/orders/ord_1/items/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkDeleteOrderItemAppMetafields(context.Background(), "ord_1", raw)
			},
		},
	}

	want := make(map[string]struct{}, len(tests))
	for _, tt := range tests {
		want[tt.method+" "+tt.path] = struct{}{}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		if _, ok := want[key]; !ok {
			t.Fatalf("unexpected request: %s", key)
		}
		delete(want, key)

		// for GETs and mutation endpoints that return JSON, respond with an empty object.
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(client); err != nil {
				t.Fatalf("call failed: %v", err)
			}
		})
	}

	if len(want) != 0 {
		t.Fatalf("missing requests: %v", want)
	}
}
