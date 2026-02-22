package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDomainsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/domains" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := DomainsListResponse{
			Items: []Domain{
				{ID: "dom_123", Host: "example.com", Primary: true, SSL: true, Status: DomainStatusActive},
				{ID: "dom_456", Host: "shop.example.com", Primary: false, SSL: true, Status: DomainStatusActive},
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

	domains, err := client.ListDomains(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListDomains failed: %v", err)
	}

	if len(domains.Items) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(domains.Items))
	}
	if domains.Items[0].ID != "dom_123" {
		t.Errorf("Unexpected domain ID: %s", domains.Items[0].ID)
	}
	if !domains.Items[0].Primary {
		t.Error("Expected Primary to be true")
	}
}

func TestDomainsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("primary") != "true" {
			t.Errorf("Expected primary=true, got %s", r.URL.Query().Get("primary"))
		}

		resp := DomainsListResponse{
			Items:      []Domain{{ID: "dom_123", Primary: true, Status: DomainStatusActive}},
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

	primary := true
	opts := &DomainsListOptions{
		Status:  DomainStatusActive,
		Primary: &primary,
	}
	domains, err := client.ListDomains(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListDomains failed: %v", err)
	}

	if len(domains.Items) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(domains.Items))
	}
}

func TestDomainsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains/dom_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		domain := Domain{
			ID:                "dom_123",
			Host:              "example.com",
			Primary:           true,
			SSL:               true,
			SSLStatus:         "active",
			Status:            DomainStatusActive,
			VerificationDNS:   "TXT _verify.example.com",
			VerificationToken: "abc123",
			Verified:          true,
		}
		_ = json.NewEncoder(w).Encode(domain)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	domain, err := client.GetDomain(context.Background(), "dom_123")
	if err != nil {
		t.Fatalf("GetDomain failed: %v", err)
	}

	if domain.ID != "dom_123" {
		t.Errorf("Unexpected domain ID: %s", domain.ID)
	}
	if !domain.Verified {
		t.Error("Expected Verified to be true")
	}
	if domain.VerificationToken != "abc123" {
		t.Errorf("Unexpected verification token: %s", domain.VerificationToken)
	}
}

func TestGetDomainEmptyID(t *testing.T) {
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
			_, err := client.GetDomain(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "domain id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDomainsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req DomainCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Host != "newdomain.com" {
			t.Errorf("Unexpected host: %s", req.Host)
		}

		domain := Domain{
			ID:     "dom_new",
			Host:   req.Host,
			Status: DomainStatusPending,
		}
		_ = json.NewEncoder(w).Encode(domain)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &DomainCreateRequest{
		Host: "newdomain.com",
	}

	domain, err := client.CreateDomain(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	if domain.ID != "dom_new" {
		t.Errorf("Unexpected domain ID: %s", domain.ID)
	}
	if domain.Status != DomainStatusPending {
		t.Errorf("Unexpected status: %s", domain.Status)
	}
}

func TestDomainsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/domains/dom_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req DomainUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Primary == nil || *req.Primary != true {
			t.Error("Expected Primary to be true")
		}

		domain := Domain{ID: "dom_123", Host: "example.com", Primary: true}
		_ = json.NewEncoder(w).Encode(domain)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	primary := true
	req := &DomainUpdateRequest{
		Primary: &primary,
	}

	domain, err := client.UpdateDomain(context.Background(), "dom_123", req)
	if err != nil {
		t.Fatalf("UpdateDomain failed: %v", err)
	}

	if !domain.Primary {
		t.Error("Expected Primary to be true")
	}
}

func TestDomainsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/domains/dom_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteDomain(context.Background(), "dom_123")
	if err != nil {
		t.Fatalf("DeleteDomain failed: %v", err)
	}
}

func TestDeleteDomainEmptyID(t *testing.T) {
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
			err := client.DeleteDomain(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "domain id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDomainsVerify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/domains/dom_123/verify" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		domain := Domain{
			ID:       "dom_123",
			Host:     "example.com",
			Status:   DomainStatusVerifying,
			Verified: false,
		}
		_ = json.NewEncoder(w).Encode(domain)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	domain, err := client.VerifyDomain(context.Background(), "dom_123")
	if err != nil {
		t.Fatalf("VerifyDomain failed: %v", err)
	}

	if domain.Status != DomainStatusVerifying {
		t.Errorf("Unexpected status: %s", domain.Status)
	}
}
