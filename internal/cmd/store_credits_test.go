package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

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

func TestStoreCreditsCreateCmd(t *testing.T) {
	if storeCreditsCreateCmd.Use != "create" {
		t.Errorf("Expected Use, got %q", storeCreditsCreateCmd.Use)
	}
}

func TestStoreCreditsListFlags(t *testing.T) {
	flags := []string{"customer-id", "page", "per-page"}
	for _, flag := range flags {
		if storeCreditsListCmd.Flags().Lookup(flag) == nil && storeCreditsCmd.PersistentFlags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestStoreCreditsCreateFlags(t *testing.T) {
	flags := []string{"customer-id", "value", "remarks", "expires-at"}
	for _, flag := range flags {
		if storeCreditsCreateCmd.Flags().Lookup(flag) == nil && storeCreditsCmd.PersistentFlags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestStoreCreditsListRunE_MissingCustomerID(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	origClientFactory := clientFactory
	defer func() { clientFactory = origClientFactory }()
	clientFactory = func(handle, accessToken string) api.APIClient {
		return &storeCreditsTestClient{}
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().Int("per-page", 0, "")

	err := storeCreditsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for missing customer-id")
	}
	if !strings.Contains(err.Error(), "customer id is required") {
		t.Errorf("Expected customer id required error, got: %v", err)
	}
}

func TestStoreCreditsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().Int("per-page", 0, "")
	err := storeCreditsListCmd.RunE(cmd, []string{})
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
	cmd.Flags().Int("value", 222, "")
	cmd.Flags().String("remarks", "test", "")
	cmd.Flags().String("expires-at", "", "")
	err := storeCreditsCreateCmd.RunE(cmd, []string{})
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
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().Int("per-page", 0, "")
	err := storeCreditsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestStoreCreditsCreateRunE_DryRun(t *testing.T) {
	origWriter := formatterWriter
	defer func() { formatterWriter = origWriter }()

	var buf bytes.Buffer
	formatterWriter = &buf
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("value", 222, "")
	cmd.Flags().String("remarks", "直播抽獎購物金", "")
	cmd.Flags().String("expires-at", "", "")
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := storeCreditsCreateCmd.RunE(cmd, []string{})
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

	listResp   json.RawMessage
	listErr    error
	createResp json.RawMessage
	createErr  error
}

func (m *storeCreditsTestClient) ListCustomerStoreCredits(ctx context.Context, customerID string, page, perPage int) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func (m *storeCreditsTestClient) UpdateCustomerStoreCredits(ctx context.Context, customerID string, req *api.StoreCreditUpdateRequest) (json.RawMessage, error) {
	return m.createResp, m.createErr
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
		name    string
		resp    json.RawMessage
		err     error
		wantErr bool
	}{
		{
			name: "successful list",
			resp: json.RawMessage(`{"items":[{"value":100,"remarks":"test"}]}`),
		},
		{
			name:    "API error",
			err:     errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storeCreditsTestClient{
				listResp: tt.resp,
				listErr:  tt.err,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "json", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().Int("page", 0, "")
			cmd.Flags().Int("per-page", 0, "")

			err := storeCreditsListCmd.RunE(cmd, []string{})

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
		name    string
		resp    json.RawMessage
		err     error
		wantErr bool
	}{
		{
			name: "successful create",
			resp: json.RawMessage(`{"id":"sc_new","value":222}`),
		},
		{
			name:    "create fails",
			err:     errors.New("failed to update store credits"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storeCreditsTestClient{
				createResp: tt.resp,
				createErr:  tt.err,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "json", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Bool("dry-run", false, "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().Int("value", 222, "")
			cmd.Flags().String("remarks", "直播抽獎購物金", "")
			cmd.Flags().String("expires-at", "", "")

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
