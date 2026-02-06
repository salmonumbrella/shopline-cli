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

// TestPriceRulesCommandSetup verifies price-rules command initialization
func TestPriceRulesCommandSetup(t *testing.T) {
	if priceRulesCmd.Use != "price-rules" {
		t.Errorf("expected Use 'price-rules', got %q", priceRulesCmd.Use)
	}
	if priceRulesCmd.Short != "Manage price rules" {
		t.Errorf("expected Short 'Manage price rules', got %q", priceRulesCmd.Short)
	}
}

// TestPriceRulesSubcommands verifies all subcommands are registered
func TestPriceRulesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List price rules",
		"get":    "Get price rule details",
		"create": "Create a price rule",
		"delete": "Delete a price rule",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range priceRulesCmd.Commands() {
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

// TestPriceRulesListFlags verifies list command flags exist with correct defaults
func TestPriceRulesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := priceRulesListCmd.Flags().Lookup(f.name)
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

// TestPriceRulesCreateFlags verifies create command flags exist
func TestPriceRulesCreateFlags(t *testing.T) {
	flags := []string{"title", "value-type", "value"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := priceRulesCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

// TestPriceRulesCreateRequiredFlags verifies that required flags are marked correctly
func TestPriceRulesCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"title", "value-type", "value"}

	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := priceRulesCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
				return
			}
			// Check if the flag has required annotation
			annotations := flag.Annotations
			if annotations == nil {
				t.Errorf("flag %q has no annotations, expected required", name)
				return
			}
			if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
				t.Errorf("flag %q is not marked as required", name)
			}
		})
	}
}

// TestPriceRulesGetCmdUse verifies the get command has correct use string
func TestPriceRulesGetCmdUse(t *testing.T) {
	if priceRulesGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", priceRulesGetCmd.Use)
	}
}

// TestPriceRulesDeleteCmdUse verifies the delete command has correct use string
func TestPriceRulesDeleteCmdUse(t *testing.T) {
	if priceRulesDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", priceRulesDeleteCmd.Use)
	}
}

