package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticlesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/articles" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ArticlesListResponse{
			Items: []Article{
				{
					ID:        "art_123",
					BlogID:    "blog_1",
					Title:     "First Post",
					Handle:    "first-post",
					Author:    "John Doe",
					Published: true,
				},
				{
					ID:        "art_456",
					BlogID:    "blog_1",
					Title:     "Second Post",
					Handle:    "second-post",
					Author:    "Jane Doe",
					Published: false,
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

	articles, err := client.ListArticles(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListArticles failed: %v", err)
	}

	if len(articles.Items) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles.Items))
	}
	if articles.Items[0].ID != "art_123" {
		t.Errorf("Unexpected article ID: %s", articles.Items[0].ID)
	}
	if articles.Items[0].Title != "First Post" {
		t.Errorf("Unexpected title: %s", articles.Items[0].Title)
	}
}

func TestArticlesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("blog_id") != "blog_1" {
			t.Errorf("Expected blog_id=blog_1, got %s", r.URL.Query().Get("blog_id"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("published") != "true" {
			t.Errorf("Expected published=true, got %s", r.URL.Query().Get("published"))
		}

		resp := ArticlesListResponse{
			Items: []Article{
				{ID: "art_123", BlogID: "blog_1", Published: true},
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

	published := true
	opts := &ArticlesListOptions{
		Page:      2,
		BlogID:    "blog_1",
		Published: &published,
	}
	articles, err := client.ListArticles(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListArticles failed: %v", err)
	}

	if len(articles.Items) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles.Items))
	}
}

func TestArticlesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/articles/art_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		article := Article{
			ID:        "art_123",
			BlogID:    "blog_1",
			Title:     "First Post",
			Handle:    "first-post",
			Author:    "John Doe",
			BodyHTML:  "<p>Hello World</p>",
			Published: true,
		}
		_ = json.NewEncoder(w).Encode(article)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	article, err := client.GetArticle(context.Background(), "art_123")
	if err != nil {
		t.Fatalf("GetArticle failed: %v", err)
	}

	if article.ID != "art_123" {
		t.Errorf("Unexpected article ID: %s", article.ID)
	}
	if article.Title != "First Post" {
		t.Errorf("Unexpected title: %s", article.Title)
	}
	if article.BodyHTML != "<p>Hello World</p>" {
		t.Errorf("Unexpected body HTML: %s", article.BodyHTML)
	}
}

func TestGetArticleEmptyID(t *testing.T) {
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
			_, err := client.GetArticle(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "article id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestArticlesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/articles" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ArticleCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.BlogID != "blog_1" {
			t.Errorf("Unexpected blog_id: %s", req.BlogID)
		}
		if req.Title != "New Article" {
			t.Errorf("Unexpected title: %s", req.Title)
		}
		if req.BodyHTML != "<p>Content</p>" {
			t.Errorf("Unexpected body: %s", req.BodyHTML)
		}

		article := Article{
			ID:        "art_new",
			BlogID:    req.BlogID,
			Title:     req.Title,
			BodyHTML:  req.BodyHTML,
			Author:    req.Author,
			Published: req.Published,
		}
		_ = json.NewEncoder(w).Encode(article)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ArticleCreateRequest{
		BlogID:    "blog_1",
		Title:     "New Article",
		BodyHTML:  "<p>Content</p>",
		Author:    "John Doe",
		Published: true,
	}
	article, err := client.CreateArticle(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}

	if article.ID != "art_new" {
		t.Errorf("Unexpected article ID: %s", article.ID)
	}
	if article.Title != "New Article" {
		t.Errorf("Unexpected title: %s", article.Title)
	}
}

func TestArticlesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/articles/art_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ArticleUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Title != "Updated Title" {
			t.Errorf("Unexpected title: %s", req.Title)
		}
		if req.Author != "Jane Doe" {
			t.Errorf("Unexpected author: %s", req.Author)
		}
		if req.BodyHTML != "<p>Updated content</p>" {
			t.Errorf("Unexpected body: %s", req.BodyHTML)
		}
		if req.Published == nil || *req.Published != false {
			t.Errorf("Unexpected published value")
		}

		article := Article{
			ID:        "art_123",
			BlogID:    "blog_1",
			Title:     req.Title,
			BodyHTML:  req.BodyHTML,
			Author:    req.Author,
			Published: *req.Published,
		}
		_ = json.NewEncoder(w).Encode(article)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	published := false
	req := &ArticleUpdateRequest{
		Title:     "Updated Title",
		Author:    "Jane Doe",
		BodyHTML:  "<p>Updated content</p>",
		Published: &published,
	}
	article, err := client.UpdateArticle(context.Background(), "art_123", req)
	if err != nil {
		t.Fatalf("UpdateArticle failed: %v", err)
	}

	if article.ID != "art_123" {
		t.Errorf("Unexpected article ID: %s", article.ID)
	}
	if article.Title != "Updated Title" {
		t.Errorf("Unexpected title: %s", article.Title)
	}
	if article.Author != "Jane Doe" {
		t.Errorf("Unexpected author: %s", article.Author)
	}
	if article.Published != false {
		t.Errorf("Expected published=false, got %v", article.Published)
	}
}

func TestUpdateArticleEmptyID(t *testing.T) {
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
			_, err := client.UpdateArticle(context.Background(), tc.id, &ArticleUpdateRequest{Title: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "article id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestArticlesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/articles/art_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteArticle(context.Background(), "art_123")
	if err != nil {
		t.Fatalf("DeleteArticle failed: %v", err)
	}
}

func TestDeleteArticleEmptyID(t *testing.T) {
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
			err := client.DeleteArticle(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "article id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
