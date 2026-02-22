package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLabelsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/labels" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := LabelsListResponse{
			Items: []Label{
				{
					ID:          "lbl_123",
					Name:        "Sale",
					Description: "Products on sale",
					Color:       "#ff0000",
					Active:      true,
				},
				{
					ID:          "lbl_456",
					Name:        "New",
					Description: "New arrivals",
					Color:       "#00ff00",
					Active:      true,
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

	labels, err := client.ListLabels(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListLabels failed: %v", err)
	}

	if len(labels.Items) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(labels.Items))
	}
	if labels.Items[0].ID != "lbl_123" {
		t.Errorf("Unexpected label ID: %s", labels.Items[0].ID)
	}
	if labels.Items[0].Name != "Sale" {
		t.Errorf("Unexpected name: %s", labels.Items[0].Name)
	}
}

func TestLabelsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", r.URL.Query().Get("active"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := LabelsListResponse{
			Items: []Label{
				{ID: "lbl_123", Name: "Sale", Active: true},
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

	active := true
	opts := &LabelsListOptions{
		Page:   2,
		Active: &active,
	}
	labels, err := client.ListLabels(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListLabels failed: %v", err)
	}

	if len(labels.Items) != 1 {
		t.Errorf("Expected 1 label, got %d", len(labels.Items))
	}
}

func TestLabelsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/labels/lbl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		label := Label{
			ID:          "lbl_123",
			Name:        "Sale",
			Description: "Products on sale",
			Color:       "#ff0000",
			Icon:        "tag",
			Active:      true,
		}
		_ = json.NewEncoder(w).Encode(label)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	label, err := client.GetLabel(context.Background(), "lbl_123")
	if err != nil {
		t.Fatalf("GetLabel failed: %v", err)
	}

	if label.ID != "lbl_123" {
		t.Errorf("Unexpected label ID: %s", label.ID)
	}
	if label.Name != "Sale" {
		t.Errorf("Unexpected name: %s", label.Name)
	}
	if label.Color != "#ff0000" {
		t.Errorf("Unexpected color: %s", label.Color)
	}
}

func TestGetLabelEmptyID(t *testing.T) {
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
			_, err := client.GetLabel(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "label id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestLabelsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/labels" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req LabelCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Sale" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Color != "#ff0000" {
			t.Errorf("Unexpected color: %s", req.Color)
		}

		label := Label{
			ID:          "lbl_new",
			Name:        req.Name,
			Description: req.Description,
			Color:       req.Color,
			Active:      req.Active,
		}
		_ = json.NewEncoder(w).Encode(label)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &LabelCreateRequest{
		Name:        "Sale",
		Description: "Products on sale",
		Color:       "#ff0000",
		Active:      true,
	}
	label, err := client.CreateLabel(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateLabel failed: %v", err)
	}

	if label.ID != "lbl_new" {
		t.Errorf("Unexpected label ID: %s", label.ID)
	}
	if label.Name != "Sale" {
		t.Errorf("Unexpected name: %s", label.Name)
	}
}

func TestLabelsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/labels/lbl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req LabelUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Clearance" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Active == nil || *req.Active != false {
			t.Errorf("Unexpected active value")
		}

		label := Label{
			ID:     "lbl_123",
			Name:   req.Name,
			Color:  req.Color,
			Active: *req.Active,
		}
		_ = json.NewEncoder(w).Encode(label)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	active := false
	req := &LabelUpdateRequest{
		Name:   "Clearance",
		Color:  "#0000ff",
		Active: &active,
	}
	label, err := client.UpdateLabel(context.Background(), "lbl_123", req)
	if err != nil {
		t.Fatalf("UpdateLabel failed: %v", err)
	}

	if label.ID != "lbl_123" {
		t.Errorf("Unexpected label ID: %s", label.ID)
	}
	if label.Name != "Clearance" {
		t.Errorf("Unexpected name: %s", label.Name)
	}
}

func TestUpdateLabelEmptyID(t *testing.T) {
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
			_, err := client.UpdateLabel(context.Background(), tc.id, &LabelUpdateRequest{Name: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "label id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestLabelsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/labels/lbl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteLabel(context.Background(), "lbl_123")
	if err != nil {
		t.Fatalf("DeleteLabel failed: %v", err)
	}
}

func TestDeleteLabelEmptyID(t *testing.T) {
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
			err := client.DeleteLabel(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "label id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
