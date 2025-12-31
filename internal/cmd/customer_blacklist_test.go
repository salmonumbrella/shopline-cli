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

func TestCustomerBlacklistCmd(t *testing.T) {
	if customerBlacklistCmd.Use != "customer-blacklist" {
		t.Errorf("Expected Use to be 'customer-blacklist', got %q", customerBlacklistCmd.Use)
	}
	if customerBlacklistCmd.Short != "Manage customer blacklist" {
		t.Errorf("Expected Short to be 'Manage customer blacklist', got %q", customerBlacklistCmd.Short)
	}
}

func TestCustomerBlacklistListCmd(t *testing.T) {
	if customerBlacklistListCmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %q", customerBlacklistListCmd.Use)
	}
	if customerBlacklistListCmd.Short != "List blacklisted customers" {
		t.Errorf("Expected Short to be 'List blacklisted customers', got %q", customerBlacklistListCmd.Short)
	}
}

func TestCustomerBlacklistGetCmd(t *testing.T) {
	if customerBlacklistGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use to be 'get <id>', got %q", customerBlacklistGetCmd.Use)
	}
	if customerBlacklistGetCmd.Short != "Get blacklist entry details" {
		t.Errorf("Expected Short to be 'Get blacklist entry details', got %q", customerBlacklistGetCmd.Short)
	}
}

func TestCustomerBlacklistCreateCmd(t *testing.T) {
	if customerBlacklistCreateCmd.Use != "create" {
		t.Errorf("Expected Use to be 'create', got %q", customerBlacklistCreateCmd.Use)
	}
	if customerBlacklistCreateCmd.Short != "Add a customer to blacklist" {
		t.Errorf("Expected Short to be 'Add a customer to blacklist', got %q", customerBlacklistCreateCmd.Short)
	}
}

func TestCustomerBlacklistDeleteCmd(t *testing.T) {
	if customerBlacklistDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use to be 'delete <id>', got %q", customerBlacklistDeleteCmd.Use)
	}
	if customerBlacklistDeleteCmd.Short != "Remove a customer from blacklist" {
		t.Errorf("Expected Short to be 'Remove a customer from blacklist', got %q", customerBlacklistDeleteCmd.Short)
	}
}

func TestCustomerBlacklistListFlags(t *testing.T) {
	flags := []string{"email", "phone", "page", "page-size"}
	for _, flag := range flags {
		if customerBlacklistListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestCustomerBlacklistCreateFlags(t *testing.T) {
	flags := []string{"customer-id", "email", "phone", "reason"}
	for _, flag := range flags {
		if customerBlacklistCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestCustomerBlacklistListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := customerBlacklistListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerBlacklistGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := customerBlacklistGetCmd.RunE(cmd, []string{"blacklist_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerBlacklistCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("reason", "fraud", "")
	err := customerBlacklistCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerBlacklistDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := customerBlacklistDeleteCmd.RunE(cmd, []string{"blacklist_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerBlacklistListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := customerBlacklistListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

func TestCustomerBlacklistGetRunE_MultipleProfiles(t *testing.T) {
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
	err := customerBlacklistGetCmd.RunE(cmd, []string{"blacklist_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles without selection, got nil")
	}
}

func TestCustomerBlacklistCreateRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("reason", "fraud", "")
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	err := customerBlacklistCreateCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run output, got: %s", output)
	}
}

func TestCustomerBlacklistDeleteRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newTestCmdWithFlags()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	err := customerBlacklistDeleteCmd.RunE(cmd, []string{"blacklist_123"})

	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run output, got: %s", output)
	}
}

// blacklistMockClient is a mock implementation for customer blacklist testing.
type blacklistMockClient struct {
	api.MockClient
	listResp   *api.CustomerBlacklistListResponse
	listErr    error
	getResp    *api.CustomerBlacklist
	getErr     error
	createResp *api.CustomerBlacklist
	createErr  error
	deleteErr  error
}

func (m *blacklistMockClient) ListCustomerBlacklist(ctx context.Context, opts *api.CustomerBlacklistListOptions) (*api.CustomerBlacklistListResponse, error) {
	return m.listResp, m.listErr
}

func (m *blacklistMockClient) GetCustomerBlacklist(ctx context.Context, id string) (*api.CustomerBlacklist, error) {
	return m.getResp, m.getErr
}

func (m *blacklistMockClient) CreateCustomerBlacklist(ctx context.Context, req *api.CustomerBlacklistCreateRequest) (*api.CustomerBlacklist, error) {
	return m.createResp, m.createErr
}

func (m *blacklistMockClient) DeleteCustomerBlacklist(ctx context.Context, id string) error {
	return m.deleteErr
}

func TestCustomerBlacklistListRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *blacklistMockClient
		wantErr    bool
		errContain string
	}{
		{
			name: "success with entries",
			mockClient: &blacklistMockClient{
				listResp: &api.CustomerBlacklistListResponse{
					Items: []api.CustomerBlacklist{
						{
							ID:         "bl_123",
							CustomerID: "cust_123",
							Email:      "bad@example.com",
							Reason:     "fraud",
							CreatedAt:  now,
						},
					},
					TotalCount: 1,
				},
			},
		},
		{
			name: "API error",
			mockClient: &blacklistMockClient{
				listErr: errors.New("API connection failed"),
			},
			wantErr:    true,
			errContain: "failed to list customer blacklist",
		},
		{
			name: "empty list",
			mockClient: &blacklistMockClient{
				listResp: &api.CustomerBlacklistListResponse{
					Items:      []api.CustomerBlacklist{},
					TotalCount: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("email", "", "")
			cmd.Flags().String("phone", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := customerBlacklistListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCustomerBlacklistGetRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *blacklistMockClient
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			mockClient: &blacklistMockClient{
				getResp: &api.CustomerBlacklist{
					ID:         "bl_123",
					CustomerID: "cust_123",
					Email:      "bad@example.com",
					Phone:      "+1234567890",
					Reason:     "fraud",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
		},
		{
			name: "not found",
			mockClient: &blacklistMockClient{
				getErr: errors.New("entry not found"),
			},
			wantErr:    true,
			errContain: "failed to get blacklist entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")

			err := customerBlacklistGetCmd.RunE(cmd, []string{"bl_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCustomerBlacklistCreateRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *blacklistMockClient
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			mockClient: &blacklistMockClient{
				createResp: &api.CustomerBlacklist{
					ID:        "bl_new",
					Email:     "bad@example.com",
					Reason:    "fraud",
					CreatedAt: now,
				},
			},
		},
		{
			name: "API error",
			mockClient: &blacklistMockClient{
				createErr: errors.New("validation failed"),
			},
			wantErr:    true,
			errContain: "failed to add to blacklist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Bool("dry-run", false, "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().String("email", "bad@example.com", "")
			cmd.Flags().String("phone", "", "")
			cmd.Flags().String("reason", "fraud", "")

			err := customerBlacklistCreateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCustomerBlacklistDeleteRunE_WithMockAPI(t *testing.T) {
	tests := []struct {
		name       string
		mockClient *blacklistMockClient
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			mockClient: &blacklistMockClient{
				deleteErr: nil,
			},
		},
		{
			name: "API error",
			mockClient: &blacklistMockClient{
				deleteErr: errors.New("entry not found"),
			},
			wantErr:    true,
			errContain: "failed to remove from blacklist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Bool("dry-run", false, "")

			err := customerBlacklistDeleteCmd.RunE(cmd, []string{"bl_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}
