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

// TestCategoriesCommandSetup verifies categories command initialization
func TestCategoriesCommandSetup(t *testing.T) {
	if categoriesCmd.Use != "categories" {
		t.Errorf("expected Use 'categories', got %q", categoriesCmd.Use)
	}
	if categoriesCmd.Short != "Manage product categories" {
		t.Errorf("expected Short 'Manage product categories', got %q", categoriesCmd.Short)
	}
}

// TestCategoriesSubcommands verifies all subcommands are registered
func TestCategoriesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":             "List categories",
		"get":              "Get category details",
		"create":           "Create a category",
		"delete":           "Delete a category",
		"products-sorting": "Manage category product sorting",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range categoriesCmd.Commands() {
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

// TestCategoriesListFlags verifies list command flags exist with correct defaults
func TestCategoriesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"parent-id", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := categoriesListCmd.Flags().Lookup(f.name)
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

// TestCategoriesCreateFlags verifies create command flags exist with correct defaults
func TestCategoriesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"title", ""},
		{"handle", ""},
		{"description", ""},
		{"parent-id", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := categoriesCreateCmd.Flags().Lookup(f.name)
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

// TestCategoriesCreateRequiredFlags verifies title is required
func TestCategoriesCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"title"}

	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := categoriesCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
				return
			}
			// Check if the flag has required annotation
			annotations := flag.Annotations
			if annotations == nil {
				t.Errorf("flag %q has no annotations, expected required", name)
				return
			}
			if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
				t.Errorf("flag %q is not marked as required", name)
			}
		})
	}
}

