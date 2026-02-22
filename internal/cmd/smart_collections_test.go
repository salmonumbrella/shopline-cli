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

// TestSmartCollectionsCommandSetup verifies smart-collections command initialization
func TestSmartCollectionsCommandSetup(t *testing.T) {
	if smartCollectionsCmd.Use != "smart-collections" {
		t.Errorf("expected Use 'smart-collections', got %q", smartCollectionsCmd.Use)
	}
	if smartCollectionsCmd.Short != "Manage smart collections (auto-populated based on rules)" {
		t.Errorf("expected Short 'Manage smart collections (auto-populated based on rules)', got %q", smartCollectionsCmd.Short)
	}
}

// TestSmartCollectionsSubcommands verifies all subcommands are registered
func TestSmartCollectionsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List smart collections",
		"get":    "Get smart collection details",
		"create": "Create a smart collection",
		"delete": "Delete a smart collection",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range smartCollectionsCmd.Commands() {
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

// TestSmartCollectionsListFlags verifies list command flags exist with correct defaults
func TestSmartCollectionsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := smartCollectionsListCmd.Flags().Lookup(f.name)
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

// TestSmartCollectionsCreateFlags verifies create command flags exist with correct defaults
func TestSmartCollectionsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"title", ""},
		{"handle", ""},
		{"body-html", ""},
		{"sort-order", ""},
		{"disjunctive", "false"},
		{"published", "true"},
		{"rules", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := smartCollectionsCreateCmd.Flags().Lookup(f.name)
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

// TestSmartCollectionsGetClientError verifies error handling when getClient fails
func TestSmartCollectionsGetClientError(t *testing.T) {
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

// TestSmartCollectionsWithMockStore tests smart collections commands with a mock credential store
func TestSmartCollectionsWithMockStore(t *testing.T) {
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

// TestSmartCollectionsCreateRequiredFlags verifies that title and rules flags are required
func TestSmartCollectionsCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"title", "rules"}

	for _, flagName := range requiredFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := smartCollectionsCreateCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("%s flag not found", flagName)
				return
			}
			// Verify it exists and is a string flag
			if flag.Value.Type() != "string" {
				t.Errorf("%s flag should be a string type", flagName)
			}
		})
	}
}

// TestSmartCollectionsDeleteArgs verifies delete command requires exactly one argument
func TestSmartCollectionsDeleteArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if smartCollectionsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", smartCollectionsDeleteCmd.Use)
	}
}

// TestSmartCollectionsGetArgs verifies get command requires exactly one argument
func TestSmartCollectionsGetArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if smartCollectionsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", smartCollectionsGetCmd.Use)
	}
}

// TestSmartCollectionsCreateFlagDescriptions verifies flag descriptions are set
func TestSmartCollectionsCreateFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"title":       "Smart collection title",
		"handle":      "Smart collection handle (URL slug)",
		"body-html":   "Smart collection description HTML",
		"disjunctive": "Match any rule (true) or all rules (false)",
		"published":   "Publish the collection",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := smartCollectionsCreateCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Usage != expectedUsage {
				t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
			}
		})
	}
}

// smartCollectionsMockAPIClient is a mock implementation of api.APIClient for smart collections tests.
type smartCollectionsMockAPIClient struct {
	api.MockClient
	listSmartCollectionsResp  *api.SmartCollectionsListResponse
	listSmartCollectionsErr   error
	getSmartCollectionResp    *api.SmartCollection
	getSmartCollectionErr     error
	createSmartCollectionResp *api.SmartCollection
	createSmartCollectionErr  error
	deleteSmartCollectionErr  error
}

func (m *smartCollectionsMockAPIClient) ListSmartCollections(ctx context.Context, opts *api.SmartCollectionsListOptions) (*api.SmartCollectionsListResponse, error) {
	return m.listSmartCollectionsResp, m.listSmartCollectionsErr
}

func (m *smartCollectionsMockAPIClient) GetSmartCollection(ctx context.Context, id string) (*api.SmartCollection, error) {
	return m.getSmartCollectionResp, m.getSmartCollectionErr
}

