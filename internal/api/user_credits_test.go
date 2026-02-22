package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserCreditsEndpoints(t *testing.T) {
	type tc struct {
		name   string
		method string
		path   string
		call   func(c *Client) error
	}

	raw := json.RawMessage(`{"credits":[{"customer_id":"cust_1","amount":1}]}`)

	tests := []tc{
		{
			name:   "ListUserCredits",
			method: http.MethodGet,
			path:   "/user_credits",
			call: func(c *Client) error {
				_, err := c.ListUserCredits(context.Background(), nil)
				return err
			},
		},
		{
			name:   "BulkUpdateUserCredits",
			method: http.MethodPost,
			path:   "/user_credits/bulk_update",
			call: func(c *Client) error {
				_, err := c.BulkUpdateUserCredits(context.Background(), raw)
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
