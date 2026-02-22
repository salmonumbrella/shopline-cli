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

// operationLogsAPIClient is a mock implementation of api.APIClient for operation logs tests.
type operationLogsAPIClient struct {
	api.MockClient

	listOperationLogsResp *api.OperationLogsListResponse
	listOperationLogsErr  error
	getOperationLogResp   *api.OperationLog
	getOperationLogErr    error
}

func (m *operationLogsAPIClient) ListOperationLogs(ctx context.Context, opts *api.OperationLogsListOptions) (*api.OperationLogsListResponse, error) {
	return m.listOperationLogsResp, m.listOperationLogsErr
}

func (m *operationLogsAPIClient) GetOperationLog(ctx context.Context, id string) (*api.OperationLog, error) {
	return m.getOperationLogResp, m.getOperationLogErr
}

// setupOperationLogsTest sets up the test environment for operation logs tests.
func setupOperationLogsTest(t *testing.T) (restore func()) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	origEnv := os.Getenv("SHOPLINE_STORE")

	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}
}

// createOperationLogsTestCmd creates a cobra.Command with all necessary flags for testing.
func createOperationLogsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "Store profile name")
	cmd.Flags().String("output", "", "Output format")
	cmd.Flags().String("color", "never", "Color mode")
	cmd.Flags().String("query", "", "JQ filter")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("page-size", 20, "Results per page")
	cmd.Flags().String("action", "", "Action filter")
	cmd.Flags().String("resource-type", "", "Resource type filter")
	cmd.Flags().String("resource-id", "", "Resource ID filter")
	cmd.Flags().String("user-id", "", "User ID filter")
	cmd.Flags().String("since", "", "Since date filter")
	cmd.Flags().String("until", "", "Until date filter")
	return cmd
}

// TestOperationLogsCommand verifies parent command initialization.
func TestOperationLogsCommand(t *testing.T) {
	if operationLogsCmd == nil {
		t.Fatal("operationLogsCmd is nil")
	}
	if operationLogsCmd.Use != "operation-logs" {
		t.Errorf("Expected Use to be 'operation-logs', got %q", operationLogsCmd.Use)
	}
	if operationLogsCmd.Short != "View operation audit logs" {
		t.Errorf("Expected Short to be 'View operation audit logs', got %q", operationLogsCmd.Short)
	}
}

// TestOperationLogsAliases verifies command aliases.
func TestOperationLogsAliases(t *testing.T) {
	expectedAliases := []string{"audit-logs", "audit", "logs", "operation-log", "ol"}
	aliases := operationLogsCmd.Aliases
	if len(aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(aliases))
	}
	// Check that all expected aliases are present (order-agnostic)
	for _, expected := range expectedAliases {
		found := false
		for _, actual := range aliases {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected alias %q not found in %v", expected, aliases)
		}
	}
}

// TestOperationLogsSubcommands verifies all subcommands are registered.
func TestOperationLogsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":     "List operation logs",
		"get <id>": "Get operation log details",
	}

	for use, short := range subcommands {
		t.Run(use, func(t *testing.T) {
			found := false
			for _, sub := range operationLogsCmd.Commands() {
				if sub.Use == use {
					found = true
					if sub.Short != short {
						t.Errorf("expected Short %q, got %q", short, sub.Short)
					}
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not found", use)
			}
		})
	}
}

// TestOperationLogsListFlags verifies list command flags exist with correct defaults.
func TestOperationLogsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
		flagType     string
	}{
		{"page", "1", "int"},
		{"page-size", "20", "int"},
		{"action", "", "string"},
		{"resource-type", "", "string"},
		{"resource-id", "", "string"},
		{"user-id", "", "string"},
		{"since", "", "string"},
		{"until", "", "string"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := operationLogsListCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
			if flag.Value.Type() != f.flagType {
				t.Errorf("expected type %q, got %q", f.flagType, flag.Value.Type())
			}
		})
	}
}

