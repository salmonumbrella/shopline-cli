package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// metafieldDefinitionsMockAPIClient is a mock implementation of api.APIClient for metafield definitions tests.
type metafieldDefinitionsMockAPIClient struct {
	api.MockClient
	listResp   *api.MetafieldDefinitionsListResponse
	listErr    error
	getResp    *api.MetafieldDefinition
	getErr     error
	createResp *api.MetafieldDefinition
	createErr  error
	deleteErr  error
}

func (m *metafieldDefinitionsMockAPIClient) ListMetafieldDefinitions(ctx context.Context, opts *api.MetafieldDefinitionsListOptions) (*api.MetafieldDefinitionsListResponse, error) {
	return m.listResp, m.listErr
}

func (m *metafieldDefinitionsMockAPIClient) GetMetafieldDefinition(ctx context.Context, id string) (*api.MetafieldDefinition, error) {
	return m.getResp, m.getErr
}

func (m *metafieldDefinitionsMockAPIClient) CreateMetafieldDefinition(ctx context.Context, req *api.MetafieldDefinitionCreateRequest) (*api.MetafieldDefinition, error) {
	return m.createResp, m.createErr
}

func (m *metafieldDefinitionsMockAPIClient) DeleteMetafieldDefinition(ctx context.Context, id string) error {
	return m.deleteErr
}

