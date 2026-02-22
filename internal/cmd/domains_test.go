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

// domainsMockAPIClient is a mock implementation of api.APIClient for domains tests.
type domainsMockAPIClient struct {
	api.MockClient
	listDomainsResp  *api.DomainsListResponse
	listDomainsErr   error
	getDomainResp    *api.Domain
	getDomainErr     error
	createDomainResp *api.Domain
	createDomainErr  error
	updateDomainResp *api.Domain
	updateDomainErr  error
	deleteDomainErr  error
	verifyDomainResp *api.Domain
	verifyDomainErr  error
}

func (m *domainsMockAPIClient) ListDomains(ctx context.Context, opts *api.DomainsListOptions) (*api.DomainsListResponse, error) {
	return m.listDomainsResp, m.listDomainsErr
}

func (m *domainsMockAPIClient) GetDomain(ctx context.Context, id string) (*api.Domain, error) {
	return m.getDomainResp, m.getDomainErr
}

func (m *domainsMockAPIClient) CreateDomain(ctx context.Context, req *api.DomainCreateRequest) (*api.Domain, error) {
	return m.createDomainResp, m.createDomainErr
}

func (m *domainsMockAPIClient) UpdateDomain(ctx context.Context, id string, req *api.DomainUpdateRequest) (*api.Domain, error) {
	return m.updateDomainResp, m.updateDomainErr
}

func (m *domainsMockAPIClient) DeleteDomain(ctx context.Context, id string) error {
	return m.deleteDomainErr
}

func (m *domainsMockAPIClient) VerifyDomain(ctx context.Context, id string) (*api.Domain, error) {
	return m.verifyDomainResp, m.verifyDomainErr
}

