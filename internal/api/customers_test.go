package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CustomersListResponse{
			Items: []Customer{
				{ID: "cust_123", Email: "alice@example.com", FirstName: "Alice", LastName: "Smith", State: "enabled"},
				{ID: "cust_456", Email: "bob@example.com", FirstName: "Bob", LastName: "Jones", State: "enabled"},
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

	customers, err := client.ListCustomers(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCustomers failed: %v", err)
	}

	if len(customers.Items) != 2 {
		t.Errorf("Expected 2 customers, got %d", len(customers.Items))
	}
	if customers.Items[0].ID != "cust_123" {
		t.Errorf("Unexpected customer ID: %s", customers.Items[0].ID)
	}
}

func TestCustomersGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customers/cust_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		creditBalance := 42.5
		customer := Customer{
			ID:            "cust_123",
			Email:         "alice@example.com",
			FirstName:     "Alice",
			LastName:      "Smith",
			State:         "enabled",
			CreditBalance: &creditBalance,
			Subscriptions: []CustomerSubscription{
				{Platform: "email", IsActive: true},
				{Platform: "sms", IsActive: false},
			},
		}
		_ = json.NewEncoder(w).Encode(customer)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	customer, err := client.GetCustomer(context.Background(), "cust_123")
	if err != nil {
		t.Fatalf("GetCustomer failed: %v", err)
	}

	if customer.ID != "cust_123" {
		t.Errorf("Unexpected customer ID: %s", customer.ID)
	}
	if customer.CreditBalance == nil || *customer.CreditBalance != 42.5 {
		t.Errorf("Unexpected credit balance: %v", customer.CreditBalance)
	}
	if len(customer.Subscriptions) != 2 {
		t.Errorf("Unexpected subscription count: %d", len(customer.Subscriptions))
	}
}

func TestGetCustomerEmptyID(t *testing.T) {
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
			_, err := client.GetCustomer(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreateCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customers" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CustomerCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Email != "newcustomer@example.com" {
			t.Errorf("Unexpected email: %s", req.Email)
		}
		if req.FirstName != "New" {
			t.Errorf("Unexpected first name: %s", req.FirstName)
		}
		if req.LastName != "Customer" {
			t.Errorf("Unexpected last name: %s", req.LastName)
		}

		customer := Customer{
			ID:        "cust_new",
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Phone:     req.Phone,
			Tags:      req.Tags,
			Note:      req.Note,
			State:     "enabled",
		}
		_ = json.NewEncoder(w).Encode(customer)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerCreateRequest{
		Email:     "newcustomer@example.com",
		FirstName: "New",
		LastName:  "Customer",
		Phone:     "+1234567890",
		Tags:      []string{"vip", "new"},
		Note:      "Test customer",
	}

	customer, err := client.CreateCustomer(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}

	if customer.ID != "cust_new" {
		t.Errorf("Unexpected customer ID: %s", customer.ID)
	}
	if customer.Email != "newcustomer@example.com" {
		t.Errorf("Unexpected email: %s", customer.Email)
	}
}

func TestCreateCustomerAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid email"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerCreateRequest{
		Email: "invalid-email",
	}

	_, err := client.CreateCustomer(context.Background(), req)
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestUpdateCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CustomerUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.FirstName == nil || *req.FirstName != "Updated" {
			t.Errorf("Unexpected first name: %v", req.FirstName)
		}

		customer := Customer{
			ID:        "cust_123",
			Email:     "alice@example.com",
			FirstName: *req.FirstName,
			LastName:  "Smith",
			State:     "enabled",
		}
		_ = json.NewEncoder(w).Encode(customer)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	firstName := "Updated"
	req := &CustomerUpdateRequest{
		FirstName: &firstName,
	}

	customer, err := client.UpdateCustomer(context.Background(), "cust_123", req)
	if err != nil {
		t.Fatalf("UpdateCustomer failed: %v", err)
	}

	if customer.FirstName != "Updated" {
		t.Errorf("Unexpected first name: %s", customer.FirstName)
	}
}

