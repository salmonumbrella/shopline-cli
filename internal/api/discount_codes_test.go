package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscountCodesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/discount_codes" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := DiscountCodesListResponse{
			Items: []DiscountCode{
				{ID: "dc_123", Code: "SAVE20", DiscountValue: 20, Status: "active"},
				{ID: "dc_456", Code: "WELCOME10", DiscountValue: 10, Status: "active"},
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

	codes, err := client.ListDiscountCodes(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListDiscountCodes failed: %v", err)
	}

	if len(codes.Items) != 2 {
		t.Errorf("Expected 2 codes, got %d", len(codes.Items))
	}
	if codes.Items[0].ID != "dc_123" {
		t.Errorf("Unexpected code ID: %s", codes.Items[0].ID)
	}
}

func TestDiscountCodesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/discount_codes/dc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		code := DiscountCode{ID: "dc_123", Code: "SAVE20", DiscountValue: 20}
		_ = json.NewEncoder(w).Encode(code)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	code, err := client.GetDiscountCode(context.Background(), "dc_123")
	if err != nil {
		t.Fatalf("GetDiscountCode failed: %v", err)
	}

	if code.ID != "dc_123" {
		t.Errorf("Unexpected code ID: %s", code.ID)
	}
}

func TestGetDiscountCodeEmptyID(t *testing.T) {
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
			_, err := client.GetDiscountCode(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "discount code id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetDiscountCodeByCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/discount_codes/lookup" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("code") != "SAVE20" {
			t.Errorf("Expected code=SAVE20, got %s", r.URL.Query().Get("code"))
		}

		code := DiscountCode{ID: "dc_123", Code: "SAVE20", DiscountValue: 20}
		_ = json.NewEncoder(w).Encode(code)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	code, err := client.GetDiscountCodeByCode(context.Background(), "SAVE20")
	if err != nil {
		t.Fatalf("GetDiscountCodeByCode failed: %v", err)
	}

	if code.Code != "SAVE20" {
		t.Errorf("Unexpected code: %s", code.Code)
	}
}

func TestDiscountCodesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/discount_codes" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		code := DiscountCode{ID: "dc_new", Code: "NEWCODE", DiscountValue: 15}
		_ = json.NewEncoder(w).Encode(code)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &DiscountCodeCreateRequest{Code: "NEWCODE", DiscountType: "percentage", DiscountValue: 15}
	code, err := client.CreateDiscountCode(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateDiscountCode failed: %v", err)
	}

	if code.ID != "dc_new" {
		t.Errorf("Unexpected code ID: %s", code.ID)
	}
}

func TestDiscountCodesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/discount_codes/dc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteDiscountCode(context.Background(), "dc_123")
	if err != nil {
		t.Fatalf("DeleteDiscountCode failed: %v", err)
	}
}

func TestDeleteDiscountCodeEmptyID(t *testing.T) {
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
			err := client.DeleteDiscountCode(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "discount code id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListDiscountCodesWithOptions(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *DiscountCodesListOptions
		expectedParams map[string]string
	}{
		{
			name: "with page",
			opts: &DiscountCodesListOptions{Page: 2},
			expectedParams: map[string]string{
				"page": "2",
			},
		},
		{
			name: "with page_size",
			opts: &DiscountCodesListOptions{PageSize: 50},
			expectedParams: map[string]string{
				"page_size": "50",
			},
		},
		{
			name: "with price_rule_id",
			opts: &DiscountCodesListOptions{PriceRuleID: "pr_123"},
			expectedParams: map[string]string{
				"price_rule_id": "pr_123",
			},
		},
		{
			name: "with status",
			opts: &DiscountCodesListOptions{Status: "active"},
			expectedParams: map[string]string{
				"status": "active",
			},
		},
		{
			name: "with all options",
			opts: &DiscountCodesListOptions{
				Page:        3,
				PageSize:    25,
				PriceRuleID: "pr_456",
				Status:      "expired",
			},
			expectedParams: map[string]string{
				"page":          "3",
				"page_size":     "25",
				"price_rule_id": "pr_456",
				"status":        "expired",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET, got %s", r.Method)
				}

				// Verify query parameters
				for key, expectedValue := range tc.expectedParams {
					actualValue := r.URL.Query().Get(key)
					if actualValue != expectedValue {
						t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
					}
				}

				resp := DiscountCodesListResponse{
					Items: []DiscountCode{
						{ID: "dc_123", Code: "SAVE20", Status: "active"},
					},
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

			codes, err := client.ListDiscountCodes(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListDiscountCodes failed: %v", err)
			}

			if len(codes.Items) != 1 {
				t.Errorf("Expected 1 code, got %d", len(codes.Items))
			}
		})
	}
}