// setupMetafieldDefinitionsMockFactories sets up mock factories for metafield definitions tests.
func setupMetafieldDefinitionsMockFactories(mockClient *metafieldDefinitionsMockAPIClient) (func(), *bytes.Buffer) {
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

// newMetafieldDefinitionsTestCmd creates a test command with common flags for metafield definitions tests.
func newMetafieldDefinitionsTestCmd() *cobra.Command {
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

func TestMetafieldDefinitionsCommandStructure(t *testing.T) {
	subcommands := metafieldDefinitionsCmd.Commands()

	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"delete": false,
	}

	for _, cmd := range subcommands {
		if _, exists := expectedCmds[cmd.Name()]; exists {
			expectedCmds[cmd.Name()] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestMetafieldDefinitionsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"owner-type", ""},
		{"namespace", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := metafieldDefinitionsListCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Expected flag %q not found on list command", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestMetafieldDefinitionsGetArgs(t *testing.T) {
	err := metafieldDefinitionsGetCmd.Args(metafieldDefinitionsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = metafieldDefinitionsGetCmd.Args(metafieldDefinitionsGetCmd, []string{"def-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestMetafieldDefinitionsDeleteArgs(t *testing.T) {
	err := metafieldDefinitionsDeleteCmd.Args(metafieldDefinitionsDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = metafieldDefinitionsDeleteCmd.Args(metafieldDefinitionsDeleteCmd, []string{"def-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestMetafieldDefinitionsCreateFlags(t *testing.T) {
	flags := []string{"name", "namespace", "key", "type", "owner-type", "description"}

	for _, flagName := range flags {
		if metafieldDefinitionsCreateCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on create command", flagName)
		}
	}
}

func TestMetafieldDefinitionsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMetafieldDefinitionsTestCmd()
	cmd.Flags().String("owner-type", "", "")
	cmd.Flags().String("namespace", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := metafieldDefinitionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestMetafieldDefinitionsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMetafieldDefinitionsTestCmd()

	err := metafieldDefinitionsGetCmd.RunE(cmd, []string{"def-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestMetafieldDefinitionsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMetafieldDefinitionsTestCmd()
	cmd.Flags().String("name", "test", "")
	cmd.Flags().String("namespace", "custom", "")
	cmd.Flags().String("key", "testkey", "")
	cmd.Flags().String("type", "single_line_text_field", "")
	cmd.Flags().String("owner-type", "product", "")
	cmd.Flags().String("description", "", "")

	err := metafieldDefinitionsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestMetafieldDefinitionsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMetafieldDefinitionsTestCmd()

	err := metafieldDefinitionsDeleteCmd.RunE(cmd, []string{"def-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestMetafieldDefinitionsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.MetafieldDefinitionsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.MetafieldDefinitionsListResponse{
				Items: []api.MetafieldDefinition{
					{
						ID:        "def_123",
						Name:      "Color",
						Namespace: "custom",
						Key:       "color",
						Type:      "single_line_text_field",
						OwnerType: "product",
					},
					{
						ID:        "def_456",
						Name:      "Size",
						Namespace: "custom",
						Key:       "size",
						Type:      "number_integer",
						OwnerType: "variant",
					},
				},
				TotalCount: 2,
			},
			wantOutput: "def_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.MetafieldDefinitionsListResponse{
				Items:      []api.MetafieldDefinition{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &metafieldDefinitionsMockAPIClient{
				listResp: tt.mockResp,
				listErr:  tt.mockErr,
			}
			cleanup, buf := setupMetafieldDefinitionsMockFactories(mockClient)
			defer cleanup()

			cmd := newMetafieldDefinitionsTestCmd()
			cmd.Flags().String("owner-type", "", "")
			cmd.Flags().String("namespace", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := metafieldDefinitionsListCmd.RunE(cmd, []string{})

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

func TestMetafieldDefinitionsListRunEWithFilters(t *testing.T) {
	mockClient := &metafieldDefinitionsMockAPIClient{
		listResp: &api.MetafieldDefinitionsListResponse{
			Items: []api.MetafieldDefinition{
				{
					ID:        "def_123",
					Name:      "Color",
					Namespace: "custom",
					Key:       "color",
					Type:      "single_line_text_field",
					OwnerType: "product",
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupMetafieldDefinitionsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldDefinitionsTestCmd()
	cmd.Flags().String("owner-type", "product", "")
	cmd.Flags().String("namespace", "custom", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")

	err := metafieldDefinitionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "def_123") {
		t.Errorf("output should contain 'def_123', got: %s", output)
	}
}

func TestMetafieldDefinitionsListRunEJSON(t *testing.T) {
	mockClient := &metafieldDefinitionsMockAPIClient{
		listResp: &api.MetafieldDefinitionsListResponse{
			Items: []api.MetafieldDefinition{
				{
					ID:        "def_123",
					Name:      "Color",
					Namespace: "custom",
					Key:       "color",
					Type:      "single_line_text_field",
					OwnerType: "product",
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupMetafieldDefinitionsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldDefinitionsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("owner-type", "", "")
	cmd.Flags().String("namespace", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := metafieldDefinitionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "def_123") {
		t.Errorf("JSON output should contain 'def_123', got: %s", output)
	}
}

func TestMetafieldDefinitionsGetRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		defID      string
		mockResp   *api.MetafieldDefinition
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:  "successful get",
			defID: "def_123",
			mockResp: &api.MetafieldDefinition{
				ID:          "def_123",
				Name:        "Color",
				Namespace:   "custom",
				Key:         "color",
				Type:        "single_line_text_field",
				OwnerType:   "product",
				Description: "Product color field",
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
		},
		{
			name:    "not found",
			defID:   "def_999",
			mockErr: errors.New("metafield definition not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &metafieldDefinitionsMockAPIClient{
				getResp: tt.mockResp,
				getErr:  tt.mockErr,
			}
			cleanup, _ := setupMetafieldDefinitionsMockFactories(mockClient)
			defer cleanup()

			cmd := newMetafieldDefinitionsTestCmd()

			err := metafieldDefinitionsGetCmd.RunE(cmd, []string{tt.defID})

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

func TestMetafieldDefinitionsGetRunEWithValidations(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &metafieldDefinitionsMockAPIClient{
		getResp: &api.MetafieldDefinition{
			ID:          "def_123",
			Name:        "Weight",
			Namespace:   "custom",
			Key:         "weight",
			Type:        "number_decimal",
			OwnerType:   "product",
			Description: "Product weight in kg",
			Validations: []api.Validation{
				{Name: "min", Type: "number", Value: "0"},
				{Name: "max", Type: "number", Value: "1000"},
			},
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, _ := setupMetafieldDefinitionsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldDefinitionsTestCmd()

	err := metafieldDefinitionsGetCmd.RunE(cmd, []string{"def_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMetafieldDefinitionsGetRunEJSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &metafieldDefinitionsMockAPIClient{
		getResp: &api.MetafieldDefinition{
			ID:          "def_123",
			Name:        "Color",
			Namespace:   "custom",
			Key:         "color",
			Type:        "single_line_text_field",
			OwnerType:   "product",
			Description: "Product color field",
			CreatedAt:   testTime,
			UpdatedAt:   testTime,
		},
	}
	cleanup, buf := setupMetafieldDefinitionsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldDefinitionsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := metafieldDefinitionsGetCmd.RunE(cmd, []string{"def_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "def_123") {
		t.Errorf("JSON output should contain 'def_123', got: %s", output)
	}
}

func TestMetafieldDefinitionsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.MetafieldDefinition
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.MetafieldDefinition{
				ID:        "def_new",
				Name:      "Material",
				Namespace: "custom",
				Key:       "material",
				Type:      "single_line_text_field",
				OwnerType: "product",
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("validation failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &metafieldDefinitionsMockAPIClient{
				createResp: tt.mockResp,
				createErr:  tt.mockErr,
			}
			cleanup, _ := setupMetafieldDefinitionsMockFactories(mockClient)
			defer cleanup()

			cmd := newMetafieldDefinitionsTestCmd()
			cmd.Flags().String("name", "Material", "")
			cmd.Flags().String("namespace", "custom", "")
			cmd.Flags().String("key", "material", "")
			cmd.Flags().String("type", "single_line_text_field", "")
			cmd.Flags().String("owner-type", "product", "")
			cmd.Flags().String("description", "Material description", "")

			err := metafieldDefinitionsCreateCmd.RunE(cmd, []string{})

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

func TestMetafieldDefinitionsCreateRunEJSON(t *testing.T) {
	mockClient := &metafieldDefinitionsMockAPIClient{
		createResp: &api.MetafieldDefinition{
			ID:        "def_new",
			Name:      "Material",
			Namespace: "custom",
			Key:       "material",
			Type:      "single_line_text_field",
			OwnerType: "product",
		},
	}
	cleanup, buf := setupMetafieldDefinitionsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldDefinitionsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("name", "Material", "")
	cmd.Flags().String("namespace", "custom", "")
	cmd.Flags().String("key", "material", "")
	cmd.Flags().String("type", "single_line_text_field", "")
	cmd.Flags().String("owner-type", "product", "")
	cmd.Flags().String("description", "", "")

	err := metafieldDefinitionsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "def_new") {
		t.Errorf("JSON output should contain 'def_new', got: %s", output)
	}
}

func TestMetafieldDefinitionsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		defID   string
		mockErr error
		wantErr bool
	}{
		{
			name:  "successful delete",
			defID: "def_123",
		},
		{
			name:    "delete error",
			defID:   "def_999",
			mockErr: errors.New("metafield definition not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &metafieldDefinitionsMockAPIClient{
				deleteErr: tt.mockErr,
			}
			cleanup, _ := setupMetafieldDefinitionsMockFactories(mockClient)
			defer cleanup()

			cmd := newMetafieldDefinitionsTestCmd()
			// Set yes=true to skip confirmation prompt
			_ = cmd.Flags().Set("yes", "true")

			err := metafieldDefinitionsDeleteCmd.RunE(cmd, []string{tt.defID})

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

func TestMetafieldDefinitionsDeleteRunENoConfirm(t *testing.T) {
	mockClient := &metafieldDefinitionsMockAPIClient{}
	cleanup, _ := setupMetafieldDefinitionsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldDefinitionsTestCmd()
	// yes is false by default in this test cmd, but we set it explicitly for clarity
	_ = cmd.Flags().Set("yes", "false")

	// When yes=false and no input provided, the confirmation will fail (empty confirm)
	// This tests the confirmation path - in a real terminal it would wait for input
	// but in tests, Scanln returns immediately with empty string
	err := metafieldDefinitionsDeleteCmd.RunE(cmd, []string{"def_123"})
	// No error expected since user "cancelled" (empty confirm != "y")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMetafieldDefinitionsCommandUseAndShort(t *testing.T) {
	if metafieldDefinitionsCmd.Use != "metafield-definitions" {
		t.Errorf("expected Use 'metafield-definitions', got %q", metafieldDefinitionsCmd.Use)
	}
	if metafieldDefinitionsCmd.Short != "Manage metafield definitions" {
		t.Errorf("expected Short 'Manage metafield definitions', got %q", metafieldDefinitionsCmd.Short)
	}
}

func TestMetafieldDefinitionsListCmdUseAndShort(t *testing.T) {
	if metafieldDefinitionsListCmd.Use != "list" {
		t.Errorf("expected Use 'list', got %q", metafieldDefinitionsListCmd.Use)
	}
	if metafieldDefinitionsListCmd.Short != "List metafield definitions" {
		t.Errorf("expected Short 'List metafield definitions', got %q", metafieldDefinitionsListCmd.Short)
	}
}

func TestMetafieldDefinitionsGetCmdUseAndShort(t *testing.T) {
	if metafieldDefinitionsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", metafieldDefinitionsGetCmd.Use)
	}
	if metafieldDefinitionsGetCmd.Short != "Get metafield definition details" {
		t.Errorf("expected Short 'Get metafield definition details', got %q", metafieldDefinitionsGetCmd.Short)
	}
}

func TestMetafieldDefinitionsCreateCmdUseAndShort(t *testing.T) {
	if metafieldDefinitionsCreateCmd.Use != "create" {
		t.Errorf("expected Use 'create', got %q", metafieldDefinitionsCreateCmd.Use)
	}
	if metafieldDefinitionsCreateCmd.Short != "Create a metafield definition" {
		t.Errorf("expected Short 'Create a metafield definition', got %q", metafieldDefinitionsCreateCmd.Short)
	}
}

func TestMetafieldDefinitionsDeleteCmdUseAndShort(t *testing.T) {
	if metafieldDefinitionsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", metafieldDefinitionsDeleteCmd.Use)
	}
	if metafieldDefinitionsDeleteCmd.Short != "Delete a metafield definition" {
		t.Errorf("expected Short 'Delete a metafield definition', got %q", metafieldDefinitionsDeleteCmd.Short)
	}
}
