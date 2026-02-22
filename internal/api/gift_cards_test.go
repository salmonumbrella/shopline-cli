package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGiftCardsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/gift_cards" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := GiftCardsListResponse{
			Items: []GiftCard{
				{ID: "gc_123", MaskedCode: "****-****-****-ABCD", InitialValue: "100.00", Balance: "75.00", Currency: "USD", Status: GiftCardStatusEnabled},
				{ID: "gc_456", MaskedCode: "****-****-****-EFGH", InitialValue: "50.00", Balance: "50.00", Currency: "USD", Status: GiftCardStatusEnabled},
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

	giftCards, err := client.ListGiftCards(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListGiftCards failed: %v", err)
	}

	if len(giftCards.Items) != 2 {
		t.Errorf("Expected 2 gift cards, got %d", len(giftCards.Items))
	}
	if giftCards.Items[0].ID != "gc_123" {
		t.Errorf("Unexpected gift card ID: %s", giftCards.Items[0].ID)
	}
	if giftCards.Items[0].Balance != "75.00" {
		t.Errorf("Unexpected balance: %s", giftCards.Items[0].Balance)
	}
}

func TestGiftCardsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "enabled" {
			t.Errorf("Expected status=enabled, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("customer_id") != "cust_123" {
			t.Errorf("Expected customer_id=cust_123, got %s", r.URL.Query().Get("customer_id"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := GiftCardsListResponse{
			Items: []GiftCard{
				{ID: "gc_123", Status: GiftCardStatusEnabled},
			},
			Page:       2,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &GiftCardsListOptions{
		Page:       2,
		Status:     "enabled",
		CustomerID: "cust_123",
	}
	giftCards, err := client.ListGiftCards(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListGiftCards failed: %v", err)
	}

	if len(giftCards.Items) != 1 {
		t.Errorf("Expected 1 gift card, got %d", len(giftCards.Items))
	}
}

func TestGiftCardsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gift_cards/gc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		giftCard := GiftCard{
			ID:           "gc_123",
			Code:         "ABCD-EFGH-IJKL-MNOP",
			MaskedCode:   "****-****-****-MNOP",
			InitialValue: "100.00",
			Balance:      "75.00",
			Currency:     "USD",
			Status:       GiftCardStatusEnabled,
			Note:         "Birthday gift",
		}
		_ = json.NewEncoder(w).Encode(giftCard)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	giftCard, err := client.GetGiftCard(context.Background(), "gc_123")
	if err != nil {
		t.Fatalf("GetGiftCard failed: %v", err)
	}

	if giftCard.ID != "gc_123" {
		t.Errorf("Unexpected gift card ID: %s", giftCard.ID)
	}
	if giftCard.Note != "Birthday gift" {
		t.Errorf("Unexpected note: %s", giftCard.Note)
	}
	if giftCard.Status != GiftCardStatusEnabled {
		t.Errorf("Unexpected status: %s", giftCard.Status)
	}
}

func TestGiftCardsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/gift_cards" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req GiftCardCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.InitialValue != "100.00" {
			t.Errorf("Unexpected initial value: %s", req.InitialValue)
		}
		if req.Currency != "USD" {
			t.Errorf("Unexpected currency: %s", req.Currency)
		}

		giftCard := GiftCard{
			ID:           "gc_new",
			Code:         "XXXX-YYYY-ZZZZ-AAAA",
			MaskedCode:   "****-****-****-AAAA",
			InitialValue: req.InitialValue,
			Balance:      req.InitialValue,
			Currency:     req.Currency,
			Status:       GiftCardStatusEnabled,
		}
		_ = json.NewEncoder(w).Encode(giftCard)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &GiftCardCreateRequest{
		InitialValue: "100.00",
		Currency:     "USD",
	}
	giftCard, err := client.CreateGiftCard(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateGiftCard failed: %v", err)
	}

	if giftCard.ID != "gc_new" {
		t.Errorf("Unexpected gift card ID: %s", giftCard.ID)
	}
	if giftCard.Balance != "100.00" {
		t.Errorf("Unexpected balance: %s", giftCard.Balance)
	}
}

func TestGiftCardsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/gift_cards/gc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteGiftCard(context.Background(), "gc_123")
	if err != nil {
		t.Fatalf("DeleteGiftCard failed: %v", err)
	}
}

func TestGetGiftCardEmptyID(t *testing.T) {
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
			_, err := client.GetGiftCard(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "gift card id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteGiftCardEmptyID(t *testing.T) {
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
			err := client.DeleteGiftCard(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "gift card id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
