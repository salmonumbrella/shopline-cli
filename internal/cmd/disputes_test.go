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

func TestDisputesCommandStructure(t *testing.T) {
	if disputesCmd == nil {
		t.Fatal("disputesCmd is nil")
	}
	if disputesCmd.Use != "disputes" {
		t.Errorf("Expected Use 'disputes', got %q", disputesCmd.Use)
	}
	subcommands := map[string]bool{"list": false, "get": false, "submit": false, "accept": false, "evidence": false}
	for _, cmd := range disputesCmd.Commands() {
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

func TestDisputesListFlags(t *testing.T) {
	cmd := disputesListCmd
	flags := []struct{ name, defaultValue string }{{"page", "1"}, {"page-size", "20"}, {"status", ""}, {"reason", ""}}
	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("Flag %q not found", f.name)
		} else if flag.DefValue != f.defaultValue {
			t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestDisputesGetRequiresArg(t *testing.T) {
	cmd := disputesGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"dispute_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
	if cmd.Args(cmd, []string{"dispute_123", "extra"}) == nil {
		t.Error("Expected error with 2 args")
	}
}

func TestDisputesSubmitRequiresArg(t *testing.T) {
	cmd := disputesSubmitCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"dispute_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
	if cmd.Args(cmd, []string{"dispute_123", "extra"}) == nil {
		t.Error("Expected error with 2 args")
	}
}

func TestDisputesAcceptRequiresArg(t *testing.T) {
	cmd := disputesAcceptCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"dispute_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
	if cmd.Args(cmd, []string{"dispute_123", "extra"}) == nil {
		t.Error("Expected error with 2 args")
	}
}

func TestDisputesEvidenceRequiresArg(t *testing.T) {
	cmd := disputesEvidenceCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"dispute_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
	if cmd.Args(cmd, []string{"dispute_123", "extra"}) == nil {
		t.Error("Expected error with 2 args")
	}
}

func TestDisputesEvidenceFlags(t *testing.T) {
	cmd := disputesEvidenceCmd
	flags := []string{
		"customer-name",
		"customer-email",
		"product-description",
		"shipping-carrier",
		"tracking-number",
		"shipping-date",
	}
	for _, name := range flags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("Flag %q not found", name)
		}
	}
}

func TestDisputesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := disputesListCmd.RunE(cmd, []string{}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDisputesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := disputesGetCmd.RunE(cmd, []string{"dispute_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDisputesSubmitGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true") // Use existing flag from newTestCmdWithFlags
	if err := disputesSubmitCmd.RunE(cmd, []string{"dispute_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDisputesAcceptGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true") // Use existing flag from newTestCmdWithFlags
	if err := disputesAcceptCmd.RunE(cmd, []string{"dispute_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDisputesEvidenceGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-name", "", "")
	cmd.Flags().String("customer-email", "", "")
	cmd.Flags().String("product-description", "", "")
	cmd.Flags().String("shipping-carrier", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().String("shipping-date", "", "")
	if err := disputesEvidenceCmd.RunE(cmd, []string{"dispute_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDisputesListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := disputesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestDisputesGetNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := disputesGetCmd.RunE(cmd, []string{"dispute_123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestDisputesListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() { secretsStoreFactory = origFactory; _ = os.Setenv("SHOPLINE_STORE", origEnv) }()
	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{"envstore", "other"}, creds: map[string]*secrets.StoreCredentials{"envstore": {Handle: "test", AccessToken: "token123"}}}, nil
	}
	cmd := newTestCmdWithFlags()
	err := disputesListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// disputesTestClient is a mock implementation for disputes testing.
type disputesTestClient struct {
	api.MockClient

	listDisputesResp          *api.DisputesListResponse
	listDisputesErr           error
	getDisputeResp            *api.Dispute
	getDisputeErr             error
	submitDisputeResp         *api.Dispute
	submitDisputeErr          error
	acceptDisputeResp         *api.Dispute
	acceptDisputeErr          error
	updateDisputeEvidenceResp *api.Dispute
	updateDisputeEvidenceErr  error
	updateDisputeEvidenceReq  *api.DisputeUpdateEvidenceRequest
}

func (m *disputesTestClient) ListDisputes(ctx context.Context, opts *api.DisputesListOptions) (*api.DisputesListResponse, error) {
	return m.listDisputesResp, m.listDisputesErr
}

func (m *disputesTestClient) GetDispute(ctx context.Context, id string) (*api.Dispute, error) {
	return m.getDisputeResp, m.getDisputeErr
}

func (m *disputesTestClient) SubmitDispute(ctx context.Context, id string) (*api.Dispute, error) {
	return m.submitDisputeResp, m.submitDisputeErr
}

func (m *disputesTestClient) AcceptDispute(ctx context.Context, id string) (*api.Dispute, error) {
	return m.acceptDisputeResp, m.acceptDisputeErr
}

func (m *disputesTestClient) UpdateDisputeEvidence(ctx context.Context, id string, req *api.DisputeUpdateEvidenceRequest) (*api.Dispute, error) {
	m.updateDisputeEvidenceReq = req
	return m.updateDisputeEvidenceResp, m.updateDisputeEvidenceErr
}

// setupDisputesTest sets up mocks for disputes testing and returns cleanup function.
func setupDisputesTest(mockClient *disputesTestClient) func() {
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

	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// TestDisputesListRunE tests the disputes list command execution with mock API.
func TestDisputesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.DisputesListResponse
		mockErr    error
		outputFmt  string
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with text output",
			mockResp: &api.DisputesListResponse{
				Items: []api.Dispute{
					{
						ID:        "dispute_123",
						OrderID:   "ord_456",
						Amount:    "100.00",
						Currency:  "USD",
						Status:    "open",
						Reason:    "fraudulent",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFmt:  "text",
			wantOutput: "dispute_123",
		},
		{
			name: "successful list with JSON output",
			mockResp: &api.DisputesListResponse{
				Items: []api.Dispute{
					{
						ID:        "dispute_json",
						OrderID:   "ord_json",
						Amount:    "200.00",
						Currency:  "EUR",
						Status:    "needs_response",
						Reason:    "product_not_received",
						CreatedAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFmt:  "json",
			wantOutput: "dispute_json",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.DisputesListResponse{
				Items:      []api.Dispute{},
				TotalCount: 0,
			},
			outputFmt: "text",
		},
		{
			name: "multiple disputes",
			mockResp: &api.DisputesListResponse{
				Items: []api.Dispute{
					{
						ID:        "dispute_1",
						OrderID:   "ord_1",
						Amount:    "50.00",
						Currency:  "USD",
						Status:    "open",
						Reason:    "fraudulent",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:        "dispute_2",
						OrderID:   "ord_2",
						Amount:    "75.00",
						Currency:  "GBP",
						Status:    "under_review",
						Reason:    "duplicate",
						CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			outputFmt:  "text",
			wantOutput: "dispute_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &disputesTestClient{
				listDisputesResp: tt.mockResp,
				listDisputesErr:  tt.mockErr,
			}

			cleanup := setupDisputesTest(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFmt, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("reason", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := disputesListCmd.RunE(cmd, []string{})

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

// TestDisputesGetRunE tests the disputes get command execution with mock API.
func TestDisputesGetRunE(t *testing.T) {
	evidenceDue := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	resolvedAt := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		disputeID string
		mockResp  *api.Dispute
		mockErr   error
		outputFmt string
		wantErr   bool
	}{
		{
			name:      "successful get with text output",
			disputeID: "dispute_123",
			mockResp: &api.Dispute{
				ID:        "dispute_123",
				OrderID:   "ord_456",
				PaymentID: "pay_789",
				Amount:    "100.00",
				Currency:  "USD",
				Status:    "open",
				Reason:    "fraudulent",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:      "successful get with JSON output",
			disputeID: "dispute_json",
			mockResp: &api.Dispute{
				ID:        "dispute_json",
				OrderID:   "ord_json",
				PaymentID: "pay_json",
				Amount:    "200.00",
				Currency:  "EUR",
				Status:    "needs_response",
				Reason:    "product_not_received",
				CreatedAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 2, 21, 15, 0, 0, 0, time.UTC),
			},
			outputFmt: "json",
		},
		{
			name:      "get with network reason code",
			disputeID: "dispute_network",
			mockResp: &api.Dispute{
				ID:                "dispute_network",
				OrderID:           "ord_network",
				PaymentID:         "pay_network",
				Amount:            "150.00",
				Currency:          "USD",
				Status:            "under_review",
				Reason:            "fraudulent",
				NetworkReasonCode: "4837",
				CreatedAt:         time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC),
				UpdatedAt:         time.Date(2024, 1, 21, 11, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:      "get with evidence due by",
			disputeID: "dispute_evidence_due",
			mockResp: &api.Dispute{
				ID:            "dispute_evidence_due",
				OrderID:       "ord_evidence",
				PaymentID:     "pay_evidence",
				Amount:        "75.00",
				Currency:      "GBP",
				Status:        "needs_response",
				Reason:        "duplicate",
				EvidenceDueBy: &evidenceDue,
				CreatedAt:     time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:      "get with evidence",
			disputeID: "dispute_with_evidence",
			mockResp: &api.Dispute{
				ID:        "dispute_with_evidence",
				OrderID:   "ord_evidence",
				PaymentID: "pay_evidence",
				Amount:    "50.00",
				Currency:  "USD",
				Status:    "under_review",
				Reason:    "product_not_received",
				Evidence: &api.DisputeEvidence{
					CustomerName:           "John Doe",
					CustomerEmail:          "john@example.com",
					ShippingCarrier:        "UPS",
					ShippingTrackingNumber: "1Z999AA10123456784",
				},
				CreatedAt: time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 12, 10, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:      "get with resolved at",
			disputeID: "dispute_resolved",
			mockResp: &api.Dispute{
				ID:         "dispute_resolved",
				OrderID:    "ord_resolved",
				PaymentID:  "pay_resolved",
				Amount:     "25.00",
				Currency:   "CAD",
				Status:     "won",
				Reason:     "fraudulent",
				ResolvedAt: &resolvedAt,
				CreatedAt:  time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:      "get with all optional fields",
			disputeID: "dispute_full",
			mockResp: &api.Dispute{
				ID:                "dispute_full",
				OrderID:           "ord_full",
				PaymentID:         "pay_full",
				Amount:            "500.00",
				Currency:          "USD",
				Status:            "lost",
				Reason:            "general",
				NetworkReasonCode: "1234",
				Evidence: &api.DisputeEvidence{
					CustomerName:           "Jane Smith",
					CustomerEmail:          "jane@example.com",
					ShippingCarrier:        "FedEx",
					ShippingTrackingNumber: "794644790200",
				},
				EvidenceDueBy: &evidenceDue,
				ResolvedAt:    &resolvedAt,
				CreatedAt:     time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:      "dispute not found",
			disputeID: "dispute_999",
			mockErr:   errors.New("dispute not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &disputesTestClient{
				getDisputeResp: tt.mockResp,
				getDisputeErr:  tt.mockErr,
			}

			cleanup := setupDisputesTest(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFmt, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := disputesGetCmd.RunE(cmd, []string{tt.disputeID})

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

// TestDisputesSubmitRunE tests the disputes submit command execution with mock API.
func TestDisputesSubmitRunE(t *testing.T) {
	tests := []struct {
		name      string
		disputeID string
		mockResp  *api.Dispute
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful submit",
			disputeID: "dispute_123",
			mockResp: &api.Dispute{
				ID:     "dispute_123",
				Status: "under_review",
			},
		},
		{
			name:      "submit fails",
			disputeID: "dispute_456",
			mockErr:   errors.New("dispute cannot be submitted"),
			wantErr:   true,
		},
		{
			name:      "submit with full response",
			disputeID: "dispute_full",
			mockResp: &api.Dispute{
				ID:        "dispute_full",
				OrderID:   "ord_full",
				PaymentID: "pay_full",
				Amount:    "100.00",
				Currency:  "USD",
				Status:    "under_review",
				Reason:    "fraudulent",
				CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &disputesTestClient{
				submitDisputeResp: tt.mockResp,
				submitDisputeErr:  tt.mockErr,
			}

			cleanup := setupDisputesTest(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := disputesSubmitCmd.RunE(cmd, []string{tt.disputeID})

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

// TestDisputesAcceptRunE tests the disputes accept command execution with mock API.
func TestDisputesAcceptRunE(t *testing.T) {
	tests := []struct {
		name      string
		disputeID string
		mockResp  *api.Dispute
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful accept",
			disputeID: "dispute_123",
			mockResp: &api.Dispute{
				ID:     "dispute_123",
				Status: "lost",
			},
		},
		{
			name:      "accept fails",
			disputeID: "dispute_456",
			mockErr:   errors.New("dispute cannot be accepted"),
			wantErr:   true,
		},
		{
			name:      "accept with full response",
			disputeID: "dispute_full",
			mockResp: &api.Dispute{
				ID:        "dispute_full",
				OrderID:   "ord_full",
				PaymentID: "pay_full",
				Amount:    "200.00",
				Currency:  "EUR",
				Status:    "lost",
				Reason:    "product_not_received",
				CreatedAt: time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 2, 5, 14, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &disputesTestClient{
				acceptDisputeResp: tt.mockResp,
				acceptDisputeErr:  tt.mockErr,
			}

			cleanup := setupDisputesTest(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := disputesAcceptCmd.RunE(cmd, []string{tt.disputeID})

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

// TestDisputesEvidenceRunE tests the disputes evidence command execution with mock API.
func TestDisputesEvidenceRunE(t *testing.T) {
	tests := []struct {
		name            string
		disputeID       string
		customerName    string
		customerEmail   string
		productDesc     string
		shippingCarrier string
		trackingNumber  string
		shippingDate    string
		mockResp        *api.Dispute
		mockErr         error
		wantErr         bool
	}{
		{
			name:          "successful update with customer info",
			disputeID:     "dispute_123",
			customerName:  "John Doe",
			customerEmail: "john@example.com",
			mockResp: &api.Dispute{
				ID:     "dispute_123",
				Status: "needs_response",
			},
		},
		{
			name:            "successful update with shipping info",
			disputeID:       "dispute_456",
			shippingCarrier: "UPS",
			trackingNumber:  "1Z999AA10123456784",
			shippingDate:    "2024-01-15",
			mockResp: &api.Dispute{
				ID:     "dispute_456",
				Status: "needs_response",
			},
		},
		{
			name:            "successful update with all fields",
			disputeID:       "dispute_789",
			customerName:    "Jane Smith",
			customerEmail:   "jane@example.com",
			productDesc:     "Widget XL - Blue",
			shippingCarrier: "FedEx",
			trackingNumber:  "794644790200",
			shippingDate:    "2024-02-01",
			mockResp: &api.Dispute{
				ID:        "dispute_789",
				OrderID:   "ord_789",
				PaymentID: "pay_789",
				Amount:    "150.00",
				Currency:  "USD",
				Status:    "needs_response",
				Reason:    "product_not_received",
				Evidence: &api.DisputeEvidence{
					CustomerName:           "Jane Smith",
					CustomerEmail:          "jane@example.com",
					ProductDescription:     "Widget XL - Blue",
					ShippingCarrier:        "FedEx",
					ShippingTrackingNumber: "794644790200",
					ShippingDate:           "2024-02-01",
				},
				CreatedAt: time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "update fails",
			disputeID: "dispute_error",
			mockErr:   errors.New("failed to update evidence"),
			wantErr:   true,
		},
		{
			name:      "update with no flags",
			disputeID: "dispute_empty",
			mockResp: &api.Dispute{
				ID:     "dispute_empty",
				Status: "needs_response",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &disputesTestClient{
				updateDisputeEvidenceResp: tt.mockResp,
				updateDisputeEvidenceErr:  tt.mockErr,
			}

			cleanup := setupDisputesTest(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("customer-name", tt.customerName, "")
			cmd.Flags().String("customer-email", tt.customerEmail, "")
			cmd.Flags().String("product-description", tt.productDesc, "")
			cmd.Flags().String("shipping-carrier", tt.shippingCarrier, "")
			cmd.Flags().String("tracking-number", tt.trackingNumber, "")
			cmd.Flags().String("shipping-date", tt.shippingDate, "")

			err := disputesEvidenceCmd.RunE(cmd, []string{tt.disputeID})

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

// TestDisputesListAPIErrorMessage verifies error wrapping in list command
func TestDisputesListAPIErrorMessage(t *testing.T) {
	mockClient := &disputesTestClient{
		listDisputesErr: errors.New("connection refused"),
	}

	cleanup := setupDisputesTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("reason", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := disputesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "failed to list disputes") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestDisputesGetAPIErrorMessage verifies error wrapping in get command
func TestDisputesGetAPIErrorMessage(t *testing.T) {
	mockClient := &disputesTestClient{
		getDisputeErr: errors.New("not found"),
	}

	cleanup := setupDisputesTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := disputesGetCmd.RunE(cmd, []string{"dispute_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "failed to get dispute") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestDisputesSubmitAPIErrorMessage verifies error wrapping in submit command
func TestDisputesSubmitAPIErrorMessage(t *testing.T) {
	mockClient := &disputesTestClient{
		submitDisputeErr: errors.New("cannot submit"),
	}

	cleanup := setupDisputesTest(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", true, "")

	err := disputesSubmitCmd.RunE(cmd, []string{"dispute_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "failed to submit dispute") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestDisputesAcceptAPIErrorMessage verifies error wrapping in accept command
func TestDisputesAcceptAPIErrorMessage(t *testing.T) {
	mockClient := &disputesTestClient{
		acceptDisputeErr: errors.New("cannot accept"),
	}

	cleanup := setupDisputesTest(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", true, "")

	err := disputesAcceptCmd.RunE(cmd, []string{"dispute_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "failed to accept dispute") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestDisputesEvidenceAPIErrorMessage verifies error wrapping in evidence command
func TestDisputesEvidenceAPIErrorMessage(t *testing.T) {
	mockClient := &disputesTestClient{
		updateDisputeEvidenceErr: errors.New("cannot update"),
	}

	cleanup := setupDisputesTest(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("customer-name", "", "")
	cmd.Flags().String("customer-email", "", "")
	cmd.Flags().String("product-description", "", "")
	cmd.Flags().String("shipping-carrier", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().String("shipping-date", "", "")

	err := disputesEvidenceCmd.RunE(cmd, []string{"dispute_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "failed to update dispute evidence") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestDisputesSubmitNoProfiles verifies error handling when no profiles exist for submit command
func TestDisputesSubmitNoProfiles(t *testing.T) {
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
	_ = cmd.Flags().Set("yes", "true") // Use existing flag from newTestCmdWithFlags
	err := disputesSubmitCmd.RunE(cmd, []string{"dispute_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestDisputesAcceptNoProfiles verifies error handling when no profiles exist for accept command
func TestDisputesAcceptNoProfiles(t *testing.T) {
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
	_ = cmd.Flags().Set("yes", "true") // Use existing flag from newTestCmdWithFlags
	err := disputesAcceptCmd.RunE(cmd, []string{"dispute_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestDisputesEvidenceNoProfiles verifies error handling when no profiles exist for evidence command
func TestDisputesEvidenceNoProfiles(t *testing.T) {
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
	cmd.Flags().String("customer-name", "", "")
	cmd.Flags().String("customer-email", "", "")
	cmd.Flags().String("product-description", "", "")
	cmd.Flags().String("shipping-carrier", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().String("shipping-date", "", "")
	err := disputesEvidenceCmd.RunE(cmd, []string{"dispute_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestDisputesSubmitCancelled tests cancellation when user doesn't confirm submit.
func TestDisputesSubmitCancelled(t *testing.T) {
	mockClient := &disputesTestClient{
		submitDisputeResp: &api.Dispute{
			ID:     "dispute_123",
			Status: "under_review",
		},
	}

	cleanup := setupDisputesTest(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", false, "") // Not skipping confirmation

	// Since Scanln will fail or return empty, the command should print "Cancelled."
	err := disputesSubmitCmd.RunE(cmd, []string{"dispute_123"})
	// The command should succeed (cancellation is not an error)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDisputesAcceptCancelled tests cancellation when user doesn't confirm accept.
func TestDisputesAcceptCancelled(t *testing.T) {
	mockClient := &disputesTestClient{
		acceptDisputeResp: &api.Dispute{
			ID:     "dispute_123",
			Status: "lost",
		},
	}

	cleanup := setupDisputesTest(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", false, "") // Not skipping confirmation

	// Since Scanln will fail or return empty, the command should print "Cancelled."
	err := disputesAcceptCmd.RunE(cmd, []string{"dispute_123"})
	// The command should succeed (cancellation is not an error)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
