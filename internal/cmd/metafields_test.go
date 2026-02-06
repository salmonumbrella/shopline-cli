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

// metafieldsMockAPIClient is a mock implementation of api.APIClient for metafields tests.
type metafieldsMockAPIClient struct {
	api.MockClient
	listMetafieldsResp  *api.MetafieldsListResponse
	listMetafieldsErr   error
	getMetafieldResp    *api.Metafield
	getMetafieldErr     error
	createMetafieldResp *api.Metafield
	createMetafieldErr  error
	deleteMetafieldErr  error
}

func (m *metafieldsMockAPIClient) ListMetafields(ctx context.Context, opts *api.MetafieldsListOptions) (*api.MetafieldsListResponse, error) {
	return m.listMetafieldsResp, m.listMetafieldsErr
}

func (m *metafieldsMockAPIClient) GetMetafield(ctx context.Context, id string) (*api.Metafield, error) {
	return m.getMetafieldResp, m.getMetafieldErr
}

func (m *metafieldsMockAPIClient) CreateMetafield(ctx context.Context, req *api.MetafieldCreateRequest) (*api.Metafield, error) {
	return m.createMetafieldResp, m.createMetafieldErr
}

func (m *metafieldsMockAPIClient) DeleteMetafield(ctx context.Context, id string) error {
	return m.deleteMetafieldErr
}

