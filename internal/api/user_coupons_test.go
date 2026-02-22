package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserCouponsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/user_coupons" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		// Verify promotion_id is sent
		if r.URL.Query().Get("promotion_id") != "promo_123" {
			t.Errorf("Expected promotion_id=promo_123, got %s", r.URL.Query().Get("promotion_id"))
		}

		resp := UserCouponsListResponse{
			Items: []UserCoupon{
				{ID: "uc_123", UserID: "user_1", CouponCode: "SAVE10", Status: "active"},
				{ID: "uc_456", UserID: "user_2", CouponCode: "FLAT20", Status: "used"},
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

	// PromotionID is required
	userCoupons, err := client.ListUserCoupons(context.Background(), &UserCouponsListOptions{
		PromotionID: "promo_123",
	})
	if err != nil {
		t.Fatalf("ListUserCoupons failed: %v", err)
	}

	if len(userCoupons.Items) != 2 {
		t.Errorf("Expected 2 user coupons, got %d", len(userCoupons.Items))
	}
	if userCoupons.Items[0].ID != "uc_123" {
		t.Errorf("Unexpected user coupon ID: %s", userCoupons.Items[0].ID)
	}
}

func TestUserCouponsListRequiresPromotionID(t *testing.T) {
	client := NewClient("token")

	// Test with nil options
	_, err := client.ListUserCoupons(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil options")
	}

	// Test with empty promotion_id
	_, err = client.ListUserCoupons(context.Background(), &UserCouponsListOptions{})
	if err == nil {
		t.Error("Expected error for empty promotion_id")
	}
}

func TestUserCouponsListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("promotion_id") != "promo_456" {
			t.Errorf("Expected promotion_id=promo_456, got %s", r.URL.Query().Get("promotion_id"))
		}
		if r.URL.Query().Get("user_id") != "user_123" {
			t.Errorf("Expected user_id=user_123, got %s", r.URL.Query().Get("user_id"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := UserCouponsListResponse{Items: []UserCoupon{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListUserCoupons(context.Background(), &UserCouponsListOptions{
		PromotionID: "promo_456",
		UserID:      "user_123",
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("ListUserCoupons failed: %v", err)
	}
}

func TestUserCouponsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user_coupons/uc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		userCoupon := UserCoupon{ID: "uc_123", UserID: "user_1", CouponCode: "SAVE10"}
		_ = json.NewEncoder(w).Encode(userCoupon)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	userCoupon, err := client.GetUserCoupon(context.Background(), "uc_123")
	if err != nil {
		t.Fatalf("GetUserCoupon failed: %v", err)
	}

	if userCoupon.ID != "uc_123" {
		t.Errorf("Unexpected user coupon ID: %s", userCoupon.ID)
	}
}

func TestGetUserCouponEmptyID(t *testing.T) {
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
			_, err := client.GetUserCoupon(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "user coupon id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUserCouponsAssign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/user_coupons" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		userCoupon := UserCoupon{ID: "uc_new", UserID: "user_1", CouponID: "cpn_123"}
		_ = json.NewEncoder(w).Encode(userCoupon)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &UserCouponAssignRequest{
		UserID:   "user_1",
		CouponID: "cpn_123",
	}

	userCoupon, err := client.AssignUserCoupon(context.Background(), req)
	if err != nil {
		t.Fatalf("AssignUserCoupon failed: %v", err)
	}

	if userCoupon.UserID != "user_1" {
		t.Errorf("Unexpected user ID: %s", userCoupon.UserID)
	}
}

func TestUserCouponsRevoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/user_coupons/uc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.RevokeUserCoupon(context.Background(), "uc_123")
	if err != nil {
		t.Fatalf("RevokeUserCoupon failed: %v", err)
	}
}

func TestRevokeUserCouponEmptyID(t *testing.T) {
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
			err := client.RevokeUserCoupon(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "user coupon id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