func TestUpdateCustomerEmptyID(t *testing.T) {
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
			_, err := client.UpdateCustomer(context.Background(), tc.id, &CustomerUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateCustomerAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "customer not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	firstName := "Updated"
	req := &CustomerUpdateRequest{
		FirstName: &firstName,
	}

	_, err := client.UpdateCustomer(context.Background(), "nonexistent", req)
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestDeleteCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCustomer(context.Background(), "cust_123")
	if err != nil {
		t.Fatalf("DeleteCustomer failed: %v", err)
	}
}

func TestDeleteCustomerEmptyID(t *testing.T) {
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
			err := client.DeleteCustomer(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteCustomerAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "customer not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCustomer(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestSearchCustomers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers/search" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("query") != "alice" {
			t.Errorf("Unexpected query param: %s", query.Get("query"))
		}
		if query.Get("page") != "1" {
			t.Errorf("Unexpected page param: %s", query.Get("page"))
		}
		if query.Get("page_size") != "10" {
			t.Errorf("Unexpected page_size param: %s", query.Get("page_size"))
		}

		resp := CustomersListResponse{
			Items: []Customer{
				{ID: "cust_123", Email: "alice@example.com", FirstName: "Alice", LastName: "Smith", State: "enabled"},
			},
			Page:       1,
			PageSize:   10,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomerSearchOptions{
		Query:    "alice",
		Page:     1,
		PageSize: 10,
	}

	customers, err := client.SearchCustomers(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchCustomers failed: %v", err)
	}

	if len(customers.Items) != 1 {
		t.Errorf("Expected 1 customer, got %d", len(customers.Items))
	}
	if customers.Items[0].Email != "alice@example.com" {
		t.Errorf("Unexpected email: %s", customers.Items[0].Email)
	}
}

func TestSearchCustomersWithEmailAndPhone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("email") != "alice@example.com" {
			t.Errorf("Unexpected email param: %s", query.Get("email"))
		}
		if query.Get("phone") != "+1234567890" {
			t.Errorf("Unexpected phone param: %s", query.Get("phone"))
		}

		resp := CustomersListResponse{
			Items:      []Customer{},
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

	opts := &CustomerSearchOptions{
		Email: "alice@example.com",
		Phone: "+1234567890",
	}

	_, err := client.SearchCustomers(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchCustomers failed: %v", err)
	}
}

func TestSearchCustomersAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomerSearchOptions{
		Query: "test",
	}

	_, err := client.SearchCustomers(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestSearchCustomersNilOptions(t *testing.T) {
	client := NewClient("token")
	_, err := client.SearchCustomers(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}
	if err.Error() != "search options are required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCustomerTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CustomerTagsUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.Add) != 2 || req.Add[0] != "vip" || req.Add[1] != "loyal" {
			t.Errorf("Unexpected add tags: %v", req.Add)
		}
		if len(req.Remove) != 1 || req.Remove[0] != "new" {
			t.Errorf("Unexpected remove tags: %v", req.Remove)
		}

		customer := Customer{
			ID:    "cust_123",
			Email: "alice@example.com",
			Tags:  []string{"vip", "loyal"},
			State: "enabled",
		}
		_ = json.NewEncoder(w).Encode(customer)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerTagsUpdateRequest{
		Add:    []string{"vip", "loyal"},
		Remove: []string{"new"},
	}

	customer, err := client.UpdateCustomerTags(context.Background(), "cust_123", req)
	if err != nil {
		t.Fatalf("UpdateCustomerTags failed: %v", err)
	}

	if len(customer.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(customer.Tags))
	}
}

func TestUpdateCustomerSubscriptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/subscriptions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if body["kind"] != "test" {
			t.Errorf("Unexpected request body: %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateCustomerSubscriptions(context.Background(), "cust_123", map[string]any{"kind": "test"})
	if err != nil {
		t.Fatalf("UpdateCustomerSubscriptions failed: %v", err)
	}
}

func TestGetLineCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers/line/line_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Customer{ID: "cust_1", Email: "line@example.com"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	c, err := client.GetLineCustomer(context.Background(), "line_123")
	if err != nil {
		t.Fatalf("GetLineCustomer failed: %v", err)
	}
	if c.ID != "cust_1" {
		t.Fatalf("unexpected id: %s", c.ID)
	}
}

func TestUpdateCustomerTagsEmptyID(t *testing.T) {
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
			_, err := client.UpdateCustomerTags(context.Background(), tc.id, &CustomerTagsUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateCustomerTagsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "customer not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerTagsUpdateRequest{
		Add: []string{"vip"},
	}

	_, err := client.UpdateCustomerTags(context.Background(), "nonexistent", req)
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestSetCustomerTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req struct {
			Tags []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.Tags) != 3 || req.Tags[0] != "premium" || req.Tags[1] != "loyal" || req.Tags[2] != "subscriber" {
			t.Errorf("Unexpected tags: %v", req.Tags)
		}

		customer := Customer{
			ID:    "cust_123",
			Email: "alice@example.com",
			Tags:  req.Tags,
			State: "enabled",
		}
		_ = json.NewEncoder(w).Encode(customer)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	tags := []string{"premium", "loyal", "subscriber"}
	customer, err := client.SetCustomerTags(context.Background(), "cust_123", tags)
	if err != nil {
		t.Fatalf("SetCustomerTags failed: %v", err)
	}

	if len(customer.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(customer.Tags))
	}
}

func TestSetCustomerTagsEmptyTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Tags []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Tags == nil || len(req.Tags) != 0 {
			t.Errorf("Expected empty tags array, got: %v", req.Tags)
		}

		customer := Customer{
			ID:    "cust_123",
			Email: "alice@example.com",
			Tags:  []string{},
			State: "enabled",
		}
		_ = json.NewEncoder(w).Encode(customer)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	customer, err := client.SetCustomerTags(context.Background(), "cust_123", []string{})
	if err != nil {
		t.Fatalf("SetCustomerTags failed: %v", err)
	}

	if len(customer.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(customer.Tags))
	}
}

