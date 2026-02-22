package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkListCustomers(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CustomersListResponse{
			Items:      make([]Customer, 20),
			Page:       1,
			PageSize:   20,
			TotalCount: 100,
			HasMore:    true,
		}
		for i := range resp.Items {
			resp.Items[i] = Customer{
				ID:        "cust_" + string(rune('a'+i)),
				Email:     "user@example.com",
				FirstName: "First",
				LastName:  "Last",
				State:     "enabled",
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ListCustomers(context.Background(), nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetCustomer(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customer := Customer{
			ID:               "cust_123",
			Email:            "user@example.com",
			FirstName:        "First",
			LastName:         "Last",
			Phone:            "+1234567890",
			AcceptsMarketing: true,
			OrdersCount:      10,
			TotalSpent:       "1500.00",
			Currency:         "USD",
			Tags:             []string{"vip", "loyal"},
			Note:             "Preferred customer",
			State:            "enabled",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		_ = json.NewEncoder(w).Encode(customer)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetCustomer(context.Background(), "cust_123")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateCustomer(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		customer := Customer{
			ID:        "cust_new",
			Email:     "new@example.com",
			FirstName: "New",
			LastName:  "Customer",
			State:     "enabled",
		}
		_ = json.NewEncoder(w).Encode(customer)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	reqBody := map[string]interface{}{
		"email":      "new@example.com",
		"first_name": "New",
		"last_name":  "Customer",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result Customer
		err := client.Post(context.Background(), "/customers", reqBody, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONParsing(b *testing.B) {
	customerJSON := []byte(`{
		"id": "cust_123",
		"email": "user@example.com",
		"first_name": "First",
		"last_name": "Last",
		"phone": "+1234567890",
		"accepts_marketing": true,
		"orders_count": 10,
		"total_spent": "1500.00",
		"currency": "USD",
		"tags": "vip,loyal",
		"note": "Preferred customer",
		"state": "enabled",
		"created_at": "2024-01-15T10:30:00Z",
		"updated_at": "2024-01-20T14:45:00Z"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var customer Customer
		if err := json.Unmarshal(customerJSON, &customer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewClient("shpat_test_token_12345")
	}
}

func BenchmarkListCustomersWithOptions(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CustomersListResponse{
			Items:      make([]Customer, 50),
			Page:       1,
			PageSize:   50,
			TotalCount: 500,
			HasMore:    true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	acceptsMarketing := true
	opts := &CustomersListOptions{
		Page:             1,
		PageSize:         50,
		State:            "enabled",
		Tags:             "vip",
		AcceptsMarketing: &acceptsMarketing,
		SortBy:           "created_at",
		SortOrder:        "desc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ListCustomers(context.Background(), opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONParsingListResponse(b *testing.B) {
	respJSON := []byte(`{
		"items": [
			{"id": "cust_1", "email": "user1@example.com", "first_name": "User", "last_name": "One", "state": "enabled"},
			{"id": "cust_2", "email": "user2@example.com", "first_name": "User", "last_name": "Two", "state": "enabled"},
			{"id": "cust_3", "email": "user3@example.com", "first_name": "User", "last_name": "Three", "state": "enabled"},
			{"id": "cust_4", "email": "user4@example.com", "first_name": "User", "last_name": "Four", "state": "enabled"},
			{"id": "cust_5", "email": "user5@example.com", "first_name": "User", "last_name": "Five", "state": "enabled"}
		],
		"page": 1,
		"page_size": 20,
		"total_count": 100,
		"has_more": true
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp CustomersListResponse
		if err := json.Unmarshal(respJSON, &resp); err != nil {
			b.Fatal(err)
		}
	}
}