// setupDomainsMockFactories sets up mock factories for domains tests.
func setupDomainsMockFactories(mockClient *domainsMockAPIClient) (func(), *bytes.Buffer) {
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

// newDomainsTestCmd creates a test command with common flags for domains tests.
func newDomainsTestCmd() *cobra.Command {
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

func TestDomainsCommandStructure(t *testing.T) {
	subcommands := domainsCmd.Commands()

	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
		"delete": false,
		"verify": false,
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

func TestDomainsAliases(t *testing.T) {
	aliases := domainsCmd.Aliases
	found := false
	for _, alias := range aliases {
		if alias == "domain" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected alias 'domain' not found")
	}
}

func TestDomainsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
		{"primary", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := domainsListCmd.Flags().Lookup(f.name)
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

func TestDomainsGetArgs(t *testing.T) {
	err := domainsGetCmd.Args(domainsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = domainsGetCmd.Args(domainsGetCmd, []string{"domain-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDomainsUpdateArgs(t *testing.T) {
	err := domainsUpdateCmd.Args(domainsUpdateCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = domainsUpdateCmd.Args(domainsUpdateCmd, []string{"domain-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDomainsDeleteArgs(t *testing.T) {
	err := domainsDeleteCmd.Args(domainsDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = domainsDeleteCmd.Args(domainsDeleteCmd, []string{"domain-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDomainsVerifyArgs(t *testing.T) {
	err := domainsVerifyCmd.Args(domainsVerifyCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = domainsVerifyCmd.Args(domainsVerifyCmd, []string{"domain-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDomainsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"host", ""},
		{"primary", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := domainsCreateCmd.Flags().Lookup(f.name)
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

func TestDomainsUpdateFlags(t *testing.T) {
	flags := []string{"primary"}

	for _, flagName := range flags {
		if domainsUpdateCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on update command", flagName)
		}
	}
}

// TestDomainsListRunE tests the domains list command with mock API.
func TestDomainsListRunE(t *testing.T) {
	verifiedAt := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name         string
		mockResp     *api.DomainsListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name: "successful list text output",
			mockResp: &api.DomainsListResponse{
				Items: []api.Domain{
					{
						ID:        "dom_123",
						Host:      "example.com",
						Primary:   true,
						SSL:       true,
						Status:    api.DomainStatusActive,
						Verified:  true,
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "dom_123",
		},
		{
			name: "successful list with non-primary domain",
			mockResp: &api.DomainsListResponse{
				Items: []api.Domain{
					{
						ID:        "dom_456",
						Host:      "shop.example.com",
						Primary:   false,
						SSL:       false,
						Status:    api.DomainStatusPending,
						Verified:  false,
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "dom_456",
		},
		{
			name: "successful list JSON output",
			mockResp: &api.DomainsListResponse{
				Items: []api.Domain{
					{
						ID:         "dom_789",
						Host:       "store.example.com",
						Primary:    true,
						SSL:        true,
						Status:     api.DomainStatusActive,
						Verified:   true,
						VerifiedAt: &verifiedAt,
						CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFormat: "json",
			wantOutput:   "dom_789",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.DomainsListResponse{
				Items:      []api.Domain{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &domainsMockAPIClient{
				listDomainsResp: tt.mockResp,
				listDomainsErr:  tt.mockErr,
			}
			cleanup, buf := setupDomainsMockFactories(mockClient)
			defer cleanup()

			cmd := newDomainsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Bool("primary", false, "")
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := domainsListCmd.RunE(cmd, []string{})

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

// TestDomainsGetRunE tests the domains get command with mock API.
func TestDomainsGetRunE(t *testing.T) {
	verifiedAt := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		domainID     string
		mockResp     *api.Domain
		mockErr      error
		outputFormat string
		wantErr      bool
	}{
		{
			name:     "successful get verified domain",
			domainID: "dom_123",
			mockResp: &api.Domain{
				ID:         "dom_123",
				Host:       "example.com",
				Primary:    true,
				SSL:        true,
				SSLStatus:  "active",
				Status:     api.DomainStatusActive,
				Verified:   true,
				VerifiedAt: &verifiedAt,
				ExpiresAt:  &expiresAt,
				CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:     "successful get unverified domain with verification info",
			domainID: "dom_456",
			mockResp: &api.Domain{
				ID:                "dom_456",
				Host:              "pending.example.com",
				Primary:           false,
				SSL:               false,
				Status:            api.DomainStatusPending,
				Verified:          false,
				VerificationDNS:   "TXT _verify.pending.example.com",
				VerificationToken: "verify-token-123",
				CreatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:     "successful get JSON output",
			domainID: "dom_789",
			mockResp: &api.Domain{
				ID:        "dom_789",
				Host:      "json.example.com",
				Primary:   true,
				SSL:       true,
				Status:    api.DomainStatusActive,
				Verified:  true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			outputFormat: "json",
		},
		{
			name:     "domain not found",
			domainID: "dom_999",
			mockErr:  errors.New("domain not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &domainsMockAPIClient{
				getDomainResp: tt.mockResp,
				getDomainErr:  tt.mockErr,
			}
			cleanup, _ := setupDomainsMockFactories(mockClient)
			defer cleanup()

			cmd := newDomainsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := domainsGetCmd.RunE(cmd, []string{tt.domainID})

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

// TestDomainsCreateRunE tests the domains create command with mock API.
func TestDomainsCreateRunE(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		primary      bool
		dryRun       bool
		mockResp     *api.Domain
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name:   "successful create",
			host:   "newdomain.com",
			dryRun: false,
			mockResp: &api.Domain{
				ID:                "dom_new",
				Host:              "newdomain.com",
				Primary:           false,
				Status:            api.DomainStatusPending,
				VerificationDNS:   "TXT _verify.newdomain.com",
				VerificationToken: "verify-abc123",
				CreatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			// Text output goes to stdout, not buffer
		},
		{
			name:    "successful create as primary",
			host:    "primary.com",
			primary: true,
			dryRun:  false,
			mockResp: &api.Domain{
				ID:        "dom_primary",
				Host:      "primary.com",
				Primary:   true,
				Status:    api.DomainStatusPending,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			// Text output goes to stdout, not buffer
		},
		{
			name:   "successful create JSON output",
			host:   "json.domain.com",
			dryRun: false,
			mockResp: &api.Domain{
				ID:        "dom_json",
				Host:      "json.domain.com",
				Status:    api.DomainStatusPending,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			outputFormat: "json",
			wantOutput:   "dom_json",
		},
		{
			name:   "dry-run mode",
			host:   "dryrun.com",
			dryRun: true,
		},
		{
			name:    "API error",
			host:    "error.com",
			dryRun:  false,
			mockErr: errors.New("domain already exists"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &domainsMockAPIClient{
				createDomainResp: tt.mockResp,
				createDomainErr:  tt.mockErr,
			}
			cleanup, buf := setupDomainsMockFactories(mockClient)
			defer cleanup()

			cmd := newDomainsTestCmd()
			cmd.Flags().String("host", "", "")
			cmd.Flags().Bool("primary", false, "")
			_ = cmd.Flags().Set("host", tt.host)
			if tt.primary {
				_ = cmd.Flags().Set("primary", "true")
			}
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := domainsCreateCmd.RunE(cmd, []string{})

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

// TestDomainsUpdateRunE tests the domains update command with mock API.
func TestDomainsUpdateRunE(t *testing.T) {
	tests := []struct {
		name         string
		domainID     string
		setPrimary   bool
		dryRun       bool
		mockResp     *api.Domain
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name:       "successful update to primary",
			domainID:   "dom_123",
			setPrimary: true,
			dryRun:     false,
			mockResp: &api.Domain{
				ID:        "dom_123",
				Host:      "example.com",
				Primary:   true,
				Status:    api.DomainStatusActive,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			// Text output goes to stdout, not buffer
		},
		{
			name:       "successful update JSON output",
			domainID:   "dom_456",
			setPrimary: true,
			dryRun:     false,
			mockResp: &api.Domain{
				ID:        "dom_456",
				Host:      "json.example.com",
				Primary:   true,
				Status:    api.DomainStatusActive,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			outputFormat: "json",
			wantOutput:   "dom_456",
		},
		{
			name:     "dry-run mode",
			domainID: "dom_789",
			dryRun:   true,
		},
		{
			name:     "API error",
			domainID: "dom_999",
			dryRun:   false,
			mockErr:  errors.New("domain not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &domainsMockAPIClient{
				updateDomainResp: tt.mockResp,
				updateDomainErr:  tt.mockErr,
			}
			cleanup, buf := setupDomainsMockFactories(mockClient)
			defer cleanup()

			cmd := newDomainsTestCmd()
			cmd.Flags().Bool("primary", false, "")
			if tt.setPrimary {
				_ = cmd.Flags().Set("primary", "true")
			}
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := domainsUpdateCmd.RunE(cmd, []string{tt.domainID})

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

// TestDomainsDeleteRunE tests the domains delete command with mock API.
func TestDomainsDeleteRunE(t *testing.T) {
	tests := []struct {
		name     string
		domainID string
		confirm  bool
		dryRun   bool
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful delete with confirmation",
			domainID: "dom_123",
			confirm:  true,
			dryRun:   false,
		},
		{
			name:     "delete without confirmation (exits early)",
			domainID: "dom_456",
			confirm:  false,
			dryRun:   false,
		},
		{
			name:     "dry-run mode",
			domainID: "dom_789",
			confirm:  false,
			dryRun:   true,
		},
		{
			name:     "API error",
			domainID: "dom_999",
			confirm:  true,
			dryRun:   false,
			mockErr:  errors.New("domain not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &domainsMockAPIClient{
				deleteDomainErr: tt.mockErr,
			}
			cleanup, _ := setupDomainsMockFactories(mockClient)
			defer cleanup()

			cmd := newDomainsTestCmd()
			if tt.confirm {
				_ = cmd.Flags().Set("yes", "true")
			} else {
				_ = cmd.Flags().Set("yes", "false")
			}
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := domainsDeleteCmd.RunE(cmd, []string{tt.domainID})

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

// TestDomainsVerifyRunE tests the domains verify command with mock API.
func TestDomainsVerifyRunE(t *testing.T) {
	verifiedAt := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		domainID     string
		dryRun       bool
		mockResp     *api.Domain
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name:     "successful verify",
			domainID: "dom_123",
			dryRun:   false,
			mockResp: &api.Domain{
				ID:         "dom_123",
				Host:       "example.com",
				Status:     api.DomainStatusActive,
				Verified:   true,
				VerifiedAt: &verifiedAt,
				CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			// Text output goes to stdout, not buffer
		},
		{
			name:     "verify pending domain",
			domainID: "dom_456",
			dryRun:   false,
			mockResp: &api.Domain{
				ID:        "dom_456",
				Host:      "pending.example.com",
				Status:    api.DomainStatusVerifying,
				Verified:  false,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			// Text output goes to stdout, not buffer
		},
		{
			name:     "verify JSON output",
			domainID: "dom_789",
			dryRun:   false,
			mockResp: &api.Domain{
				ID:        "dom_789",
				Host:      "json.example.com",
				Status:    api.DomainStatusActive,
				Verified:  true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			outputFormat: "json",
			wantOutput:   "dom_789",
		},
		{
			name:     "dry-run mode",
			domainID: "dom_dry",
			dryRun:   true,
		},
		{
			name:     "API error",
			domainID: "dom_999",
			dryRun:   false,
			mockErr:  errors.New("verification failed"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &domainsMockAPIClient{
				verifyDomainResp: tt.mockResp,
				verifyDomainErr:  tt.mockErr,
			}
			cleanup, buf := setupDomainsMockFactories(mockClient)
			defer cleanup()

			cmd := newDomainsTestCmd()
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := domainsVerifyCmd.RunE(cmd, []string{tt.domainID})

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

// TestDomainsListGetClientError verifies error handling when getClient fails.
func TestDomainsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Bool("primary", false, "")

	err := domainsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestDomainsGetGetClientError verifies error handling when getClient fails.
func TestDomainsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := domainsGetCmd.RunE(cmd, []string{"domain-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestDomainsCreateGetClientError verifies error handling when getClient fails.
func TestDomainsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("host", "example.com", "")
	cmd.Flags().Bool("primary", false, "")

	err := domainsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestDomainsUpdateGetClientError verifies error handling when getClient fails.
func TestDomainsUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Bool("primary", false, "")

	err := domainsUpdateCmd.RunE(cmd, []string{"domain-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestDomainsDeleteGetClientError verifies error handling when getClient fails.
func TestDomainsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := domainsDeleteCmd.RunE(cmd, []string{"domain-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestDomainsVerifyGetClientError verifies error handling when getClient fails.
func TestDomainsVerifyGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := domainsVerifyCmd.RunE(cmd, []string{"domain-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestDomainsDeleteNoConfirmation verifies delete exits early without confirmation.
func TestDomainsDeleteNoConfirmation(t *testing.T) {
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
	// yes defaults to false, so delete should exit early

	err := domainsDeleteCmd.RunE(cmd, []string{"domain-id"})
	if err != nil {
		t.Errorf("Expected nil when confirmation not provided, got: %v", err)
	}
}

// TestDomainsCommandSetup verifies domains command initialization.
func TestDomainsCommandSetup(t *testing.T) {
	if domainsCmd.Use != "domains" {
		t.Errorf("expected Use 'domains', got %q", domainsCmd.Use)
	}
	if domainsCmd.Short != "Manage domains" {
		t.Errorf("expected Short 'Manage domains', got %q", domainsCmd.Short)
	}
}

// TestDomainsSubcommands verifies all subcommands are registered.
func TestDomainsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List domains",
		"get":    "Get domain details",
		"create": "Create a domain",
		"update": "Update a domain",
		"delete": "Delete a domain",
		"verify": "Verify domain ownership",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range domainsCmd.Commands() {
				cmdName := strings.Split(sub.Use, " ")[0]
				if cmdName == name {
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

// TestDomainsCreateDryRun verifies dry-run mode works for create.
func TestDomainsCreateDryRun(t *testing.T) {
	mockClient := &domainsMockAPIClient{}
	cleanup, _ := setupDomainsMockFactories(mockClient)
	defer cleanup()

	cmd := newDomainsTestCmd()
	cmd.Flags().String("host", "", "")
	cmd.Flags().Bool("primary", false, "")
	_ = cmd.Flags().Set("host", "dryrun.example.com")
	_ = cmd.Flags().Set("dry-run", "true")

	err := domainsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestDomainsUpdateDryRun verifies dry-run mode works for update.
func TestDomainsUpdateDryRun(t *testing.T) {
	mockClient := &domainsMockAPIClient{}
	cleanup, _ := setupDomainsMockFactories(mockClient)
	defer cleanup()

	cmd := newDomainsTestCmd()
	cmd.Flags().Bool("primary", false, "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := domainsUpdateCmd.RunE(cmd, []string{"dom_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestDomainsDeleteDryRun verifies dry-run mode works for delete.
func TestDomainsDeleteDryRun(t *testing.T) {
	mockClient := &domainsMockAPIClient{}
	cleanup, _ := setupDomainsMockFactories(mockClient)
	defer cleanup()

	cmd := newDomainsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := domainsDeleteCmd.RunE(cmd, []string{"dom_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestDomainsVerifyDryRun verifies dry-run mode works for verify.
func TestDomainsVerifyDryRun(t *testing.T) {
	mockClient := &domainsMockAPIClient{}
	cleanup, _ := setupDomainsMockFactories(mockClient)
	defer cleanup()

	cmd := newDomainsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := domainsVerifyCmd.RunE(cmd, []string{"dom_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}
