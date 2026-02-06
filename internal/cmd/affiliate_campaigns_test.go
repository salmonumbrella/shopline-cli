package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// affiliateCampaignsMockAPIClient is a mock implementation for affiliate campaigns tests.
type affiliateCampaignsMockAPIClient struct {
	api.MockClient
	listAffiliateCampaignsResp  *api.AffiliateCampaignsListResponse
	listAffiliateCampaignsErr   error
	getAffiliateCampaignResp    *api.AffiliateCampaign
	getAffiliateCampaignErr     error
	createAffiliateCampaignResp *api.AffiliateCampaign
	createAffiliateCampaignErr  error
	deleteAffiliateCampaignErr  error

	getOrdersResp  json.RawMessage
	getOrdersErr   error
	getSummaryResp json.RawMessage
	getSummaryErr  error
	getRankingResp json.RawMessage
	getRankingErr  error
	exportResp     json.RawMessage
	exportErr      error
}

func (m *affiliateCampaignsMockAPIClient) ListAffiliateCampaigns(ctx context.Context, opts *api.AffiliateCampaignsListOptions) (*api.AffiliateCampaignsListResponse, error) {
	return m.listAffiliateCampaignsResp, m.listAffiliateCampaignsErr
}

func (m *affiliateCampaignsMockAPIClient) GetAffiliateCampaign(ctx context.Context, id string) (*api.AffiliateCampaign, error) {
	return m.getAffiliateCampaignResp, m.getAffiliateCampaignErr
}

func (m *affiliateCampaignsMockAPIClient) CreateAffiliateCampaign(ctx context.Context, req *api.AffiliateCampaignCreateRequest) (*api.AffiliateCampaign, error) {
	return m.createAffiliateCampaignResp, m.createAffiliateCampaignErr
}

func (m *affiliateCampaignsMockAPIClient) DeleteAffiliateCampaign(ctx context.Context, id string) error {
	return m.deleteAffiliateCampaignErr
}

func (m *affiliateCampaignsMockAPIClient) GetAffiliateCampaignOrders(ctx context.Context, id string, opts *api.AffiliateCampaignOrdersOptions) (json.RawMessage, error) {
	return m.getOrdersResp, m.getOrdersErr
}

func (m *affiliateCampaignsMockAPIClient) GetAffiliateCampaignSummary(ctx context.Context, id string) (json.RawMessage, error) {
	return m.getSummaryResp, m.getSummaryErr
}

func (m *affiliateCampaignsMockAPIClient) GetAffiliateCampaignProductsSalesRanking(ctx context.Context, id string, opts *api.AffiliateCampaignProductsSalesRankingOptions) (json.RawMessage, error) {
	return m.getRankingResp, m.getRankingErr
}

func (m *affiliateCampaignsMockAPIClient) ExportAffiliateCampaignReport(ctx context.Context, id string, body any) (json.RawMessage, error) {
	return m.exportResp, m.exportErr
}

