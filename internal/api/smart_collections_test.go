package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSmartCollectionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/smart_collections" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := SmartCollectionsListResponse{
			Items: []SmartCollection{
				{
					ID:          "sc_123",
					Title:       "Sale Items",
					Disjunctive: false,
					Rules: []Rule{
						{Column: "tag", Relation: "equals", Condition: "sale"},
					},
				},
				{
					ID:          "sc_456",
					Title:       "New Arrivals",
					Disjunctive: true,
					Rules: []Rule{
						{Column: "created_at", Relation: "greater_than", Condition: "2024-01-01"},
					},
				},
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

	collections, err := client.ListSmartCollections(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListSmartCollections failed: %v", err)
	}

	if len(collections.Items) != 2 {
		t.Errorf("Expected 2 smart collections, got %d", len(collections.Items))
	}
	if collections.Items[0].ID != "sc_123" {
		t.Errorf("Unexpected smart collection ID: %s", collections.Items[0].ID)
	}
	if len(collections.Items[0].Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(collections.Items[0].Rules))
	}
}

func TestSmartCollectionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/smart_collections/sc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		collection := SmartCollection{
			ID:          "sc_123",
			Title:       "Sale Items",
			Handle:      "sale-items",
			Disjunctive: false,
			Rules: []Rule{
				{Column: "tag", Relation: "equals", Condition: "sale"},
			},
			Published: true,
		}
		_ = json.NewEncoder(w).Encode(collection)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	collection, err := client.GetSmartCollection(context.Background(), "sc_123")
	if err != nil {
		t.Fatalf("GetSmartCollection failed: %v", err)
	}

	if collection.ID != "sc_123" {
		t.Errorf("Unexpected smart collection ID: %s", collection.ID)
	}
	if collection.Title != "Sale Items" {
		t.Errorf("Unexpected title: %s", collection.Title)
	}
	if len(collection.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(collection.Rules))
	}
}

func TestGetSmartCollectionEmptyID(t *testing.T) {
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
			_, err := client.GetSmartCollection(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "smart collection id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSmartCollectionsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/smart_collections" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req SmartCollectionCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Title != "New Smart Collection" {
			t.Errorf("Unexpected title in request: %s", req.Title)
		}
		if len(req.Rules) != 1 {
			t.Errorf("Expected 1 rule in request, got %d", len(req.Rules))
		}

		collection := SmartCollection{
			ID:          "sc_new",
			Title:       req.Title,
			Handle:      "new-smart-collection",
			Disjunctive: req.Disjunctive,
			Rules:       req.Rules,
			Published:   req.Published,
		}
		_ = json.NewEncoder(w).Encode(collection)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &SmartCollectionCreateRequest{
		Title:       "New Smart Collection",
		Disjunctive: false,
		Rules: []Rule{
			{Column: "vendor", Relation: "equals", Condition: "Nike"},
		},
		Published: true,
	}
	collection, err := client.CreateSmartCollection(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSmartCollection failed: %v", err)
	}

	if collection.ID != "sc_new" {
		t.Errorf("Unexpected smart collection ID: %s", collection.ID)
	}
	if len(collection.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(collection.Rules))
	}
}

func TestSmartCollectionsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/smart_collections/sc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteSmartCollection(context.Background(), "sc_123")
	if err != nil {
		t.Fatalf("DeleteSmartCollection failed: %v", err)
	}
}

func TestDeleteSmartCollectionEmptyID(t *testing.T) {
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
			err := client.DeleteSmartCollection(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "smart collection id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSmartCollectionsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/smart_collections/sc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		collection := SmartCollection{
			ID:          "sc_123",
			Title:       "Updated Smart Collection",
			Handle:      "updated-smart-collection",
			Disjunctive: true,
			Rules: []Rule{
				{Column: "vendor", Relation: "equals", Condition: "Adidas"},
			},
			Published: true,
		}
		_ = json.NewEncoder(w).Encode(collection)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	published := true
	req := &SmartCollectionUpdateRequest{
		Title:     "Updated Smart Collection",
		Published: &published,
	}
	collection, err := client.UpdateSmartCollection(context.Background(), "sc_123", req)
	if err != nil {
		t.Fatalf("UpdateSmartCollection failed: %v", err)
	}

	if collection.ID != "sc_123" {
		t.Errorf("Unexpected smart collection ID: %s", collection.ID)
	}
	if collection.Title != "Updated Smart Collection" {
		t.Errorf("Unexpected title: %s", collection.Title)
	}
}
