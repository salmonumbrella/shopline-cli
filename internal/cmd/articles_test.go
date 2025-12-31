package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func TestArticlesCommandSetup(t *testing.T) {
	if articlesCmd.Use != "articles" {
		t.Errorf("expected Use 'articles', got %q", articlesCmd.Use)
	}
	if articlesCmd.Short != "Manage blog articles" {
		t.Errorf("expected Short 'Manage blog articles', got %q", articlesCmd.Short)
	}
}

func TestArticlesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List articles",
		"get":    "Get article details",
		"create": "Create an article",
		"delete": "Delete an article",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range articlesCmd.Commands() {
				if sub.Use == name || (len(sub.Use) > len(name) && sub.Use[:len(name)] == name) {
					found = true
					if sub.Short != short {
						t.Errorf("expected Short %q, got %q", short, sub.Short)
					}
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not found", name)
			}
		})
	}
}

func TestArticlesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_, err := getClient(cmd)
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

func TestArticlesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"blog-id", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := articlesListCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestArticlesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"blog-id", ""},
		{"title", ""},
		{"body", ""},
		{"author", ""},
		{"tags", ""},
		{"published", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := articlesCreateCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestArticlesWithMockStore(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")

	store := &mockStore{
		names: []string{"teststore"},
		creds: map[string]*secrets.StoreCredentials{
			"teststore": {Handle: "test-handle", AccessToken: "test-token"},
		},
	}

	secretsStoreFactory = func() (CredentialStore, error) {
		return store, nil
	}

	cmd := newTestCmdWithFlags()
	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

// articlesMockAPIClient is a mock implementation of api.APIClient for articles tests.
type articlesMockAPIClient struct {
	api.MockClient
	listArticlesResp  *api.ArticlesListResponse
	listArticlesErr   error
	getArticleResp    *api.Article
	getArticleErr     error
	createArticleResp *api.Article
	createArticleErr  error
	deleteArticleErr  error
}

func (m *articlesMockAPIClient) ListArticles(ctx context.Context, opts *api.ArticlesListOptions) (*api.ArticlesListResponse, error) {
	return m.listArticlesResp, m.listArticlesErr
}

func (m *articlesMockAPIClient) GetArticle(ctx context.Context, id string) (*api.Article, error) {
	return m.getArticleResp, m.getArticleErr
}

func (m *articlesMockAPIClient) CreateArticle(ctx context.Context, req *api.ArticleCreateRequest) (*api.Article, error) {
	return m.createArticleResp, m.createArticleErr
}

func (m *articlesMockAPIClient) DeleteArticle(ctx context.Context, id string) error {
	return m.deleteArticleErr
}

// setupArticlesMockFactories sets up mock factories for articles tests.
func setupArticlesMockFactories(mockClient *articlesMockAPIClient) (func(), *bytes.Buffer) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	buf := new(bytes.Buffer)
	formatterWriter = buf

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	cleanup := func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	return cleanup, buf
}

// newArticlesTestCmd creates a test command with common flags for articles tests.
func newArticlesTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

// TestArticlesListRunE tests the articles list command with mock API.
func TestArticlesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ArticlesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ArticlesListResponse{
				Items: []api.Article{
					{
						ID:        "art_123",
						BlogID:    "blog_456",
						Title:     "Test Article",
						Author:    "John Doe",
						Published: true,
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "art_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ArticlesListResponse{
				Items:      []api.Article{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple articles",
			mockResp: &api.ArticlesListResponse{
				Items: []api.Article{
					{
						ID:        "art_001",
						BlogID:    "blog_001",
						Title:     "First Article",
						Author:    "Author One",
						Published: true,
						CreatedAt: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
					},
					{
						ID:        "art_002",
						BlogID:    "blog_001",
						Title:     "Second Article",
						Author:    "Author Two",
						Published: false,
						CreatedAt: time.Date(2024, 1, 11, 11, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "art_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &articlesMockAPIClient{
				listArticlesResp: tt.mockResp,
				listArticlesErr:  tt.mockErr,
			}
			cleanup, buf := setupArticlesMockFactories(mockClient)
			defer cleanup()

			cmd := newArticlesTestCmd()
			cmd.Flags().String("blog-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := articlesListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			output := buf.String()
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestArticlesListJSONOutput tests the articles list command with JSON output.
func TestArticlesListJSONOutput(t *testing.T) {
	mockClient := &articlesMockAPIClient{
		listArticlesResp: &api.ArticlesListResponse{
			Items: []api.Article{
				{
					ID:        "art_json",
					BlogID:    "blog_json",
					Title:     "JSON Article",
					Author:    "JSON Author",
					Published: true,
					CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupArticlesMockFactories(mockClient)
	defer cleanup()

	cmd := newArticlesTestCmd()
	cmd.Flags().String("blog-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := articlesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "art_json") {
		t.Errorf("JSON output should contain article ID, got %q", output)
	}
}

// TestArticlesGetRunE tests the articles get command with mock API.
func TestArticlesGetRunE(t *testing.T) {
	tests := []struct {
		name      string
		articleID string
		mockResp  *api.Article
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			articleID: "art_123",
			mockResp: &api.Article{
				ID:          "art_123",
				BlogID:      "blog_456",
				Title:       "Test Article",
				Handle:      "test-article",
				Author:      "John Doe",
				BodyHTML:    "<p>Article content</p>",
				Tags:        "tech,news",
				Published:   true,
				PublishedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				CreatedAt:   time.Date(2024, 1, 14, 8, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:      "article not found",
			articleID: "art_999",
			mockErr:   errors.New("article not found"),
			wantErr:   true,
		},
		{
			name:      "get article with image",
			articleID: "art_img",
			mockResp: &api.Article{
				ID:        "art_img",
				BlogID:    "blog_456",
				Title:     "Article with Image",
				Handle:    "article-with-image",
				Author:    "Image Author",
				BodyHTML:  "<p>Content with image</p>",
				Published: true,
				Image: &api.Image{
					Src:    "https://example.com/image.jpg",
					Alt:    "Article image",
					Width:  800,
					Height: 600,
				},
				CreatedAt: time.Date(2024, 1, 14, 8, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:      "get unpublished article",
			articleID: "art_unpub",
			mockResp: &api.Article{
				ID:        "art_unpub",
				BlogID:    "blog_456",
				Title:     "Unpublished Article",
				Handle:    "unpublished-article",
				Author:    "Draft Author",
				BodyHTML:  "<p>Draft content</p>",
				Published: false,
				CreatedAt: time.Date(2024, 1, 14, 8, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &articlesMockAPIClient{
				getArticleResp: tt.mockResp,
				getArticleErr:  tt.mockErr,
			}
			cleanup, _ := setupArticlesMockFactories(mockClient)
			defer cleanup()

			cmd := newArticlesTestCmd()

			err := articlesGetCmd.RunE(cmd, []string{tt.articleID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestArticlesGetJSONOutput tests the articles get command with JSON output.
func TestArticlesGetJSONOutput(t *testing.T) {
	mockClient := &articlesMockAPIClient{
		getArticleResp: &api.Article{
			ID:        "art_json_get",
			BlogID:    "blog_456",
			Title:     "JSON Get Article",
			Handle:    "json-get-article",
			Author:    "JSON Author",
			BodyHTML:  "<p>JSON content</p>",
			Published: true,
			CreatedAt: time.Date(2024, 1, 14, 8, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupArticlesMockFactories(mockClient)
	defer cleanup()

	cmd := newArticlesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := articlesGetCmd.RunE(cmd, []string{"art_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "art_json_get") {
		t.Errorf("JSON output should contain article ID, got %q", output)
	}
}

// TestArticlesCreateRunE tests the articles create command with mock API.
func TestArticlesCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]string
		mockResp *api.Article
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			flags: map[string]string{
				"blog-id": "blog_123",
				"title":   "New Article",
				"body":    "<p>Article body</p>",
				"author":  "Test Author",
				"tags":    "tag1,tag2",
			},
			mockResp: &api.Article{
				ID:        "art_new",
				BlogID:    "blog_123",
				Title:     "New Article",
				Author:    "Test Author",
				BodyHTML:  "<p>Article body</p>",
				Tags:      "tag1,tag2",
				Published: false,
				CreatedAt: time.Now(),
			},
		},
		{
			name: "create with published flag",
			flags: map[string]string{
				"blog-id":   "blog_123",
				"title":     "Published Article",
				"body":      "<p>Published content</p>",
				"published": "true",
			},
			mockResp: &api.Article{
				ID:        "art_pub",
				BlogID:    "blog_123",
				Title:     "Published Article",
				BodyHTML:  "<p>Published content</p>",
				Published: true,
				CreatedAt: time.Now(),
			},
		},
		{
			name: "API error",
			flags: map[string]string{
				"blog-id": "blog_123",
				"title":   "Failed Article",
				"body":    "<p>Will fail</p>",
			},
			mockErr: errors.New("failed to create article"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &articlesMockAPIClient{
				createArticleResp: tt.mockResp,
				createArticleErr:  tt.mockErr,
			}
			cleanup, _ := setupArticlesMockFactories(mockClient)
			defer cleanup()

			cmd := newArticlesTestCmd()
			cmd.Flags().String("blog-id", "", "")
			cmd.Flags().String("title", "", "")
			cmd.Flags().String("body", "", "")
			cmd.Flags().String("author", "", "")
			cmd.Flags().String("tags", "", "")
			cmd.Flags().Bool("published", false, "")

			for key, val := range tt.flags {
				_ = cmd.Flags().Set(key, val)
			}

			err := articlesCreateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestArticlesCreateJSONOutput tests the articles create command with JSON output.
func TestArticlesCreateJSONOutput(t *testing.T) {
	mockClient := &articlesMockAPIClient{
		createArticleResp: &api.Article{
			ID:        "art_json_create",
			BlogID:    "blog_123",
			Title:     "JSON Created Article",
			Published: false,
			CreatedAt: time.Now(),
		},
	}
	cleanup, buf := setupArticlesMockFactories(mockClient)
	defer cleanup()

	cmd := newArticlesTestCmd()
	cmd.Flags().String("blog-id", "", "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("author", "", "")
	cmd.Flags().String("tags", "", "")
	cmd.Flags().Bool("published", false, "")
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("blog-id", "blog_123")
	_ = cmd.Flags().Set("title", "JSON Created Article")
	_ = cmd.Flags().Set("body", "<p>Content</p>")

	err := articlesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "art_json_create") {
		t.Errorf("JSON output should contain article ID, got %q", output)
	}
}

// TestArticlesCreateDryRun tests the articles create command with dry-run flag.
func TestArticlesCreateDryRun(t *testing.T) {
	// No mock client needed since dry-run doesn't call API
	mockClient := &articlesMockAPIClient{}
	cleanup, _ := setupArticlesMockFactories(mockClient)
	defer cleanup()

	cmd := newArticlesTestCmd()
	cmd.Flags().String("blog-id", "", "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("author", "", "")
	cmd.Flags().String("tags", "", "")
	cmd.Flags().Bool("published", false, "")
	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("blog-id", "blog_123")
	_ = cmd.Flags().Set("title", "Dry Run Article")
	_ = cmd.Flags().Set("body", "<p>Content</p>")

	err := articlesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// dry-run returns early without error and prints to stdout (not captured in buffer)
}

// TestArticlesDeleteRunE tests the articles delete command with mock API.
func TestArticlesDeleteRunE(t *testing.T) {
	tests := []struct {
		name      string
		articleID string
		mockErr   error
		useYes    bool
		wantErr   bool
	}{
		{
			name:      "successful delete",
			articleID: "art_delete",
			useYes:    true,
		},
		{
			name:      "API error",
			articleID: "art_fail",
			useYes:    true,
			mockErr:   errors.New("failed to delete article"),
			wantErr:   true,
		},
		{
			name:      "delete without confirmation",
			articleID: "art_noconfirm",
			useYes:    false,
			// This path prints a confirmation prompt but does not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &articlesMockAPIClient{
				deleteArticleErr: tt.mockErr,
			}
			cleanup, _ := setupArticlesMockFactories(mockClient)
			defer cleanup()

			cmd := newArticlesTestCmd()
			if tt.useYes {
				_ = cmd.Flags().Set("yes", "true")
			} else {
				_ = cmd.Flags().Set("yes", "false")
			}

			err := articlesDeleteCmd.RunE(cmd, []string{tt.articleID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestArticlesDeleteDryRun tests the articles delete command with dry-run flag.
func TestArticlesDeleteDryRun(t *testing.T) {
	mockClient := &articlesMockAPIClient{}
	cleanup, _ := setupArticlesMockFactories(mockClient)
	defer cleanup()

	cmd := newArticlesTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := articlesDeleteCmd.RunE(cmd, []string{"art_dryrun"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// dry-run returns early without error and prints to stdout (not captured in buffer)
}

// TestArticlesListWithBlogID tests the articles list command with blog-id filter.
func TestArticlesListWithBlogID(t *testing.T) {
	mockClient := &articlesMockAPIClient{
		listArticlesResp: &api.ArticlesListResponse{
			Items: []api.Article{
				{
					ID:        "art_filtered",
					BlogID:    "blog_specific",
					Title:     "Filtered Article",
					Author:    "Filter Author",
					Published: true,
					CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupArticlesMockFactories(mockClient)
	defer cleanup()

	cmd := newArticlesTestCmd()
	cmd.Flags().String("blog-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("blog-id", "blog_specific")

	err := articlesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "art_filtered") {
		t.Errorf("output should contain filtered article, got %q", output)
	}
}

// TestArticlesListPublishedStatus tests published status display in list.
func TestArticlesListPublishedStatus(t *testing.T) {
	tests := []struct {
		name       string
		published  bool
		wantOutput string
	}{
		{
			name:       "published article shows Yes",
			published:  true,
			wantOutput: "Yes",
		},
		{
			name:       "unpublished article shows No",
			published:  false,
			wantOutput: "No",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &articlesMockAPIClient{
				listArticlesResp: &api.ArticlesListResponse{
					Items: []api.Article{
						{
							ID:        "art_status",
							BlogID:    "blog_123",
							Title:     "Status Test",
							Author:    "Author",
							Published: tt.published,
							CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						},
					},
					TotalCount: 1,
				},
			}
			cleanup, buf := setupArticlesMockFactories(mockClient)
			defer cleanup()

			cmd := newArticlesTestCmd()
			cmd.Flags().String("blog-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := articlesListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}
