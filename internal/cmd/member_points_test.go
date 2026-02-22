package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func TestMemberPointsCmd(t *testing.T) {
	if memberPointsCmd.Use != "member-points" {
		t.Errorf("Expected Use to be 'member-points', got %q", memberPointsCmd.Use)
	}
}

func TestMemberPointsGetCmd(t *testing.T) {
	if memberPointsGetCmd.Use != "get" {
		t.Errorf("Expected Use to be 'get', got %q", memberPointsGetCmd.Use)
	}
}

func TestMemberPointsTransactionsCmd(t *testing.T) {
	if memberPointsTransactionsCmd.Use != "transactions" {
		t.Errorf("Expected Use to be 'transactions', got %q", memberPointsTransactionsCmd.Use)
	}
}

func TestMemberPointsAdjustCmd(t *testing.T) {
	if memberPointsAdjustCmd.Use != "adjust" {
		t.Errorf("Expected Use to be 'adjust', got %q", memberPointsAdjustCmd.Use)
	}
}

func TestMemberPointsPersistentFlags(t *testing.T) {
	if memberPointsCmd.PersistentFlags().Lookup("customer-id") == nil {
		t.Error("Expected persistent flag 'customer-id' to be defined")
	}
}

func TestMemberPointsTransactionsFlags(t *testing.T) {
	flags := []string{"page", "page-size", "type"}
	for _, flag := range flags {
		if memberPointsTransactionsCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestMemberPointsAdjustFlags(t *testing.T) {
	flags := []string{"points", "description"}
	for _, flag := range flags {
		if memberPointsAdjustCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestMemberPointsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	err := memberPointsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMemberPointsTransactionsRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("type", "", "")
	err := memberPointsTransactionsCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMemberPointsAdjustRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("points", 100, "")
	cmd.Flags().String("description", "bonus points", "")
	err := memberPointsAdjustCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMemberPointsGetRunE_NoProfiles(t *testing.T) {
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
	err := memberPointsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// memberPointsTestClient is a mock implementation for member points testing.
type memberPointsTestClient struct {
	api.MockClient

	getMemberPointsResp        *api.MemberPoints
	getMemberPointsErr         error
	listPointsTransactionsResp *api.PointsTransactionsListResponse
	listPointsTransactionsErr  error
	adjustMemberPointsResp     *api.MemberPoints
	adjustMemberPointsErr      error

	getHistoryResp json.RawMessage
	getHistoryErr  error
	updateResp     json.RawMessage
	updateErr      error
	rulesResp      json.RawMessage
	rulesErr       error
	bulkResp       json.RawMessage
	bulkErr        error
}

func (m *memberPointsTestClient) GetMemberPoints(ctx context.Context, customerID string) (*api.MemberPoints, error) {
	return m.getMemberPointsResp, m.getMemberPointsErr
}

func (m *memberPointsTestClient) ListPointsTransactions(ctx context.Context, customerID string, opts *api.PointsTransactionsListOptions) (*api.PointsTransactionsListResponse, error) {
	return m.listPointsTransactionsResp, m.listPointsTransactionsErr
}

func (m *memberPointsTestClient) AdjustMemberPoints(ctx context.Context, customerID string, points int, description string) (*api.MemberPoints, error) {
	return m.adjustMemberPointsResp, m.adjustMemberPointsErr
}

func (m *memberPointsTestClient) GetCustomerMemberPointsHistory(ctx context.Context, customerID string) (json.RawMessage, error) {
	return m.getHistoryResp, m.getHistoryErr
}

func (m *memberPointsTestClient) UpdateCustomerMemberPoints(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	return m.updateResp, m.updateErr
}

func (m *memberPointsTestClient) ListMemberPointRules(ctx context.Context) (json.RawMessage, error) {
	return m.rulesResp, m.rulesErr
}

func (m *memberPointsTestClient) BulkUpdateMemberPoints(ctx context.Context, body any) (json.RawMessage, error) {
	return m.bulkResp, m.bulkErr
}

// setupMemberPointsTest configures mocks for member points command tests.
func setupMemberPointsTest(t *testing.T) (restore func()) {
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

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// TestMemberPointsGetRunE tests the member points get command execution with mock API.
func TestMemberPointsGetRunE(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	tests := []struct {
		name       string
		mockResp   *api.MemberPoints
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful get",
			mockResp: &api.MemberPoints{
				CustomerID:      "cust_123",
				TotalPoints:     1500,
				AvailablePoints: 1200,
				PendingPoints:   200,
				ExpiredPoints:   100,
				UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantOutput: "cust_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("customer not found"),
			wantErr: true,
		},
		{
			name: "zero points",
			mockResp: &api.MemberPoints{
				CustomerID:      "cust_456",
				TotalPoints:     0,
				AvailablePoints: 0,
				PendingPoints:   0,
				ExpiredPoints:   0,
				UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantOutput: "cust_456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &memberPointsTestClient{
				getMemberPointsResp: tt.mockResp,
				getMemberPointsErr:  tt.mockErr,
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
			cmd.Flags().String("customer-id", "cust_123", "")

			err := memberPointsGetCmd.RunE(cmd, []string{})

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

// TestMemberPointsGetRunE_JSON tests the member points get command with JSON output.
func TestMemberPointsGetRunE_JSON(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		getMemberPointsResp: &api.MemberPoints{
			CustomerID:      "cust_123",
			TotalPoints:     1500,
			AvailablePoints: 1200,
			PendingPoints:   200,
			ExpiredPoints:   100,
			UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("customer-id", "cust_123", "")

	err := memberPointsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "customer_id") {
		t.Errorf("JSON output should contain customer_id, got: %s", output)
	}
}

// TestMemberPointsTransactionsRunE tests the member points transactions command execution with mock API.
func TestMemberPointsTransactionsRunE(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	tests := []struct {
		name       string
		mockResp   *api.PointsTransactionsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.PointsTransactionsListResponse{
				Items: []api.PointsTransaction{
					{
						ID:          "txn_123",
						CustomerID:  "cust_456",
						Type:        "earn",
						Points:      100,
						Balance:     1100,
						Description: "Purchase reward",
						OrderID:     "order_789",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "txn_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.PointsTransactionsListResponse{
				Items:      []api.PointsTransaction{},
				TotalCount: 0,
			},
		},
		{
			name: "transaction without order ID",
			mockResp: &api.PointsTransactionsListResponse{
				Items: []api.PointsTransaction{
					{
						ID:          "txn_456",
						CustomerID:  "cust_789",
						Type:        "adjust",
						Points:      -50,
						Balance:     950,
						Description: "Manual adjustment",
						OrderID:     "",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "txn_456",
		},
		{
			name: "transaction with long description",
			mockResp: &api.PointsTransactionsListResponse{
				Items: []api.PointsTransaction{
					{
						ID:          "txn_789",
						CustomerID:  "cust_101",
						Type:        "redeem",
						Points:      -200,
						Balance:     800,
						Description: "This is a very long description that should be truncated to fit in the table",
						OrderID:     "order_202",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "txn_789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &memberPointsTestClient{
				listPointsTransactionsResp: tt.mockResp,
				listPointsTransactionsErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			// Capture stdout for text output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("type", "", "")

			err := memberPointsTransactionsCmd.RunE(cmd, []string{})

			_ = w.Close()
			os.Stdout = oldStdout
			var stdoutBuf bytes.Buffer
			_, _ = stdoutBuf.ReadFrom(r)

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
		})
	}
}

// TestMemberPointsTransactionsRunE_JSON tests the member points transactions command with JSON output.
func TestMemberPointsTransactionsRunE_JSON(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		listPointsTransactionsResp: &api.PointsTransactionsListResponse{
			Items: []api.PointsTransaction{
				{
					ID:          "txn_123",
					CustomerID:  "cust_456",
					Type:        "earn",
					Points:      100,
					Balance:     1100,
					Description: "Purchase reward",
					OrderID:     "order_789",
					CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("type", "", "")

	err := memberPointsTransactionsCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "txn_123") {
		t.Errorf("JSON output should contain transaction ID, got: %s", output)
	}
}

// TestMemberPointsTransactionsRunE_WithFilters tests the member points transactions command with filters.
func TestMemberPointsTransactionsRunE_WithFilters(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		listPointsTransactionsResp: &api.PointsTransactionsListResponse{
			Items: []api.PointsTransaction{
				{
					ID:          "txn_123",
					CustomerID:  "cust_456",
					Type:        "earn",
					Points:      100,
					Balance:     1100,
					Description: "Purchase reward",
					OrderID:     "order_789",
					CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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

	// Capture stdout for text output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")
	cmd.Flags().String("type", "earn", "")

	err := memberPointsTransactionsCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout
	var stdoutBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(r)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMemberPointsAdjustRunE tests the member points adjust command execution with mock API.
func TestMemberPointsAdjustRunE(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	tests := []struct {
		name        string
		points      int
		description string
		mockResp    *api.MemberPoints
		mockErr     error
		wantErr     bool
		wantOutput  string
	}{
		{
			name:        "successful positive adjustment",
			points:      100,
			description: "Bonus points",
			mockResp: &api.MemberPoints{
				CustomerID:      "cust_123",
				TotalPoints:     1600,
				AvailablePoints: 1300,
				PendingPoints:   200,
				ExpiredPoints:   100,
				UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantOutput: "1600",
		},
		{
			name:        "successful negative adjustment",
			points:      -50,
			description: "Points deduction",
			mockResp: &api.MemberPoints{
				CustomerID:      "cust_123",
				TotalPoints:     1450,
				AvailablePoints: 1150,
				PendingPoints:   200,
				ExpiredPoints:   100,
				UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantOutput: "1450",
		},
		{
			name:        "API error",
			points:      100,
			description: "Bonus",
			mockErr:     errors.New("insufficient points"),
			wantErr:     true,
		},
		{
			name:        "zero point adjustment",
			points:      0,
			description: "No change",
			mockResp: &api.MemberPoints{
				CustomerID:      "cust_123",
				TotalPoints:     1500,
				AvailablePoints: 1200,
				PendingPoints:   200,
				ExpiredPoints:   100,
				UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantOutput: "1500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &memberPointsTestClient{
				adjustMemberPointsResp: tt.mockResp,
				adjustMemberPointsErr:  tt.mockErr,
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
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().Int("points", tt.points, "")
			cmd.Flags().String("description", tt.description, "")

			err := memberPointsAdjustCmd.RunE(cmd, []string{})

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

// TestMemberPointsAdjustRunE_JSON tests the member points adjust command with JSON output.
func TestMemberPointsAdjustRunE_JSON(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		adjustMemberPointsResp: &api.MemberPoints{
			CustomerID:      "cust_123",
			TotalPoints:     1600,
			AvailablePoints: 1300,
			PendingPoints:   200,
			ExpiredPoints:   100,
			UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("points", 100, "")
	cmd.Flags().String("description", "Bonus points", "")

	err := memberPointsAdjustCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "customer_id") {
		t.Errorf("JSON output should contain customer_id, got: %s", output)
	}
}

// TestMemberPointsTransactionsRunE_NoProfiles tests the member points transactions command with no profiles.
func TestMemberPointsTransactionsRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("type", "", "")
	err := memberPointsTransactionsCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestMemberPointsAdjustRunE_NoProfiles tests the member points adjust command with no profiles.
func TestMemberPointsAdjustRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().Int("points", 100, "")
	cmd.Flags().String("description", "test", "")
	err := memberPointsAdjustCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestMemberPointsGetRunE_TextOutputAllFields tests all fields are displayed in text output.
func TestMemberPointsGetRunE_TextOutputAllFields(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		getMemberPointsResp: &api.MemberPoints{
			CustomerID:      "cust_test_123",
			TotalPoints:     2500,
			AvailablePoints: 2000,
			PendingPoints:   300,
			ExpiredPoints:   200,
			UpdatedAt:       time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_test_123", "")

	err := memberPointsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	expectedFields := []string{
		"Customer ID:",
		"cust_test_123",
		"Total Points:",
		"2500",
		"Available Points:",
		"2000",
		"Pending Points:",
		"300",
		"Expired Points:",
		"200",
		"Updated:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("output should contain %q, got: %s", field, output)
		}
	}
}

// TestMemberPointsTransactionsRunE_DescriptionTruncation tests that long descriptions are truncated.
func TestMemberPointsTransactionsRunE_DescriptionTruncation(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	// Test with description exactly at 25 characters (should not truncate)
	mockClient := &memberPointsTestClient{
		listPointsTransactionsResp: &api.PointsTransactionsListResponse{
			Items: []api.PointsTransaction{
				{
					ID:          "txn_exact",
					CustomerID:  "cust_456",
					Type:        "earn",
					Points:      100,
					Balance:     1100,
					Description: "This is exactly25chars!!", // 25 characters
					OrderID:     "order_789",
					CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("type", "", "")

	err := memberPointsTransactionsCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout
	var stdoutBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(r)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMemberPointsAdjustRunE_LargePointValue tests adjustment with large point values.
func TestMemberPointsAdjustRunE_LargePointValue(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		adjustMemberPointsResp: &api.MemberPoints{
			CustomerID:      "cust_123",
			TotalPoints:     1000000,
			AvailablePoints: 999000,
			PendingPoints:   1000,
			ExpiredPoints:   0,
			UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("points", 999999, "")
	cmd.Flags().String("description", "Large bonus", "")

	err := memberPointsAdjustCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "999999") {
		t.Errorf("output should contain the adjustment amount, got: %s", output)
	}
}

func TestMemberPointsHistoryRunE_JSON(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		getHistoryResp: json.RawMessage(`{"items":[{"id":"h_1"}]}`),
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
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("customer-id", "cust_123", "")

	if err := memberPointsHistoryCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"h_1"`)) {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestMemberPointsUpdateRunE_JSON(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		updateResp: json.RawMessage(`{"ok":true}`),
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
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().Int("points", 0, "")
	cmd.Flags().String("description", "", "")
	_ = cmd.Flags().Set("points", "10")

	if err := memberPointsUpdateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"ok": true`)) {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestMemberPointsRulesListRunE_JSON(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		rulesResp: json.RawMessage(`{"items":[{"id":"r_1"}]}`),
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
	cmd.Flags().Bool("items-only", false, "")

	if err := memberPointsRulesListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"r_1"`)) {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestMemberPointsBulkUpdateRunE_JSON(t *testing.T) {
	restore := setupMemberPointsTest(t)
	defer restore()

	mockClient := &memberPointsTestClient{
		bulkResp: json.RawMessage(`{"ok":true}`),
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
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("body", `{"items":[]}`, "")
	cmd.Flags().String("body-file", "", "")

	if err := memberPointsBulkUpdateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"ok": true`)) {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}
