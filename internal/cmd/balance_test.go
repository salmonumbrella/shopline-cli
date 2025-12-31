package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func TestBalanceGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := balanceGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestBalanceTransactionsGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().String("source-type", "", "")

	err := balanceTransactionsCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestBalanceTransactionGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := balanceTransactionGetCmd.RunE(cmd, []string{"txn-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestBalanceTransactionsFlags(t *testing.T) {
	flags := balanceTransactionsCmd.Flags()

	if flags.Lookup("page") == nil {
		t.Error("Expected page flag")
	}
	if flags.Lookup("page-size") == nil {
		t.Error("Expected page-size flag")
	}
	if flags.Lookup("type") == nil {
		t.Error("Expected type flag")
	}
	if flags.Lookup("source-type") == nil {
		t.Error("Expected source-type flag")
	}
}

func TestBalanceCommandStructure(t *testing.T) {
	if balanceCmd.Use != "balance" {
		t.Errorf("Expected Use 'balance', got %s", balanceCmd.Use)
	}

	subcommands := balanceCmd.Commands()
	expectedCmds := map[string]bool{
		"get":          false,
		"transactions": false,
		"transaction":  false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %s not found", name)
		}
	}
}

// balanceTestClient is a mock implementation for balance testing.
type balanceTestClient struct {
	api.MockClient

	getBalanceResp              *api.Balance
	getBalanceErr               error
	listBalanceTransactionsResp *api.BalanceTransactionsListResponse
	listBalanceTransactionsErr  error
	getBalanceTransactionResp   *api.BalanceTransaction
	getBalanceTransactionErr    error
}

func (m *balanceTestClient) GetBalance(ctx context.Context) (*api.Balance, error) {
	return m.getBalanceResp, m.getBalanceErr
}

func (m *balanceTestClient) ListBalanceTransactions(ctx context.Context, opts *api.BalanceTransactionsListOptions) (*api.BalanceTransactionsListResponse, error) {
	return m.listBalanceTransactionsResp, m.listBalanceTransactionsErr
}

func (m *balanceTestClient) GetBalanceTransaction(ctx context.Context, id string) (*api.BalanceTransaction, error) {
	return m.getBalanceTransactionResp, m.getBalanceTransactionErr
}

// setupBalanceTest sets up the test environment for balance commands.
func setupBalanceTest(t *testing.T) (cleanup func(), setBuf func() *bytes.Buffer) {
	t.Helper()

	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	cleanup = func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	setBuf = func() *bytes.Buffer {
		var buf bytes.Buffer
		formatterWriter = &buf
		return &buf
	}

	return cleanup, setBuf
}

// TestBalanceGetRunE tests the balance get command execution with mock API.
func TestBalanceGetRunE(t *testing.T) {
	cleanup, setBuf := setupBalanceTest(t)
	defer cleanup()

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.Balance
		mockErr    error
		output     string
		wantErr    bool
		wantOutput []string // Only used for JSON output (which goes through formatter)
	}{
		{
			name:   "successful get text output with reserved",
			output: "",
			mockResp: &api.Balance{
				Currency:  "USD",
				Available: "1000.00",
				Pending:   "250.00",
				Reserved:  "50.00",
				Total:     "1300.00",
				UpdatedAt: fixedTime,
			},
			// Text output goes to stdout via fmt.Printf, not formatter buffer
		},
		{
			name:   "successful get text output without reserved",
			output: "",
			mockResp: &api.Balance{
				Currency:  "EUR",
				Available: "500.00",
				Pending:   "100.00",
				Reserved:  "",
				Total:     "600.00",
				UpdatedAt: fixedTime,
			},
		},
		{
			name:   "successful get JSON output",
			output: "json",
			mockResp: &api.Balance{
				Currency:  "USD",
				Available: "1000.00",
				Pending:   "250.00",
				Total:     "1250.00",
				UpdatedAt: fixedTime,
			},
			wantOutput: []string{"USD", "1000.00"},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &balanceTestClient{
				getBalanceResp: tt.mockResp,
				getBalanceErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			buf := setBuf()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := balanceGetCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "failed to get balance") {
					t.Errorf("expected 'failed to get balance' in error, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check formatter output for JSON (text output goes to stdout via fmt.Printf)
			if len(tt.wantOutput) > 0 {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output %q should contain %q", output, want)
					}
				}
			}
		})
	}
}

