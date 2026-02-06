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

// TestChannelsCommandSetup verifies channels command initialization
func TestChannelsCommandSetup(t *testing.T) {
	if channelsCmd.Use != "channels" {
		t.Errorf("expected Use 'channels', got %q", channelsCmd.Use)
	}
	if channelsCmd.Short != "Manage sales channels" {
		t.Errorf("expected Short 'Manage sales channels', got %q", channelsCmd.Short)
	}
}

// TestChannelsSubcommands verifies all subcommands are registered
func TestChannelsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":      "List sales channels",
		"get":       "Get channel details",
		"create":    "Create a sales channel",
		"delete":    "Delete a sales channel",
		"products":  "List products in a channel",
		"prices":    "Manage product channel prices",
		"publish":   "Publish a product to a channel",
		"unpublish": "Unpublish a product from a channel",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range channelsCmd.Commands() {
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

// TestChannelsListFlags verifies list command flags exist with correct defaults
func TestChannelsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"active", "false"},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := channelsListCmd.Flags().Lookup(f.name)
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

// TestChannelsCreateFlags verifies create command flags exist with correct defaults
func TestChannelsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"handle", ""},
		{"type", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := channelsCreateCmd.Flags().Lookup(f.name)
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

// TestChannelsCreateRequiredFlags verifies name and type are required
func TestChannelsCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"name", "type"}

	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := channelsCreateCmd.Flags().Lookup(name)
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

// TestChannelsProductsFlags verifies products subcommand flags exist with correct defaults
func TestChannelsProductsFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := channelsProductsCmd.Flags().Lookup(f.name)
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

