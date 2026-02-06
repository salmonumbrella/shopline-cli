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

// TestTaxonomiesCommandSetup verifies taxonomies command initialization
func TestTaxonomiesCommandSetup(t *testing.T) {
	if taxonomiesCmd.Use != "taxonomies" {
		t.Errorf("expected Use 'taxonomies', got %q", taxonomiesCmd.Use)
	}
	if taxonomiesCmd.Short != "Manage product taxonomies/categories" {
		t.Errorf("expected Short 'Manage product taxonomies/categories', got %q", taxonomiesCmd.Short)
	}
}

// TestTaxonomiesSubcommands verifies all subcommands are registered
func TestTaxonomiesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List taxonomies",
		"get":    "Get taxonomy details",
		"create": "Create a taxonomy",
		"delete": "Delete a taxonomy",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range taxonomiesCmd.Commands() {
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

// TestTaxonomiesListFlags verifies list command flags exist with correct defaults
func TestTaxonomiesListFlags(t *testing.T) {
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
			flag := taxonomiesListCmd.Flags().Lookup(f.name)
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

// TestTaxonomiesCreateFlags verifies create command flags exist with correct defaults
func TestTaxonomiesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"handle", ""},
		{"description", ""},
		{"parent-id", ""},
		{"position", "0"},
		{"active", "true"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := taxonomiesCreateCmd.Flags().Lookup(f.name)
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

// TestTaxonomiesGetClientError verifies error handling when getClient fails
func TestTaxonomiesGetClientError(t *testing.T) {
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

// TestTaxonomiesWithMockStore tests taxonomies commands with a mock credential store
func TestTaxonomiesWithMockStore(t *testing.T) {
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

// TestTaxonomiesCreateRequiredFlags verifies that name flag is required
func TestTaxonomiesCreateRequiredFlags(t *testing.T) {
	flag := taxonomiesCreateCmd.Flags().Lookup("name")
	if flag == nil {
		t.Error("name flag not found")
		return
	}
	// Verify it exists and is a string flag
	if flag.Value.Type() != "string" {
		t.Error("name flag should be a string type")
	}
}

// TestTaxonomiesDeleteArgs verifies delete command requires exactly one argument
func TestTaxonomiesDeleteArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if taxonomiesDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", taxonomiesDeleteCmd.Use)
	}

	// Test Args validation
	err := taxonomiesDeleteCmd.Args(taxonomiesDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = taxonomiesDeleteCmd.Args(taxonomiesDeleteCmd, []string{"tax-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestTaxonomiesGetArgs verifies get command requires exactly one argument
func TestTaxonomiesGetArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if taxonomiesGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", taxonomiesGetCmd.Use)
	}

	// Test Args validation
	err := taxonomiesGetCmd.Args(taxonomiesGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = taxonomiesGetCmd.Args(taxonomiesGetCmd, []string{"tax-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestTaxonomiesListFlagDescriptions verifies flag descriptions are set
func TestTaxonomiesListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"parent-id": "Filter by parent taxonomy ID",
		"page":      "Page number",
		"page-size": "Results per page",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := taxonomiesListCmd.Flags().Lookup(flagName)
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

// TestTaxonomiesCreateFlagDescriptions verifies create flag descriptions
func TestTaxonomiesCreateFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"name":        "Taxonomy name (required)",
		"handle":      "URL handle (auto-generated if not provided)",
		"description": "Taxonomy description",
		"parent-id":   "Parent taxonomy ID for nested categories",
		"position":    "Position in the list",
		"active":      "Taxonomy active status",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := taxonomiesCreateCmd.Flags().Lookup(flagName)
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

// TestTaxonomiesCreateFlagTypes verifies flag types are correct
func TestTaxonomiesCreateFlagTypes(t *testing.T) {
	flags := map[string]string{
		"name":        "string",
		"handle":      "string",
		"description": "string",
		"parent-id":   "string",
		"position":    "int",
		"active":      "bool",
	}

	for flagName, expectedType := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := taxonomiesCreateCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Value.Type() != expectedType {
				t.Errorf("expected type %q, got %q", expectedType, flag.Value.Type())
			}
		})
	}
}

// taxonomiesMockAPIClient is a mock implementation of api.APIClient for taxonomies tests.
type taxonomiesMockAPIClient struct {
	api.MockClient
	listTaxonomiesResp *api.TaxonomiesListResponse
	listTaxonomiesErr  error
	getTaxonomyResp    *api.Taxonomy
	getTaxonomyErr     error
	createTaxonomyResp *api.Taxonomy
	createTaxonomyErr  error
	deleteTaxonomyErr  error
}

func (m *taxonomiesMockAPIClient) ListTaxonomies(ctx context.Context, opts *api.TaxonomiesListOptions) (*api.TaxonomiesListResponse, error) {
	return m.listTaxonomiesResp, m.listTaxonomiesErr
}

func (m *taxonomiesMockAPIClient) GetTaxonomy(ctx context.Context, id string) (*api.Taxonomy, error) {
	return m.getTaxonomyResp, m.getTaxonomyErr
}

func (m *taxonomiesMockAPIClient) CreateTaxonomy(ctx context.Context, req *api.TaxonomyCreateRequest) (*api.Taxonomy, error) {
	return m.createTaxonomyResp, m.createTaxonomyErr
}

func (m *taxonomiesMockAPIClient) DeleteTaxonomy(ctx context.Context, id string) error {
	return m.deleteTaxonomyErr
}

// setupTaxonomiesMockFactories sets up mock factories for taxonomies tests.
func setupTaxonomiesMockFactories(mockClient *taxonomiesMockAPIClient) (func(), *bytes.Buffer) {
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

// newTaxonomiesTestCmd creates a test command with common flags for taxonomies tests.
func newTaxonomiesTestCmd() *cobra.Command {
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

// TestTaxonomiesListRunE tests the taxonomies list command with mock API.
func TestTaxonomiesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.TaxonomiesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.TaxonomiesListResponse{
				Items: []api.Taxonomy{
					{
						ID:           "tax_123",
						Name:         "Electronics",
						Handle:       "electronics",
						Level:        1,
						ProductCount: 50,
						Active:       true,
						CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "tax_123",
		},
		{
			name: "successful list with multiple taxonomies",
			mockResp: &api.TaxonomiesListResponse{
				Items: []api.Taxonomy{
					{
						ID:           "tax_001",
						Name:         "Clothing",
						Handle:       "clothing",
						Level:        1,
						ProductCount: 100,
						Active:       true,
						CreatedAt:    time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
					},
					{
						ID:           "tax_002",
						Name:         "Shoes",
						Handle:       "shoes",
						ParentID:     "tax_001",
						Level:        2,
						ProductCount: 25,
						Active:       false,
						CreatedAt:    time.Date(2024, 1, 12, 9, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "tax_001",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.TaxonomiesListResponse{
				Items:      []api.Taxonomy{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxonomiesMockAPIClient{
				listTaxonomiesResp: tt.mockResp,
				listTaxonomiesErr:  tt.mockErr,
			}
			cleanup, buf := setupTaxonomiesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxonomiesTestCmd()
			cmd.Flags().String("parent-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := taxonomiesListCmd.RunE(cmd, []string{})

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

// TestTaxonomiesListRunEWithJSON tests JSON output format.
func TestTaxonomiesListRunEWithJSON(t *testing.T) {
	mockClient := &taxonomiesMockAPIClient{
		listTaxonomiesResp: &api.TaxonomiesListResponse{
			Items: []api.Taxonomy{
				{ID: "tax_json", Name: "JSON Taxonomy", Handle: "json-taxonomy"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupTaxonomiesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxonomiesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := taxonomiesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_json") {
		t.Errorf("JSON output should contain taxonomy ID, got: %s", output)
	}
}

// TestTaxonomiesListRunEWithParentID tests list filtering by parent ID.
func TestTaxonomiesListRunEWithParentID(t *testing.T) {
	mockClient := &taxonomiesMockAPIClient{
		listTaxonomiesResp: &api.TaxonomiesListResponse{
			Items: []api.Taxonomy{
				{
					ID:       "tax_child",
					Name:     "Child Category",
					Handle:   "child-category",
					ParentID: "tax_parent",
					Level:    2,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupTaxonomiesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxonomiesTestCmd()
	cmd.Flags().String("parent-id", "", "")
	_ = cmd.Flags().Set("parent-id", "tax_parent")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := taxonomiesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_child") {
		t.Errorf("output should contain child taxonomy ID, got: %s", output)
	}
}

// TestTaxonomiesGetRunE tests the taxonomies get command with mock API.
func TestTaxonomiesGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		taxonomyID string
		mockResp   *api.Taxonomy
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful get",
			taxonomyID: "tax_123",
			mockResp: &api.Taxonomy{
				ID:           "tax_123",
				Name:         "Electronics",
				Handle:       "electronics",
				Description:  "Electronic products",
				Level:        1,
				Position:     1,
				Path:         "electronics",
				FullPath:     "Electronics",
				ProductCount: 50,
				Active:       true,
				CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "successful get with parent",
			taxonomyID: "tax_456",
			mockResp: &api.Taxonomy{
				ID:           "tax_456",
				Name:         "Laptops",
				Handle:       "laptops",
				Description:  "Laptop computers",
				ParentID:     "tax_123",
				Level:        2,
				Position:     1,
				Path:         "electronics/laptops",
				FullPath:     "Electronics > Laptops",
				ProductCount: 20,
				Active:       true,
				CreatedAt:    time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "taxonomy not found",
			taxonomyID: "tax_999",
			mockErr:    errors.New("taxonomy not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxonomiesMockAPIClient{
				getTaxonomyResp: tt.mockResp,
				getTaxonomyErr:  tt.mockErr,
			}
			cleanup, _ := setupTaxonomiesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxonomiesTestCmd()

			err := taxonomiesGetCmd.RunE(cmd, []string{tt.taxonomyID})

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

// TestTaxonomiesGetRunEWithJSON tests JSON output format for get command.
func TestTaxonomiesGetRunEWithJSON(t *testing.T) {
	mockClient := &taxonomiesMockAPIClient{
		getTaxonomyResp: &api.Taxonomy{
			ID:          "tax_json",
			Name:        "JSON Test Taxonomy",
			Handle:      "json-test",
			Description: "Test taxonomy for JSON output",
			Active:      true,
		},
	}
	cleanup, buf := setupTaxonomiesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxonomiesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := taxonomiesGetCmd.RunE(cmd, []string{"tax_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_json") {
		t.Errorf("JSON output should contain taxonomy ID, got: %s", output)
	}
}

// TestTaxonomiesCreateRunE tests the taxonomies create command with mock API.
func TestTaxonomiesCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.Taxonomy
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Taxonomy{
				ID:     "tax_new",
				Name:   "New Taxonomy",
				Handle: "new-taxonomy",
				Level:  1,
				Active: true,
			},
		},
		{
			name: "successful create with all options",
			mockResp: &api.Taxonomy{
				ID:          "tax_full",
				Name:        "Full Taxonomy",
				Handle:      "full-taxonomy",
				Description: "Complete taxonomy with all options",
				ParentID:    "tax_parent",
				Level:       2,
				Position:    5,
				Active:      true,
			},
		},
		{
			name:    "create fails - validation error",
			mockErr: errors.New("validation error: name is required"),
			wantErr: true,
		},
		{
			name:    "create fails - duplicate handle",
			mockErr: errors.New("handle already exists"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxonomiesMockAPIClient{
				createTaxonomyResp: tt.mockResp,
				createTaxonomyErr:  tt.mockErr,
			}
			cleanup, _ := setupTaxonomiesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxonomiesTestCmd()
			cmd.Flags().String("name", "New Taxonomy", "")
			cmd.Flags().String("handle", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("parent-id", "", "")
			cmd.Flags().Int("position", 0, "")
			cmd.Flags().Bool("active", true, "")

			err := taxonomiesCreateCmd.RunE(cmd, []string{})

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

// TestTaxonomiesCreateRunEWithJSON tests JSON output format for create command.
func TestTaxonomiesCreateRunEWithJSON(t *testing.T) {
	mockClient := &taxonomiesMockAPIClient{
		createTaxonomyResp: &api.Taxonomy{
			ID:     "tax_json_new",
			Name:   "JSON New Taxonomy",
			Handle: "json-new-taxonomy",
			Level:  1,
			Active: true,
		},
	}
	cleanup, buf := setupTaxonomiesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxonomiesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("name", "JSON New Taxonomy", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("position", 0, "")
	cmd.Flags().Bool("active", true, "")

	err := taxonomiesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_json_new") {
		t.Errorf("JSON output should contain taxonomy ID, got: %s", output)
	}
}

// TestTaxonomiesCreateRunEDryRun tests dry-run mode for create command.
func TestTaxonomiesCreateRunEDryRun(t *testing.T) {
	// No mock client needed for dry-run since it doesn't make API calls
	origSecretsFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origSecretsFactory }()

	cmd := newTaxonomiesTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("name", "Test Taxonomy", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("position", 0, "")
	cmd.Flags().Bool("active", true, "")

	err := taxonomiesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestTaxonomiesDeleteRunE tests the taxonomies delete command with mock API.
func TestTaxonomiesDeleteRunE(t *testing.T) {
	tests := []struct {
		name       string
		taxonomyID string
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful delete",
			taxonomyID: "tax_123",
			mockErr:    nil,
		},
		{
			name:       "delete fails - taxonomy not found",
			taxonomyID: "tax_999",
			mockErr:    errors.New("taxonomy not found"),
			wantErr:    true,
		},
		{
			name:       "delete fails - has children",
			taxonomyID: "tax_parent",
			mockErr:    errors.New("cannot delete taxonomy with children"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxonomiesMockAPIClient{
				deleteTaxonomyErr: tt.mockErr,
			}
			cleanup, _ := setupTaxonomiesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxonomiesTestCmd()

			err := taxonomiesDeleteCmd.RunE(cmd, []string{tt.taxonomyID})

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

// TestTaxonomiesDeleteRunEDryRun tests dry-run mode for delete command.
func TestTaxonomiesDeleteRunEDryRun(t *testing.T) {
	// No mock client needed for dry-run since it doesn't make API calls
	origSecretsFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origSecretsFactory }()

	cmd := newTaxonomiesTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := taxonomiesDeleteCmd.RunE(cmd, []string{"tax_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestTaxonomiesDeleteRunENoConfirmation tests delete without --yes flag.
func TestTaxonomiesDeleteRunENoConfirmation(t *testing.T) {
	cmd := newTaxonomiesTestCmd()
	_ = cmd.Flags().Set("yes", "false")

	// This should just print a message and return nil (asking for confirmation)
	err := taxonomiesDeleteCmd.RunE(cmd, []string{"tax_123"})
	if err != nil {
		t.Errorf("unexpected error when confirmation not provided: %v", err)
	}
}

// TestTaxonomiesListGetClientError verifies list command error handling when getClient fails.
func TestTaxonomiesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := taxonomiesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestTaxonomiesGetGetClientError verifies get command error handling when getClient fails.
func TestTaxonomiesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := taxonomiesGetCmd.RunE(cmd, []string{"tax_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestTaxonomiesCreateGetClientError verifies create command error handling when getClient fails.
func TestTaxonomiesCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("position", 0, "")
	cmd.Flags().Bool("active", true, "")

	err := taxonomiesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestTaxonomiesDeleteGetClientError verifies delete command error handling when getClient fails.
func TestTaxonomiesDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	// Must provide --yes flag to skip confirmation and actually try to get client
	_ = cmd.Flags().Set("yes", "true")

	err := taxonomiesDeleteCmd.RunE(cmd, []string{"tax_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestTaxonomiesListWithNoProfiles verifies error when no profiles are configured.
func TestTaxonomiesListWithNoProfiles(t *testing.T) {
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

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := taxonomiesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for no profiles, got nil")
	}
}

// TestTaxonomiesGetWithMultipleProfiles verifies error when multiple profiles exist without selection.
func TestTaxonomiesGetWithMultipleProfiles(t *testing.T) {
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

	cmd := newTestCmdWithFlags()

	err := taxonomiesGetCmd.RunE(cmd, []string{"tax_123"})
	if err == nil {
		t.Fatal("expected error for multiple profiles, got nil")
	}
}
