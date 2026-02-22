package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetafieldDefinitionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/metafield_definitions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := MetafieldDefinitionsListResponse{
			Items: []MetafieldDefinition{
				{ID: "md_123", Name: "Color", Namespace: "custom", Key: "color", Type: "single_line_text_field"},
				{ID: "md_456", Name: "Size", Namespace: "custom", Key: "size", Type: "single_line_text_field"},
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

	defs, err := client.ListMetafieldDefinitions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListMetafieldDefinitions failed: %v", err)
	}

	if len(defs.Items) != 2 {
		t.Errorf("Expected 2 metafield definitions, got %d", len(defs.Items))
	}
	if defs.Items[0].ID != "md_123" {
		t.Errorf("Unexpected metafield definition ID: %s", defs.Items[0].ID)
	}
}

func TestMetafieldDefinitionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metafield_definitions/md_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		def := MetafieldDefinition{
			ID:        "md_123",
			Name:      "Color",
			Namespace: "custom",
			Key:       "color",
			Type:      "single_line_text_field",
			OwnerType: "product",
		}
		_ = json.NewEncoder(w).Encode(def)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	def, err := client.GetMetafieldDefinition(context.Background(), "md_123")
	if err != nil {
		t.Fatalf("GetMetafieldDefinition failed: %v", err)
	}

	if def.ID != "md_123" {
		t.Errorf("Unexpected metafield definition ID: %s", def.ID)
	}
	if def.OwnerType != "product" {
		t.Errorf("Unexpected owner type: %s", def.OwnerType)
	}
}

func TestGetMetafieldDefinitionEmptyID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{name: "empty string", id: ""},
		{name: "whitespace only", id: "   "},
		{name: "tab character", id: "\t"},
		{name: "newline", id: "\n"},
	}

	client := NewClient("token")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetMetafieldDefinition(context.Background(), tt.id)
			if err == nil {
				t.Error("Expected error for empty/whitespace ID, got nil")
			}
			if err != nil && err.Error() != "metafield definition id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMetafieldDefinitionsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		def := MetafieldDefinition{ID: "md_new", Name: "Material", Namespace: "custom", Key: "material"}
		_ = json.NewEncoder(w).Encode(def)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MetafieldDefinitionCreateRequest{
		Name:      "Material",
		Namespace: "custom",
		Key:       "material",
		Type:      "single_line_text_field",
		OwnerType: "product",
	}

	def, err := client.CreateMetafieldDefinition(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateMetafieldDefinition failed: %v", err)
	}

	if def.ID != "md_new" {
		t.Errorf("Unexpected metafield definition ID: %s", def.ID)
	}
}

func TestMetafieldDefinitionsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		def := MetafieldDefinition{ID: "md_123", Name: "Updated Color", Namespace: "custom", Key: "color"}
		_ = json.NewEncoder(w).Encode(def)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	name := "Updated Color"
	req := &MetafieldDefinitionUpdateRequest{
		Name: &name,
	}

	def, err := client.UpdateMetafieldDefinition(context.Background(), "md_123", req)
	if err != nil {
		t.Fatalf("UpdateMetafieldDefinition failed: %v", err)
	}

	if def.Name != "Updated Color" {
		t.Errorf("Unexpected metafield definition name: %s", def.Name)
	}
}

func TestMetafieldDefinitionsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/metafield_definitions/md_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteMetafieldDefinition(context.Background(), "md_123")
	if err != nil {
		t.Fatalf("DeleteMetafieldDefinition failed: %v", err)
	}
}

func TestListMetafieldDefinitionsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/metafield_definitions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters
		query := r.URL.Query()
		if page := query.Get("page"); page != "2" {
			t.Errorf("Expected page=2, got %s", page)
		}
		if pageSize := query.Get("page_size"); pageSize != "10" {
			t.Errorf("Expected page_size=10, got %s", pageSize)
		}
		if ownerType := query.Get("owner_type"); ownerType != "product" {
			t.Errorf("Expected owner_type=product, got %s", ownerType)
		}
		if namespace := query.Get("namespace"); namespace != "custom" {
			t.Errorf("Expected namespace=custom, got %s", namespace)
		}

		resp := MetafieldDefinitionsListResponse{
			Items: []MetafieldDefinition{
				{ID: "md_789", Name: "Weight", Namespace: "custom", Key: "weight", Type: "number_decimal", OwnerType: "product"},
			},
			Page:       2,
			PageSize:   10,
			TotalCount: 15,
			HasMore:    true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &MetafieldDefinitionsListOptions{
		Page:      2,
		PageSize:  10,
		OwnerType: "product",
		Namespace: "custom",
	}

	defs, err := client.ListMetafieldDefinitions(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListMetafieldDefinitions failed: %v", err)
	}

	if len(defs.Items) != 1 {
		t.Errorf("Expected 1 metafield definition, got %d", len(defs.Items))
	}
	if defs.Page != 2 {
		t.Errorf("Expected page 2, got %d", defs.Page)
	}
	if defs.PageSize != 10 {
		t.Errorf("Expected page_size 10, got %d", defs.PageSize)
	}
	if !defs.HasMore {
		t.Error("Expected HasMore to be true")
	}
}