// TestBalanceTransactionsRunE tests the balance transactions command execution with mock API.
func TestBalanceTransactionsRunE(t *testing.T) {
	cleanup, setBuf := setupBalanceTest(t)
	defer cleanup()

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.BalanceTransactionsListResponse
		mockErr    error
		output     string
		wantErr    bool
		wantOutput []string // For table/JSON output (via formatter)
	}{
		{
			name:   "successful list text output",
			output: "",
			mockResp: &api.BalanceTransactionsListResponse{
				Items: []api.BalanceTransaction{
					{
						ID:        "txn_123",
						Type:      "payment",
						Amount:    "100.00",
						Currency:  "USD",
						Net:       "97.00",
						Status:    "available",
						CreatedAt: fixedTime,
					},
					{
						ID:        "txn_456",
						Type:      "refund",
						Amount:    "-50.00",
						Currency:  "USD",
						Net:       "-50.00",
						Status:    "pending",
						CreatedAt: fixedTime,
					},
				},
				TotalCount: 2,
			},
			// Table output goes through formatter, "Showing X of Y" goes to stdout
			wantOutput: []string{"txn_123", "txn_456", "payment", "refund", "100.00"},
		},
		{
			name:   "successful list JSON output",
			output: "json",
			mockResp: &api.BalanceTransactionsListResponse{
				Items: []api.BalanceTransaction{
					{
						ID:        "txn_789",
						Type:      "payout",
						Amount:    "500.00",
						Currency:  "EUR",
						Net:       "500.00",
						Status:    "available",
						CreatedAt: fixedTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: []string{"txn_789", "payout"},
		},
		{
			name:   "empty list",
			output: "",
			mockResp: &api.BalanceTransactionsListResponse{
				Items:      []api.BalanceTransaction{},
				TotalCount: 0,
			},
			// Empty table still has headers
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &balanceTestClient{
				listBalanceTransactionsResp: tt.mockResp,
				listBalanceTransactionsErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			buf := setBuf()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("type", "", "")
			cmd.Flags().String("source-type", "", "")

			err := balanceTransactionsCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "failed to list balance transactions") {
					t.Errorf("expected 'failed to list balance transactions' in error, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check formatter output (table or JSON)
			if len(tt.wantOutput) > 0 {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output %q should contain %q", output, want)
					}
				}
			}
		})
	}
}

// TestBalanceTransactionGetRunE tests the balance transaction get command execution with mock API.
func TestBalanceTransactionGetRunE(t *testing.T) {
	cleanup, setBuf := setupBalanceTest(t)
	defer cleanup()

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	availableOn := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		txnID      string
		mockResp   *api.BalanceTransaction
		mockErr    error
		output     string
		wantErr    bool
		wantOutput []string // Only for JSON output (text goes to stdout via fmt.Printf)
	}{
		{
			name:   "successful get text output full details",
			txnID:  "txn_123",
			output: "",
			mockResp: &api.BalanceTransaction{
				ID:          "txn_123",
				Type:        "payment",
				Amount:      "100.00",
				Currency:    "USD",
				Net:         "97.00",
				Fee:         "3.00",
				Status:      "available",
				Description: "Payment for order #1234",
				SourceID:    "ord_1234",
				SourceType:  "order",
				AvailableOn: &availableOn,
				CreatedAt:   fixedTime,
			},
			// Text output goes to stdout via fmt.Printf, not formatter buffer
		},
		{
			name:   "successful get text output minimal details",
			txnID:  "txn_456",
			output: "",
			mockResp: &api.BalanceTransaction{
				ID:        "txn_456",
				Type:      "payout",
				Amount:    "500.00",
				Currency:  "EUR",
				Net:       "500.00",
				Fee:       "",
				Status:    "pending",
				CreatedAt: fixedTime,
			},
		},
		{
			name:   "successful get JSON output",
			txnID:  "txn_789",
			output: "json",
			mockResp: &api.BalanceTransaction{
				ID:        "txn_789",
				Type:      "refund",
				Amount:    "-50.00",
				Currency:  "USD",
				Net:       "-50.00",
				Status:    "available",
				CreatedAt: fixedTime,
			},
			wantOutput: []string{"txn_789", "refund", "-50.00"},
		},
		{
			name:    "transaction not found",
			txnID:   "txn_999",
			mockErr: errors.New("transaction not found"),
			wantErr: true,
		},
		{
			name:    "API error",
			txnID:   "txn_000",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &balanceTestClient{
				getBalanceTransactionResp: tt.mockResp,
				getBalanceTransactionErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			buf := setBuf()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := balanceTransactionGetCmd.RunE(cmd, []string{tt.txnID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "failed to get balance transaction") {
					t.Errorf("expected 'failed to get balance transaction' in error, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check formatter output for JSON (text goes to stdout via fmt.Printf)
			if len(tt.wantOutput) > 0 {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output %q should contain %q", output, want)
					}
				}
			}
		})
	}
}

// TestBalanceTransactionGetRequiresArg tests that balance transaction get requires an argument.
func TestBalanceTransactionGetRequiresArg(t *testing.T) {
	cmd := balanceTransactionGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"txn_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

// TestBalanceTransactionsWithFilters tests that filters are passed to the API.
func TestBalanceTransactionsWithFilters(t *testing.T) {
	cleanup, setBuf := setupBalanceTest(t)
	defer cleanup()

	var capturedOpts *api.BalanceTransactionsListOptions
	mockClient := &balanceTestClient{
		listBalanceTransactionsResp: &api.BalanceTransactionsListResponse{
			Items:      []api.BalanceTransaction{},
			TotalCount: 0,
		},
	}

	// Override the mock to capture the options
	originalClientFactory := clientFactory
	clientFactory = func(handle, accessToken string) api.APIClient {
		return &balanceTestClientWithCapture{
			balanceTestClient: mockClient,
			captureOpts:       func(opts *api.BalanceTransactionsListOptions) { capturedOpts = opts },
		}
	}
	defer func() { clientFactory = originalClientFactory }()

	_ = setBuf()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")
	cmd.Flags().String("type", "payment", "")
	cmd.Flags().String("source-type", "order", "")

	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "50")
	_ = cmd.Flags().Set("type", "payment")
	_ = cmd.Flags().Set("source-type", "order")

	err := balanceTransactionsCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if capturedOpts == nil {
		t.Fatal("Options were not captured")
	}
	if capturedOpts.Page != 2 {
		t.Errorf("Expected page 2, got %d", capturedOpts.Page)
	}
	if capturedOpts.PageSize != 50 {
		t.Errorf("Expected page-size 50, got %d", capturedOpts.PageSize)
	}
	if capturedOpts.Type != "payment" {
		t.Errorf("Expected type 'payment', got %s", capturedOpts.Type)
	}
	if capturedOpts.SourceType != "order" {
		t.Errorf("Expected source-type 'order', got %s", capturedOpts.SourceType)
	}
}

// balanceTestClientWithCapture wraps balanceTestClient to capture options.
type balanceTestClientWithCapture struct {
	*balanceTestClient
	captureOpts func(opts *api.BalanceTransactionsListOptions)
}

func (m *balanceTestClientWithCapture) ListBalanceTransactions(ctx context.Context, opts *api.BalanceTransactionsListOptions) (*api.BalanceTransactionsListResponse, error) {
	if m.captureOpts != nil {
		m.captureOpts(opts)
	}
	return m.balanceTestClient.ListBalanceTransactions(ctx, opts)
}
