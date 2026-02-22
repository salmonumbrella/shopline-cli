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

func TestGiftsCmd(t *testing.T) {
	if giftsCmd.Use != "gifts" {
		t.Errorf("Expected Use 'gifts', got %q", giftsCmd.Use)
	}
	if giftsCmd.Short != "Manage gift promotions" {
		t.Errorf("Expected Short 'Manage gift promotions', got %q", giftsCmd.Short)
	}
}

func TestGiftsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":                   "List gift promotions",
		"get":                    "Get gift details",
		"create":                 "Create a gift promotion",
		"update":                 "Update a gift promotion",
		"update-quantity":        "Update gift quantity (documented endpoint)",
		"update-quantity-by-sku": "Bulk update gift quantity by SKU (documented endpoint)",
		"stocks":                 "Manage gift stocks (documented endpoints)",
		"activate":               "Activate a gift promotion",
		"deactivate":             "Deactivate a gift promotion",
		"delete":                 "Delete a gift promotion",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range giftsCmd.Commands() {
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

func TestGiftsListCmd(t *testing.T) {
	if giftsListCmd.Use != "list" {
		t.Errorf("Expected Use 'list', got %q", giftsListCmd.Use)
	}
}

func TestGiftsGetCmd(t *testing.T) {
	if giftsGetCmd.Use != "get [id]" {
		t.Errorf("Expected Use 'get [id]', got %q", giftsGetCmd.Use)
	}
}

func TestGiftsCreateCmd(t *testing.T) {
	if giftsCreateCmd.Use != "create" {
		t.Errorf("Expected Use 'create', got %q", giftsCreateCmd.Use)
	}
}

func TestGiftsActivateCmd(t *testing.T) {
	if giftsActivateCmd.Use != "activate <id>" {
		t.Errorf("Expected Use 'activate <id>', got %q", giftsActivateCmd.Use)
	}
}

func TestGiftsDeactivateCmd(t *testing.T) {
	if giftsDeactivateCmd.Use != "deactivate <id>" {
		t.Errorf("Expected Use 'deactivate <id>', got %q", giftsDeactivateCmd.Use)
	}
}

func TestGiftsDeleteCmd(t *testing.T) {
	if giftsDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use 'delete <id>', got %q", giftsDeleteCmd.Use)
	}
}