func TestGetDiscountCodeByCodeEmptyCode(t *testing.T) {
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
			_, err := client.GetDiscountCodeByCode(context.Background(), tc.code)
			if err == nil {
				t.Error("Expected error for empty code, got nil")
			}
			if err != nil && err.Error() != "discount code is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetDiscountCodeByCodeURLEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/discount_codes/lookup" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		// Verify the code with special characters is properly URL-encoded
		codeParam := r.URL.Query().Get("code")
		if codeParam != "SAVE 20%" {
			t.Errorf("Expected code=SAVE 20%%, got %s", codeParam)
		}

		code := DiscountCode{ID: "dc_123", Code: "SAVE 20%", DiscountValue: 20}
		_ = json.NewEncoder(w).Encode(code)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	code, err := client.GetDiscountCodeByCode(context.Background(), "SAVE 20%")
	if err != nil {
		t.Fatalf("GetDiscountCodeByCode failed: %v", err)
	}

	if code.Code != "SAVE 20%" {
		t.Errorf("Unexpected code: %s", code.Code)
	}
}

func TestDiscountCodesListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListDiscountCodes(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for server error response, got nil")
	}
}

func TestDiscountCodesGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetDiscountCode(context.Background(), "dc_notfound")
	if err == nil {
		t.Error("Expected error for not found response, got nil")
	}
}

func TestGetDiscountCodeByCodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetDiscountCodeByCode(context.Background(), "INVALID")
	if err == nil {
		t.Error("Expected error for not found response, got nil")
	}
}

func TestDiscountCodesCreateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid discount code"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &DiscountCodeCreateRequest{Code: "BADCODE", DiscountType: "invalid", DiscountValue: -10}
	_, err := client.CreateDiscountCode(context.Background(), req)
	if err == nil {
		t.Error("Expected error for bad request response, got nil")
	}
}

func TestDiscountCodesCreateRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Verify request body
		var reqBody DiscountCodeCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if reqBody.Code != "TESTCODE" {
			t.Errorf("Expected code=TESTCODE, got %s", reqBody.Code)
		}
		if reqBody.DiscountType != "percentage" {
			t.Errorf("Expected discount_type=percentage, got %s", reqBody.DiscountType)
		}
		if reqBody.DiscountValue != 25.0 {
			t.Errorf("Expected discount_value=25.0, got %f", reqBody.DiscountValue)
		}
		if reqBody.UsageLimit != 100 {
			t.Errorf("Expected usage_limit=100, got %d", reqBody.UsageLimit)
		}
		if reqBody.MinPurchase != 50.0 {
			t.Errorf("Expected min_purchase=50.0, got %f", reqBody.MinPurchase)
		}

		code := DiscountCode{
			ID:            "dc_created",
			Code:          reqBody.Code,
			DiscountType:  reqBody.DiscountType,
			DiscountValue: reqBody.DiscountValue,
			UsageLimit:    reqBody.UsageLimit,
			MinPurchase:   reqBody.MinPurchase,
		}
		_ = json.NewEncoder(w).Encode(code)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &DiscountCodeCreateRequest{
		Code:          "TESTCODE",
		DiscountType:  "percentage",
		DiscountValue: 25.0,
		UsageLimit:    100,
		MinPurchase:   50.0,
	}
	code, err := client.CreateDiscountCode(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateDiscountCode failed: %v", err)
	}

	if code.ID != "dc_created" {
		t.Errorf("Unexpected code ID: %s", code.ID)
	}
	if code.DiscountValue != 25.0 {
		t.Errorf("Unexpected discount value: %f", code.DiscountValue)
	}
}

func TestDiscountCodesDeleteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteDiscountCode(context.Background(), "dc_notfound")
	if err == nil {
		t.Error("Expected error for not found response, got nil")
	}
}
