package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMarketsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/markets" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := MarketsListResponse{
			Items: []Market{
				{ID: "mkt_123", Name: "US Market", Handle: "us", Enabled: true, Primary: true},
				{ID: "mkt_456", Name: "EU Market", Handle: "eu", Enabled: true},
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

	markets, err := client.ListMarkets(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListMarkets failed: %v", err)
	}

	if len(markets.Items) != 2 {
		t.Errorf("Expected 2 markets, got %d", len(markets.Items))
	}
	if markets.Items[0].ID != "mkt_123" {
		t.Errorf("Unexpected market ID: %s", markets.Items[0].ID)
	}
}

func TestMarketsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/mkt_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		market := Market{ID: "mkt_123", Name: "US Market", Handle: "us"}
		_ = json.NewEncoder(w).Encode(market)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	market, err := client.GetMarket(context.Background(), "mkt_123")
	if err != nil {
		t.Fatalf("GetMarket failed: %v", err)
	}

	if market.ID != "mkt_123" {
		t.Errorf("Unexpected market ID: %s", market.ID)
	}
}

func TestGetMarketEmptyID(t *testing.T) {
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
			_, err := client.GetMarket(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "market id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMarketsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		market := Market{ID: "mkt_new", Name: "New Market"}
		_ = json.NewEncoder(w).Encode(market)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MarketCreateRequest{
		Name:      "New Market",
		Countries: []string{"US", "CA"},
	}

	market, err := client.CreateMarket(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateMarket failed: %v", err)
	}

	if market.ID != "mkt_new" {
		t.Errorf("Unexpected market ID: %s", market.ID)
	}
}

func TestMarketsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/markets/mkt_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteMarket(context.Background(), "mkt_123")
	if err != nil {
		t.Fatalf("DeleteMarket failed: %v", err)
	}
}

func TestListMarketsWithOptions(t *testing.T) {
	testCases := []struct {
		name          string
		opts          *MarketsListOptions
		expectedQuery map[string]string
	}{
		{
			name:          "page only",
			opts:          &MarketsListOptions{Page: 2},
			expectedQuery: map[string]string{"page": "2"},
		},
		{
			name:          "page_size only",
			opts:          &MarketsListOptions{PageSize: 50},
			expectedQuery: map[string]string{"page_size": "50"},
		},
		{
			name:          "all options combined",
			opts:          &MarketsListOptions{Page: 3, PageSize: 25},
			expectedQuery: map[string]string{"page": "3", "page_size": "25"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET, got %s", r.Method)
				}

				query := r.URL.Query()
				for key, expectedValue := range tc.expectedQuery {
					if got := query.Get(key); got != expectedValue {
						t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, got)
					}
				}

				resp := MarketsListResponse{
					Items:      []Market{{ID: "mkt_123", Name: "Test Market"}},
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

			_, err := client.ListMarkets(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListMarkets failed: %v", err)
			}
		})
	}
}

func TestDeleteMarketEmptyID(t *testing.T) {
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
			err := client.DeleteMarket(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "market id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
