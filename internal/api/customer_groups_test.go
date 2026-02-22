package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomerGroupsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer_groups" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CustomerGroupsListResponse{
			Items: []CustomerGroup{
				{ID: "grp_123", Name: "VIP", CustomerCount: 50},
				{ID: "grp_456", Name: "Wholesale", CustomerCount: 25},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	groups, err := client.ListCustomerGroups(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCustomerGroups failed: %v", err)
	}

	if len(groups.Items) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups.Items))
	}
	if groups.Items[0].ID != "grp_123" {
		t.Errorf("Unexpected group ID: %s", groups.Items[0].ID)
	}
}

func TestCustomerGroupsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer_groups/grp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		group := CustomerGroup{ID: "grp_123", Name: "VIP", CustomerCount: 50}
		_ = json.NewEncoder(w).Encode(group)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	group, err := client.GetCustomerGroup(context.Background(), "grp_123")
	if err != nil {
		t.Fatalf("GetCustomerGroup failed: %v", err)
	}

	if group.ID != "grp_123" {
		t.Errorf("Unexpected group ID: %s", group.ID)
	}
}

func TestGetCustomerGroupEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetCustomerGroup(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer group id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCustomerGroupsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer_groups" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		group := CustomerGroup{ID: "grp_new", Name: "New Group"}
		_ = json.NewEncoder(w).Encode(group)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerGroupCreateRequest{Name: "New Group"}
	group, err := client.CreateCustomerGroup(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCustomerGroup failed: %v", err)
	}

	if group.ID != "grp_new" {
		t.Errorf("Unexpected group ID: %s", group.ID)
	}
}

func TestCustomerGroupsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/customer_groups/grp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCustomerGroup(context.Background(), "grp_123")
	if err != nil {
		t.Fatalf("DeleteCustomerGroup failed: %v", err)
	}
}

func TestDeleteCustomerGroupEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteCustomerGroup(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer group id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCustomerGroupsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customer_groups/grp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		group := CustomerGroup{ID: "grp_123", Name: "Updated VIP"}
		_ = json.NewEncoder(w).Encode(group)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerGroupUpdateRequest{Name: "Updated VIP"}
	group, err := client.UpdateCustomerGroup(context.Background(), "grp_123", req)
	if err != nil {
		t.Fatalf("UpdateCustomerGroup failed: %v", err)
	}

	if group.ID != "grp_123" {
		t.Errorf("Unexpected group ID: %s", group.ID)
	}
	if group.Name != "Updated VIP" {
		t.Errorf("Unexpected group name: %s", group.Name)
	}
}

func TestUpdateCustomerGroupEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &CustomerGroupUpdateRequest{Name: "Test"}
			_, err := client.UpdateCustomerGroup(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer group id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCustomerGroupsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}
		if query.Get("sort_by") != "name" {
			t.Errorf("Expected sort_by=name, got %s", query.Get("sort_by"))
		}
		if query.Get("sort_order") != "asc" {
			t.Errorf("Expected sort_order=asc, got %s", query.Get("sort_order"))
		}

		resp := CustomerGroupsListResponse{
			Items:      []CustomerGroup{{ID: "grp_123", Name: "VIP"}},
			Page:       2,
			PageSize:   50,
			TotalCount: 100,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomerGroupsListOptions{
		Page:      2,
		PageSize:  50,
		SortBy:    "name",
		SortOrder: "asc",
	}
	groups, err := client.ListCustomerGroups(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListCustomerGroups failed: %v", err)
	}

	if len(groups.Items) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups.Items))
	}
}

func TestSearchCustomerGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer_groups/search" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("query") != "VIP" {
			t.Errorf("Expected query=VIP, got %s", query.Get("query"))
		}
		if query.Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "25" {
			t.Errorf("Expected page_size=25, got %s", query.Get("page_size"))
		}

		resp := CustomerGroupsListResponse{
			Items: []CustomerGroup{
				{ID: "grp_123", Name: "VIP Customers", CustomerCount: 100},
				{ID: "grp_456", Name: "VIP Gold", CustomerCount: 50},
			},
			Page:       1,
			PageSize:   25,
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomerGroupSearchOptions{
		Query:    "VIP",
		Page:     1,
		PageSize: 25,
	}
	groups, err := client.SearchCustomerGroups(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchCustomerGroups failed: %v", err)
	}

	if len(groups.Items) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups.Items))
	}
	if groups.Items[0].ID != "grp_123" {
		t.Errorf("Unexpected group ID: %s", groups.Items[0].ID)
	}
	if groups.Items[0].Name != "VIP Customers" {
		t.Errorf("Unexpected group name: %s", groups.Items[0].Name)
	}
}

func TestSearchCustomerGroupsEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CustomerGroupsListResponse{
			Items:      []CustomerGroup{},
			Page:       1,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomerGroupSearchOptions{Query: "nonexistent"}
	groups, err := client.SearchCustomerGroups(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchCustomerGroups failed: %v", err)
	}

	if len(groups.Items) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(groups.Items))
	}
}

func TestSearchCustomerGroupsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomerGroupSearchOptions{Query: "VIP"}
	_, err := client.SearchCustomerGroups(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
}

func TestSearchCustomerGroupsMinimalOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("query") != "test" {
			t.Errorf("Expected query=test, got %s", query.Get("query"))
		}
		// Page and page_size should not be present when zero
		if query.Get("page") != "" && query.Get("page") != "0" {
			t.Errorf("Expected no page param or 0, got %s", query.Get("page"))
		}

		resp := CustomerGroupsListResponse{
			Items:      []CustomerGroup{{ID: "grp_1", Name: "Test Group"}},
			Page:       1,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomerGroupSearchOptions{Query: "test"}
	groups, err := client.SearchCustomerGroups(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchCustomerGroups failed: %v", err)
	}

	if len(groups.Items) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups.Items))
	}
}

func TestGetCustomerGroupIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer_groups/grp_123/customer_ids" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CustomerGroupIDsResponse{
			CustomerIDs: []string{"cust_1", "cust_2", "cust_3", "cust_4", "cust_5"},
			TotalCount:  5,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	resp, err := client.GetCustomerGroupIDs(context.Background(), "grp_123")
	if err != nil {
		t.Fatalf("GetCustomerGroupIDs failed: %v", err)
	}

	if len(resp.CustomerIDs) != 5 {
		t.Errorf("Expected 5 customer IDs, got %d", len(resp.CustomerIDs))
	}
	if resp.TotalCount != 5 {
		t.Errorf("Expected total count 5, got %d", resp.TotalCount)
	}
	if resp.CustomerIDs[0] != "cust_1" {
		t.Errorf("Unexpected first customer ID: %s", resp.CustomerIDs[0])
	}
}

func TestGetCustomerGroupIDsEmptyGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CustomerGroupIDsResponse{
			CustomerIDs: []string{},
			TotalCount:  0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	resp, err := client.GetCustomerGroupIDs(context.Background(), "grp_empty")
	if err != nil {
		t.Fatalf("GetCustomerGroupIDs failed: %v", err)
	}

	if len(resp.CustomerIDs) != 0 {
		t.Errorf("Expected 0 customer IDs, got %d", len(resp.CustomerIDs))
	}
	if resp.TotalCount != 0 {
		t.Errorf("Expected total count 0, got %d", resp.TotalCount)
	}
}

func TestGetCustomerGroupIDsEmptyGroupID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name    string
		groupID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetCustomerGroupIDs(context.Background(), tc.groupID)
			if err == nil {
				t.Error("Expected error for empty group ID, got nil")
			}
			if err != nil && err.Error() != "customer group id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetCustomerGroupIDsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "customer group not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetCustomerGroupIDs(context.Background(), "grp_invalid")
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
}

func TestGetCustomerGroupIDsLargeGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a large group with many customer IDs
		customerIDs := make([]string, 100)
		for i := 0; i < 100; i++ {
			customerIDs[i] = "cust_" + string(rune('0'+i%10)) + string(rune('0'+i/10))
		}
		resp := CustomerGroupIDsResponse{
			CustomerIDs: customerIDs,
			TotalCount:  100,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	resp, err := client.GetCustomerGroupIDs(context.Background(), "grp_large")
	if err != nil {
		t.Fatalf("GetCustomerGroupIDs failed: %v", err)
	}

	if len(resp.CustomerIDs) != 100 {
		t.Errorf("Expected 100 customer IDs, got %d", len(resp.CustomerIDs))
	}
	if resp.TotalCount != 100 {
		t.Errorf("Expected total count 100, got %d", resp.TotalCount)
	}
}

func TestGetCustomerGroupChildren(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer_groups/grp_parent/customer_group_children" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "grp_child_1", "name": "Child 1"},
				{"id": "grp_child_2", "name": "Child 2"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetCustomerGroupChildren(context.Background(), "grp_parent")
	if err != nil {
		t.Fatalf("GetCustomerGroupChildren failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("failed to unmarshal raw response: %v", err)
	}
	if _, ok := got["items"]; !ok {
		t.Fatalf("expected items key in response, got %v", got)
	}
}

func TestGetCustomerGroupChildCustomerIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer_groups/grp_parent/customer_group_children/grp_child/customer_ids" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CustomerGroupIDsResponse{
			CustomerIDs: []string{"cust_1", "cust_2"},
			TotalCount:  2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	resp, err := client.GetCustomerGroupChildCustomerIDs(context.Background(), "grp_parent", "grp_child")
	if err != nil {
		t.Fatalf("GetCustomerGroupChildCustomerIDs failed: %v", err)
	}
	if len(resp.CustomerIDs) != 2 {
		t.Errorf("Expected 2 customer IDs, got %d", len(resp.CustomerIDs))
	}
	if resp.TotalCount != 2 {
		t.Errorf("Expected total count 2, got %d", resp.TotalCount)
	}
}

func TestGetCustomerGroupChildrenEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.GetCustomerGroupChildren(context.Background(), "   ")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestGetCustomerGroupChildCustomerIDsEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.GetCustomerGroupChildCustomerIDs(context.Background(), "grp_parent", " ")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}
