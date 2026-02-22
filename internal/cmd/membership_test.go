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

func TestMembershipCmd(t *testing.T) {
	if membershipCmd.Use != "membership" {
		t.Errorf("Expected Use to be 'membership', got %q", membershipCmd.Use)
	}
	if membershipCmd.Short != "Manage membership tiers" {
		t.Errorf("Expected Short to be 'Manage membership tiers', got %q", membershipCmd.Short)
	}
}

func TestMembershipListCmd(t *testing.T) {
	if membershipListCmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %q", membershipListCmd.Use)
	}
	if membershipListCmd.Short != "List membership tiers" {
		t.Errorf("Expected Short to be 'List membership tiers', got %q", membershipListCmd.Short)
	}
}

func TestMembershipGetCmd(t *testing.T) {
	if membershipGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use to be 'get <id>', got %q", membershipGetCmd.Use)
	}
	if membershipGetCmd.Short != "Get membership tier details" {
		t.Errorf("Expected Short to be 'Get membership tier details', got %q", membershipGetCmd.Short)
	}
}

func TestMembershipCreateCmd(t *testing.T) {
	if membershipCreateCmd.Use != "create" {
		t.Errorf("Expected Use to be 'create', got %q", membershipCreateCmd.Use)
	}
	if membershipCreateCmd.Short != "Create a membership tier" {
		t.Errorf("Expected Short to be 'Create a membership tier', got %q", membershipCreateCmd.Short)
	}
}

func TestMembershipDeleteCmd(t *testing.T) {
	if membershipDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use to be 'delete <id>', got %q", membershipDeleteCmd.Use)
	}
	if membershipDeleteCmd.Short != "Delete a membership tier" {
		t.Errorf("Expected Short to be 'Delete a membership tier', got %q", membershipDeleteCmd.Short)
	}
}

