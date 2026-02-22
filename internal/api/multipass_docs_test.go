package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMultipassSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/secret" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"secret": "s"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetMultipassSecret(context.Background())
	if err != nil {
		t.Fatalf("GetMultipassSecret failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["secret"] != "s" {
		t.Fatalf("expected secret, got %v", got["secret"])
	}
}

func TestCreateMultipassSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/secret" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"secret": "new"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.CreateMultipassSecret(context.Background(), nil)
	if err != nil {
		t.Fatalf("CreateMultipassSecret failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["secret"] != "new" {
		t.Fatalf("expected secret=new, got %v", got["secret"])
	}
}

func TestListMultipassLinkings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/linkings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("customer_ids") != "c1,c2" {
			t.Errorf("Expected customer_ids=c1,c2, got %s", r.URL.Query().Get("customer_ids"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.ListMultipassLinkings(context.Background(), []string{"c1", "c2"})
	if err != nil {
		t.Fatalf("ListMultipassLinkings failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if _, ok := got["items"]; !ok {
		t.Fatalf("expected items, got %v", got)
	}
}

func TestUpdateMultipassCustomerLinking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/customers/cust_123/linkings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["ok"] != true {
			t.Fatalf("expected ok=true in body, got %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"updated": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.UpdateMultipassCustomerLinking(context.Background(), "cust_123", map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("UpdateMultipassCustomerLinking failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["updated"] != true {
		t.Fatalf("expected updated=true, got %v", got["updated"])
	}
}

func TestUpdateMultipassCustomerLinkingEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.UpdateMultipassCustomerLinking(context.Background(), " ", map[string]any{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "customer id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMultipassCustomerLinking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/customers/cust_123/linkings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"deleted": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.DeleteMultipassCustomerLinking(context.Background(), "cust_123")
	if err != nil {
		t.Fatalf("DeleteMultipassCustomerLinking failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["deleted"] != true {
		t.Fatalf("expected deleted=true, got %v", got["deleted"])
	}
}

func TestDeleteMultipassCustomerLinkingEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.DeleteMultipassCustomerLinking(context.Background(), " ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "customer id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}