// TestPriceRulesGetArgs verifies get command requires exactly 1 argument
func TestPriceRulesGetArgs(t *testing.T) {
	err := priceRulesGetCmd.Args(priceRulesGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = priceRulesGetCmd.Args(priceRulesGetCmd, []string{"price-rule-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestPriceRulesDeleteArgs verifies delete command requires exactly 1 argument
func TestPriceRulesDeleteArgs(t *testing.T) {
	err := priceRulesDeleteCmd.Args(priceRulesDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = priceRulesDeleteCmd.Args(priceRulesDeleteCmd, []string{"price-rule-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// priceRulesMockAPIClient is a mock implementation of api.APIClient for price rules tests.
type priceRulesMockAPIClient struct {
	api.MockClient
	listPriceRulesResp  *api.PriceRulesListResponse
	listPriceRulesErr   error
	getPriceRuleResp    *api.PriceRule
	getPriceRuleErr     error
	createPriceRuleResp *api.PriceRule
	createPriceRuleErr  error
	deletePriceRuleErr  error
}

func (m *priceRulesMockAPIClient) ListPriceRules(ctx context.Context, opts *api.PriceRulesListOptions) (*api.PriceRulesListResponse, error) {
	return m.listPriceRulesResp, m.listPriceRulesErr
}

func (m *priceRulesMockAPIClient) GetPriceRule(ctx context.Context, id string) (*api.PriceRule, error) {
	return m.getPriceRuleResp, m.getPriceRuleErr
}

func (m *priceRulesMockAPIClient) CreatePriceRule(ctx context.Context, req *api.PriceRuleCreateRequest) (*api.PriceRule, error) {
	return m.createPriceRuleResp, m.createPriceRuleErr
}

func (m *priceRulesMockAPIClient) DeletePriceRule(ctx context.Context, id string) error {
	return m.deletePriceRuleErr
}

// setupPriceRulesMockFactories sets up mock factories for price rules tests.
func setupPriceRulesMockFactories(mockClient *priceRulesMockAPIClient) (func(), *bytes.Buffer) {
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

// newPriceRulesTestCmd creates a test command with common flags for price rules tests.
func newPriceRulesTestCmd() *cobra.Command {
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

// TestPriceRulesListRunE tests the price-rules list command with mock API.
func TestPriceRulesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.PriceRulesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with usage limit",
			mockResp: &api.PriceRulesListResponse{
				Items: []api.PriceRule{
					{
						ID:                "pr_123",
						Title:             "Summer Sale",
						ValueType:         "percentage",
						Value:             "-20",
						TargetType:        "line_item",
						CustomerSelection: "all",
						UsageLimit:        100,
						OncePerCustomer:   true,
						CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "pr_123",
		},
		{
			name: "successful list without usage limit",
			mockResp: &api.PriceRulesListResponse{
				Items: []api.PriceRule{
					{
						ID:                "pr_456",
						Title:             "VIP Discount",
						ValueType:         "fixed_amount",
						Value:             "-10.00",
						TargetType:        "shipping_line",
						CustomerSelection: "prerequisite",
						UsageLimit:        0, // unlimited
						OncePerCustomer:   false,
						CreatedAt:         time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "pr_456",
		},
		{
			name: "successful list with multiple items",
			mockResp: &api.PriceRulesListResponse{
				Items: []api.PriceRule{
					{
						ID:                "pr_aaa",
						Title:             "First Discount",
						ValueType:         "percentage",
						Value:             "-15",
						TargetType:        "line_item",
						CustomerSelection: "all",
						UsageLimit:        50,
						OncePerCustomer:   true,
					},
					{
						ID:                "pr_bbb",
						Title:             "Second Discount",
						ValueType:         "fixed_amount",
						Value:             "-5.00",
						TargetType:        "line_item",
						CustomerSelection: "all",
						UsageLimit:        0,
						OncePerCustomer:   false,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "pr_aaa",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.PriceRulesListResponse{
				Items:      []api.PriceRule{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &priceRulesMockAPIClient{
				listPriceRulesResp: tt.mockResp,
				listPriceRulesErr:  tt.mockErr,
			}
			cleanup, buf := setupPriceRulesMockFactories(mockClient)
			defer cleanup()

			cmd := newPriceRulesTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := priceRulesListCmd.RunE(cmd, []string{})

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

// TestPriceRulesListRunE_JSONOutput tests the list command with JSON output.
func TestPriceRulesListRunE_JSONOutput(t *testing.T) {
	mockClient := &priceRulesMockAPIClient{
		listPriceRulesResp: &api.PriceRulesListResponse{
			Items: []api.PriceRule{
				{
					ID:                "pr_json",
					Title:             "JSON Test",
					ValueType:         "percentage",
					Value:             "-25",
					TargetType:        "line_item",
					CustomerSelection: "all",
					CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupPriceRulesMockFactories(mockClient)
	defer cleanup()

	cmd := newPriceRulesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := priceRulesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "pr_json") {
		t.Errorf("JSON output should contain price rule ID")
	}
}

// TestPriceRulesGetRunE tests the price-rules get command with mock API.
func TestPriceRulesGetRunE(t *testing.T) {
	tests := []struct {
		name        string
		priceRuleID string
		mockResp    *api.PriceRule
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful get with all fields",
			priceRuleID: "pr_123",
			mockResp: &api.PriceRule{
				ID:                "pr_123",
				Title:             "Holiday Sale",
				ValueType:         "percentage",
				Value:             "-20",
				TargetType:        "line_item",
				TargetSelection:   "all",
				AllocationMethod:  "across",
				CustomerSelection: "all",
				OncePerCustomer:   true,
				UsageLimit:        100,
				StartsAt:          time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				EndsAt:            time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				CreatedAt:         time.Date(2024, 11, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:        "successful get with minimal fields",
			priceRuleID: "pr_minimal",
			mockResp: &api.PriceRule{
				ID:                "pr_minimal",
				Title:             "Simple Discount",
				ValueType:         "fixed_amount",
				Value:             "-5.00",
				TargetType:        "line_item",
				TargetSelection:   "entitled",
				AllocationMethod:  "each",
				CustomerSelection: "all",
				OncePerCustomer:   false,
				UsageLimit:        0,
				CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:        "successful get without dates",
			priceRuleID: "pr_nodates",
			mockResp: &api.PriceRule{
				ID:                "pr_nodates",
				Title:             "No Dates Discount",
				ValueType:         "percentage",
				Value:             "-10",
				TargetType:        "shipping_line",
				TargetSelection:   "all",
				AllocationMethod:  "across",
				CustomerSelection: "prerequisite",
				OncePerCustomer:   false,
				UsageLimit:        0,
				CreatedAt:         time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:        "price rule not found",
			priceRuleID: "pr_999",
			mockErr:     errors.New("price rule not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &priceRulesMockAPIClient{
				getPriceRuleResp: tt.mockResp,
				getPriceRuleErr:  tt.mockErr,
			}
			cleanup, _ := setupPriceRulesMockFactories(mockClient)
			defer cleanup()

			cmd := newPriceRulesTestCmd()

			err := priceRulesGetCmd.RunE(cmd, []string{tt.priceRuleID})

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

// TestPriceRulesGetRunE_JSONOutput tests the get command with JSON output.
func TestPriceRulesGetRunE_JSONOutput(t *testing.T) {
	mockClient := &priceRulesMockAPIClient{
		getPriceRuleResp: &api.PriceRule{
			ID:                "pr_json_get",
			Title:             "JSON Get Test",
			ValueType:         "percentage",
			Value:             "-15",
			TargetType:        "line_item",
			TargetSelection:   "all",
			AllocationMethod:  "across",
			CustomerSelection: "all",
			CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupPriceRulesMockFactories(mockClient)
	defer cleanup()

	cmd := newPriceRulesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := priceRulesGetCmd.RunE(cmd, []string{"pr_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "pr_json_get") {
		t.Errorf("JSON output should contain price rule ID")
	}
}

// TestPriceRulesGetRunE_WithUsageLimit tests get command output with usage limit.
func TestPriceRulesGetRunE_WithUsageLimit(t *testing.T) {
	mockClient := &priceRulesMockAPIClient{
		getPriceRuleResp: &api.PriceRule{
			ID:                "pr_with_limit",
			Title:             "Limited Use",
			ValueType:         "percentage",
			Value:             "-10",
			TargetType:        "line_item",
			TargetSelection:   "all",
			AllocationMethod:  "across",
			CustomerSelection: "all",
			UsageLimit:        50,
			OncePerCustomer:   true,
			CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupPriceRulesMockFactories(mockClient)
	defer cleanup()

	cmd := newPriceRulesTestCmd()

	err := priceRulesGetCmd.RunE(cmd, []string{"pr_with_limit"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPriceRulesGetRunE_WithStartsAt tests get command output with starts_at date.
func TestPriceRulesGetRunE_WithStartsAt(t *testing.T) {
	mockClient := &priceRulesMockAPIClient{
		getPriceRuleResp: &api.PriceRule{
			ID:                "pr_starts",
			Title:             "Starts Later",
			ValueType:         "percentage",
			Value:             "-20",
			TargetType:        "line_item",
			TargetSelection:   "all",
			AllocationMethod:  "across",
			CustomerSelection: "all",
			StartsAt:          time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupPriceRulesMockFactories(mockClient)
	defer cleanup()

	cmd := newPriceRulesTestCmd()

	err := priceRulesGetCmd.RunE(cmd, []string{"pr_starts"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPriceRulesGetRunE_WithEndsAt tests get command output with ends_at date.
func TestPriceRulesGetRunE_WithEndsAt(t *testing.T) {
	mockClient := &priceRulesMockAPIClient{
		getPriceRuleResp: &api.PriceRule{
			ID:                "pr_ends",
			Title:             "Ends Soon",
			ValueType:         "percentage",
			Value:             "-25",
			TargetType:        "line_item",
			TargetSelection:   "all",
			AllocationMethod:  "across",
			CustomerSelection: "all",
			EndsAt:            time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupPriceRulesMockFactories(mockClient)
	defer cleanup()

	cmd := newPriceRulesTestCmd()

	err := priceRulesGetCmd.RunE(cmd, []string{"pr_ends"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPriceRulesCreateRunE tests the price-rules create command with mock API.
func TestPriceRulesCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.PriceRule
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.PriceRule{
				ID:                "pr_new",
				Title:             "New Discount",
				ValueType:         "percentage",
				Value:             "-15",
				TargetType:        "line_item",
				CustomerSelection: "all",
				CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("failed to create price rule"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &priceRulesMockAPIClient{
				createPriceRuleResp: tt.mockResp,
				createPriceRuleErr:  tt.mockErr,
			}
			cleanup, _ := setupPriceRulesMockFactories(mockClient)
			defer cleanup()

			cmd := newPriceRulesTestCmd()
			cmd.Flags().String("title", "", "")
			cmd.Flags().String("value-type", "", "")
			cmd.Flags().String("value", "", "")
			_ = cmd.Flags().Set("title", "New Discount")
			_ = cmd.Flags().Set("value-type", "percentage")
			_ = cmd.Flags().Set("value", "-15")

			err := priceRulesCreateCmd.RunE(cmd, []string{})

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

// TestPriceRulesCreateRunE_FixedAmount tests create command with fixed_amount value type.
func TestPriceRulesCreateRunE_FixedAmount(t *testing.T) {
	mockClient := &priceRulesMockAPIClient{
		createPriceRuleResp: &api.PriceRule{
			ID:        "pr_fixed",
			Title:     "Fixed Discount",
			ValueType: "fixed_amount",
			Value:     "-10.00",
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupPriceRulesMockFactories(mockClient)
	defer cleanup()

	cmd := newPriceRulesTestCmd()
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("value-type", "", "")
	cmd.Flags().String("value", "", "")
	_ = cmd.Flags().Set("title", "Fixed Discount")
	_ = cmd.Flags().Set("value-type", "fixed_amount")
	_ = cmd.Flags().Set("value", "-10.00")

	err := priceRulesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPriceRulesDeleteRunE tests the price-rules delete command with mock API.
func TestPriceRulesDeleteRunE(t *testing.T) {
	tests := []struct {
		name        string
		priceRuleID string
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful delete",
			priceRuleID: "pr_123",
			mockErr:     nil,
		},
		{
			name:        "delete error",
			priceRuleID: "pr_456",
			mockErr:     errors.New("failed to delete"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &priceRulesMockAPIClient{
				deletePriceRuleErr: tt.mockErr,
			}
			cleanup, _ := setupPriceRulesMockFactories(mockClient)
			defer cleanup()

			cmd := newPriceRulesTestCmd()

			err := priceRulesDeleteCmd.RunE(cmd, []string{tt.priceRuleID})

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

// TestPriceRulesListRunE_GetClientFails verifies error handling when getClient fails
func TestPriceRulesListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := priceRulesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestPriceRulesGetRunE_GetClientFails verifies error handling when getClient fails
func TestPriceRulesGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := priceRulesGetCmd.RunE(cmd, []string{"price_rule_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestPriceRulesCreateRunE_GetClientFails verifies error handling when getClient fails
func TestPriceRulesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("value-type", "", "")
	cmd.Flags().String("value", "", "")
	_ = cmd.Flags().Set("title", "20% Off")
	_ = cmd.Flags().Set("value-type", "percentage")
	_ = cmd.Flags().Set("value", "-20")

	err := priceRulesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestPriceRulesDeleteRunE_GetClientFails verifies error handling when getClient fails
func TestPriceRulesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := priceRulesDeleteCmd.RunE(cmd, []string{"price_rule_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestPriceRulesListRunE_NoProfiles verifies error when no profiles are configured
func TestPriceRulesListRunE_NoProfiles(t *testing.T) {
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
	err := priceRulesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestPriceRulesGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestPriceRulesGetRunE_MultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	err := priceRulesGetCmd.RunE(cmd, []string{"price_rule_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestPriceRulesCreateRunE_NoProfiles verifies error when no profiles are configured
func TestPriceRulesCreateRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("value-type", "", "")
	cmd.Flags().String("value", "", "")
	_ = cmd.Flags().Set("title", "Test")
	_ = cmd.Flags().Set("value-type", "percentage")
	_ = cmd.Flags().Set("value", "-10")

	err := priceRulesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestPriceRulesDeleteRunE_NoProfiles verifies error when no profiles are configured
func TestPriceRulesDeleteRunE_NoProfiles(t *testing.T) {
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
	_ = cmd.Flags().Set("yes", "true")

	err := priceRulesDeleteCmd.RunE(cmd, []string{"pr_123"})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestPriceRulesDeleteRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestPriceRulesDeleteRunE_MultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := priceRulesDeleteCmd.RunE(cmd, []string{"pr_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestPriceRulesWithMockStore tests price rules commands with a mock credential store
func TestPriceRulesWithMockStore(t *testing.T) {
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

// TestPriceRulesListRunE_AllBranches tests list command covering all format branches.
func TestPriceRulesListRunE_AllBranches(t *testing.T) {
	// Test with OncePerCustomer=true and UsageLimit>0
	mockClient := &priceRulesMockAPIClient{
		listPriceRulesResp: &api.PriceRulesListResponse{
			Items: []api.PriceRule{
				{
					ID:                "pr_branch1",
					Title:             "Branch Test 1",
					ValueType:         "percentage",
					Value:             "-10",
					TargetType:        "line_item",
					CustomerSelection: "all",
					UsageLimit:        25,
					OncePerCustomer:   true,
				},
				{
					ID:                "pr_branch2",
					Title:             "Branch Test 2",
					ValueType:         "fixed_amount",
					Value:             "-5.00",
					TargetType:        "shipping_line",
					CustomerSelection: "prerequisite",
					UsageLimit:        0,     // unlimited, should show "-"
					OncePerCustomer:   false, // should show "No"
				},
			},
			TotalCount: 2,
		},
	}
	cleanup, buf := setupPriceRulesMockFactories(mockClient)
	defer cleanup()

	cmd := newPriceRulesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := priceRulesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	// Verify both rules are in output
	if !strings.Contains(output, "pr_branch1") {
		t.Error("output should contain pr_branch1")
	}
	if !strings.Contains(output, "pr_branch2") {
		t.Error("output should contain pr_branch2")
	}
}
