package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPromotionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/promotions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PromotionsListResponse{
			Items: []Promotion{
				{ID: "promo_123", Title: "Summer Sale", Status: "active", DiscountValue: 20},
				{ID: "promo_456", Title: "Black Friday", Status: "scheduled", DiscountValue: 30},
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

	promotions, err := client.ListPromotions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPromotions failed: %v", err)
	}

	if len(promotions.Items) != 2 {
		t.Errorf("Expected 2 promotions, got %d", len(promotions.Items))
	}
	if promotions.Items[0].ID != "promo_123" {
		t.Errorf("Unexpected promotion ID: %s", promotions.Items[0].ID)
	}
}

func TestPromotionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/promotions/promo_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		promotion := Promotion{ID: "promo_123", Title: "Summer Sale", Status: "active"}
		_ = json.NewEncoder(w).Encode(promotion)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	promotion, err := client.GetPromotion(context.Background(), "promo_123")
	if err != nil {
		t.Fatalf("GetPromotion failed: %v", err)
	}

	if promotion.ID != "promo_123" {
		t.Errorf("Unexpected promotion ID: %s", promotion.ID)
	}
}

func TestGetPromotionEmptyID(t *testing.T) {
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
			_, err := client.GetPromotion(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "promotion id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestPromotionsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/promotions/promo_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeletePromotion(context.Background(), "promo_123")
	if err != nil {
		t.Fatalf("DeletePromotion failed: %v", err)
	}
}

func TestDeletePromotionEmptyID(t *testing.T) {
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
			err := client.DeletePromotion(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "promotion id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestPromotionsActivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/promotions/promo_123/activate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		promotion := Promotion{ID: "promo_123", Title: "Summer Sale", Status: "active"}
		_ = json.NewEncoder(w).Encode(promotion)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	promotion, err := client.ActivatePromotion(context.Background(), "promo_123")
	if err != nil {
		t.Fatalf("ActivatePromotion failed: %v", err)
	}

	if promotion.Status != "active" {
		t.Errorf("Expected active status, got: %s", promotion.Status)
	}
}

func TestPromotionsDeactivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/promotions/promo_123/deactivate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		promotion := Promotion{ID: "promo_123", Title: "Summer Sale", Status: "inactive"}
		_ = json.NewEncoder(w).Encode(promotion)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	promotion, err := client.DeactivatePromotion(context.Background(), "promo_123")
	if err != nil {
		t.Fatalf("DeactivatePromotion failed: %v", err)
	}

	if promotion.Status != "inactive" {
		t.Errorf("Expected inactive status, got: %s", promotion.Status)
	}
}

func TestPromotionsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/promotions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		promotion := Promotion{ID: "promo_new", Title: "New Promotion", Status: "draft"}
		_ = json.NewEncoder(w).Encode(promotion)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &PromotionCreateRequest{
		Title:         "New Promotion",
		Type:          "percentage",
		DiscountType:  "percentage",
		DiscountValue: 15,
	}

	promotion, err := client.CreatePromotion(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePromotion failed: %v", err)
	}

	if promotion.ID != "promo_new" {
		t.Errorf("Unexpected promotion ID: %s", promotion.ID)
	}
}

func TestActivatePromotionEmptyID(t *testing.T) {
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
			_, err := client.ActivatePromotion(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "promotion id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeactivatePromotionEmptyID(t *testing.T) {
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
			_, err := client.DeactivatePromotion(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "promotion id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestPromotionsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", query.Get("page_size"))
		}
		if query.Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", query.Get("status"))
		}
		if query.Get("type") != "percentage" {
			t.Errorf("Expected type=percentage, got %s", query.Get("type"))
		}

		resp := PromotionsListResponse{
			Items:      []Promotion{{ID: "promo_123", Title: "Summer Sale"}},
			Page:       2,
			PageSize:   10,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &PromotionsListOptions{
		Page:     2,
		PageSize: 10,
		Status:   "active",
		Type:     "percentage",
	}
	promotions, err := client.ListPromotions(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListPromotions failed: %v", err)
	}

	if len(promotions.Items) != 1 {
		t.Errorf("Expected 1 promotion, got %d", len(promotions.Items))
	}
}

func TestSearchPromotions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/promotions/search" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("query") != "summer" {
			t.Errorf("Expected query=summer, got %s", query.Get("query"))
		}
		if query.Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", query.Get("status"))
		}
		if query.Get("type") != "percentage" {
			t.Errorf("Expected type=percentage, got %s", query.Get("type"))
		}
		if query.Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "25" {
			t.Errorf("Expected page_size=25, got %s", query.Get("page_size"))
		}

		resp := PromotionsListResponse{
			Items: []Promotion{
				{ID: "promo_123", Title: "Summer Sale", Status: "active", DiscountValue: 20},
			},
			Page:       1,
			PageSize:   25,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &PromotionSearchOptions{
		Query:    "summer",
		Status:   "active",
		Type:     "percentage",
		Page:     1,
		PageSize: 25,
	}
	promotions, err := client.SearchPromotions(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchPromotions failed: %v", err)
	}

	if len(promotions.Items) != 1 {
		t.Errorf("Expected 1 promotion, got %d", len(promotions.Items))
	}
	if promotions.Items[0].ID != "promo_123" {
		t.Errorf("Unexpected promotion ID: %s", promotions.Items[0].ID)
	}
}

func TestSearchPromotionsEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PromotionsListResponse{
			Items:      []Promotion{},
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

	opts := &PromotionSearchOptions{Query: "nonexistent"}
	promotions, err := client.SearchPromotions(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchPromotions failed: %v", err)
	}

	if len(promotions.Items) != 0 {
		t.Errorf("Expected 0 promotions, got %d", len(promotions.Items))
	}
}

func TestSearchPromotionsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &PromotionSearchOptions{Query: "summer"}
	_, err := client.SearchPromotions(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
}

func TestSendCoupon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/promotions/send-coupon" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CouponSendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.PromotionID != "promo_123" {
			t.Errorf("Unexpected promotion ID: %s", req.PromotionID)
		}
		if len(req.CustomerIDs) != 2 {
			t.Errorf("Expected 2 customer IDs, got %d", len(req.CustomerIDs))
		}
		if req.CustomerIDs[0] != "cust_1" || req.CustomerIDs[1] != "cust_2" {
			t.Errorf("Unexpected customer IDs: %v", req.CustomerIDs)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.SendCoupon(context.Background(), "promo_123", []string{"cust_1", "cust_2"})
	if err != nil {
		t.Fatalf("SendCoupon failed: %v", err)
	}
}

func TestSendCouponEmptyPromotionID(t *testing.T) {
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
			err := client.SendCoupon(context.Background(), tc.id, []string{"cust_1"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "promotion id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSendCouponEmptyCustomerIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name        string
		customerIDs []string
	}{
		{"nil slice", nil},
		{"empty slice", []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.SendCoupon(context.Background(), "promo_123", tc.customerIDs)
			if err == nil {
				t.Error("Expected error for empty customer IDs, got nil")
			}
			if err != nil && err.Error() != "at least one customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSendCouponAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid promotion"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.SendCoupon(context.Background(), "promo_123", []string{"cust_1"})
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
}

func TestRedeemCoupon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/promotions/redeem-coupon" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CouponRedeemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Code != "SUMMER20" {
			t.Errorf("Unexpected code: %s", req.Code)
		}
		if req.CustomerID != "cust_123" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}
		if req.OrderID != "order_456" {
			t.Errorf("Unexpected order ID: %s", req.OrderID)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.RedeemCoupon(context.Background(), "SUMMER20", "cust_123", "order_456")
	if err != nil {
		t.Fatalf("RedeemCoupon failed: %v", err)
	}
}

func TestRedeemCouponWithoutOrderID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CouponRedeemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.OrderID != "" {
			t.Errorf("Expected empty order ID, got: %s", req.OrderID)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.RedeemCoupon(context.Background(), "SUMMER20", "cust_123", "")
	if err != nil {
		t.Fatalf("RedeemCoupon failed: %v", err)
	}
}

func TestRedeemCouponEmptyCode(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		code string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.RedeemCoupon(context.Background(), tc.code, "cust_123", "order_456")
			if err == nil {
				t.Error("Expected error for empty code, got nil")
			}
			if err != nil && err.Error() != "coupon code is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestRedeemCouponEmptyCustomerID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name       string
		customerID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.RedeemCoupon(context.Background(), "SUMMER20", tc.customerID, "order_456")
			if err == nil {
				t.Error("Expected error for empty customer ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestRedeemCouponAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "coupon already redeemed"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.RedeemCoupon(context.Background(), "SUMMER20", "cust_123", "order_456")
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
}

func TestClaimCoupon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/promotions/claim-coupon" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CouponClaimRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Code != "WELCOME10" {
			t.Errorf("Unexpected code: %s", req.Code)
		}
		if req.CustomerID != "cust_789" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.ClaimCoupon(context.Background(), "WELCOME10", "cust_789")
	if err != nil {
		t.Fatalf("ClaimCoupon failed: %v", err)
	}
}

func TestClaimCouponEmptyCode(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		code string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.ClaimCoupon(context.Background(), tc.code, "cust_123")
			if err == nil {
				t.Error("Expected error for empty code, got nil")
			}
			if err != nil && err.Error() != "coupon code is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestClaimCouponEmptyCustomerID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name       string
		customerID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.ClaimCoupon(context.Background(), "WELCOME10", tc.customerID)
			if err == nil {
				t.Error("Expected error for empty customer ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestClaimCouponAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "coupon not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.ClaimCoupon(context.Background(), "INVALID", "cust_123")
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
}