func TestUpdateMetafieldDefinitionEmptyID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{name: "empty string", id: ""},
		{name: "whitespace only", id: "   "},
		{name: "tab character", id: "\t"},
		{name: "newline", id: "\n"},
	}

	client := NewClient("token")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := "Updated Name"
			req := &MetafieldDefinitionUpdateRequest{
				Name: &name,
			}

			_, err := client.UpdateMetafieldDefinition(context.Background(), tt.id, req)
			if err == nil {
				t.Error("Expected error for empty/whitespace ID, got nil")
			}
			if err != nil && err.Error() != "metafield definition id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteMetafieldDefinitionEmptyID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{name: "empty string", id: ""},
		{name: "whitespace only", id: "   "},
		{name: "tab character", id: "\t"},
		{name: "newline", id: "\n"},
	}

	client := NewClient("token")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.DeleteMetafieldDefinition(context.Background(), tt.id)
			if err == nil {
				t.Error("Expected error for empty/whitespace ID, got nil")
			}
			if err != nil && err.Error() != "metafield definition id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreateMetafieldDefinitionWithValidations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var reqBody MetafieldDefinitionCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify validations were sent
		if len(reqBody.Validations) != 2 {
			t.Errorf("Expected 2 validations, got %d", len(reqBody.Validations))
		}
		if reqBody.Validations[0].Name != "min" {
			t.Errorf("Expected first validation name 'min', got %s", reqBody.Validations[0].Name)
		}
		if reqBody.Validations[1].Name != "max" {
			t.Errorf("Expected second validation name 'max', got %s", reqBody.Validations[1].Name)
		}

		def := MetafieldDefinition{
			ID:        "md_validated",
			Name:      "Price",
			Namespace: "custom",
			Key:       "price",
			Type:      "number_decimal",
			OwnerType: "product",
			Validations: []Validation{
				{Name: "min", Type: "number", Value: "0"},
				{Name: "max", Type: "number", Value: "10000"},
			},
		}
		_ = json.NewEncoder(w).Encode(def)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MetafieldDefinitionCreateRequest{
		Name:      "Price",
		Namespace: "custom",
		Key:       "price",
		Type:      "number_decimal",
		OwnerType: "product",
		Validations: []Validation{
			{Name: "min", Type: "number", Value: "0"},
			{Name: "max", Type: "number", Value: "10000"},
		},
	}

	def, err := client.CreateMetafieldDefinition(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateMetafieldDefinition failed: %v", err)
	}

	if def.ID != "md_validated" {
		t.Errorf("Unexpected metafield definition ID: %s", def.ID)
	}
	if len(def.Validations) != 2 {
		t.Errorf("Expected 2 validations in response, got %d", len(def.Validations))
	}
}

func TestMetafieldDefinitionsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Test List API error
	_, err := client.ListMetafieldDefinitions(context.Background(), nil)
	if err == nil {
		t.Error("Expected error from ListMetafieldDefinitions")
	}

	// Test Get API error
	_, err = client.GetMetafieldDefinition(context.Background(), "md_123")
	if err == nil {
		t.Error("Expected error from GetMetafieldDefinition")
	}

	// Test Create API error
	req := &MetafieldDefinitionCreateRequest{
		Name:      "Test",
		Namespace: "custom",
		Key:       "test",
		Type:      "single_line_text_field",
		OwnerType: "product",
	}
	_, err = client.CreateMetafieldDefinition(context.Background(), req)
	if err == nil {
		t.Error("Expected error from CreateMetafieldDefinition")
	}

	// Test Update API error
	name := "Updated"
	updateReq := &MetafieldDefinitionUpdateRequest{Name: &name}
	_, err = client.UpdateMetafieldDefinition(context.Background(), "md_123", updateReq)
	if err == nil {
		t.Error("Expected error from UpdateMetafieldDefinition")
	}

	// Test Delete API error
	err = client.DeleteMetafieldDefinition(context.Background(), "md_123")
	if err == nil {
		t.Error("Expected error from DeleteMetafieldDefinition")
	}
}

func TestMetafieldDefinitionsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetMetafieldDefinition(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent metafield definition")
	}
}

func TestUpdateMetafieldDefinitionWithDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/metafield_definitions/md_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var reqBody MetafieldDefinitionUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify both name and description were sent
		if reqBody.Name == nil {
			t.Error("Expected name in request body")
		} else if *reqBody.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %s", *reqBody.Name)
		}
		if reqBody.Description == nil {
			t.Error("Expected description in request body")
		} else if *reqBody.Description != "Updated description" {
			t.Errorf("Expected description 'Updated description', got %s", *reqBody.Description)
		}

		def := MetafieldDefinition{
			ID:          "md_123",
			Name:        "Updated Name",
			Description: "Updated description",
			Namespace:   "custom",
			Key:         "test",
		}
		_ = json.NewEncoder(w).Encode(def)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	name := "Updated Name"
	desc := "Updated description"
	req := &MetafieldDefinitionUpdateRequest{
		Name:        &name,
		Description: &desc,
	}

	def, err := client.UpdateMetafieldDefinition(context.Background(), "md_123", req)
	if err != nil {
		t.Fatalf("UpdateMetafieldDefinition failed: %v", err)
	}

	if def.Name != "Updated Name" {
		t.Errorf("Unexpected name: %s", def.Name)
	}
	if def.Description != "Updated description" {
		t.Errorf("Unexpected description: %s", def.Description)
	}
}

func TestCreateMetafieldDefinitionRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/metafield_definitions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var reqBody MetafieldDefinitionCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify all fields were sent correctly
		if reqBody.Name != "Test Field" {
			t.Errorf("Expected name 'Test Field', got %s", reqBody.Name)
		}
		if reqBody.Namespace != "my_namespace" {
			t.Errorf("Expected namespace 'my_namespace', got %s", reqBody.Namespace)
		}
		if reqBody.Key != "test_key" {
			t.Errorf("Expected key 'test_key', got %s", reqBody.Key)
		}
		if reqBody.Description != "A test description" {
			t.Errorf("Expected description 'A test description', got %s", reqBody.Description)
		}
		if reqBody.Type != "multi_line_text_field" {
			t.Errorf("Expected type 'multi_line_text_field', got %s", reqBody.Type)
		}
		if reqBody.OwnerType != "order" {
			t.Errorf("Expected owner_type 'order', got %s", reqBody.OwnerType)
		}

		def := MetafieldDefinition{
			ID:          "md_created",
			Name:        reqBody.Name,
			Namespace:   reqBody.Namespace,
			Key:         reqBody.Key,
			Description: reqBody.Description,
			Type:        reqBody.Type,
			OwnerType:   reqBody.OwnerType,
		}
		_ = json.NewEncoder(w).Encode(def)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MetafieldDefinitionCreateRequest{
		Name:        "Test Field",
		Namespace:   "my_namespace",
		Key:         "test_key",
		Description: "A test description",
		Type:        "multi_line_text_field",
		OwnerType:   "order",
	}

	def, err := client.CreateMetafieldDefinition(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateMetafieldDefinition failed: %v", err)
	}

	if def.ID != "md_created" {
		t.Errorf("Unexpected ID: %s", def.ID)
	}
	if def.OwnerType != "order" {
		t.Errorf("Unexpected owner_type: %s", def.OwnerType)
	}
}

func TestListMetafieldDefinitionsPartialOptions(t *testing.T) {
	tests := []struct {
		name           string
		opts           *MetafieldDefinitionsListOptions
		expectedParams map[string]string
	}{
		{
			name: "only page",
			opts: &MetafieldDefinitionsListOptions{Page: 3},
			expectedParams: map[string]string{
				"page": "3",
			},
		},
		{
			name: "only page_size",
			opts: &MetafieldDefinitionsListOptions{PageSize: 50},
			expectedParams: map[string]string{
				"page_size": "50",
			},
		},
		{
			name: "only owner_type",
			opts: &MetafieldDefinitionsListOptions{OwnerType: "variant"},
			expectedParams: map[string]string{
				"owner_type": "variant",
			},
		},
		{
			name: "only namespace",
			opts: &MetafieldDefinitionsListOptions{Namespace: "inventory"},
			expectedParams: map[string]string{
				"namespace": "inventory",
			},
		},
		{
			name:           "zero values ignored",
			opts:           &MetafieldDefinitionsListOptions{Page: 0, PageSize: 0},
			expectedParams: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()

				for key, expected := range tt.expectedParams {
					if got := query.Get(key); got != expected {
						t.Errorf("Expected %s=%s, got %s", key, expected, got)
					}
				}

				// Verify params NOT in expectedParams are not set
				for _, key := range []string{"page", "page_size", "owner_type", "namespace"} {
					if _, exists := tt.expectedParams[key]; !exists {
						if got := query.Get(key); got != "" {
							t.Errorf("Expected %s to be empty, got %s", key, got)
						}
					}
				}

				resp := MetafieldDefinitionsListResponse{
					Items: []MetafieldDefinition{},
				}
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient("token")
			client.BaseURL = server.URL
			client.SetUseOpenAPI(false)

			_, err := client.ListMetafieldDefinitions(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("ListMetafieldDefinitions failed: %v", err)
			}
		})
	}
}

func TestMetafieldDefinitionVerifyHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization 'Bearer test-token', got %s", auth)
		}

		// Verify content type for POST/PUT
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
			}
		}

		def := MetafieldDefinition{ID: "md_test"}
		_ = json.NewEncoder(w).Encode(def)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Test GET
	_, _ = client.GetMetafieldDefinition(context.Background(), "md_test")

	// Test POST
	req := &MetafieldDefinitionCreateRequest{
		Name:      "Test",
		Namespace: "custom",
		Key:       "test",
		Type:      "single_line_text_field",
		OwnerType: "product",
	}
	_, _ = client.CreateMetafieldDefinition(context.Background(), req)
}
