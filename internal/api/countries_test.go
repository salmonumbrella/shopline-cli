package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCountriesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/countries" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CountriesListResponse{
			Items: []Country{
				{Code: "US", Name: "United States", Tax: 0.0},
				{Code: "CA", Name: "Canada", Tax: 5.0, TaxName: "GST"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	countries, err := client.ListCountries(context.Background())
	if err != nil {
		t.Fatalf("ListCountries failed: %v", err)
	}

	if len(countries.Items) != 2 {
		t.Errorf("Expected 2 countries, got %d", len(countries.Items))
	}
	if countries.Items[0].Code != "US" {
		t.Errorf("Unexpected country code: %s", countries.Items[0].Code)
	}
}

func TestCountriesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/countries/US" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		country := Country{
			Code: "US",
			Name: "United States",
			Provinces: []Province{
				{Code: "CA", Name: "California", Tax: 7.25},
				{Code: "NY", Name: "New York", Tax: 8.0},
			},
		}
		_ = json.NewEncoder(w).Encode(country)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	country, err := client.GetCountry(context.Background(), "US")
	if err != nil {
		t.Fatalf("GetCountry failed: %v", err)
	}

	if country.Code != "US" {
		t.Errorf("Unexpected country code: %s", country.Code)
	}
	if len(country.Provinces) != 2 {
		t.Errorf("Expected 2 provinces, got %d", len(country.Provinces))
	}
}

func TestGetCountryEmptyCode(t *testing.T) {
	client := NewClient("token")

	_, err := client.GetCountry(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty code, got nil")
	}
	if err != nil && err.Error() != "country code is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
