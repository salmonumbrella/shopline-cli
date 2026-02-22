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

func TestCustomerGroupsCmd(t *testing.T) {
	if customerGroupsCmd.Use != "customer-groups" {
		t.Errorf("Expected Use to be 'customer-groups', got %q", customerGroupsCmd.Use)
	}
}

func TestCustomerGroupsListCmd(t *testing.T) {
	if customerGroupsListCmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %q", customerGroupsListCmd.Use)
	}
}

func TestCustomerGroupsGetCmd(t *testing.T) {
	if customerGroupsGetCmd.Use != "get [id]" {
		t.Errorf("Expected Use to be 'get [id]', got %q", customerGroupsGetCmd.Use)
	}
}

func TestCustomerGroupsCreateCmd(t *testing.T) {
	if customerGroupsCreateCmd.Use != "create" {
		t.Errorf("Expected Use to be 'create', got %q", customerGroupsCreateCmd.Use)
	}
}

func TestCustomerGroupsDeleteCmd(t *testing.T) {
	if customerGroupsDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use to be 'delete <id>', got %q", customerGroupsDeleteCmd.Use)
	}
}

func TestCustomerGroupsListFlags(t *testing.T) {
	flags := []string{"page", "page-size"}
	for _, flag := range flags {
		if customerGroupsListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestCustomerGroupsCreateFlags(t *testing.T) {
	flags := []string{"name", "description"}
	for _, flag := range flags {
		if customerGroupsCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestCustomerGroupsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := customerGroupsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerGroupsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := customerGroupsGetCmd.RunE(cmd, []string{"group_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerGroupsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test Group", "")
	cmd.Flags().String("description", "Test description", "")
	err := customerGroupsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerGroupsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := customerGroupsDeleteCmd.RunE(cmd, []string{"group_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerGroupsListRunE_NoProfiles(t *testing.T) {
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
	err := customerGroupsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// groupsMockClient is a mock implementation for customer groups testing.
type groupsMockClient struct {
	api.MockClient
	listResp   *api.CustomerGroupsListResponse
	listErr    error
	searchResp *api.CustomerGroupsListResponse
	searchErr  error
	getResp    *api.CustomerGroup
	getErr     error
	createResp *api.CustomerGroup
	createErr  error
	deleteErr  error
}

func (m *groupsMockClient) ListCustomerGroups(ctx context.Context, opts *api.CustomerGroupsListOptions) (*api.CustomerGroupsListResponse, error) {
	return m.listResp, m.listErr
}

func (m *groupsMockClient) SearchCustomerGroups(ctx context.Context, opts *api.CustomerGroupSearchOptions) (*api.CustomerGroupsListResponse, error) {
	return m.searchResp, m.searchErr
}

func (m *groupsMockClient) GetCustomerGroup(ctx context.Context, id string) (*api.CustomerGroup, error) {
	return m.getResp, m.getErr
}

func (m *groupsMockClient) CreateCustomerGroup(ctx context.Context, req *api.CustomerGroupCreateRequest) (*api.CustomerGroup, error) {
	return m.createResp, m.createErr
}

func (m *groupsMockClient) DeleteCustomerGroup(ctx context.Context, id string) error {
	return m.deleteErr
}

func TestCustomerGroupsListRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *groupsMockClient
		wantErr    bool
		errContain string
	}{
		{
			name: "success with groups",
			mockClient: &groupsMockClient{
				listResp: &api.CustomerGroupsListResponse{
					Items: []api.CustomerGroup{
						{
							ID:            "grp_123",
							Name:          "VIP Customers",
							Description:   "VIP customers with special privileges",
							CustomerCount: 50,
							CreatedAt:     now,
						},
					},
					TotalCount: 1,
				},
			},
		},
		{
			name: "API error",
			mockClient: &groupsMockClient{
				listErr: errors.New("API connection failed"),
			},
			wantErr:    true,
			errContain: "failed to list customer groups",
		},
		{
			name: "empty list",
			mockClient: &groupsMockClient{
				listResp: &api.CustomerGroupsListResponse{
					Items:      []api.CustomerGroup{},
					TotalCount: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := customerGroupsListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCustomerGroupsGetRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *groupsMockClient
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			mockClient: &groupsMockClient{
				getResp: &api.CustomerGroup{
					ID:            "grp_123",
					Name:          "VIP Customers",
					Description:   "VIP customers",
					CustomerCount: 50,
					CreatedAt:     now,
					UpdatedAt:     now,
				},
			},
		},
		{
			name: "not found",
			mockClient: &groupsMockClient{
				getErr: errors.New("group not found"),
			},
			wantErr:    true,
			errContain: "failed to get customer group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")

			err := customerGroupsGetCmd.RunE(cmd, []string{"grp_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCustomerGroupsCreateRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *groupsMockClient
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			mockClient: &groupsMockClient{
				createResp: &api.CustomerGroup{
					ID:          "grp_new",
					Name:        "New Group",
					Description: "Test group",
					CreatedAt:   now,
				},
			},
		},
		{
			name: "API error",
			mockClient: &groupsMockClient{
				createErr: errors.New("validation failed"),
			},
			wantErr:    true,
			errContain: "failed to create customer group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("name", "New Group", "")
			cmd.Flags().String("description", "Test group", "")

			err := customerGroupsCreateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

// setupGroupsMockFactories sets up mock factories for customer groups tests.
func setupGroupsMockFactories(mockClient *groupsMockClient) (func(), *bytes.Buffer) {
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

// newGroupsTestCmd creates a test command with common flags for customer groups tests.
func newGroupsTestCmd() *cobra.Command {
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

func TestCustomerGroupsDeleteRunE_WithMockAPI(t *testing.T) {
	tests := []struct {
		name       string
		mockClient *groupsMockClient
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			mockClient: &groupsMockClient{
				deleteErr: nil,
			},
		},
		{
			name: "API error",
			mockClient: &groupsMockClient{
				deleteErr: errors.New("group not found"),
			},
			wantErr:    true,
			errContain: "failed to delete customer group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().BoolP("yes", "y", true, "") // Skip confirmation

			err := customerGroupsDeleteCmd.RunE(cmd, []string{"grp_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

// TestCustomerGroupsGetByFlag tests the --by flag on the customer-groups get command.
func TestCustomerGroupsGetByFlag(t *testing.T) {
	now := time.Now()

	t.Run("resolves customer group by name", func(t *testing.T) {
		mockClient := &groupsMockClient{
			searchResp: &api.CustomerGroupsListResponse{
				Items:      []api.CustomerGroup{{ID: "grp_found", Name: "VIP Customers"}},
				TotalCount: 1,
			},
			getResp: &api.CustomerGroup{
				ID:        "grp_found",
				Name:      "VIP Customers",
				CreatedAt: now,
			},
		}
		cleanup, buf := setupGroupsMockFactories(mockClient)
		defer cleanup()

		cmd := newGroupsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "VIP Customers")

		if err := customerGroupsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "grp_found") {
			t.Errorf("expected output to contain 'grp_found', got: %s", buf.String())
		}
	})

	t.Run("errors when no match", func(t *testing.T) {
		mockClient := &groupsMockClient{
			searchResp: &api.CustomerGroupsListResponse{
				Items:      []api.CustomerGroup{},
				TotalCount: 0,
			},
		}
		cleanup, _ := setupGroupsMockFactories(mockClient)
		defer cleanup()

		cmd := newGroupsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "nonexistent")

		err := customerGroupsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when no customer group found")
		}
		if !strings.Contains(err.Error(), "no customer group found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("errors when search fails", func(t *testing.T) {
		mockClient := &groupsMockClient{
			searchErr: errors.New("API error"),
		}
		cleanup, _ := setupGroupsMockFactories(mockClient)
		defer cleanup()

		cmd := newGroupsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "VIP Customers")

		err := customerGroupsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when search fails")
		}
		if !strings.Contains(err.Error(), "search failed") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("warns on multiple matches", func(t *testing.T) {
		mockClient := &groupsMockClient{
			searchResp: &api.CustomerGroupsListResponse{
				Items: []api.CustomerGroup{
					{ID: "grp_1", Name: "VIP A"},
					{ID: "grp_2", Name: "VIP B"},
				},
				TotalCount: 2,
			},
			getResp: &api.CustomerGroup{
				ID:        "grp_1",
				Name:      "VIP A",
				CreatedAt: now,
			},
		}
		cleanup, buf := setupGroupsMockFactories(mockClient)
		defer cleanup()

		cmd := newGroupsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "VIP")

		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)

		if err := customerGroupsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "grp_1") {
			t.Errorf("expected output to contain 'grp_1', got: %s", buf.String())
		}
		if !strings.Contains(stderr.String(), "2 matches found") {
			t.Errorf("expected stderr warning about multiple matches, got: %s", stderr.String())
		}
	})

	t.Run("positional arg takes precedence over --by", func(t *testing.T) {
		mockClient := &groupsMockClient{
			getResp: &api.CustomerGroup{
				ID:        "grp_direct",
				Name:      "Direct Group",
				CreatedAt: now,
			},
		}
		cleanup, buf := setupGroupsMockFactories(mockClient)
		defer cleanup()

		cmd := newGroupsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "should-not-be-used")

		if err := customerGroupsGetCmd.RunE(cmd, []string{"grp_direct"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "grp_direct") {
			t.Errorf("expected output to contain 'grp_direct', got: %s", buf.String())
		}
	})

	t.Run("errors with no arg and no --by", func(t *testing.T) {
		mockClient := &groupsMockClient{}
		cleanup, _ := setupGroupsMockFactories(mockClient)
		defer cleanup()

		cmd := newGroupsTestCmd()
		cmd.Flags().String("by", "", "")

		err := customerGroupsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error with no arg and no --by")
		}
		if !strings.Contains(err.Error(), "provide a resource ID") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
