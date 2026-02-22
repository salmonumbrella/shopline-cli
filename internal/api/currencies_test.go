package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCurrenciesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/currencies" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CurrenciesListResponse{
			Items: []Currency{
				{Code: "USD", Name: "US Dollar", Symbol: "$", Primary: true, Enabled: true},
				{Code: "EUR", Name: "Euro", Symbol: "EUR", Primary: false, Enabled: true},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	currencies, err := client.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListCurrencies failed: %v", err)
	}

	if len(currencies.Items) != 2 {
		t.Errorf("Expected 2 currencies, got %d", len(currencies.Items))
	}
	if currencies.Items[0].Code != "USD" {
		t.Errorf("Unexpected currency code: %s", currencies.Items[0].Code)
	}
}

func TestCurrenciesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/currencies/USD" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		currency := Currency{Code: "USD", Name: "US Dollar", Symbol: "$"}
		_ = json.NewEncoder(w).Encode(currency)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	currency, err := client.GetCurrency(context.Background(), "USD")
	if err != nil {
		t.Fatalf("GetCurrency failed: %v", err)
	}

	if currency.Code != "USD" {
		t.Errorf("Unexpected currency code: %s", currency.Code)
	}
}

func TestGetCurrencyEmptyCode(t *testing.T) {
	client := NewClient("token")

	_, err := client.GetCurrency(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty code, got nil")
	}
	if err != nil && err.Error() != "currency code is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestCurrenciesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		currency := Currency{Code: "EUR", Enabled: true, ExchangeRate: 0.85}
		_ = json.NewEncoder(w).Encode(currency)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	enabled := true
	rate := 0.85
	req := &CurrencyUpdateRequest{
		Enabled:      &enabled,
		ExchangeRate: &rate,
	}

	currency, err := client.UpdateCurrency(context.Background(), "EUR", req)
	if err != nil {
		t.Fatalf("UpdateCurrency failed: %v", err)
	}

	if currency.ExchangeRate != 0.85 {
		t.Errorf("Unexpected exchange rate: %f", currency.ExchangeRate)
	}
}
