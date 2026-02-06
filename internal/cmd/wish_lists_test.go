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

// wishListsMockAPIClient is a mock implementation of api.APIClient for wish lists tests.
type wishListsMockAPIClient struct {
	api.MockClient // embed base mock for unimplemented methods

	listWishListsResp     *api.WishListsListResponse
	listWishListsErr      error
	getWishListResp       *api.WishList
	getWishListErr        error
	createWishListResp    *api.WishList
	createWishListErr     error
	deleteWishListErr     error
	addWishListItemResp   *api.WishListItem
	addWishListItemErr    error
	removeWishListItemErr error
}

func (m *wishListsMockAPIClient) ListWishLists(ctx context.Context, opts *api.WishListsListOptions) (*api.WishListsListResponse, error) {
	return m.listWishListsResp, m.listWishListsErr
}

func (m *wishListsMockAPIClient) GetWishList(ctx context.Context, id string) (*api.WishList, error) {
	return m.getWishListResp, m.getWishListErr
}

func (m *wishListsMockAPIClient) CreateWishList(ctx context.Context, req *api.WishListCreateRequest) (*api.WishList, error) {
	return m.createWishListResp, m.createWishListErr
}

func (m *wishListsMockAPIClient) DeleteWishList(ctx context.Context, id string) error {
	return m.deleteWishListErr
}

func (m *wishListsMockAPIClient) AddWishListItem(ctx context.Context, wishListID string, req *api.WishListItemCreateRequest) (*api.WishListItem, error) {
	return m.addWishListItemResp, m.addWishListItemErr
}

func (m *wishListsMockAPIClient) RemoveWishListItem(ctx context.Context, wishListID, itemID string) error {
	return m.removeWishListItemErr
}