// TestOperationLogsGetArgs verifies get command requires exactly one argument.
func TestOperationLogsGetArgs(t *testing.T) {
	if operationLogsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", operationLogsGetCmd.Use)
	}
	if operationLogsGetCmd.Args == nil {
		t.Fatal("Expected Args validator on get command")
	}
	err := operationLogsGetCmd.Args(operationLogsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = operationLogsGetCmd.Args(operationLogsGetCmd, []string{"log_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
	err = operationLogsGetCmd.Args(operationLogsGetCmd, []string{"log_123", "log_456"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

// TestOperationLogsListRunE tests the list command execution with mock API.
func TestOperationLogsListRunE(t *testing.T) {
	restore := setupOperationLogsTest(t)
	defer restore()

	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		mockResp     *api.OperationLogsListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantContains []string
	}{
		{
			name: "successful list with text output",
			mockResp: &api.OperationLogsListResponse{
				Items: []api.OperationLog{
					{
						ID:           "log_123",
						Action:       api.OperationLogActionCreate,
						ResourceType: "product",
						ResourceID:   "prod_456",
						UserEmail:    "admin@example.com",
						IPAddress:    "192.168.1.1",
						CreatedAt:    testTime,
					},
				},
				TotalCount: 1,
			},
			wantContains: []string{"log_123", "create", "product:prod_456", "admin@example.com"},
		},
		{
			name: "successful list with JSON output",
			mockResp: &api.OperationLogsListResponse{
				Items: []api.OperationLog{
					{
						ID:           "log_123",
						Action:       api.OperationLogActionUpdate,
						ResourceType: "order",
						ResourceID:   "ord_789",
						UserEmail:    "user@example.com",
						IPAddress:    "10.0.0.1",
						CreatedAt:    testTime,
					},
				},
				TotalCount: 1,
			},
			outputFormat: "json",
		},
		{
			name: "list with empty resource ID",
			mockResp: &api.OperationLogsListResponse{
				Items: []api.OperationLog{
					{
						ID:           "log_124",
						Action:       api.OperationLogActionLogin,
						ResourceType: "session",
						ResourceID:   "",
						UserEmail:    "user@example.com",
						IPAddress:    "10.0.0.2",
						CreatedAt:    testTime,
					},
				},
				TotalCount: 1,
			},
			wantContains: []string{"session"},
		},
		{
			name: "list using UserName when UserEmail is empty",
			mockResp: &api.OperationLogsListResponse{
				Items: []api.OperationLog{
					{
						ID:           "log_125",
						Action:       api.OperationLogActionDelete,
						ResourceType: "customer",
						ResourceID:   "cust_001",
						UserEmail:    "",
						UserName:     "John Doe",
						IPAddress:    "172.16.0.1",
						CreatedAt:    testTime,
					},
				},
				TotalCount: 1,
			},
			wantContains: []string{"John Doe"},
		},
		{
			name: "empty list",
			mockResp: &api.OperationLogsListResponse{
				Items:      []api.OperationLog{},
				TotalCount: 0,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &operationLogsAPIClient{
				listOperationLogsResp: tt.mockResp,
				listOperationLogsErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := createOperationLogsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := operationLogsListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
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

// TestOperationLogsListRunE_DateFilters tests date filter parsing.
func TestOperationLogsListRunE_DateFilters(t *testing.T) {
	restore := setupOperationLogsTest(t)
	defer restore()

	tests := []struct {
		name    string
		since   string
		until   string
		wantErr bool
		errMsg  string
	}{
		{
			name:  "valid RFC3339 dates",
			since: "2024-01-01T00:00:00Z",
			until: "2024-12-31T23:59:59Z",
		},
		{
			name:  "valid short dates",
			since: "2024-01-01",
			until: "2024-12-31",
		},
		{
			name:  "only since date",
			since: "2024-06-01",
		},
		{
			name:  "only until date",
			until: "2024-06-30",
		},
		{
			name:    "invalid since date",
			since:   "not-a-date",
			wantErr: true,
			errMsg:  "invalid from date format",
		},
		{
			name:    "invalid until date",
			until:   "also-not-a-date",
			wantErr: true,
			errMsg:  "invalid to date format",
		},
		{
			name:    "partially valid since date",
			since:   "2024/01/01",
			wantErr: true,
			errMsg:  "invalid from date format",
		},
		{
			name:    "partially valid until date",
			until:   "01-01-2024",
			wantErr: true,
			errMsg:  "invalid to date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &operationLogsAPIClient{
				listOperationLogsResp: &api.OperationLogsListResponse{
					Items:      []api.OperationLog{},
					TotalCount: 0,
				},
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := createOperationLogsTestCmd()
			if tt.since != "" {
				_ = cmd.Flags().Set("since", tt.since)
			}
			if tt.until != "" {
				_ = cmd.Flags().Set("until", tt.until)
			}

			err := operationLogsListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestOperationLogsListRunE_GetClientFails tests error handling when getClient fails.
func TestOperationLogsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("action", "", "")
	cmd.Flags().String("resource-type", "", "")
	cmd.Flags().String("resource-id", "", "")
	cmd.Flags().String("user-id", "", "")
	cmd.Flags().String("since", "", "")
	cmd.Flags().String("until", "", "")

	err := operationLogsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestOperationLogsGetRunE tests the get command execution with mock API.
func TestOperationLogsGetRunE(t *testing.T) {
	restore := setupOperationLogsTest(t)
	defer restore()

	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		id           string
		mockResp     *api.OperationLog
		mockErr      error
		outputFormat string
		wantErr      bool
		wantContains []string
	}{
		{
			name: "successful get with text output - full details",
			id:   "log_123",
			mockResp: &api.OperationLog{
				ID:           "log_123",
				Action:       api.OperationLogActionUpdate,
				ResourceType: "product",
				ResourceID:   "prod_456",
				ResourceName: "Awesome Widget",
				UserID:       "user_789",
				UserEmail:    "admin@example.com",
				UserName:     "Admin User",
				IPAddress:    "192.168.1.1",
				UserAgent:    "Mozilla/5.0",
				CreatedAt:    testTime,
				Changes: map[string]api.Change{
					"price": {From: 10.00, To: 15.00},
					"title": {From: "Old Title", To: "New Title"},
				},
				Metadata: map[string]string{
					"source":   "web",
					"category": "electronics",
				},
			},
			wantContains: []string{
				"log_123",
				"update",
				"product",
				"prod_456",
				"Awesome Widget",
				"admin@example.com",
				"Admin User",
				"Mozilla/5.0",
				"Changes:",
				"Metadata:",
				"source",
				"web",
			},
		},
		{
			name: "successful get with JSON output",
			id:   "log_124",
			mockResp: &api.OperationLog{
				ID:           "log_124",
				Action:       api.OperationLogActionCreate,
				ResourceType: "order",
				ResourceID:   "ord_789",
				UserID:       "user_001",
				UserEmail:    "user@example.com",
				IPAddress:    "10.0.0.1",
				CreatedAt:    testTime,
			},
			outputFormat: "json",
		},
		{
			name: "get without optional fields",
			id:   "log_125",
			mockResp: &api.OperationLog{
				ID:           "log_125",
				Action:       api.OperationLogActionLogin,
				ResourceType: "session",
				ResourceID:   "",
				ResourceName: "",
				UserID:       "user_002",
				UserEmail:    "login@example.com",
				UserName:     "",
				IPAddress:    "172.16.0.1",
				UserAgent:    "",
				CreatedAt:    testTime,
				Changes:      nil,
				Metadata:     nil,
			},
			wantContains: []string{"log_125", "login", "session"},
		},
		{
			name: "get with changes but no metadata",
			id:   "log_126",
			mockResp: &api.OperationLog{
				ID:           "log_126",
				Action:       api.OperationLogActionDelete,
				ResourceType: "customer",
				ResourceID:   "cust_001",
				UserID:       "user_003",
				UserEmail:    "support@example.com",
				IPAddress:    "192.168.0.5",
				CreatedAt:    testTime,
				Changes: map[string]api.Change{
					"status": {From: "active", To: "deleted"},
				},
				Metadata: nil,
			},
			wantContains: []string{"Changes:", "status"},
		},
		{
			name: "get with metadata but no changes",
			id:   "log_127",
			mockResp: &api.OperationLog{
				ID:           "log_127",
				Action:       api.OperationLogActionExport,
				ResourceType: "report",
				ResourceID:   "rpt_001",
				UserID:       "user_004",
				UserEmail:    "analyst@example.com",
				IPAddress:    "10.10.10.1",
				CreatedAt:    testTime,
				Changes:      nil,
				Metadata: map[string]string{
					"format": "csv",
					"rows":   "10000",
				},
			},
			wantContains: []string{"Metadata:", "format", "csv"},
		},
		{
			name:    "not found",
			id:      "log_999",
			mockErr: errors.New("operation log not found"),
			wantErr: true,
		},
		{
			name:    "API error",
			id:      "log_error",
			mockErr: errors.New("internal server error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &operationLogsAPIClient{
				getOperationLogResp: tt.mockResp,
				getOperationLogErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := createOperationLogsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := operationLogsGetCmd.RunE(cmd, []string{tt.id})

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
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

// TestOperationLogsGetRunE_GetClientFails tests error handling when getClient fails.
func TestOperationLogsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := operationLogsGetCmd.RunE(cmd, []string{"log_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestOperationLogsListRunE_WithFilters tests list command with various filter combinations.
func TestOperationLogsListRunE_WithFilters(t *testing.T) {
	restore := setupOperationLogsTest(t)
	defer restore()

	mockClient := &operationLogsAPIClient{
		listOperationLogsResp: &api.OperationLogsListResponse{
			Items:      []api.OperationLog{},
			TotalCount: 0,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	tests := []struct {
		name         string
		action       string
		resourceType string
		resourceID   string
		userID       string
		page         string
		pageSize     string
	}{
		{
			name:   "filter by action",
			action: "create",
		},
		{
			name:         "filter by resource type",
			resourceType: "product",
		},
		{
			name:       "filter by resource ID",
			resourceID: "prod_123",
		},
		{
			name:   "filter by user ID",
			userID: "user_456",
		},
		{
			name:         "multiple filters",
			action:       "update",
			resourceType: "order",
			userID:       "user_789",
		},
		{
			name:     "custom pagination",
			page:     "2",
			pageSize: "50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := createOperationLogsTestCmd()
			if tt.action != "" {
				_ = cmd.Flags().Set("action", tt.action)
			}
			if tt.resourceType != "" {
				_ = cmd.Flags().Set("resource-type", tt.resourceType)
			}
			if tt.resourceID != "" {
				_ = cmd.Flags().Set("resource-id", tt.resourceID)
			}
			if tt.userID != "" {
				_ = cmd.Flags().Set("user-id", tt.userID)
			}
			if tt.page != "" {
				_ = cmd.Flags().Set("page", tt.page)
			}
			if tt.pageSize != "" {
				_ = cmd.Flags().Set("page-size", tt.pageSize)
			}

			err := operationLogsListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestOperationLogsListFlagDescriptions verifies flag descriptions are set.
func TestOperationLogsListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"page":          "Page number",
		"page-size":     "Results per page",
		"action":        "Filter by action (create, update, delete, login, logout, export, import)",
		"resource-type": "Filter by resource type (product, order, customer, etc.)",
		"resource-id":   "Filter by resource ID",
		"user-id":       "Filter by user ID",
		"since":         "Filter by start date (YYYY-MM-DD or RFC3339)",
		"until":         "Filter by end date (YYYY-MM-DD or RFC3339)",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := operationLogsListCmd.Flags().Lookup(flagName)
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