func TestMembershipListFlags(t *testing.T) {
	flags := []string{"page", "page-size"}
	for _, flag := range flags {
		if membershipListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestMembershipCreateFlags(t *testing.T) {
	flags := []string{"name", "level", "description", "min-points", "max-points", "discount"}
	for _, flag := range flags {
		if membershipCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestMembershipListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := membershipListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMembershipGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := membershipGetCmd.RunE(cmd, []string{"tier_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMembershipCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Gold", "")
	cmd.Flags().Int("level", 1, "")
	cmd.Flags().String("description", "Gold tier", "")
	cmd.Flags().Int("min-points", 1000, "")
	cmd.Flags().Int("max-points", 5000, "")
	cmd.Flags().Float64("discount", 0.1, "")
	err := membershipCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMembershipDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := membershipDeleteCmd.RunE(cmd, []string{"tier_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMembershipListRunE_NoProfiles(t *testing.T) {
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
	err := membershipListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

func TestMembershipGetRunE_MultipleProfiles(t *testing.T) {
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
	err := membershipGetCmd.RunE(cmd, []string{"tier_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles without selection, got nil")
	}
}

// membershipMockClient is a mock implementation for membership testing.
type membershipMockClient struct {
	api.MockClient
	listResp   *api.MembershipTiersListResponse
	listErr    error
	getResp    *api.MembershipTier
	getErr     error
	createResp *api.MembershipTier
	createErr  error
	deleteErr  error
}

func (m *membershipMockClient) ListMembershipTiers(ctx context.Context, opts *api.MembershipTiersListOptions) (*api.MembershipTiersListResponse, error) {
	return m.listResp, m.listErr
}

func (m *membershipMockClient) GetMembershipTier(ctx context.Context, id string) (*api.MembershipTier, error) {
	return m.getResp, m.getErr
}

func (m *membershipMockClient) CreateMembershipTier(ctx context.Context, req *api.MembershipTierCreateRequest) (*api.MembershipTier, error) {
	return m.createResp, m.createErr
}

func (m *membershipMockClient) DeleteMembershipTier(ctx context.Context, id string) error {
	return m.deleteErr
}

func setupMembershipTest(t *testing.T, mockClient *membershipMockClient) (*bytes.Buffer, func()) {
	t.Helper()
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
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}
	buf := new(bytes.Buffer)
	formatterWriter = buf

	cleanup := func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	return buf, cleanup
}

func TestMembershipListRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		mockClient  *membershipMockClient
		outputJSON  bool
		wantErr     bool
		errContain  string
		wantContain string
	}{
		{
			name: "success with tiers",
			mockClient: &membershipMockClient{
				listResp: &api.MembershipTiersListResponse{
					Items: []api.MembershipTier{
						{
							ID:        "tier_123",
							Name:      "Gold",
							Level:     1,
							MinPoints: 1000,
							MaxPoints: 5000,
							Discount:  0.1,
							CreatedAt: now,
						},
					},
					TotalCount: 1,
				},
			},
		},
		{
			name: "success with tier max points zero",
			mockClient: &membershipMockClient{
				listResp: &api.MembershipTiersListResponse{
					Items: []api.MembershipTier{
						{
							ID:        "tier_456",
							Name:      "Platinum",
							Level:     2,
							MinPoints: 5000,
							MaxPoints: 0,
							Discount:  0,
							CreatedAt: now,
						},
					},
					TotalCount: 1,
				},
			},
		},
		{
			name: "success with JSON output",
			mockClient: &membershipMockClient{
				listResp: &api.MembershipTiersListResponse{
					Items: []api.MembershipTier{
						{
							ID:        "tier_789",
							Name:      "Silver",
							Level:     0,
							MinPoints: 0,
							MaxPoints: 999,
							Discount:  0.05,
							CreatedAt: now,
						},
					},
					TotalCount: 1,
				},
			},
			outputJSON: true,
		},
		{
			name: "API error",
			mockClient: &membershipMockClient{
				listErr: errors.New("API connection failed"),
			},
			wantErr:    true,
			errContain: "failed to list membership tiers",
		},
		{
			name: "empty list",
			mockClient: &membershipMockClient{
				listResp: &api.MembershipTiersListResponse{
					Items:      []api.MembershipTier{},
					TotalCount: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupMembershipTest(t, tt.mockClient)
			defer cleanup()

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			outputVal := "text"
			if tt.outputJSON {
				outputVal = "json"
			}
			cmd.Flags().StringP("output", "o", outputVal, "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := membershipListCmd.RunE(cmd, []string{})

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

func TestMembershipGetRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *membershipMockClient
		outputJSON bool
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			mockClient: &membershipMockClient{
				getResp: &api.MembershipTier{
					ID:          "tier_123",
					Name:        "Gold",
					Level:       1,
					Description: "Gold tier with special benefits",
					MinPoints:   1000,
					MaxPoints:   5000,
					Discount:    0.1,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
		},
		{
			name: "success with max points zero",
			mockClient: &membershipMockClient{
				getResp: &api.MembershipTier{
					ID:          "tier_456",
					Name:        "Platinum",
					Level:       2,
					Description: "Highest tier",
					MinPoints:   5000,
					MaxPoints:   0,
					Discount:    0,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
		},
		{
			name: "success with JSON output",
			mockClient: &membershipMockClient{
				getResp: &api.MembershipTier{
					ID:          "tier_789",
					Name:        "Silver",
					Level:       0,
					Description: "Entry tier",
					MinPoints:   0,
					MaxPoints:   999,
					Discount:    0.05,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
			outputJSON: true,
		},
		{
			name: "not found",
			mockClient: &membershipMockClient{
				getErr: errors.New("tier not found"),
			},
			wantErr:    true,
			errContain: "failed to get membership tier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupMembershipTest(t, tt.mockClient)
			defer cleanup()

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			outputVal := "text"
			if tt.outputJSON {
				outputVal = "json"
			}
			cmd.Flags().StringP("output", "o", outputVal, "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")

			err := membershipGetCmd.RunE(cmd, []string{"tier_123"})

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

func TestMembershipCreateRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *membershipMockClient
		outputJSON bool
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			mockClient: &membershipMockClient{
				createResp: &api.MembershipTier{
					ID:          "tier_new",
					Name:        "New Tier",
					Level:       3,
					Description: "New tier description",
					MinPoints:   10000,
					MaxPoints:   20000,
					Discount:    0.15,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
		},
		{
			name: "success with JSON output",
			mockClient: &membershipMockClient{
				createResp: &api.MembershipTier{
					ID:          "tier_new_json",
					Name:        "JSON Tier",
					Level:       4,
					Description: "JSON tier description",
					MinPoints:   20000,
					MaxPoints:   50000,
					Discount:    0.2,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
			outputJSON: true,
		},
		{
			name: "API error",
			mockClient: &membershipMockClient{
				createErr: errors.New("validation failed"),
			},
			wantErr:    true,
			errContain: "failed to create membership tier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupMembershipTest(t, tt.mockClient)
			defer cleanup()

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			outputVal := "text"
			if tt.outputJSON {
				outputVal = "json"
			}
			cmd.Flags().StringP("output", "o", outputVal, "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("name", "New Tier", "")
			cmd.Flags().Int("level", 3, "")
			cmd.Flags().String("description", "New tier description", "")
			cmd.Flags().Int("min-points", 10000, "")
			cmd.Flags().Int("max-points", 20000, "")
			cmd.Flags().Float64("discount", 0.15, "")

			err := membershipCreateCmd.RunE(cmd, []string{})

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

func TestMembershipDeleteRunE_WithMockAPI(t *testing.T) {
	tests := []struct {
		name        string
		mockClient  *membershipMockClient
		skipConfirm bool
		wantErr     bool
		errContain  string
	}{
		{
			name: "success with yes flag",
			mockClient: &membershipMockClient{
				deleteErr: nil,
			},
			skipConfirm: true,
		},
		{
			name: "API error",
			mockClient: &membershipMockClient{
				deleteErr: errors.New("tier not found"),
			},
			skipConfirm: true,
			wantErr:     true,
			errContain:  "failed to delete membership tier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupMembershipTest(t, tt.mockClient)
			defer cleanup()

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().BoolP("yes", "y", tt.skipConfirm, "")

			err := membershipDeleteCmd.RunE(cmd, []string{"tier_123"})

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

func TestMembershipDeleteRunE_WithoutConfirmation(t *testing.T) {
	mockClient := &membershipMockClient{
		deleteErr: nil,
	}

	_, cleanup := setupMembershipTest(t, mockClient)
	defer cleanup()

	// Redirect stdin to simulate user declining
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write "n" to stdin to decline confirmation
	_, _ = w.WriteString("n\n")
	_ = w.Close()

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "text", "")
	cmd.Flags().String("color", "auto", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().BoolP("yes", "y", false, "")

	err := membershipDeleteCmd.RunE(cmd, []string{"tier_123"})
	// Should not return an error, just cancel
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestMembershipCreateRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("name", "Gold", "")
	cmd.Flags().Int("level", 1, "")
	cmd.Flags().String("description", "Gold tier", "")
	cmd.Flags().Int("min-points", 1000, "")
	cmd.Flags().Int("max-points", 5000, "")
	cmd.Flags().Float64("discount", 0.1, "")
	err := membershipCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

func TestMembershipDeleteRunE_NoProfiles(t *testing.T) {
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
	err := membershipDeleteCmd.RunE(cmd, []string{"tier_123"})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

func TestMembershipListRunE_MultipleProfiles(t *testing.T) {
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := membershipListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for multiple profiles without selection, got nil")
	}
}

func TestMembershipCreateRunE_MultipleProfiles(t *testing.T) {
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
	cmd.Flags().String("name", "Gold", "")
	cmd.Flags().Int("level", 1, "")
	cmd.Flags().String("description", "Gold tier", "")
	cmd.Flags().Int("min-points", 1000, "")
	cmd.Flags().Int("max-points", 5000, "")
	cmd.Flags().Float64("discount", 0.1, "")
	err := membershipCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for multiple profiles without selection, got nil")
	}
}

func TestMembershipDeleteRunE_MultipleProfiles(t *testing.T) {
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
	err := membershipDeleteCmd.RunE(cmd, []string{"tier_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles without selection, got nil")
	}
}

func TestMembershipGetRunE_NoProfiles(t *testing.T) {
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
	err := membershipGetCmd.RunE(cmd, []string{"tier_123"})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}
