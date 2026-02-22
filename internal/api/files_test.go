package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/files" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := FilesListResponse{
			Items: []File{
				{
					ID:       "file_123",
					Filename: "product-image.jpg",
					MimeType: "image/jpeg",
					FileSize: 102400,
					URL:      "https://cdn.example.com/files/product-image.jpg",
					Status:   FileStatusReady,
				},
				{
					ID:       "file_456",
					Filename: "catalog.pdf",
					MimeType: "application/pdf",
					FileSize: 512000,
					URL:      "https://cdn.example.com/files/catalog.pdf",
					Status:   FileStatusReady,
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

	files, err := client.ListFiles(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files.Items) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files.Items))
	}
	if files.Items[0].ID != "file_123" {
		t.Errorf("Unexpected file ID: %s", files.Items[0].ID)
	}
	if files.Items[0].Filename != "product-image.jpg" {
		t.Errorf("Unexpected filename: %s", files.Items[0].Filename)
	}
}

func TestFilesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("content_type") != "image" {
			t.Errorf("Expected content_type=image, got %s", r.URL.Query().Get("content_type"))
		}
		if r.URL.Query().Get("status") != "ready" {
			t.Errorf("Expected status=ready, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := FilesListResponse{
			Items: []File{
				{ID: "file_123", Filename: "image.jpg", Status: FileStatusReady},
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

	opts := &FilesListOptions{
		Page:        2,
		ContentType: "image",
		Status:      "ready",
	}
	files, err := client.ListFiles(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files.Items) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files.Items))
	}
}

func TestFilesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/files/file_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		file := File{
			ID:       "file_123",
			Filename: "product-image.jpg",
			MimeType: "image/jpeg",
			FileSize: 102400,
			URL:      "https://cdn.example.com/files/product-image.jpg",
			Alt:      "Product image",
			Status:   FileStatusReady,
			Width:    800,
			Height:   600,
		}
		_ = json.NewEncoder(w).Encode(file)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	file, err := client.GetFile(context.Background(), "file_123")
	if err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}

	if file.ID != "file_123" {
		t.Errorf("Unexpected file ID: %s", file.ID)
	}
	if file.Filename != "product-image.jpg" {
		t.Errorf("Unexpected filename: %s", file.Filename)
	}
	if file.FileSize != 102400 {
		t.Errorf("Unexpected file size: %d", file.FileSize)
	}
}

func TestGetFileEmptyID(t *testing.T) {
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
			_, err := client.GetFile(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "file id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestFilesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/files" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req FileCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Filename != "new-image.jpg" {
			t.Errorf("Unexpected filename: %s", req.Filename)
		}
		if req.URL != "https://example.com/image.jpg" {
			t.Errorf("Unexpected URL: %s", req.URL)
		}

		file := File{
			ID:       "file_new",
			Filename: req.Filename,
			URL:      req.URL,
			Alt:      req.Alt,
			Status:   FileStatusPending,
		}
		_ = json.NewEncoder(w).Encode(file)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &FileCreateRequest{
		Filename: "new-image.jpg",
		URL:      "https://example.com/image.jpg",
		Alt:      "New product image",
	}
	file, err := client.CreateFile(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	if file.ID != "file_new" {
		t.Errorf("Unexpected file ID: %s", file.ID)
	}
	if file.Filename != "new-image.jpg" {
		t.Errorf("Unexpected filename: %s", file.Filename)
	}
}

func TestFilesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/files/file_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req FileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Filename != "renamed-image.jpg" {
			t.Errorf("Unexpected filename: %s", req.Filename)
		}
		if req.Alt != "Updated alt text" {
			t.Errorf("Unexpected alt: %s", req.Alt)
		}

		file := File{
			ID:       "file_123",
			Filename: req.Filename,
			Alt:      req.Alt,
			Status:   FileStatusReady,
		}
		_ = json.NewEncoder(w).Encode(file)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &FileUpdateRequest{
		Filename: "renamed-image.jpg",
		Alt:      "Updated alt text",
	}
	file, err := client.UpdateFile(context.Background(), "file_123", req)
	if err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}

	if file.ID != "file_123" {
		t.Errorf("Unexpected file ID: %s", file.ID)
	}
	if file.Filename != "renamed-image.jpg" {
		t.Errorf("Unexpected filename: %s", file.Filename)
	}
}

func TestUpdateFileEmptyID(t *testing.T) {
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
			_, err := client.UpdateFile(context.Background(), tc.id, &FileUpdateRequest{Filename: "test.jpg"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "file id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestFilesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/files/file_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteFile(context.Background(), "file_123")
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
}

func TestDeleteFileEmptyID(t *testing.T) {
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
			err := client.DeleteFile(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "file id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
