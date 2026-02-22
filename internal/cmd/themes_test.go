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

func TestThemesCommandSetup(t *testing.T) {
	if themesCmd.Use != "themes" {
		t.Errorf("expected Use 'themes', got %q", themesCmd.Use)
	}
	if themesCmd.Short != "Manage themes" {
		t.Errorf("expected Short 'Manage themes', got %q", themesCmd.Short)
	}
}

func TestThemesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List themes",
		"get":    "Get theme details",
		"create": "Create a theme",
		"delete": "Delete a theme",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range themesCmd.Commands() {
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

func TestThemesListGetClientError(t *testing.T) {
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

func TestThemesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"role", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := themesListCmd.Flags().Lookup(f.name)
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

func TestThemesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"role", ""},
		{"src", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := themesCreateCmd.Flags().Lookup(f.name)
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

func TestThemesWithMockStore(t *testing.T) {
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

// themesMockAPIClient is a mock implementation of api.APIClient for themes tests.
type themesMockAPIClient struct {
	api.MockClient
	listThemesResp  *api.ThemesListResponse
	listThemesErr   error
	getThemeResp    *api.Theme
	getThemeErr     error
	createThemeResp *api.Theme
	createThemeErr  error
	deleteThemeErr  error
}

func (m *themesMockAPIClient) ListThemes(ctx context.Context, opts *api.ThemesListOptions) (*api.ThemesListResponse, error) {
	return m.listThemesResp, m.listThemesErr
}

func (m *themesMockAPIClient) GetTheme(ctx context.Context, id string) (*api.Theme, error) {
	return m.getThemeResp, m.getThemeErr
}

func (m *themesMockAPIClient) CreateTheme(ctx context.Context, req *api.ThemeCreateRequest) (*api.Theme, error) {
	return m.createThemeResp, m.createThemeErr
}

func (m *themesMockAPIClient) DeleteTheme(ctx context.Context, id string) error {
	return m.deleteThemeErr
}

// setupThemesMockFactories sets up mock factories for themes tests.
func setupThemesMockFactories(mockClient *themesMockAPIClient) (func(), *bytes.Buffer) {
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

// newThemesTestCmd creates a test command with common flags.
func newThemesTestCmd() *cobra.Command {
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

// TestThemesListRunE tests the themes list command with mock API.
func TestThemesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ThemesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ThemesListResponse{
				Items: []api.Theme{
					{
						ID:          "theme_123",
						Name:        "Dawn",
						Role:        "main",
						Previewable: true,
						Processing:  false,
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "theme_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ThemesListResponse{
				Items:      []api.Theme{},
				TotalCount: 0,
			},
		},
		{
			name: "theme processing",
			mockResp: &api.ThemesListResponse{
				Items: []api.Theme{
					{
						ID:          "theme_456",
						Name:        "Custom Theme",
						Role:        "unpublished",
						Previewable: false,
						Processing:  true,
						CreatedAt:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "theme_456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &themesMockAPIClient{
				listThemesResp: tt.mockResp,
				listThemesErr:  tt.mockErr,
			}
			cleanup, buf := setupThemesMockFactories(mockClient)
			defer cleanup()

			cmd := newThemesTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("role", "", "")

			err := themesListCmd.RunE(cmd, []string{})

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

// TestThemesListRunEWithJSON tests JSON output format.
func TestThemesListRunEWithJSON(t *testing.T) {
	mockClient := &themesMockAPIClient{
		listThemesResp: &api.ThemesListResponse{
			Items: []api.Theme{
				{ID: "theme_json", Name: "JSON Theme"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupThemesMockFactories(mockClient)
	defer cleanup()

	cmd := newThemesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("role", "", "")

	err := themesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "theme_json") {
		t.Errorf("JSON output should contain theme ID, got: %s", output)
	}
}

// TestThemesGetRunE tests the themes get command with mock API.
func TestThemesGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		themeID  string
		mockResp *api.Theme
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get",
			themeID: "theme_123",
			mockResp: &api.Theme{
				ID:          "theme_123",
				Name:        "Dawn",
				Role:        "main",
				Previewable: true,
				Processing:  false,
				CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "theme not found",
			themeID: "theme_999",
			mockErr: errors.New("theme not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &themesMockAPIClient{
				getThemeResp: tt.mockResp,
				getThemeErr:  tt.mockErr,
			}
			cleanup, _ := setupThemesMockFactories(mockClient)
			defer cleanup()

			cmd := newThemesTestCmd()

			err := themesGetCmd.RunE(cmd, []string{tt.themeID})

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

// TestThemesGetRunEWithJSON tests JSON output format for get command.
func TestThemesGetRunEWithJSON(t *testing.T) {
	mockClient := &themesMockAPIClient{
		getThemeResp: &api.Theme{
			ID:   "theme_json",
			Name: "JSON Test Theme",
		},
	}
	cleanup, buf := setupThemesMockFactories(mockClient)
	defer cleanup()

	cmd := newThemesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := themesGetCmd.RunE(cmd, []string{"theme_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "theme_json") {
		t.Errorf("JSON output should contain theme ID, got: %s", output)
	}
}

// TestThemesCreateRunE tests the themes create command with mock API.
func TestThemesCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.Theme
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Theme{
				ID:   "theme_new",
				Name: "New Theme",
				Role: "unpublished",
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
			mockClient := &themesMockAPIClient{
				createThemeResp: tt.mockResp,
				createThemeErr:  tt.mockErr,
			}
			cleanup, _ := setupThemesMockFactories(mockClient)
			defer cleanup()

			cmd := newThemesTestCmd()
			cmd.Flags().String("name", "New Theme", "")
			cmd.Flags().String("role", "", "")
			cmd.Flags().String("src", "", "")

			err := themesCreateCmd.RunE(cmd, []string{})

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

// TestThemesCreateRunEWithJSON tests JSON output format for create command.
func TestThemesCreateRunEWithJSON(t *testing.T) {
	mockClient := &themesMockAPIClient{
		createThemeResp: &api.Theme{
			ID:   "theme_json_new",
			Name: "JSON New Theme",
			Role: "unpublished",
		},
	}
	cleanup, buf := setupThemesMockFactories(mockClient)
	defer cleanup()

	cmd := newThemesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("name", "JSON New Theme", "")
	cmd.Flags().String("role", "", "")
	cmd.Flags().String("src", "", "")

	err := themesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "theme_json_new") {
		t.Errorf("JSON output should contain theme ID, got: %s", output)
	}
}

// TestThemesDeleteRunE tests the themes delete command with mock API.
func TestThemesDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		themeID string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			themeID: "theme_123",
			mockErr: nil,
		},
		{
			name:    "delete fails",
			themeID: "theme_456",
			mockErr: errors.New("theme is main theme"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &themesMockAPIClient{
				deleteThemeErr: tt.mockErr,
			}
			cleanup, _ := setupThemesMockFactories(mockClient)
			defer cleanup()

			cmd := newThemesTestCmd()

			err := themesDeleteCmd.RunE(cmd, []string{tt.themeID})

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

// TestThemesDeleteRunEDryRun tests the dry-run mode of delete command.
func TestThemesDeleteRunEDryRun(t *testing.T) {
	mockClient := &themesMockAPIClient{}
	cleanup, _ := setupThemesMockFactories(mockClient)
	defer cleanup()

	cmd := newThemesTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := themesDeleteCmd.RunE(cmd, []string{"theme_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