// setupAffiliateCampaignsMockFactories sets up mock factories for affiliate campaigns tests.
func setupAffiliateCampaignsMockFactories(mockClient *affiliateCampaignsMockAPIClient) (func(), *bytes.Buffer) {
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

// newAffiliateCampaignsTestCmd creates a test command with common flags for affiliate campaigns tests.
func newAffiliateCampaignsTestCmd() *cobra.Command {
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

// TestAffiliateCampaignsListRunE tests the list command with mock API.
func TestAffiliateCampaignsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.AffiliateCampaignsListResponse
		mockErr    error
		output     string
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful list with percentage commission",
			output: "text",
			mockResp: &api.AffiliateCampaignsListResponse{
				Items: []api.AffiliateCampaign{
					{
						ID:              "camp_123",
						Name:            "Summer Sale Campaign",
						Status:          "active",
						CommissionType:  "percentage",
						CommissionValue: 10.5,
						TotalSales:      150,
						CreatedAt:       time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "camp_123",
		},
		{
			name:   "successful list with fixed commission",
			output: "text",
			mockResp: &api.AffiliateCampaignsListResponse{
				Items: []api.AffiliateCampaign{
					{
						ID:              "camp_456",
						Name:            "Fixed Bonus Campaign",
						Status:          "paused",
						CommissionType:  "fixed",
						CommissionValue: 25.00,
						TotalSales:      75,
						CreatedAt:       time.Date(2024, 5, 1, 8, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "camp_456",
		},
		{
			name:   "successful list JSON output",
			output: "json",
			mockResp: &api.AffiliateCampaignsListResponse{
				Items: []api.AffiliateCampaign{
					{
						ID:              "camp_789",
						Name:            "JSON Campaign",
						Status:          "active",
						CommissionType:  "percentage",
						CommissionValue: 5.0,
						TotalSales:      200,
						CreatedAt:       time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "camp_789",
		},
		{
			name:    "API error",
			output:  "text",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name:   "empty list",
			output: "text",
			mockResp: &api.AffiliateCampaignsListResponse{
				Items:      []api.AffiliateCampaign{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &affiliateCampaignsMockAPIClient{
				listAffiliateCampaignsResp: tt.mockResp,
				listAffiliateCampaignsErr:  tt.mockErr,
			}
			cleanup, buf := setupAffiliateCampaignsMockFactories(mockClient)
			defer cleanup()

			cmd := newAffiliateCampaignsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("status", "", "")
			_ = cmd.Flags().Set("output", tt.output)

			err := affiliateCampaignsListCmd.RunE(cmd, []string{})

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

// TestAffiliateCampaignsGetRunE tests the get command with mock API.
func TestAffiliateCampaignsGetRunE(t *testing.T) {
	baseTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		campaignID string
		output     string
		mockResp   *api.AffiliateCampaign
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful get with all fields",
			campaignID: "camp_123",
			output:     "text",
			mockResp: &api.AffiliateCampaign{
				ID:              "camp_123",
				Name:            "Full Campaign",
				Description:     "A comprehensive campaign description",
				Status:          "active",
				CommissionType:  "percentage",
				CommissionValue: 15.0,
				TotalClicks:     5000,
				TotalSales:      250,
				TotalRevenue:    12500.50,
				StartDate:       startDate,
				EndDate:         endDate,
				CreatedAt:       baseTime,
				UpdatedAt:       baseTime.Add(24 * time.Hour),
			},
		},
		{
			name:       "successful get without optional fields",
			campaignID: "camp_456",
			output:     "text",
			mockResp: &api.AffiliateCampaign{
				ID:              "camp_456",
				Name:            "Minimal Campaign",
				Status:          "paused",
				CommissionType:  "fixed",
				CommissionValue: 10.0,
				TotalClicks:     100,
				TotalSales:      10,
				TotalRevenue:    500.00,
				CreatedAt:       baseTime,
				UpdatedAt:       baseTime,
			},
		},
		{
			name:       "successful get JSON output",
			campaignID: "camp_789",
			output:     "json",
			mockResp: &api.AffiliateCampaign{
				ID:              "camp_789",
				Name:            "JSON Campaign",
				Status:          "active",
				CommissionType:  "percentage",
				CommissionValue: 8.5,
				TotalClicks:     1000,
				TotalSales:      50,
				TotalRevenue:    2500.00,
				CreatedAt:       baseTime,
				UpdatedAt:       baseTime,
			},
		},
		{
			name:       "campaign not found",
			campaignID: "camp_999",
			output:     "text",
			mockErr:    errors.New("campaign not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &affiliateCampaignsMockAPIClient{
				getAffiliateCampaignResp: tt.mockResp,
				getAffiliateCampaignErr:  tt.mockErr,
			}
			cleanup, _ := setupAffiliateCampaignsMockFactories(mockClient)
			defer cleanup()

			cmd := newAffiliateCampaignsTestCmd()
			_ = cmd.Flags().Set("output", tt.output)

			err := affiliateCampaignsGetCmd.RunE(cmd, []string{tt.campaignID})

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

// TestAffiliateCampaignsCreateRunE tests the create command with mock API.
func TestAffiliateCampaignsCreateRunE(t *testing.T) {
	tests := []struct {
		name            string
		campaignName    string
		description     string
		commissionType  string
		commissionValue float64
		output          string
		mockResp        *api.AffiliateCampaign
		mockErr         error
		wantErr         bool
	}{
		{
			name:            "successful create percentage commission",
			campaignName:    "New Campaign",
			description:     "A new affiliate campaign",
			commissionType:  "percentage",
			commissionValue: 12.5,
			output:          "text",
			mockResp: &api.AffiliateCampaign{
				ID:              "camp_new_123",
				Name:            "New Campaign",
				Description:     "A new affiliate campaign",
				Status:          "active",
				CommissionType:  "percentage",
				CommissionValue: 12.5,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
		{
			name:            "successful create fixed commission",
			campaignName:    "Fixed Campaign",
			description:     "",
			commissionType:  "fixed",
			commissionValue: 50.00,
			output:          "text",
			mockResp: &api.AffiliateCampaign{
				ID:              "camp_new_456",
				Name:            "Fixed Campaign",
				Status:          "active",
				CommissionType:  "fixed",
				CommissionValue: 50.00,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
		{
			name:            "successful create JSON output",
			campaignName:    "JSON Campaign",
			description:     "Created via CLI",
			commissionType:  "percentage",
			commissionValue: 5.0,
			output:          "json",
			mockResp: &api.AffiliateCampaign{
				ID:              "camp_new_789",
				Name:            "JSON Campaign",
				Description:     "Created via CLI",
				Status:          "active",
				CommissionType:  "percentage",
				CommissionValue: 5.0,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
		{
			name:            "create fails",
			campaignName:    "Fail Campaign",
			commissionType:  "percentage",
			commissionValue: 10.0,
			output:          "text",
			mockErr:         errors.New("validation error: name already exists"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &affiliateCampaignsMockAPIClient{
				createAffiliateCampaignResp: tt.mockResp,
				createAffiliateCampaignErr:  tt.mockErr,
			}
			cleanup, _ := setupAffiliateCampaignsMockFactories(mockClient)
			defer cleanup()

			cmd := newAffiliateCampaignsTestCmd()
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("commission-type", "percentage", "")
			cmd.Flags().Float64("commission-value", 0, "")
			_ = cmd.Flags().Set("name", tt.campaignName)
			_ = cmd.Flags().Set("description", tt.description)
			_ = cmd.Flags().Set("commission-type", tt.commissionType)
			_ = cmd.Flags().Set("commission-value", formatFloat(tt.commissionValue))
			_ = cmd.Flags().Set("output", tt.output)

			err := affiliateCampaignsCreateCmd.RunE(cmd, []string{})

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

// formatFloat formats a float64 for flag setting.
func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(
		strings.Replace(fmt.Sprintf("%f", f), ".", ".", 1),
		"0"), ".")
}

// TestAffiliateCampaignsDeleteRunE tests the delete command with mock API.
func TestAffiliateCampaignsDeleteRunE(t *testing.T) {
	tests := []struct {
		name       string
		campaignID string
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful delete",
			campaignID: "camp_123",
			mockErr:    nil,
		},
		{
			name:       "delete fails",
			campaignID: "camp_456",
			mockErr:    errors.New("campaign not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &affiliateCampaignsMockAPIClient{
				deleteAffiliateCampaignErr: tt.mockErr,
			}
			cleanup, _ := setupAffiliateCampaignsMockFactories(mockClient)
			defer cleanup()

			cmd := newAffiliateCampaignsTestCmd()
			_ = cmd.Flags().Set("yes", "true")

			err := affiliateCampaignsDeleteCmd.RunE(cmd, []string{tt.campaignID})

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

// TestAffiliateCampaignsListGetClientError tests error handling when getClient fails.
func TestAffiliateCampaignsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")

	err := affiliateCampaignsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestAffiliateCampaignsGetGetClientError tests error handling when getClient fails.
func TestAffiliateCampaignsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := affiliateCampaignsGetCmd.RunE(cmd, []string{"campaign-123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestAffiliateCampaignsCreateDryRun tests the create command with dry-run mode.
func TestAffiliateCampaignsCreateDryRun(t *testing.T) {
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
	cmd.Flags().String("name", "Test Campaign", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("commission-type", "percentage", "")
	cmd.Flags().Float64("commission-value", 10.0, "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := affiliateCampaignsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run: %v", err)
	}
}

// TestAffiliateCampaignsCreateGetClientError tests error handling when getClient fails.
func TestAffiliateCampaignsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test Campaign", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("commission-type", "percentage", "")
	cmd.Flags().Float64("commission-value", 10.0, "")

	err := affiliateCampaignsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestAffiliateCampaignsDeleteDryRun tests the delete command with dry-run mode.
func TestAffiliateCampaignsDeleteDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")

	err := affiliateCampaignsDeleteCmd.RunE(cmd, []string{"campaign-123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run: %v", err)
	}
}

// TestAffiliateCampaignsDeleteNoConfirmation tests that delete requires confirmation.
func TestAffiliateCampaignsDeleteNoConfirmation(t *testing.T) {
	cmd := newTestCmdWithFlags()
	// Don't set --yes flag, should print message and return nil

	err := affiliateCampaignsDeleteCmd.RunE(cmd, []string{"campaign-123"})
	if err != nil {
		t.Errorf("unexpected error without confirmation: %v", err)
	}
}

// TestAffiliateCampaignsDeleteGetClientError tests error handling when getClient fails.
func TestAffiliateCampaignsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := affiliateCampaignsDeleteCmd.RunE(cmd, []string{"campaign-123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestAffiliateCampaignsListFlags verifies list command flags exist with correct defaults.
func TestAffiliateCampaignsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := affiliateCampaignsListCmd.Flags().Lookup(f.name)
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

// TestAffiliateCampaignsCreateFlags verifies create command flags.
func TestAffiliateCampaignsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"description", ""},
		{"commission-type", "percentage"},
		{"commission-value", "0"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := affiliateCampaignsCreateCmd.Flags().Lookup(f.name)
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

// TestAffiliateCampaignsDeleteFlags verifies delete command flags.
func TestAffiliateCampaignsDeleteFlags(t *testing.T) {
	flag := affiliateCampaignsDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("flag 'yes' not found")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}

// TestAffiliateCampaignsCommandStructure verifies command structure.
func TestAffiliateCampaignsCommandStructure(t *testing.T) {
	if affiliateCampaignsCmd.Use != "affiliate-campaigns" {
		t.Errorf("expected Use 'affiliate-campaigns', got %s", affiliateCampaignsCmd.Use)
	}

	if affiliateCampaignsCmd.Short != "Manage affiliate marketing campaigns" {
		t.Errorf("expected Short 'Manage affiliate marketing campaigns', got %s", affiliateCampaignsCmd.Short)
	}

	subcommands := affiliateCampaignsCmd.Commands()
	expectedCmds := map[string]bool{
		"list":                   false,
		"get":                    false,
		"create":                 false,
		"delete":                 false,
		"orders":                 false,
		"summary":                false,
		"products-sales-ranking": false,
		"export-report":          false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("expected subcommand %s not found", name)
		}
	}
}

// TestAffiliateCampaignsSubcommandDescriptions verifies subcommand descriptions.
func TestAffiliateCampaignsSubcommandDescriptions(t *testing.T) {
	subcommands := map[string]string{
		"list":                   "List affiliate campaigns",
		"get":                    "Get affiliate campaign details",
		"create":                 "Create an affiliate campaign",
		"delete":                 "Delete an affiliate campaign",
		"orders":                 "Get affiliate campaign orders (documented endpoint; raw JSON)",
		"summary":                "Get affiliate campaign summary (documented endpoint; raw JSON)",
		"products-sales-ranking": "Get products sales ranking of campaign (documented endpoint; raw JSON)",
		"export-report":          "Export affiliate campaign report to partner (documented endpoint; raw JSON body)",
	}

	for use, short := range subcommands {
		t.Run(use, func(t *testing.T) {
			found := false
			for _, cmd := range affiliateCampaignsCmd.Commands() {
				if getBaseUse(cmd.Use) == use {
					found = true
					if cmd.Short != short {
						t.Errorf("expected Short %q, got %q", short, cmd.Short)
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

// TestAffiliateCampaignsGetArgsValidation tests get command argument validation.
func TestAffiliateCampaignsGetArgsValidation(t *testing.T) {
	if affiliateCampaignsGetCmd.Args == nil {
		t.Error("expected Args to be set")
		return
	}

	// Test with wrong number of args
	err := affiliateCampaignsGetCmd.Args(affiliateCampaignsGetCmd, []string{})
	if err == nil {
		t.Error("expected error for missing argument")
	}

	err = affiliateCampaignsGetCmd.Args(affiliateCampaignsGetCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error for too many arguments")
	}

	// Test with correct number of args
	err = affiliateCampaignsGetCmd.Args(affiliateCampaignsGetCmd, []string{"id1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestAffiliateCampaignsDeleteArgsValidation tests delete command argument validation.
func TestAffiliateCampaignsDeleteArgsValidation(t *testing.T) {
	if affiliateCampaignsDeleteCmd.Args == nil {
		t.Error("expected Args to be set")
		return
	}

	// Test with wrong number of args
	err := affiliateCampaignsDeleteCmd.Args(affiliateCampaignsDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error for missing argument")
	}

	err = affiliateCampaignsDeleteCmd.Args(affiliateCampaignsDeleteCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error for too many arguments")
	}

	// Test with correct number of args
	err = affiliateCampaignsDeleteCmd.Args(affiliateCampaignsDeleteCmd, []string{"id1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAffiliateCampaignsOrdersRunE(t *testing.T) {
	mockClient := &affiliateCampaignsMockAPIClient{
		getOrdersResp: json.RawMessage(`{"items":[]}`),
	}
	cleanup, buf := setupAffiliateCampaignsMockFactories(mockClient)
	defer cleanup()

	cmd := newAffiliateCampaignsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	if err := affiliateCampaignsOrdersCmd.RunE(cmd, []string{"camp_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"items\"") {
		t.Fatalf("expected items in output, got %q", buf.String())
	}
}

func TestAffiliateCampaignsSummaryRunE(t *testing.T) {
	mockClient := &affiliateCampaignsMockAPIClient{
		getSummaryResp: json.RawMessage(`{"ok":true}`),
	}
	cleanup, buf := setupAffiliateCampaignsMockFactories(mockClient)
	defer cleanup()

	cmd := newAffiliateCampaignsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	if err := affiliateCampaignsSummaryCmd.RunE(cmd, []string{"camp_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"ok\"") {
		t.Fatalf("expected ok in output, got %q", buf.String())
	}
}

func TestAffiliateCampaignsProductsSalesRankingRunE(t *testing.T) {
	mockClient := &affiliateCampaignsMockAPIClient{
		getRankingResp: json.RawMessage(`{"items":[]}`),
	}
	cleanup, buf := setupAffiliateCampaignsMockFactories(mockClient)
	defer cleanup()

	cmd := newAffiliateCampaignsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	if err := affiliateCampaignsProductsSalesRankingCmd.RunE(cmd, []string{"camp_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"items\"") {
		t.Fatalf("expected items in output, got %q", buf.String())
	}
}

func TestAffiliateCampaignsExportReportRunE(t *testing.T) {
	mockClient := &affiliateCampaignsMockAPIClient{
		exportResp: json.RawMessage(`{"job_id":"job_1"}`),
	}
	cleanup, buf := setupAffiliateCampaignsMockFactories(mockClient)
	defer cleanup()

	cmd := newAffiliateCampaignsTestCmd()
	addJSONBodyFlags(cmd)
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"email":"test@example.com"}`)

	if err := affiliateCampaignsExportReportCmd.RunE(cmd, []string{"camp_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"job_id\"") {
		t.Fatalf("expected job_id in output, got %q", buf.String())
	}
}
