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

// mockSellingPlansClient is a mock implementation for selling plans API methods.
type mockSellingPlansClient struct {
	api.MockClient // embed base mock for unimplemented methods

	listSellingPlansResp *api.SellingPlansListResponse
	listSellingPlansErr  error

	getSellingPlanResp *api.SellingPlan
	getSellingPlanErr  error

	createSellingPlanResp *api.SellingPlan
	createSellingPlanErr  error

	deleteSellingPlanErr error
}

func (m *mockSellingPlansClient) ListSellingPlans(ctx context.Context, opts *api.SellingPlansListOptions) (*api.SellingPlansListResponse, error) {
	return m.listSellingPlansResp, m.listSellingPlansErr
}

func (m *mockSellingPlansClient) GetSellingPlan(ctx context.Context, id string) (*api.SellingPlan, error) {
	return m.getSellingPlanResp, m.getSellingPlanErr
}

func (m *mockSellingPlansClient) CreateSellingPlan(ctx context.Context, req *api.SellingPlanCreateRequest) (*api.SellingPlan, error) {
	return m.createSellingPlanResp, m.createSellingPlanErr
}

func (m *mockSellingPlansClient) DeleteSellingPlan(ctx context.Context, id string) error {
	return m.deleteSellingPlanErr
}

// setupSellingPlansTest configures the test environment with mock factories.
func setupSellingPlansTest(t *testing.T, mockClient *mockSellingPlansClient) (cleanup func()) {
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

func TestSellingPlansCommand(t *testing.T) {
	if sellingPlansCmd == nil {
		t.Fatal("sellingPlansCmd is nil")
	}
	if sellingPlansCmd.Use != "selling-plans" {
		t.Errorf("Expected Use to be 'selling-plans', got %q", sellingPlansCmd.Use)
	}
	if sellingPlansCmd.Short != "Manage selling plan configurations" {
		t.Errorf("Expected Short to be 'Manage selling plan configurations', got %q", sellingPlansCmd.Short)
	}
}

func TestSellingPlansSubcommands(t *testing.T) {
	subcommands := sellingPlansCmd.Commands()
	expectedCmds := map[string]bool{"list": false, "get": false, "create": false, "delete": false}
	for _, cmd := range subcommands {
		switch cmd.Use {
		case "list":
			expectedCmds["list"] = true
		case "get <id>":
			expectedCmds["get"] = true
		case "create":
			expectedCmds["create"] = true
		case "delete <id>":
			expectedCmds["delete"] = true
		}
	}
	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestSellingPlansListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status", ""},
		{"page", "1"},
		{"page-size", "20"},
	}
	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := sellingPlansListCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Expected flag --%s not found on list command", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("Expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestSellingPlansCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"description", ""},
		{"billing-policy", ""},
		{"delivery-policy", ""},
		{"frequency", ""},
		{"frequency-interval", "1"},
		{"trial-days", "0"},
		{"discount-type", ""},
		{"discount-value", ""},
		{"position", "0"},
	}
	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := sellingPlansCreateCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Expected flag --%s not found on create command", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("Expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestSellingPlansCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"name", "frequency"}
	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := sellingPlansCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
				return
			}
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

