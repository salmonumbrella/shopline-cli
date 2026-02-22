package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductReviewsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/product_reviews" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ProductReviewsListResponse{
			Items: []ProductReview{
				{
					ID:           "rev_123",
					ProductID:    "prod_456",
					CustomerName: "John Doe",
					Rating:       5,
					Title:        "Great product!",
					Content:      "Very satisfied with this purchase.",
					Status:       "approved",
					Verified:     true,
				},
				{
					ID:           "rev_456",
					ProductID:    "prod_789",
					CustomerName: "Jane Smith",
					Rating:       4,
					Title:        "Good quality",
					Content:      "Would recommend.",
					Status:       "approved",
					Verified:     false,
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

	reviews, err := client.ListProductReviews(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListProductReviews failed: %v", err)
	}

	if len(reviews.Items) != 2 {
		t.Errorf("Expected 2 reviews, got %d", len(reviews.Items))
	}
	if reviews.Items[0].ID != "rev_123" {
		t.Errorf("Unexpected review ID: %s", reviews.Items[0].ID)
	}
	if reviews.Items[0].Rating != 5 {
		t.Errorf("Unexpected rating: %d", reviews.Items[0].Rating)
	}
}

func TestProductReviewsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("product_id") != "prod_123" {
			t.Errorf("Expected product_id=prod_123, got %s", r.URL.Query().Get("product_id"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("rating") != "5" {
			t.Errorf("Expected rating=5, got %s", r.URL.Query().Get("rating"))
		}

		resp := ProductReviewsListResponse{
			Items: []ProductReview{
				{ID: "rev_123", ProductID: "prod_123", Rating: 5},
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

	opts := &ProductReviewsListOptions{
		Page:      2,
		ProductID: "prod_123",
		Rating:    5,
	}
	reviews, err := client.ListProductReviews(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListProductReviews failed: %v", err)
	}

	if len(reviews.Items) != 1 {
		t.Errorf("Expected 1 review, got %d", len(reviews.Items))
	}
}

func TestProductReviewsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/product_reviews/rev_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		review := ProductReview{
			ID:           "rev_123",
			ProductID:    "prod_456",
			CustomerID:   "cust_789",
			CustomerName: "John Doe",
			Rating:       5,
			Title:        "Great product!",
			Content:      "Very satisfied with this purchase.",
			Status:       "approved",
			Verified:     true,
			HelpfulCount: 10,
		}
		_ = json.NewEncoder(w).Encode(review)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	review, err := client.GetProductReview(context.Background(), "rev_123")
	if err != nil {
		t.Fatalf("GetProductReview failed: %v", err)
	}

	if review.ID != "rev_123" {
		t.Errorf("Unexpected review ID: %s", review.ID)
	}
	if review.Rating != 5 {
		t.Errorf("Unexpected rating: %d", review.Rating)
	}
	if !review.Verified {
		t.Error("Expected verified to be true")
	}
}

func TestProductReviewsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/product_reviews" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductReviewCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.ProductID != "prod_123" {
			t.Errorf("Unexpected product ID: %s", req.ProductID)
		}
		if req.Rating != 5 {
			t.Errorf("Unexpected rating: %d", req.Rating)
		}

		review := ProductReview{
			ID:           "rev_new",
			ProductID:    req.ProductID,
			CustomerName: req.CustomerName,
			Rating:       req.Rating,
			Title:        req.Title,
			Content:      req.Content,
			Status:       "pending",
		}
		_ = json.NewEncoder(w).Encode(review)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductReviewCreateRequest{
		ProductID:    "prod_123",
		CustomerName: "John Doe",
		Rating:       5,
		Title:        "Excellent!",
		Content:      "Best purchase ever.",
	}
	review, err := client.CreateProductReview(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateProductReview failed: %v", err)
	}

	if review.ID != "rev_new" {
		t.Errorf("Unexpected review ID: %s", review.ID)
	}
	if review.Rating != 5 {
		t.Errorf("Unexpected rating: %d", review.Rating)
	}
}

func TestProductReviewsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/product_reviews/rev_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProductReview(context.Background(), "rev_123")
	if err != nil {
		t.Fatalf("DeleteProductReview failed: %v", err)
	}
}

func TestGetProductReviewEmptyID(t *testing.T) {
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
			_, err := client.GetProductReview(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product review id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteProductReviewEmptyID(t *testing.T) {
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
			err := client.DeleteProductReview(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product review id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
