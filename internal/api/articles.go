package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Article represents a Shopline blog article.
type Article struct {
	ID          string    `json:"id"`
	BlogID      string    `json:"blog_id"`
	Title       string    `json:"title"`
	Handle      string    `json:"handle"`
	Author      string    `json:"author"`
	BodyHTML    string    `json:"body_html"`
	SummaryHTML string    `json:"summary_html"`
	Tags        string    `json:"tags"`
	Image       *Image    `json:"image"`
	Published   bool      `json:"published"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Image represents an article image.
type Image struct {
	Src    string `json:"src"`
	Alt    string `json:"alt"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// ArticlesListOptions contains options for listing articles.
type ArticlesListOptions struct {
	Page      int
	PageSize  int
	BlogID    string
	Published *bool
}

// ArticlesListResponse is the paginated response for articles.
type ArticlesListResponse = ListResponse[Article]

// ArticleCreateRequest contains the data for creating an article.
type ArticleCreateRequest struct {
	BlogID    string `json:"blog_id"`
	Title     string `json:"title"`
	Author    string `json:"author,omitempty"`
	BodyHTML  string `json:"body_html"`
	Tags      string `json:"tags,omitempty"`
	Published bool   `json:"published"`
}

// ArticleUpdateRequest contains the data for updating an article.
type ArticleUpdateRequest struct {
	Title     string `json:"title,omitempty"`
	Author    string `json:"author,omitempty"`
	BodyHTML  string `json:"body_html,omitempty"`
	Tags      string `json:"tags,omitempty"`
	Published *bool  `json:"published,omitempty"`
}

// ListArticles retrieves a list of articles.
func (c *Client) ListArticles(ctx context.Context, opts *ArticlesListOptions) (*ArticlesListResponse, error) {
	path := "/articles"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("blog_id", opts.BlogID).
			BoolPtr("published", opts.Published).
			Build()
	}

	var resp ArticlesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetArticle retrieves a single article by ID.
func (c *Client) GetArticle(ctx context.Context, id string) (*Article, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("article id is required")
	}
	var article Article
	if err := c.Get(ctx, fmt.Sprintf("/articles/%s", id), &article); err != nil {
		return nil, err
	}
	return &article, nil
}

// CreateArticle creates a new article.
func (c *Client) CreateArticle(ctx context.Context, req *ArticleCreateRequest) (*Article, error) {
	var article Article
	if err := c.Post(ctx, "/articles", req, &article); err != nil {
		return nil, err
	}
	return &article, nil
}

// UpdateArticle updates an existing article.
func (c *Client) UpdateArticle(ctx context.Context, id string, req *ArticleUpdateRequest) (*Article, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("article id is required")
	}
	var article Article
	if err := c.Put(ctx, fmt.Sprintf("/articles/%s", id), req, &article); err != nil {
		return nil, err
	}
	return &article, nil
}

// DeleteArticle deletes an article.
func (c *Client) DeleteArticle(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("article id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/articles/%s", id))
}
