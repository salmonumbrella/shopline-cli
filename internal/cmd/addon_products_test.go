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

// TestAddonProductsCommandSetup verifies addon-products command initialization
func TestAddonProductsCommandSetup(t *testing.T) {
	if addonProductsCmd.Use != "addon-products" {
		t.Errorf("expected Use 'addon-products', got %q", addonProductsCmd.Use)
	}
	if addonProductsCmd.Short != "Manage add-on product bundles" {
		t.Errorf("expected Short 'Manage add-on product bundles', got %q", addonProductsCmd.Short)
	}
}

// TestAddonProductsSubcommands verifies all subcommands are registered
func TestAddonProductsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List add-on products",
		"get":    "Get add-on product details",
		"create": "Create an add-on product",
		"delete": "Delete an add-on product",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range addonProductsCmd.Commands() {
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

// TestAddonProductsListFlags verifies list command flags exist with correct defaults
func TestAddonProductsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"product-id", ""},
		{"status", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := addonProductsListCmd.Flags().Lookup(f.name)
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

// TestAddonProductsCreateFlags verifies create command flags exist with correct defaults
func TestAddonProductsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
		required     bool
	}{
		{"title", "", true},
		{"product-id", "", true},
		{"variant-id", "", false},
		{"price", "", false},
		{"quantity", "1", false},
		{"position", "0", false},
		{"description", "", false},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := addonProductsCreateCmd.Flags().Lookup(f.name)
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

// TestAddonProductsGetCmd verifies get command configuration
func TestAddonProductsGetCmd(t *testing.T) {
	if addonProductsGetCmd.Use != "get [id]" {
		t.Errorf("expected Use 'get [id]', got %q", addonProductsGetCmd.Use)
	}
	if addonProductsGetCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestAddonProductsDeleteCmd verifies delete command configuration
func TestAddonProductsDeleteCmd(t *testing.T) {
	if addonProductsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", addonProductsDeleteCmd.Use)
	}
	if addonProductsDeleteCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// addonProductsMockAPIClient is a mock implementation of api.APIClient for addon products tests.
type addonProductsMockAPIClient struct {
	api.MockClient
	listAddonProductsResp   *api.AddonProductsListResponse
	listAddonProductsErr    error
	searchAddonProductsResp *api.AddonProductsListResponse
	searchAddonProductsErr  error
	getAddonProductResp     *api.AddonProduct
	getAddonProductErr      error
	createAddonProductResp  *api.AddonProduct
	createAddonProductErr   error
	deleteAddonProductErr   error
}

func (m *addonProductsMockAPIClient) ListAddonProducts(ctx context.Context, opts *api.AddonProductsListOptions) (*api.AddonProductsListResponse, error) {
	return m.listAddonProductsResp, m.listAddonProductsErr
}

func (m *addonProductsMockAPIClient) SearchAddonProducts(ctx context.Context, opts *api.AddonProductSearchOptions) (*api.AddonProductsListResponse, error) {
	return m.searchAddonProductsResp, m.searchAddonProductsErr
}

func (m *addonProductsMockAPIClient) GetAddonProduct(ctx context.Context, id string) (*api.AddonProduct, error) {
	return m.getAddonProductResp, m.getAddonProductErr
}

func (m *addonProductsMockAPIClient) CreateAddonProduct(ctx context.Context, req *api.AddonProductCreateRequest) (*api.AddonProduct, error) {
	return m.createAddonProductResp, m.createAddonProductErr
}

func (m *addonProductsMockAPIClient) DeleteAddonProduct(ctx context.Context, id string) error {
	return m.deleteAddonProductErr
}

// setupAddonProductsMockFactories sets up mock factories for addon products tests.
func setupAddonProductsMockFactories(mockClient *addonProductsMockAPIClient) (func(), *bytes.Buffer) {
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

// newAddonProductsTestCmd creates a test command with common flags for addon products tests.
func newAddonProductsTestCmd() *cobra.Command {
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

// TestAddonProductsListRunE tests the addon products list command with mock API.
func TestAddonProductsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.AddonProductsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.AddonProductsListResponse{
				Items: []api.AddonProduct{
					{
						ID:        "addon_123",
						Title:     "Test Addon",
						ProductID: "prod_123",
						Price:     "9.99",
						Currency:  "USD",
						Status:    "active",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "addon_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.AddonProductsListResponse{
				Items:      []api.AddonProduct{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &addonProductsMockAPIClient{
				listAddonProductsResp: tt.mockResp,
				listAddonProductsErr:  tt.mockErr,
			}
			cleanup, buf := setupAddonProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newAddonProductsTestCmd()
			cmd.Flags().String("product-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := addonProductsListCmd.RunE(cmd, []string{})

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

// TestAddonProductsGetRunE tests the addon products get command with mock API.
func TestAddonProductsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockResp *api.AddonProduct
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "addon_123",
			mockResp: &api.AddonProduct{
				ID:          "addon_123",
				Title:       "Test Addon",
				ProductID:   "prod_123",
				VariantID:   "var_123",
				Price:       "9.99",
				Currency:    "USD",
				Quantity:    1,
				Position:    1,
				Status:      "active",
				Description: "Test description",
			},
		},
		{
			name:    "not found",
			id:      "addon_999",
			mockErr: errors.New("addon product not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &addonProductsMockAPIClient{
				getAddonProductResp: tt.mockResp,
				getAddonProductErr:  tt.mockErr,
			}
			cleanup, _ := setupAddonProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newAddonProductsTestCmd()

			err := addonProductsGetCmd.RunE(cmd, []string{tt.id})

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

// TestAddonProductsCreateRunE tests the addon products create command with mock API.
func TestAddonProductsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		dryRun   bool
		mockResp *api.AddonProduct
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "dry run",
			dryRun: true,
		},
		{
			name: "successful create",
			mockResp: &api.AddonProduct{
				ID:        "addon_new",
				Title:     "New Addon",
				ProductID: "prod_123",
				Price:     "9.99",
				Currency:  "USD",
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("create failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &addonProductsMockAPIClient{
				createAddonProductResp: tt.mockResp,
				createAddonProductErr:  tt.mockErr,
			}
			cleanup, _ := setupAddonProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newAddonProductsTestCmd()
			cmd.Flags().String("title", "New Addon", "")
			cmd.Flags().String("product-id", "prod_123", "")
			cmd.Flags().String("variant-id", "", "")
			cmd.Flags().String("price", "9.99", "")
			cmd.Flags().Int("quantity", 1, "")
			cmd.Flags().Int("position", 0, "")
			cmd.Flags().String("description", "", "")
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := addonProductsCreateCmd.RunE(cmd, []string{})

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

// TestAddonProductsDeleteRunE tests the addon products delete command with mock API.
func TestAddonProductsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		dryRun  bool
		mockErr error
		wantErr bool
	}{
		{
			name:   "dry run",
			id:     "addon_123",
			dryRun: true,
		},
		{
			name: "successful delete",
			id:   "addon_123",
		},
		{
			name:    "delete fails",
			id:      "addon_123",
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &addonProductsMockAPIClient{
				deleteAddonProductErr: tt.mockErr,
			}
			cleanup, _ := setupAddonProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newAddonProductsTestCmd()
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := addonProductsDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestAddonProductsGetByFlag tests the --by flag on the addon-products get command.
func TestAddonProductsGetByFlag(t *testing.T) {
	t.Run("resolves addon product by title", func(t *testing.T) {
		mockClient := &addonProductsMockAPIClient{
			searchAddonProductsResp: &api.AddonProductsListResponse{
				Items:      []api.AddonProduct{{ID: "addon_found", Title: "Extra Warranty"}},
				TotalCount: 1,
			},
			getAddonProductResp: &api.AddonProduct{
				ID:    "addon_found",
				Title: "Extra Warranty",
			},
		}
		cleanup, buf := setupAddonProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newAddonProductsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Extra Warranty")

		if err := addonProductsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "addon_found") {
			t.Errorf("expected output to contain 'addon_found', got: %s", buf.String())
		}
	})

	t.Run("errors when no match", func(t *testing.T) {
		mockClient := &addonProductsMockAPIClient{
			searchAddonProductsResp: &api.AddonProductsListResponse{
				Items:      []api.AddonProduct{},
				TotalCount: 0,
			},
		}
		cleanup, _ := setupAddonProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newAddonProductsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "nonexistent")

		err := addonProductsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when no addon product found")
		}
		if !strings.Contains(err.Error(), "no addon product found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("errors when search fails", func(t *testing.T) {
		mockClient := &addonProductsMockAPIClient{
			searchAddonProductsErr: errors.New("API error"),
		}
		cleanup, _ := setupAddonProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newAddonProductsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Extra Warranty")

		err := addonProductsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when search fails")
		}
		if !strings.Contains(err.Error(), "search failed") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("warns on multiple matches", func(t *testing.T) {
		mockClient := &addonProductsMockAPIClient{
			searchAddonProductsResp: &api.AddonProductsListResponse{
				Items: []api.AddonProduct{
					{ID: "addon_1", Title: "Extra Warranty A"},
					{ID: "addon_2", Title: "Extra Warranty B"},
				},
				TotalCount: 2,
			},
			getAddonProductResp: &api.AddonProduct{
				ID:    "addon_1",
				Title: "Extra Warranty A",
			},
		}
		cleanup, buf := setupAddonProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newAddonProductsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Extra Warranty")

		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)

		if err := addonProductsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "addon_1") {
			t.Errorf("expected output to contain 'addon_1', got: %s", buf.String())
		}
		if !strings.Contains(stderr.String(), "2 matches found") {
			t.Errorf("expected stderr warning about multiple matches, got: %s", stderr.String())
		}
	})

	t.Run("positional arg takes precedence over --by", func(t *testing.T) {
		mockClient := &addonProductsMockAPIClient{
			getAddonProductResp: &api.AddonProduct{
				ID:    "addon_direct",
				Title: "Direct Addon",
			},
		}
		cleanup, buf := setupAddonProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newAddonProductsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "should-not-be-used")

		if err := addonProductsGetCmd.RunE(cmd, []string{"addon_direct"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "addon_direct") {
			t.Errorf("expected output to contain 'addon_direct', got: %s", buf.String())
		}
	})

	t.Run("errors with no arg and no --by", func(t *testing.T) {
		mockClient := &addonProductsMockAPIClient{}
		cleanup, _ := setupAddonProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newAddonProductsTestCmd()
		cmd.Flags().String("by", "", "")

		err := addonProductsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error with no arg and no --by")
		}
		if !strings.Contains(err.Error(), "provide a resource ID") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
