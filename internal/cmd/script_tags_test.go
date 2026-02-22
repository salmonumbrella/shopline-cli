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

func TestScriptTagsCommandSetup(t *testing.T) {
	if scriptTagsCmd.Use != "script-tags" {
		t.Errorf("expected Use 'script-tags', got %q", scriptTagsCmd.Use)
	}
	if scriptTagsCmd.Short != "Manage script tags" {
		t.Errorf("expected Short 'Manage script tags', got %q", scriptTagsCmd.Short)
	}
}

func TestScriptTagsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List script tags",
		"get":    "Get script tag details",
		"create": "Create a script tag",
		"delete": "Delete a script tag",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range scriptTagsCmd.Commands() {
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

func TestScriptTagsListGetClientError(t *testing.T) {
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

func TestScriptTagsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"src", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := scriptTagsListCmd.Flags().Lookup(f.name)
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

func TestScriptTagsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"src", ""},
		{"event", ""},
		{"display-scope", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := scriptTagsCreateCmd.Flags().Lookup(f.name)
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

func TestScriptTagsWithMockStore(t *testing.T) {
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

// TestScriptTagsGetArgs verifies get command requires exactly 1 argument
func TestScriptTagsGetArgs(t *testing.T) {
	err := scriptTagsGetCmd.Args(scriptTagsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = scriptTagsGetCmd.Args(scriptTagsGetCmd, []string{"tag-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestScriptTagsDeleteArgs verifies delete command requires exactly 1 argument
func TestScriptTagsDeleteArgs(t *testing.T) {
	err := scriptTagsDeleteCmd.Args(scriptTagsDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = scriptTagsDeleteCmd.Args(scriptTagsDeleteCmd, []string{"tag-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// scriptTagsMockAPIClient is a mock implementation of api.APIClient for script tags tests.
type scriptTagsMockAPIClient struct {
	api.MockClient
	listScriptTagsResp  *api.ScriptTagsListResponse
	listScriptTagsErr   error
	getScriptTagResp    *api.ScriptTag
	getScriptTagErr     error
	createScriptTagResp *api.ScriptTag
	createScriptTagErr  error
	deleteScriptTagErr  error
}

func (m *scriptTagsMockAPIClient) ListScriptTags(ctx context.Context, opts *api.ScriptTagsListOptions) (*api.ScriptTagsListResponse, error) {
	return m.listScriptTagsResp, m.listScriptTagsErr
}

func (m *scriptTagsMockAPIClient) GetScriptTag(ctx context.Context, id string) (*api.ScriptTag, error) {
	return m.getScriptTagResp, m.getScriptTagErr
}

func (m *scriptTagsMockAPIClient) CreateScriptTag(ctx context.Context, req *api.ScriptTagCreateRequest) (*api.ScriptTag, error) {
	return m.createScriptTagResp, m.createScriptTagErr
}

func (m *scriptTagsMockAPIClient) DeleteScriptTag(ctx context.Context, id string) error {
	return m.deleteScriptTagErr
}

// setupScriptTagsMockFactories sets up mock factories for script tags tests.
func setupScriptTagsMockFactories(mockClient *scriptTagsMockAPIClient) (func(), *bytes.Buffer) {
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

// newScriptTagsTestCmd creates a test command with common flags for script tags tests.
func newScriptTagsTestCmd() *cobra.Command {
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

// TestScriptTagsListRunE tests the script tags list command with mock API.
func TestScriptTagsListRunE(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.ScriptTagsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ScriptTagsListResponse{
				Items: []api.ScriptTag{
					{
						ID:           "tag_123",
						Src:          "https://example.com/script.js",
						Event:        "onload",
						DisplayScope: "all",
						CreatedAt:    testTime,
						UpdatedAt:    testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "tag_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ScriptTagsListResponse{
				Items:      []api.ScriptTag{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple script tags",
			mockResp: &api.ScriptTagsListResponse{
				Items: []api.ScriptTag{
					{
						ID:           "tag_001",
						Src:          "https://cdn.example.com/analytics.js",
						Event:        "onload",
						DisplayScope: "online_store",
						CreatedAt:    testTime,
					},
					{
						ID:           "tag_002",
						Src:          "https://cdn.example.com/tracking.js",
						Event:        "onload",
						DisplayScope: "all",
						CreatedAt:    testTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "tag_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &scriptTagsMockAPIClient{
				listScriptTagsResp: tt.mockResp,
				listScriptTagsErr:  tt.mockErr,
			}
			cleanup, buf := setupScriptTagsMockFactories(mockClient)
			defer cleanup()

			cmd := newScriptTagsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("src", "", "")

			err := scriptTagsListCmd.RunE(cmd, []string{})

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

// TestScriptTagsListRunEWithJSON tests JSON output format for list command.
func TestScriptTagsListRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &scriptTagsMockAPIClient{
		listScriptTagsResp: &api.ScriptTagsListResponse{
			Items: []api.ScriptTag{
				{
					ID:           "tag_json",
					Src:          "https://example.com/script.js",
					Event:        "onload",
					DisplayScope: "all",
					CreatedAt:    testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupScriptTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newScriptTagsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("src", "", "")

	err := scriptTagsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tag_json") {
		t.Errorf("JSON output should contain script tag ID, got: %s", output)
	}
}

// TestScriptTagsListRunEGetClientError verifies list command error handling when getClient fails
func TestScriptTagsListRunEGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newScriptTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("src", "", "")

	err := scriptTagsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestScriptTagsGetRunE tests the script tags get command with mock API.
func TestScriptTagsGetRunE(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		tagID    string
		mockResp *api.ScriptTag
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful get",
			tagID: "tag_123",
			mockResp: &api.ScriptTag{
				ID:           "tag_123",
				Src:          "https://example.com/script.js",
				Event:        "onload",
				DisplayScope: "all",
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
			},
		},
		{
			name:    "script tag not found",
			tagID:   "tag_999",
			mockErr: errors.New("script tag not found"),
			wantErr: true,
		},
		{
			name:  "script tag with different display scope",
			tagID: "tag_456",
			mockResp: &api.ScriptTag{
				ID:           "tag_456",
				Src:          "https://cdn.example.com/tracker.js",
				Event:        "onload",
				DisplayScope: "online_store",
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &scriptTagsMockAPIClient{
				getScriptTagResp: tt.mockResp,
				getScriptTagErr:  tt.mockErr,
			}
			cleanup, _ := setupScriptTagsMockFactories(mockClient)
			defer cleanup()

			cmd := newScriptTagsTestCmd()

			err := scriptTagsGetCmd.RunE(cmd, []string{tt.tagID})

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

// TestScriptTagsGetRunEWithJSON tests JSON output format for get command.
func TestScriptTagsGetRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &scriptTagsMockAPIClient{
		getScriptTagResp: &api.ScriptTag{
			ID:           "tag_json",
			Src:          "https://example.com/script.js",
			Event:        "onload",
			DisplayScope: "all",
			CreatedAt:    testTime,
			UpdatedAt:    testTime,
		},
	}
	cleanup, buf := setupScriptTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newScriptTagsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := scriptTagsGetCmd.RunE(cmd, []string{"tag_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tag_json") {
		t.Errorf("JSON output should contain script tag ID, got: %s", output)
	}
}

// TestScriptTagsGetRunEGetClientError verifies get command error handling when getClient fails
func TestScriptTagsGetRunEGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newScriptTagsTestCmd()

	err := scriptTagsGetCmd.RunE(cmd, []string{"tag-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestScriptTagsCreateRunE tests the script tags create command with mock API.
func TestScriptTagsCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		src      string
		event    string
		scope    string
		mockResp *api.ScriptTag
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful create",
			src:   "https://example.com/script.js",
			event: "onload",
			scope: "all",
			mockResp: &api.ScriptTag{
				ID:           "tag_new",
				Src:          "https://example.com/script.js",
				Event:        "onload",
				DisplayScope: "all",
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
			},
		},
		{
			name:    "API error",
			src:     "https://example.com/script.js",
			mockErr: errors.New("failed to create script tag"),
			wantErr: true,
		},
		{
			name:  "create with online_store scope",
			src:   "https://cdn.example.com/tracker.js",
			event: "onload",
			scope: "online_store",
			mockResp: &api.ScriptTag{
				ID:           "tag_store",
				Src:          "https://cdn.example.com/tracker.js",
				Event:        "onload",
				DisplayScope: "online_store",
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
			},
		},
		{
			name:  "create with minimal parameters",
			src:   "https://example.com/minimal.js",
			event: "",
			scope: "",
			mockResp: &api.ScriptTag{
				ID:           "tag_minimal",
				Src:          "https://example.com/minimal.js",
				Event:        "",
				DisplayScope: "",
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &scriptTagsMockAPIClient{
				createScriptTagResp: tt.mockResp,
				createScriptTagErr:  tt.mockErr,
			}
			cleanup, _ := setupScriptTagsMockFactories(mockClient)
			defer cleanup()

			cmd := newScriptTagsTestCmd()
			cmd.Flags().String("src", "", "")
			cmd.Flags().String("event", "", "")
			cmd.Flags().String("display-scope", "", "")
			_ = cmd.Flags().Set("src", tt.src)
			if tt.event != "" {
				_ = cmd.Flags().Set("event", tt.event)
			}
			if tt.scope != "" {
				_ = cmd.Flags().Set("display-scope", tt.scope)
			}

			err := scriptTagsCreateCmd.RunE(cmd, []string{})

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

// TestScriptTagsCreateRunEWithJSON tests JSON output format for create command.
func TestScriptTagsCreateRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &scriptTagsMockAPIClient{
		createScriptTagResp: &api.ScriptTag{
			ID:           "tag_json_create",
			Src:          "https://example.com/script.js",
			Event:        "onload",
			DisplayScope: "all",
			CreatedAt:    testTime,
			UpdatedAt:    testTime,
		},
	}
	cleanup, buf := setupScriptTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newScriptTagsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("src", "", "")
	cmd.Flags().String("event", "", "")
	cmd.Flags().String("display-scope", "", "")
	_ = cmd.Flags().Set("src", "https://example.com/script.js")

	err := scriptTagsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tag_json_create") {
		t.Errorf("JSON output should contain script tag ID, got: %s", output)
	}
}

// TestScriptTagsCreateRunEGetClientError verifies create command error handling when getClient fails
func TestScriptTagsCreateRunEGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newScriptTagsTestCmd()
	cmd.Flags().String("src", "", "")
	cmd.Flags().String("event", "", "")
	cmd.Flags().String("display-scope", "", "")
	_ = cmd.Flags().Set("src", "https://example.com/script.js")

	err := scriptTagsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestScriptTagsDeleteRunE tests the script tags delete command with mock API.
func TestScriptTagsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		tagID   string
		dryRun  bool
		yes     bool
		mockErr error
		wantErr bool
	}{
		{
			name:  "successful delete with yes flag",
			tagID: "tag_123",
			yes:   true,
		},
		{
			name:    "API error",
			tagID:   "tag_123",
			yes:     true,
			mockErr: errors.New("failed to delete"),
			wantErr: true,
		},
		{
			name:   "dry run mode",
			tagID:  "tag_123",
			dryRun: true,
			yes:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &scriptTagsMockAPIClient{
				deleteScriptTagErr: tt.mockErr,
			}
			cleanup, _ := setupScriptTagsMockFactories(mockClient)
			defer cleanup()

			cmd := newScriptTagsTestCmd()
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}
			if tt.yes {
				_ = cmd.Flags().Set("yes", "true")
			}

			err := scriptTagsDeleteCmd.RunE(cmd, []string{tt.tagID})

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

// TestScriptTagsDeleteRunEGetClientError verifies delete command error handling when getClient fails
func TestScriptTagsDeleteRunEGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newScriptTagsTestCmd()
	_ = cmd.Flags().Set("yes", "true")

	err := scriptTagsDeleteCmd.RunE(cmd, []string{"tag-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestScriptTagsListNoProfiles verifies list command error handling when no profiles exist
func TestScriptTagsListNoProfiles(t *testing.T) {
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

	cmd := newScriptTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("src", "", "")

	err := scriptTagsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestScriptTagsGetMultipleProfiles verifies get command error handling when multiple profiles exist
func TestScriptTagsGetMultipleProfiles(t *testing.T) {
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

	cmd := newScriptTagsTestCmd()

	err := scriptTagsGetCmd.RunE(cmd, []string{"tag-id"})
	if err == nil {
		t.Error("expected error for multiple profiles")
	}
}

// TestScriptTagsDeleteDryRunOutput verifies dry-run output for delete command
func TestScriptTagsDeleteDryRunOutput(t *testing.T) {
	mockClient := &scriptTagsMockAPIClient{}
	cleanup, _ := setupScriptTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newScriptTagsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := scriptTagsDeleteCmd.RunE(cmd, []string{"tag_dryrun"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestScriptTagsListWithSrcFilter tests list command with src filter
func TestScriptTagsListWithSrcFilter(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &scriptTagsMockAPIClient{
		listScriptTagsResp: &api.ScriptTagsListResponse{
			Items: []api.ScriptTag{
				{
					ID:           "tag_filtered",
					Src:          "https://specific.example.com/script.js",
					Event:        "onload",
					DisplayScope: "all",
					CreatedAt:    testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupScriptTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newScriptTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("src", "", "")
	_ = cmd.Flags().Set("src", "https://specific.example.com")

	err := scriptTagsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tag_filtered") {
		t.Errorf("output should contain filtered tag ID, got: %s", output)
	}
}

// TestScriptTagsListPagination tests list command with pagination options
func TestScriptTagsListPagination(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &scriptTagsMockAPIClient{
		listScriptTagsResp: &api.ScriptTagsListResponse{
			Items: []api.ScriptTag{
				{
					ID:           "tag_page2",
					Src:          "https://example.com/script.js",
					Event:        "onload",
					DisplayScope: "all",
					CreatedAt:    testTime,
				},
			},
			Page:       2,
			PageSize:   10,
			TotalCount: 15,
			HasMore:    false,
		},
	}
	cleanup, buf := setupScriptTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newScriptTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("src", "", "")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := scriptTagsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tag_page2") {
		t.Errorf("output should contain paginated tag ID, got: %s", output)
	}
}
