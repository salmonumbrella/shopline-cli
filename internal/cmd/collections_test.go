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

// TestCollectionsCommandSetup verifies collections command initialization
func TestCollectionsCommandSetup(t *testing.T) {
	if collectionsCmd.Use != "collections" {
		t.Errorf("expected Use 'collections', got %q", collectionsCmd.Use)
	}
	if collectionsCmd.Short != "Manage product collections" {
		t.Errorf("expected Short 'Manage product collections', got %q", collectionsCmd.Short)
	}
}

// TestCollectionsSubcommands verifies all subcommands are registered
func TestCollectionsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List collections",
		"get":    "Get collection details",
		"create": "Create a collection",
		"delete": "Delete a collection",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range collectionsCmd.Commands() {
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

// TestCollectionsListFlags verifies list command flags exist with correct defaults
func TestCollectionsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"title", ""},
		{"handle", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := collectionsListCmd.Flags().Lookup(f.name)
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

// TestCollectionsCreateFlags verifies create command flags exist with correct defaults
func TestCollectionsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"title", ""},
		{"handle", ""},
		{"description", ""},
		{"sort-order", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := collectionsCreateCmd.Flags().Lookup(f.name)
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

// TestCollectionsCreateRequiredFlags verifies title is required
func TestCollectionsCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"title"}

	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := collectionsCreateCmd.Flags().Lookup(name)
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

