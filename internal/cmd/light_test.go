package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

func TestToLightOrder(t *testing.T) {
	order := &api.Order{
		ID:            "order-1",
		OrderNumber:   "SL-1001",
		Status:        "open",
		PaymentStatus: "paid",
		FulfillStatus: "unfulfilled",
		TotalPrice:    "99.99",
		Currency:      "CAD",
		CustomerName:  "Test User",
		CustomerEmail: "test@example.com",
		CustomerID:    "cust-1",
		Note:          "Some note",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	light := toLightOrder(order)
	if light.ID != "order-1" {
		t.Errorf("ID = %q, want %q", light.ID, "order-1")
	}
	if light.CustomerName != "Test User" {
		t.Errorf("CustomerName = %q, want %q", light.CustomerName, "Test User")
	}

	// Verify light output has fewer JSON lines than full order
	lightJSON, _ := json.MarshalIndent(light, "", "  ")
	fullJSON, _ := json.MarshalIndent(order, "", "  ")
	lightSize := len(lightJSON)
	fullSize := len(fullJSON)
	if lightSize >= fullSize {
		t.Errorf("light JSON (%d bytes) should be smaller than full JSON (%d bytes)", lightSize, fullSize)
	}
}

func TestToLightProduct(t *testing.T) {
	product := &api.Product{
		ID:     "prod-1",
		Title:  "Test Product",
		Status: "active",
		Vendor: "Test Vendor",
		Handle: "test-product",
	}

	light := toLightProduct(product)
	if light.ID != "prod-1" {
		t.Errorf("ID = %q, want %q", light.ID, "prod-1")
	}
	if light.Title != "Test Product" {
		t.Errorf("Title = %q, want %q", light.Title, "Test Product")
	}
}

func TestToLightCustomer(t *testing.T) {
	customer := &api.Customer{
		ID:          "cust-1",
		Email:       "test@example.com",
		FirstName:   "Test",
		LastName:    "User",
		Phone:       "+1234567890",
		OrdersCount: 5,
		TotalSpent:  "500.00",
		Note:        "VIP customer",
		State:       "enabled",
	}

	light := toLightCustomer(customer)
	if light.OrdersCount != 5 {
		t.Errorf("OrdersCount = %d, want %d", light.OrdersCount, 5)
	}
	if light.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", light.Email, "test@example.com")
	}
}

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"short ASCII", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated ASCII", "hello world foo", 10, "hello w..."},
		{"empty string", "", 5, ""},
		{"CJK truncated", "一二三四五六七八九十", 7, "一二三四..."},
		{"CJK fits", "一二三四五", 5, "一二三四五"},
		{"mixed runes", "Hello世界Foo", 8, "Hello..."},
		{"max equals 3 long string", "hello", 3, "..."},
		{"max equals 3 short string", "hi", 3, "hi"},
		{"max equals 2 long string", "hello", 2, ".."},
		{"max equals 1 long string", "hello", 1, "."},
		{"max equals 0 long string", "hello", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateRunes(tt.input, tt.max)
			if got != tt.expected {
				t.Errorf("truncateRunes(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.expected)
			}
		})
	}
}

func TestToLightOrderTruncatesLongName(t *testing.T) {
	longName := "Bartholomew Jebediah Aloysius Featherington III of Canterbury"
	order := &api.Order{
		ID:           "order-1",
		OrderNumber:  "SL-1001",
		Status:       "open",
		Currency:     "CAD",
		CustomerName: longName,
	}
	light := toLightOrder(order)
	if len([]rune(light.CustomerName)) > 50 {
		t.Errorf("CustomerName rune length = %d, want <= 50", len([]rune(light.CustomerName)))
	}
	if light.CustomerName[len(light.CustomerName)-3:] != "..." {
		t.Errorf("CustomerName should end with '...', got %q", light.CustomerName)
	}
}

func TestToLightProductTruncatesLongTitle(t *testing.T) {
	longTitle := "Super Ultra Premium Deluxe Mega Extra Special Limited Edition Collector's Item Gold Plated Widget Set Version 2.0"
	product := &api.Product{
		ID:     "prod-1",
		Title:  longTitle,
		Status: "active",
		Vendor: "Test",
		Handle: "test",
	}
	light := toLightProduct(product)
	if len([]rune(light.Title)) > 80 {
		t.Errorf("Title rune length = %d, want <= 80", len([]rune(light.Title)))
	}
}

func TestLightCouponListShape(t *testing.T) {
	resp := api.ListResponse[lightCoupon]{
		Items: []lightCoupon{{ID: "1", Code: "TEST"}},
		Pagination: api.Pagination{
			CurrentPage: 1,
			PerPage:     20,
			TotalCount:  1,
		},
		Page:       1,
		PageSize:   20,
		TotalCount: 1,
		HasMore:    false,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"items", "pagination", "page", "page_size", "total_count", "has_more"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing key %q in light coupon list response", key)
		}
	}
}

func TestToLightSlice(t *testing.T) {
	customers := []api.Customer{
		{ID: "1", Email: "a@test.com", FirstName: "A"},
		{ID: "2", Email: "b@test.com", FirstName: "B"},
		{ID: "3", Email: "c@test.com", FirstName: "C"},
	}

	result := toLightSlice(customers, toLightCustomer)
	if len(result) != 3 {
		t.Fatalf("len = %d, want 3", len(result))
	}
	if result[0].ID != "1" {
		t.Errorf("result[0].ID = %q, want %q", result[0].ID, "1")
	}
	if result[2].Email != "c@test.com" {
		t.Errorf("result[2].Email = %q, want %q", result[2].Email, "c@test.com")
	}
}