// setupWishListsMockFactories sets up mock factories for wish lists tests.
func setupWishListsMockFactories(mockClient *wishListsMockAPIClient) (func(), *bytes.Buffer) {
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

// newWishListsTestCmd creates a test command with common flags for wish lists tests.
func newWishListsTestCmd() *cobra.Command {
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

// TestWishListsListRunE tests the wish lists list command with mock API.
func TestWishListsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.WishListsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.WishListsListResponse{
				Items: []api.WishList{
					{
						ID:         "wl_123",
						CustomerID: "cust_456",
						Name:       "My Wishlist",
						ItemCount:  5,
						IsPublic:   true,
						CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "wl_123",
		},
		{
			name: "list with private wishlist",
			mockResp: &api.WishListsListResponse{
				Items: []api.WishList{
					{
						ID:         "wl_789",
						CustomerID: "cust_101",
						Name:       "Private List",
						ItemCount:  3,
						IsPublic:   false,
						CreatedAt:  time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "wl_789",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.WishListsListResponse{
				Items:      []api.WishList{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &wishListsMockAPIClient{
				listWishListsResp: tt.mockResp,
				listWishListsErr:  tt.mockErr,
			}
			cleanup, buf := setupWishListsMockFactories(mockClient)
			defer cleanup()

			cmd := newWishListsTestCmd()
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := wishListsListCmd.RunE(cmd, []string{})

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

// TestWishListsListJSONOutput tests JSON output for wish lists list command.
func TestWishListsListJSONOutput(t *testing.T) {
	mockClient := &wishListsMockAPIClient{
		listWishListsResp: &api.WishListsListResponse{
			Items: []api.WishList{
				{
					ID:         "wl_json",
					CustomerID: "cust_json",
					Name:       "JSON Wishlist",
					ItemCount:  2,
					IsPublic:   true,
					CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupWishListsMockFactories(mockClient)
	defer cleanup()

	cmd := newWishListsTestCmd()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := wishListsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "wl_json") {
		t.Errorf("JSON output should contain ID, got %q", output)
	}
}

// TestWishListsGetRunE tests the wish lists get command with mock API.
func TestWishListsGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		wishListID string
		mockResp   *api.WishList
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful get",
			wishListID: "wl_123",
			mockResp: &api.WishList{
				ID:          "wl_123",
				CustomerID:  "cust_456",
				Name:        "My Wishlist",
				Description: "A test wishlist",
				IsDefault:   true,
				IsPublic:    false,
				ItemCount:   3,
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "get with items",
			wishListID: "wl_456",
			mockResp: &api.WishList{
				ID:          "wl_456",
				CustomerID:  "cust_789",
				Name:        "Items Wishlist",
				Description: "Has items",
				IsDefault:   false,
				IsPublic:    true,
				ShareURL:    "https://shop.com/wishlist/abc123",
				ItemCount:   2,
				Items: []api.WishListItem{
					{
						ID:           "item_1",
						ProductID:    "prod_1",
						Title:        "Test Product",
						VariantTitle: "Size M",
						Price:        "29.99",
						Currency:     "USD",
						Available:    true,
						Notes:        "Want this!",
					},
					{
						ID:        "item_2",
						ProductID: "prod_2",
						Title:     "Another Product",
						Price:     "49.99",
						Currency:  "USD",
						Available: false,
					},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "wishlist not found",
			wishListID: "wl_999",
			mockErr:    errors.New("wishlist not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &wishListsMockAPIClient{
				getWishListResp: tt.mockResp,
				getWishListErr:  tt.mockErr,
			}
			cleanup, _ := setupWishListsMockFactories(mockClient)
			defer cleanup()

			cmd := newWishListsTestCmd()

			err := wishListsGetCmd.RunE(cmd, []string{tt.wishListID})

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

// TestWishListsGetJSONOutput tests JSON output for wish lists get command.
func TestWishListsGetJSONOutput(t *testing.T) {
	mockClient := &wishListsMockAPIClient{
		getWishListResp: &api.WishList{
			ID:          "wl_json",
			CustomerID:  "cust_json",
			Name:        "JSON Wishlist",
			Description: "Test description",
			IsDefault:   false,
			IsPublic:    true,
			ShareURL:    "https://shop.com/wishlist/json",
			ItemCount:   1,
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupWishListsMockFactories(mockClient)
	defer cleanup()

	cmd := newWishListsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := wishListsGetCmd.RunE(cmd, []string{"wl_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "wl_json") {
		t.Errorf("JSON output should contain ID, got %q", output)
	}
}

// TestWishListsCreateRunE tests the wish lists create command with mock API.
func TestWishListsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.WishList
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.WishList{
				ID:         "wl_new",
				CustomerID: "cust_123",
				Name:       "New Wishlist",
				IsDefault:  false,
				IsPublic:   false,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		{
			name: "create public wishlist",
			mockResp: &api.WishList{
				ID:         "wl_public",
				CustomerID: "cust_123",
				Name:       "Public Wishlist",
				IsDefault:  false,
				IsPublic:   true,
				ShareURL:   "https://shop.com/wishlist/public",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("failed to create wishlist"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &wishListsMockAPIClient{
				createWishListResp: tt.mockResp,
				createWishListErr:  tt.mockErr,
			}
			cleanup, _ := setupWishListsMockFactories(mockClient)
			defer cleanup()

			cmd := newWishListsTestCmd()
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().String("name", "Test Wishlist", "")
			cmd.Flags().String("description", "Test description", "")
			cmd.Flags().Bool("default", false, "")
			cmd.Flags().Bool("public", false, "")

			err := wishListsCreateCmd.RunE(cmd, []string{})

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

// TestWishListsCreateJSONOutput tests JSON output for wish lists create command.
func TestWishListsCreateJSONOutput(t *testing.T) {
	mockClient := &wishListsMockAPIClient{
		createWishListResp: &api.WishList{
			ID:         "wl_json_create",
			CustomerID: "cust_json",
			Name:       "JSON Created Wishlist",
			IsDefault:  false,
			IsPublic:   false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}
	cleanup, buf := setupWishListsMockFactories(mockClient)
	defer cleanup()

	cmd := newWishListsTestCmd()
	cmd.Flags().String("customer-id", "cust_json", "")
	cmd.Flags().String("name", "JSON Wishlist", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("default", false, "")
	cmd.Flags().Bool("public", false, "")
	_ = cmd.Flags().Set("output", "json")

	err := wishListsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "wl_json_create") {
		t.Errorf("JSON output should contain ID, got %q", output)
	}
}

// TestWishListsCreateGetClientError tests create command when getClient fails.
func TestWishListsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust-123", "")
	cmd.Flags().String("name", "My Wishlist", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("default", false, "")
	cmd.Flags().Bool("public", false, "")

	err := wishListsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestWishListsDeleteRunE tests the wish lists delete command with mock API.
func TestWishListsDeleteRunE(t *testing.T) {
	tests := []struct {
		name       string
		wishListID string
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful delete",
			wishListID: "wl_123",
			mockErr:    nil,
		},
		{
			name:       "delete fails",
			wishListID: "wl_456",
			mockErr:    errors.New("wishlist not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &wishListsMockAPIClient{
				deleteWishListErr: tt.mockErr,
			}
			cleanup, _ := setupWishListsMockFactories(mockClient)
			defer cleanup()

			cmd := newWishListsTestCmd()

			err := wishListsDeleteCmd.RunE(cmd, []string{tt.wishListID})

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

// TestWishListsAddItemRunE tests the wish lists add-item command with mock API.
func TestWishListsAddItemRunE(t *testing.T) {
	tests := []struct {
		name       string
		wishListID string
		mockResp   *api.WishListItem
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful add item",
			wishListID: "wl_123",
			mockResp: &api.WishListItem{
				ID:        "item_new",
				ProductID: "prod_456",
				Title:     "New Product",
				Price:     "39.99",
				Currency:  "USD",
				Available: true,
			},
		},
		{
			name:       "add item with variant",
			wishListID: "wl_123",
			mockResp: &api.WishListItem{
				ID:           "item_variant",
				ProductID:    "prod_789",
				VariantID:    "var_101",
				Title:        "Variant Product",
				VariantTitle: "Size L",
				Price:        "49.99",
				Currency:     "USD",
				Available:    true,
			},
		},
		{
			name:       "API error",
			wishListID: "wl_123",
			mockErr:    errors.New("product not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &wishListsMockAPIClient{
				addWishListItemResp: tt.mockResp,
				addWishListItemErr:  tt.mockErr,
			}
			cleanup, _ := setupWishListsMockFactories(mockClient)
			defer cleanup()

			cmd := newWishListsTestCmd()
			cmd.Flags().String("product-id", "prod_456", "")
			cmd.Flags().String("variant-id", "", "")
			cmd.Flags().Int("quantity", 1, "")
			cmd.Flags().Int("priority", 0, "")
			cmd.Flags().String("notes", "", "")

			err := wishListsAddItemCmd.RunE(cmd, []string{tt.wishListID})

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

// TestWishListsAddItemJSONOutput tests JSON output for wish lists add-item command.
func TestWishListsAddItemJSONOutput(t *testing.T) {
	mockClient := &wishListsMockAPIClient{
		addWishListItemResp: &api.WishListItem{
			ID:        "item_json",
			ProductID: "prod_json",
			Title:     "JSON Product",
			Price:     "59.99",
			Currency:  "USD",
			Available: true,
		},
	}
	cleanup, buf := setupWishListsMockFactories(mockClient)
	defer cleanup()

	cmd := newWishListsTestCmd()
	cmd.Flags().String("product-id", "prod_json", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().Int("quantity", 1, "")
	cmd.Flags().Int("priority", 0, "")
	cmd.Flags().String("notes", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := wishListsAddItemCmd.RunE(cmd, []string{"wl_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "item_json") {
		t.Errorf("JSON output should contain ID, got %q", output)
	}
}

// TestWishListsRemoveItemRunE tests the wish lists remove-item command with mock API.
func TestWishListsRemoveItemRunE(t *testing.T) {
	tests := []struct {
		name       string
		wishListID string
		itemID     string
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful remove item",
			wishListID: "wl_123",
			itemID:     "item_456",
			mockErr:    nil,
		},
		{
			name:       "remove fails",
			wishListID: "wl_789",
			itemID:     "item_nonexistent",
			mockErr:    errors.New("item not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &wishListsMockAPIClient{
				removeWishListItemErr: tt.mockErr,
			}
			cleanup, _ := setupWishListsMockFactories(mockClient)
			defer cleanup()

			cmd := newWishListsTestCmd()

			err := wishListsRemoveItemCmd.RunE(cmd, []string{tt.wishListID, tt.itemID})

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

// Test existing dry-run and getClient error tests
func TestWishListsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := wishListsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestWishListsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := wishListsGetCmd.RunE(cmd, []string{"wishlist-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestWishListsCreateDryRun(t *testing.T) {
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
	cmd.Flags().String("customer-id", "cust-123", "")
	cmd.Flags().String("name", "My Wishlist", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("default", false, "")
	cmd.Flags().Bool("public", false, "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := wishListsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error in dry-run: %v", err)
	}
}

func TestWishListsDeleteDryRun(t *testing.T) {
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
	_ = cmd.Flags().Set("dry-run", "true")

	err := wishListsDeleteCmd.RunE(cmd, []string{"wishlist-123"})
	if err != nil {
		t.Errorf("Unexpected error in dry-run: %v", err)
	}
}

func TestWishListsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := wishListsDeleteCmd.RunE(cmd, []string{"wishlist-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestWishListsAddItemDryRun(t *testing.T) {
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
	cmd.Flags().String("product-id", "prod-123", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().Int("quantity", 1, "")
	cmd.Flags().Int("priority", 0, "")
	cmd.Flags().String("notes", "", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := wishListsAddItemCmd.RunE(cmd, []string{"wishlist-123"})
	if err != nil {
		t.Errorf("Unexpected error in dry-run: %v", err)
	}
}

func TestWishListsAddItemGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("product-id", "prod-123", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().Int("quantity", 1, "")
	cmd.Flags().Int("priority", 0, "")
	cmd.Flags().String("notes", "", "")

	err := wishListsAddItemCmd.RunE(cmd, []string{"wishlist-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestWishListsRemoveItemDryRun(t *testing.T) {
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
	_ = cmd.Flags().Set("dry-run", "true")

	err := wishListsRemoveItemCmd.RunE(cmd, []string{"wishlist-123", "item-456"})
	if err != nil {
		t.Errorf("Unexpected error in dry-run: %v", err)
	}
}

func TestWishListsRemoveItemGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := wishListsRemoveItemCmd.RunE(cmd, []string{"wishlist-123", "item-456"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestWishListsListFlags(t *testing.T) {
	flags := wishListsListCmd.Flags()

	if flags.Lookup("customer-id") == nil {
		t.Error("Expected customer-id flag")
	}
	if flags.Lookup("page") == nil {
		t.Error("Expected page flag")
	}
	if flags.Lookup("page-size") == nil {
		t.Error("Expected page-size flag")
	}
}

func TestWishListsCreateFlags(t *testing.T) {
	flags := wishListsCreateCmd.Flags()

	if flags.Lookup("customer-id") == nil {
		t.Error("Expected customer-id flag")
	}
	if flags.Lookup("name") == nil {
		t.Error("Expected name flag")
	}
	if flags.Lookup("description") == nil {
		t.Error("Expected description flag")
	}
	if flags.Lookup("default") == nil {
		t.Error("Expected default flag")
	}
	if flags.Lookup("public") == nil {
		t.Error("Expected public flag")
	}
}

func TestWishListsAddItemFlags(t *testing.T) {
	flags := wishListsAddItemCmd.Flags()

	if flags.Lookup("product-id") == nil {
		t.Error("Expected product-id flag")
	}
	if flags.Lookup("variant-id") == nil {
		t.Error("Expected variant-id flag")
	}
	if flags.Lookup("quantity") == nil {
		t.Error("Expected quantity flag")
	}
	if flags.Lookup("priority") == nil {
		t.Error("Expected priority flag")
	}
	if flags.Lookup("notes") == nil {
		t.Error("Expected notes flag")
	}
}

func TestWishListsCommandStructure(t *testing.T) {
	if wishListsCmd.Use != "wish-lists" {
		t.Errorf("Expected Use 'wish-lists', got %s", wishListsCmd.Use)
	}

	subcommands := wishListsCmd.Commands()
	expectedCmds := map[string]bool{
		"list":        false,
		"get":         false,
		"create":      false,
		"delete":      false,
		"add-item":    false,
		"remove-item": false,
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

// TestWishListsListWithCustomerIDFilter tests list command with customer-id filter.
func TestWishListsListWithCustomerIDFilter(t *testing.T) {
	mockClient := &wishListsMockAPIClient{
		listWishListsResp: &api.WishListsListResponse{
			Items: []api.WishList{
				{
					ID:         "wl_filtered",
					CustomerID: "cust_specific",
					Name:       "Filtered Wishlist",
					ItemCount:  1,
					IsPublic:   false,
					CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupWishListsMockFactories(mockClient)
	defer cleanup()

	cmd := newWishListsTestCmd()
	cmd.Flags().String("customer-id", "cust_specific", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("customer-id", "cust_specific")

	err := wishListsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "wl_filtered") {
		t.Errorf("output should contain filtered wishlist ID, got %q", output)
	}
}
