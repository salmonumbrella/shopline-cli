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

func TestAbandonedCheckoutsCommandStructure(t *testing.T) {
	if abandonedCheckoutsCmd == nil {
		t.Fatal("abandonedCheckoutsCmd is nil")
	}
	if abandonedCheckoutsCmd.Use != "abandoned-checkouts" {
		t.Errorf("Expected Use 'abandoned-checkouts', got %q", abandonedCheckoutsCmd.Use)
	}
	subcommands := map[string]bool{"list": false, "get": false, "send-recovery": false}
	for _, cmd := range abandonedCheckoutsCmd.Commands() {
		for key := range subcommands {
			if strings.HasPrefix(cmd.Use, key) {
				subcommands[key] = true
			}
		}
	}
	for name, found := range subcommands {
		if !found {
			t.Errorf("Subcommand %q not found", name)
		}
	}
}

func TestAbandonedCheckoutsListFlags(t *testing.T) {
	cmd := abandonedCheckoutsListCmd
	flags := []struct{ name, defaultValue string }{{"status", ""}, {"customer-id", ""}, {"page", "1"}, {"page-size", "20"}}
	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("Flag %q not found", f.name)
		} else if flag.DefValue != f.defaultValue {
			t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestAbandonedCheckoutsGetRequiresArg(t *testing.T) {
	cmd := abandonedCheckoutsGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"checkout_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestAbandonedCheckoutsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := abandonedCheckoutsListCmd.RunE(cmd, []string{}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAbandonedCheckoutsListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := abandonedCheckoutsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestAbandonedCheckoutsListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() { secretsStoreFactory = origFactory; _ = os.Setenv("SHOPLINE_STORE", origEnv) }()
	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{"envstore", "other"}, creds: map[string]*secrets.StoreCredentials{"envstore": {Handle: "test", AccessToken: "token123"}}}, nil
	}
	cmd := newTestCmdWithFlags()
	err := abandonedCheckoutsListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// abandonedCheckoutsMockAPIClient is a mock implementation of api.APIClient for abandoned checkouts tests.
type abandonedCheckoutsMockAPIClient struct {
	api.MockClient
	listAbandonedCheckoutsResp *api.AbandonedCheckoutsListResponse
	listAbandonedCheckoutsErr  error
	getAbandonedCheckoutResp   *api.AbandonedCheckout
	getAbandonedCheckoutErr    error
	sendRecoveryEmailErr       error
}

func (m *abandonedCheckoutsMockAPIClient) ListAbandonedCheckouts(ctx context.Context, opts *api.AbandonedCheckoutsListOptions) (*api.AbandonedCheckoutsListResponse, error) {
	return m.listAbandonedCheckoutsResp, m.listAbandonedCheckoutsErr
}

func (m *abandonedCheckoutsMockAPIClient) GetAbandonedCheckout(ctx context.Context, id string) (*api.AbandonedCheckout, error) {
	return m.getAbandonedCheckoutResp, m.getAbandonedCheckoutErr
}

func (m *abandonedCheckoutsMockAPIClient) SendAbandonedCheckoutRecoveryEmail(ctx context.Context, id string) error {
	return m.sendRecoveryEmailErr
}

