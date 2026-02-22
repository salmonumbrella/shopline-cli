package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCouponsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/coupons" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CouponsListResponse{
			Items: []Coupon{
				{ID: "cpn_123", Code: "SAVE10", DiscountType: "percentage", DiscountValue: 10},
				{ID: "cpn_456", Code: "FLAT20", DiscountType: "fixed_amount", DiscountValue: 20},
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

	coupons, err := client.ListCoupons(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCoupons failed: %v", err)
	}

	if len(coupons.Items) != 2 {
		t.Errorf("Expected 2 coupons, got %d", len(coupons.Items))
	}
	if coupons.Items[0].Code != "SAVE10" {
		t.Errorf("Unexpected coupon code: %s", coupons.Items[0].Code)
	}
}

func TestCouponsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/coupons/cpn_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		coupon := Coupon{ID: "cpn_123", Code: "SAVE10", DiscountType: "percentage"}
		_ = json.NewEncoder(w).Encode(coupon)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	coupon, err := client.GetCoupon(context.Background(), "cpn_123")
	if err != nil {
		t.Fatalf("GetCoupon failed: %v", err)
	}

	if coupon.ID != "cpn_123" {
		t.Errorf("Unexpected coupon ID: %s", coupon.ID)
	}
}

func TestGetCouponEmptyID(t *testing.T) {
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
			_, err := client.GetCoupon(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "coupon id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetCouponByCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/coupons/code/SAVE10" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		coupon := Coupon{ID: "cpn_123", Code: "SAVE10"}
		_ = json.NewEncoder(w).Encode(coupon)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	coupon, err := client.GetCouponByCode(context.Background(), "SAVE10")
	if err != nil {
		t.Fatalf("GetCouponByCode failed: %v", err)
	}

	if coupon.Code != "SAVE10" {
		t.Errorf("Unexpected coupon code: %s", coupon.Code)
	}
}

func TestCouponsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		coupon := Coupon{ID: "cpn_new", Code: "NEWCODE"}
		_ = json.NewEncoder(w).Encode(coupon)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CouponCreateRequest{
		Code:          "NEWCODE",
		Title:         "New Coupon",
		DiscountType:  "percentage",
		DiscountValue: 15,
	}

	coupon, err := client.CreateCoupon(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCoupon failed: %v", err)
	}

	if coupon.Code != "NEWCODE" {
		t.Errorf("Unexpected coupon code: %s", coupon.Code)
	}
}

func TestCouponsActivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/coupons/cpn_123/activate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		coupon := Coupon{ID: "cpn_123", Status: "active"}
		_ = json.NewEncoder(w).Encode(coupon)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	coupon, err := client.ActivateCoupon(context.Background(), "cpn_123")
	if err != nil {
		t.Fatalf("ActivateCoupon failed: %v", err)
	}

	if coupon.Status != "active" {
		t.Errorf("Unexpected status: %s", coupon.Status)
	}
}

func TestCouponsDeactivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/coupons/cpn_123/deactivate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		coupon := Coupon{ID: "cpn_123", Status: "inactive"}
		_ = json.NewEncoder(w).Encode(coupon)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	coupon, err := client.DeactivateCoupon(context.Background(), "cpn_123")
	if err != nil {
		t.Fatalf("DeactivateCoupon failed: %v", err)
	}

	if coupon.Status != "inactive" {
		t.Errorf("Unexpected status: %s", coupon.Status)
	}
}

func TestCouponsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/coupons/cpn_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		coupon := Coupon{ID: "cpn_123", Code: "SAVE10", Title: "Updated Title"}
		_ = json.NewEncoder(w).Encode(coupon)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CouponUpdateRequest{Title: "Updated Title"}
	coupon, err := client.UpdateCoupon(context.Background(), "cpn_123", req)
	if err != nil {
		t.Fatalf("UpdateCoupon failed: %v", err)
	}

	if coupon.Title != "Updated Title" {
		t.Errorf("Unexpected title: %s", coupon.Title)
	}
}

func TestCouponsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/coupons/cpn_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCoupon(context.Background(), "cpn_123")
	if err != nil {
		t.Fatalf("DeleteCoupon failed: %v", err)
	}
}

func TestGetCouponByCodeEmptyCode(t *testing.T) {
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
			_, err := client.GetCouponByCode(context.Background(), tc.code)
			if err == nil {
				t.Error("Expected error for empty code, got nil")
			}
			if err != nil && err.Error() != "coupon code is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateCouponEmptyID(t *testing.T) {
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
			_, err := client.UpdateCoupon(context.Background(), tc.id, &CouponUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "coupon id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestActivateCouponEmptyID(t *testing.T) {
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
			_, err := client.ActivateCoupon(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "coupon id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeactivateCouponEmptyID(t *testing.T) {
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
			_, err := client.DeactivateCoupon(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "coupon id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteCouponEmptyID(t *testing.T) {
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
			err := client.DeleteCoupon(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "coupon id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCouponsListWithOptions(t *testing.T) {
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
		if query.Get("code") != "SAVE10" {
			t.Errorf("Expected code=SAVE10, got %s", query.Get("code"))
		}

		resp := CouponsListResponse{
			Items:      []Coupon{{ID: "cpn_123", Code: "SAVE10"}},
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

	opts := &CouponsListOptions{
		Page:     2,
		PageSize: 10,
		Status:   "active",
		Code:     "SAVE10",
	}
	coupons, err := client.ListCoupons(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListCoupons failed: %v", err)
	}

	if len(coupons.Items) != 1 {
		t.Errorf("Expected 1 coupon, got %d", len(coupons.Items))
	}
}
