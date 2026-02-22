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

// mockBulkOperationsClient is a mock implementation for bulk operations testing.
type mockBulkOperationsClient struct {
	api.MockClient

	listBulkOperationsResp *api.BulkOperationsListResponse
	listBulkOperationsErr  error

	getBulkOperationResp *api.BulkOperation
	getBulkOperationErr  error

	getCurrentBulkOperationResp *api.BulkOperation
	getCurrentBulkOperationErr  error

	createBulkQueryResp *api.BulkOperation
	createBulkQueryErr  error

	cancelBulkOperationResp *api.BulkOperation
	cancelBulkOperationErr  error
}

func (m *mockBulkOperationsClient) ListBulkOperations(ctx context.Context, opts *api.BulkOperationsListOptions) (*api.BulkOperationsListResponse, error) {
	return m.listBulkOperationsResp, m.listBulkOperationsErr
}

func (m *mockBulkOperationsClient) GetBulkOperation(ctx context.Context, id string) (*api.BulkOperation, error) {
	return m.getBulkOperationResp, m.getBulkOperationErr
}

func (m *mockBulkOperationsClient) GetCurrentBulkOperation(ctx context.Context) (*api.BulkOperation, error) {
	return m.getCurrentBulkOperationResp, m.getCurrentBulkOperationErr
}

func (m *mockBulkOperationsClient) CreateBulkQuery(ctx context.Context, req *api.BulkOperationCreateRequest) (*api.BulkOperation, error) {
	return m.createBulkQueryResp, m.createBulkQueryErr
}

func (m *mockBulkOperationsClient) CancelBulkOperation(ctx context.Context, id string) (*api.BulkOperation, error) {
	return m.cancelBulkOperationResp, m.cancelBulkOperationErr
}

// setupBulkOperationsTest sets up the test environment with mock factories.
func setupBulkOperationsTest(t *testing.T, mockClient *mockBulkOperationsClient) (cleanup func(), buf *bytes.Buffer) {
	t.Helper()

	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	buf = new(bytes.Buffer)

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

	formatterWriter = buf

	cleanup = func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	return cleanup, buf
}

// newBulkOperationsTestCmd creates a test command with standard flags.
func newBulkOperationsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("graphql", "", "")
	return cmd
}

func TestBulkOperationsCommand(t *testing.T) {
	if bulkOperationsCmd == nil {
		t.Fatal("bulkOperationsCmd is nil")
	}
	if bulkOperationsCmd.Use != "bulk-operations" {
		t.Errorf("Expected Use to be 'bulk-operations', got %q", bulkOperationsCmd.Use)
	}
}

func TestBulkOperationsSubcommands(t *testing.T) {
	subcommands := bulkOperationsCmd.Commands()
	expectedCmds := map[string]bool{"list": false, "get": false, "current": false, "query": false, "cancel": false}
	for _, cmd := range subcommands {
		switch cmd.Use {
		case "list":
			expectedCmds["list"] = true
		case "get <id>":
			expectedCmds["get"] = true
		case "current":
			expectedCmds["current"] = true
		case "query":
			expectedCmds["query"] = true
		case "cancel <id>":
			expectedCmds["cancel"] = true
		}
	}
	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestBulkOperationsListFlags(t *testing.T) {
	flags := []string{"status", "type", "page", "page-size"}
	for _, flag := range flags {
		if bulkOperationsListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on list command", flag)
		}
	}
}

func TestBulkOperationsQueryFlags(t *testing.T) {
	flags := []string{"graphql"}
	for _, flag := range flags {
		if bulkOperationsQueryCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on query command", flag)
		}
	}
}