// setupAbandonedCheckoutsMockFactories sets up mock factories for abandoned checkouts tests.
func setupAbandonedCheckoutsMockFactories(mockClient *abandonedCheckoutsMockAPIClient) (func(), *bytes.Buffer) {
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

// newAbandonedCheckoutsTestCmd creates a test command with common flags.
func newAbandonedCheckoutsTestCmd() *cobra.Command {
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

// TestAbandonedCheckoutsListRunE tests the abandoned checkouts list command with mock API.
func TestAbandonedCheckoutsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.AbandonedCheckoutsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.AbandonedCheckoutsListResponse{
				Items: []api.AbandonedCheckout{
					{
						ID:         "checkout_123",
						Email:      "customer@example.com",
						TotalPrice: "99.99",
						Currency:   "USD",
						LineItems: []api.AbandonedCheckoutLineItem{
							{Title: "Product 1", Quantity: 2, Price: 49.99},
						},
						RecoveryEmailSentCount: 1,
						CreatedAt:              time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "checkout_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.AbandonedCheckoutsListResponse{
				Items:      []api.AbandonedCheckout{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple checkouts",
			mockResp: &api.AbandonedCheckoutsListResponse{
				Items: []api.AbandonedCheckout{
					{ID: "checkout_1", Email: "user1@example.com", TotalPrice: "50.00", Currency: "USD", CreatedAt: time.Now()},
					{ID: "checkout_2", Email: "user2@example.com", TotalPrice: "75.00", Currency: "USD", CreatedAt: time.Now()},
				},
				TotalCount: 2,
			},
			wantOutput: "checkout_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &abandonedCheckoutsMockAPIClient{
				listAbandonedCheckoutsResp: tt.mockResp,
				listAbandonedCheckoutsErr:  tt.mockErr,
			}
			cleanup, buf := setupAbandonedCheckoutsMockFactories(mockClient)
			defer cleanup()

			cmd := newAbandonedCheckoutsTestCmd()
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := abandonedCheckoutsListCmd.RunE(cmd, []string{})

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

// TestAbandonedCheckoutsListRunEWithJSON tests JSON output format.
func TestAbandonedCheckoutsListRunEWithJSON(t *testing.T) {
	mockClient := &abandonedCheckoutsMockAPIClient{
		listAbandonedCheckoutsResp: &api.AbandonedCheckoutsListResponse{
			Items: []api.AbandonedCheckout{
				{ID: "checkout_json", Email: "json@example.com"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupAbandonedCheckoutsMockFactories(mockClient)
	defer cleanup()

	cmd := newAbandonedCheckoutsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := abandonedCheckoutsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "checkout_json") {
		t.Errorf("JSON output should contain checkout ID, got: %s", output)
	}
}

// TestAbandonedCheckoutsGetRunE tests the abandoned checkouts get command with mock API.
func TestAbandonedCheckoutsGetRunE(t *testing.T) {
	completedAt := time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)
	closedAt := time.Date(2024, 2, 2, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		checkoutID string
		mockResp   *api.AbandonedCheckout
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful get",
			checkoutID: "checkout_123",
			mockResp: &api.AbandonedCheckout{
				ID:                     "checkout_123",
				Email:                  "customer@example.com",
				Phone:                  "+1234567890",
				CustomerID:             "cust_123",
				CustomerLocale:         "en-US",
				TotalPrice:             "99.99",
				SubtotalPrice:          "89.99",
				TotalTax:               "10.00",
				TotalDiscounts:         "5.00",
				Currency:               "USD",
				RecoveryEmailSentCount: 2,
				RecoveryURL:            "https://example.com/recover/abc",
				CreatedAt:              time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:              time.Date(2024, 1, 20, 15, 0, 0, 0, time.UTC),
				LineItems: []api.AbandonedCheckoutLineItem{
					{Title: "Product 1", VariantName: "Large", Quantity: 2, Price: 44.99},
				},
			},
		},
		{
			name:       "checkout not found",
			checkoutID: "checkout_999",
			mockErr:    errors.New("checkout not found"),
			wantErr:    true,
		},
		{
			name:       "get checkout with completed and closed dates",
			checkoutID: "checkout_456",
			mockResp: &api.AbandonedCheckout{
				ID:          "checkout_456",
				Email:       "completed@example.com",
				TotalPrice:  "50.00",
				Currency:    "USD",
				CompletedAt: &completedAt,
				ClosedAt:    &closedAt,
			},
		},
		{
			name:       "get checkout without phone",
			checkoutID: "checkout_789",
			mockResp: &api.AbandonedCheckout{
				ID:         "checkout_789",
				Email:      "nophone@example.com",
				Phone:      "",
				TotalPrice: "75.00",
				Currency:   "USD",
			},
		},
		{
			name:       "get checkout with zero discounts",
			checkoutID: "checkout_nodiscount",
			mockResp: &api.AbandonedCheckout{
				ID:             "checkout_nodiscount",
				Email:          "nodiscount@example.com",
				TotalPrice:     "100.00",
				TotalDiscounts: "0",
				Currency:       "USD",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &abandonedCheckoutsMockAPIClient{
				getAbandonedCheckoutResp: tt.mockResp,
				getAbandonedCheckoutErr:  tt.mockErr,
			}
			cleanup, _ := setupAbandonedCheckoutsMockFactories(mockClient)
			defer cleanup()

			cmd := newAbandonedCheckoutsTestCmd()

			err := abandonedCheckoutsGetCmd.RunE(cmd, []string{tt.checkoutID})

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

// TestAbandonedCheckoutsGetRunEWithJSON tests JSON output format for get command.
func TestAbandonedCheckoutsGetRunEWithJSON(t *testing.T) {
	mockClient := &abandonedCheckoutsMockAPIClient{
		getAbandonedCheckoutResp: &api.AbandonedCheckout{
			ID:         "checkout_json",
			Email:      "json@example.com",
			TotalPrice: "50.00",
			Currency:   "USD",
		},
	}
	cleanup, buf := setupAbandonedCheckoutsMockFactories(mockClient)
	defer cleanup()

	cmd := newAbandonedCheckoutsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := abandonedCheckoutsGetCmd.RunE(cmd, []string{"checkout_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "checkout_json") {
		t.Errorf("JSON output should contain checkout ID, got: %s", output)
	}
}

// TestAbandonedCheckoutsSendRecoveryRunE tests the send-recovery command.
func TestAbandonedCheckoutsSendRecoveryRunE(t *testing.T) {
	tests := []struct {
		name       string
		checkoutID string
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful send",
			checkoutID: "checkout_123",
			mockErr:    nil,
		},
		{
			name:       "send fails",
			checkoutID: "checkout_456",
			mockErr:    errors.New("email already sent"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &abandonedCheckoutsMockAPIClient{
				sendRecoveryEmailErr: tt.mockErr,
			}
			cleanup, _ := setupAbandonedCheckoutsMockFactories(mockClient)
			defer cleanup()

			cmd := newAbandonedCheckoutsTestCmd()

			err := abandonedCheckoutsSendRecoveryCmd.RunE(cmd, []string{tt.checkoutID})

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
