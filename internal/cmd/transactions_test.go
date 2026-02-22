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

// mockTransactionsClient is a mock implementation for transactions API methods.
type mockTransactionsClient struct {
	api.MockClient // embed base mock for unimplemented methods

	listTransactionsResp      *api.TransactionsListResponse
	listTransactionsErr       error
	getTransactionResp        *api.Transaction
	getTransactionErr         error
	listOrderTransactionsResp *api.TransactionsListResponse
	listOrderTransactionsErr  error
}

func (m *mockTransactionsClient) ListTransactions(ctx context.Context, opts *api.TransactionsListOptions) (*api.TransactionsListResponse, error) {
	return m.listTransactionsResp, m.listTransactionsErr
}

func (m *mockTransactionsClient) GetTransaction(ctx context.Context, id string) (*api.Transaction, error) {
	return m.getTransactionResp, m.getTransactionErr
}

func (m *mockTransactionsClient) ListOrderTransactions(ctx context.Context, orderID string) (*api.TransactionsListResponse, error) {
	return m.listOrderTransactionsResp, m.listOrderTransactionsErr
}

// setupTransactionsTest sets up the test environment for transactions tests.
func setupTransactionsTest(t *testing.T, mockClient *mockTransactionsClient) (*bytes.Buffer, func()) {
	t.Helper()

	// Save original factories
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	// Setup mock credential store
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	// Setup mock API client
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	// Capture output
	buf := new(bytes.Buffer)
	formatterWriter = buf

	// Return cleanup function
	cleanup := func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	return buf, cleanup
}

// TestTransactionsCommandSetup verifies transactions command initialization
func TestTransactionsCommandSetup(t *testing.T) {
	if transactionsCmd.Use != "transactions" {
		t.Errorf("expected Use 'transactions', got %q", transactionsCmd.Use)
	}
	if transactionsCmd.Short != "Manage payment transactions" {
		t.Errorf("expected Short 'Manage payment transactions', got %q", transactionsCmd.Short)
	}
}

// TestTransactionsSubcommands verifies all subcommands are registered
func TestTransactionsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":  "List transactions",
		"get":   "Get transaction details",
		"order": "List transactions for an order",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range transactionsCmd.Commands() {
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

// TestTransactionsListFlags verifies list command flags exist with correct defaults
func TestTransactionsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
		{"kind", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := transactionsListCmd.Flags().Lookup(f.name)
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

// TestTransactionsGetClientError verifies error handling when getClient fails
func TestTransactionsGetClientError(t *testing.T) {
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

// TestTransactionsWithMockStore tests transactions commands with a mock credential store
func TestTransactionsWithMockStore(t *testing.T) {
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

// TestTransactionsGetArgs verifies get command requires exactly one argument
func TestTransactionsGetArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if transactionsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", transactionsGetCmd.Use)
	}
}

// TestTransactionsOrderArgs verifies order command requires exactly one argument
func TestTransactionsOrderArgs(t *testing.T) {
	// Check the Use field includes <order-id> which indicates required argument
	if transactionsOrderCmd.Use != "order <order-id>" {
		t.Errorf("expected Use 'order <order-id>', got %q", transactionsOrderCmd.Use)
	}
}

// TestTransactionsListFlagDescriptions verifies flag descriptions are set
func TestTransactionsListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"page":      "Page number",
		"page-size": "Results per page",
		"status":    "Filter by status (success, failure, pending)",
		"kind":      "Filter by kind (sale, refund, capture, void)",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := transactionsListCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Usage != expectedUsage {
				t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
			}
		})
	}
}

// TestTransactionsListFlagTypes verifies flag types are correct
func TestTransactionsListFlagTypes(t *testing.T) {
	flags := map[string]string{
		"page":      "int",
		"page-size": "int",
		"status":    "string",
		"kind":      "string",
	}

	for flagName, expectedType := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := transactionsListCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Value.Type() != expectedType {
				t.Errorf("expected type %q, got %q", expectedType, flag.Value.Type())
			}
		})
	}
}

// TestTransactionsOrderShort verifies order subcommand description
func TestTransactionsOrderShort(t *testing.T) {
	expectedShort := "List transactions for an order"
	if transactionsOrderCmd.Short != expectedShort {
		t.Errorf("expected Short %q, got %q", expectedShort, transactionsOrderCmd.Short)
	}
}

