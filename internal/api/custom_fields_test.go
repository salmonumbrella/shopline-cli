package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomFieldsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/custom_fields" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// API returns array directly, not wrapped in response struct
		items := []CustomField{
			{ID: "cf_123", Type: CustomFieldTypeText},
			{ID: "cf_456", Type: CustomFieldTypeSelect},
		}
		_ = json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fields, err := client.ListCustomFields(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCustomFields failed: %v", err)
	}

	if len(fields.Items) != 2 {
		t.Errorf("Expected 2 custom fields, got %d", len(fields.Items))
	}
	if fields.Items[0].ID != "cf_123" {
		t.Errorf("Unexpected custom field ID: %s", fields.Items[0].ID)
	}
	if fields.Items[0].Type != CustomFieldTypeText {
		t.Errorf("Unexpected type: %s", fields.Items[0].Type)
	}
}

func TestCustomFieldsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("owner_type") != "product" {
			t.Errorf("Expected owner_type=product, got %s", r.URL.Query().Get("owner_type"))
		}
		if r.URL.Query().Get("type") != "text" {
			t.Errorf("Expected type=text, got %s", r.URL.Query().Get("type"))
		}

		// API returns array directly
		items := []CustomField{{ID: "cf_123", Type: CustomFieldTypeText}}
		_ = json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomFieldsListOptions{
		OwnerType: CustomFieldOwnerProduct,
		Type:      CustomFieldTypeText,
	}
	fields, err := client.ListCustomFields(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListCustomFields failed: %v", err)
	}

	if len(fields.Items) != 1 {
		t.Errorf("Expected 1 custom field, got %d", len(fields.Items))
	}
}

func TestCustomFieldsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/custom_fields/cf_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		field := CustomField{
			ID:   "cf_123",
			Type: CustomFieldTypeSelect,
			Options: map[string]interface{}{
				"values": []string{"Red", "Blue", "Green"},
			},
		}
		_ = json.NewEncoder(w).Encode(field)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	field, err := client.GetCustomField(context.Background(), "cf_123")
	if err != nil {
		t.Fatalf("GetCustomField failed: %v", err)
	}

	if field.ID != "cf_123" {
		t.Errorf("Unexpected custom field ID: %s", field.ID)
	}
	if field.Options == nil {
		t.Error("Expected Options to be non-nil")
	}
}

func TestGetCustomFieldEmptyID(t *testing.T) {
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
			_, err := client.GetCustomField(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "custom field id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCustomFieldsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req CustomFieldCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Material" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Key != "material" {
			t.Errorf("Unexpected key: %s", req.Key)
		}
		if req.Type != CustomFieldTypeText {
			t.Errorf("Unexpected type: %s", req.Type)
		}

		field := CustomField{
			ID:   "cf_new",
			Type: req.Type,
		}
		_ = json.NewEncoder(w).Encode(field)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomFieldCreateRequest{
		Name:      "Material",
		Key:       "material",
		Type:      CustomFieldTypeText,
		OwnerType: CustomFieldOwnerProduct,
	}

	field, err := client.CreateCustomField(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCustomField failed: %v", err)
	}

	if field.ID != "cf_new" {
		t.Errorf("Unexpected custom field ID: %s", field.ID)
	}
}

func TestCustomFieldsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/custom_fields/cf_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CustomFieldUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		field := CustomField{ID: "cf_123", Type: CustomFieldTypeText}
		_ = json.NewEncoder(w).Encode(field)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	visible := true
	req := &CustomFieldUpdateRequest{
		Name:    "Updated Color",
		Visible: &visible,
	}

	field, err := client.UpdateCustomField(context.Background(), "cf_123", req)
	if err != nil {
		t.Fatalf("UpdateCustomField failed: %v", err)
	}

	if field.ID != "cf_123" {
		t.Errorf("Unexpected ID: %s", field.ID)
	}
}

func TestCustomFieldsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/custom_fields/cf_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCustomField(context.Background(), "cf_123")
	if err != nil {
		t.Fatalf("DeleteCustomField failed: %v", err)
	}
}

func TestDeleteCustomFieldEmptyID(t *testing.T) {
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
			err := client.DeleteCustomField(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "custom field id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
