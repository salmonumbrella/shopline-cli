package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func TestCustomFieldsCommandStructure(t *testing.T) {
	subcommands := customFieldsCmd.Commands()

	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
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

func TestCustomFieldsAliases(t *testing.T) {
	aliases := customFieldsCmd.Aliases
	expectedAliases := map[string]bool{
		"custom-field": false,
		"cf":           false,
	}

	for _, alias := range aliases {
		if _, exists := expectedAliases[alias]; exists {
			expectedAliases[alias] = true
		}
	}

	for alias, found := range expectedAliases {
		if !found {
			t.Errorf("Expected alias %q not found", alias)
		}
	}
}

func TestCustomFieldsListFlags(t *testing.T) {
	flags := []string{"page", "page-size", "owner-type", "type"}

	for _, flagName := range flags {
		if customFieldsListCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on list command", flagName)
		}
	}
}

func TestCustomFieldsGetArgs(t *testing.T) {
	err := customFieldsGetCmd.Args(customFieldsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = customFieldsGetCmd.Args(customFieldsGetCmd, []string{"cf-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestCustomFieldsUpdateArgs(t *testing.T) {
	err := customFieldsUpdateCmd.Args(customFieldsUpdateCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = customFieldsUpdateCmd.Args(customFieldsUpdateCmd, []string{"cf-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestCustomFieldsDeleteArgs(t *testing.T) {
	err := customFieldsDeleteCmd.Args(customFieldsDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = customFieldsDeleteCmd.Args(customFieldsDeleteCmd, []string{"cf-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestCustomFieldsCreateFlags(t *testing.T) {
	flags := []string{"name", "key", "description", "type", "owner-type", "required", "searchable", "visible", "default-value", "options", "validation", "position"}

	for _, flagName := range flags {
		if customFieldsCreateCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on create command", flagName)
		}
	}
}

func TestCustomFieldsUpdateFlags(t *testing.T) {
	flags := []string{"name", "description", "required", "searchable", "visible", "default-value", "options", "validation", "position"}

	for _, flagName := range flags {
		if customFieldsUpdateCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on update command", flagName)
		}
	}
}

// customFieldsMockAPIClient is a mock implementation of api.APIClient for custom fields tests.
type customFieldsMockAPIClient struct {
	api.MockClient
	listCustomFieldsResp  *api.CustomFieldsListResponse
	listCustomFieldsErr   error
	getCustomFieldResp    *api.CustomField
	getCustomFieldErr     error
	createCustomFieldResp *api.CustomField
	createCustomFieldErr  error
	updateCustomFieldResp *api.CustomField
	updateCustomFieldErr  error
	deleteCustomFieldErr  error
}

func (m *customFieldsMockAPIClient) ListCustomFields(ctx context.Context, opts *api.CustomFieldsListOptions) (*api.CustomFieldsListResponse, error) {
	return m.listCustomFieldsResp, m.listCustomFieldsErr
}

func (m *customFieldsMockAPIClient) GetCustomField(ctx context.Context, id string) (*api.CustomField, error) {
	return m.getCustomFieldResp, m.getCustomFieldErr
}

func (m *customFieldsMockAPIClient) CreateCustomField(ctx context.Context, req *api.CustomFieldCreateRequest) (*api.CustomField, error) {
	return m.createCustomFieldResp, m.createCustomFieldErr
}

func (m *customFieldsMockAPIClient) UpdateCustomField(ctx context.Context, id string, req *api.CustomFieldUpdateRequest) (*api.CustomField, error) {
	return m.updateCustomFieldResp, m.updateCustomFieldErr
}

func (m *customFieldsMockAPIClient) DeleteCustomField(ctx context.Context, id string) error {
	return m.deleteCustomFieldErr
}

// setupCustomFieldsMockFactories sets up mock factories for custom fields tests.
func setupCustomFieldsMockFactories(mockClient *customFieldsMockAPIClient) (func(), *bytes.Buffer) {
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

// newCustomFieldsTestCmd creates a test command with common flags for custom fields tests.
func newCustomFieldsTestCmd() *cobra.Command {
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

// TestCustomFieldsListRunE tests the custom fields list command with mock API.
func TestCustomFieldsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CustomFieldsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CustomFieldsListResponse{
				Items: []api.CustomField{
					{
						ID:   "cf_123",
						Type: api.CustomFieldTypeText,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "cf_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CustomFieldsListResponse{
				Items:      []api.CustomField{},
				TotalCount: 0,
			},
		},
		{
			name: "list with multiple fields",
			mockResp: &api.CustomFieldsListResponse{
				Items: []api.CustomField{
					{
						ID:   "cf_001",
						Type: api.CustomFieldTypeText,
					},
					{
						ID:   "cf_002",
						Type: api.CustomFieldTypeNumber,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "cf_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customFieldsMockAPIClient{
				listCustomFieldsResp: tt.mockResp,
				listCustomFieldsErr:  tt.mockErr,
			}
			cleanup, buf := setupCustomFieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomFieldsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("owner-type", "", "")
			cmd.Flags().String("type", "", "")

			err := customFieldsListCmd.RunE(cmd, []string{})

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

// TestCustomFieldsListRunEJSON tests the custom fields list command with JSON output.
func TestCustomFieldsListRunEJSON(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{
		listCustomFieldsResp: &api.CustomFieldsListResponse{
			Items: []api.CustomField{
				{
					ID:   "cf_json_123",
					Type: api.CustomFieldTypeSelect,
					Options: map[string]interface{}{
						"values": []string{"opt1", "opt2"},
					},
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("owner-type", "", "")
	cmd.Flags().String("type", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := customFieldsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cf_json_123") {
		t.Errorf("JSON output should contain field ID, got: %q", output)
	}
}

// TestCustomFieldsGetRunE tests the custom fields get command with mock API.
func TestCustomFieldsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		fieldID  string
		mockResp *api.CustomField
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get",
			fieldID: "cf_123",
			mockResp: &api.CustomField{
				ID:   "cf_123",
				Type: api.CustomFieldTypeText,
			},
		},
		{
			name:    "field not found",
			fieldID: "cf_999",
			mockErr: errors.New("custom field not found"),
			wantErr: true,
		},
		{
			name:    "get with options",
			fieldID: "cf_select",
			mockResp: &api.CustomField{
				ID:   "cf_select",
				Type: api.CustomFieldTypeSelect,
				Options: map[string]interface{}{
					"values": []string{"option1", "option2", "option3"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customFieldsMockAPIClient{
				getCustomFieldResp: tt.mockResp,
				getCustomFieldErr:  tt.mockErr,
			}
			cleanup, _ := setupCustomFieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomFieldsTestCmd()

			err := customFieldsGetCmd.RunE(cmd, []string{tt.fieldID})

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

// TestCustomFieldsGetRunEJSON tests the custom fields get command with JSON output.
func TestCustomFieldsGetRunEJSON(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{
		getCustomFieldResp: &api.CustomField{
			ID:   "cf_json_get",
			Type: api.CustomFieldTypeBoolean,
		},
	}
	cleanup, buf := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := customFieldsGetCmd.RunE(cmd, []string{"cf_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cf_json_get") {
		t.Errorf("JSON output should contain field ID, got: %q", output)
	}
}

// TestCustomFieldsCreateRunE tests the custom fields create command with mock API.
func TestCustomFieldsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.CustomField
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.CustomField{
				ID:   "cf_new",
				Type: api.CustomFieldTypeText,
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("failed to create custom field"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customFieldsMockAPIClient{
				createCustomFieldResp: tt.mockResp,
				createCustomFieldErr:  tt.mockErr,
			}
			cleanup, _ := setupCustomFieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomFieldsTestCmd()
			cmd.Flags().String("name", "Test Field", "")
			cmd.Flags().String("key", "test_field", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("type", "text", "")
			cmd.Flags().String("owner-type", "product", "")
			cmd.Flags().Bool("required", false, "")
			cmd.Flags().Bool("searchable", false, "")
			cmd.Flags().Bool("visible", true, "")
			cmd.Flags().String("default-value", "", "")
			cmd.Flags().String("options", "", "")
			cmd.Flags().String("validation", "", "")
			cmd.Flags().Int("position", 0, "")

			err := customFieldsCreateCmd.RunE(cmd, []string{})

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

// TestCustomFieldsCreateDryRun tests the custom fields create command in dry-run mode.
func TestCustomFieldsCreateDryRun(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{}
	cleanup, _ := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	cmd.Flags().String("name", "Dry Run Field", "")
	cmd.Flags().String("key", "dry_run", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("type", "text", "")
	cmd.Flags().String("owner-type", "product", "")
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("searchable", false, "")
	cmd.Flags().Bool("visible", true, "")
	cmd.Flags().String("default-value", "", "")
	cmd.Flags().String("options", "", "")
	cmd.Flags().String("validation", "", "")
	cmd.Flags().Int("position", 0, "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := customFieldsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run: %v", err)
	}
}

// TestCustomFieldsCreateWithOptions tests creating a field with comma-separated options.
func TestCustomFieldsCreateWithOptions(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{
		createCustomFieldResp: &api.CustomField{
			ID:   "cf_with_opts",
			Type: api.CustomFieldTypeSelect,
			Options: map[string]interface{}{
				"values": []string{"red", "green", "blue"},
			},
		},
	}
	cleanup, _ := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	cmd.Flags().String("name", "Select Field", "")
	cmd.Flags().String("key", "select_field", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("type", "select", "")
	cmd.Flags().String("owner-type", "product", "")
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("searchable", false, "")
	cmd.Flags().Bool("visible", true, "")
	cmd.Flags().String("default-value", "", "")
	cmd.Flags().String("options", "red, green, blue", "")
	cmd.Flags().String("validation", "", "")
	cmd.Flags().Int("position", 0, "")

	err := customFieldsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestCustomFieldsCreateRunEJSON tests the custom fields create command with JSON output.
func TestCustomFieldsCreateRunEJSON(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{
		createCustomFieldResp: &api.CustomField{
			ID:   "cf_json_create",
			Type: api.CustomFieldTypeNumber,
		},
	}
	cleanup, buf := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	cmd.Flags().String("name", "JSON Create Field", "")
	cmd.Flags().String("key", "json_create", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("type", "number", "")
	cmd.Flags().String("owner-type", "variant", "")
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("searchable", false, "")
	cmd.Flags().Bool("visible", true, "")
	cmd.Flags().String("default-value", "", "")
	cmd.Flags().String("options", "", "")
	cmd.Flags().String("validation", "", "")
	cmd.Flags().Int("position", 0, "")
	_ = cmd.Flags().Set("output", "json")

	err := customFieldsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cf_json_create") {
		t.Errorf("JSON output should contain field ID, got: %q", output)
	}
}

// TestCustomFieldsUpdateRunE tests the custom fields update command with mock API.
func TestCustomFieldsUpdateRunE(t *testing.T) {
	tests := []struct {
		name     string
		fieldID  string
		mockResp *api.CustomField
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful update",
			fieldID: "cf_123",
			mockResp: &api.CustomField{
				ID:   "cf_123",
				Type: api.CustomFieldTypeText,
			},
		},
		{
			name:    "update fails",
			fieldID: "cf_999",
			mockErr: errors.New("custom field not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customFieldsMockAPIClient{
				updateCustomFieldResp: tt.mockResp,
				updateCustomFieldErr:  tt.mockErr,
			}
			cleanup, _ := setupCustomFieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomFieldsTestCmd()
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().Bool("required", false, "")
			cmd.Flags().Bool("searchable", false, "")
			cmd.Flags().Bool("visible", false, "")
			cmd.Flags().String("default-value", "", "")
			cmd.Flags().String("options", "", "")
			cmd.Flags().String("validation", "", "")
			cmd.Flags().Int("position", 0, "")

			err := customFieldsUpdateCmd.RunE(cmd, []string{tt.fieldID})

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

// TestCustomFieldsUpdateDryRun tests the custom fields update command in dry-run mode.
func TestCustomFieldsUpdateDryRun(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{}
	cleanup, _ := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	cmd.Flags().String("name", "Updated Name", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("searchable", false, "")
	cmd.Flags().Bool("visible", false, "")
	cmd.Flags().String("default-value", "", "")
	cmd.Flags().String("options", "", "")
	cmd.Flags().String("validation", "", "")
	cmd.Flags().Int("position", 0, "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := customFieldsUpdateCmd.RunE(cmd, []string{"cf_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run: %v", err)
	}
}

// TestCustomFieldsUpdateRunEJSON tests the custom fields update command with JSON output.
func TestCustomFieldsUpdateRunEJSON(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{
		updateCustomFieldResp: &api.CustomField{
			ID:   "cf_json_update",
			Type: api.CustomFieldTypeText,
		},
	}
	cleanup, buf := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	cmd.Flags().String("name", "JSON Updated Field", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("searchable", false, "")
	cmd.Flags().Bool("visible", false, "")
	cmd.Flags().String("default-value", "", "")
	cmd.Flags().String("options", "", "")
	cmd.Flags().String("validation", "", "")
	cmd.Flags().Int("position", 0, "")
	_ = cmd.Flags().Set("output", "json")

	err := customFieldsUpdateCmd.RunE(cmd, []string{"cf_json_update"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "cf_json_update") {
		t.Errorf("JSON output should contain field ID, got: %q", output)
	}
}

// TestCustomFieldsDeleteRunE tests the custom fields delete command with mock API.
func TestCustomFieldsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		fieldID string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			fieldID: "cf_123",
			mockErr: nil,
		},
		{
			name:    "delete fails",
			fieldID: "cf_456",
			mockErr: errors.New("custom field not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customFieldsMockAPIClient{
				deleteCustomFieldErr: tt.mockErr,
			}
			cleanup, _ := setupCustomFieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomFieldsTestCmd()

			err := customFieldsDeleteCmd.RunE(cmd, []string{tt.fieldID})

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

// TestCustomFieldsDeleteDryRun tests the custom fields delete command in dry-run mode.
func TestCustomFieldsDeleteDryRun(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{}
	cleanup, _ := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("yes", "false") // dry-run should work without confirmation

	err := customFieldsDeleteCmd.RunE(cmd, []string{"cf_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run: %v", err)
	}
}

// TestCustomFieldsDeleteNoConfirmation tests that delete exits early without --yes flag.
func TestCustomFieldsDeleteNoConfirmation(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{}
	cleanup, _ := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	_ = cmd.Flags().Set("yes", "false")

	// Without --yes flag, delete should exit early without error
	err := customFieldsDeleteCmd.RunE(cmd, []string{"cf-id"})
	if err != nil {
		t.Errorf("Expected nil when confirmation not provided, got: %v", err)
	}
}

// TestCustomFieldsListGetClientError verifies error handling when getClient fails on list.
func TestCustomFieldsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("owner-type", "", "")
	cmd.Flags().String("type", "", "")

	err := customFieldsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestCustomFieldsGetGetClientError verifies error handling when getClient fails on get.
func TestCustomFieldsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := customFieldsGetCmd.RunE(cmd, []string{"cf-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestCustomFieldsCreateGetClientError verifies error handling when getClient fails on create.
func TestCustomFieldsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("key", "test", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("type", "text", "")
	cmd.Flags().String("owner-type", "product", "")
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("searchable", false, "")
	cmd.Flags().Bool("visible", true, "")
	cmd.Flags().String("default-value", "", "")
	cmd.Flags().String("options", "", "")
	cmd.Flags().String("validation", "", "")
	cmd.Flags().Int("position", 0, "")

	err := customFieldsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestCustomFieldsUpdateGetClientError verifies error handling when getClient fails on update.
func TestCustomFieldsUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("searchable", false, "")
	cmd.Flags().Bool("visible", false, "")
	cmd.Flags().String("default-value", "", "")
	cmd.Flags().String("options", "", "")
	cmd.Flags().String("validation", "", "")
	cmd.Flags().Int("position", 0, "")

	err := customFieldsUpdateCmd.RunE(cmd, []string{"cf-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestCustomFieldsDeleteGetClientError verifies error handling when getClient fails on delete.
func TestCustomFieldsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	// Set --yes to true so the command proceeds to getClient instead of exiting early
	_ = cmd.Flags().Set("yes", "true")

	err := customFieldsDeleteCmd.RunE(cmd, []string{"cf-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestCustomFieldsListFlagsDefaults tests list command flag default values.
func TestCustomFieldsListFlagsDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"owner-type", ""},
		{"type", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := customFieldsListCmd.Flags().Lookup(f.name)
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

// TestCustomFieldsCreateFlagsDefaults tests create command flag default values.
func TestCustomFieldsCreateFlagsDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"key", ""},
		{"description", ""},
		{"type", "text"},
		{"owner-type", "product"},
		{"required", "false"},
		{"searchable", "false"},
		{"visible", "true"},
		{"default-value", ""},
		{"options", ""},
		{"validation", ""},
		{"position", "0"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := customFieldsCreateCmd.Flags().Lookup(f.name)
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

// TestCustomFieldsDeleteFlagsDefaults tests delete command flag default values.
func TestCustomFieldsDeleteFlagsDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"yes", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := customFieldsDeleteCmd.Flags().Lookup(f.name)
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

// TestCustomFieldsCommandSetup verifies custom-fields command initialization.
func TestCustomFieldsCommandSetup(t *testing.T) {
	if customFieldsCmd.Use != "custom-fields" {
		t.Errorf("expected Use 'custom-fields', got %q", customFieldsCmd.Use)
	}
	if customFieldsCmd.Short != "Manage custom field definitions" {
		t.Errorf("expected Short 'Manage custom field definitions', got %q", customFieldsCmd.Short)
	}
}

// TestCustomFieldsSubcommands verifies all subcommands are registered.
func TestCustomFieldsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List custom fields",
		"get":    "Get custom field details",
		"create": "Create a custom field",
		"update": "Update a custom field",
		"delete": "Delete a custom field",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range customFieldsCmd.Commands() {
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

// TestCustomFieldsUpdateWithChangedFlags tests update command with various changed flags.
func TestCustomFieldsUpdateWithChangedFlags(t *testing.T) {
	mockClient := &customFieldsMockAPIClient{
		updateCustomFieldResp: &api.CustomField{
			ID:   "cf_update_flags",
			Type: api.CustomFieldTypeText,
		},
	}
	cleanup, _ := setupCustomFieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomFieldsTestCmd()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("searchable", false, "")
	cmd.Flags().Bool("visible", false, "")
	cmd.Flags().String("default-value", "", "")
	cmd.Flags().String("options", "", "")
	cmd.Flags().String("validation", "", "")
	cmd.Flags().Int("position", 0, "")

	// Simulate changed flags
	_ = cmd.Flags().Set("name", "Updated Name")
	_ = cmd.Flags().Set("visible", "false")
	_ = cmd.Flags().Set("required", "true")
	_ = cmd.Flags().Set("position", "5")
	_ = cmd.Flags().Set("options", "opt1, opt2")

	err := customFieldsUpdateCmd.RunE(cmd, []string{"cf_update_flags"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
