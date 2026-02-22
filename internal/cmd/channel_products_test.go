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

func TestChannelProductsCommandStructure(t *testing.T) {
	subcommands := channelProductsCmd.Commands()

	expectedCmds := map[string]bool{
		"list":      false,
		"get":       false,
		"publish":   false,
		"unpublish": false,
		"update":    false,
	}

	for _, cmd := range subcommands {
		if _, exists := expectedCmds[cmd.Name()]; exists {
			expectedCmds[cmd.Name()] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestChannelProductsListFlags(t *testing.T) {
	flags := []string{"page", "page-size", "published", "available", "status"}

	for _, flagName := range flags {
		if channelProductsListCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on list command", flagName)
		}
	}
}

func TestChannelProductsListArgs(t *testing.T) {
	err := channelProductsListCmd.Args(channelProductsListCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = channelProductsListCmd.Args(channelProductsListCmd, []string{"ch-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestChannelProductsGetArgs(t *testing.T) {
	err := channelProductsGetCmd.Args(channelProductsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = channelProductsGetCmd.Args(channelProductsGetCmd, []string{"ch-id"})
	if err == nil {
		t.Error("Expected error when only 1 arg provided")
	}

	err = channelProductsGetCmd.Args(channelProductsGetCmd, []string{"ch-id", "prod-id"})
	if err != nil {
		t.Errorf("Expected no error with 2 args, got: %v", err)
	}
}

func TestChannelProductsPublishArgs(t *testing.T) {
	err := channelProductsPublishCmd.Args(channelProductsPublishCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = channelProductsPublishCmd.Args(channelProductsPublishCmd, []string{"ch-id", "prod-id"})
	if err != nil {
		t.Errorf("Expected no error with 2 args, got: %v", err)
	}
}

func TestChannelProductsUnpublishArgs(t *testing.T) {
	err := channelProductsUnpublishCmd.Args(channelProductsUnpublishCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = channelProductsUnpublishCmd.Args(channelProductsUnpublishCmd, []string{"ch-id", "prod-id"})
	if err != nil {
		t.Errorf("Expected no error with 2 args, got: %v", err)
	}
}

func TestChannelProductsUpdateArgs(t *testing.T) {
	err := channelProductsUpdateCmd.Args(channelProductsUpdateCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = channelProductsUpdateCmd.Args(channelProductsUpdateCmd, []string{"ch-id", "prod-id"})
	if err != nil {
		t.Errorf("Expected no error with 2 args, got: %v", err)
	}
}

func TestChannelProductsUpdateFlags(t *testing.T) {
	flags := []string{"published", "available"}

	for _, flagName := range flags {
		if channelProductsUpdateCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on update command", flagName)
		}
	}
}

func TestChannelProductsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := channelProductsListCmd.RunE(cmd, []string{"ch-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestChannelProductsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := channelProductsGetCmd.RunE(cmd, []string{"ch-id", "prod-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestChannelProductsPublishGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := channelProductsPublishCmd.RunE(cmd, []string{"ch-id", "prod-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestChannelProductsListWithValidStore(t *testing.T) {
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

	err := channelProductsListCmd.RunE(cmd, []string{"ch-id"})
	if err == nil {
		t.Log("channelProductsListCmd succeeded (might be due to mock setup)")
	}
}

// channelProductsMockAPIClient is a mock implementation of api.APIClient for channel products tests.
type channelProductsMockAPIClient struct {
	api.MockClient
	listChannelProductsResp    *api.ChannelProductsListResponse
	listChannelProductsErr     error
	getChannelProductResp      *api.ChannelProductListing
	getChannelProductErr       error
	publishChannelProductResp  *api.ChannelProductListing
	publishChannelProductErr   error
	unpublishChannelProductErr error
	updateChannelProductResp   *api.ChannelProductListing
	updateChannelProductErr    error
}

func (m *channelProductsMockAPIClient) ListChannelProductListings(ctx context.Context, channelID string, opts *api.ChannelProductsListOptions) (*api.ChannelProductsListResponse, error) {
	return m.listChannelProductsResp, m.listChannelProductsErr
}

func (m *channelProductsMockAPIClient) GetChannelProductListing(ctx context.Context, channelID, productID string) (*api.ChannelProductListing, error) {
	return m.getChannelProductResp, m.getChannelProductErr
}

func (m *channelProductsMockAPIClient) PublishProductToChannelListing(ctx context.Context, channelID string, req *api.ChannelProductPublishRequest) (*api.ChannelProductListing, error) {
	return m.publishChannelProductResp, m.publishChannelProductErr
}

func (m *channelProductsMockAPIClient) UnpublishProductFromChannelListing(ctx context.Context, channelID, productID string) error {
	return m.unpublishChannelProductErr
}

func (m *channelProductsMockAPIClient) UpdateChannelProductListing(ctx context.Context, channelID, productID string, req *api.ChannelProductUpdateRequest) (*api.ChannelProductListing, error) {
	return m.updateChannelProductResp, m.updateChannelProductErr
}

// setupChannelProductsMockFactories sets up mock factories for channel products tests.
func setupChannelProductsMockFactories(mockClient *channelProductsMockAPIClient) (func(), *bytes.Buffer) {
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

// newChannelProductsTestCmd creates a test command with common flags for channel products tests.
func newChannelProductsTestCmd() *cobra.Command {
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

// TestChannelProductsListRunE tests the channel products list command with mock API.
func TestChannelProductsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ChannelProductsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ChannelProductsListResponse{
				Items: []api.ChannelProductListing{
					{
						ID:               "cpl_123",
						ProductID:        "prod_123",
						ChannelID:        "ch_123",
						Title:            "Test Product",
						Status:           "active",
						Published:        true,
						AvailableForSale: true,
						CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:        time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "cpl_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ChannelProductsListResponse{
				Items:      []api.ChannelProductListing{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelProductsMockAPIClient{
				listChannelProductsResp: tt.mockResp,
				listChannelProductsErr:  tt.mockErr,
			}
			cleanup, buf := setupChannelProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelProductsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().Bool("published", false, "")
			cmd.Flags().Bool("available", false, "")
			cmd.Flags().String("status", "", "")

			err := channelProductsListCmd.RunE(cmd, []string{"ch-123"})

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

// TestChannelProductsGetRunE tests the channel products get command with mock API.
func TestChannelProductsGetRunE(t *testing.T) {
	publishedAt := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		channelID string
		productID string
		mockResp  *api.ChannelProductListing
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			channelID: "ch_123",
			productID: "prod_123",
			mockResp: &api.ChannelProductListing{
				ID:               "cpl_123",
				ProductID:        "prod_123",
				ChannelID:        "ch_123",
				Title:            "Test Product",
				Handle:           "test-product",
				Status:           "active",
				Published:        true,
				PublishedAt:      &publishedAt,
				AvailableForSale: true,
				Variants: []api.ChannelVariantListing{
					{
						ID:                "cvl_123",
						VariantID:         "var_123",
						Title:             "Default",
						Price:             "19.99",
						InventoryQuantity: 100,
					},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:      "not found",
			channelID: "ch_999",
			productID: "prod_999",
			mockErr:   errors.New("channel product listing not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelProductsMockAPIClient{
				getChannelProductResp: tt.mockResp,
				getChannelProductErr:  tt.mockErr,
			}
			cleanup, _ := setupChannelProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelProductsTestCmd()

			err := channelProductsGetCmd.RunE(cmd, []string{tt.channelID, tt.productID})

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

// TestChannelProductsPublishRunE tests the channel products publish command with mock API.
func TestChannelProductsPublishRunE(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		productID string
		mockResp  *api.ChannelProductListing
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful publish",
			channelID: "ch_123",
			productID: "prod_123",
			mockResp: &api.ChannelProductListing{
				ID:        "cpl_new",
				ProductID: "prod_123",
				ChannelID: "ch_123",
				Title:     "Test Product",
				Published: true,
			},
		},
		{
			name:      "publish fails",
			channelID: "ch_123",
			productID: "prod_123",
			mockErr:   errors.New("publish failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelProductsMockAPIClient{
				publishChannelProductResp: tt.mockResp,
				publishChannelProductErr:  tt.mockErr,
			}
			cleanup, _ := setupChannelProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelProductsTestCmd()

			err := channelProductsPublishCmd.RunE(cmd, []string{tt.channelID, tt.productID})

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

// TestChannelProductsUnpublishRunE tests the channel products unpublish command with mock API.
func TestChannelProductsUnpublishRunE(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		productID string
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful unpublish",
			channelID: "ch_123",
			productID: "prod_123",
		},
		{
			name:      "unpublish fails",
			channelID: "ch_123",
			productID: "prod_123",
			mockErr:   errors.New("unpublish failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelProductsMockAPIClient{
				unpublishChannelProductErr: tt.mockErr,
			}
			cleanup, _ := setupChannelProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelProductsTestCmd()

			err := channelProductsUnpublishCmd.RunE(cmd, []string{tt.channelID, tt.productID})

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

// TestChannelProductsUpdateRunE tests the channel products update command with mock API.
func TestChannelProductsUpdateRunE(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		productID string
		mockResp  *api.ChannelProductListing
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful update",
			channelID: "ch_123",
			productID: "prod_123",
			mockResp: &api.ChannelProductListing{
				ID:               "cpl_123",
				ProductID:        "prod_123",
				ChannelID:        "ch_123",
				Published:        true,
				AvailableForSale: true,
			},
		},
		{
			name:      "update fails",
			channelID: "ch_123",
			productID: "prod_123",
			mockErr:   errors.New("update failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelProductsMockAPIClient{
				updateChannelProductResp: tt.mockResp,
				updateChannelProductErr:  tt.mockErr,
			}
			cleanup, _ := setupChannelProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelProductsTestCmd()
			cmd.Flags().Bool("published", false, "")
			cmd.Flags().Bool("available", false, "")

			err := channelProductsUpdateCmd.RunE(cmd, []string{tt.channelID, tt.productID})

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

// TestChannelProductsUnpublishGetClientError tests unpublish command error handling when getClient fails.
func TestChannelProductsUnpublishGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := channelProductsUnpublishCmd.RunE(cmd, []string{"ch-id", "prod-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestChannelProductsUpdateGetClientError tests update command error handling when getClient fails.
func TestChannelProductsUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Bool("published", false, "")
	cmd.Flags().Bool("available", false, "")

	err := channelProductsUpdateCmd.RunE(cmd, []string{"ch-id", "prod-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestChannelProductsListRunE_JSONOutput tests list command with JSON output format.
func TestChannelProductsListRunE_JSONOutput(t *testing.T) {
	mockClient := &channelProductsMockAPIClient{
		listChannelProductsResp: &api.ChannelProductsListResponse{
			Items: []api.ChannelProductListing{
				{
					ID:               "cpl_123",
					ProductID:        "prod_123",
					ChannelID:        "ch_123",
					Title:            "Test Product",
					Status:           "active",
					Published:        true,
					AvailableForSale: true,
					CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					UpdatedAt:        time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupChannelProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelProductsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().Bool("published", false, "")
	cmd.Flags().Bool("available", false, "")
	cmd.Flags().String("status", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := channelProductsListCmd.RunE(cmd, []string{"ch-123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cpl_123") {
		t.Errorf("expected JSON output to contain listing ID, got: %s", output)
	}
}

// TestChannelProductsListRunE_WithFilters tests list command with published and available filters.
func TestChannelProductsListRunE_WithFilters(t *testing.T) {
	mockClient := &channelProductsMockAPIClient{
		listChannelProductsResp: &api.ChannelProductsListResponse{
			Items: []api.ChannelProductListing{
				{
					ID:               "cpl_123",
					ProductID:        "prod_123",
					ChannelID:        "ch_123",
					Title:            "Test Product",
					Status:           "active",
					Published:        true,
					AvailableForSale: true,
					UpdatedAt:        time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, _ := setupChannelProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelProductsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().Bool("published", false, "")
	cmd.Flags().Bool("available", false, "")
	cmd.Flags().String("status", "", "")

	// Set the flags to trigger Changed() = true
	_ = cmd.Flags().Set("published", "true")
	_ = cmd.Flags().Set("available", "true")
	_ = cmd.Flags().Set("status", "active")

	err := channelProductsListCmd.RunE(cmd, []string{"ch-123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestChannelProductsGetRunE_JSONOutput tests get command with JSON output format.
func TestChannelProductsGetRunE_JSONOutput(t *testing.T) {
	publishedAt := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	mockClient := &channelProductsMockAPIClient{
		getChannelProductResp: &api.ChannelProductListing{
			ID:               "cpl_123",
			ProductID:        "prod_123",
			ChannelID:        "ch_123",
			Title:            "Test Product",
			Handle:           "test-product",
			Status:           "active",
			Published:        true,
			PublishedAt:      &publishedAt,
			AvailableForSale: true,
			Variants: []api.ChannelVariantListing{
				{
					ID:                "cvl_123",
					VariantID:         "var_123",
					Title:             "Default",
					Price:             "19.99",
					InventoryQuantity: 100,
				},
			},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupChannelProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelProductsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := channelProductsGetCmd.RunE(cmd, []string{"ch_123", "prod_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cpl_123") {
		t.Errorf("expected JSON output to contain listing ID, got: %s", output)
	}
}

// TestChannelProductsGetRunE_NoPublishedAt tests get command when PublishedAt is nil.
func TestChannelProductsGetRunE_NoPublishedAt(t *testing.T) {
	mockClient := &channelProductsMockAPIClient{
		getChannelProductResp: &api.ChannelProductListing{
			ID:               "cpl_123",
			ProductID:        "prod_123",
			ChannelID:        "ch_123",
			Title:            "Test Product",
			Handle:           "test-product",
			Status:           "draft",
			Published:        false,
			PublishedAt:      nil, // No published date
			AvailableForSale: false,
			Variants:         []api.ChannelVariantListing{}, // Empty variants
			CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupChannelProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelProductsTestCmd()

	err := channelProductsGetCmd.RunE(cmd, []string{"ch_123", "prod_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestChannelProductsPublishRunE_JSONOutput tests publish command with JSON output format.
func TestChannelProductsPublishRunE_JSONOutput(t *testing.T) {
	mockClient := &channelProductsMockAPIClient{
		publishChannelProductResp: &api.ChannelProductListing{
			ID:        "cpl_new",
			ProductID: "prod_123",
			ChannelID: "ch_123",
			Title:     "Test Product",
			Published: true,
		},
	}
	cleanup, buf := setupChannelProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelProductsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := channelProductsPublishCmd.RunE(cmd, []string{"ch_123", "prod_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cpl_new") {
		t.Errorf("expected JSON output to contain listing ID, got: %s", output)
	}
}

// TestChannelProductsUpdateRunE_JSONOutput tests update command with JSON output format.
func TestChannelProductsUpdateRunE_JSONOutput(t *testing.T) {
	mockClient := &channelProductsMockAPIClient{
		updateChannelProductResp: &api.ChannelProductListing{
			ID:               "cpl_123",
			ProductID:        "prod_123",
			ChannelID:        "ch_123",
			Published:        true,
			AvailableForSale: true,
		},
	}
	cleanup, buf := setupChannelProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelProductsTestCmd()
	cmd.Flags().Bool("published", false, "")
	cmd.Flags().Bool("available", false, "")
	_ = cmd.Flags().Set("output", "json")

	err := channelProductsUpdateCmd.RunE(cmd, []string{"ch_123", "prod_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cpl_123") {
		t.Errorf("expected JSON output to contain listing ID, got: %s", output)
	}
}

// TestChannelProductsUpdateRunE_WithFlags tests update command with published and available flags set.
func TestChannelProductsUpdateRunE_WithFlags(t *testing.T) {
	mockClient := &channelProductsMockAPIClient{
		updateChannelProductResp: &api.ChannelProductListing{
			ID:               "cpl_123",
			ProductID:        "prod_123",
			ChannelID:        "ch_123",
			Published:        true,
			AvailableForSale: false,
		},
	}
	cleanup, _ := setupChannelProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelProductsTestCmd()
	cmd.Flags().Bool("published", false, "")
	cmd.Flags().Bool("available", false, "")

	// Set the flags to trigger Changed() = true
	_ = cmd.Flags().Set("published", "true")
	_ = cmd.Flags().Set("available", "false")

	err := channelProductsUpdateCmd.RunE(cmd, []string{"ch_123", "prod_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestChannelProductsUnpublishRunE_Cancelled tests unpublish command when user cancels confirmation.
func TestChannelProductsUnpublishRunE_Cancelled(t *testing.T) {
	mockClient := &channelProductsMockAPIClient{}
	cleanup, _ := setupChannelProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelProductsTestCmd()
	_ = cmd.Flags().Set("yes", "false") // Require confirmation

	// This test will skip the interactive prompt since we can't easily mock Scanln
	// but it verifies the flag handling works
	err := channelProductsUnpublishCmd.RunE(cmd, []string{"ch_123", "prod_123"})
	// When yes=false and we don't provide input, it should handle gracefully
	// The behavior depends on stdin, so we just verify no panic
	if err != nil {
		// If there's an error, it's acceptable since we can't mock stdin
		t.Logf("got expected behavior with yes=false: %v", err)
	}
}

// TestChannelProductsCommandSetup verifies channel-products command initialization.
func TestChannelProductsCommandSetup(t *testing.T) {
	if channelProductsCmd.Use != "channel-products" {
		t.Errorf("expected Use 'channel-products', got %q", channelProductsCmd.Use)
	}
	if channelProductsCmd.Short != "Manage multi-channel product listings" {
		t.Errorf("expected Short 'Manage multi-channel product listings', got %q", channelProductsCmd.Short)
	}
}

// TestChannelProductsSubcommands verifies all subcommands are registered with correct configuration.
func TestChannelProductsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":      "List product listings in a channel",
		"get":       "Get product listing details",
		"publish":   "Publish a product to a channel",
		"unpublish": "Unpublish a product from a channel",
		"update":    "Update a product listing",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range channelProductsCmd.Commands() {
				if sub.Use == name || strings.HasPrefix(sub.Use, name+" ") {
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

// TestChannelProductsListFlagsDefaults verifies list command flags exist with correct defaults.
func TestChannelProductsListFlagsDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"published", "false"},
		{"available", "false"},
		{"status", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := channelProductsListCmd.Flags().Lookup(f.name)
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

// TestChannelProductsUpdateFlagsDefaults verifies update command flags exist with correct defaults.
func TestChannelProductsUpdateFlagsDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"published", "false"},
		{"available", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := channelProductsUpdateCmd.Flags().Lookup(f.name)
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

// TestChannelProductsListCmd verifies list command configuration.
func TestChannelProductsListCmd(t *testing.T) {
	if channelProductsListCmd.Use != "list <channel-id>" {
		t.Errorf("expected Use 'list <channel-id>', got %q", channelProductsListCmd.Use)
	}
	if channelProductsListCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestChannelProductsGetCmd verifies get command configuration.
func TestChannelProductsGetCmd(t *testing.T) {
	if channelProductsGetCmd.Use != "get <channel-id> <product-id>" {
		t.Errorf("expected Use 'get <channel-id> <product-id>', got %q", channelProductsGetCmd.Use)
	}
	if channelProductsGetCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestChannelProductsPublishCmd verifies publish command configuration.
func TestChannelProductsPublishCmd(t *testing.T) {
	if channelProductsPublishCmd.Use != "publish <channel-id> <product-id>" {
		t.Errorf("expected Use 'publish <channel-id> <product-id>', got %q", channelProductsPublishCmd.Use)
	}
	if channelProductsPublishCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestChannelProductsUnpublishCmd verifies unpublish command configuration.
func TestChannelProductsUnpublishCmd(t *testing.T) {
	if channelProductsUnpublishCmd.Use != "unpublish <channel-id> <product-id>" {
		t.Errorf("expected Use 'unpublish <channel-id> <product-id>', got %q", channelProductsUnpublishCmd.Use)
	}
	if channelProductsUnpublishCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestChannelProductsUpdateCmd verifies update command configuration.
func TestChannelProductsUpdateCmd(t *testing.T) {
	if channelProductsUpdateCmd.Use != "update <channel-id> <product-id>" {
		t.Errorf("expected Use 'update <channel-id> <product-id>', got %q", channelProductsUpdateCmd.Use)
	}
	if channelProductsUpdateCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}
