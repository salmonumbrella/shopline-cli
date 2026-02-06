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

// Note: floatToString is defined in cmd_test.go

func TestOrderRisksCommandStructure(t *testing.T) {
	if orderRisksCmd == nil {
		t.Fatal("orderRisksCmd is nil")
	}
	if orderRisksCmd.Use != "order-risks" {
		t.Errorf("Expected Use 'order-risks', got %q", orderRisksCmd.Use)
	}
	subcommands := map[string]bool{"list": false, "get": false, "create": false, "delete": false}
	for _, cmd := range orderRisksCmd.Commands() {
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

func TestOrderRisksListFlags(t *testing.T) {
	cmd := orderRisksListCmd
	flags := []struct{ name, defaultValue string }{{"page", "1"}, {"page-size", "20"}}
	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("Flag %q not found", f.name)
		} else if flag.DefValue != f.defaultValue {
			t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestOrderRisksListRequiresArg(t *testing.T) {
	cmd := orderRisksListCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"order_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestOrderRisksGetRequiresTwoArgs(t *testing.T) {
	cmd := orderRisksGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if cmd.Args(cmd, []string{"order_123"}) == nil {
		t.Error("Expected error with 1 arg")
	}
	if err := cmd.Args(cmd, []string{"order_123", "risk_456"}); err != nil {
		t.Errorf("Expected no error with 2 args, got: %v", err)
	}
}

func TestOrderRisksCreateRequiresArg(t *testing.T) {
	cmd := orderRisksCreateCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"order_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestOrderRisksDeleteRequiresTwoArgs(t *testing.T) {
	cmd := orderRisksDeleteCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if cmd.Args(cmd, []string{"order_123"}) == nil {
		t.Error("Expected error with 1 arg")
	}
	if err := cmd.Args(cmd, []string{"order_123", "risk_456"}); err != nil {
		t.Errorf("Expected no error with 2 args, got: %v", err)
	}
}

func TestOrderRisksCreateFlags(t *testing.T) {
	cmd := orderRisksCreateCmd
	flags := []struct{ name, defaultValue string }{
		{"score", "0"},
		{"recommendation", ""},
		{"source", ""},
		{"message", ""},
		{"display", "false"},
		{"cause-cancel", "false"},
	}
	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("Flag %q not found", f.name)
		} else if flag.DefValue != f.defaultValue {
			t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestOrderRisksListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := orderRisksListCmd.RunE(cmd, []string{"order_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOrderRisksListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := orderRisksListCmd.RunE(cmd, []string{"order_123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestOrderRisksListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() { secretsStoreFactory = origFactory; _ = os.Setenv("SHOPLINE_STORE", origEnv) }()
	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{"envstore", "other"}, creds: map[string]*secrets.StoreCredentials{"envstore": {Handle: "test", AccessToken: "token123"}}}, nil
	}
	cmd := newTestCmdWithFlags()
	err := orderRisksListCmd.RunE(cmd, []string{"order_123"})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// orderRisksTestClient is a mock implementation for order risks testing.
type orderRisksTestClient struct {
	api.MockClient

	listOrderRisksResp  *api.OrderRisksListResponse
	listOrderRisksErr   error
	getOrderRiskResp    *api.OrderRisk
	getOrderRiskErr     error
	createOrderRiskResp *api.OrderRisk
	createOrderRiskErr  error
	deleteOrderRiskErr  error
}

func (m *orderRisksTestClient) ListOrderRisks(ctx context.Context, orderID string, opts *api.OrderRisksListOptions) (*api.OrderRisksListResponse, error) {
	return m.listOrderRisksResp, m.listOrderRisksErr
}

func (m *orderRisksTestClient) GetOrderRisk(ctx context.Context, orderID, riskID string) (*api.OrderRisk, error) {
	return m.getOrderRiskResp, m.getOrderRiskErr
}

func (m *orderRisksTestClient) CreateOrderRisk(ctx context.Context, orderID string, req *api.OrderRiskCreateRequest) (*api.OrderRisk, error) {
	return m.createOrderRiskResp, m.createOrderRiskErr
}

func (m *orderRisksTestClient) DeleteOrderRisk(ctx context.Context, orderID, riskID string) error {
	return m.deleteOrderRiskErr
}

// setupOrderRisksTest sets up the test environment for order risks tests.
func setupOrderRisksTest(t *testing.T) (cleanup func()) {
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

// newOrderRisksTestCmd creates a test command with necessary flags.
func newOrderRisksTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().Float64("score", 0, "")
	cmd.Flags().String("recommendation", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().String("message", "", "")
	cmd.Flags().Bool("display", false, "")
	cmd.Flags().Bool("cause-cancel", false, "")
	cmd.Flags().Bool("yes", false, "")
	return cmd
}

// TestOrderRisksListRunE tests the order risks list command execution with mock API.
func TestOrderRisksListRunE(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.OrderRisksListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.OrderRisksListResponse{
				Items: []api.OrderRisk{
					{
						ID:             "risk_123",
						OrderID:        "order_456",
						Score:          0.85,
						Recommendation: "investigate",
						Source:         "fraud_detection",
						Message:        "High risk detected",
						Display:        true,
						CauseCancel:    false,
						CreatedAt:      testTime,
						UpdatedAt:      testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "risk_123",
		},
		{
			name: "multiple risks",
			mockResp: &api.OrderRisksListResponse{
				Items: []api.OrderRisk{
					{
						ID:             "risk_001",
						OrderID:        "order_456",
						Score:          0.50,
						Recommendation: "accept",
						Source:         "manual",
						Display:        false,
						CauseCancel:    false,
						CreatedAt:      testTime,
						UpdatedAt:      testTime,
					},
					{
						ID:             "risk_002",
						OrderID:        "order_456",
						Score:          0.95,
						Recommendation: "cancel",
						Source:         "ai_detector",
						Display:        true,
						CauseCancel:    true,
						CreatedAt:      testTime,
						UpdatedAt:      testTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "risk_001",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.OrderRisksListResponse{
				Items:      []api.OrderRisk{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &orderRisksTestClient{
				listOrderRisksResp: tt.mockResp,
				listOrderRisksErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newOrderRisksTestCmd()

			err := orderRisksListCmd.RunE(cmd, []string{"order_456"})

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

// TestOrderRisksListJSONOutput tests JSON output format for the list command.
func TestOrderRisksListJSONOutput(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &orderRisksTestClient{
		listOrderRisksResp: &api.OrderRisksListResponse{
			Items: []api.OrderRisk{
				{
					ID:             "risk_123",
					OrderID:        "order_456",
					Score:          0.85,
					Recommendation: "investigate",
					Source:         "fraud_detection",
					Display:        true,
					CauseCancel:    false,
					CreatedAt:      testTime,
					UpdatedAt:      testTime,
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

	cmd := newOrderRisksTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := orderRisksListCmd.RunE(cmd, []string{"order_456"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "risk_123") {
		t.Errorf("JSON output should contain risk ID, got: %s", output)
	}
}

// TestOrderRisksGetRunE tests the order risks get command execution with mock API.
func TestOrderRisksGetRunE(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		orderID  string
		riskID   string
		mockResp *api.OrderRisk
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get",
			orderID: "order_456",
			riskID:  "risk_123",
			mockResp: &api.OrderRisk{
				ID:             "risk_123",
				OrderID:        "order_456",
				Score:          0.85,
				Recommendation: "investigate",
				Source:         "fraud_detection",
				Message:        "High risk detected",
				Display:        true,
				CauseCancel:    false,
				CreatedAt:      testTime,
				UpdatedAt:      testTime,
			},
		},
		{
			name:    "get with message",
			orderID: "order_789",
			riskID:  "risk_456",
			mockResp: &api.OrderRisk{
				ID:             "risk_456",
				OrderID:        "order_789",
				Score:          0.50,
				Recommendation: "accept",
				Source:         "manual",
				Message:        "Manual review passed",
				Display:        false,
				CauseCancel:    false,
				CreatedAt:      testTime,
				UpdatedAt:      testTime,
			},
		},
		{
			name:    "risk not found",
			orderID: "order_456",
			riskID:  "risk_999",
			mockErr: errors.New("risk not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &orderRisksTestClient{
				getOrderRiskResp: tt.mockResp,
				getOrderRiskErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newOrderRisksTestCmd()

			err := orderRisksGetCmd.RunE(cmd, []string{tt.orderID, tt.riskID})

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

// TestOrderRisksGetJSONOutput tests JSON output format for the get command.
func TestOrderRisksGetJSONOutput(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &orderRisksTestClient{
		getOrderRiskResp: &api.OrderRisk{
			ID:             "risk_123",
			OrderID:        "order_456",
			Score:          0.85,
			Recommendation: "investigate",
			Source:         "fraud_detection",
			Message:        "High risk detected",
			Display:        true,
			CauseCancel:    false,
			CreatedAt:      testTime,
			UpdatedAt:      testTime,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newOrderRisksTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := orderRisksGetCmd.RunE(cmd, []string{"order_456", "risk_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "risk_123") {
		t.Errorf("JSON output should contain risk ID, got: %s", output)
	}
}

// TestOrderRisksGetWithEmptyMessage tests get command when message is empty.
func TestOrderRisksGetWithEmptyMessage(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &orderRisksTestClient{
		getOrderRiskResp: &api.OrderRisk{
			ID:             "risk_123",
			OrderID:        "order_456",
			Score:          0.85,
			Recommendation: "investigate",
			Source:         "fraud_detection",
			Message:        "", // Empty message
			Display:        true,
			CauseCancel:    false,
			CreatedAt:      testTime,
			UpdatedAt:      testTime,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newOrderRisksTestCmd()

	err := orderRisksGetCmd.RunE(cmd, []string{"order_456", "risk_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestOrderRisksCreateRunE tests the order risks create command execution with mock API.
func TestOrderRisksCreateRunE(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		orderID        string
		score          float64
		recommendation string
		source         string
		message        string
		display        bool
		causeCancel    bool
		mockResp       *api.OrderRisk
		mockErr        error
		wantErr        bool
	}{
		{
			name:           "successful create",
			orderID:        "order_456",
			score:          0.85,
			recommendation: "investigate",
			source:         "fraud_detection",
			message:        "High risk detected",
			display:        true,
			causeCancel:    false,
			mockResp: &api.OrderRisk{
				ID:             "risk_new_123",
				OrderID:        "order_456",
				Score:          0.85,
				Recommendation: "investigate",
				Source:         "fraud_detection",
				Message:        "High risk detected",
				Display:        true,
				CauseCancel:    false,
				CreatedAt:      testTime,
				UpdatedAt:      testTime,
			},
		},
		{
			name:           "create with cause cancel",
			orderID:        "order_789",
			score:          0.95,
			recommendation: "cancel",
			source:         "ai_detector",
			display:        true,
			causeCancel:    true,
			mockResp: &api.OrderRisk{
				ID:             "risk_new_456",
				OrderID:        "order_789",
				Score:          0.95,
				Recommendation: "cancel",
				Source:         "ai_detector",
				Display:        true,
				CauseCancel:    true,
				CreatedAt:      testTime,
				UpdatedAt:      testTime,
			},
		},
		{
			name:           "create fails",
			orderID:        "order_456",
			score:          0.50,
			recommendation: "accept",
			mockErr:        errors.New("failed to create risk"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &orderRisksTestClient{
				createOrderRiskResp: tt.mockResp,
				createOrderRiskErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newOrderRisksTestCmd()
			_ = cmd.Flags().Set("score", floatToString(tt.score))
			_ = cmd.Flags().Set("recommendation", tt.recommendation)
			if tt.source != "" {
				_ = cmd.Flags().Set("source", tt.source)
			}
			if tt.message != "" {
				_ = cmd.Flags().Set("message", tt.message)
			}
			if tt.display {
				_ = cmd.Flags().Set("display", "true")
			}
			if tt.causeCancel {
				_ = cmd.Flags().Set("cause-cancel", "true")
			}

			err := orderRisksCreateCmd.RunE(cmd, []string{tt.orderID})

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

// TestOrderRisksCreateJSONOutput tests JSON output format for the create command.
func TestOrderRisksCreateJSONOutput(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &orderRisksTestClient{
		createOrderRiskResp: &api.OrderRisk{
			ID:             "risk_new_123",
			OrderID:        "order_456",
			Score:          0.85,
			Recommendation: "investigate",
			Source:         "fraud_detection",
			Display:        true,
			CauseCancel:    false,
			CreatedAt:      testTime,
			UpdatedAt:      testTime,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newOrderRisksTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("score", "0.85")
	_ = cmd.Flags().Set("recommendation", "investigate")

	err := orderRisksCreateCmd.RunE(cmd, []string{"order_456"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "risk_new_123") {
		t.Errorf("JSON output should contain risk ID, got: %s", output)
	}
}

// TestOrderRisksDeleteRunE tests the order risks delete command execution with mock API.
func TestOrderRisksDeleteRunE(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	tests := []struct {
		name    string
		orderID string
		riskID  string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			orderID: "order_456",
			riskID:  "risk_123",
		},
		{
			name:    "delete fails",
			orderID: "order_456",
			riskID:  "risk_999",
			mockErr: errors.New("risk not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &orderRisksTestClient{
				deleteOrderRiskErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := newOrderRisksTestCmd()
			_ = cmd.Flags().Set("yes", "true") // Skip confirmation

			err := orderRisksDeleteCmd.RunE(cmd, []string{tt.orderID, tt.riskID})

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

// TestOrderRisksDeleteWithoutConfirmation tests the delete command skips when not confirmed.
func TestOrderRisksDeleteWithoutConfirmation(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	mockClient := &orderRisksTestClient{}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	// Mock stdin to simulate user not confirming
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write "n" to stdin to simulate user declining
	go func() {
		_, _ = w.WriteString("n\n")
		_ = w.Close()
	}()

	cmd := newOrderRisksTestCmd()
	// Don't set --yes flag to trigger confirmation prompt

	err := orderRisksDeleteCmd.RunE(cmd, []string{"order_456", "risk_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestOrderRisksDeleteWithConfirmation tests the delete command proceeds when confirmed with "y".
func TestOrderRisksDeleteWithConfirmation(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	mockClient := &orderRisksTestClient{}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	// Mock stdin to simulate user confirming with "y"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write "y" to stdin to simulate user confirming
	go func() {
		_, _ = w.WriteString("y\n")
		_ = w.Close()
	}()

	cmd := newOrderRisksTestCmd()
	// Don't set --yes flag to trigger confirmation prompt

	err := orderRisksDeleteCmd.RunE(cmd, []string{"order_456", "risk_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestOrderRisksDeleteWithUppercaseConfirmation tests the delete command proceeds when confirmed with "Y".
func TestOrderRisksDeleteWithUppercaseConfirmation(t *testing.T) {
	cleanup := setupOrderRisksTest(t)
	defer cleanup()

	mockClient := &orderRisksTestClient{}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	// Mock stdin to simulate user confirming with "Y"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write "Y" to stdin to simulate user confirming
	go func() {
		_, _ = w.WriteString("Y\n")
		_ = w.Close()
	}()

	cmd := newOrderRisksTestCmd()
	// Don't set --yes flag to trigger confirmation prompt

	err := orderRisksDeleteCmd.RunE(cmd, []string{"order_456", "risk_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestOrderRisksGetClientError tests error handling when client creation fails.
func TestOrderRisksGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }

	cmd := newTestCmdWithFlags()
	if err := orderRisksGetCmd.RunE(cmd, []string{"order_123", "risk_456"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestOrderRisksCreateClientError tests error handling when client creation fails.
func TestOrderRisksCreateClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }

	cmd := newTestCmdWithFlags()
	if err := orderRisksCreateCmd.RunE(cmd, []string{"order_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestOrderRisksDeleteClientError tests error handling when client creation fails.
func TestOrderRisksDeleteClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }

	cmd := newTestCmdWithFlags()
	if err := orderRisksDeleteCmd.RunE(cmd, []string{"order_123", "risk_456"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestOrderRisksGetNoProfiles tests get command with no profiles configured.
func TestOrderRisksGetNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := orderRisksGetCmd.RunE(cmd, []string{"order_123", "risk_456"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestOrderRisksCreateNoProfiles tests create command with no profiles configured.
func TestOrderRisksCreateNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := orderRisksCreateCmd.RunE(cmd, []string{"order_123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestOrderRisksDeleteNoProfiles tests delete command with no profiles configured.
func TestOrderRisksDeleteNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := orderRisksDeleteCmd.RunE(cmd, []string{"order_123", "risk_456"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}