// setupMetafieldsMockFactories sets up mock factories for metafields tests.
func setupMetafieldsMockFactories(mockClient *metafieldsMockAPIClient) (func(), *bytes.Buffer) {
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

// newMetafieldsTestCmd creates a test command with common flags for metafields tests.
func newMetafieldsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

func TestMetafieldsCommandStructure(t *testing.T) {
	subcommands := metafieldsCmd.Commands()

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

func TestMetafieldsListFlags(t *testing.T) {
	flags := []string{"namespace", "key", "owner-type", "owner-id", "page", "page-size"}

	for _, flagName := range flags {
		if metafieldsListCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on list command", flagName)
		}
	}
}

func TestMetafieldsGetArgs(t *testing.T) {
	err := metafieldsGetCmd.Args(metafieldsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = metafieldsGetCmd.Args(metafieldsGetCmd, []string{"mf-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestMetafieldsDeleteArgs(t *testing.T) {
	err := metafieldsDeleteCmd.Args(metafieldsDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = metafieldsDeleteCmd.Args(metafieldsDeleteCmd, []string{"mf-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestMetafieldsCreateFlags(t *testing.T) {
	flags := []string{"namespace", "key", "value", "type", "description", "owner-type", "owner-id"}

	for _, flagName := range flags {
		if metafieldsCreateCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on create command", flagName)
		}
	}
}

func TestMetafieldsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := metafieldsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestMetafieldsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := metafieldsGetCmd.RunE(cmd, []string{"mf-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestMetafieldsListRunE tests the metafields list command with mock API.
func TestMetafieldsListRunE(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name       string
		mockResp   *api.MetafieldsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.MetafieldsListResponse{
				Items: []api.Metafield{
					{
						ID:        "mf_123",
						Namespace: "custom",
						Key:       "test_key",
						Value:     "test_value",
						ValueType: "string",
						OwnerType: "product",
						OwnerID:   "prod_456",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "mf_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.MetafieldsListResponse{
				Items:      []api.Metafield{},
				TotalCount: 0,
			},
		},
		{
			name: "list with long value truncation",
			mockResp: &api.MetafieldsListResponse{
				Items: []api.Metafield{
					{
						ID:        "mf_456",
						Namespace: "custom",
						Key:       "long_key",
						Value:     "This is a very long value that should be truncated",
						ValueType: "string",
						OwnerType: "product",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "mf_456",
		},
		{
			name: "list with owner type only",
			mockResp: &api.MetafieldsListResponse{
				Items: []api.Metafield{
					{
						ID:        "mf_789",
						Namespace: "app",
						Key:       "setting",
						Value:     "value",
						ValueType: "string",
						OwnerType: "shop",
						OwnerID:   "",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "shop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &metafieldsMockAPIClient{
				listMetafieldsResp: tt.mockResp,
				listMetafieldsErr:  tt.mockErr,
			}
			cleanup, buf := setupMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newMetafieldsTestCmd()
			cmd.Flags().String("namespace", "", "")
			cmd.Flags().String("key", "", "")
			cmd.Flags().String("owner-type", "", "")
			cmd.Flags().String("owner-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := metafieldsListCmd.RunE(cmd, []string{})

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

// TestMetafieldsListRunEWithFilters tests list command with various filter flags.
func TestMetafieldsListRunEWithFilters(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &metafieldsMockAPIClient{
		listMetafieldsResp: &api.MetafieldsListResponse{
			Items: []api.Metafield{
				{
					ID:        "mf_filtered",
					Namespace: "custom",
					Key:       "color",
					Value:     "red",
					ValueType: "string",
					OwnerType: "product",
					OwnerID:   "prod_123",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupMetafieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldsTestCmd()
	cmd.Flags().String("namespace", "custom", "")
	cmd.Flags().String("key", "color", "")
	cmd.Flags().String("owner-type", "product", "")
	cmd.Flags().String("owner-id", "prod_123", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")

	err := metafieldsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "mf_filtered") {
		t.Errorf("output should contain filtered metafield ID")
	}
}

// TestMetafieldsListRunEJSONOutput tests list command with JSON output.
func TestMetafieldsListRunEJSONOutput(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &metafieldsMockAPIClient{
		listMetafieldsResp: &api.MetafieldsListResponse{
			Items: []api.Metafield{
				{
					ID:        "mf_json",
					Namespace: "custom",
					Key:       "test",
					Value:     "value",
					ValueType: "string",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupMetafieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldsTestCmd()
	cmd.Flags().String("namespace", "", "")
	cmd.Flags().String("key", "", "")
	cmd.Flags().String("owner-type", "", "")
	cmd.Flags().String("owner-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := metafieldsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "mf_json") {
		t.Errorf("JSON output should contain metafield ID, got: %s", output)
	}
}

// TestMetafieldsGetRunE tests the metafields get command with mock API.
func TestMetafieldsGetRunE(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name        string
		metafieldID string
		mockResp    *api.Metafield
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful get",
			metafieldID: "mf_123",
			mockResp: &api.Metafield{
				ID:          "mf_123",
				Namespace:   "custom",
				Key:         "test_key",
				Value:       "test_value",
				ValueType:   "string",
				Description: "A test metafield",
				OwnerType:   "product",
				OwnerID:     "prod_456",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			name:        "get with no owner",
			metafieldID: "mf_shop",
			mockResp: &api.Metafield{
				ID:          "mf_shop",
				Namespace:   "app",
				Key:         "setting",
				Value:       "enabled",
				ValueType:   "boolean",
				Description: "App setting",
				OwnerType:   "",
				OwnerID:     "",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			name:        "metafield not found",
			metafieldID: "mf_999",
			mockErr:     errors.New("metafield not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &metafieldsMockAPIClient{
				getMetafieldResp: tt.mockResp,
				getMetafieldErr:  tt.mockErr,
			}
			cleanup, _ := setupMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newMetafieldsTestCmd()

			err := metafieldsGetCmd.RunE(cmd, []string{tt.metafieldID})

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

// TestMetafieldsGetRunEJSONOutput tests get command with JSON output.
func TestMetafieldsGetRunEJSONOutput(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &metafieldsMockAPIClient{
		getMetafieldResp: &api.Metafield{
			ID:          "mf_json",
			Namespace:   "custom",
			Key:         "test",
			Value:       "value",
			ValueType:   "string",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	cleanup, buf := setupMetafieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := metafieldsGetCmd.RunE(cmd, []string{"mf_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "mf_json") {
		t.Errorf("JSON output should contain metafield ID, got: %s", output)
	}
}

// TestMetafieldsCreateRunE tests the metafields create command with mock API.
func TestMetafieldsCreateRunE(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name     string
		mockResp *api.Metafield
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Metafield{
				ID:        "mf_new",
				Namespace: "custom",
				Key:       "new_key",
				Value:     "new_value",
				ValueType: "string",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:    "create API error",
			mockErr: errors.New("validation failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &metafieldsMockAPIClient{
				createMetafieldResp: tt.mockResp,
				createMetafieldErr:  tt.mockErr,
			}
			cleanup, _ := setupMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newMetafieldsTestCmd()
			cmd.Flags().String("namespace", "custom", "")
			cmd.Flags().String("key", "new_key", "")
			cmd.Flags().String("value", "new_value", "")
			cmd.Flags().String("type", "string", "")
			cmd.Flags().String("description", "A new metafield", "")
			cmd.Flags().String("owner-type", "product", "")
			cmd.Flags().String("owner-id", "prod_123", "")

			err := metafieldsCreateCmd.RunE(cmd, []string{})

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

// TestMetafieldsCreateRunEJSONOutput tests create command with JSON output.
func TestMetafieldsCreateRunEJSONOutput(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &metafieldsMockAPIClient{
		createMetafieldResp: &api.Metafield{
			ID:        "mf_created",
			Namespace: "custom",
			Key:       "test",
			Value:     "value",
			ValueType: "string",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	cleanup, buf := setupMetafieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldsTestCmd()
	cmd.Flags().String("namespace", "custom", "")
	cmd.Flags().String("key", "test", "")
	cmd.Flags().String("value", "value", "")
	cmd.Flags().String("type", "string", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("owner-type", "", "")
	cmd.Flags().String("owner-id", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := metafieldsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "mf_created") {
		t.Errorf("JSON output should contain metafield ID, got: %s", output)
	}
}

// TestMetafieldsCreateGetClientError tests create command when getClient fails.
func TestMetafieldsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("namespace", "custom", "")
	cmd.Flags().String("key", "test", "")
	cmd.Flags().String("value", "value", "")
	cmd.Flags().String("type", "string", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("owner-type", "", "")
	cmd.Flags().String("owner-id", "", "")

	err := metafieldsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestMetafieldsDeleteRunE tests the metafields delete command with mock API.
func TestMetafieldsDeleteRunE(t *testing.T) {
	tests := []struct {
		name        string
		metafieldID string
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful delete",
			metafieldID: "mf_123",
			mockErr:     nil,
		},
		{
			name:        "delete API error",
			metafieldID: "mf_999",
			mockErr:     errors.New("metafield not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &metafieldsMockAPIClient{
				deleteMetafieldErr: tt.mockErr,
			}
			cleanup, _ := setupMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newMetafieldsTestCmd()

			err := metafieldsDeleteCmd.RunE(cmd, []string{tt.metafieldID})

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

// TestMetafieldsDeleteGetClientError tests delete command when getClient fails.
func TestMetafieldsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := metafieldsDeleteCmd.RunE(cmd, []string{"mf_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestMetafieldsDeleteWithYesFlag tests delete command with yes flag set.
func TestMetafieldsDeleteWithYesFlag(t *testing.T) {
	mockClient := &metafieldsMockAPIClient{
		deleteMetafieldErr: nil,
	}
	cleanup, _ := setupMetafieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldsTestCmd()
	_ = cmd.Flags().Set("yes", "true")

	err := metafieldsDeleteCmd.RunE(cmd, []string{"mf_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMetafieldsCommandUse tests command Use fields.
func TestMetafieldsCommandUse(t *testing.T) {
	if metafieldsCmd.Use != "metafields" {
		t.Errorf("expected Use 'metafields', got %q", metafieldsCmd.Use)
	}
	if metafieldsCmd.Short != "Manage metafields" {
		t.Errorf("expected Short 'Manage metafields', got %q", metafieldsCmd.Short)
	}
}

// TestMetafieldsListCommandUse tests list command Use field.
func TestMetafieldsListCommandUse(t *testing.T) {
	if metafieldsListCmd.Use != "list" {
		t.Errorf("expected Use 'list', got %q", metafieldsListCmd.Use)
	}
	if metafieldsListCmd.Short != "List metafields" {
		t.Errorf("expected Short 'List metafields', got %q", metafieldsListCmd.Short)
	}
}

// TestMetafieldsGetCommandUse tests get command Use field.
func TestMetafieldsGetCommandUse(t *testing.T) {
	if metafieldsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", metafieldsGetCmd.Use)
	}
	if metafieldsGetCmd.Short != "Get metafield details" {
		t.Errorf("expected Short 'Get metafield details', got %q", metafieldsGetCmd.Short)
	}
}

// TestMetafieldsCreateCommandUse tests create command Use field.
func TestMetafieldsCreateCommandUse(t *testing.T) {
	if metafieldsCreateCmd.Use != "create" {
		t.Errorf("expected Use 'create', got %q", metafieldsCreateCmd.Use)
	}
	if metafieldsCreateCmd.Short != "Create a metafield" {
		t.Errorf("expected Short 'Create a metafield', got %q", metafieldsCreateCmd.Short)
	}
}

// TestMetafieldsDeleteCommandUse tests delete command Use field.
func TestMetafieldsDeleteCommandUse(t *testing.T) {
	if metafieldsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", metafieldsDeleteCmd.Use)
	}
	if metafieldsDeleteCmd.Short != "Delete a metafield" {
		t.Errorf("expected Short 'Delete a metafield', got %q", metafieldsDeleteCmd.Short)
	}
}

// TestMetafieldsGetWithOwnerType tests get command output with owner type and ID.
func TestMetafieldsGetWithOwnerType(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &metafieldsMockAPIClient{
		getMetafieldResp: &api.Metafield{
			ID:          "mf_with_owner",
			Namespace:   "custom",
			Key:         "test",
			Value:       "value",
			ValueType:   "string",
			Description: "Test description",
			OwnerType:   "product",
			OwnerID:     "prod_123",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	cleanup, _ := setupMetafieldsMockFactories(mockClient)
	defer cleanup()

	cmd := newMetafieldsTestCmd()

	err := metafieldsGetCmd.RunE(cmd, []string{"mf_with_owner"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
