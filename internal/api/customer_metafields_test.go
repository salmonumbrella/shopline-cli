package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomerMetafieldsEndpoints(t *testing.T) {
	type tc struct {
		name   string
		method string
		path   string
		call   func(c *Client) error
	}

	raw := json.RawMessage(`{"foo":"bar"}`)

	tests := []tc{
		{
			name:   "ListCustomerMetafields",
			method: http.MethodGet,
			path:   "/customers/cust_1/metafields",
			call: func(c *Client) error {
				_, err := c.ListCustomerMetafields(context.Background(), "cust_1")
				return err
			},
		},
		{
			name:   "GetCustomerMetafield",
			method: http.MethodGet,
			path:   "/customers/cust_1/metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.GetCustomerMetafield(context.Background(), "cust_1", "mf_1")
				return err
			},
		},
		{
			name:   "CreateCustomerMetafield",
			method: http.MethodPost,
			path:   "/customers/cust_1/metafields",
			call: func(c *Client) error {
				_, err := c.CreateCustomerMetafield(context.Background(), "cust_1", raw)
				return err
			},
		},
		{
			name:   "UpdateCustomerMetafield",
			method: http.MethodPut,
			path:   "/customers/cust_1/metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.UpdateCustomerMetafield(context.Background(), "cust_1", "mf_1", raw)
				return err
			},
		},
		{
			name:   "DeleteCustomerMetafield",
			method: http.MethodDelete,
			path:   "/customers/cust_1/metafields/mf_1",
			call: func(c *Client) error {
				return c.DeleteCustomerMetafield(context.Background(), "cust_1", "mf_1")
			},
		},
		{
			name:   "BulkCreateCustomerMetafields",
			method: http.MethodPost,
			path:   "/customers/cust_1/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkCreateCustomerMetafields(context.Background(), "cust_1", raw)
			},
		},
		{
			name:   "BulkUpdateCustomerMetafields",
			method: http.MethodPut,
			path:   "/customers/cust_1/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkUpdateCustomerMetafields(context.Background(), "cust_1", raw)
			},
		},
		{
			name:   "BulkDeleteCustomerMetafields",
			method: http.MethodDelete,
			path:   "/customers/cust_1/metafields/bulk",
			call: func(c *Client) error {
				return c.BulkDeleteCustomerMetafields(context.Background(), "cust_1", raw)
			},
		},

		{
			name:   "ListCustomerAppMetafields",
			method: http.MethodGet,
			path:   "/customers/cust_1/app_metafields",
			call: func(c *Client) error {
				_, err := c.ListCustomerAppMetafields(context.Background(), "cust_1")
				return err
			},
		},
		{
			name:   "GetCustomerAppMetafield",
			method: http.MethodGet,
			path:   "/customers/cust_1/app_metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.GetCustomerAppMetafield(context.Background(), "cust_1", "mf_1")
				return err
			},
		},
		{
			name:   "CreateCustomerAppMetafield",
			method: http.MethodPost,
			path:   "/customers/cust_1/app_metafields",
			call: func(c *Client) error {
				_, err := c.CreateCustomerAppMetafield(context.Background(), "cust_1", raw)
				return err
			},
		},
		{
			name:   "UpdateCustomerAppMetafield",
			method: http.MethodPut,
			path:   "/customers/cust_1/app_metafields/mf_1",
			call: func(c *Client) error {
				_, err := c.UpdateCustomerAppMetafield(context.Background(), "cust_1", "mf_1", raw)
				return err
			},
		},
		{
			name:   "DeleteCustomerAppMetafield",
			method: http.MethodDelete,
			path:   "/customers/cust_1/app_metafields/mf_1",
			call: func(c *Client) error {
				return c.DeleteCustomerAppMetafield(context.Background(), "cust_1", "mf_1")
			},
		},
		{
			name:   "BulkCreateCustomerAppMetafields",
			method: http.MethodPost,
			path:   "/customers/cust_1/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkCreateCustomerAppMetafields(context.Background(), "cust_1", raw)
			},
		},
		{
			name:   "BulkUpdateCustomerAppMetafields",
			method: http.MethodPut,
			path:   "/customers/cust_1/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkUpdateCustomerAppMetafields(context.Background(), "cust_1", raw)
			},
		},
		{
			name:   "BulkDeleteCustomerAppMetafields",
			method: http.MethodDelete,
			path:   "/customers/cust_1/app_metafields/bulk",
			call: func(c *Client) error {
				return c.BulkDeleteCustomerAppMetafields(context.Background(), "cust_1", raw)
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
