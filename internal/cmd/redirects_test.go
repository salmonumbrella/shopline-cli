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

func TestRedirectsCommandSetup(t *testing.T) {
	if redirectsCmd.Use != "redirects" {
		t.Errorf("expected Use 'redirects', got %q", redirectsCmd.Use)
	}
	if redirectsCmd.Short != "Manage URL redirects" {
		t.Errorf("expected Short 'Manage URL redirects', got %q", redirectsCmd.Short)
	}
}

func TestRedirectsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List redirects",
		"get":    "Get redirect details",
		"create": "Create a redirect",
		"delete": "Delete a redirect",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range redirectsCmd.Commands() {
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

func TestRedirectsListGetClientError(t *testing.T) {
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

func TestRedirectsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"path", ""},
		{"target", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := redirectsListCmd.Flags().Lookup(f.name)
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

func TestRedirectsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"path", ""},
		{"target", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := redirectsCreateCmd.Flags().Lookup(f.name)
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

func TestRedirectsDeleteFlags(t *testing.T) {
	flag := redirectsDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("flag 'yes' not found")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}

func TestRedirectsWithMockStore(t *testing.T) {
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

// redirectsMockAPIClient is a mock implementation of api.APIClient for redirects tests.
type redirectsMockAPIClient struct {
	api.MockClient
	listRedirectsResp  *api.RedirectsListResponse
	listRedirectsErr   error
	getRedirectResp    *api.Redirect
	getRedirectErr     error
	createRedirectResp *api.Redirect
	createRedirectErr  error
	deleteRedirectErr  error
}

func (m *redirectsMockAPIClient) ListRedirects(ctx context.Context, opts *api.RedirectsListOptions) (*api.RedirectsListResponse, error) {
	return m.listRedirectsResp, m.listRedirectsErr
}

func (m *redirectsMockAPIClient) GetRedirect(ctx context.Context, id string) (*api.Redirect, error) {
	return m.getRedirectResp, m.getRedirectErr
}

func (m *redirectsMockAPIClient) CreateRedirect(ctx context.Context, req *api.RedirectCreateRequest) (*api.Redirect, error) {
	return m.createRedirectResp, m.createRedirectErr
}

func (m *redirectsMockAPIClient) DeleteRedirect(ctx context.Context, id string) error {
	return m.deleteRedirectErr
}

// setupRedirectsMockFactories sets up mock factories for redirects tests.
func setupRedirectsMockFactories(mockClient *redirectsMockAPIClient) (func(), *bytes.Buffer) {
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

// newRedirectsTestCmd creates a test command with common flags for redirects tests.
func newRedirectsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

// TestRedirectsListRunE tests the redirects list command with mock API.
func TestRedirectsListRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.RedirectsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.RedirectsListResponse{
				Items: []api.Redirect{
					{
						ID:        "redir_123",
						Path:      "/old-page",
						Target:    "/new-page",
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "redir_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.RedirectsListResponse{
				Items:      []api.Redirect{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple redirects",
			mockResp: &api.RedirectsListResponse{
				Items: []api.Redirect{
					{
						ID:        "redir_1",
						Path:      "/old-path-1",
						Target:    "/new-path-1",
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					},
					{
						ID:        "redir_2",
						Path:      "/old-path-2",
						Target:    "/new-path-2",
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "redir_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &redirectsMockAPIClient{
				listRedirectsResp: tt.mockResp,
				listRedirectsErr:  tt.mockErr,
			}
			cleanup, buf := setupRedirectsMockFactories(mockClient)
			defer cleanup()

			cmd := newRedirectsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("path", "", "")
			cmd.Flags().String("target", "", "")

			err := redirectsListCmd.RunE(cmd, []string{})

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

// TestRedirectsListRunEJSON tests the redirects list command with JSON output.
func TestRedirectsListRunEJSON(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &redirectsMockAPIClient{
		listRedirectsResp: &api.RedirectsListResponse{
			Items: []api.Redirect{
				{
					ID:        "redir_123",
					Path:      "/old-page",
					Target:    "/new-page",
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupRedirectsMockFactories(mockClient)
	defer cleanup()

	cmd := newRedirectsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("path", "", "")
	cmd.Flags().String("target", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := redirectsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "redir_123") {
		t.Errorf("JSON output should contain redirect ID, got %q", output)
	}
}

// TestRedirectsGetRunE tests the redirects get command with mock API.
func TestRedirectsGetRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		redirectID string
		mockResp   *api.Redirect
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful get",
			redirectID: "redir_123",
			mockResp: &api.Redirect{
				ID:        "redir_123",
				Path:      "/old-page",
				Target:    "/new-page",
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
		},
		{
			name:       "redirect not found",
			redirectID: "redir_999",
			mockErr:    errors.New("redirect not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &redirectsMockAPIClient{
				getRedirectResp: tt.mockResp,
				getRedirectErr:  tt.mockErr,
			}
			cleanup, _ := setupRedirectsMockFactories(mockClient)
			defer cleanup()

			cmd := newRedirectsTestCmd()

			err := redirectsGetCmd.RunE(cmd, []string{tt.redirectID})

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

// TestRedirectsGetRunEJSON tests the redirects get command with JSON output.
func TestRedirectsGetRunEJSON(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &redirectsMockAPIClient{
		getRedirectResp: &api.Redirect{
			ID:        "redir_123",
			Path:      "/old-page",
			Target:    "/new-page",
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
	}
	cleanup, buf := setupRedirectsMockFactories(mockClient)
	defer cleanup()

	cmd := newRedirectsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := redirectsGetCmd.RunE(cmd, []string{"redir_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "redir_123") {
		t.Errorf("JSON output should contain redirect ID, got %q", output)
	}
}

// TestRedirectsGetClientError tests the redirects get command when client fails.
func TestRedirectsGetClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origSecretsFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("credential store error")
	}

	cmd := newRedirectsTestCmd()
	err := redirectsGetCmd.RunE(cmd, []string{"redir_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestRedirectsCreateRunE tests the redirects create command with mock API.
func TestRedirectsCreateRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		path       string
		target     string
		mockResp   *api.Redirect
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful create",
			path:   "/old-page",
			target: "/new-page",
			mockResp: &api.Redirect{
				ID:        "redir_new",
				Path:      "/old-page",
				Target:    "/new-page",
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			wantOutput: "redir_new",
		},
		{
			name:    "API error",
			path:    "/old-page",
			target:  "/new-page",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &redirectsMockAPIClient{
				createRedirectResp: tt.mockResp,
				createRedirectErr:  tt.mockErr,
			}
			cleanup, _ := setupRedirectsMockFactories(mockClient)
			defer cleanup()

			cmd := newRedirectsTestCmd()
			cmd.Flags().String("path", "", "")
			cmd.Flags().String("target", "", "")
			_ = cmd.Flags().Set("path", tt.path)
			_ = cmd.Flags().Set("target", tt.target)

			err := redirectsCreateCmd.RunE(cmd, []string{})

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

// TestRedirectsCreateRunEJSON tests the redirects create command with JSON output.
func TestRedirectsCreateRunEJSON(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &redirectsMockAPIClient{
		createRedirectResp: &api.Redirect{
			ID:        "redir_new",
			Path:      "/old-page",
			Target:    "/new-page",
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
	}
	cleanup, buf := setupRedirectsMockFactories(mockClient)
	defer cleanup()

	cmd := newRedirectsTestCmd()
	cmd.Flags().String("path", "", "")
	cmd.Flags().String("target", "", "")
	_ = cmd.Flags().Set("path", "/old-page")
	_ = cmd.Flags().Set("target", "/new-page")
	_ = cmd.Flags().Set("output", "json")

	err := redirectsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "redir_new") {
		t.Errorf("JSON output should contain redirect ID, got %q", output)
	}
}

// TestRedirectsCreateDryRun tests the redirects create command with dry-run flag.
func TestRedirectsCreateDryRun(t *testing.T) {
	cmd := newRedirectsTestCmd()
	cmd.Flags().String("path", "", "")
	cmd.Flags().String("target", "", "")
	_ = cmd.Flags().Set("path", "/old-page")
	_ = cmd.Flags().Set("target", "/new-page")
	_ = cmd.Flags().Set("dry-run", "true")

	err := redirectsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRedirectsCreateClientError tests the redirects create command when client fails.
func TestRedirectsCreateClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origSecretsFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("credential store error")
	}

	cmd := newRedirectsTestCmd()
	cmd.Flags().String("path", "", "")
	cmd.Flags().String("target", "", "")
	_ = cmd.Flags().Set("path", "/old-page")
	_ = cmd.Flags().Set("target", "/new-page")

	err := redirectsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestRedirectsDeleteRunE tests the redirects delete command with mock API.
func TestRedirectsDeleteRunE(t *testing.T) {
	tests := []struct {
		name       string
		redirectID string
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful delete",
			redirectID: "redir_123",
			mockErr:    nil,
		},
		{
			name:       "redirect not found",
			redirectID: "redir_999",
			mockErr:    errors.New("redirect not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &redirectsMockAPIClient{
				deleteRedirectErr: tt.mockErr,
			}
			cleanup, _ := setupRedirectsMockFactories(mockClient)
			defer cleanup()

			cmd := newRedirectsTestCmd()

			err := redirectsDeleteCmd.RunE(cmd, []string{tt.redirectID})

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

// TestRedirectsDeleteDryRun tests the redirects delete command with dry-run flag.
func TestRedirectsDeleteDryRun(t *testing.T) {
	cmd := newRedirectsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := redirectsDeleteCmd.RunE(cmd, []string{"redir_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRedirectsDeleteNoConfirmation tests the redirects delete command without confirmation.
func TestRedirectsDeleteNoConfirmation(t *testing.T) {
	cmd := newRedirectsTestCmd()
	_ = cmd.Flags().Set("yes", "false")

	err := redirectsDeleteCmd.RunE(cmd, []string{"redir_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRedirectsDeleteClientError tests the redirects delete command when client fails.
func TestRedirectsDeleteClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origSecretsFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("credential store error")
	}

	cmd := newRedirectsTestCmd()

	err := redirectsDeleteCmd.RunE(cmd, []string{"redir_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestRedirectsListClientError tests the redirects list command when client fails.
func TestRedirectsListClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origSecretsFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("credential store error")
	}

	cmd := newRedirectsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("path", "", "")
	cmd.Flags().String("target", "", "")

	err := redirectsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestRedirectsListWithFilters tests the redirects list command with filter flags.
func TestRedirectsListWithFilters(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &redirectsMockAPIClient{
		listRedirectsResp: &api.RedirectsListResponse{
			Items: []api.Redirect{
				{
					ID:        "redir_filtered",
					Path:      "/specific-path",
					Target:    "/specific-target",
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupRedirectsMockFactories(mockClient)
	defer cleanup()

	cmd := newRedirectsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("path", "", "")
	cmd.Flags().String("target", "", "")
	_ = cmd.Flags().Set("path", "/specific-path")
	_ = cmd.Flags().Set("target", "/specific-target")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := redirectsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "redir_filtered") {
		t.Errorf("output should contain filtered redirect, got %q", output)
	}
}
