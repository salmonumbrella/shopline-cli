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

func TestCompanyCreditsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := companyCreditsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCreditsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := companyCreditsGetCmd.RunE(cmd, []string{"credit-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCreditsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("company-id", "comp-123", "")
	cmd.Flags().Float64("credit-limit", 10000.00, "")
	cmd.Flags().String("currency", "USD", "")

	err := companyCreditsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCreditsAdjustGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Float64("amount", 500.00, "")
	cmd.Flags().String("description", "Adjustment", "")
	cmd.Flags().String("reference-id", "", "")

	err := companyCreditsAdjustCmd.RunE(cmd, []string{"credit-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCreditsTransactionsGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := companyCreditsTransactionsCmd.RunE(cmd, []string{"credit-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCreditsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	// yes flag already added by newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := companyCreditsDeleteCmd.RunE(cmd, []string{"credit-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCreditsListFlags(t *testing.T) {
	flags := companyCreditsListCmd.Flags()

	if flags.Lookup("page") == nil {
		t.Error("Expected page flag")
	}
	if flags.Lookup("page-size") == nil {
		t.Error("Expected page-size flag")
	}
	if flags.Lookup("company-id") == nil {
		t.Error("Expected company-id flag")
	}
	if flags.Lookup("status") == nil {
		t.Error("Expected status flag")
	}
}

func TestCompanyCreditsCommandStructure(t *testing.T) {
	if companyCreditsCmd.Use != "company-credits" {
		t.Errorf("Expected Use 'company-credits', got %s", companyCreditsCmd.Use)
	}

	subcommands := companyCreditsCmd.Commands()
	expectedCmds := map[string]bool{
		"list":         false,
		"get":          false,
		"create":       false,
		"adjust":       false,
		"transactions": false,
		"delete":       false,
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

// companyCreditsTestClient is a mock implementation for company credits testing.
type companyCreditsTestClient struct {
	api.MockClient

	listCompanyCreditsResp            *api.CompanyCreditsListResponse
	listCompanyCreditsErr             error
	getCompanyCreditResp              *api.CompanyCredit
	getCompanyCreditErr               error
	createCompanyCreditResp           *api.CompanyCredit
	createCompanyCreditErr            error
	adjustCompanyCreditResp           *api.CompanyCredit
	adjustCompanyCreditErr            error
	listCompanyCreditTransactionsResp *api.CompanyCreditTransactionsListResponse
	listCompanyCreditTransactionsErr  error
	deleteCompanyCreditErr            error
}

func (m *companyCreditsTestClient) ListCompanyCredits(ctx context.Context, opts *api.CompanyCreditsListOptions) (*api.CompanyCreditsListResponse, error) {
	return m.listCompanyCreditsResp, m.listCompanyCreditsErr
}

func (m *companyCreditsTestClient) GetCompanyCredit(ctx context.Context, id string) (*api.CompanyCredit, error) {
	return m.getCompanyCreditResp, m.getCompanyCreditErr
}

func (m *companyCreditsTestClient) CreateCompanyCredit(ctx context.Context, req *api.CompanyCreditCreateRequest) (*api.CompanyCredit, error) {
	return m.createCompanyCreditResp, m.createCompanyCreditErr
}

func (m *companyCreditsTestClient) AdjustCompanyCredit(ctx context.Context, id string, req *api.CompanyCreditAdjustRequest) (*api.CompanyCredit, error) {
	return m.adjustCompanyCreditResp, m.adjustCompanyCreditErr
}

func (m *companyCreditsTestClient) ListCompanyCreditTransactions(ctx context.Context, creditID string, page, pageSize int) (*api.CompanyCreditTransactionsListResponse, error) {
	return m.listCompanyCreditTransactionsResp, m.listCompanyCreditTransactionsErr
}

func (m *companyCreditsTestClient) DeleteCompanyCredit(ctx context.Context, id string) error {
	return m.deleteCompanyCreditErr
}

// TestCompanyCreditsListRunE tests the company credits list command execution with mock API.
func TestCompanyCreditsListRunE(t *testing.T) {
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
		mockResp   *api.CompanyCreditsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CompanyCreditsListResponse{
				Items: []api.CompanyCredit{
					{
						ID:            "cc_123",
						CompanyID:     "comp_456",
						CompanyName:   "Acme Corp",
						CreditBalance: 5000.00,
						CreditLimit:   10000.00,
						Currency:      "USD",
						Status:        "active",
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "cc_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CompanyCreditsListResponse{
				Items:      []api.CompanyCredit{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCreditsTestClient{
				listCompanyCreditsResp: tt.mockResp,
				listCompanyCreditsErr:  tt.mockErr,
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
			cmd.Flags().String("company-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := companyCreditsListCmd.RunE(cmd, []string{})

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

// TestCompanyCreditsGetRunE tests the company credits get command execution with mock API.
func TestCompanyCreditsGetRunE(t *testing.T) {
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
		creditID string
		mockResp *api.CompanyCredit
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful get",
			creditID: "cc_123",
			mockResp: &api.CompanyCredit{
				ID:            "cc_123",
				CompanyID:     "comp_456",
				CompanyName:   "Acme Corp",
				CreditBalance: 5000.00,
				CreditLimit:   10000.00,
				Currency:      "USD",
				Status:        "active",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:     "credit not found",
			creditID: "cc_999",
			mockErr:  errors.New("company credit not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCreditsTestClient{
				getCompanyCreditResp: tt.mockResp,
				getCompanyCreditErr:  tt.mockErr,
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

			err := companyCreditsGetCmd.RunE(cmd, []string{tt.creditID})

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

// TestCompanyCreditsCreateRunE tests the company credits create command execution with mock API.
func TestCompanyCreditsCreateRunE(t *testing.T) {
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
		mockResp *api.CompanyCredit
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.CompanyCredit{
				ID:          "cc_new",
				CompanyID:   "comp_123",
				CreditLimit: 10000.00,
				Currency:    "USD",
				Status:      "active",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("failed to create company credit"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCreditsTestClient{
				createCompanyCreditResp: tt.mockResp,
				createCompanyCreditErr:  tt.mockErr,
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
			cmd.Flags().String("company-id", "comp_123", "")
			cmd.Flags().Float64("credit-limit", 10000.00, "")
			cmd.Flags().String("currency", "USD", "")

			err := companyCreditsCreateCmd.RunE(cmd, []string{})

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

// TestCompanyCreditsAdjustRunE tests the company credits adjust command execution with mock API.
func TestCompanyCreditsAdjustRunE(t *testing.T) {
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
		creditID string
		mockResp *api.CompanyCredit
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful adjust",
			creditID: "cc_123",
			mockResp: &api.CompanyCredit{
				ID:            "cc_123",
				CompanyID:     "comp_456",
				CreditBalance: 5500.00,
				CreditLimit:   10000.00,
				Currency:      "USD",
			},
		},
		{
			name:     "adjust fails",
			creditID: "cc_456",
			mockErr:  errors.New("failed to adjust company credit"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCreditsTestClient{
				adjustCompanyCreditResp: tt.mockResp,
				adjustCompanyCreditErr:  tt.mockErr,
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
			cmd.Flags().Float64("amount", 500.00, "")
			cmd.Flags().String("description", "Adjustment", "")
			cmd.Flags().String("reference-id", "", "")

			err := companyCreditsAdjustCmd.RunE(cmd, []string{tt.creditID})

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

// TestCompanyCreditsTransactionsRunE tests the company credit transactions command execution with mock API.
func TestCompanyCreditsTransactionsRunE(t *testing.T) {
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
		creditID   string
		mockResp   *api.CompanyCreditTransactionsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:     "successful list",
			creditID: "cc_123",
			mockResp: &api.CompanyCreditTransactionsListResponse{
				Items: []api.CompanyCreditTransaction{
					{
						ID:          "tx_123",
						Type:        "credit",
						Amount:      500.00,
						Balance:     5500.00,
						Description: "Credit adjustment",
						ReferenceID: "ref_001",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "tx_123",
		},
		{
			name:     "API error",
			creditID: "cc_456",
			mockErr:  errors.New("API unavailable"),
			wantErr:  true,
		},
		{
			name:     "empty list",
			creditID: "cc_789",
			mockResp: &api.CompanyCreditTransactionsListResponse{
				Items:      []api.CompanyCreditTransaction{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCreditsTestClient{
				listCompanyCreditTransactionsResp: tt.mockResp,
				listCompanyCreditTransactionsErr:  tt.mockErr,
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
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := companyCreditsTransactionsCmd.RunE(cmd, []string{tt.creditID})

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

// TestCompanyCreditsDeleteRunE tests the company credits delete command execution with mock API.
func TestCompanyCreditsDeleteRunE(t *testing.T) {
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
		name     string
		creditID string
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful delete",
			creditID: "cc_123",
			mockErr:  nil,
		},
		{
			name:     "delete fails",
			creditID: "cc_456",
			mockErr:  errors.New("company credit not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCreditsTestClient{
				deleteCompanyCreditErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := companyCreditsDeleteCmd.RunE(cmd, []string{tt.creditID})

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
