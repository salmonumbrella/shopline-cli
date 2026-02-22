package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetafieldsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/metafields" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := MetafieldsListResponse{
			Items: []Metafield{
				{ID: "mf_123", Namespace: "custom", Key: "color", Value: "red"},
				{ID: "mf_456", Namespace: "custom", Key: "size", Value: "large"},
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

	metafields, err := client.ListMetafields(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListMetafields failed: %v", err)
	}

	if len(metafields.Items) != 2 {
		t.Errorf("Expected 2 metafields, got %d", len(metafields.Items))
	}
	if metafields.Items[0].ID != "mf_123" {
		t.Errorf("Unexpected metafield ID: %s", metafields.Items[0].ID)
	}
}

func TestMetafieldsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metafields/mf_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		metafield := Metafield{ID: "mf_123", Namespace: "custom", Key: "color", Value: "red"}
		_ = json.NewEncoder(w).Encode(metafield)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	metafield, err := client.GetMetafield(context.Background(), "mf_123")
	if err != nil {
		t.Fatalf("GetMetafield failed: %v", err)
	}

	if metafield.ID != "mf_123" {
		t.Errorf("Unexpected metafield ID: %s", metafield.ID)
	}
}

func TestGetMetafieldEmptyID(t *testing.T) {
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
			_, err := client.GetMetafield(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "metafield id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMetafieldsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/metafields" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		metafield := Metafield{ID: "mf_new", Namespace: "custom", Key: "weight", Value: "100"}
		_ = json.NewEncoder(w).Encode(metafield)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MetafieldCreateRequest{Namespace: "custom", Key: "weight", Value: "100", ValueType: "integer"}
	metafield, err := client.CreateMetafield(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateMetafield failed: %v", err)
	}

	if metafield.ID != "mf_new" {
		t.Errorf("Unexpected metafield ID: %s", metafield.ID)
	}
}

func TestMetafieldsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/metafields/mf_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteMetafield(context.Background(), "mf_123")
	if err != nil {
		t.Fatalf("DeleteMetafield failed: %v", err)
	}
}

func TestDeleteMetafieldEmptyID(t *testing.T) {
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
			err := client.DeleteMetafield(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "metafield id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMetafieldsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/metafields/mf_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		metafield := Metafield{ID: "mf_123", Namespace: "custom", Key: "color", Value: "blue"}
		_ = json.NewEncoder(w).Encode(metafield)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MetafieldUpdateRequest{Value: "blue"}
	metafield, err := client.UpdateMetafield(context.Background(), "mf_123", req)
	if err != nil {
		t.Fatalf("UpdateMetafield failed: %v", err)
	}

	if metafield.ID != "mf_123" {
		t.Errorf("Unexpected metafield ID: %s", metafield.ID)
	}
	if metafield.Value != "blue" {
		t.Errorf("Unexpected metafield value: %s", metafield.Value)
	}
}

func TestUpdateMetafieldEmptyID(t *testing.T) {
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
			req := &MetafieldUpdateRequest{Value: "test"}
			_, err := client.UpdateMetafield(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "metafield id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMetafieldsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}
		if query.Get("namespace") != "custom" {
			t.Errorf("Expected namespace=custom, got %s", query.Get("namespace"))
		}
		if query.Get("key") != "color" {
			t.Errorf("Expected key=color, got %s", query.Get("key"))
		}
		if query.Get("owner_id") != "prod_123" {
			t.Errorf("Expected owner_id=prod_123, got %s", query.Get("owner_id"))
		}
		if query.Get("owner_type") != "product" {
			t.Errorf("Expected owner_type=product, got %s", query.Get("owner_type"))
		}

		resp := MetafieldsListResponse{
			Items:      []Metafield{{ID: "mf_123", Namespace: "custom", Key: "color"}},
			Page:       2,
			PageSize:   50,
			TotalCount: 100,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &MetafieldsListOptions{
		Page:      2,
		PageSize:  50,
		Namespace: "custom",
		Key:       "color",
		OwnerID:   "prod_123",
		OwnerType: "product",
	}
	metafields, err := client.ListMetafields(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListMetafields failed: %v", err)
	}

	if len(metafields.Items) != 1 {
		t.Errorf("Expected 1 metafield, got %d", len(metafields.Items))
	}
}