func TestGiftsListFlags(t *testing.T) {
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
			flag := giftsListCmd.Flags().Lookup(f.name)
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

func TestGiftsCreateFlags(t *testing.T) {
	flags := []string{"title", "description", "gift-product-id", "gift-variant-id", "trigger-type", "trigger-value", "quantity", "limit-per-user", "starts-at", "ends-at"}
	for _, flag := range flags {
		if giftsCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestGiftsDeleteFlags(t *testing.T) {
	flag := giftsDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("Expected flag 'yes'")
	}
	if flag.DefValue != "false" {
		t.Errorf("Expected default 'false', got %q", flag.DefValue)
	}
}

func TestGiftsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	err := giftsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := giftsGetCmd.RunE(cmd, []string{"gift_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("title", "Test Gift", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("gift-product-id", "prod_123", "")
	cmd.Flags().String("gift-variant-id", "", "")
	cmd.Flags().String("trigger-type", "min_purchase", "")
	cmd.Flags().Float64("trigger-value", 100.0, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	err := giftsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftsActivateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := giftsActivateCmd.RunE(cmd, []string{"gift_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftsDeactivateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := giftsDeactivateCmd.RunE(cmd, []string{"gift_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")
	err := giftsDeleteCmd.RunE(cmd, []string{"gift_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftsListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	err := giftsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftsCreateRunE_InvalidStartsAt(t *testing.T) {
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
	cmd.Flags().String("title", "Test Gift", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("gift-product-id", "prod_123", "")
	cmd.Flags().String("gift-variant-id", "", "")
	cmd.Flags().String("trigger-type", "min_purchase", "")
	cmd.Flags().Float64("trigger-value", 100.0, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "invalid-date", "")
	cmd.Flags().String("ends-at", "", "")
	err := giftsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for invalid starts-at format")
	}
}

func TestGiftsCreateRunE_InvalidEndsAt(t *testing.T) {
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
	cmd.Flags().String("title", "Test Gift", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("gift-product-id", "prod_123", "")
	cmd.Flags().String("gift-variant-id", "", "")
	cmd.Flags().String("trigger-type", "min_purchase", "")
	cmd.Flags().Float64("trigger-value", 100.0, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "invalid-date", "")
	err := giftsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for invalid ends-at format")
	}
}

// giftsMockAPIClient is a mock implementation of api.APIClient for gifts tests.
type giftsMockAPIClient struct {
	api.MockClient
	listGiftsResp      *api.GiftsListResponse
	listGiftsErr       error
	searchGiftsResp    *api.GiftsListResponse
	searchGiftsErr     error
	getGiftResp        *api.Gift
	getGiftErr         error
	createGiftResp     *api.Gift
	createGiftErr      error
	updateGiftResp     *api.Gift
	updateGiftErr      error
	updateQtyResp      *api.Gift
	updateQtyErr       error
	updateBySKUErr     error
	getStocksResp      json.RawMessage
	getStocksErr       error
	updateStocksResp   json.RawMessage
	updateStocksErr    error
	activateGiftResp   *api.Gift
	activateGiftErr    error
	deactivateGiftResp *api.Gift
	deactivateGiftErr  error
	deleteGiftErr      error
}

func (m *giftsMockAPIClient) ListGifts(ctx context.Context, opts *api.GiftsListOptions) (*api.GiftsListResponse, error) {
	return m.listGiftsResp, m.listGiftsErr
}

func (m *giftsMockAPIClient) SearchGifts(ctx context.Context, opts *api.GiftSearchOptions) (*api.GiftsListResponse, error) {
	return m.searchGiftsResp, m.searchGiftsErr
}

func (m *giftsMockAPIClient) GetGift(ctx context.Context, id string) (*api.Gift, error) {
	return m.getGiftResp, m.getGiftErr
}

func (m *giftsMockAPIClient) CreateGift(ctx context.Context, req *api.GiftCreateRequest) (*api.Gift, error) {
	return m.createGiftResp, m.createGiftErr
}

func (m *giftsMockAPIClient) UpdateGift(ctx context.Context, id string, req *api.GiftUpdateRequest) (*api.Gift, error) {
	return m.updateGiftResp, m.updateGiftErr
}

func (m *giftsMockAPIClient) UpdateGiftQuantity(ctx context.Context, id string, quantity int) (*api.Gift, error) {
	return m.updateQtyResp, m.updateQtyErr
}

func (m *giftsMockAPIClient) UpdateGiftsQuantityBySKU(ctx context.Context, sku string, quantity int) error {
	return m.updateBySKUErr
}

func (m *giftsMockAPIClient) GetGiftStocks(ctx context.Context, id string) (json.RawMessage, error) {
	return m.getStocksResp, m.getStocksErr
}

func (m *giftsMockAPIClient) UpdateGiftStocks(ctx context.Context, id string, body any) (json.RawMessage, error) {
	return m.updateStocksResp, m.updateStocksErr
}

func (m *giftsMockAPIClient) ActivateGift(ctx context.Context, id string) (*api.Gift, error) {
	return m.activateGiftResp, m.activateGiftErr
}

func (m *giftsMockAPIClient) DeactivateGift(ctx context.Context, id string) (*api.Gift, error) {
	return m.deactivateGiftResp, m.deactivateGiftErr
}

func (m *giftsMockAPIClient) DeleteGift(ctx context.Context, id string) error {
	return m.deleteGiftErr
}

// setupGiftsMockFactories sets up mock factories for gifts tests.
func setupGiftsMockFactories(mockClient *giftsMockAPIClient) (func(), *bytes.Buffer) {
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

// newGiftsTestCmd creates a test command with common flags for gifts tests.
func newGiftsTestCmd() *cobra.Command {
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

// TestGiftsListRunE tests the gifts list command with mock API.
func TestGiftsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.GiftsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.GiftsListResponse{
				Items: []api.Gift{
					{
						ID:              "gift_123",
						Title:           "Free Item Promo",
						GiftProductName: "Sample Product",
						TriggerType:     "min_purchase",
						TriggerValue:    50.00,
						Quantity:        100,
						QuantityUsed:    25,
						Status:          "active",
						StartsAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						EndsAt:          time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "gift_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.GiftsListResponse{
				Items:      []api.Gift{},
				TotalCount: 0,
			},
		},
		{
			name: "gift with unlimited quantity",
			mockResp: &api.GiftsListResponse{
				Items: []api.Gift{
					{
						ID:              "gift_unlimited",
						Title:           "Unlimited Gift",
						GiftProductName: "Free Sample",
						TriggerType:     "min_purchase",
						TriggerValue:    100.00,
						Quantity:        0,
						QuantityUsed:    50,
						Status:          "active",
					},
				},
				TotalCount: 1,
			},
			wantOutput: "gift_unlimited",
		},
		{
			name: "gift with zero dates",
			mockResp: &api.GiftsListResponse{
				Items: []api.Gift{
					{
						ID:              "gift_nodates",
						Title:           "No Dates Gift",
						GiftProductName: "Test Product",
						TriggerType:     "product_purchase",
						TriggerValue:    1.00,
						Status:          "scheduled",
					},
				},
				TotalCount: 1,
			},
			wantOutput: "gift_nodates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftsMockAPIClient{
				listGiftsResp: tt.mockResp,
				listGiftsErr:  tt.mockErr,
			}
			cleanup, buf := setupGiftsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("status", "", "")

			err := giftsListCmd.RunE(cmd, []string{})

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

// TestGiftsListRunE_JSONOutput tests the gifts list command with JSON output.
func TestGiftsListRunE_JSONOutput(t *testing.T) {
	mockClient := &giftsMockAPIClient{
		listGiftsResp: &api.GiftsListResponse{
			Items: []api.Gift{
				{
					ID:              "gift_json",
					Title:           "JSON Gift",
					GiftProductName: "Test Product",
					TriggerType:     "min_purchase",
					TriggerValue:    75.00,
					Status:          "active",
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupGiftsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := giftsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "gift_json") {
		t.Errorf("output %q should contain 'gift_json'", output)
	}
}

// TestGiftsGetRunE tests the gifts get command with mock API.
func TestGiftsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		giftID   string
		mockResp *api.Gift
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "successful get",
			giftID: "gift_123",
			mockResp: &api.Gift{
				ID:              "gift_123",
				Title:           "Test Gift",
				Description:     "A test gift promotion",
				GiftProductID:   "prod_456",
				GiftProductName: "Free Sample",
				GiftVariantID:   "var_789",
				TriggerType:     "min_purchase",
				TriggerValue:    50.00,
				Quantity:        100,
				QuantityUsed:    25,
				LimitPerUser:    2,
				Status:          "active",
				StartsAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndsAt:          time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				CreatedAt:       time.Date(2023, 12, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "gift not found",
			giftID:  "gift_999",
			mockErr: errors.New("gift not found"),
			wantErr: true,
		},
		{
			name:   "gift without variant",
			giftID: "gift_novariant",
			mockResp: &api.Gift{
				ID:              "gift_novariant",
				Title:           "No Variant Gift",
				Description:     "Gift without variant",
				GiftProductID:   "prod_123",
				GiftProductName: "Test Product",
				GiftVariantID:   "",
				TriggerType:     "product_purchase",
				TriggerValue:    1.00,
				Status:          "active",
				CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:   "gift with unlimited quantity",
			giftID: "gift_unlimited",
			mockResp: &api.Gift{
				ID:              "gift_unlimited",
				Title:           "Unlimited Gift",
				Description:     "Gift with unlimited quantity",
				GiftProductID:   "prod_123",
				GiftProductName: "Test Product",
				TriggerType:     "min_purchase",
				TriggerValue:    100.00,
				Quantity:        0,
				QuantityUsed:    75,
				LimitPerUser:    0,
				Status:          "active",
				CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:   "gift without dates",
			giftID: "gift_nodates",
			mockResp: &api.Gift{
				ID:              "gift_nodates",
				Title:           "No Dates Gift",
				Description:     "Gift without start/end dates",
				GiftProductID:   "prod_123",
				GiftProductName: "Test Product",
				TriggerType:     "min_purchase",
				TriggerValue:    50.00,
				Status:          "scheduled",
				CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftsMockAPIClient{
				getGiftResp: tt.mockResp,
				getGiftErr:  tt.mockErr,
			}
			cleanup, _ := setupGiftsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftsTestCmd()

			err := giftsGetCmd.RunE(cmd, []string{tt.giftID})

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

// TestGiftsGetRunE_JSONOutput tests the gifts get command with JSON output.
func TestGiftsGetRunE_JSONOutput(t *testing.T) {
	mockClient := &giftsMockAPIClient{
		getGiftResp: &api.Gift{
			ID:              "gift_json",
			Title:           "JSON Gift",
			GiftProductName: "Test Product",
			TriggerType:     "min_purchase",
			TriggerValue:    50.00,
			Status:          "active",
			CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupGiftsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := giftsGetCmd.RunE(cmd, []string{"gift_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "gift_json") {
		t.Errorf("output %q should contain 'gift_json'", output)
	}
}

// TestGiftsCreateRunE tests the gifts create command with mock API.
func TestGiftsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.Gift
		mockErr  error
		wantErr  bool
		startsAt string
		endsAt   string
	}{
		{
			name: "successful create",
			mockResp: &api.Gift{
				ID:              "gift_new",
				Title:           "New Gift",
				GiftProductID:   "prod_123",
				GiftProductName: "Test Product",
				TriggerType:     "min_purchase",
				TriggerValue:    100.00,
				Status:          "active",
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("failed to create gift"),
			wantErr: true,
		},
		{
			name: "create with dates",
			mockResp: &api.Gift{
				ID:              "gift_dated",
				Title:           "Dated Gift",
				GiftProductID:   "prod_123",
				GiftProductName: "Test Product",
				TriggerType:     "min_purchase",
				TriggerValue:    100.00,
				Status:          "scheduled",
			},
			startsAt: "2024-01-01T00:00:00Z",
			endsAt:   "2024-12-31T23:59:59Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftsMockAPIClient{
				createGiftResp: tt.mockResp,
				createGiftErr:  tt.mockErr,
			}
			cleanup, _ := setupGiftsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftsTestCmd()
			cmd.Flags().String("title", "Test Gift", "")
			cmd.Flags().String("description", "A test gift", "")
			cmd.Flags().String("gift-product-id", "prod_123", "")
			cmd.Flags().String("gift-variant-id", "var_456", "")
			cmd.Flags().String("trigger-type", "min_purchase", "")
			cmd.Flags().Float64("trigger-value", 100.0, "")
			cmd.Flags().Int("quantity", 50, "")
			cmd.Flags().Int("limit-per-user", 2, "")
			cmd.Flags().String("starts-at", tt.startsAt, "")
			cmd.Flags().String("ends-at", tt.endsAt, "")

			err := giftsCreateCmd.RunE(cmd, []string{})

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

// TestGiftsCreateRunE_JSONOutput tests the gifts create command with JSON output.
func TestGiftsCreateRunE_JSONOutput(t *testing.T) {
	mockClient := &giftsMockAPIClient{
		createGiftResp: &api.Gift{
			ID:              "gift_json_create",
			Title:           "JSON Gift",
			GiftProductID:   "prod_123",
			GiftProductName: "Test Product",
			TriggerType:     "min_purchase",
			TriggerValue:    100.00,
			Status:          "active",
		},
	}
	cleanup, buf := setupGiftsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftsTestCmd()
	cmd.Flags().String("title", "Test Gift", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("gift-product-id", "prod_123", "")
	cmd.Flags().String("gift-variant-id", "", "")
	cmd.Flags().String("trigger-type", "min_purchase", "")
	cmd.Flags().Float64("trigger-value", 100.0, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := giftsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "gift_json_create") {
		t.Errorf("output %q should contain 'gift_json_create'", output)
	}
}

// TestGiftsActivateRunE tests the gifts activate command with mock API.
func TestGiftsActivateRunE(t *testing.T) {
	tests := []struct {
		name     string
		giftID   string
		mockResp *api.Gift
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "successful activate",
			giftID: "gift_123",
			mockResp: &api.Gift{
				ID:     "gift_123",
				Title:  "Activated Gift",
				Status: "active",
			},
		},
		{
			name:    "gift not found",
			giftID:  "gift_999",
			mockErr: errors.New("gift not found"),
			wantErr: true,
		},
		{
			name:    "already active",
			giftID:  "gift_active",
			mockErr: errors.New("gift is already active"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftsMockAPIClient{
				activateGiftResp: tt.mockResp,
				activateGiftErr:  tt.mockErr,
			}
			cleanup, _ := setupGiftsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftsTestCmd()

			err := giftsActivateCmd.RunE(cmd, []string{tt.giftID})

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

// TestGiftsDeactivateRunE tests the gifts deactivate command with mock API.
func TestGiftsDeactivateRunE(t *testing.T) {
	tests := []struct {
		name     string
		giftID   string
		mockResp *api.Gift
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "successful deactivate",
			giftID: "gift_123",
			mockResp: &api.Gift{
				ID:     "gift_123",
				Title:  "Deactivated Gift",
				Status: "inactive",
			},
		},
		{
			name:    "gift not found",
			giftID:  "gift_999",
			mockErr: errors.New("gift not found"),
			wantErr: true,
		},
		{
			name:    "already inactive",
			giftID:  "gift_inactive",
			mockErr: errors.New("gift is already inactive"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftsMockAPIClient{
				deactivateGiftResp: tt.mockResp,
				deactivateGiftErr:  tt.mockErr,
			}
			cleanup, _ := setupGiftsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftsTestCmd()

			err := giftsDeactivateCmd.RunE(cmd, []string{tt.giftID})

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

// TestGiftsDeleteRunE tests the gifts delete command with mock API.
func TestGiftsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		giftID  string
		mockErr error
		wantErr bool
	}{
		{
			name:   "successful delete",
			giftID: "gift_123",
		},
		{
			name:    "gift not found",
			giftID:  "gift_999",
			mockErr: errors.New("gift not found"),
			wantErr: true,
		},
		{
			name:    "delete fails",
			giftID:  "gift_fail",
			mockErr: errors.New("failed to delete gift"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftsMockAPIClient{
				deleteGiftErr: tt.mockErr,
			}
			cleanup, _ := setupGiftsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftsTestCmd()
			_ = cmd.Flags().Set("yes", "true")

			err := giftsDeleteCmd.RunE(cmd, []string{tt.giftID})

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

func TestGiftsUpdateRunE_JSON(t *testing.T) {
	mockClient := &giftsMockAPIClient{
		updateGiftResp: &api.Gift{ID: "gift_123", Title: "Updated Gift"},
	}
	cleanup, buf := setupGiftsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftsTestCmd()
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("gift-product-id", "", "")
	cmd.Flags().String("gift-variant-id", "", "")
	cmd.Flags().String("trigger-type", "", "")
	cmd.Flags().Float64("trigger-value", 0, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("title", "Updated Gift")

	if err := giftsUpdateCmd.RunE(cmd, []string{"gift_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Updated Gift") {
		t.Fatalf("expected updated gift in output, got %q", buf.String())
	}
}

func TestGiftsUpdateQuantityRunE_JSON(t *testing.T) {
	mockClient := &giftsMockAPIClient{
		updateQtyResp: &api.Gift{ID: "gift_123", Quantity: 100},
	}
	cleanup, buf := setupGiftsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftsTestCmd()
	cmd.Flags().Int("quantity", 0, "")
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("quantity", "100")

	if err := giftsUpdateQuantityCmd.RunE(cmd, []string{"gift_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"quantity\"") {
		t.Fatalf("expected quantity in output, got %q", buf.String())
	}
}

func TestGiftsUpdateQuantityBySKURunE_JSON(t *testing.T) {
	mockClient := &giftsMockAPIClient{}
	cleanup, buf := setupGiftsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftsTestCmd()
	cmd.Flags().String("sku", "", "")
	cmd.Flags().Int("quantity", 0, "")
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("sku", "SKU-123")
	_ = cmd.Flags().Set("quantity", "50")

	if err := giftsUpdateQuantityBySKUCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"ok\"") {
		t.Fatalf("expected ok in output, got %q", buf.String())
	}
}

func TestGiftsStocksGetRunE_JSON(t *testing.T) {
	mockClient := &giftsMockAPIClient{
		getStocksResp: json.RawMessage(`{"items":[]}`),
	}
	cleanup, buf := setupGiftsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	if err := giftsStocksGetCmd.RunE(cmd, []string{"gift_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"items\"") {
		t.Fatalf("expected items in output, got %q", buf.String())
	}
}

func TestGiftsStocksUpdateRunE_JSON(t *testing.T) {
	mockClient := &giftsMockAPIClient{
		updateStocksResp: json.RawMessage(`{"updated":true}`),
	}
	cleanup, buf := setupGiftsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftsTestCmd()
	addJSONBodyFlags(cmd)
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"ok":true}`)

	if err := giftsStocksUpdateCmd.RunE(cmd, []string{"gift_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"updated\"") {
		t.Fatalf("expected updated in output, got %q", buf.String())
	}
}

// TestGiftsGetByFlag tests the --by flag on the gifts get command.
func TestGiftsGetByFlag(t *testing.T) {
	t.Run("resolves gift by title", func(t *testing.T) {
		mockClient := &giftsMockAPIClient{
			searchGiftsResp: &api.GiftsListResponse{
				Items:      []api.Gift{{ID: "gift_found", Title: "Free Item"}},
				TotalCount: 1,
			},
			getGiftResp: &api.Gift{
				ID:        "gift_found",
				Title:     "Free Item",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}
		cleanup, buf := setupGiftsMockFactories(mockClient)
		defer cleanup()

		cmd := newGiftsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Free Item")

		if err := giftsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "gift_found") {
			t.Errorf("expected output to contain 'gift_found', got: %s", buf.String())
		}
	})

	t.Run("errors when no match", func(t *testing.T) {
		mockClient := &giftsMockAPIClient{
			searchGiftsResp: &api.GiftsListResponse{
				Items:      []api.Gift{},
				TotalCount: 0,
			},
		}
		cleanup, _ := setupGiftsMockFactories(mockClient)
		defer cleanup()

		cmd := newGiftsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "nonexistent")

		err := giftsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when no gift found")
		}
		if !strings.Contains(err.Error(), "no gift found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("errors when search fails", func(t *testing.T) {
		mockClient := &giftsMockAPIClient{
			searchGiftsErr: errors.New("API error"),
		}
		cleanup, _ := setupGiftsMockFactories(mockClient)
		defer cleanup()

		cmd := newGiftsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Free Item")

		err := giftsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when search fails")
		}
		if !strings.Contains(err.Error(), "search failed") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("warns on multiple matches", func(t *testing.T) {
		mockClient := &giftsMockAPIClient{
			searchGiftsResp: &api.GiftsListResponse{
				Items: []api.Gift{
					{ID: "gift_1", Title: "Free Item A"},
					{ID: "gift_2", Title: "Free Item B"},
				},
				TotalCount: 2,
			},
			getGiftResp: &api.Gift{
				ID:        "gift_1",
				Title:     "Free Item A",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}
		cleanup, buf := setupGiftsMockFactories(mockClient)
		defer cleanup()

		cmd := newGiftsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Free Item")

		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)

		if err := giftsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "gift_1") {
			t.Errorf("expected output to contain 'gift_1', got: %s", buf.String())
		}
		if !strings.Contains(stderr.String(), "2 matches found") {
			t.Errorf("expected stderr warning about multiple matches, got: %s", stderr.String())
		}
	})

	t.Run("positional arg takes precedence over --by", func(t *testing.T) {
		mockClient := &giftsMockAPIClient{
			getGiftResp: &api.Gift{
				ID:        "gift_direct",
				Title:     "Direct Gift",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}
		cleanup, buf := setupGiftsMockFactories(mockClient)
		defer cleanup()

		cmd := newGiftsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "should-not-be-used")

		if err := giftsGetCmd.RunE(cmd, []string{"gift_direct"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "gift_direct") {
			t.Errorf("expected output to contain 'gift_direct', got: %s", buf.String())
		}
	})

	t.Run("errors with no arg and no --by", func(t *testing.T) {
		mockClient := &giftsMockAPIClient{}
		cleanup, _ := setupGiftsMockFactories(mockClient)
		defer cleanup()

		cmd := newGiftsTestCmd()
		cmd.Flags().String("by", "", "")

		err := giftsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error with no arg and no --by")
		}
		if !strings.Contains(err.Error(), "provide a resource ID") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
