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

func TestStoreCreditsCmd(t *testing.T) {
	if storeCreditsCmd.Use != "store-credits" {
		t.Errorf("Expected Use to be 'store-credits', got %q", storeCreditsCmd.Use)
	}
}
func TestStoreCreditsListCmd(t *testing.T) {
	if storeCreditsListCmd.Use != "list" {
		t.Errorf("Expected Use, got %q", storeCreditsListCmd.Use)
	}
}
func TestStoreCreditsGetCmd(t *testing.T) {
	if storeCreditsGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use, got %q", storeCreditsGetCmd.Use)
	}
}
func TestStoreCreditsCreateCmd(t *testing.T) {
	if storeCreditsCreateCmd.Use != "create" {
		t.Errorf("Expected Use, got %q", storeCreditsCreateCmd.Use)
	}
}
func TestStoreCreditsDeleteCmd(t *testing.T) {
	if storeCreditsDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use, got %q", storeCreditsDeleteCmd.Use)
	}
}
func TestStoreCreditsListFlags(t *testing.T) {
	flags := []string{"customer-id", "page", "page-size"}
	for _, flag := range flags {
		if storeCreditsListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}
func TestStoreCreditsCreateFlags(t *testing.T) {
	flags := []string{"customer-id", "amount", "currency", "description"}
	for _, flag := range flags {
		if storeCreditsCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}
func TestStoreCreditsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := storeCreditsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}
func TestStoreCreditsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := storeCreditsGetCmd.RunE(cmd, []string{"credit_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}
func TestStoreCreditsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("amount", "50.00", "")
	cmd.Flags().String("currency", "USD", "")
	cmd.Flags().String("description", "Refund", "")
	err := storeCreditsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}
func TestStoreCreditsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := storeCreditsDeleteCmd.RunE(cmd, []string{"credit_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}
func TestStoreCreditsListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := storeCreditsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}
func TestStoreCreditsCreateRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("amount", "50.00", "")
	cmd.Flags().String("currency", "USD", "")
	cmd.Flags().String("description", "Refund", "")
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := storeCreditsCreateCmd.RunE(cmd, []string{})
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
func TestStoreCreditsDeleteRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd := newTestCmdWithFlags()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := storeCreditsDeleteCmd.RunE(cmd, []string{"credit_123"})
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

// storeCreditsTestClient is a mock implementation for store credits testing.
type storeCreditsTestClient struct {
	api.MockClient

	listStoreCreditsResp  *api.StoreCreditsListResponse
	listStoreCreditsErr   error
	getStoreCreditResp    *api.StoreCredit
	getStoreCreditErr     error
	createStoreCreditResp *api.StoreCredit
	createStoreCreditErr  error
	deleteStoreCreditErr  error
}

func (m *storeCreditsTestClient) ListStoreCredits(ctx context.Context, opts *api.StoreCreditsListOptions) (*api.StoreCreditsListResponse, error) {
	return m.listStoreCreditsResp, m.listStoreCreditsErr
}

func (m *storeCreditsTestClient) GetStoreCredit(ctx context.Context, id string) (*api.StoreCredit, error) {
	return m.getStoreCreditResp, m.getStoreCreditErr
}

func (m *storeCreditsTestClient) CreateStoreCredit(ctx context.Context, req *api.StoreCreditCreateRequest) (*api.StoreCredit, error) {
	return m.createStoreCreditResp, m.createStoreCreditErr
}

func (m *storeCreditsTestClient) DeleteStoreCredit(ctx context.Context, id string) error {
	return m.deleteStoreCreditErr
}

// TestStoreCreditsListRunE tests the store credits list command execution with mock API.
func TestStoreCreditsListRunE(t *testing.T) {
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
		mockResp   *api.StoreCreditsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.StoreCreditsListResponse{
				Items: []api.StoreCredit{
					{
						ID:         "sc_123",
						CustomerID: "cust_456",
						Amount:     "100.00",
						Balance:    "75.00",
						Currency:   "USD",
						CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sc_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.StoreCreditsListResponse{
				Items:      []api.StoreCredit{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storeCreditsTestClient{
				listStoreCreditsResp: tt.mockResp,
				listStoreCreditsErr:  tt.mockErr,
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
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := storeCreditsListCmd.RunE(cmd, []string{})

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

// TestStoreCreditsGetRunE tests the store credits get command execution with mock API.
func TestStoreCreditsGetRunE(t *testing.T) {
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
		creditID string
		mockResp *api.StoreCredit
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful get",
			creditID: "sc_123",
			mockResp: &api.StoreCredit{
				ID:          "sc_123",
				CustomerID:  "cust_456",
				Amount:      "100.00",
				Balance:     "75.00",
				Currency:    "USD",
				Description: "Store credit refund",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:     "credit not found",
			creditID: "sc_999",
			mockErr:  errors.New("store credit not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storeCreditsTestClient{
				getStoreCreditResp: tt.mockResp,
				getStoreCreditErr:  tt.mockErr,
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

			err := storeCreditsGetCmd.RunE(cmd, []string{tt.creditID})

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

// TestStoreCreditsCreateRunE tests the store credits create command execution with mock API.
func TestStoreCreditsCreateRunE(t *testing.T) {
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
		mockResp *api.StoreCredit
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.StoreCredit{
				ID:         "sc_new",
				CustomerID: "cust_123",
				Amount:     "50.00",
				Balance:    "50.00",
				Currency:   "USD",
				CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("failed to create store credit"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storeCreditsTestClient{
				createStoreCreditResp: tt.mockResp,
				createStoreCreditErr:  tt.mockErr,
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
			cmd.Flags().Bool("dry-run", false, "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().String("amount", "50.00", "")
			cmd.Flags().String("currency", "USD", "")
			cmd.Flags().String("description", "Refund", "")

			err := storeCreditsCreateCmd.RunE(cmd, []string{})

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

// TestStoreCreditsDeleteRunE tests the store credits delete command execution with mock API.
func TestStoreCreditsDeleteRunE(t *testing.T) {
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
		name     string
		creditID string
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful delete",
			creditID: "sc_123",
			mockErr:  nil,
		},
		{
			name:     "delete fails",
			creditID: "sc_456",
			mockErr:  errors.New("store credit not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storeCreditsTestClient{
				deleteStoreCreditErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("dry-run", false, "")

			err := storeCreditsDeleteCmd.RunE(cmd, []string{tt.creditID})

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
