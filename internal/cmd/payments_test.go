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

// TestPaymentsCommandSetup verifies payments command initialization
func TestPaymentsCommandSetup(t *testing.T) {
	if paymentsCmd.Use != "payments" {
		t.Errorf("expected Use 'payments', got %q", paymentsCmd.Use)
	}
	if paymentsCmd.Short != "Manage payments" {
		t.Errorf("expected Short 'Manage payments', got %q", paymentsCmd.Short)
	}
}

// TestPaymentsSubcommands verifies all subcommands are registered
func TestPaymentsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":            "List payments",
		"get":             "Get payment details",
		"order":           "List payments for an order",
		"account-summary": "Get payments account summary (via Admin API)",
		"payouts":         "List payment payouts (via Admin API)",
		"capture":         "Capture an authorized payment",
		"void":            "Void an authorized payment",
		"refund":          "Refund a captured payment",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range paymentsCmd.Commands() {
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

// TestPaymentsListFlags verifies list command flags exist with correct defaults
func TestPaymentsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
		{"gateway", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := paymentsListCmd.Flags().Lookup(f.name)
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

func TestPaymentsGetCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"pay_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"pay_1", "pay_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paymentsGetCmd.Args(paymentsGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentsOrderCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"ord_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"ord_1", "ord_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paymentsOrderCmd.Args(paymentsOrderCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentsCaptureCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"pay_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"pay_1", "pay_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paymentsCaptureCmd.Args(paymentsCaptureCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentsCaptureCmdFlags(t *testing.T) {
	amountFlag := paymentsCaptureCmd.Flags().Lookup("amount")
	if amountFlag == nil {
		t.Error("Missing --amount flag")
	}
}

func TestPaymentsVoidCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"pay_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"pay_1", "pay_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paymentsVoidCmd.Args(paymentsVoidCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentsRefundCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"pay_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"pay_1", "pay_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paymentsRefundCmd.Args(paymentsRefundCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentsRefundCmdFlags(t *testing.T) {
	flags := []string{"amount", "reason"}
	for _, flagName := range flags {
		flag := paymentsRefundCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Missing --%s flag", flagName)
		}
	}
}

func TestPaymentsPayoutsFlags(t *testing.T) {
	flag := paymentsPayoutsCmd.Flags().Lookup("from")
	if flag == nil {
		t.Fatal("flag 'from' not found")
	}
	if flag.DefValue != "0" {
		t.Errorf("expected default '0', got %q", flag.DefValue)
	}

	if flag.Annotations == nil {
		t.Fatal("flag 'from' has no annotations (expected required)")
	}
	if _, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; !ok {
		t.Error("flag 'from' is not marked as required")
	}
}

func TestPaymentsAccountSummaryRunE_NoAdminToken(t *testing.T) {
	t.Setenv("SHOPLINE_ADMIN_BASE_URL", "https://test.example.com")
	origToken := os.Getenv("SHOPLINE_ADMIN_TOKEN")
	origMerchant := os.Getenv("SHOPLINE_ADMIN_MERCHANT_ID")
	defer func() {
		_ = os.Setenv("SHOPLINE_ADMIN_TOKEN", origToken)
		_ = os.Setenv("SHOPLINE_ADMIN_MERCHANT_ID", origMerchant)
	}()
	_ = os.Unsetenv("SHOPLINE_ADMIN_TOKEN")
	_ = os.Unsetenv("SHOPLINE_ADMIN_MERCHANT_ID")

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")

	err := paymentsAccountSummaryCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPaymentsPayoutsRunE_NoAdminToken(t *testing.T) {
	t.Setenv("SHOPLINE_ADMIN_BASE_URL", "https://test.example.com")
	origToken := os.Getenv("SHOPLINE_ADMIN_TOKEN")
	origMerchant := os.Getenv("SHOPLINE_ADMIN_MERCHANT_ID")
	defer func() {
		_ = os.Setenv("SHOPLINE_ADMIN_TOKEN", origToken)
		_ = os.Setenv("SHOPLINE_ADMIN_MERCHANT_ID", origMerchant)
	}()
	_ = os.Unsetenv("SHOPLINE_ADMIN_TOKEN")
	_ = os.Unsetenv("SHOPLINE_ADMIN_MERCHANT_ID")

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")
	cmd.Flags().Int64("from", 1704067200000, "")

	err := paymentsPayoutsCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPaymentsPayoutsRunE_InvalidFrom(t *testing.T) {
	t.Setenv("SHOPLINE_ADMIN_BASE_URL", "https://test.example.com")
	t.Setenv("SHOPLINE_ADMIN_TOKEN", "test-token")
	t.Setenv("SHOPLINE_ADMIN_MERCHANT_ID", "test-merchant")

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")
	cmd.Flags().Int64("from", 0, "")

	err := paymentsPayoutsCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "--from must be a positive Unix timestamp in milliseconds" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPaymentsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(paymentsListCmd)

	err := paymentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestPaymentsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(paymentsGetCmd)

	err := paymentsGetCmd.RunE(cmd, []string{"pay_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestPaymentsOrderGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(paymentsOrderCmd)

	err := paymentsOrderCmd.RunE(cmd, []string{"ord_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestPaymentsCaptureGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(paymentsCaptureCmd)
	cmd.Flags().String("amount", "", "")

	err := paymentsCaptureCmd.RunE(cmd, []string{"pay_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestPaymentsVoidGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(paymentsVoidCmd)
	// yes flag already added by newTestCmdWithFlags()

	err := paymentsVoidCmd.RunE(cmd, []string{"pay_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestPaymentsRefundGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(paymentsRefundCmd)
	cmd.Flags().String("amount", "", "")
	cmd.Flags().String("reason", "", "")
	// yes flag already added by newTestCmdWithFlags()

	err := paymentsRefundCmd.RunE(cmd, []string{"pay_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestPaymentsGetClientError verifies error handling when getClient fails
func TestPaymentsGetClientError(t *testing.T) {
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

// TestPaymentsWithMockStore tests payments commands with a mock credential store
func TestPaymentsWithMockStore(t *testing.T) {
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

// TestPaymentsListWithValidStore tests list command execution with valid store
func TestPaymentsListWithValidStore(t *testing.T) {
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

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(paymentsListCmd)

	// This will fail at the API call level, but validates the client setup works
	err := paymentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("paymentsListCmd succeeded (might be due to mock setup)")
	}
}

// paymentsTestClient is a mock implementation for payments testing.
type paymentsTestClient struct {
	api.MockClient

	listPaymentsResp      *api.PaymentsListResponse
	listPaymentsErr       error
	getPaymentResp        *api.Payment
	getPaymentErr         error
	listOrderPaymentsResp *api.PaymentsListResponse
	listOrderPaymentsErr  error
	capturePaymentResp    *api.Payment
	capturePaymentErr     error
	voidPaymentResp       *api.Payment
	voidPaymentErr        error
	refundPaymentResp     *api.Payment
	refundPaymentErr      error
}

func (m *paymentsTestClient) ListPayments(ctx context.Context, opts *api.PaymentsListOptions) (*api.PaymentsListResponse, error) {
	return m.listPaymentsResp, m.listPaymentsErr
}

func (m *paymentsTestClient) GetPayment(ctx context.Context, id string) (*api.Payment, error) {
	return m.getPaymentResp, m.getPaymentErr
}

func (m *paymentsTestClient) ListOrderPayments(ctx context.Context, orderID string) (*api.PaymentsListResponse, error) {
	return m.listOrderPaymentsResp, m.listOrderPaymentsErr
}

func (m *paymentsTestClient) CapturePayment(ctx context.Context, id string, amount string) (*api.Payment, error) {
	return m.capturePaymentResp, m.capturePaymentErr
}

func (m *paymentsTestClient) VoidPayment(ctx context.Context, id string) (*api.Payment, error) {
	return m.voidPaymentResp, m.voidPaymentErr
}

func (m *paymentsTestClient) RefundPayment(ctx context.Context, id string, amount string, reason string) (*api.Payment, error) {
	return m.refundPaymentResp, m.refundPaymentErr
}

// TestPaymentsListRunE tests the payments list command execution with mock API.
func TestPaymentsListRunE(t *testing.T) {
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
		mockResp   *api.PaymentsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.PaymentsListResponse{
				Items: []api.Payment{
					{
						ID:            "pay_123",
						OrderID:       "ord_456",
						Amount:        "99.99",
						Currency:      "USD",
						Status:        "captured",
						Gateway:       "stripe",
						PaymentMethod: "credit_card",
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "pay_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.PaymentsListResponse{
				Items:      []api.Payment{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &paymentsTestClient{
				listPaymentsResp: tt.mockResp,
				listPaymentsErr:  tt.mockErr,
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
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("gateway", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := paymentsListCmd.RunE(cmd, []string{})

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

// TestPaymentsGetRunE tests the payments get command execution with mock API.
func TestPaymentsGetRunE(t *testing.T) {
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
		name      string
		paymentID string
		mockResp  *api.Payment
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			paymentID: "pay_123",
			mockResp: &api.Payment{
				ID:            "pay_123",
				OrderID:       "ord_456",
				Amount:        "99.99",
				Currency:      "USD",
				Status:        "captured",
				Gateway:       "stripe",
				PaymentMethod: "credit_card",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
			},
		},
		{
			name:      "payment not found",
			paymentID: "pay_999",
			mockErr:   errors.New("payment not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &paymentsTestClient{
				getPaymentResp: tt.mockResp,
				getPaymentErr:  tt.mockErr,
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

			err := paymentsGetCmd.RunE(cmd, []string{tt.paymentID})

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

// TestPaymentsOrderRunE tests the payments order command execution with mock API.
func TestPaymentsOrderRunE(t *testing.T) {
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
		orderID  string
		mockResp *api.PaymentsListResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful list order payments",
			orderID: "ord_123",
			mockResp: &api.PaymentsListResponse{
				Items: []api.Payment{
					{
						ID:       "pay_123",
						Amount:   "50.00",
						Currency: "USD",
						Status:   "captured",
					},
				},
				TotalCount: 1,
			},
		},
		{
			name:    "order not found",
			orderID: "ord_999",
			mockErr: errors.New("order not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &paymentsTestClient{
				listOrderPaymentsResp: tt.mockResp,
				listOrderPaymentsErr:  tt.mockErr,
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

			err := paymentsOrderCmd.RunE(cmd, []string{tt.orderID})

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

// TestPaymentsCaptureRunE tests the payments capture command execution with mock API.
func TestPaymentsCaptureRunE(t *testing.T) {
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
		name      string
		paymentID string
		mockResp  *api.Payment
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful capture",
			paymentID: "pay_123",
			mockResp: &api.Payment{
				ID:     "pay_123",
				Status: "captured",
			},
		},
		{
			name:      "capture fails",
			paymentID: "pay_456",
			mockErr:   errors.New("payment already captured"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &paymentsTestClient{
				capturePaymentResp: tt.mockResp,
				capturePaymentErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("amount", "", "")

			err := paymentsCaptureCmd.RunE(cmd, []string{tt.paymentID})

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

// TestPaymentsVoidRunE tests the payments void command execution with mock API.
func TestPaymentsVoidRunE(t *testing.T) {
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
		name      string
		paymentID string
		mockResp  *api.Payment
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful void",
			paymentID: "pay_123",
			mockResp: &api.Payment{
				ID:     "pay_123",
				Status: "voided",
			},
		},
		{
			name:      "void fails",
			paymentID: "pay_456",
			mockErr:   errors.New("payment cannot be voided"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &paymentsTestClient{
				voidPaymentResp: tt.mockResp,
				voidPaymentErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := paymentsVoidCmd.RunE(cmd, []string{tt.paymentID})

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

// TestPaymentsRefundRunE tests the payments refund command execution with mock API.
func TestPaymentsRefundRunE(t *testing.T) {
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
		name      string
		paymentID string
		mockResp  *api.Payment
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful refund",
			paymentID: "pay_123",
			mockResp: &api.Payment{
				ID:     "pay_123",
				Status: "refunded",
			},
		},
		{
			name:      "refund fails",
			paymentID: "pay_456",
			mockErr:   errors.New("insufficient funds for refund"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &paymentsTestClient{
				refundPaymentResp: tt.mockResp,
				refundPaymentErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("amount", "", "")
			cmd.Flags().String("reason", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := paymentsRefundCmd.RunE(cmd, []string{tt.paymentID})

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

// TestPaymentsListJSONOutput tests JSON output format for list command.
func TestPaymentsListJSONOutput(t *testing.T) {
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

	mockClient := &paymentsTestClient{
		listPaymentsResp: &api.PaymentsListResponse{
			Items: []api.Payment{
				{
					ID:            "pay_json_123",
					OrderID:       "ord_456",
					Amount:        "99.99",
					Currency:      "USD",
					Status:        "captured",
					Gateway:       "stripe",
					PaymentMethod: "credit_card",
					CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
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
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("gateway", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := paymentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "pay_json_123") {
		t.Errorf("expected JSON output to contain payment ID, got: %s", output)
	}
}

// TestPaymentsGetJSONOutput tests JSON output format for get command.
func TestPaymentsGetJSONOutput(t *testing.T) {
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

	mockClient := &paymentsTestClient{
		getPaymentResp: &api.Payment{
			ID:            "pay_json_456",
			OrderID:       "ord_789",
			Amount:        "150.00",
			Currency:      "EUR",
			Status:        "authorized",
			Gateway:       "paypal",
			PaymentMethod: "paypal",
			CreatedAt:     time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2024, 2, 20, 14, 5, 0, 0, time.UTC),
		},
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

	err := paymentsGetCmd.RunE(cmd, []string{"pay_json_456"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "pay_json_456") {
		t.Errorf("expected JSON output to contain payment ID, got: %s", output)
	}
}

// TestPaymentsOrderJSONOutput tests JSON output format for order command.
func TestPaymentsOrderJSONOutput(t *testing.T) {
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

	mockClient := &paymentsTestClient{
		listOrderPaymentsResp: &api.PaymentsListResponse{
			Items: []api.Payment{
				{
					ID:       "pay_order_123",
					Amount:   "75.00",
					Currency: "GBP",
					Status:   "captured",
				},
			},
			TotalCount: 1,
		},
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

	err := paymentsOrderCmd.RunE(cmd, []string{"ord_test"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "pay_order_123") {
		t.Errorf("expected JSON output to contain payment ID, got: %s", output)
	}
}

// TestPaymentsGetWithTransactionID tests payment display with TransactionID.
func TestPaymentsGetWithTransactionID(t *testing.T) {
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

	mockClient := &paymentsTestClient{
		getPaymentResp: &api.Payment{
			ID:            "pay_tx_123",
			OrderID:       "ord_789",
			Amount:        "200.00",
			Currency:      "USD",
			Status:        "captured",
			Gateway:       "stripe",
			PaymentMethod: "credit_card",
			TransactionID: "txn_abc123xyz",
			CreatedAt:     time.Date(2024, 3, 10, 9, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2024, 3, 10, 9, 5, 0, 0, time.UTC),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "") // text output
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := paymentsGetCmd.RunE(cmd, []string{"pay_tx_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Transaction ID: txn_abc123xyz") {
		t.Errorf("expected output to contain Transaction ID, got: %s", output)
	}
}

// TestPaymentsGetWithErrorMessage tests payment display with ErrorMessage.
func TestPaymentsGetWithErrorMessage(t *testing.T) {
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

	mockClient := &paymentsTestClient{
		getPaymentResp: &api.Payment{
			ID:            "pay_err_123",
			OrderID:       "ord_789",
			Amount:        "50.00",
			Currency:      "USD",
			Status:        "failed",
			Gateway:       "stripe",
			PaymentMethod: "credit_card",
			ErrorMessage:  "Card declined",
			CreatedAt:     time.Date(2024, 3, 12, 11, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2024, 3, 12, 11, 0, 0, 0, time.UTC),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "") // text output
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := paymentsGetCmd.RunE(cmd, []string{"pay_err_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Error:          Card declined") {
		t.Errorf("expected output to contain Error message, got: %s", output)
	}
}

// TestPaymentsGetWithCreditCard tests payment display with CreditCard details.
func TestPaymentsGetWithCreditCard(t *testing.T) {
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

	mockClient := &paymentsTestClient{
		getPaymentResp: &api.Payment{
			ID:            "pay_cc_123",
			OrderID:       "ord_789",
			Amount:        "125.00",
			Currency:      "USD",
			Status:        "captured",
			Gateway:       "stripe",
			PaymentMethod: "credit_card",
			CreditCard: &api.CreditCard{
				Brand:       "Visa",
				Last4:       "4242",
				ExpiryMonth: 12,
				ExpiryYear:  2025,
			},
			CreatedAt: time.Date(2024, 4, 5, 8, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 4, 5, 8, 5, 0, 0, time.UTC),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "") // text output
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := paymentsGetCmd.RunE(cmd, []string{"pay_cc_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Card:           Visa ****4242 (12/2025)") {
		t.Errorf("expected output to contain Card details, got: %s", output)
	}
}

// TestPaymentsVoidConfirmationCancelled tests void command when user cancels.
func TestPaymentsVoidConfirmationCancelled(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origStdin := os.Stdin
	origStdout := os.Stdout
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &paymentsTestClient{
		voidPaymentResp: &api.Payment{
			ID:     "pay_void_cancel",
			Status: "voided",
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	// Simulate user typing "n" for cancel
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("n\n")
	_ = w.Close()
	os.Stdin = r

	// Capture stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", false, "") // Don't skip confirmation

	err := paymentsVoidCmd.RunE(cmd, []string{"pay_void_cancel"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_ = wOut.Close()
	var stdout bytes.Buffer
	_, _ = stdout.ReadFrom(rOut)
	output := stdout.String()

	if !strings.Contains(output, "Cancelled") {
		t.Errorf("expected 'Cancelled' in output, got: %s", output)
	}
}

// TestPaymentsRefundConfirmationCancelled tests refund command when user cancels.
func TestPaymentsRefundConfirmationCancelled(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origStdin := os.Stdin
	origStdout := os.Stdout
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &paymentsTestClient{
		refundPaymentResp: &api.Payment{
			ID:     "pay_refund_cancel",
			Status: "refunded",
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	// Simulate user typing "n" for cancel
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("n\n")
	_ = w.Close()
	os.Stdin = r

	// Capture stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("amount", "", "")
	cmd.Flags().String("reason", "", "")
	cmd.Flags().Bool("yes", false, "") // Don't skip confirmation

	err := paymentsRefundCmd.RunE(cmd, []string{"pay_refund_cancel"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_ = wOut.Close()
	var stdout bytes.Buffer
	_, _ = stdout.ReadFrom(rOut)
	output := stdout.String()

	if !strings.Contains(output, "Cancelled") {
		t.Errorf("expected 'Cancelled' in output, got: %s", output)
	}
}

// TestPaymentsVoidConfirmationAccepted tests void command when user confirms with 'y'.
func TestPaymentsVoidConfirmationAccepted(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origStdin := os.Stdin
	origStdout := os.Stdout
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &paymentsTestClient{
		voidPaymentResp: &api.Payment{
			ID:     "pay_void_confirm",
			Status: "voided",
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	// Simulate user typing "y" to confirm
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("y\n")
	_ = w.Close()
	os.Stdin = r

	// Capture stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", false, "") // Don't skip confirmation

	err := paymentsVoidCmd.RunE(cmd, []string{"pay_void_confirm"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_ = wOut.Close()
	var stdout bytes.Buffer
	_, _ = stdout.ReadFrom(rOut)
	output := stdout.String()

	if !strings.Contains(output, "Voided payment pay_void_confirm") {
		t.Errorf("expected 'Voided payment' in output, got: %s", output)
	}
}

// TestPaymentsRefundConfirmationAccepted tests refund command when user confirms with 'Y'.
func TestPaymentsRefundConfirmationAccepted(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origStdin := os.Stdin
	origStdout := os.Stdout
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &paymentsTestClient{
		refundPaymentResp: &api.Payment{
			ID:     "pay_refund_confirm",
			Status: "refunded",
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	// Simulate user typing "Y" to confirm (uppercase)
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("Y\n")
	_ = w.Close()
	os.Stdin = r

	// Capture stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("amount", "50.00", "")
	cmd.Flags().String("reason", "customer request", "")
	cmd.Flags().Bool("yes", false, "") // Don't skip confirmation

	err := paymentsRefundCmd.RunE(cmd, []string{"pay_refund_confirm"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_ = wOut.Close()
	var stdout bytes.Buffer
	_, _ = stdout.ReadFrom(rOut)
	output := stdout.String()

	if !strings.Contains(output, "Refunded payment pay_refund_confirm") {
		t.Errorf("expected 'Refunded payment' in output, got: %s", output)
	}
}
