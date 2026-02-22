package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssetsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/themes/thm_123/assets" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := AssetsListResponse{
			Items: []Asset{
				{Key: "layout/theme.liquid", ThemeID: "thm_123", ContentType: "text/x-liquid"},
				{Key: "templates/index.liquid", ThemeID: "thm_123", ContentType: "text/x-liquid"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	assets, err := client.ListAssets(context.Background(), "thm_123")
	if err != nil {
		t.Fatalf("ListAssets failed: %v", err)
	}

	if len(assets.Items) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(assets.Items))
	}
	if assets.Items[0].Key != "layout/theme.liquid" {
		t.Errorf("Unexpected asset key: %s", assets.Items[0].Key)
	}
}

func TestListAssetsEmptyThemeID(t *testing.T) {
	client := NewClient("token")

	_, err := client.ListAssets(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty theme ID, got nil")
	}
	if err != nil && err.Error() != "theme id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestAssetsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "layout/theme.liquid" {
			t.Errorf("Unexpected key: %s", r.URL.Query().Get("key"))
		}

		asset := Asset{Key: "layout/theme.liquid", ThemeID: "thm_123", Value: "{{ content_for_layout }}"}
		_ = json.NewEncoder(w).Encode(asset)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	asset, err := client.GetAsset(context.Background(), "thm_123", "layout/theme.liquid")
	if err != nil {
		t.Fatalf("GetAsset failed: %v", err)
	}

	if asset.Key != "layout/theme.liquid" {
		t.Errorf("Unexpected asset key: %s", asset.Key)
	}
}

func TestGetAssetEmptyKey(t *testing.T) {
	client := NewClient("token")

	_, err := client.GetAsset(context.Background(), "thm_123", "")
	if err == nil {
		t.Error("Expected error for empty key, got nil")
	}
	if err != nil && err.Error() != "asset key is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestAssetsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		asset := Asset{Key: "layout/theme.liquid", Value: "new content"}
		_ = json.NewEncoder(w).Encode(asset)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &AssetUpdateRequest{
		Key:   "layout/theme.liquid",
		Value: "new content",
	}

	asset, err := client.UpdateAsset(context.Background(), "thm_123", req)
	if err != nil {
		t.Fatalf("UpdateAsset failed: %v", err)
	}

	if asset.Value != "new content" {
		t.Errorf("Unexpected asset value: %s", asset.Value)
	}
}

func TestAssetsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Query().Get("key") != "assets/custom.css" {
			t.Errorf("Unexpected key: %s", r.URL.Query().Get("key"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteAsset(context.Background(), "thm_123", "assets/custom.css")
	if err != nil {
		t.Fatalf("DeleteAsset failed: %v", err)
	}
}
