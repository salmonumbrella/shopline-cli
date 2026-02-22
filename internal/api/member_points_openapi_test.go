package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMemberPointsOpenAPIEndpoints(t *testing.T) {
	type tc struct {
		name   string
		method string
		path   string
		call   func(c *Client) error
	}

	raw := json.RawMessage(`{"foo":"bar"}`)

	tests := []tc{
		{
			name:   "GetCustomersMembershipInfo",
			method: http.MethodGet,
			path:   "/customers/membership_info",
			call: func(c *Client) error {
				_, err := c.GetCustomersMembershipInfo(context.Background())
				return err
			},
		},
		{
			name:   "GetCustomerMemberPointsHistory",
			method: http.MethodGet,
			path:   "/customers/cust_1/member_points",
			call: func(c *Client) error {
				_, err := c.GetCustomerMemberPointsHistory(context.Background(), "cust_1")
				return err
			},
		},
		{
			name:   "UpdateCustomerMemberPoints",
			method: http.MethodPost,
			path:   "/customers/cust_1/member_points",
			call: func(c *Client) error {
				_, err := c.UpdateCustomerMemberPoints(context.Background(), "cust_1", raw)
				return err
			},
		},
		{
			name:   "GetCustomerMembershipTierActionLogs",
			method: http.MethodGet,
			path:   "/customers/cust_1/membership_tier/action_logs",
			call: func(c *Client) error {
				_, err := c.GetCustomerMembershipTierActionLogs(context.Background(), "cust_1")
				return err
			},
		},
		{
			name:   "ListMemberPointRules",
			method: http.MethodGet,
			path:   "/member_point_rules",
			call: func(c *Client) error {
				_, err := c.ListMemberPointRules(context.Background())
				return err
			},
		},
		{
			name:   "BulkUpdateMemberPoints",
			method: http.MethodPost,
			path:   "/member_points/bulk_update",
			call: func(c *Client) error {
				_, err := c.BulkUpdateMemberPoints(context.Background(), raw)
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
