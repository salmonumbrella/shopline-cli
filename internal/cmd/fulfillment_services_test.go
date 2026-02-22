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

func TestFulfillmentServicesCmd(t *testing.T) {
	if fulfillmentServicesCmd.Use != "fulfillment-services" {
		t.Errorf("Expected Use 'fulfillment-services', got %q", fulfillmentServicesCmd.Use)
	}
	if fulfillmentServicesCmd.Short != "Manage fulfillment services" {
		t.Errorf("Expected Short 'Manage fulfillment services', got %q", fulfillmentServicesCmd.Short)
	}
}

func TestFulfillmentServicesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List fulfillment services",
		"get":    "Get fulfillment service details",
		"create": "Create a fulfillment service",
		"delete": "Delete a fulfillment service",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range fulfillmentServicesCmd.Commands() {
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

func TestFulfillmentServicesListCmd(t *testing.T) {
	if fulfillmentServicesListCmd.Use != "list" {
		t.Errorf("Expected Use 'list', got %q", fulfillmentServicesListCmd.Use)
	}
}

func TestFulfillmentServicesGetCmd(t *testing.T) {
	if fulfillmentServicesGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use 'get <id>', got %q", fulfillmentServicesGetCmd.Use)
	}
}

func TestFulfillmentServicesCreateCmd(t *testing.T) {
	if fulfillmentServicesCreateCmd.Use != "create" {
		t.Errorf("Expected Use 'create', got %q", fulfillmentServicesCreateCmd.Use)
	}
}

func TestFulfillmentServicesDeleteCmd(t *testing.T) {
	if fulfillmentServicesDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use 'delete <id>', got %q", fulfillmentServicesDeleteCmd.Use)
	}
}

func TestFulfillmentServicesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := fulfillmentServicesListCmd.Flags().Lookup(f.name)
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

func TestFulfillmentServicesCreateFlags(t *testing.T) {
	flags := []string{"name", "callback-url", "inventory-management", "tracking-support"}
	for _, flag := range flags {
		if fulfillmentServicesCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestFulfillmentServicesListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := fulfillmentServicesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentServicesGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := fulfillmentServicesGetCmd.RunE(cmd, []string{"fs_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentServicesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test Service", "")
	cmd.Flags().String("callback-url", "https://example.com/callback", "")
	cmd.Flags().Bool("inventory-management", false, "")
	cmd.Flags().Bool("tracking-support", false, "")
	err := fulfillmentServicesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentServicesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")
	err := fulfillmentServicesDeleteCmd.RunE(cmd, []string{"fs_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentServicesListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := fulfillmentServicesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentServicesCreateRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test Service", "")
	cmd.Flags().String("callback-url", "https://example.com/callback", "")
	cmd.Flags().Bool("inventory-management", false, "")
	cmd.Flags().Bool("tracking-support", false, "")
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := fulfillmentServicesCreateCmd.RunE(cmd, []string{})
	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run, got: %s", output)
	}
}

func TestFulfillmentServicesDeleteRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd := newTestCmdWithFlags()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := fulfillmentServicesDeleteCmd.RunE(cmd, []string{"fs_123"})
	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run, got: %s", output)
	}
}

func TestFulfillmentServicesDeleteRunE_NoConfirmation(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd := newTestCmdWithFlags()
	// yes is already false by default from newTestCmdWithFlags
	err := fulfillmentServicesDeleteCmd.RunE(cmd, []string{"fs_123"})
	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Are you sure") {
		t.Errorf("Expected confirmation prompt, got: %s", output)
	}
}

// fulfillmentServicesAPIClient is a mock implementation for fulfillment services testing.
type fulfillmentServicesAPIClient struct {
	api.MockClient

	listFulfillmentServicesResp *api.FulfillmentServicesListResponse
	listFulfillmentServicesErr  error

	getFulfillmentServiceResp *api.FulfillmentService
	getFulfillmentServiceErr  error

	createFulfillmentServiceResp *api.FulfillmentService
	createFulfillmentServiceErr  error

	deleteFulfillmentServiceErr error
}

func (m *fulfillmentServicesAPIClient) ListFulfillmentServices(ctx context.Context, opts *api.FulfillmentServicesListOptions) (*api.FulfillmentServicesListResponse, error) {
	return m.listFulfillmentServicesResp, m.listFulfillmentServicesErr
}

func (m *fulfillmentServicesAPIClient) GetFulfillmentService(ctx context.Context, id string) (*api.FulfillmentService, error) {
	return m.getFulfillmentServiceResp, m.getFulfillmentServiceErr
}

func (m *fulfillmentServicesAPIClient) CreateFulfillmentService(ctx context.Context, req *api.FulfillmentServiceCreateRequest) (*api.FulfillmentService, error) {
	return m.createFulfillmentServiceResp, m.createFulfillmentServiceErr
}

func (m *fulfillmentServicesAPIClient) DeleteFulfillmentService(ctx context.Context, id string) error {
	return m.deleteFulfillmentServiceErr
}

func TestFulfillmentServicesListRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name       string
		mockResp   *api.FulfillmentServicesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.FulfillmentServicesListResponse{
				Items: []api.FulfillmentService{
					{
						ID:                  "fs_123",
						Name:                "Test Service",
						CallbackURL:         "https://example.com/callback",
						InventoryManagement: true,
						TrackingSupport:     true,
						CreatedAt:           time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "fs_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.FulfillmentServicesListResponse{
				Items:      []api.FulfillmentService{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentServicesAPIClient{
				listFulfillmentServicesResp: tt.mockResp,
				listFulfillmentServicesErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := fulfillmentServicesListCmd.RunE(cmd, []string{})

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

func TestFulfillmentServicesGetRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name     string
		fsID     string
		mockResp *api.FulfillmentService
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			fsID: "fs_123",
			mockResp: &api.FulfillmentService{
				ID:                  "fs_123",
				Name:                "Test Service",
				CallbackURL:         "https://example.com/callback",
				InventoryManagement: true,
				TrackingSupport:     true,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			},
		},
		{
			name:    "service not found",
			fsID:    "fs_999",
			mockErr: errors.New("fulfillment service not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentServicesAPIClient{
				getFulfillmentServiceResp: tt.mockResp,
				getFulfillmentServiceErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := fulfillmentServicesGetCmd.RunE(cmd, []string{tt.fsID})

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

func TestFulfillmentServicesCreateRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name     string
		mockResp *api.FulfillmentService
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.FulfillmentService{
				ID:                  "fs_new",
				Name:                "New Service",
				CallbackURL:         "https://example.com/new",
				InventoryManagement: true,
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("invalid request"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentServicesAPIClient{
				createFulfillmentServiceResp: tt.mockResp,
				createFulfillmentServiceErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("name", "New Service", "")
			cmd.Flags().String("callback-url", "https://example.com/new", "")
			cmd.Flags().Bool("inventory-management", true, "")
			cmd.Flags().Bool("tracking-support", false, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := fulfillmentServicesCreateCmd.RunE(cmd, []string{})

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

func TestFulfillmentServicesDeleteRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name    string
		fsID    string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			fsID:    "fs_123",
			mockErr: nil,
		},
		{
			name:    "delete fails",
			fsID:    "fs_999",
			mockErr: errors.New("service not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentServicesAPIClient{
				deleteFulfillmentServiceErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := fulfillmentServicesDeleteCmd.RunE(cmd, []string{tt.fsID})

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