// TestChannelsGetArgs verifies get command requires exactly 1 argument
func TestChannelsGetArgs(t *testing.T) {
	err := channelsGetCmd.Args(channelsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = channelsGetCmd.Args(channelsGetCmd, []string{"ch-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestChannelsDeleteArgs verifies delete command requires exactly 1 argument
func TestChannelsDeleteArgs(t *testing.T) {
	err := channelsDeleteCmd.Args(channelsDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = channelsDeleteCmd.Args(channelsDeleteCmd, []string{"ch-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestChannelsProductsArgs verifies products command requires exactly 1 argument
func TestChannelsProductsArgs(t *testing.T) {
	err := channelsProductsCmd.Args(channelsProductsCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = channelsProductsCmd.Args(channelsProductsCmd, []string{"ch-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestChannelsPublishArgs verifies publish command requires exactly 2 arguments
func TestChannelsPublishArgs(t *testing.T) {
	err := channelsPublishCmd.Args(channelsPublishCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = channelsPublishCmd.Args(channelsPublishCmd, []string{"ch-id"})
	if err == nil {
		t.Error("expected error when only 1 arg provided")
	}

	err = channelsPublishCmd.Args(channelsPublishCmd, []string{"ch-id", "prod-id"})
	if err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}
}

// TestChannelsUnpublishArgs verifies unpublish command requires exactly 2 arguments
func TestChannelsUnpublishArgs(t *testing.T) {
	err := channelsUnpublishCmd.Args(channelsUnpublishCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = channelsUnpublishCmd.Args(channelsUnpublishCmd, []string{"ch-id", "prod-id"})
	if err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}
}

// TestChannelsGetClientError verifies error handling when getClient fails
func TestChannelsGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_, err := getClient(cmd)
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestChannelsListGetClientError verifies list command error handling when getClient fails
func TestChannelsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(channelsListCmd)

	err := channelsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestChannelsGetGetClientError verifies get command error handling when getClient fails
func TestChannelsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(channelsGetCmd)

	err := channelsGetCmd.RunE(cmd, []string{"ch-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestChannelsCreateGetClientError verifies create command error handling when getClient fails
func TestChannelsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(channelsCreateCmd)

	err := channelsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestChannelsDeleteGetClientError verifies delete command error handling when getClient fails
func TestChannelsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(channelsDeleteCmd)

	err := channelsDeleteCmd.RunE(cmd, []string{"ch-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestChannelsProductsGetClientError verifies products command error handling when getClient fails
func TestChannelsProductsGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(channelsProductsCmd)

	err := channelsProductsCmd.RunE(cmd, []string{"ch-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestChannelsPublishGetClientError verifies publish command error handling when getClient fails
func TestChannelsPublishGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(channelsPublishCmd)

	err := channelsPublishCmd.RunE(cmd, []string{"ch-id", "prod-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestChannelsUnpublishGetClientError verifies unpublish command error handling when getClient fails
func TestChannelsUnpublishGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(channelsUnpublishCmd)

	err := channelsUnpublishCmd.RunE(cmd, []string{"ch-id", "prod-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestChannelsWithMockStore tests channels commands with a mock credential store
func TestChannelsWithMockStore(t *testing.T) {
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

// channelsMockAPIClient is a mock implementation of api.APIClient for channels tests.
type channelsMockAPIClient struct {
	api.MockClient
	listChannelsResp        *api.ChannelsListResponse
	listChannelsErr         error
	getChannelResp          *api.Channel
	getChannelErr           error
	createChannelResp       *api.Channel
	createChannelErr        error
	deleteChannelErr        error
	listChannelProductsResp *api.ChannelProductsResponse
	listChannelProductsErr  error
	publishProductErr       error
	unpublishProductErr     error
}

func (m *channelsMockAPIClient) ListChannels(ctx context.Context, opts *api.ChannelsListOptions) (*api.ChannelsListResponse, error) {
	return m.listChannelsResp, m.listChannelsErr
}

func (m *channelsMockAPIClient) GetChannel(ctx context.Context, id string) (*api.Channel, error) {
	return m.getChannelResp, m.getChannelErr
}

func (m *channelsMockAPIClient) CreateChannel(ctx context.Context, req *api.ChannelCreateRequest) (*api.Channel, error) {
	return m.createChannelResp, m.createChannelErr
}

func (m *channelsMockAPIClient) DeleteChannel(ctx context.Context, id string) error {
	return m.deleteChannelErr
}

func (m *channelsMockAPIClient) ListChannelProducts(ctx context.Context, channelID string, page, pageSize int) (*api.ChannelProductsResponse, error) {
	return m.listChannelProductsResp, m.listChannelProductsErr
}

func (m *channelsMockAPIClient) PublishProductToChannel(ctx context.Context, channelID string, req *api.ChannelPublishProductRequest) error {
	return m.publishProductErr
}

func (m *channelsMockAPIClient) UnpublishProductFromChannel(ctx context.Context, channelID, productID string) error {
	return m.unpublishProductErr
}

// setupChannelsMockFactories sets up mock factories for channels tests.
func setupChannelsMockFactories(mockClient *channelsMockAPIClient) (func(), *bytes.Buffer) {
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

// newChannelsTestCmd creates a test command with common flags for channels tests.
func newChannelsTestCmd() *cobra.Command {
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

// TestChannelsListRunE tests the channels list command with mock API.
func TestChannelsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ChannelsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ChannelsListResponse{
				Items: []api.Channel{
					{
						ID:           "ch_123",
						Name:         "Online Store",
						Handle:       "online-store",
						Type:         "online_store",
						Active:       true,
						ProductCount: 100,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "ch_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ChannelsListResponse{
				Items:      []api.Channel{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple channels",
			mockResp: &api.ChannelsListResponse{
				Items: []api.Channel{
					{
						ID:           "ch_123",
						Name:         "Online Store",
						Handle:       "online-store",
						Type:         "online_store",
						Active:       true,
						ProductCount: 100,
					},
					{
						ID:           "ch_456",
						Name:         "Point of Sale",
						Handle:       "pos",
						Type:         "point_of_sale",
						Active:       false,
						ProductCount: 50,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "ch_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelsMockAPIClient{
				listChannelsResp: tt.mockResp,
				listChannelsErr:  tt.mockErr,
			}
			cleanup, buf := setupChannelsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelsTestCmd()
			cmd.Flags().Bool("active", false, "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := channelsListCmd.RunE(cmd, []string{})

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

// TestChannelsListRunEJSON tests the channels list command with JSON output.
func TestChannelsListRunEJSON(t *testing.T) {
	mockClient := &channelsMockAPIClient{
		listChannelsResp: &api.ChannelsListResponse{
			Items: []api.Channel{
				{
					ID:     "ch_123",
					Name:   "Online Store",
					Handle: "online-store",
					Type:   "online_store",
					Active: true,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupChannelsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelsTestCmd()
	cmd.Flags().Bool("active", false, "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := channelsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "ch_123") {
		t.Errorf("JSON output should contain channel ID, got %q", output)
	}
}

// TestChannelsListActiveFilter tests the list command with active filter.
func TestChannelsListActiveFilter(t *testing.T) {
	mockClient := &channelsMockAPIClient{
		listChannelsResp: &api.ChannelsListResponse{
			Items: []api.Channel{
				{
					ID:     "ch_123",
					Name:   "Active Channel",
					Active: true,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, _ := setupChannelsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelsTestCmd()
	cmd.Flags().Bool("active", false, "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("active", "true")

	err := channelsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestChannelsGetRunE tests the channels get command with mock API.
func TestChannelsGetRunE(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		mockResp  *api.Channel
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			channelID: "ch_123",
			mockResp: &api.Channel{
				ID:                        "ch_123",
				Name:                      "Online Store",
				Handle:                    "online-store",
				Type:                      "online_store",
				Active:                    true,
				ProductCount:              100,
				CollectionCount:           10,
				SupportsRemoteFulfillment: true,
				CreatedAt:                 time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:                 time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
			},
		},
		{
			name:      "channel not found",
			channelID: "ch_999",
			mockErr:   errors.New("channel not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelsMockAPIClient{
				getChannelResp: tt.mockResp,
				getChannelErr:  tt.mockErr,
			}
			cleanup, _ := setupChannelsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelsTestCmd()

			err := channelsGetCmd.RunE(cmd, []string{tt.channelID})

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

// TestChannelsGetRunEJSON tests the channels get command with JSON output.
func TestChannelsGetRunEJSON(t *testing.T) {
	mockClient := &channelsMockAPIClient{
		getChannelResp: &api.Channel{
			ID:           "ch_123",
			Name:         "Online Store",
			Handle:       "online-store",
			Type:         "online_store",
			Active:       true,
			ProductCount: 100,
			CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupChannelsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := channelsGetCmd.RunE(cmd, []string{"ch_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "ch_123") {
		t.Errorf("JSON output should contain channel ID, got %q", output)
	}
}

// TestChannelsCreateRunE tests the channels create command with mock API.
func TestChannelsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.Channel
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Channel{
				ID:     "ch_new",
				Name:   "New Channel",
				Handle: "new-channel",
				Type:   "online_store",
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("failed to create channel"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelsMockAPIClient{
				createChannelResp: tt.mockResp,
				createChannelErr:  tt.mockErr,
			}
			cleanup, _ := setupChannelsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelsTestCmd()
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("handle", "", "")
			cmd.Flags().String("type", "", "")
			_ = cmd.Flags().Set("name", "New Channel")
			_ = cmd.Flags().Set("type", "online_store")

			err := channelsCreateCmd.RunE(cmd, []string{})

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

// TestChannelsCreateRunEJSON tests the channels create command with JSON output.
func TestChannelsCreateRunEJSON(t *testing.T) {
	mockClient := &channelsMockAPIClient{
		createChannelResp: &api.Channel{
			ID:     "ch_new",
			Name:   "New Channel",
			Handle: "new-channel",
			Type:   "online_store",
		},
	}
	cleanup, buf := setupChannelsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelsTestCmd()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("handle", "", "")
	cmd.Flags().String("type", "", "")
	_ = cmd.Flags().Set("name", "New Channel")
	_ = cmd.Flags().Set("type", "online_store")
	_ = cmd.Flags().Set("output", "json")

	err := channelsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "ch_new") {
		t.Errorf("JSON output should contain channel ID, got %q", output)
	}
}

// TestChannelsDeleteRunE tests the channels delete command with mock API.
func TestChannelsDeleteRunE(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful delete",
			channelID: "ch_123",
			mockErr:   nil,
		},
		{
			name:      "delete error",
			channelID: "ch_999",
			mockErr:   errors.New("channel not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelsMockAPIClient{
				deleteChannelErr: tt.mockErr,
			}
			cleanup, _ := setupChannelsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelsTestCmd()

			err := channelsDeleteCmd.RunE(cmd, []string{tt.channelID})

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

// TestChannelsProductsRunE tests the channels products command with mock API.
func TestChannelsProductsRunE(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		mockResp   *api.ChannelProductsResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:      "successful list products",
			channelID: "ch_123",
			mockResp: &api.ChannelProductsResponse{
				Items: []api.ChannelProduct{
					{
						ProductID: "prod_123",
						Published: true,
					},
					{
						ProductID: "prod_456",
						Published: false,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "prod_123",
		},
		{
			name:      "API error",
			channelID: "ch_123",
			mockErr:   errors.New("API error"),
			wantErr:   true,
		},
		{
			name:      "empty products list",
			channelID: "ch_123",
			mockResp: &api.ChannelProductsResponse{
				Items:      []api.ChannelProduct{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelsMockAPIClient{
				listChannelProductsResp: tt.mockResp,
				listChannelProductsErr:  tt.mockErr,
			}
			cleanup, buf := setupChannelsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := channelsProductsCmd.RunE(cmd, []string{tt.channelID})

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

// TestChannelsProductsRunEJSON tests the channels products command with JSON output.
func TestChannelsProductsRunEJSON(t *testing.T) {
	mockClient := &channelsMockAPIClient{
		listChannelProductsResp: &api.ChannelProductsResponse{
			Items: []api.ChannelProduct{
				{
					ProductID: "prod_123",
					Published: true,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupChannelsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := channelsProductsCmd.RunE(cmd, []string{"ch_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "prod_123") {
		t.Errorf("JSON output should contain product ID, got %q", output)
	}
}

// TestChannelsPublishRunE tests the channels publish command with mock API.
func TestChannelsPublishRunE(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		productID string
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful publish",
			channelID: "ch_123",
			productID: "prod_456",
			mockErr:   nil,
		},
		{
			name:      "publish error",
			channelID: "ch_123",
			productID: "prod_999",
			mockErr:   errors.New("product not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelsMockAPIClient{
				publishProductErr: tt.mockErr,
			}
			cleanup, _ := setupChannelsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelsTestCmd()

			err := channelsPublishCmd.RunE(cmd, []string{tt.channelID, tt.productID})

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

// TestChannelsUnpublishRunE tests the channels unpublish command with mock API.
func TestChannelsUnpublishRunE(t *testing.T) {
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
			productID: "prod_456",
			mockErr:   nil,
		},
		{
			name:      "unpublish error",
			channelID: "ch_123",
			productID: "prod_999",
			mockErr:   errors.New("product not in channel"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &channelsMockAPIClient{
				unpublishProductErr: tt.mockErr,
			}
			cleanup, _ := setupChannelsMockFactories(mockClient)
			defer cleanup()

			cmd := newChannelsTestCmd()

			err := channelsUnpublishCmd.RunE(cmd, []string{tt.channelID, tt.productID})

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

// TestChannelsDeleteYesFlag tests delete command with --yes flag skipping confirmation
func TestChannelsDeleteYesFlag(t *testing.T) {
	mockClient := &channelsMockAPIClient{
		deleteChannelErr: nil,
	}
	cleanup, _ := setupChannelsMockFactories(mockClient)
	defer cleanup()

	cmd := newChannelsTestCmd()
	// yes flag is already set to true in newChannelsTestCmd

	err := channelsDeleteCmd.RunE(cmd, []string{"ch_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
