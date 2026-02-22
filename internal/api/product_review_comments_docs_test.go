package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductReviewCommentsDocs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/product_review_comments":
			if got := r.URL.Query().Get("page"); got != "2" {
				t.Fatalf("expected page=2, got %q", got)
			}
			if got := r.URL.Query().Get("page_size"); got != "50" {
				t.Fatalf("expected page_size=50, got %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
			return
		case r.Method == http.MethodGet && r.URL.Path == "/product_review_comments/cmt_1":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "cmt_1"})
			return
		case r.Method == http.MethodPost && r.URL.Path == "/product_review_comments":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["ok"] != true {
				t.Fatalf("expected ok=true, got %v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "cmt_new"})
			return
		case r.Method == http.MethodPut && r.URL.Path == "/product_review_comments/cmt_1":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["ok"] != true {
				t.Fatalf("expected ok=true, got %v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "cmt_1", "updated": true})
			return
		case r.Method == http.MethodDelete && r.URL.Path == "/product_review_comments/cmt_1":
			_ = json.NewEncoder(w).Encode(map[string]any{"deleted": true})
			return
		case r.Method == http.MethodPost && r.URL.Path == "/product_review_comments/bulk":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			return
		case r.Method == http.MethodPut && r.URL.Path == "/product_review_comments/bulk":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			return
		case r.Method == http.MethodDelete && r.URL.Path == "/product_review_comments/bulk":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["ok"] != true {
				t.Fatalf("expected ok=true, got %v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			return
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListProductReviewComments(context.Background(), &ProductReviewCommentsListOptions{Page: 2, PageSize: 50})
	if err != nil {
		t.Fatalf("ListProductReviewComments failed: %v", err)
	}
	_, err = client.GetProductReviewComment(context.Background(), "cmt_1")
	if err != nil {
		t.Fatalf("GetProductReviewComment failed: %v", err)
	}
	_, err = client.CreateProductReviewComment(context.Background(), map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("CreateProductReviewComment failed: %v", err)
	}
	_, err = client.UpdateProductReviewComment(context.Background(), "cmt_1", map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("UpdateProductReviewComment failed: %v", err)
	}
	_, err = client.DeleteProductReviewComment(context.Background(), "cmt_1")
	if err != nil {
		t.Fatalf("DeleteProductReviewComment failed: %v", err)
	}
	_, err = client.BulkCreateProductReviewComments(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("BulkCreateProductReviewComments failed: %v", err)
	}
	_, err = client.BulkUpdateProductReviewComments(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("BulkUpdateProductReviewComments failed: %v", err)
	}
	_, err = client.BulkDeleteProductReviewComments(context.Background(), map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("BulkDeleteProductReviewComments failed: %v", err)
	}
}

func TestProductReviewCommentsDocsEmptyID(t *testing.T) {
	client := NewClient("token")
	client.SetUseOpenAPI(false)

	if _, err := client.GetProductReviewComment(context.Background(), " "); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.UpdateProductReviewComment(context.Background(), " ", map[string]any{}); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.DeleteProductReviewComment(context.Background(), " "); err == nil {
		t.Fatalf("expected error")
	}
}
