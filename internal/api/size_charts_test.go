package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSizeChartsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/size_charts" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := SizeChartsListResponse{
			Items: []SizeChart{
				{
					ID:          "sc_123",
					Name:        "T-Shirt Sizes",
					Description: "Size chart for t-shirts",
					Unit:        "cm",
					Headers:     []string{"Chest", "Length", "Sleeve"},
					Active:      true,
				},
				{
					ID:          "sc_456",
					Name:        "Pants Sizes",
					Description: "Size chart for pants",
					Unit:        "inches",
					Headers:     []string{"Waist", "Inseam", "Hip"},
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

	sizeCharts, err := client.ListSizeCharts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListSizeCharts failed: %v", err)
	}

	if len(sizeCharts.Items) != 2 {
		t.Errorf("Expected 2 size charts, got %d", len(sizeCharts.Items))
	}
	if sizeCharts.Items[0].ID != "sc_123" {
		t.Errorf("Unexpected size chart ID: %s", sizeCharts.Items[0].ID)
	}
	if sizeCharts.Items[0].Name != "T-Shirt Sizes" {
		t.Errorf("Unexpected name: %s", sizeCharts.Items[0].Name)
	}
}

func TestSizeChartsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", r.URL.Query().Get("active"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := SizeChartsListResponse{
			Items: []SizeChart{
				{ID: "sc_123", Name: "T-Shirt Sizes", Active: true},
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
	opts := &SizeChartsListOptions{
		Page:   2,
		Active: &active,
	}
	sizeCharts, err := client.ListSizeCharts(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListSizeCharts failed: %v", err)
	}

	if len(sizeCharts.Items) != 1 {
		t.Errorf("Expected 1 size chart, got %d", len(sizeCharts.Items))
	}
}

func TestSizeChartsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/size_charts/sc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		sizeChart := SizeChart{
			ID:          "sc_123",
			Name:        "T-Shirt Sizes",
			Description: "Size chart for t-shirts",
			Unit:        "cm",
			Headers:     []string{"Chest", "Length", "Sleeve"},
			Rows: []SizeChartRow{
				{Size: "S", Values: []string{"86-91", "66", "33"}},
				{Size: "M", Values: []string{"91-97", "69", "34"}},
				{Size: "L", Values: []string{"97-102", "72", "35"}},
			},
			ProductIDs: []string{"prod_1", "prod_2"},
			Active:     true,
		}
		_ = json.NewEncoder(w).Encode(sizeChart)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	sizeChart, err := client.GetSizeChart(context.Background(), "sc_123")
	if err != nil {
		t.Fatalf("GetSizeChart failed: %v", err)
	}

	if sizeChart.ID != "sc_123" {
		t.Errorf("Unexpected size chart ID: %s", sizeChart.ID)
	}
	if sizeChart.Name != "T-Shirt Sizes" {
		t.Errorf("Unexpected name: %s", sizeChart.Name)
	}
	if len(sizeChart.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(sizeChart.Rows))
	}
	if sizeChart.Rows[0].Size != "S" {
		t.Errorf("Unexpected first row size: %s", sizeChart.Rows[0].Size)
	}
}

func TestGetSizeChartEmptyID(t *testing.T) {
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
			_, err := client.GetSizeChart(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "size chart id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSizeChartsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/size_charts" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req SizeChartCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Shoe Sizes" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Unit != "EU" {
			t.Errorf("Unexpected unit: %s", req.Unit)
		}

		sizeChart := SizeChart{
			ID:          "sc_new",
			Name:        req.Name,
			Description: req.Description,
			Unit:        req.Unit,
			Headers:     req.Headers,
			Rows:        req.Rows,
			Active:      req.Active,
		}
		_ = json.NewEncoder(w).Encode(sizeChart)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &SizeChartCreateRequest{
		Name:        "Shoe Sizes",
		Description: "Size chart for shoes",
		Unit:        "EU",
		Headers:     []string{"EU", "US", "UK"},
		Rows: []SizeChartRow{
			{Size: "38", Values: []string{"38", "7", "5"}},
			{Size: "39", Values: []string{"39", "8", "6"}},
		},
		Active: true,
	}
	sizeChart, err := client.CreateSizeChart(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSizeChart failed: %v", err)
	}

	if sizeChart.ID != "sc_new" {
		t.Errorf("Unexpected size chart ID: %s", sizeChart.ID)
	}
	if sizeChart.Name != "Shoe Sizes" {
		t.Errorf("Unexpected name: %s", sizeChart.Name)
	}
}

func TestSizeChartsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/size_charts/sc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req SizeChartUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Updated T-Shirt Sizes" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Active == nil || *req.Active != false {
			t.Errorf("Unexpected active value")
		}

		sizeChart := SizeChart{
			ID:     "sc_123",
			Name:   req.Name,
			Unit:   "cm",
			Active: *req.Active,
		}
		_ = json.NewEncoder(w).Encode(sizeChart)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	active := false
	req := &SizeChartUpdateRequest{
		Name:   "Updated T-Shirt Sizes",
		Active: &active,
	}
	sizeChart, err := client.UpdateSizeChart(context.Background(), "sc_123", req)
	if err != nil {
		t.Fatalf("UpdateSizeChart failed: %v", err)
	}

	if sizeChart.ID != "sc_123" {
		t.Errorf("Unexpected size chart ID: %s", sizeChart.ID)
	}
	if sizeChart.Name != "Updated T-Shirt Sizes" {
		t.Errorf("Unexpected name: %s", sizeChart.Name)
	}
}

func TestUpdateSizeChartEmptyID(t *testing.T) {
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
			_, err := client.UpdateSizeChart(context.Background(), tc.id, &SizeChartUpdateRequest{Name: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "size chart id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSizeChartsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/size_charts/sc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteSizeChart(context.Background(), "sc_123")
	if err != nil {
		t.Fatalf("DeleteSizeChart failed: %v", err)
	}
}

func TestDeleteSizeChartEmptyID(t *testing.T) {
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
			err := client.DeleteSizeChart(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "size chart id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