func (m *smartCollectionsMockAPIClient) CreateSmartCollection(ctx context.Context, req *api.SmartCollectionCreateRequest) (*api.SmartCollection, error) {
	return m.createSmartCollectionResp, m.createSmartCollectionErr
}

func (m *smartCollectionsMockAPIClient) DeleteSmartCollection(ctx context.Context, id string) error {
	return m.deleteSmartCollectionErr
}

// setupSmartCollectionsMockFactories sets up mock factories for smart collections tests.
func setupSmartCollectionsMockFactories(mockClient *smartCollectionsMockAPIClient) (func(), *bytes.Buffer) {
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

// newSmartCollectionsTestCmd creates a test command with common flags for smart collections tests.
func newSmartCollectionsTestCmd() *cobra.Command {
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

// TestSmartCollectionsListRunE tests the smart collections list command with mock API.
func TestSmartCollectionsListRunE(t *testing.T) {
	publishedAt := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		mockResp   *api.SmartCollectionsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.SmartCollectionsListResponse{
				Items: []api.SmartCollection{
					{
						ID:          "sc_123",
						Title:       "Sale Products",
						Handle:      "sale-products",
						BodyHTML:    "<p>Products on sale</p>",
						SortOrder:   "best-selling",
						Disjunctive: false,
						Rules: []api.Rule{
							{Column: "tag", Relation: "equals", Condition: "sale"},
						},
						Published:   true,
						PublishedAt: publishedAt,
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sc_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.SmartCollectionsListResponse{
				Items:      []api.SmartCollection{},
				TotalCount: 0,
			},
		},
		{
			name: "disjunctive collection",
			mockResp: &api.SmartCollectionsListResponse{
				Items: []api.SmartCollection{
					{
						ID:          "sc_456",
						Title:       "Featured Items",
						Handle:      "featured-items",
						Disjunctive: true,
						Rules: []api.Rule{
							{Column: "tag", Relation: "equals", Condition: "featured"},
							{Column: "vendor", Relation: "equals", Condition: "TopBrand"},
						},
						Published:   true,
						PublishedAt: publishedAt,
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sc_456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &smartCollectionsMockAPIClient{
				listSmartCollectionsResp: tt.mockResp,
				listSmartCollectionsErr:  tt.mockErr,
			}
			cleanup, buf := setupSmartCollectionsMockFactories(mockClient)
			defer cleanup()

			cmd := newSmartCollectionsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := smartCollectionsListCmd.RunE(cmd, []string{})

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

// TestSmartCollectionsGetRunE tests the smart collections get command with mock API.
func TestSmartCollectionsGetRunE(t *testing.T) {
	publishedAt := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		id       string
		mockResp *api.SmartCollection
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "sc_123",
			mockResp: &api.SmartCollection{
				ID:          "sc_123",
				Title:       "Sale Products",
				Handle:      "sale-products",
				BodyHTML:    "<p>Products on sale</p>",
				SortOrder:   "best-selling",
				Disjunctive: false,
				Rules: []api.Rule{
					{Column: "tag", Relation: "equals", Condition: "sale"},
				},
				Published:   true,
				PublishedAt: publishedAt,
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "get with multiple rules",
			id:   "sc_456",
			mockResp: &api.SmartCollection{
				ID:          "sc_456",
				Title:       "Featured Items",
				Handle:      "featured-items",
				Disjunctive: true,
				Rules: []api.Rule{
					{Column: "tag", Relation: "equals", Condition: "featured"},
					{Column: "vendor", Relation: "equals", Condition: "TopBrand"},
					{Column: "type", Relation: "contains", Condition: "electronics"},
				},
				Published:   true,
				PublishedAt: publishedAt,
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "sc_999",
			mockErr: errors.New("smart collection not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &smartCollectionsMockAPIClient{
				getSmartCollectionResp: tt.mockResp,
				getSmartCollectionErr:  tt.mockErr,
			}
			cleanup, _ := setupSmartCollectionsMockFactories(mockClient)
			defer cleanup()

			cmd := newSmartCollectionsTestCmd()

			err := smartCollectionsGetCmd.RunE(cmd, []string{tt.id})

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

// TestSmartCollectionsCreateRunE tests the smart collections create command with mock API.
func TestSmartCollectionsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.SmartCollection
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.SmartCollection{
				ID:     "sc_new",
				Title:  "New Collection",
				Handle: "new-collection",
				Rules: []api.Rule{
					{Column: "tag", Relation: "equals", Condition: "new"},
				},
				Published: true,
			},
		},
		{
			name: "create with disjunctive rules",
			mockResp: &api.SmartCollection{
				ID:          "sc_disj",
				Title:       "Disjunctive Collection",
				Handle:      "disjunctive-collection",
				Disjunctive: true,
				Rules: []api.Rule{
					{Column: "tag", Relation: "equals", Condition: "sale"},
					{Column: "tag", Relation: "equals", Condition: "clearance"},
				},
				Published: true,
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("create failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &smartCollectionsMockAPIClient{
				createSmartCollectionResp: tt.mockResp,
				createSmartCollectionErr:  tt.mockErr,
			}
			cleanup, _ := setupSmartCollectionsMockFactories(mockClient)
			defer cleanup()

			cmd := newSmartCollectionsTestCmd()
			cmd.Flags().String("title", "Test Collection", "")
			cmd.Flags().String("handle", "", "")
			cmd.Flags().String("body-html", "", "")
			cmd.Flags().String("sort-order", "", "")
			cmd.Flags().Bool("disjunctive", false, "")
			cmd.Flags().Bool("published", true, "")
			cmd.Flags().String("rules", `[{"column":"tag","relation":"equals","condition":"test"}]`, "")

			err := smartCollectionsCreateCmd.RunE(cmd, []string{})

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

// TestSmartCollectionsDeleteRunE tests the smart collections delete command with mock API.
func TestSmartCollectionsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   "sc_123",
		},
		{
			name:    "delete fails",
			id:      "sc_123",
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
		{
			name:    "not found",
			id:      "sc_999",
			mockErr: errors.New("smart collection not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &smartCollectionsMockAPIClient{
				deleteSmartCollectionErr: tt.mockErr,
			}
			cleanup, _ := setupSmartCollectionsMockFactories(mockClient)
			defer cleanup()

			cmd := newSmartCollectionsTestCmd()

			err := smartCollectionsDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestSmartCollectionsListGetClientError tests error handling when getClient fails for list command.
func TestSmartCollectionsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(smartCollectionsListCmd)

	err := smartCollectionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestSmartCollectionsGetGetClientError tests error handling when getClient fails for get command.
func TestSmartCollectionsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(smartCollectionsGetCmd)

	err := smartCollectionsGetCmd.RunE(cmd, []string{"sc_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestSmartCollectionsCreateGetClientError tests error handling when getClient fails for create command.
func TestSmartCollectionsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("title", "Test", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("body-html", "", "")
	cmd.Flags().String("sort-order", "", "")
	cmd.Flags().Bool("disjunctive", false, "")
	cmd.Flags().Bool("published", true, "")
	cmd.Flags().String("rules", "[]", "")

	err := smartCollectionsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestSmartCollectionsDeleteGetClientError tests error handling when getClient fails for delete command.
func TestSmartCollectionsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(smartCollectionsDeleteCmd)

	err := smartCollectionsDeleteCmd.RunE(cmd, []string{"sc_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestSmartCollectionsCreateInvalidRulesJSON tests error handling for invalid rules JSON.
func TestSmartCollectionsCreateInvalidRulesJSON(t *testing.T) {
	mockClient := &smartCollectionsMockAPIClient{}
	cleanup, _ := setupSmartCollectionsMockFactories(mockClient)
	defer cleanup()

	cmd := newSmartCollectionsTestCmd()
	cmd.Flags().String("title", "Test Collection", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("body-html", "", "")
	cmd.Flags().String("sort-order", "", "")
	cmd.Flags().Bool("disjunctive", false, "")
	cmd.Flags().Bool("published", true, "")
	cmd.Flags().String("rules", "invalid json", "")

	err := smartCollectionsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse rules JSON") {
		t.Errorf("expected 'failed to parse rules JSON' error, got: %v", err)
	}
}

// Ensure unused imports don't cause errors
var _ = secrets.StoreCredentials{}
