package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomerStoreCreditsEndpoints(t *testing.T) {
	type tc struct {
		name   string
		method string
		path   string
		call   func(c *Client) error
	}

	raw := json.RawMessage(`{"amount":1}`)

	tests := []tc{
		{
			name:   "GetCustomerStoreCredits",
			method: http.MethodGet,
			path:   "/customers/cust_1/store_credits",
			call: func(c *Client) error {
				_, err := c.GetCustomerStoreCredits(context.Background(), "cust_1")
				return err
			},
		},
		{
			name:   "CreateCustomerStoreCredits",
			method: http.MethodPost,
			path:   "/customers/cust_1/store_credits",
			call: func(c *Client) error {
				_, err := c.CreateCustomerStoreCredits(context.Background(), "cust_1", raw)
				return err
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

	client := NewClient("test", "token")
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