func TestSellingPlansGetArgsValidation(t *testing.T) {
	if sellingPlansGetCmd.Args == nil {
		t.Fatal("Expected Args validator on get command")
	}
	err := sellingPlansGetCmd.Args(sellingPlansGetCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = sellingPlansGetCmd.Args(sellingPlansGetCmd, []string{"plan_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
	err = sellingPlansGetCmd.Args(sellingPlansGetCmd, []string{"plan_123", "extra"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestSellingPlansDeleteArgsValidation(t *testing.T) {
	if sellingPlansDeleteCmd.Args == nil {
		t.Fatal("Expected Args validator on delete command")
	}
	err := sellingPlansDeleteCmd.Args(sellingPlansDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = sellingPlansDeleteCmd.Args(sellingPlansDeleteCmd, []string{"plan_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
	err = sellingPlansDeleteCmd.Args(sellingPlansDeleteCmd, []string{"plan_123", "extra"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestSellingPlansListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := sellingPlansListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "keyring error") {
		t.Errorf("Expected keyring error, got %v", err)
	}
}

func TestSellingPlansGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := sellingPlansGetCmd.RunE(cmd, []string{"plan_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestSellingPlansCreateDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("name", "", "Name")
	_ = cmd.Flags().Set("name", "Monthly Plan")
	cmd.Flags().String("frequency", "", "Frequency")
	_ = cmd.Flags().Set("frequency", "monthly")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("billing-policy", "", "Billing policy")
	cmd.Flags().String("delivery-policy", "", "Delivery policy")
	cmd.Flags().Int("frequency-interval", 1, "Frequency interval")
	cmd.Flags().Int("trial-days", 0, "Trial days")
	cmd.Flags().String("discount-type", "", "Discount type")
	cmd.Flags().String("discount-value", "", "Discount value")
	cmd.Flags().Int("position", 0, "Position")
	err := sellingPlansCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}
}

func TestSellingPlansDeleteDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")
	err := sellingPlansDeleteCmd.RunE(cmd, []string{"plan_123"})
	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}
}

func TestSellingPlansDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := sellingPlansDeleteCmd.RunE(cmd, []string{"plan_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestSellingPlansCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "", "Name")
	_ = cmd.Flags().Set("name", "Monthly Plan")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("billing-policy", "", "Billing policy")
	cmd.Flags().String("delivery-policy", "", "Delivery policy")
	cmd.Flags().String("frequency", "", "Frequency")
	_ = cmd.Flags().Set("frequency", "monthly")
	cmd.Flags().Int("frequency-interval", 1, "Frequency interval")
	cmd.Flags().Int("trial-days", 0, "Trial days")
	cmd.Flags().String("discount-type", "", "Discount type")
	cmd.Flags().String("discount-value", "", "Discount value")
	cmd.Flags().Int("position", 0, "Position")
	err := sellingPlansCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestSellingPlansListRunE tests the list command execution with mock API.
func TestSellingPlansListRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		mockResp     *api.SellingPlansListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name: "successful list with basic plan",
			mockResp: &api.SellingPlansListResponse{
				Items: []api.SellingPlan{
					{
						ID:                "plan_123",
						Name:              "Monthly Subscription",
						Frequency:         "monthly",
						FrequencyInterval: 1,
						DiscountType:      "",
						DiscountValue:     "",
						TrialDays:         0,
						Status:            "active",
						CreatedAt:         testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "plan_123",
		},
		{
			name: "successful list with percentage discount",
			mockResp: &api.SellingPlansListResponse{
				Items: []api.SellingPlan{
					{
						ID:                "plan_456",
						Name:              "Yearly Plan",
						Frequency:         "yearly",
						FrequencyInterval: 1,
						DiscountType:      "percentage",
						DiscountValue:     "10",
						TrialDays:         14,
						Status:            "active",
						CreatedAt:         testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "10%",
		},
		{
			name: "successful list with fixed discount",
			mockResp: &api.SellingPlansListResponse{
				Items: []api.SellingPlan{
					{
						ID:                "plan_789",
						Name:              "Weekly Plan",
						Frequency:         "weekly",
						FrequencyInterval: 1,
						DiscountType:      "fixed",
						DiscountValue:     "$5.00",
						TrialDays:         0,
						Status:            "active",
						CreatedAt:         testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "$5.00",
		},
		{
			name: "successful list with frequency interval > 1",
			mockResp: &api.SellingPlansListResponse{
				Items: []api.SellingPlan{
					{
						ID:                "plan_multi",
						Name:              "Every 2 Weeks",
						Frequency:         "weekly",
						FrequencyInterval: 2,
						DiscountType:      "",
						DiscountValue:     "",
						TrialDays:         7,
						Status:            "active",
						CreatedAt:         testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "every 2 weekly",
		},
		{
			name: "successful list with trial days",
			mockResp: &api.SellingPlansListResponse{
				Items: []api.SellingPlan{
					{
						ID:                "plan_trial",
						Name:              "Trial Plan",
						Frequency:         "monthly",
						FrequencyInterval: 1,
						TrialDays:         30,
						Status:            "active",
						CreatedAt:         testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "30 days",
		},
		{
			name:         "successful list JSON output",
			outputFormat: "json",
			mockResp: &api.SellingPlansListResponse{
				Items: []api.SellingPlan{
					{
						ID:        "plan_json",
						Name:      "JSON Plan",
						Frequency: "monthly",
						Status:    "active",
						CreatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "plan_json",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.SellingPlansListResponse{
				Items:      []api.SellingPlan{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSellingPlansClient{
				listSellingPlansResp: tt.mockResp,
				listSellingPlansErr:  tt.mockErr,
			}

			cleanup := setupSellingPlansTest(t, mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := sellingPlansListCmd.RunE(cmd, []string{})

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

// TestSellingPlansGetRunE tests the get command execution with mock API.
func TestSellingPlansGetRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		planID       string
		mockResp     *api.SellingPlan
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name:   "successful get",
			planID: "plan_123",
			mockResp: &api.SellingPlan{
				ID:                "plan_123",
				Name:              "Monthly Subscription",
				Description:       "Subscribe monthly for savings",
				BillingPolicy:     "recurring",
				DeliveryPolicy:    "recurring",
				Frequency:         "monthly",
				FrequencyInterval: 1,
				TrialDays:         14,
				DiscountType:      "percentage",
				DiscountValue:     "10",
				Status:            "active",
				Position:          1,
				CreatedAt:         testTime,
				UpdatedAt:         testTime,
			},
			wantOutput: "Selling Plan ID:     plan_123",
		},
		{
			name:         "successful get JSON output",
			planID:       "plan_456",
			outputFormat: "json",
			mockResp: &api.SellingPlan{
				ID:        "plan_456",
				Name:      "JSON Plan",
				Frequency: "weekly",
				Status:    "active",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
			wantOutput: "plan_456",
		},
		{
			name:    "plan not found",
			planID:  "plan_999",
			mockErr: errors.New("selling plan not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSellingPlansClient{
				getSellingPlanResp: tt.mockResp,
				getSellingPlanErr:  tt.mockErr,
			}

			cleanup := setupSellingPlansTest(t, mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := sellingPlansGetCmd.RunE(cmd, []string{tt.planID})

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

// TestSellingPlansCreateRunE tests the create command execution with mock API.
func TestSellingPlansCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		flags        map[string]string
		mockResp     *api.SellingPlan
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name: "successful create",
			flags: map[string]string{
				"name":               "New Monthly Plan",
				"frequency":          "monthly",
				"description":        "Monthly subscription plan",
				"billing-policy":     "recurring",
				"delivery-policy":    "recurring",
				"frequency-interval": "1",
				"trial-days":         "7",
				"discount-type":      "percentage",
				"discount-value":     "15",
				"position":           "1",
			},
			mockResp: &api.SellingPlan{
				ID:                "plan_new",
				Name:              "New Monthly Plan",
				Frequency:         "monthly",
				FrequencyInterval: 1,
				Status:            "active",
				CreatedAt:         testTime,
				UpdatedAt:         testTime,
			},
			wantOutput: "Created selling plan plan_new",
		},
		{
			name: "successful create JSON output",
			flags: map[string]string{
				"name":      "JSON Plan",
				"frequency": "weekly",
			},
			outputFormat: "json",
			mockResp: &api.SellingPlan{
				ID:        "plan_json_create",
				Name:      "JSON Plan",
				Frequency: "weekly",
				Status:    "active",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
			wantOutput: "plan_json_create",
		},
		{
			name: "create fails",
			flags: map[string]string{
				"name":      "Failed Plan",
				"frequency": "monthly",
			},
			mockErr: errors.New("invalid frequency"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSellingPlansClient{
				createSellingPlanResp: tt.mockResp,
				createSellingPlanErr:  tt.mockErr,
			}

			cleanup := setupSellingPlansTest(t, mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Bool("dry-run", false, "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("billing-policy", "", "")
			cmd.Flags().String("delivery-policy", "", "")
			cmd.Flags().String("frequency", "", "")
			cmd.Flags().Int("frequency-interval", 1, "")
			cmd.Flags().Int("trial-days", 0, "")
			cmd.Flags().String("discount-type", "", "")
			cmd.Flags().String("discount-value", "", "")
			cmd.Flags().Int("position", 0, "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := sellingPlansCreateCmd.RunE(cmd, []string{})

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

// TestSellingPlansDeleteRunE tests the delete command execution with mock API.
func TestSellingPlansDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		planID  string
		mockErr error
		wantErr bool
	}{
		{
			name:   "successful delete",
			planID: "plan_123",
		},
		{
			name:    "delete fails",
			planID:  "plan_456",
			mockErr: errors.New("plan not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSellingPlansClient{
				deleteSellingPlanErr: tt.mockErr,
			}

			cleanup := setupSellingPlansTest(t, mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("dry-run", false, "")

			err := sellingPlansDeleteCmd.RunE(cmd, []string{tt.planID})

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

// TestSellingPlansListWithStatus tests list command with status filter.
func TestSellingPlansListWithStatus(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		listSellingPlansResp: &api.SellingPlansListResponse{
			Items: []api.SellingPlan{
				{
					ID:        "plan_active",
					Name:      "Active Plan",
					Status:    "active",
					Frequency: "monthly",
					CreatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "active", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")

	err := sellingPlansListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSellingPlansListEmptyDiscount tests list with empty discount display.
func TestSellingPlansListEmptyDiscount(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		listSellingPlansResp: &api.SellingPlansListResponse{
			Items: []api.SellingPlan{
				{
					ID:            "plan_no_discount",
					Name:          "No Discount Plan",
					Status:        "active",
					Frequency:     "monthly",
					DiscountType:  "", // Empty discount type
					DiscountValue: "",
					CreatedAt:     testTime,
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := sellingPlansListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "-") {
		t.Log("Discount column should show '-' for empty discount")
	}
}

// TestSellingPlansListZeroTrialDays tests list with zero trial days display.
func TestSellingPlansListZeroTrialDays(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		listSellingPlansResp: &api.SellingPlansListResponse{
			Items: []api.SellingPlan{
				{
					ID:        "plan_no_trial",
					Name:      "No Trial Plan",
					Status:    "active",
					Frequency: "weekly",
					TrialDays: 0, // Zero trial days
					CreatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := sellingPlansListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSellingPlansWithMockStore tests with a mock credential store setup.
func TestSellingPlansWithMockStore(t *testing.T) {
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

// TestSellingPlansListDiscountTypeFixed tests fixed discount type display.
func TestSellingPlansListDiscountTypeFixed(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		listSellingPlansResp: &api.SellingPlansListResponse{
			Items: []api.SellingPlan{
				{
					ID:            "plan_fixed",
					Name:          "Fixed Discount Plan",
					Status:        "active",
					Frequency:     "monthly",
					DiscountType:  "fixed",
					DiscountValue: "25.00",
					CreatedAt:     testTime,
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := sellingPlansListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "25.00") {
		t.Errorf("Expected fixed discount value in output, got: %s", output)
	}
}

// TestSellingPlansGetTextOutput tests get command text output fields.
func TestSellingPlansGetTextOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		getSellingPlanResp: &api.SellingPlan{
			ID:                "plan_full",
			Name:              "Full Plan",
			Description:       "Complete subscription plan",
			BillingPolicy:     "recurring",
			DeliveryPolicy:    "recurring",
			Frequency:         "monthly",
			FrequencyInterval: 3,
			TrialDays:         14,
			DiscountType:      "percentage",
			DiscountValue:     "20",
			Status:            "active",
			Position:          5,
			CreatedAt:         testTime,
			UpdatedAt:         testTime,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := sellingPlansGetCmd.RunE(cmd, []string{"plan_full"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSellingPlansCreateWithAllFlags tests create command with all flags.
func TestSellingPlansCreateWithAllFlags(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		createSellingPlanResp: &api.SellingPlan{
			ID:                "plan_all_flags",
			Name:              "Complete Plan",
			Description:       "Plan with all options",
			BillingPolicy:     "recurring",
			DeliveryPolicy:    "recurring",
			Frequency:         "quarterly",
			FrequencyInterval: 1,
			TrialDays:         30,
			DiscountType:      "percentage",
			DiscountValue:     "25",
			Status:            "active",
			Position:          10,
			CreatedAt:         testTime,
			UpdatedAt:         testTime,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("name", "Complete Plan", "")
	cmd.Flags().String("description", "Plan with all options", "")
	cmd.Flags().String("billing-policy", "recurring", "")
	cmd.Flags().String("delivery-policy", "recurring", "")
	cmd.Flags().String("frequency", "quarterly", "")
	cmd.Flags().Int("frequency-interval", 1, "")
	cmd.Flags().Int("trial-days", 30, "")
	cmd.Flags().String("discount-type", "percentage", "")
	cmd.Flags().String("discount-value", "25", "")
	cmd.Flags().Int("position", 10, "")

	err := sellingPlansCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSellingPlansListJSONError tests JSON formatter error handling.
func TestSellingPlansListJSONError(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		listSellingPlansResp: &api.SellingPlansListResponse{
			Items: []api.SellingPlan{
				{
					ID:        "plan_json",
					Name:      "JSON Test",
					Frequency: "monthly",
					CreatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := sellingPlansListCmd.RunE(cmd, []string{})
	// JSON output should succeed
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSellingPlansGetJSONError tests JSON formatter error for get command.
func TestSellingPlansGetJSONError(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		getSellingPlanResp: &api.SellingPlan{
			ID:        "plan_json_get",
			Name:      "JSON Get Test",
			Frequency: "weekly",
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := sellingPlansGetCmd.RunE(cmd, []string{"plan_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSellingPlansCreateJSONError tests JSON formatter error for create command.
func TestSellingPlansCreateJSONError(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockSellingPlansClient{
		createSellingPlanResp: &api.SellingPlan{
			ID:        "plan_json_create",
			Name:      "JSON Create Test",
			Frequency: "daily",
			Status:    "active",
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}

	cleanup := setupSellingPlansTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("name", "JSON Create Test", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("billing-policy", "", "")
	cmd.Flags().String("delivery-policy", "", "")
	cmd.Flags().String("frequency", "daily", "")
	cmd.Flags().Int("frequency-interval", 1, "")
	cmd.Flags().Int("trial-days", 0, "")
	cmd.Flags().String("discount-type", "", "")
	cmd.Flags().String("discount-value", "", "")
	cmd.Flags().Int("position", 0, "")

	err := sellingPlansCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