// TestCategoriesGetArgs verifies get command requires exactly 1 argument
func TestCategoriesGetArgs(t *testing.T) {
	err := categoriesGetCmd.Args(categoriesGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = categoriesGetCmd.Args(categoriesGetCmd, []string{"cat-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCategoriesDeleteArgs verifies delete command requires exactly 1 argument
func TestCategoriesDeleteArgs(t *testing.T) {
	err := categoriesDeleteCmd.Args(categoriesDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = categoriesDeleteCmd.Args(categoriesDeleteCmd, []string{"cat-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCategoriesGetClientError verifies error handling when getClient fails
func TestCategoriesGetClientError(t *testing.T) {
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

// TestCategoriesListGetClientError verifies list command error handling when getClient fails
func TestCategoriesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(categoriesListCmd)

	err := categoriesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCategoriesGetGetClientError verifies get command error handling when getClient fails
func TestCategoriesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(categoriesGetCmd)

	err := categoriesGetCmd.RunE(cmd, []string{"cat-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCategoriesCreateGetClientError verifies create command error handling when getClient fails
func TestCategoriesCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(categoriesCreateCmd)

	err := categoriesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCategoriesDeleteGetClientError verifies delete command error handling when getClient fails
func TestCategoriesDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(categoriesDeleteCmd)

	err := categoriesDeleteCmd.RunE(cmd, []string{"cat-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCategoriesWithMockStore tests categories commands with a mock credential store
func TestCategoriesWithMockStore(t *testing.T) {
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

// categoriesMockAPIClient is a mock implementation of api.APIClient for categories tests.
type categoriesMockAPIClient struct {
	api.MockClient
	listCategoriesResp *api.CategoriesListResponse
	listCategoriesErr  error
	getCategoryResp    *api.Category
	getCategoryErr     error
	createCategoryResp *api.Category
	createCategoryErr  error
	deleteCategoryErr  error
}

func (m *categoriesMockAPIClient) ListCategories(ctx context.Context, opts *api.CategoriesListOptions) (*api.CategoriesListResponse, error) {
	return m.listCategoriesResp, m.listCategoriesErr
}

func (m *categoriesMockAPIClient) GetCategory(ctx context.Context, id string) (*api.Category, error) {
	return m.getCategoryResp, m.getCategoryErr
}

func (m *categoriesMockAPIClient) CreateCategory(ctx context.Context, req *api.CategoryCreateRequest) (*api.Category, error) {
	return m.createCategoryResp, m.createCategoryErr
}

func (m *categoriesMockAPIClient) DeleteCategory(ctx context.Context, id string) error {
	return m.deleteCategoryErr
}

// setupCategoriesMockFactories sets up mock factories for categories tests.
func setupCategoriesMockFactories(mockClient *categoriesMockAPIClient) (func(), *bytes.Buffer) {
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

// newCategoriesTestCmd creates a test command with common flags for categories tests.
func newCategoriesTestCmd() *cobra.Command {
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

// TestCategoriesListRunE tests the categories list command with mock API.
func TestCategoriesListRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.CategoriesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CategoriesListResponse{
				Items: []api.Category{
					{
						ID:        "cat_123",
						Title:     "Electronics",
						Handle:    "electronics",
						ParentID:  "",
						Position:  1,
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "cat_123",
		},
		{
			name: "list with multiple categories",
			mockResp: &api.CategoriesListResponse{
				Items: []api.Category{
					{
						ID:        "cat_123",
						Title:     "Electronics",
						Handle:    "electronics",
						ParentID:  "",
						Position:  1,
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
					{
						ID:        "cat_456",
						Title:     "Clothing",
						Handle:    "clothing",
						ParentID:  "",
						Position:  2,
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "cat_456",
		},
		{
			name: "list with parent category",
			mockResp: &api.CategoriesListResponse{
				Items: []api.Category{
					{
						ID:        "cat_789",
						Title:     "Smartphones",
						Handle:    "smartphones",
						ParentID:  "cat_123",
						Position:  1,
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "cat_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CategoriesListResponse{
				Items:      []api.Category{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &categoriesMockAPIClient{
				listCategoriesResp: tt.mockResp,
				listCategoriesErr:  tt.mockErr,
			}
			cleanup, buf := setupCategoriesMockFactories(mockClient)
			defer cleanup()

			cmd := newCategoriesTestCmd()
			cmd.Flags().String("parent-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := categoriesListCmd.RunE(cmd, []string{})

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

// TestCategoriesListRunEJSON tests the categories list command with JSON output.
func TestCategoriesListRunEJSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &categoriesMockAPIClient{
		listCategoriesResp: &api.CategoriesListResponse{
			Items: []api.Category{
				{
					ID:        "cat_123",
					Title:     "Electronics",
					Handle:    "electronics",
					ParentID:  "",
					Position:  1,
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCategoriesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := categoriesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cat_123") {
		t.Errorf("JSON output should contain category ID, got: %s", output)
	}
}

// TestCategoriesGetRunE tests the categories get command with mock API.
func TestCategoriesGetRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		categoryID string
		mockResp   *api.Category
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "successful get",
			categoryID: "cat_123",
			mockResp: &api.Category{
				ID:          "cat_123",
				Title:       "Electronics",
				Handle:      "electronics",
				Description: "All electronic products",
				ParentID:    "",
				Position:    1,
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
			wantOutput: "cat_123",
		},
		{
			name:       "get with parent",
			categoryID: "cat_789",
			mockResp: &api.Category{
				ID:          "cat_789",
				Title:       "Smartphones",
				Handle:      "smartphones",
				Description: "Mobile phones and accessories",
				ParentID:    "cat_123",
				Position:    1,
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
			wantOutput: "Smartphones",
		},
		{
			name:       "category not found",
			categoryID: "cat_999",
			mockErr:    errors.New("category not found"),
			wantErr:    true,
		},
		{
			name:       "API error",
			categoryID: "cat_123",
			mockErr:    errors.New("API unavailable"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &categoriesMockAPIClient{
				getCategoryResp: tt.mockResp,
				getCategoryErr:  tt.mockErr,
			}
			cleanup, _ := setupCategoriesMockFactories(mockClient)
			defer cleanup()

			cmd := newCategoriesTestCmd()

			err := categoriesGetCmd.RunE(cmd, []string{tt.categoryID})

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
			// Note: Text output goes to stdout via fmt.Printf, not the buffer.
			// This test verifies the command runs without error for text output.
		})
	}
}

// TestCategoriesGetRunEJSON tests the categories get command with JSON output.
func TestCategoriesGetRunEJSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &categoriesMockAPIClient{
		getCategoryResp: &api.Category{
			ID:          "cat_123",
			Title:       "Electronics",
			Handle:      "electronics",
			Description: "All electronic products",
			ParentID:    "",
			Position:    1,
			CreatedAt:   testTime,
			UpdatedAt:   testTime,
		},
	}
	cleanup, buf := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCategoriesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := categoriesGetCmd.RunE(cmd, []string{"cat_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cat_123") {
		t.Errorf("JSON output should contain category ID, got: %s", output)
	}
}

// TestCategoriesCreateRunE tests the categories create command with mock API.
func TestCategoriesCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		title      string
		handle     string
		desc       string
		parentID   string
		mockResp   *api.Category
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful create",
			title:  "Electronics",
			handle: "electronics",
			desc:   "All electronic products",
			mockResp: &api.Category{
				ID:          "cat_new",
				Title:       "Electronics",
				Handle:      "electronics",
				Description: "All electronic products",
				ParentID:    "",
				Position:    1,
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
			wantOutput: "cat_new",
		},
		{
			name:     "create with parent",
			title:    "Smartphones",
			handle:   "smartphones",
			desc:     "Mobile phones",
			parentID: "cat_123",
			mockResp: &api.Category{
				ID:          "cat_child",
				Title:       "Smartphones",
				Handle:      "smartphones",
				Description: "Mobile phones",
				ParentID:    "cat_123",
				Position:    1,
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
			wantOutput: "cat_child",
		},
		{
			name:  "create with minimal fields",
			title: "Books",
			mockResp: &api.Category{
				ID:        "cat_minimal",
				Title:     "Books",
				Handle:    "books",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
			wantOutput: "Books",
		},
		{
			name:    "API error",
			title:   "Invalid",
			mockErr: errors.New("validation error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &categoriesMockAPIClient{
				createCategoryResp: tt.mockResp,
				createCategoryErr:  tt.mockErr,
			}
			cleanup, _ := setupCategoriesMockFactories(mockClient)
			defer cleanup()

			cmd := newCategoriesTestCmd()
			cmd.Flags().String("title", "", "")
			cmd.Flags().String("handle", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("parent-id", "", "")

			if tt.title != "" {
				_ = cmd.Flags().Set("title", tt.title)
			}
			if tt.handle != "" {
				_ = cmd.Flags().Set("handle", tt.handle)
			}
			if tt.desc != "" {
				_ = cmd.Flags().Set("description", tt.desc)
			}
			if tt.parentID != "" {
				_ = cmd.Flags().Set("parent-id", tt.parentID)
			}

			err := categoriesCreateCmd.RunE(cmd, []string{})

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
			// Note: Text output goes to stdout via fmt.Printf, not the buffer.
			// This test verifies the command runs without error for text output.
		})
	}
}

// TestCategoriesCreateRunEJSON tests the categories create command with JSON output.
func TestCategoriesCreateRunEJSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &categoriesMockAPIClient{
		createCategoryResp: &api.Category{
			ID:          "cat_new",
			Title:       "Electronics",
			Handle:      "electronics",
			Description: "All electronic products",
			ParentID:    "",
			Position:    1,
			CreatedAt:   testTime,
			UpdatedAt:   testTime,
		},
	}
	cleanup, buf := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCategoriesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("parent-id", "", "")
	_ = cmd.Flags().Set("title", "Electronics")

	err := categoriesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cat_new") {
		t.Errorf("JSON output should contain category ID, got: %s", output)
	}
}

// TestCategoriesDeleteRunE tests the categories delete command with mock API.
func TestCategoriesDeleteRunE(t *testing.T) {
	tests := []struct {
		name       string
		categoryID string
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "successful delete",
			categoryID: "cat_123",
			wantOutput: "Deleted",
		},
		{
			name:       "delete non-existent category",
			categoryID: "cat_999",
			mockErr:    errors.New("category not found"),
			wantErr:    true,
		},
		{
			name:       "API error",
			categoryID: "cat_123",
			mockErr:    errors.New("API unavailable"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &categoriesMockAPIClient{
				deleteCategoryErr: tt.mockErr,
			}
			cleanup, _ := setupCategoriesMockFactories(mockClient)
			defer cleanup()

			cmd := newCategoriesTestCmd()
			// Set yes flag to skip confirmation
			_ = cmd.Flags().Set("yes", "true")

			err := categoriesDeleteCmd.RunE(cmd, []string{tt.categoryID})

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

// TestCategoriesDeleteConfirmationDenied tests the delete command when user denies confirmation.
func TestCategoriesDeleteConfirmationDenied(t *testing.T) {
	mockClient := &categoriesMockAPIClient{}
	cleanup, _ := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	// Create a pipe to simulate stdin with "n" input
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write "n" to stdin (user declines)
	go func() {
		_, _ = w.WriteString("n\n")
		_ = w.Close()
	}()

	cmd := newCategoriesTestCmd()
	_ = cmd.Flags().Set("yes", "false")

	// The command should return nil (cancelled, not an error)
	err := categoriesDeleteCmd.RunE(cmd, []string{"cat_123"})
	if err != nil {
		t.Errorf("expected nil error for cancelled delete, got: %v", err)
	}
}

// TestCategoriesDeleteConfirmationAccepted tests the delete command when user confirms.
func TestCategoriesDeleteConfirmationAccepted(t *testing.T) {
	mockClient := &categoriesMockAPIClient{}
	cleanup, _ := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	// Create a pipe to simulate stdin with "y" input
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write "y" to stdin (user confirms)
	go func() {
		_, _ = w.WriteString("y\n")
		_ = w.Close()
	}()

	cmd := newCategoriesTestCmd()
	_ = cmd.Flags().Set("yes", "false")

	// The command should succeed
	err := categoriesDeleteCmd.RunE(cmd, []string{"cat_123"})
	if err != nil {
		t.Errorf("expected nil error for confirmed delete, got: %v", err)
	}
}

// TestCategoriesListWithParentIDFilter tests list command with parent-id filter.
func TestCategoriesListWithParentIDFilter(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &categoriesMockAPIClient{
		listCategoriesResp: &api.CategoriesListResponse{
			Items: []api.Category{
				{
					ID:        "cat_child",
					Title:     "Smartphones",
					Handle:    "smartphones",
					ParentID:  "cat_electronics",
					Position:  1,
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCategoriesTestCmd()
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("parent-id", "cat_electronics")

	err := categoriesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cat_child") {
		t.Errorf("output should contain filtered category, got: %s", output)
	}
}

// TestCategoriesListPagination tests list command with pagination options.
func TestCategoriesListPagination(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &categoriesMockAPIClient{
		listCategoriesResp: &api.CategoriesListResponse{
			Items: []api.Category{
				{
					ID:        "cat_page2",
					Title:     "Books",
					Handle:    "books",
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
			},
			TotalCount: 50,
		},
	}
	cleanup, buf := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCategoriesTestCmd()
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := categoriesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cat_page2") {
		t.Errorf("output should contain paginated category, got: %s", output)
	}
}

// TestCategoriesGetTextOutput tests the get command with text output format.
func TestCategoriesGetTextOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &categoriesMockAPIClient{
		getCategoryResp: &api.Category{
			ID:          "cat_123",
			Title:       "Electronics",
			Handle:      "electronics",
			Description: "Electronic devices and accessories",
			ParentID:    "cat_parent",
			Position:    5,
			CreatedAt:   testTime,
			UpdatedAt:   testTime,
		},
	}
	cleanup, _ := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCategoriesTestCmd()
	// Text output is default (empty or "text")

	err := categoriesGetCmd.RunE(cmd, []string{"cat_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	// Note: Text output goes to stdout via fmt.Printf, not the buffer.
	// This test verifies the command runs without error for text output.
}

// TestCategoriesCreateAllFields tests create command with all optional fields.
func TestCategoriesCreateAllFields(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &categoriesMockAPIClient{
		createCategoryResp: &api.Category{
			ID:          "cat_full",
			Title:       "Home & Garden",
			Handle:      "home-garden",
			Description: "Home improvement and garden supplies",
			ParentID:    "cat_parent",
			Position:    3,
			CreatedAt:   testTime,
			UpdatedAt:   testTime,
		},
	}
	cleanup, _ := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCategoriesTestCmd()
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("parent-id", "", "")

	_ = cmd.Flags().Set("title", "Home & Garden")
	_ = cmd.Flags().Set("handle", "home-garden")
	_ = cmd.Flags().Set("description", "Home improvement and garden supplies")
	_ = cmd.Flags().Set("parent-id", "cat_parent")

	err := categoriesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	// Note: Text output goes to stdout via fmt.Printf, not the buffer.
	// This test verifies the command runs without error for text output.
}

// TestCategoriesListEmptyResponse tests list command handles empty response correctly.
func TestCategoriesListEmptyResponse(t *testing.T) {
	mockClient := &categoriesMockAPIClient{
		listCategoriesResp: &api.CategoriesListResponse{
			Items:      []api.Category{},
			TotalCount: 0,
		},
	}
	cleanup, _ := setupCategoriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCategoriesTestCmd()
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := categoriesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error for empty list: %v", err)
	}
}

// TestCategoriesListWithValidStore tests list command execution with valid store
func TestCategoriesListWithValidStore(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(categoriesListCmd)

	// This will fail at the API call level, but validates the client setup works
	err := categoriesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("categoriesListCmd succeeded (might be due to mock setup)")
	}
}

// TestCategoriesErrorMessages tests that error messages are properly wrapped.
func TestCategoriesErrorMessages(t *testing.T) {
	baseErr := errors.New("connection refused")

	tests := []struct {
		name    string
		setup   func(*categoriesMockAPIClient)
		run     func(*cobra.Command) error
		wantMsg string
	}{
		{
			name: "list error wrapping",
			setup: func(m *categoriesMockAPIClient) {
				m.listCategoriesErr = baseErr
			},
			run: func(cmd *cobra.Command) error {
				cmd.Flags().String("parent-id", "", "")
				cmd.Flags().Int("page", 1, "")
				cmd.Flags().Int("page-size", 20, "")
				return categoriesListCmd.RunE(cmd, []string{})
			},
			wantMsg: "failed to list categories",
		},
		{
			name: "get error wrapping",
			setup: func(m *categoriesMockAPIClient) {
				m.getCategoryErr = baseErr
			},
			run: func(cmd *cobra.Command) error {
				return categoriesGetCmd.RunE(cmd, []string{"cat_123"})
			},
			wantMsg: "failed to get category",
		},
		{
			name: "create error wrapping",
			setup: func(m *categoriesMockAPIClient) {
				m.createCategoryErr = baseErr
			},
			run: func(cmd *cobra.Command) error {
				cmd.Flags().String("title", "", "")
				cmd.Flags().String("handle", "", "")
				cmd.Flags().String("description", "", "")
				cmd.Flags().String("parent-id", "", "")
				return categoriesCreateCmd.RunE(cmd, []string{})
			},
			wantMsg: "failed to create category",
		},
		{
			name: "delete error wrapping",
			setup: func(m *categoriesMockAPIClient) {
				m.deleteCategoryErr = baseErr
			},
			run: func(cmd *cobra.Command) error {
				_ = cmd.Flags().Set("yes", "true")
				return categoriesDeleteCmd.RunE(cmd, []string{"cat_123"})
			},
			wantMsg: "failed to delete category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &categoriesMockAPIClient{}
			tt.setup(mockClient)

			cleanup, _ := setupCategoriesMockFactories(mockClient)
			defer cleanup()

			cmd := newCategoriesTestCmd()
			err := tt.run(cmd)

			if err == nil {
				t.Error("expected error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error %q should contain %q", err.Error(), tt.wantMsg)
			}
		})
	}
}
