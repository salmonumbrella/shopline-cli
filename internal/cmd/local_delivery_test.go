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

// TestLocalDeliveryCommandSetup verifies local-delivery command initialization
func TestLocalDeliveryCommandSetup(t *testing.T) {
	if localDeliveryCmd.Use != "local-delivery" {
		t.Errorf("expected Use 'local-delivery', got %q", localDeliveryCmd.Use)
	}
	if localDeliveryCmd.Short != "Manage local delivery options" {
		t.Errorf("expected Short 'Manage local delivery options', got %q", localDeliveryCmd.Short)
	}
}

// TestLocalDeliverySubcommands verifies all subcommands are registered
func TestLocalDeliverySubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List local delivery options",
		"get":    "Get local delivery option details",
		"create": "Create a local delivery option",
		"delete": "Delete a local delivery option",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range localDeliveryCmd.Commands() {
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

// TestLocalDeliveryListFlags verifies list command flags exist with correct defaults
func TestLocalDeliveryListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"location-id", ""},
		{"active", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := localDeliveryListCmd.Flags().Lookup(f.name)
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

// TestLocalDeliveryCreateFlags verifies create command flags exist with correct defaults
func TestLocalDeliveryCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"description", ""},
		{"price", ""},
		{"free-above", ""},
		{"active", "true"},
		{"location-id", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := localDeliveryCreateCmd.Flags().Lookup(f.name)
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

// TestLocalDeliveryGetArgs verifies get command requires exactly 1 argument
func TestLocalDeliveryGetArgs(t *testing.T) {
	err := localDeliveryGetCmd.Args(localDeliveryGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = localDeliveryGetCmd.Args(localDeliveryGetCmd, []string{"delivery_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}

	err = localDeliveryGetCmd.Args(localDeliveryGetCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error when more than 1 arg provided")
	}
}

// TestLocalDeliveryDeleteArgs verifies delete command requires exactly 1 argument
func TestLocalDeliveryDeleteArgs(t *testing.T) {
	err := localDeliveryDeleteCmd.Args(localDeliveryDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = localDeliveryDeleteCmd.Args(localDeliveryDeleteCmd, []string{"delivery_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestLocalDeliveryListGetClientError verifies list command error handling when getClient fails
func TestLocalDeliveryListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := localDeliveryListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestLocalDeliveryGetGetClientError verifies get command error handling when getClient fails
func TestLocalDeliveryGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := localDeliveryGetCmd.RunE(cmd, []string{"delivery_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestLocalDeliveryCreateGetClientError verifies create command error handling when getClient fails
func TestLocalDeliveryCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("price", "5.00", "")
	cmd.Flags().String("free-above", "", "")
	cmd.Flags().Bool("active", true, "")
	cmd.Flags().String("location-id", "", "")

	err := localDeliveryCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestLocalDeliveryDeleteGetClientError verifies delete command error handling when getClient fails
func TestLocalDeliveryDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := localDeliveryDeleteCmd.RunE(cmd, []string{"delivery_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestLocalDeliveryListNoProfiles verifies list command error handling when no profiles exist
func TestLocalDeliveryListNoProfiles(t *testing.T) {
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
	err := localDeliveryListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("expected 'no store profiles' error, got: %v", err)
	}
}

// TestLocalDeliveryWithMockStore tests local delivery commands with a mock credential store
func TestLocalDeliveryWithMockStore(t *testing.T) {
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

// TestLocalDeliveryListWithEnvVar verifies list command uses SHOPLINE_STORE env var
func TestLocalDeliveryListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Setenv("SHOPLINE_STORE", "envstore")

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"envstore", "other"},
			creds: map[string]*secrets.StoreCredentials{
				"envstore": {Handle: "test", AccessToken: "token123"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	err := localDeliveryListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// localDeliveryMockAPIClient is a mock implementation of api.APIClient for local delivery tests.
type localDeliveryMockAPIClient struct {
	api.MockClient
	listLocalDeliveryResp   *api.LocalDeliveryListResponse
	listLocalDeliveryErr    error
	getLocalDeliveryResp    *api.LocalDeliveryOption
	getLocalDeliveryErr     error
	createLocalDeliveryResp *api.LocalDeliveryOption
	createLocalDeliveryErr  error
	deleteLocalDeliveryErr  error
}

func (m *localDeliveryMockAPIClient) ListLocalDeliveryOptions(ctx context.Context, opts *api.LocalDeliveryListOptions) (*api.LocalDeliveryListResponse, error) {
	return m.listLocalDeliveryResp, m.listLocalDeliveryErr
}

func (m *localDeliveryMockAPIClient) GetLocalDeliveryOption(ctx context.Context, id string) (*api.LocalDeliveryOption, error) {
	return m.getLocalDeliveryResp, m.getLocalDeliveryErr
}

func (m *localDeliveryMockAPIClient) CreateLocalDeliveryOption(ctx context.Context, req *api.LocalDeliveryCreateRequest) (*api.LocalDeliveryOption, error) {
	return m.createLocalDeliveryResp, m.createLocalDeliveryErr
}

func (m *localDeliveryMockAPIClient) DeleteLocalDeliveryOption(ctx context.Context, id string) error {
	return m.deleteLocalDeliveryErr
}

// setupLocalDeliveryMockFactories sets up mock factories for local delivery tests.
func setupLocalDeliveryMockFactories(mockClient *localDeliveryMockAPIClient) (func(), *bytes.Buffer) {
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

// newLocalDeliveryTestCmd creates a test command with common flags for local delivery tests.
func newLocalDeliveryTestCmd() *cobra.Command {
	cmd := newTestCmdWithFlags()
	cmd.SetContext(context.Background())
	_ = cmd.Flags().Set("yes", "true")
	return cmd
}

// TestLocalDeliveryListRunE tests the local delivery list command with mock API.
func TestLocalDeliveryListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.LocalDeliveryListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.LocalDeliveryListResponse{
				Items: []api.LocalDeliveryOption{
					{
						ID:               "ld_123",
						Name:             "Local Delivery",
						Price:            "5.00",
						Currency:         "USD",
						FreeAbove:        "50.00",
						Active:           true,
						DeliveryTimeMin:  1,
						DeliveryTimeMax:  2,
						DeliveryTimeUnit: "days",
						Zones: []api.LocalDeliveryZone{
							{ID: "zone_1", Name: "Downtown", Type: "zip_code"},
						},
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "ld_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.LocalDeliveryListResponse{
				Items:      []api.LocalDeliveryOption{},
				TotalCount: 0,
			},
		},
		{
			name: "option without delivery time",
			mockResp: &api.LocalDeliveryListResponse{
				Items: []api.LocalDeliveryOption{
					{
						ID:        "ld_456",
						Name:      "Express",
						Price:     "10.00",
						Currency:  "USD",
						FreeAbove: "",
						Active:    true,
						Zones:     []api.LocalDeliveryZone{},
					},
				},
				TotalCount: 1,
			},
			wantOutput: "Express",
		},
		{
			name: "multiple options",
			mockResp: &api.LocalDeliveryListResponse{
				Items: []api.LocalDeliveryOption{
					{ID: "ld_1", Name: "Standard", Price: "5.00", Currency: "USD", Active: true},
					{ID: "ld_2", Name: "Express", Price: "10.00", Currency: "USD", Active: false},
				},
				TotalCount: 2,
			},
			wantOutput: "Standard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &localDeliveryMockAPIClient{
				listLocalDeliveryResp: tt.mockResp,
				listLocalDeliveryErr:  tt.mockErr,
			}
			cleanup, buf := setupLocalDeliveryMockFactories(mockClient)
			defer cleanup()

			cmd := newLocalDeliveryTestCmd()
			cmd.Flags().String("location-id", "", "")
			cmd.Flags().String("active", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := localDeliveryListCmd.RunE(cmd, []string{})

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

// TestLocalDeliveryListRunEWithJSON tests JSON output format.
func TestLocalDeliveryListRunEWithJSON(t *testing.T) {
	mockClient := &localDeliveryMockAPIClient{
		listLocalDeliveryResp: &api.LocalDeliveryListResponse{
			Items: []api.LocalDeliveryOption{
				{ID: "ld_json", Name: "JSON Test", Price: "5.00", Currency: "USD"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupLocalDeliveryMockFactories(mockClient)
	defer cleanup()

	cmd := newLocalDeliveryTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().String("active", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := localDeliveryListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ld_json") {
		t.Errorf("JSON output should contain option ID, got: %s", output)
	}
}

// TestLocalDeliveryListWithActiveFilter tests active filter parsing.
func TestLocalDeliveryListWithActiveFilter(t *testing.T) {
	tests := []struct {
		name       string
		activeFlag string
	}{
		{"active=true", "true"},
		{"active=false", "false"},
		{"no filter", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &localDeliveryMockAPIClient{
				listLocalDeliveryResp: &api.LocalDeliveryListResponse{
					Items:      []api.LocalDeliveryOption{},
					TotalCount: 0,
				},
			}
			cleanup, _ := setupLocalDeliveryMockFactories(mockClient)
			defer cleanup()

			cmd := newLocalDeliveryTestCmd()
			cmd.Flags().String("location-id", "", "")
			cmd.Flags().String("active", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			if tt.activeFlag != "" {
				_ = cmd.Flags().Set("active", tt.activeFlag)
			}

			err := localDeliveryListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestLocalDeliveryGetRunE tests the local delivery get command with mock API.
func TestLocalDeliveryGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		optionID string
		mockResp *api.LocalDeliveryOption
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful get",
			optionID: "ld_123",
			mockResp: &api.LocalDeliveryOption{
				ID:               "ld_123",
				Name:             "Local Delivery",
				Description:      "Same day delivery",
				Active:           true,
				Price:            "5.00",
				Currency:         "USD",
				FreeAbove:        "50.00",
				MinOrderAmount:   "10.00",
				MaxOrderAmount:   "500.00",
				DeliveryTimeMin:  1,
				DeliveryTimeMax:  2,
				DeliveryTimeUnit: "days",
				LocationID:       "loc_456",
				Zones: []api.LocalDeliveryZone{
					{ID: "zone_1", Name: "Downtown", Type: "zip_code", ZipCodes: []string{"10001", "10002"}},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:     "option not found",
			optionID: "ld_999",
			mockErr:  errors.New("local delivery option not found"),
			wantErr:  true,
		},
		{
			name:     "option with distance zone",
			optionID: "ld_789",
			mockResp: &api.LocalDeliveryOption{
				ID:       "ld_789",
				Name:     "Distance Based",
				Active:   true,
				Price:    "7.00",
				Currency: "USD",
				Zones: []api.LocalDeliveryZone{
					{ID: "zone_2", Name: "5km radius", Type: "radius", MinDistance: 0, MaxDistance: 5.0},
				},
				CreatedAt: time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:     "minimal option",
			optionID: "ld_minimal",
			mockResp: &api.LocalDeliveryOption{
				ID:        "ld_minimal",
				Name:      "Basic",
				Active:    false,
				Price:     "3.00",
				Currency:  "EUR",
				CreatedAt: time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &localDeliveryMockAPIClient{
				getLocalDeliveryResp: tt.mockResp,
				getLocalDeliveryErr:  tt.mockErr,
			}
			cleanup, _ := setupLocalDeliveryMockFactories(mockClient)
			defer cleanup()

			cmd := newLocalDeliveryTestCmd()

			err := localDeliveryGetCmd.RunE(cmd, []string{tt.optionID})

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

// TestLocalDeliveryGetRunEWithJSON tests JSON output format for get command.
func TestLocalDeliveryGetRunEWithJSON(t *testing.T) {
	mockClient := &localDeliveryMockAPIClient{
		getLocalDeliveryResp: &api.LocalDeliveryOption{
			ID:        "ld_json",
			Name:      "JSON Test",
			Price:     "5.00",
			Currency:  "USD",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	cleanup, buf := setupLocalDeliveryMockFactories(mockClient)
	defer cleanup()

	cmd := newLocalDeliveryTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := localDeliveryGetCmd.RunE(cmd, []string{"ld_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ld_json") {
		t.Errorf("JSON output should contain option ID, got: %s", output)
	}
}

// TestLocalDeliveryCreateRunE tests the local delivery create command with mock API.
func TestLocalDeliveryCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]string
		mockResp *api.LocalDeliveryOption
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			flags: map[string]string{
				"name":  "New Delivery",
				"price": "5.00",
			},
			mockResp: &api.LocalDeliveryOption{
				ID:        "ld_new",
				Name:      "New Delivery",
				Price:     "5.00",
				Currency:  "USD",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "create with all options",
			flags: map[string]string{
				"name":        "Full Option",
				"description": "Full delivery option",
				"price":       "10.00",
				"free-above":  "100.00",
				"location-id": "loc_123",
			},
			mockResp: &api.LocalDeliveryOption{
				ID:         "ld_full",
				Name:       "Full Option",
				Price:      "10.00",
				FreeAbove:  "100.00",
				LocationID: "loc_123",
				Active:     true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		{
			name: "API error",
			flags: map[string]string{
				"name":  "Error Delivery",
				"price": "5.00",
			},
			mockErr: errors.New("failed to create"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &localDeliveryMockAPIClient{
				createLocalDeliveryResp: tt.mockResp,
				createLocalDeliveryErr:  tt.mockErr,
			}
			cleanup, _ := setupLocalDeliveryMockFactories(mockClient)
			defer cleanup()

			cmd := newLocalDeliveryTestCmd()
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("price", "", "")
			cmd.Flags().String("free-above", "", "")
			cmd.Flags().Bool("active", true, "")
			cmd.Flags().String("location-id", "", "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			err := localDeliveryCreateCmd.RunE(cmd, []string{})

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

// TestLocalDeliveryCreateRunEWithJSON tests JSON output format for create command.
func TestLocalDeliveryCreateRunEWithJSON(t *testing.T) {
	mockClient := &localDeliveryMockAPIClient{
		createLocalDeliveryResp: &api.LocalDeliveryOption{
			ID:        "ld_json_create",
			Name:      "JSON Create Test",
			Price:     "5.00",
			Currency:  "USD",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	cleanup, buf := setupLocalDeliveryMockFactories(mockClient)
	defer cleanup()

	cmd := newLocalDeliveryTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("name", "JSON Create Test", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("price", "5.00", "")
	cmd.Flags().String("free-above", "", "")
	cmd.Flags().Bool("active", true, "")
	cmd.Flags().String("location-id", "", "")

	err := localDeliveryCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ld_json_create") {
		t.Errorf("JSON output should contain option ID, got: %s", output)
	}
}

// TestLocalDeliveryDeleteRunE tests the local delivery delete command with mock API.
func TestLocalDeliveryDeleteRunE(t *testing.T) {
	tests := []struct {
		name     string
		optionID string
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful delete",
			optionID: "ld_123",
			mockErr:  nil,
		},
		{
			name:     "option not found",
			optionID: "ld_999",
			mockErr:  errors.New("local delivery option not found"),
			wantErr:  true,
		},
		{
			name:     "API error",
			optionID: "ld_error",
			mockErr:  errors.New("delete failed"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &localDeliveryMockAPIClient{
				deleteLocalDeliveryErr: tt.mockErr,
			}
			cleanup, _ := setupLocalDeliveryMockFactories(mockClient)
			defer cleanup()

			cmd := newLocalDeliveryTestCmd()

			err := localDeliveryDeleteCmd.RunE(cmd, []string{tt.optionID})

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

// TestLocalDeliveryGetDisplayFields tests the get command executes successfully with all fields.
func TestLocalDeliveryGetDisplayFields(t *testing.T) {
	mockClient := &localDeliveryMockAPIClient{
		getLocalDeliveryResp: &api.LocalDeliveryOption{
			ID:               "ld_display",
			Name:             "Display Test",
			Description:      "Test description",
			Active:           true,
			Price:            "5.00",
			Currency:         "USD",
			FreeAbove:        "50.00",
			MinOrderAmount:   "10.00",
			MaxOrderAmount:   "200.00",
			DeliveryTimeMin:  1,
			DeliveryTimeMax:  3,
			DeliveryTimeUnit: "hours",
			LocationID:       "loc_abc",
			Zones: []api.LocalDeliveryZone{
				{ID: "z1", Name: "Zone 1", Type: "zip_code", ZipCodes: []string{"12345", "67890"}},
				{ID: "z2", Name: "Zone 2", Type: "radius", MinDistance: 0, MaxDistance: 10.5},
			},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupLocalDeliveryMockFactories(mockClient)
	defer cleanup()

	cmd := newLocalDeliveryTestCmd()

	err := localDeliveryGetCmd.RunE(cmd, []string{"ld_display"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Text output goes to stdout via fmt.Printf, not captured in buffer
	// This test verifies the command executes successfully with all field types
}

// TestLocalDeliveryGetOptionalFields tests get command with various optional field combinations.
func TestLocalDeliveryGetOptionalFields(t *testing.T) {
	tests := []struct {
		name   string
		option *api.LocalDeliveryOption
	}{
		{
			name: "with description",
			option: &api.LocalDeliveryOption{
				ID:          "ld_1",
				Name:        "Test",
				Description: "Has description",
				Price:       "5.00",
				Currency:    "USD",
				Active:      true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		{
			name: "without description",
			option: &api.LocalDeliveryOption{
				ID:        "ld_2",
				Name:      "Test",
				Price:     "5.00",
				Currency:  "USD",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "with free above zero",
			option: &api.LocalDeliveryOption{
				ID:        "ld_3",
				Name:      "Test",
				Price:     "5.00",
				Currency:  "USD",
				FreeAbove: "0",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "with free above value",
			option: &api.LocalDeliveryOption{
				ID:        "ld_4",
				Name:      "Test",
				Price:     "5.00",
				Currency:  "USD",
				FreeAbove: "50.00",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &localDeliveryMockAPIClient{
				getLocalDeliveryResp: tt.option,
			}
			cleanup, _ := setupLocalDeliveryMockFactories(mockClient)
			defer cleanup()

			cmd := newLocalDeliveryTestCmd()
			err := localDeliveryGetCmd.RunE(cmd, []string{tt.option.ID})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Text output goes to stdout via fmt.Printf
			// This test verifies the command executes successfully with various field combinations
		})
	}
}

// TestLocalDeliveryListTableOutput tests table output formatting.
func TestLocalDeliveryListTableOutput(t *testing.T) {
	mockClient := &localDeliveryMockAPIClient{
		listLocalDeliveryResp: &api.LocalDeliveryListResponse{
			Items: []api.LocalDeliveryOption{
				{
					ID:               "ld_table",
					Name:             "Table Test",
					Price:            "5.00",
					Currency:         "USD",
					FreeAbove:        "50.00",
					Active:           true,
					DeliveryTimeMin:  1,
					DeliveryTimeMax:  2,
					DeliveryTimeUnit: "days",
					Zones:            []api.LocalDeliveryZone{{}, {}},
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupLocalDeliveryMockFactories(mockClient)
	defer cleanup()

	cmd := newLocalDeliveryTestCmd()
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().String("active", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := localDeliveryListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	expectedElements := []string{
		"ld_table",
		"Table Test",
		"5.00 USD",
		"50.00",
		"true",
		"2",
		"1-2 days",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("table output should contain %q, got: %s", elem, output)
		}
	}
}
