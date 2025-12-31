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

func TestPagesCommandSetup(t *testing.T) {
	if pagesCmd.Use != "pages" {
		t.Errorf("expected Use 'pages', got %q", pagesCmd.Use)
	}
	if pagesCmd.Short != "Manage pages" {
		t.Errorf("expected Short 'Manage pages', got %q", pagesCmd.Short)
	}
}

func TestPagesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List pages",
		"get":    "Get page details",
		"create": "Create a page",
		"delete": "Delete a page",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range pagesCmd.Commands() {
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

func TestPagesListGetClientError(t *testing.T) {
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

func TestPagesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"title", ""},
		{"published", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := pagesListCmd.Flags().Lookup(f.name)
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

func TestPagesGetFlags(t *testing.T) {
	flag := pagesGetCmd.Flags().Lookup("body")
	if flag == nil {
		t.Error("flag 'body' not found")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}

func TestPagesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"title", ""},
		{"body", ""},
		{"handle", ""},
		{"author", ""},
		{"published", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := pagesCreateCmd.Flags().Lookup(f.name)
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

func TestPagesDeleteFlags(t *testing.T) {
	flag := pagesDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("flag 'yes' not found")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}

func TestPagesWithMockStore(t *testing.T) {
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

// pagesMockAPIClient is a mock implementation of api.APIClient for pages tests.
type pagesMockAPIClient struct {
	api.MockClient
	listPagesResp   *api.PagesListResponse
	listPagesErr    error
	getPageResp     *api.Page
	getPageErr      error
	createPageResp  *api.Page
	createPageErr   error
	deletePageErr   error
	deletedPageID   string
	createPageInput *api.PageCreateRequest
}

func (m *pagesMockAPIClient) ListPages(ctx context.Context, opts *api.PagesListOptions) (*api.PagesListResponse, error) {
	return m.listPagesResp, m.listPagesErr
}

func (m *pagesMockAPIClient) GetPage(ctx context.Context, id string) (*api.Page, error) {
	return m.getPageResp, m.getPageErr
}

func (m *pagesMockAPIClient) CreatePage(ctx context.Context, req *api.PageCreateRequest) (*api.Page, error) {
	m.createPageInput = req
	return m.createPageResp, m.createPageErr
}

func (m *pagesMockAPIClient) DeletePage(ctx context.Context, id string) error {
	m.deletedPageID = id
	return m.deletePageErr
}

// setupPagesMockFactories sets up mock factories for pages tests.
func setupPagesMockFactories(mockClient *pagesMockAPIClient) (func(), *bytes.Buffer) {
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

// newPagesTestCmd creates a test command with common flags for pages tests.
func newPagesTestCmd() *cobra.Command {
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

// TestPagesListRunE tests the pages list command with mock API.
func TestPagesListRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.PagesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.PagesListResponse{
				Items: []api.Page{
					{
						ID:        "page_123",
						Title:     "About Us",
						Handle:    "about-us",
						Published: true,
						Author:    "John Doe",
						CreatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "page_123",
		},
		{
			name: "multiple pages",
			mockResp: &api.PagesListResponse{
				Items: []api.Page{
					{
						ID:        "page_1",
						Title:     "Page 1",
						Handle:    "page-1",
						Published: true,
						Author:    "Author 1",
						CreatedAt: testTime,
					},
					{
						ID:        "page_2",
						Title:     "Page 2",
						Handle:    "page-2",
						Published: false,
						Author:    "Author 2",
						CreatedAt: testTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "page_1",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.PagesListResponse{
				Items:      []api.Page{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pagesMockAPIClient{
				listPagesResp: tt.mockResp,
				listPagesErr:  tt.mockErr,
			}
			cleanup, buf := setupPagesMockFactories(mockClient)
			defer cleanup()

			cmd := newPagesTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("title", "", "")
			cmd.Flags().Bool("published", false, "")

			err := pagesListCmd.RunE(cmd, []string{})

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

// TestPagesListRunE_JSON tests the pages list command with JSON output.
func TestPagesListRunE_JSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &pagesMockAPIClient{
		listPagesResp: &api.PagesListResponse{
			Items: []api.Page{
				{
					ID:        "page_123",
					Title:     "About Us",
					Handle:    "about-us",
					Published: true,
					Author:    "John Doe",
					CreatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupPagesMockFactories(mockClient)
	defer cleanup()

	cmd := newPagesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().Bool("published", false, "")

	err := pagesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "page_123") {
		t.Errorf("JSON output should contain page ID, got: %s", output)
	}
}

// TestPagesGetRunE tests the pages get command with mock API.
func TestPagesGetRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	publishedAt := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		pageID   string
		mockResp *api.Page
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "successful get",
			pageID: "page_123",
			mockResp: &api.Page{
				ID:             "page_123",
				Title:          "About Us",
				Handle:         "about-us",
				Author:         "John Doe",
				Published:      true,
				PublishedAt:    publishedAt,
				TemplateSuffix: "custom",
				BodyHTML:       "<p>About us content</p>",
				CreatedAt:      testTime,
				UpdatedAt:      testTime,
			},
		},
		{
			name:   "page without published at",
			pageID: "page_456",
			mockResp: &api.Page{
				ID:        "page_456",
				Title:     "Draft Page",
				Handle:    "draft-page",
				Author:    "Jane Doe",
				Published: false,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		{
			name:    "page not found",
			pageID:  "page_999",
			mockErr: errors.New("page not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pagesMockAPIClient{
				getPageResp: tt.mockResp,
				getPageErr:  tt.mockErr,
			}
			cleanup, _ := setupPagesMockFactories(mockClient)
			defer cleanup()

			cmd := newPagesTestCmd()
			cmd.Flags().Bool("body", false, "")

			err := pagesGetCmd.RunE(cmd, []string{tt.pageID})

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

// TestPagesGetRunE_JSON tests the pages get command with JSON output.
func TestPagesGetRunE_JSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &pagesMockAPIClient{
		getPageResp: &api.Page{
			ID:        "page_123",
			Title:     "About Us",
			Handle:    "about-us",
			Author:    "John Doe",
			Published: true,
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, buf := setupPagesMockFactories(mockClient)
	defer cleanup()

	cmd := newPagesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Bool("body", false, "")

	err := pagesGetCmd.RunE(cmd, []string{"page_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "page_123") {
		t.Errorf("JSON output should contain page ID, got: %s", output)
	}
}

// TestPagesGetRunE_WithBody tests the pages get command with --body flag.
func TestPagesGetRunE_WithBody(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &pagesMockAPIClient{
		getPageResp: &api.Page{
			ID:        "page_123",
			Title:     "About Us",
			Handle:    "about-us",
			Author:    "John Doe",
			Published: true,
			BodyHTML:  "<p>This is the body content</p>",
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, _ := setupPagesMockFactories(mockClient)
	defer cleanup()

	cmd := newPagesTestCmd()
	cmd.Flags().Bool("body", false, "")
	_ = cmd.Flags().Set("body", "true")

	err := pagesGetCmd.RunE(cmd, []string{"page_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPagesCreateRunE tests the pages create command with mock API.
func TestPagesCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		title     string
		body      string
		handle    string
		author    string
		published bool
		mockResp  *api.Page
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful create",
			title:     "New Page",
			body:      "<p>Content</p>",
			handle:    "new-page",
			author:    "Author",
			published: true,
			mockResp: &api.Page{
				ID:        "page_new",
				Title:     "New Page",
				Handle:    "new-page",
				Author:    "Author",
				Published: true,
				BodyHTML:  "<p>Content</p>",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		{
			name:      "create with minimal fields",
			title:     "Minimal Page",
			body:      "<p>Minimal content</p>",
			published: false,
			mockResp: &api.Page{
				ID:        "page_minimal",
				Title:     "Minimal Page",
				Handle:    "minimal-page",
				Published: false,
				BodyHTML:  "<p>Minimal content</p>",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		{
			name:    "API error",
			title:   "Error Page",
			body:    "<p>Error content</p>",
			mockErr: errors.New("failed to create page"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pagesMockAPIClient{
				createPageResp: tt.mockResp,
				createPageErr:  tt.mockErr,
			}
			cleanup, _ := setupPagesMockFactories(mockClient)
			defer cleanup()

			cmd := newPagesTestCmd()
			cmd.Flags().String("title", "", "")
			cmd.Flags().String("body", "", "")
			cmd.Flags().String("handle", "", "")
			cmd.Flags().String("author", "", "")
			cmd.Flags().Bool("published", false, "")

			_ = cmd.Flags().Set("title", tt.title)
			_ = cmd.Flags().Set("body", tt.body)
			if tt.handle != "" {
				_ = cmd.Flags().Set("handle", tt.handle)
			}
			if tt.author != "" {
				_ = cmd.Flags().Set("author", tt.author)
			}
			if tt.published {
				_ = cmd.Flags().Set("published", "true")
			}

			err := pagesCreateCmd.RunE(cmd, []string{})

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

// TestPagesCreateRunE_DryRun tests the pages create command with --dry-run flag.
func TestPagesCreateRunE_DryRun(t *testing.T) {
	mockClient := &pagesMockAPIClient{}
	cleanup, _ := setupPagesMockFactories(mockClient)
	defer cleanup()

	cmd := newPagesTestCmd()
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("author", "", "")
	cmd.Flags().Bool("published", false, "")

	_ = cmd.Flags().Set("title", "Test Page")
	_ = cmd.Flags().Set("body", "<p>Test content</p>")
	_ = cmd.Flags().Set("dry-run", "true")

	err := pagesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// In dry-run mode, CreatePage should not be called
	if mockClient.createPageInput != nil {
		t.Error("CreatePage should not be called in dry-run mode")
	}
}

// TestPagesCreateRunE_JSON tests the pages create command with JSON output.
func TestPagesCreateRunE_JSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &pagesMockAPIClient{
		createPageResp: &api.Page{
			ID:        "page_json",
			Title:     "JSON Page",
			Handle:    "json-page",
			Published: true,
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, buf := setupPagesMockFactories(mockClient)
	defer cleanup()

	cmd := newPagesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("author", "", "")
	cmd.Flags().Bool("published", false, "")

	_ = cmd.Flags().Set("title", "JSON Page")
	_ = cmd.Flags().Set("body", "<p>Content</p>")

	err := pagesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "page_json") {
		t.Errorf("JSON output should contain page ID, got: %s", output)
	}
}

// TestPagesDeleteRunE tests the pages delete command with mock API.
func TestPagesDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		pageID  string
		yes     bool
		dryRun  bool
		mockErr error
		wantErr bool
	}{
		{
			name:   "successful delete with confirmation",
			pageID: "page_123",
			yes:    true,
		},
		{
			name:   "delete without confirmation",
			pageID: "page_456",
			yes:    false,
			// Should return early without error (prompts user)
		},
		{
			name:   "dry run delete",
			pageID: "page_789",
			yes:    true,
			dryRun: true,
		},
		{
			name:    "API error",
			pageID:  "page_error",
			yes:     true,
			mockErr: errors.New("failed to delete page"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pagesMockAPIClient{
				deletePageErr: tt.mockErr,
			}
			cleanup, _ := setupPagesMockFactories(mockClient)
			defer cleanup()

			cmd := newPagesTestCmd()
			if tt.yes {
				_ = cmd.Flags().Set("yes", "true")
			} else {
				_ = cmd.Flags().Set("yes", "false")
			}
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := pagesDeleteCmd.RunE(cmd, []string{tt.pageID})

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

			// Verify delete was called only when not dry-run and confirmed
			if tt.yes && !tt.dryRun && mockClient.deletedPageID != tt.pageID && tt.mockErr == nil {
				t.Errorf("expected DeletePage to be called with %q, got %q", tt.pageID, mockClient.deletedPageID)
			}

			// Verify delete was NOT called in dry-run mode
			if tt.dryRun && mockClient.deletedPageID != "" {
				t.Error("DeletePage should not be called in dry-run mode")
			}
		})
	}
}

// TestPagesListRunE_NoProfiles tests error when no store profiles exist.
func TestPagesListRunE_NoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newPagesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().Bool("published", false, "")

	err := pagesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestPagesGetRunE_NoProfiles tests error when no store profiles exist.
func TestPagesGetRunE_NoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newPagesTestCmd()
	cmd.Flags().Bool("body", false, "")

	err := pagesGetCmd.RunE(cmd, []string{"page_123"})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestPagesCreateRunE_NoProfiles tests error when no store profiles exist.
func TestPagesCreateRunE_NoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newPagesTestCmd()
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("author", "", "")
	cmd.Flags().Bool("published", false, "")

	_ = cmd.Flags().Set("title", "Test Page")
	_ = cmd.Flags().Set("body", "<p>Test</p>")

	err := pagesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestPagesDeleteRunE_NoProfiles tests error when no store profiles exist.
func TestPagesDeleteRunE_NoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newPagesTestCmd()

	err := pagesDeleteCmd.RunE(cmd, []string{"page_123"})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestPagesListRunE_MultipleProfiles tests error when multiple profiles exist without selection.
func TestPagesListRunE_MultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newPagesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().Bool("published", false, "")

	err := pagesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestPagesListRunE_PublishedFilter tests the published filter handling.
func TestPagesListRunE_PublishedFilter(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &pagesMockAPIClient{
		listPagesResp: &api.PagesListResponse{
			Items: []api.Page{
				{
					ID:        "page_published",
					Title:     "Published Page",
					Handle:    "published-page",
					Published: true,
					Author:    "Author",
					CreatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupPagesMockFactories(mockClient)
	defer cleanup()

	cmd := newPagesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().Bool("published", false, "")
	// Mark the published flag as changed to trigger the tri-state behavior
	_ = cmd.Flags().Set("published", "true")

	err := pagesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "page_published") {
		t.Errorf("output should contain page_published, got: %s", output)
	}
}

// TestPagesCreateRunE_VerifyInput tests that create passes correct input to API.
func TestPagesCreateRunE_VerifyInput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &pagesMockAPIClient{
		createPageResp: &api.Page{
			ID:        "page_new",
			Title:     "Test Title",
			Handle:    "test-handle",
			Author:    "Test Author",
			Published: true,
			BodyHTML:  "<p>Test Body</p>",
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, _ := setupPagesMockFactories(mockClient)
	defer cleanup()

	cmd := newPagesTestCmd()
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("author", "", "")
	cmd.Flags().Bool("published", false, "")

	_ = cmd.Flags().Set("title", "Test Title")
	_ = cmd.Flags().Set("body", "<p>Test Body</p>")
	_ = cmd.Flags().Set("handle", "test-handle")
	_ = cmd.Flags().Set("author", "Test Author")
	_ = cmd.Flags().Set("published", "true")

	err := pagesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Verify the input passed to CreatePage
	if mockClient.createPageInput == nil {
		t.Fatal("CreatePage was not called")
	}
	if mockClient.createPageInput.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", mockClient.createPageInput.Title)
	}
	if mockClient.createPageInput.BodyHTML != "<p>Test Body</p>" {
		t.Errorf("expected body '<p>Test Body</p>', got %q", mockClient.createPageInput.BodyHTML)
	}
	if mockClient.createPageInput.Handle != "test-handle" {
		t.Errorf("expected handle 'test-handle', got %q", mockClient.createPageInput.Handle)
	}
	if mockClient.createPageInput.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got %q", mockClient.createPageInput.Author)
	}
	if !mockClient.createPageInput.Published {
		t.Error("expected published to be true")
	}
}
