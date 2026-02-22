package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMediasList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/medias" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := MediasListResponse{
			Items: []Media{
				{
					ID:        "med_123",
					ProductID: "prod_1",
					MediaType: MediaTypeImage,
					Src:       "https://example.com/image1.jpg",
					Alt:       "Product image 1",
					Width:     800,
					Height:    600,
				},
				{
					ID:        "med_456",
					ProductID: "prod_1",
					MediaType: MediaTypeVideo,
					Src:       "https://example.com/video1.mp4",
					Alt:       "Product video",
					Duration:  120,
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

	medias, err := client.ListMedias(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListMedias failed: %v", err)
	}

	if len(medias.Items) != 2 {
		t.Errorf("Expected 2 medias, got %d", len(medias.Items))
	}
	if medias.Items[0].ID != "med_123" {
		t.Errorf("Unexpected media ID: %s", medias.Items[0].ID)
	}
	if medias.Items[0].MediaType != MediaTypeImage {
		t.Errorf("Unexpected media type: %s", medias.Items[0].MediaType)
	}
}

func TestMediasListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("product_id") != "prod_1" {
			t.Errorf("Expected product_id=prod_1, got %s", r.URL.Query().Get("product_id"))
		}
		if r.URL.Query().Get("media_type") != "image" {
			t.Errorf("Expected media_type=image, got %s", r.URL.Query().Get("media_type"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := MediasListResponse{
			Items: []Media{
				{ID: "med_123", ProductID: "prod_1", MediaType: MediaTypeImage},
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

	opts := &MediasListOptions{
		Page:      2,
		ProductID: "prod_1",
		MediaType: "image",
	}
	medias, err := client.ListMedias(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListMedias failed: %v", err)
	}

	if len(medias.Items) != 1 {
		t.Errorf("Expected 1 media, got %d", len(medias.Items))
	}
}

func TestMediasGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/medias/med_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		media := Media{
			ID:        "med_123",
			ProductID: "prod_1",
			MediaType: MediaTypeImage,
			Src:       "https://example.com/image1.jpg",
			Alt:       "Product image 1",
			Width:     800,
			Height:    600,
			MimeType:  "image/jpeg",
			FileSize:  102400,
		}
		_ = json.NewEncoder(w).Encode(media)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	media, err := client.GetMedia(context.Background(), "med_123")
	if err != nil {
		t.Fatalf("GetMedia failed: %v", err)
	}

	if media.ID != "med_123" {
		t.Errorf("Unexpected media ID: %s", media.ID)
	}
	if media.Src != "https://example.com/image1.jpg" {
		t.Errorf("Unexpected src: %s", media.Src)
	}
	if media.Width != 800 {
		t.Errorf("Unexpected width: %d", media.Width)
	}
}

func TestGetMediaEmptyID(t *testing.T) {
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
			_, err := client.GetMedia(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "media id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMediasCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/medias" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req MediaCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.ProductID != "prod_1" {
			t.Errorf("Unexpected product_id: %s", req.ProductID)
		}
		if req.MediaType != MediaTypeImage {
			t.Errorf("Unexpected media_type: %s", req.MediaType)
		}
		if req.Src != "https://example.com/image.jpg" {
			t.Errorf("Unexpected src: %s", req.Src)
		}

		media := Media{
			ID:        "med_new",
			ProductID: req.ProductID,
			MediaType: req.MediaType,
			Src:       req.Src,
			Alt:       req.Alt,
		}
		_ = json.NewEncoder(w).Encode(media)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MediaCreateRequest{
		ProductID: "prod_1",
		MediaType: MediaTypeImage,
		Src:       "https://example.com/image.jpg",
		Alt:       "Product image",
	}
	media, err := client.CreateMedia(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateMedia failed: %v", err)
	}

	if media.ID != "med_new" {
		t.Errorf("Unexpected media ID: %s", media.ID)
	}
	if media.Src != "https://example.com/image.jpg" {
		t.Errorf("Unexpected src: %s", media.Src)
	}
}

func TestMediasUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/medias/med_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req MediaUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Alt != "Updated alt text" {
			t.Errorf("Unexpected alt: %s", req.Alt)
		}
		if req.Position != 2 {
			t.Errorf("Unexpected position: %d", req.Position)
		}

		media := Media{
			ID:        "med_123",
			ProductID: "prod_1",
			MediaType: MediaTypeImage,
			Alt:       req.Alt,
			Position:  req.Position,
		}
		_ = json.NewEncoder(w).Encode(media)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MediaUpdateRequest{
		Alt:      "Updated alt text",
		Position: 2,
	}
	media, err := client.UpdateMedia(context.Background(), "med_123", req)
	if err != nil {
		t.Fatalf("UpdateMedia failed: %v", err)
	}

	if media.ID != "med_123" {
		t.Errorf("Unexpected media ID: %s", media.ID)
	}
	if media.Alt != "Updated alt text" {
		t.Errorf("Unexpected alt: %s", media.Alt)
	}
}

func TestUpdateMediaEmptyID(t *testing.T) {
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
			_, err := client.UpdateMedia(context.Background(), tc.id, &MediaUpdateRequest{Alt: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "media id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMediasDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/medias/med_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteMedia(context.Background(), "med_123")
	if err != nil {
		t.Fatalf("DeleteMedia failed: %v", err)
	}
}

func TestDeleteMediaEmptyID(t *testing.T) {
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
			err := client.DeleteMedia(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "media id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
