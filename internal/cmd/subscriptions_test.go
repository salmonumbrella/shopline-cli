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

// subscriptionsMockAPIClient is a mock implementation of api.APIClient for subscriptions tests.
type subscriptionsMockAPIClient struct {
	api.MockClient
	listSubscriptionsResp  *api.SubscriptionsListResponse
	listSubscriptionsErr   error
	getSubscriptionResp    *api.Subscription
	getSubscriptionErr     error
	createSubscriptionResp *api.Subscription
	createSubscriptionErr  error
	deleteSubscriptionErr  error
}

func (m *subscriptionsMockAPIClient) ListSubscriptions(ctx context.Context, opts *api.SubscriptionsListOptions) (*api.SubscriptionsListResponse, error) {
	return m.listSubscriptionsResp, m.listSubscriptionsErr
}

func (m *subscriptionsMockAPIClient) GetSubscription(ctx context.Context, id string) (*api.Subscription, error) {
	return m.getSubscriptionResp, m.getSubscriptionErr
}

func (m *subscriptionsMockAPIClient) CreateSubscription(ctx context.Context, req *api.SubscriptionCreateRequest) (*api.Subscription, error) {
	return m.createSubscriptionResp, m.createSubscriptionErr
}

func (m *subscriptionsMockAPIClient) DeleteSubscription(ctx context.Context, id string) error {
	return m.deleteSubscriptionErr
}

// setupSubscriptionsMockFactories sets up mock factories for subscriptions tests.
func setupSubscriptionsMockFactories(mockClient *subscriptionsMockAPIClient) (func(), *bytes.Buffer) {
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

// newSubscriptionsTestCmd creates a test command with common flags for subscriptions tests.
func newSubscriptionsTestCmd() *cobra.Command {
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

func TestSubscriptionsCmd(t *testing.T) {
	if subscriptionsCmd.Use != "subscriptions" {
		t.Errorf("Expected Use 'subscriptions', got %q", subscriptionsCmd.Use)
	}
	if subscriptionsCmd.Short != "Manage customer subscriptions" {
		t.Errorf("Expected Short 'Manage customer subscriptions', got %q", subscriptionsCmd.Short)
	}
}

func TestSubscriptionsListCmd(t *testing.T) {
	if subscriptionsListCmd.Use != "list" {
		t.Errorf("Expected Use 'list', got %q", subscriptionsListCmd.Use)
	}
	if subscriptionsListCmd.Short != "List subscriptions" {
		t.Errorf("Expected Short 'List subscriptions', got %q", subscriptionsListCmd.Short)
	}
}

func TestSubscriptionsGetCmd(t *testing.T) {
	if subscriptionsGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use 'get <id>', got %q", subscriptionsGetCmd.Use)
	}
	if subscriptionsGetCmd.Short != "Get subscription details" {
		t.Errorf("Expected Short 'Get subscription details', got %q", subscriptionsGetCmd.Short)
	}
}

func TestSubscriptionsCreateCmd(t *testing.T) {
	if subscriptionsCreateCmd.Use != "create" {
		t.Errorf("Expected Use 'create', got %q", subscriptionsCreateCmd.Use)
	}
	if subscriptionsCreateCmd.Short != "Create a subscription" {
		t.Errorf("Expected Short 'Create a subscription', got %q", subscriptionsCreateCmd.Short)
	}
}

func TestSubscriptionsDeleteCmd(t *testing.T) {
	if subscriptionsDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use 'delete <id>', got %q", subscriptionsDeleteCmd.Use)
	}
	if subscriptionsDeleteCmd.Short != "Cancel a subscription" {
		t.Errorf("Expected Short 'Cancel a subscription', got %q", subscriptionsDeleteCmd.Short)
	}
}

func TestSubscriptionsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"customer-id", ""},
		{"product-id", ""},
		{"status", ""},
		{"page", "1"},
		{"page-size", "20"},
	}
	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := subscriptionsListCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Expected flag --%s not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("Expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestSubscriptionsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"customer-id", ""},
		{"product-id", ""},
		{"variant-id", ""},
		{"interval", "month"},
		{"interval-count", "1"},
	}
	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := subscriptionsCreateCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Expected flag --%s not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("Expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestSubscriptionsGetArgsValidation(t *testing.T) {
	if subscriptionsGetCmd.Args == nil {
		t.Fatal("Expected Args validator on get command")
	}
	err := subscriptionsGetCmd.Args(subscriptionsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = subscriptionsGetCmd.Args(subscriptionsGetCmd, []string{"sub_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestSubscriptionsDeleteArgsValidation(t *testing.T) {
	if subscriptionsDeleteCmd.Args == nil {
		t.Fatal("Expected Args validator on delete command")
	}
	err := subscriptionsDeleteCmd.Args(subscriptionsDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = subscriptionsDeleteCmd.Args(subscriptionsDeleteCmd, []string{"sub_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestSubscriptionsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("product-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := subscriptionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "keyring error") && !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

func TestSubscriptionsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := subscriptionsGetCmd.RunE(cmd, []string{"sub_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSubscriptionsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("product-id", "prod_456", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().String("interval", "month", "")
	cmd.Flags().Int("interval-count", 1, "")
	err := subscriptionsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSubscriptionsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := subscriptionsDeleteCmd.RunE(cmd, []string{"sub_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSubscriptionsListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("product-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := subscriptionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSubscriptionsCreateRunE_DryRun(t *testing.T) {
	origWriter := formatterWriter
	defer func() { formatterWriter = origWriter }()

	var buf bytes.Buffer
	formatterWriter = &buf
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("product-id", "prod_456", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().String("interval", "month", "")
	cmd.Flags().Int("interval-count", 1, "")
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := subscriptionsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run, got: %s", output)
	}
	if !strings.Contains(output, "cust_123") {
		t.Errorf("Expected customer ID in output, got: %s", output)
	}
	if !strings.Contains(output, "prod_456") {
		t.Errorf("Expected product ID in output, got: %s", output)
	}
}

func TestSubscriptionsDeleteRunE_DryRun(t *testing.T) {
	origWriter := formatterWriter
	defer func() { formatterWriter = origWriter }()

	var buf bytes.Buffer
	formatterWriter = &buf
	cmd := newTestCmdWithFlags()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := subscriptionsDeleteCmd.RunE(cmd, []string{"sub_123"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run, got: %s", output)
	}
	if !strings.Contains(output, "sub_123") {
		t.Errorf("Expected subscription ID in output, got: %s", output)
	}
}

// TestSubscriptionsListRunE tests the subscriptions list command with mock API.
func TestSubscriptionsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.SubscriptionsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with single interval",
			mockResp: &api.SubscriptionsListResponse{
				Items: []api.Subscription{
					{
						ID:            "sub_123",
						CustomerID:    "cust_456",
						ProductID:     "prod_789",
						Status:        api.SubscriptionStatusActive,
						Interval:      "month",
						IntervalCount: 1,
						Price:         "19.99",
						Currency:      "USD",
						NextBillingAt: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sub_123",
		},
		{
			name: "successful list with multiple interval count",
			mockResp: &api.SubscriptionsListResponse{
				Items: []api.Subscription{
					{
						ID:            "sub_456",
						CustomerID:    "cust_789",
						ProductID:     "prod_012",
						Status:        api.SubscriptionStatusActive,
						Interval:      "week",
						IntervalCount: 2,
						Price:         "9.99",
						Currency:      "USD",
						NextBillingAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "2 weeks",
		},
		{
			name: "successful list without currency",
			mockResp: &api.SubscriptionsListResponse{
				Items: []api.Subscription{
					{
						ID:            "sub_789",
						CustomerID:    "cust_012",
						ProductID:     "prod_345",
						Status:        api.SubscriptionStatusPaused,
						Interval:      "month",
						IntervalCount: 1,
						Price:         "29.99",
						Currency:      "",
						NextBillingAt: time.Time{},
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sub_789",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.SubscriptionsListResponse{
				Items:      []api.Subscription{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &subscriptionsMockAPIClient{
				listSubscriptionsResp: tt.mockResp,
				listSubscriptionsErr:  tt.mockErr,
			}
			cleanup, buf := setupSubscriptionsMockFactories(mockClient)
			defer cleanup()

			cmd := newSubscriptionsTestCmd()
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().String("product-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := subscriptionsListCmd.RunE(cmd, []string{})

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

// TestSubscriptionsListRunE_JSONOutput tests the subscriptions list command with JSON output.
func TestSubscriptionsListRunE_JSONOutput(t *testing.T) {
	mockClient := &subscriptionsMockAPIClient{
		listSubscriptionsResp: &api.SubscriptionsListResponse{
			Items: []api.Subscription{
				{
					ID:         "sub_123",
					CustomerID: "cust_456",
					ProductID:  "prod_789",
					Status:     api.SubscriptionStatusActive,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupSubscriptionsMockFactories(mockClient)
	defer cleanup()

	cmd := newSubscriptionsTestCmd()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("product-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := subscriptionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sub_123") {
		t.Errorf("JSON output should contain subscription ID, got: %s", output)
	}
}

// TestSubscriptionsGetRunE tests the subscriptions get command with mock API.
func TestSubscriptionsGetRunE(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		mockResp       *api.Subscription
		mockErr        error
		wantErr        bool
	}{
		{
			name:           "successful get with all fields",
			subscriptionID: "sub_123",
			mockResp: &api.Subscription{
				ID:            "sub_123",
				CustomerID:    "cust_456",
				ProductID:     "prod_789",
				VariantID:     "var_012",
				Status:        api.SubscriptionStatusActive,
				Interval:      "month",
				IntervalCount: 1,
				Price:         "19.99",
				Currency:      "USD",
				NextBillingAt: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
				CancelledAt:   time.Time{},
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:           "successful get with cancelled subscription",
			subscriptionID: "sub_456",
			mockResp: &api.Subscription{
				ID:            "sub_456",
				CustomerID:    "cust_789",
				ProductID:     "prod_012",
				VariantID:     "var_345",
				Status:        api.SubscriptionStatusCancelled,
				Interval:      "year",
				IntervalCount: 1,
				Price:         "99.99",
				Currency:      "USD",
				NextBillingAt: time.Time{},
				CancelledAt:   time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:           "subscription not found",
			subscriptionID: "sub_999",
			mockErr:        errors.New("subscription not found"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &subscriptionsMockAPIClient{
				getSubscriptionResp: tt.mockResp,
				getSubscriptionErr:  tt.mockErr,
			}
			cleanup, _ := setupSubscriptionsMockFactories(mockClient)
			defer cleanup()

			cmd := newSubscriptionsTestCmd()

			err := subscriptionsGetCmd.RunE(cmd, []string{tt.subscriptionID})

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

// TestSubscriptionsGetRunE_JSONOutput tests the subscriptions get command with JSON output.
func TestSubscriptionsGetRunE_JSONOutput(t *testing.T) {
	mockClient := &subscriptionsMockAPIClient{
		getSubscriptionResp: &api.Subscription{
			ID:            "sub_123",
			CustomerID:    "cust_456",
			ProductID:     "prod_789",
			Status:        api.SubscriptionStatusActive,
			Interval:      "month",
			IntervalCount: 1,
			Price:         "19.99",
			Currency:      "USD",
			NextBillingAt: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupSubscriptionsMockFactories(mockClient)
	defer cleanup()

	cmd := newSubscriptionsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := subscriptionsGetCmd.RunE(cmd, []string{"sub_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sub_123") {
		t.Errorf("JSON output should contain subscription ID, got: %s", output)
	}
}

// TestSubscriptionsCreateRunE tests the subscriptions create command with mock API.
func TestSubscriptionsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.Subscription
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Subscription{
				ID:            "sub_new",
				CustomerID:    "cust_123",
				ProductID:     "prod_456",
				VariantID:     "var_789",
				Status:        api.SubscriptionStatusActive,
				Interval:      "month",
				IntervalCount: 1,
				Price:         "29.99",
				Currency:      "USD",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("failed to create subscription"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &subscriptionsMockAPIClient{
				createSubscriptionResp: tt.mockResp,
				createSubscriptionErr:  tt.mockErr,
			}
			cleanup, _ := setupSubscriptionsMockFactories(mockClient)
			defer cleanup()

			cmd := newSubscriptionsTestCmd()
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().String("product-id", "prod_456", "")
			cmd.Flags().String("variant-id", "var_789", "")
			cmd.Flags().String("interval", "month", "")
			cmd.Flags().Int("interval-count", 1, "")

			err := subscriptionsCreateCmd.RunE(cmd, []string{})

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

// TestSubscriptionsCreateRunE_JSONOutput tests the subscriptions create command with JSON output.
func TestSubscriptionsCreateRunE_JSONOutput(t *testing.T) {
	mockClient := &subscriptionsMockAPIClient{
		createSubscriptionResp: &api.Subscription{
			ID:            "sub_new",
			CustomerID:    "cust_123",
			ProductID:     "prod_456",
			Status:        api.SubscriptionStatusActive,
			Interval:      "month",
			IntervalCount: 1,
		},
	}
	cleanup, buf := setupSubscriptionsMockFactories(mockClient)
	defer cleanup()

	cmd := newSubscriptionsTestCmd()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("product-id", "prod_456", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().String("interval", "month", "")
	cmd.Flags().Int("interval-count", 1, "")
	_ = cmd.Flags().Set("output", "json")

	err := subscriptionsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sub_new") {
		t.Errorf("JSON output should contain subscription ID, got: %s", output)
	}
}

// TestSubscriptionsDeleteRunE tests the subscriptions delete command with mock API.
func TestSubscriptionsDeleteRunE(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		mockErr        error
		wantErr        bool
	}{
		{
			name:           "successful delete",
			subscriptionID: "sub_123",
			mockErr:        nil,
		},
		{
			name:           "delete fails",
			subscriptionID: "sub_456",
			mockErr:        errors.New("subscription already cancelled"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &subscriptionsMockAPIClient{
				deleteSubscriptionErr: tt.mockErr,
			}
			cleanup, _ := setupSubscriptionsMockFactories(mockClient)
			defer cleanup()

			cmd := newSubscriptionsTestCmd()

			err := subscriptionsDeleteCmd.RunE(cmd, []string{tt.subscriptionID})

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

// TestSubscriptionsListWithFilters tests list command with various filter flags.
func TestSubscriptionsListWithFilters(t *testing.T) {
	mockClient := &subscriptionsMockAPIClient{
		listSubscriptionsResp: &api.SubscriptionsListResponse{
			Items:      []api.Subscription{},
			TotalCount: 0,
		},
	}
	cleanup, _ := setupSubscriptionsMockFactories(mockClient)
	defer cleanup()

	cmd := newSubscriptionsTestCmd()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("product-id", "prod_456", "")
	cmd.Flags().String("status", "active", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")

	_ = cmd.Flags().Set("customer-id", "cust_123")
	_ = cmd.Flags().Set("product-id", "prod_456")
	_ = cmd.Flags().Set("status", "active")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "50")

	err := subscriptionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error with filters: %v", err)
	}
}

// TestSubscriptionsCreateWithAllFlags tests create command with all flags set.
func TestSubscriptionsCreateWithAllFlags(t *testing.T) {
	mockClient := &subscriptionsMockAPIClient{
		createSubscriptionResp: &api.Subscription{
			ID:            "sub_new",
			CustomerID:    "cust_123",
			ProductID:     "prod_456",
			VariantID:     "var_789",
			Status:        api.SubscriptionStatusActive,
			Interval:      "week",
			IntervalCount: 2,
		},
	}
	cleanup, _ := setupSubscriptionsMockFactories(mockClient)
	defer cleanup()

	cmd := newSubscriptionsTestCmd()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("product-id", "", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().String("interval", "", "")
	cmd.Flags().Int("interval-count", 1, "")

	_ = cmd.Flags().Set("customer-id", "cust_123")
	_ = cmd.Flags().Set("product-id", "prod_456")
	_ = cmd.Flags().Set("variant-id", "var_789")
	_ = cmd.Flags().Set("interval", "week")
	_ = cmd.Flags().Set("interval-count", "2")

	err := subscriptionsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSubscriptionsSubcommands verifies all subcommands are registered.
func TestSubscriptionsSubcommands(t *testing.T) {
	subcommands := subscriptionsCmd.Commands()
	expectedCmds := map[string]bool{"list": false, "get": false, "create": false, "delete": false}
	for _, cmd := range subcommands {
		switch {
		case cmd.Use == "list":
			expectedCmds["list"] = true
		case strings.HasPrefix(cmd.Use, "get"):
			expectedCmds["get"] = true
		case cmd.Use == "create":
			expectedCmds["create"] = true
		case strings.HasPrefix(cmd.Use, "delete"):
			expectedCmds["delete"] = true
		}
	}
	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}