// TestCollectionsGetArgs verifies get command requires exactly 1 argument
func TestCollectionsGetArgs(t *testing.T) {
	err := collectionsGetCmd.Args(collectionsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = collectionsGetCmd.Args(collectionsGetCmd, []string{"coll-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCollectionsDeleteArgs verifies delete command requires exactly 1 argument
func TestCollectionsDeleteArgs(t *testing.T) {
	err := collectionsDeleteCmd.Args(collectionsDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = collectionsDeleteCmd.Args(collectionsDeleteCmd, []string{"coll-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCollectionsGetClientError verifies error handling when getClient fails
func TestCollectionsGetClientError(t *testing.T) {
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

// TestCollectionsListGetClientError verifies list command error handling when getClient fails
func TestCollectionsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(collectionsListCmd)

	err := collectionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCollectionsGetGetClientError verifies get command error handling when getClient fails
func TestCollectionsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(collectionsGetCmd)

	err := collectionsGetCmd.RunE(cmd, []string{"coll-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCollectionsCreateGetClientError verifies create command error handling when getClient fails
func TestCollectionsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(collectionsCreateCmd)

	err := collectionsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCollectionsDeleteGetClientError verifies delete command error handling when getClient fails
func TestCollectionsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(collectionsDeleteCmd)

	err := collectionsDeleteCmd.RunE(cmd, []string{"coll-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCollectionsWithMockStore tests collections commands with a mock credential store
func TestCollectionsWithMockStore(t *testing.T) {
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

// TestCollectionsListWithValidStore tests list command execution with valid store
func TestCollectionsListWithValidStore(t *testing.T) {
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
	cmd.AddCommand(collectionsListCmd)

	// This will fail at the API call level, but validates the client setup works
	err := collectionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("collectionsListCmd succeeded (might be due to mock setup)")
	}
}

// collectionsMockAPIClient is a mock implementation of api.APIClient for collections tests.
type collectionsMockAPIClient struct {
	api.MockClient
	listCollectionsResp  *api.CollectionsListResponse
	listCollectionsErr   error
	getCollectionResp    *api.Collection
	getCollectionErr     error
	createCollectionResp *api.Collection
	createCollectionErr  error
	deleteCollectionErr  error
}

func (m *collectionsMockAPIClient) ListCollections(ctx context.Context, opts *api.CollectionsListOptions) (*api.CollectionsListResponse, error) {
	return m.listCollectionsResp, m.listCollectionsErr
}

func (m *collectionsMockAPIClient) GetCollection(ctx context.Context, id string) (*api.Collection, error) {
	return m.getCollectionResp, m.getCollectionErr
}

func (m *collectionsMockAPIClient) CreateCollection(ctx context.Context, req *api.CollectionCreateRequest) (*api.Collection, error) {
	return m.createCollectionResp, m.createCollectionErr
}

func (m *collectionsMockAPIClient) DeleteCollection(ctx context.Context, id string) error {
	return m.deleteCollectionErr
}

// setupCollectionsMockFactories sets up mock factories for collections tests.
func setupCollectionsMockFactories(mockClient *collectionsMockAPIClient) (func(), *bytes.Buffer) {
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

// newCollectionsTestCmd creates a test command with common flags.
func newCollectionsTestCmd() *cobra.Command {
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

// TestCollectionsListRunE tests the collections list command with mock API.
func TestCollectionsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CollectionsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CollectionsListResponse{
				Items: []api.Collection{
					{
						ID:            "coll_123",
						Title:         "Summer Collection",
						Handle:        "summer-collection",
						ProductsCount: 25,
						SortOrder:     "best-selling",
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "coll_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CollectionsListResponse{
				Items:      []api.Collection{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &collectionsMockAPIClient{
				listCollectionsResp: tt.mockResp,
				listCollectionsErr:  tt.mockErr,
			}
			cleanup, buf := setupCollectionsMockFactories(mockClient)
			defer cleanup()

			cmd := newCollectionsTestCmd()
			cmd.Flags().String("title", "", "")
			cmd.Flags().String("handle", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := collectionsListCmd.RunE(cmd, []string{})

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

// TestCollectionsListRunEWithJSON tests JSON output format.
func TestCollectionsListRunEWithJSON(t *testing.T) {
	mockClient := &collectionsMockAPIClient{
		listCollectionsResp: &api.CollectionsListResponse{
			Items: []api.Collection{
				{ID: "coll_json", Title: "JSON Collection"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCollectionsMockFactories(mockClient)
	defer cleanup()

	cmd := newCollectionsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := collectionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "coll_json") {
		t.Errorf("JSON output should contain collection ID, got: %s", output)
	}
}

// TestCollectionsGetRunE tests the collections get command with mock API.
func TestCollectionsGetRunE(t *testing.T) {
	tests := []struct {
		name         string
		collectionID string
		mockResp     *api.Collection
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "successful get",
			collectionID: "coll_123",
			mockResp: &api.Collection{
				ID:             "coll_123",
				Title:          "Summer Collection",
				Handle:         "summer-collection",
				Description:    "Our summer products",
				SortOrder:      "best-selling",
				ProductsCount:  25,
				PublishedScope: "global",
				PublishedAt:    time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
				CreatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:      time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:         "collection not found",
			collectionID: "coll_999",
			mockErr:      errors.New("collection not found"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &collectionsMockAPIClient{
				getCollectionResp: tt.mockResp,
				getCollectionErr:  tt.mockErr,
			}
			cleanup, _ := setupCollectionsMockFactories(mockClient)
			defer cleanup()

			cmd := newCollectionsTestCmd()

			err := collectionsGetCmd.RunE(cmd, []string{tt.collectionID})

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

// TestCollectionsGetRunEWithJSON tests JSON output format for get command.
func TestCollectionsGetRunEWithJSON(t *testing.T) {
	mockClient := &collectionsMockAPIClient{
		getCollectionResp: &api.Collection{
			ID:    "coll_json",
			Title: "JSON Test Collection",
		},
	}
	cleanup, buf := setupCollectionsMockFactories(mockClient)
	defer cleanup()

	cmd := newCollectionsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := collectionsGetCmd.RunE(cmd, []string{"coll_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "coll_json") {
		t.Errorf("JSON output should contain collection ID, got: %s", output)
	}
}

// TestCollectionsCreateRunE tests the collections create command with mock API.
func TestCollectionsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.Collection
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Collection{
				ID:     "coll_new",
				Title:  "New Collection",
				Handle: "new-collection",
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("validation error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &collectionsMockAPIClient{
				createCollectionResp: tt.mockResp,
				createCollectionErr:  tt.mockErr,
			}
			cleanup, _ := setupCollectionsMockFactories(mockClient)
			defer cleanup()

			cmd := newCollectionsTestCmd()
			cmd.Flags().String("title", "New Collection", "")
			cmd.Flags().String("handle", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("sort-order", "", "")

			err := collectionsCreateCmd.RunE(cmd, []string{})

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

// TestCollectionsCreateRunEWithJSON tests JSON output format for create command.
func TestCollectionsCreateRunEWithJSON(t *testing.T) {
	mockClient := &collectionsMockAPIClient{
		createCollectionResp: &api.Collection{
			ID:     "coll_json_new",
			Title:  "JSON New Collection",
			Handle: "json-new-collection",
		},
	}
	cleanup, buf := setupCollectionsMockFactories(mockClient)
	defer cleanup()

	cmd := newCollectionsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("title", "JSON New Collection", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("sort-order", "", "")

	err := collectionsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "coll_json_new") {
		t.Errorf("JSON output should contain collection ID, got: %s", output)
	}
}

// TestCollectionsDeleteRunE tests the collections delete command with mock API.
func TestCollectionsDeleteRunE(t *testing.T) {
	tests := []struct {
		name         string
		collectionID string
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "successful delete",
			collectionID: "coll_123",
			mockErr:      nil,
		},
		{
			name:         "delete fails",
			collectionID: "coll_456",
			mockErr:      errors.New("collection not found"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &collectionsMockAPIClient{
				deleteCollectionErr: tt.mockErr,
			}
			cleanup, _ := setupCollectionsMockFactories(mockClient)
			defer cleanup()

			cmd := newCollectionsTestCmd()

			err := collectionsDeleteCmd.RunE(cmd, []string{tt.collectionID})

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