// TestTransactionsListRunE tests the transactions list command execution.
func TestTransactionsListRunE(t *testing.T) {
	tests := []struct {
		name         string
		mockResp     *api.TransactionsListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantErrMsg   string
		wantOutput   string
	}{
		{
			name: "successful list text output",
			mockResp: &api.TransactionsListResponse{
				Items: []api.Transaction{
					{
						ID:        "txn_123",
						OrderID:   "ord_456",
						Kind:      "sale",
						Status:    "success",
						Amount:    "99.99",
						Currency:  "USD",
						Gateway:   "stripe",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:        "txn_789",
						OrderID:   "ord_101",
						Kind:      "refund",
						Status:    "pending",
						Amount:    "25.00",
						Currency:  "EUR",
						Gateway:   "paypal",
						CreatedAt: time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			outputFormat: "text",
			wantOutput:   "txn_123",
		},
		{
			name: "successful list json output",
			mockResp: &api.TransactionsListResponse{
				Items: []api.Transaction{
					{
						ID:        "txn_123",
						OrderID:   "ord_456",
						Kind:      "sale",
						Status:    "success",
						Amount:    "99.99",
						Currency:  "USD",
						Gateway:   "stripe",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFormat: "json",
			wantOutput:   `"id": "txn_123"`,
		},
		{
			name: "empty list",
			mockResp: &api.TransactionsListResponse{
				Items:      []api.Transaction{},
				TotalCount: 0,
			},
			outputFormat: "text",
		},
		{
			name:         "API error",
			mockErr:      errors.New("API unavailable"),
			outputFormat: "text",
			wantErr:      true,
			wantErrMsg:   "failed to list transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockTransactionsClient{
				listTransactionsResp: tt.mockResp,
				listTransactionsErr:  tt.mockErr,
			}

			buf, cleanup := setupTransactionsTest(t, mockClient)
			defer cleanup()

			// Create command with flags
			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("kind", "", "")

			err := transactionsListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
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

// TestTransactionsListRunEWithFilters tests list command with filter flags.
func TestTransactionsListRunEWithFilters(t *testing.T) {
	mockClient := &mockTransactionsClient{
		listTransactionsResp: &api.TransactionsListResponse{
			Items: []api.Transaction{
				{
					ID:        "txn_filtered",
					OrderID:   "ord_111",
					Kind:      "refund",
					Status:    "success",
					Amount:    "50.00",
					Currency:  "USD",
					Gateway:   "stripe",
					CreatedAt: time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}

	buf, cleanup := setupTransactionsTest(t, mockClient)
	defer cleanup()

	// Create command with filter flags set
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 10, "")
	cmd.Flags().String("status", "success", "")
	cmd.Flags().String("kind", "refund", "")

	err := transactionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "txn_filtered") {
		t.Errorf("output should contain filtered transaction ID")
	}
}

// TestTransactionsListRunEGetClientError tests list command when getClient fails.
func TestTransactionsListRunEGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("kind", "", "")

	err := transactionsListCmd.RunE(cmd, []string{})

	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestTransactionsGetRunE tests the transactions get command execution.
func TestTransactionsGetRunE(t *testing.T) {
	tests := []struct {
		name          string
		transactionID string
		mockResp      *api.Transaction
		mockErr       error
		outputFormat  string
		wantErr       bool
		wantErrMsg    string
		wantOutput    string
	}{
		{
			name:          "successful get text output",
			transactionID: "txn_123",
			mockResp: &api.Transaction{
				ID:        "txn_123",
				OrderID:   "ord_456",
				Kind:      "sale",
				Status:    "success",
				Amount:    "99.99",
				Currency:  "USD",
				Gateway:   "stripe",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			outputFormat: "text",
			wantOutput:   "", // Text output goes to stdout, not captured in buffer
		},
		{
			name:          "successful get json output",
			transactionID: "txn_123",
			mockResp: &api.Transaction{
				ID:        "txn_123",
				OrderID:   "ord_456",
				Kind:      "sale",
				Status:    "success",
				Amount:    "99.99",
				Currency:  "USD",
				Gateway:   "stripe",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			outputFormat: "json",
			wantOutput:   `"id": "txn_123"`,
		},
		{
			name:          "get with error code",
			transactionID: "txn_err",
			mockResp: &api.Transaction{
				ID:        "txn_err",
				OrderID:   "ord_789",
				Kind:      "sale",
				Status:    "failure",
				Amount:    "50.00",
				Currency:  "USD",
				Gateway:   "stripe",
				ErrorCode: "card_declined",
				CreatedAt: time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
			outputFormat: "text",
			wantOutput:   "", // Text output goes to stdout, not captured in buffer
		},
		{
			name:          "get with message",
			transactionID: "txn_msg",
			mockResp: &api.Transaction{
				ID:        "txn_msg",
				OrderID:   "ord_999",
				Kind:      "refund",
				Status:    "success",
				Amount:    "25.00",
				Currency:  "EUR",
				Gateway:   "paypal",
				Message:   "Refund processed successfully",
				CreatedAt: time.Date(2024, 1, 17, 12, 30, 0, 0, time.UTC),
			},
			outputFormat: "text",
			wantOutput:   "", // Text output goes to stdout, not captured in buffer
		},
		{
			name:          "get with both error code and message",
			transactionID: "txn_full",
			mockResp: &api.Transaction{
				ID:        "txn_full",
				OrderID:   "ord_full",
				Kind:      "capture",
				Status:    "failure",
				Amount:    "100.00",
				Currency:  "GBP",
				Gateway:   "adyen",
				ErrorCode: "insufficient_funds",
				Message:   "Card has insufficient funds",
				CreatedAt: time.Date(2024, 1, 18, 8, 15, 0, 0, time.UTC),
			},
			outputFormat: "text",
			wantOutput:   "", // Text output goes to stdout, not captured in buffer
		},
		{
			name:          "transaction not found",
			transactionID: "txn_notfound",
			mockErr:       errors.New("transaction not found"),
			outputFormat:  "text",
			wantErr:       true,
			wantErrMsg:    "failed to get transaction",
		},
		{
			name:          "API error",
			transactionID: "txn_error",
			mockErr:       errors.New("API unavailable"),
			outputFormat:  "text",
			wantErr:       true,
			wantErrMsg:    "failed to get transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockTransactionsClient{
				getTransactionResp: tt.mockResp,
				getTransactionErr:  tt.mockErr,
			}

			buf, cleanup := setupTransactionsTest(t, mockClient)
			defer cleanup()

			// Create command with flags
			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := transactionsGetCmd.RunE(cmd, []string{tt.transactionID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
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

// TestTransactionsGetRunEGetClientError tests get command when getClient fails.
func TestTransactionsGetRunEGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := transactionsGetCmd.RunE(cmd, []string{"txn_123"})

	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestTransactionsOrderRunE tests the transactions order command execution.
func TestTransactionsOrderRunE(t *testing.T) {
	tests := []struct {
		name         string
		orderID      string
		mockResp     *api.TransactionsListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantErrMsg   string
		wantOutput   string
	}{
		{
			name:    "successful order transactions text output",
			orderID: "ord_456",
			mockResp: &api.TransactionsListResponse{
				Items: []api.Transaction{
					{
						ID:        "txn_111",
						OrderID:   "ord_456",
						Kind:      "sale",
						Status:    "success",
						Amount:    "150.00",
						Currency:  "USD",
						Gateway:   "stripe",
						CreatedAt: time.Date(2024, 1, 20, 15, 0, 0, 0, time.UTC),
					},
					{
						ID:        "txn_222",
						OrderID:   "ord_456",
						Kind:      "refund",
						Status:    "success",
						Amount:    "30.00",
						Currency:  "USD",
						Gateway:   "stripe",
						CreatedAt: time.Date(2024, 1, 21, 10, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			outputFormat: "text",
			wantOutput:   "txn_111",
		},
		{
			name:    "successful order transactions json output",
			orderID: "ord_789",
			mockResp: &api.TransactionsListResponse{
				Items: []api.Transaction{
					{
						ID:        "txn_333",
						OrderID:   "ord_789",
						Kind:      "capture",
						Status:    "success",
						Amount:    "75.00",
						Currency:  "EUR",
						Gateway:   "paypal",
						CreatedAt: time.Date(2024, 1, 22, 9, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFormat: "json",
			wantOutput:   `"id": "txn_333"`,
		},
		{
			name:    "order with no transactions",
			orderID: "ord_empty",
			mockResp: &api.TransactionsListResponse{
				Items:      []api.Transaction{},
				TotalCount: 0,
			},
			outputFormat: "text",
		},
		{
			name:         "order not found",
			orderID:      "ord_notfound",
			mockErr:      errors.New("order not found"),
			outputFormat: "text",
			wantErr:      true,
			wantErrMsg:   "failed to list order transactions",
		},
		{
			name:         "API error",
			orderID:      "ord_error",
			mockErr:      errors.New("API unavailable"),
			outputFormat: "text",
			wantErr:      true,
			wantErrMsg:   "failed to list order transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockTransactionsClient{
				listOrderTransactionsResp: tt.mockResp,
				listOrderTransactionsErr:  tt.mockErr,
			}

			buf, cleanup := setupTransactionsTest(t, mockClient)
			defer cleanup()

			// Create command with flags
			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := transactionsOrderCmd.RunE(cmd, []string{tt.orderID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
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

// TestTransactionsOrderRunEGetClientError tests order command when getClient fails.
func TestTransactionsOrderRunEGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := transactionsOrderCmd.RunE(cmd, []string{"ord_123"})

	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestTransactionsGetCmdArgs verifies the get command argument validation.
func TestTransactionsGetCmdArgs(t *testing.T) {
	// Test with no args - should fail
	if err := transactionsGetCmd.Args(transactionsGetCmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}

	// Test with one arg - should succeed
	if err := transactionsGetCmd.Args(transactionsGetCmd, []string{"txn_123"}); err != nil {
		t.Errorf("expected no error with one arg, got: %v", err)
	}

	// Test with two args - should fail
	if err := transactionsGetCmd.Args(transactionsGetCmd, []string{"txn_123", "extra"}); err == nil {
		t.Error("expected error with two args")
	}
}

// TestTransactionsOrderCmdArgs verifies the order command argument validation.
func TestTransactionsOrderCmdArgs(t *testing.T) {
	// Test with no args - should fail
	if err := transactionsOrderCmd.Args(transactionsOrderCmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}

	// Test with one arg - should succeed
	if err := transactionsOrderCmd.Args(transactionsOrderCmd, []string{"ord_123"}); err != nil {
		t.Errorf("expected no error with one arg, got: %v", err)
	}

	// Test with two args - should fail
	if err := transactionsOrderCmd.Args(transactionsOrderCmd, []string{"ord_123", "extra"}); err == nil {
		t.Error("expected error with two args")
	}
}

// TestTransactionsListOutputFormats tests both text and JSON output formats for list.
func TestTransactionsListOutputFormats(t *testing.T) {
	testTransaction := api.Transaction{
		ID:        "txn_format_test",
		OrderID:   "ord_format_test",
		Kind:      "sale",
		Status:    "success",
		Amount:    "123.45",
		Currency:  "USD",
		Gateway:   "stripe",
		CreatedAt: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name         string
		outputFormat string
		wantContains []string
	}{
		{
			name:         "text format contains table data",
			outputFormat: "text",
			wantContains: []string{"txn_format_test", "ord_format_test", "sale", "success"},
		},
		{
			name:         "json format contains JSON structure",
			outputFormat: "json",
			wantContains: []string{`"id"`, `"order_id"`, `"kind"`, `"status"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockTransactionsClient{
				listTransactionsResp: &api.TransactionsListResponse{
					Items:      []api.Transaction{testTransaction},
					TotalCount: 1,
				},
			}

			buf, cleanup := setupTransactionsTest(t, mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("kind", "", "")

			err := transactionsListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

// TestTransactionsGetTextOutputFields tests that the get command executes successfully with all fields.
func TestTransactionsGetTextOutputFields(t *testing.T) {
	mockClient := &mockTransactionsClient{
		getTransactionResp: &api.Transaction{
			ID:        "txn_field_test",
			OrderID:   "ord_field_test",
			Kind:      "capture",
			Status:    "failure",
			Amount:    "999.99",
			Currency:  "GBP",
			Gateway:   "adyen",
			ErrorCode: "declined",
			Message:   "Transaction declined by issuer",
			CreatedAt: time.Date(2024, 4, 15, 16, 45, 0, 0, time.UTC),
		},
	}

	_, cleanup := setupTransactionsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	// Execute the command - text output goes to stdout via fmt.Printf
	// This test verifies the command executes without error when all fields are present
	err := transactionsGetCmd.RunE(cmd, []string{"txn_field_test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestTransactionsOrderTextOutputColumns tests that order transactions table has correct columns.
func TestTransactionsOrderTextOutputColumns(t *testing.T) {
	mockClient := &mockTransactionsClient{
		listOrderTransactionsResp: &api.TransactionsListResponse{
			Items: []api.Transaction{
				{
					ID:        "txn_col_test",
					OrderID:   "ord_col_test",
					Kind:      "void",
					Status:    "pending",
					Amount:    "50.00",
					Currency:  "CAD",
					Gateway:   "square",
					CreatedAt: time.Date(2024, 5, 1, 8, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}

	buf, cleanup := setupTransactionsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := transactionsOrderCmd.RunE(cmd, []string{"ord_col_test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that transaction data is present (note: order transactions table omits ORDER column)
	expectedData := []string{"txn_col_test", "void", "pending", "50.00", "CAD", "square"}
	for _, data := range expectedData {
		if !strings.Contains(output, data) {
			t.Errorf("output should contain %q", data)
		}
	}
}

// TestTransactionsGetNoOptionalFields tests get command when optional fields are empty.
func TestTransactionsGetNoOptionalFields(t *testing.T) {
	mockClient := &mockTransactionsClient{
		getTransactionResp: &api.Transaction{
			ID:        "txn_minimal",
			OrderID:   "ord_minimal",
			Kind:      "sale",
			Status:    "success",
			Amount:    "10.00",
			Currency:  "USD",
			Gateway:   "stripe",
			ErrorCode: "", // No error code
			Message:   "", // No message
			CreatedAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
		},
	}

	_, cleanup := setupTransactionsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	// Execute the command - text output goes to stdout via fmt.Printf
	// This test verifies the command executes without error when optional fields are empty
	err := transactionsGetCmd.RunE(cmd, []string{"txn_minimal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