func TestBulkOperationsGetArgsValidation(t *testing.T) {
	if bulkOperationsGetCmd.Args == nil {
		t.Fatal("Expected Args validator on get command")
	}
	err := bulkOperationsGetCmd.Args(bulkOperationsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = bulkOperationsGetCmd.Args(bulkOperationsGetCmd, []string{"op_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestBulkOperationsCancelArgsValidation(t *testing.T) {
	if bulkOperationsCancelCmd.Args == nil {
		t.Fatal("Expected Args validator on cancel command")
	}
	err := bulkOperationsCancelCmd.Args(bulkOperationsCancelCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = bulkOperationsCancelCmd.Args(bulkOperationsCancelCmd, []string{"op_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestBulkOperationsListRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		mockResp     *api.BulkOperationsListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantErrMsg   string
		wantOutput   []string
	}{
		{
			name: "successful list text format",
			mockResp: &api.BulkOperationsListResponse{
				Items: []api.BulkOperation{
					{
						ID:          "op_123",
						Type:        "query",
						Status:      "completed",
						ObjectCount: 1000,
						FileSize:    2048,
						CreatedAt:   createdAt,
					},
					{
						ID:          "op_456",
						Type:        "mutation",
						Status:      "running",
						ObjectCount: 500,
						FileSize:    1024,
						CreatedAt:   createdAt,
					},
				},
				TotalCount: 2,
			},
			wantOutput: []string{"op_123", "op_456", "query", "mutation", "completed", "running"},
		},
		{
			name: "successful list json format",
			mockResp: &api.BulkOperationsListResponse{
				Items: []api.BulkOperation{
					{
						ID:          "op_789",
						Type:        "query",
						Status:      "completed",
						ObjectCount: 100,
						FileSize:    512,
						CreatedAt:   createdAt,
					},
				},
				TotalCount: 1,
			},
			outputFormat: "json",
			wantOutput:   []string{"op_789"},
		},
		{
			name: "empty list",
			mockResp: &api.BulkOperationsListResponse{
				Items:      []api.BulkOperation{},
				TotalCount: 0,
			},
			wantOutput: []string{},
		},
		{
			name:       "API error",
			mockErr:    errors.New("API unavailable"),
			wantErr:    true,
			wantErrMsg: "failed to list bulk operations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBulkOperationsClient{
				listBulkOperationsResp: tt.mockResp,
				listBulkOperationsErr:  tt.mockErr,
			}

			cleanup, buf := setupBulkOperationsTest(t, mockClient)
			defer cleanup()

			cmd := newBulkOperationsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := bulkOperationsListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got %q", want, output)
				}
			}
		})
	}
}

func TestBulkOperationsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := bulkOperationsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestBulkOperationsGetRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	completedAt := time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		operationID  string
		mockResp     *api.BulkOperation
		mockErr      error
		outputFormat string
		wantErr      bool
		wantErrMsg   string
		wantOutput   []string // Only checked for JSON output
	}{
		{
			name:        "successful get text format",
			operationID: "op_123",
			mockResp: &api.BulkOperation{
				ID:          "op_123",
				Type:        "query",
				Status:      "completed",
				ObjectCount: 1000,
				FileSize:    2048,
				URL:         "https://example.com/results.jsonl",
				CreatedAt:   createdAt,
				CompletedAt: &completedAt,
			},
		},
		{
			name:        "successful get with error code",
			operationID: "op_err",
			mockResp: &api.BulkOperation{
				ID:          "op_err",
				Type:        "query",
				Status:      "failed",
				ObjectCount: 0,
				FileSize:    0,
				ErrorCode:   "INTERNAL_ERROR",
				CreatedAt:   createdAt,
			},
		},
		{
			name:        "successful get with partial data URL",
			operationID: "op_partial",
			mockResp: &api.BulkOperation{
				ID:             "op_partial",
				Type:           "query",
				Status:         "failed",
				ObjectCount:    500,
				FileSize:       1024,
				PartialDataURL: "https://example.com/partial.jsonl",
				CreatedAt:      createdAt,
			},
		},
		{
			name:         "successful get json format",
			operationID:  "op_456",
			outputFormat: "json",
			mockResp: &api.BulkOperation{
				ID:          "op_456",
				Type:        "mutation",
				Status:      "running",
				ObjectCount: 500,
				FileSize:    1024,
				CreatedAt:   createdAt,
			},
			wantOutput: []string{"op_456", "mutation"},
		},
		{
			name:        "operation not found",
			operationID: "op_999",
			mockErr:     errors.New("not found"),
			wantErr:     true,
			wantErrMsg:  "failed to get bulk operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBulkOperationsClient{
				getBulkOperationResp: tt.mockResp,
				getBulkOperationErr:  tt.mockErr,
			}

			cleanup, buf := setupBulkOperationsTest(t, mockClient)
			defer cleanup()

			cmd := newBulkOperationsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := bulkOperationsGetCmd.RunE(cmd, []string{tt.operationID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check output for JSON format (text uses fmt.Printf to stdout)
			if tt.outputFormat == "json" {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got %q", want, output)
					}
				}
			}
		})
	}
}

func TestBulkOperationsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := bulkOperationsGetCmd.RunE(cmd, []string{"op_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestBulkOperationsCurrentRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		mockResp     *api.BulkOperation
		mockErr      error
		outputFormat string
		wantErr      bool
		wantErrMsg   string
		wantOutput   []string // Only checked for JSON output
	}{
		{
			name: "successful current text format",
			mockResp: &api.BulkOperation{
				ID:          "op_current",
				Type:        "query",
				Status:      "running",
				ObjectCount: 500,
				CreatedAt:   createdAt,
			},
		},
		{
			name: "no current operation",
			mockResp: &api.BulkOperation{
				ID: "", // Empty ID indicates no running operation
			},
		},
		{
			name:         "successful current json format",
			outputFormat: "json",
			mockResp: &api.BulkOperation{
				ID:          "op_json",
				Type:        "mutation",
				Status:      "running",
				ObjectCount: 100,
				CreatedAt:   createdAt,
			},
			wantOutput: []string{"op_json"},
		},
		{
			name:       "API error",
			mockErr:    errors.New("API unavailable"),
			wantErr:    true,
			wantErrMsg: "failed to get current bulk operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBulkOperationsClient{
				getCurrentBulkOperationResp: tt.mockResp,
				getCurrentBulkOperationErr:  tt.mockErr,
			}

			cleanup, buf := setupBulkOperationsTest(t, mockClient)
			defer cleanup()

			cmd := newBulkOperationsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := bulkOperationsCurrentCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check output for JSON format (text uses fmt.Printf to stdout)
			if tt.outputFormat == "json" {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got %q", want, output)
					}
				}
			}
		})
	}
}

func TestBulkOperationsCurrentGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := bulkOperationsCurrentCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestBulkOperationsQueryRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		graphqlQuery string
		mockResp     *api.BulkOperation
		mockErr      error
		outputFormat string
		wantErr      bool
		wantErrMsg   string
		wantOutput   []string // Only checked for JSON output
	}{
		{
			name:         "successful query text format",
			graphqlQuery: "{ products { edges { node { id } } } }",
			mockResp: &api.BulkOperation{
				ID:        "op_query_123",
				Type:      "query",
				Status:    "created",
				CreatedAt: createdAt,
			},
		},
		{
			name:         "successful query json format",
			graphqlQuery: "{ orders { edges { node { id } } } }",
			outputFormat: "json",
			mockResp: &api.BulkOperation{
				ID:        "op_query_456",
				Type:      "query",
				Status:    "running",
				CreatedAt: createdAt,
			},
			wantOutput: []string{"op_query_456"},
		},
		{
			name:         "query creation error",
			graphqlQuery: "{ invalid query }",
			mockErr:      errors.New("invalid query"),
			wantErr:      true,
			wantErrMsg:   "failed to create bulk query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBulkOperationsClient{
				createBulkQueryResp: tt.mockResp,
				createBulkQueryErr:  tt.mockErr,
			}

			cleanup, buf := setupBulkOperationsTest(t, mockClient)
			defer cleanup()

			cmd := newBulkOperationsTestCmd()
			_ = cmd.Flags().Set("graphql", tt.graphqlQuery)
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := bulkOperationsQueryCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check output for JSON format (text uses fmt.Printf to stdout)
			if tt.outputFormat == "json" {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got %q", want, output)
					}
				}
			}
		})
	}
}

func TestBulkOperationsQueryGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := bulkOperationsQueryCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestBulkOperationsQueryWithClient(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{"test": {Handle: "test", AccessToken: "token"}},
		}, nil
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("graphql", "", "GraphQL query")
	_ = cmd.Flags().Set("graphql", "{ products { edges { node { id } } } }")
	err := bulkOperationsQueryCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("Query succeeded unexpectedly")
	}
}

func TestBulkOperationsCancelRunE(t *testing.T) {
	tests := []struct {
		name         string
		operationID  string
		mockResp     *api.BulkOperation
		mockErr      error
		outputFormat string
		wantErr      bool
		wantErrMsg   string
		wantOutput   []string // Only checked for JSON output
	}{
		{
			name:        "successful cancel text format",
			operationID: "op_to_cancel",
			mockResp: &api.BulkOperation{
				ID:     "op_to_cancel",
				Status: "cancelled",
			},
		},
		{
			name:         "successful cancel json format",
			operationID:  "op_cancel_json",
			outputFormat: "json",
			mockResp: &api.BulkOperation{
				ID:     "op_cancel_json",
				Status: "cancelled",
			},
			wantOutput: []string{"op_cancel_json"},
		},
		{
			name:        "cancel error - already completed",
			operationID: "op_completed",
			mockErr:     errors.New("operation already completed"),
			wantErr:     true,
			wantErrMsg:  "failed to cancel bulk operation",
		},
		{
			name:        "cancel error - not found",
			operationID: "op_not_found",
			mockErr:     errors.New("not found"),
			wantErr:     true,
			wantErrMsg:  "failed to cancel bulk operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBulkOperationsClient{
				cancelBulkOperationResp: tt.mockResp,
				cancelBulkOperationErr:  tt.mockErr,
			}

			cleanup, buf := setupBulkOperationsTest(t, mockClient)
			defer cleanup()

			cmd := newBulkOperationsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := bulkOperationsCancelCmd.RunE(cmd, []string{tt.operationID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check output for JSON format (text uses fmt.Printf to stdout)
			if tt.outputFormat == "json" {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got %q", want, output)
					}
				}
			}
		})
	}
}

func TestBulkOperationsCancelGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := bulkOperationsCancelCmd.RunE(cmd, []string{"op_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "bytes below unit",
			bytes:    500,
			expected: "500 B",
		},
		{
			name:     "exactly 1 KB",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "1.5 KB",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "exactly 1 MB",
			bytes:    1048576,
			expected: "1.0 MB",
		},
		{
			name:     "exactly 1 GB",
			bytes:    1073741824,
			expected: "1.0 GB",
		},
		{
			name:     "exactly 1 TB",
			bytes:    1099511627776,
			expected: "1.0 TB",
		},
		{
			name:     "2.5 MB",
			bytes:    2621440,
			expected: "2.5 MB",
		},
		{
			name:     "exactly 1 PB",
			bytes:    1125899906842624,
			expected: "1.0 PB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestBulkOperationsListWithFilters(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockBulkOperationsClient{
		listBulkOperationsResp: &api.BulkOperationsListResponse{
			Items: []api.BulkOperation{
				{
					ID:          "op_filtered",
					Type:        "query",
					Status:      "completed",
					ObjectCount: 100,
					FileSize:    512,
					CreatedAt:   createdAt,
				},
			},
			TotalCount: 1,
		},
	}

	cleanup, buf := setupBulkOperationsTest(t, mockClient)
	defer cleanup()

	cmd := newBulkOperationsTestCmd()
	_ = cmd.Flags().Set("status", "completed")
	_ = cmd.Flags().Set("type", "query")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "50")

	err := bulkOperationsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "op_filtered") {
		t.Errorf("output should contain 'op_filtered', got %q", output)
	}
}

func TestBulkOperationsGetWithOptionalFields(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	completedAt := time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		mockResp *api.BulkOperation
	}{
		{
			name: "operation with all optional fields",
			mockResp: &api.BulkOperation{
				ID:             "op_full",
				Type:           "query",
				Status:         "completed",
				ObjectCount:    1000,
				FileSize:       2048,
				URL:            "https://example.com/results.jsonl",
				ErrorCode:      "",
				PartialDataURL: "",
				CreatedAt:      createdAt,
				CompletedAt:    &completedAt,
			},
		},
		{
			name: "operation without completed time",
			mockResp: &api.BulkOperation{
				ID:          "op_running",
				Type:        "query",
				Status:      "running",
				ObjectCount: 500,
				FileSize:    1024,
				CreatedAt:   createdAt,
				CompletedAt: nil,
			},
		},
		{
			name: "operation with error and partial data",
			mockResp: &api.BulkOperation{
				ID:             "op_error",
				Type:           "query",
				Status:         "failed",
				ObjectCount:    100,
				FileSize:       256,
				ErrorCode:      "TIMEOUT",
				PartialDataURL: "https://example.com/partial.jsonl",
				CreatedAt:      createdAt,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBulkOperationsClient{
				getBulkOperationResp: tt.mockResp,
			}

			cleanup, _ := setupBulkOperationsTest(t, mockClient)
			defer cleanup()

			cmd := newBulkOperationsTestCmd()
			err := bulkOperationsGetCmd.RunE(cmd, []string{tt.mockResp.ID})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// Text output goes to stdout via fmt.Printf - we just verify no error occurred
		})
	}
}

func TestBulkOperationsListEmptyResponse(t *testing.T) {
	mockClient := &mockBulkOperationsClient{
		listBulkOperationsResp: &api.BulkOperationsListResponse{
			Items:      []api.BulkOperation{},
			TotalCount: 0,
		},
	}

	cleanup, _ := setupBulkOperationsTest(t, mockClient)
	defer cleanup()

	cmd := newBulkOperationsTestCmd()
	err := bulkOperationsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBulkOperationsListLargeFileSize(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockBulkOperationsClient{
		listBulkOperationsResp: &api.BulkOperationsListResponse{
			Items: []api.BulkOperation{
				{
					ID:          "op_large",
					Type:        "query",
					Status:      "completed",
					ObjectCount: 1000000,
					FileSize:    1073741824, // 1 GB
					CreatedAt:   createdAt,
				},
			},
			TotalCount: 1,
		},
	}

	cleanup, buf := setupBulkOperationsTest(t, mockClient)
	defer cleanup()

	cmd := newBulkOperationsTestCmd()
	err := bulkOperationsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "1.0 GB") {
		t.Errorf("output should contain '1.0 GB', got %q", output)
	}
}