func TestSetCustomerTagsEmptyID(t *testing.T) {
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
			_, err := client.SetCustomerTags(context.Background(), tc.id, []string{"tag"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSetCustomerTagsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "customer not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.SetCustomerTags(context.Background(), "nonexistent", []string{"tag"})
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestUpdateCustomerStoreCreditsViaStoreCreditsAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/store_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req StoreCreditUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Value != 50 {
			t.Errorf("Unexpected value: %d", req.Value)
		}
		if req.Remarks != "Loyalty bonus" {
			t.Errorf("Unexpected remarks: %s", req.Remarks)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"sc_new","value":50}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StoreCreditUpdateRequest{Value: 50, Remarks: "Loyalty bonus", Type: "manual_credit"}
	resp, err := client.UpdateCustomerStoreCredits(context.Background(), "cust_123", req)
	if err != nil {
		t.Fatalf("UpdateCustomerStoreCredits failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
}

func TestUpdateCustomerStoreCreditsNegativeValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/store_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var req StoreCreditUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Value != -25 {
			t.Errorf("Unexpected value: %d", req.Value)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"sc_deduct","value":-25}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StoreCreditUpdateRequest{Value: -25, Remarks: "Refund deduction", Type: "manual_credit"}
	_, err := client.UpdateCustomerStoreCredits(context.Background(), "cust_123", req)
	if err != nil {
		t.Fatalf("UpdateCustomerStoreCredits failed: %v", err)
	}
}

func TestUpdateCustomerStoreCreditsEmptyIDViaCustomers(t *testing.T) {
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
			req := &StoreCreditUpdateRequest{Value: 10, Remarks: "test"}
			_, err := client.UpdateCustomerStoreCredits(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateCustomerStoreCreditsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "insufficient credits"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StoreCreditUpdateRequest{Value: -1000, Remarks: "Large deduction", Type: "manual_credit"}
	_, err := client.UpdateCustomerStoreCredits(context.Background(), "cust_123", req)
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestGetCustomerPromotions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/promotions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CustomerPromotionsResponse{
			Items: []interface{}{
				map[string]interface{}{
					"id":       "promo_1",
					"name":     "Summer Sale",
					"discount": 20,
				},
				map[string]interface{}{
					"id":       "promo_2",
					"name":     "Loyalty Reward",
					"discount": 10,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	promotions, err := client.GetCustomerPromotions(context.Background(), "cust_123")
	if err != nil {
		t.Fatalf("GetCustomerPromotions failed: %v", err)
	}

	if len(promotions.Items) != 2 {
		t.Errorf("Expected 2 promotions, got %d", len(promotions.Items))
	}
}

func TestGetCustomerPromotionsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CustomerPromotionsResponse{
			Items: []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	promotions, err := client.GetCustomerPromotions(context.Background(), "cust_123")
	if err != nil {
		t.Fatalf("GetCustomerPromotions failed: %v", err)
	}

	if len(promotions.Items) != 0 {
		t.Errorf("Expected 0 promotions, got %d", len(promotions.Items))
	}
}

func TestGetCustomerPromotionsEmptyID(t *testing.T) {
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
			_, err := client.GetCustomerPromotions(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetCustomerPromotionsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "customer not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetCustomerPromotions(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestGetCustomerCouponPromotions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/coupon_promotions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "promo_coupon_1"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetCustomerCouponPromotions(context.Background(), "cust_123")
	if err != nil {
		t.Fatalf("GetCustomerCouponPromotions failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("failed to unmarshal raw response: %v", err)
	}
	if _, ok := got["items"]; !ok {
		t.Fatalf("expected items key in response, got %v", got)
	}
}

func TestGetCustomerCouponPromotionsEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.GetCustomerCouponPromotions(context.Background(), " ")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "customer id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCustomerCouponPromotionsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "customer not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetCustomerCouponPromotions(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}
