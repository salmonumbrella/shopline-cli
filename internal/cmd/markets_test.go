package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// TestMarketsCommandSetup verifies markets command initialization
func TestMarketsCommandSetup(t *testing.T) {
	if marketsCmd.Use != "markets" {
		t.Errorf("expected Use 'markets', got %q", marketsCmd.Use)
	}
	if marketsCmd.Short != "Manage markets (regions)" {
		t.Errorf("expected Short 'Manage markets (regions)', got %q", marketsCmd.Short)
	}
}

// TestMarketsSubcommands verifies all subcommands are registered
func TestMarketsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List markets",
		"get":    "Get market details",
		"create": "Create a market",
		"delete": "Delete a market",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range marketsCmd.Commands() {
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

// TestMarketsListFlags verifies list command flags exist with correct defaults
func TestMarketsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := marketsListCmd.Flags().Lookup(f.name)
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

// TestMarketsCreateFlags verifies create command flags exist with correct defaults
func TestMarketsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"handle", ""},
		{"enabled", "true"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := marketsCreateCmd.Flags().Lookup(f.name)
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

// TestMarketsGetArgs verifies get command argument validation
func TestMarketsGetArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"mkt_123"}, false},
		{"too many args", []string{"mkt_123", "mkt_456"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := marketsGetCmd.Args(marketsGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMarketsDeleteArgs verifies delete command argument validation
func TestMarketsDeleteArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"mkt_123"}, false},
		{"too many args", []string{"mkt_123", "mkt_456"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := marketsDeleteCmd.Args(marketsDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// marketsMockAPIClient is a mock implementation of api.APIClient for markets tests.
type marketsMockAPIClient struct {
	api.MockClient
	listMarketsResp  *api.MarketsListResponse
	listMarketsErr   error
	getMarketResp    *api.Market
	getMarketErr     error
	createMarketResp *api.Market
	createMarketErr  error
	deleteMarketErr  error
}

func (m *marketsMockAPIClient) ListMarkets(ctx context.Context, opts *api.MarketsListOptions) (*api.MarketsListResponse, error) {
	return m.listMarketsResp, m.listMarketsErr
}

func (m *marketsMockAPIClient) GetMarket(ctx context.Context, id string) (*api.Market, error) {
	return m.getMarketResp, m.getMarketErr
}

func (m *marketsMockAPIClient) CreateMarket(ctx context.Context, req *api.MarketCreateRequest) (*api.Market, error) {
	return m.createMarketResp, m.createMarketErr
}

func (m *marketsMockAPIClient) DeleteMarket(ctx context.Context, id string) error {
	return m.deleteMarketErr
}

// setupMarketsMockFactories sets up mock factories for markets tests.
func setupMarketsMockFactories(mockClient *marketsMockAPIClient) (func(), *bytes.Buffer) {
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

// newMarketsTestCmd creates a test command with common flags for markets tests.
func newMarketsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

// captureMarketsOutput captures both the formatter buffer and stdout during execution.
func captureMarketsOutput(formatterBuf *bytes.Buffer, fn func() error) (string, error) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := fn()

	_ = w.Close()
	os.Stdout = oldStdout

	var stdoutBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, r)

	// Combine formatter buffer and stdout
	return formatterBuf.String() + stdoutBuf.String(), err
}

// TestMarketsListRunE tests the markets list command with mock API.
func TestMarketsListRunE(t *testing.T) {
	tests := []struct {
		name        string
		mockResp    *api.MarketsListResponse
		mockErr     error
		outputJSON  bool
		wantErr     bool
		wantOutput  string
		wantMissing string
	}{
		{
			name: "successful list with multiple markets",
			mockResp: &api.MarketsListResponse{
				Items: []api.Market{
					{
						ID:         "mkt_001",
						Name:       "United States",
						Handle:     "us",
						Primary:    true,
						Enabled:    true,
						Countries:  []string{"US"},
						Currencies: []string{"USD"},
						Languages:  []string{"en"},
						CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:         "mkt_002",
						Name:       "Europe",
						Handle:     "eu",
						Primary:    false,
						Enabled:    true,
						Countries:  []string{"DE", "FR", "IT"},
						Currencies: []string{"EUR"},
						Languages:  []string{"de", "fr", "it"},
						CreatedAt:  time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
				Page:       1,
				PageSize:   20,
			},
			wantOutput: "mkt_001",
		},
		{
			name: "successful list with JSON output",
			mockResp: &api.MarketsListResponse{
				Items: []api.Market{
					{
						ID:        "mkt_001",
						Name:      "United States",
						Handle:    "us",
						Primary:   true,
						Enabled:   true,
						Countries: []string{"US"},
					},
				},
				TotalCount: 1,
			},
			outputJSON: true,
			wantOutput: "mkt_001",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.MarketsListResponse{
				Items:      []api.Market{},
				TotalCount: 0,
			},
			wantOutput: "0 of 0 markets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &marketsMockAPIClient{
				listMarketsResp: tt.mockResp,
				listMarketsErr:  tt.mockErr,
			}
			cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
			defer cleanup()

			cmd := newMarketsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			if tt.outputJSON {
				_ = cmd.Flags().Set("output", "json")
			}

			output, err := captureMarketsOutput(formatterBuf, func() error {
				return marketsListCmd.RunE(cmd, []string{})
			})

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

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
			if tt.wantMissing != "" && strings.Contains(output, tt.wantMissing) {
				t.Errorf("output %q should not contain %q", output, tt.wantMissing)
			}
		})
	}
}

// TestMarketsListRunETableOutput tests that table output shows correct columns
func TestMarketsListRunETableOutput(t *testing.T) {
	mockClient := &marketsMockAPIClient{
		listMarketsResp: &api.MarketsListResponse{
			Items: []api.Market{
				{
					ID:        "mkt_test",
					Name:      "Test Market",
					Handle:    "test-mkt",
					Primary:   false,
					Enabled:   true,
					Countries: []string{"CA", "MX"},
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
	defer cleanup()

	cmd := newMarketsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	output, err := captureMarketsOutput(formatterBuf, func() error {
		return marketsListCmd.RunE(cmd, []string{})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check table headers
	expectedHeaders := []string{"ID", "NAME", "HANDLE", "PRIMARY", "ENABLED", "COUNTRIES"}
	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("output should contain header %q", header)
		}
	}

	// Check row content
	expectedValues := []string{"mkt_test", "Test Market", "test-mkt", "false", "true", "2"}
	for _, val := range expectedValues {
		if !strings.Contains(output, val) {
			t.Errorf("output should contain value %q", val)
		}
	}
}

// TestMarketsGetRunE tests the markets get command with mock API.
func TestMarketsGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		marketID   string
		mockResp   *api.Market
		mockErr    error
		outputJSON bool
		wantErr    bool
		wantOutput string
	}{
		{
			name:     "successful get",
			marketID: "mkt_123",
			mockResp: &api.Market{
				ID:         "mkt_123",
				Name:       "United States",
				Handle:     "us",
				Primary:    true,
				Enabled:    true,
				Countries:  []string{"US"},
				Currencies: []string{"USD"},
				Languages:  []string{"en"},
			},
			wantOutput: "Market ID:   mkt_123",
		},
		{
			name:     "successful get JSON output",
			marketID: "mkt_123",
			mockResp: &api.Market{
				ID:         "mkt_123",
				Name:       "United States",
				Handle:     "us",
				Primary:    true,
				Enabled:    true,
				Countries:  []string{"US"},
				Currencies: []string{"USD"},
				Languages:  []string{"en"},
			},
			outputJSON: true,
			wantOutput: "mkt_123",
		},
		{
			name:     "market not found",
			marketID: "mkt_999",
			mockErr:  errors.New("market not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &marketsMockAPIClient{
				getMarketResp: tt.mockResp,
				getMarketErr:  tt.mockErr,
			}
			cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
			defer cleanup()

			cmd := newMarketsTestCmd()
			if tt.outputJSON {
				_ = cmd.Flags().Set("output", "json")
			}

			output, err := captureMarketsOutput(formatterBuf, func() error {
				return marketsGetCmd.RunE(cmd, []string{tt.marketID})
			})

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

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestMarketsGetRunEDetailedOutput tests that get displays all market fields
func TestMarketsGetRunEDetailedOutput(t *testing.T) {
	mockClient := &marketsMockAPIClient{
		getMarketResp: &api.Market{
			ID:         "mkt_detailed",
			Name:       "European Union",
			Handle:     "eu-market",
			Primary:    false,
			Enabled:    true,
			Countries:  []string{"DE", "FR", "IT", "ES"},
			Currencies: []string{"EUR"},
			Languages:  []string{"de", "fr", "it", "es"},
		},
	}
	cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
	defer cleanup()

	cmd := newMarketsTestCmd()

	output, err := captureMarketsOutput(formatterBuf, func() error {
		return marketsGetCmd.RunE(cmd, []string{"mkt_detailed"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFields := []string{
		"Market ID:",
		"Name:",
		"Handle:",
		"Primary:",
		"Enabled:",
		"Countries:",
		"Currencies:",
		"Languages:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("output should contain field %q", field)
		}
	}
}

// TestMarketsCreateRunE tests the markets create command with mock API.
func TestMarketsCreateRunE(t *testing.T) {
	tests := []struct {
		name       string
		marketName string
		handle     string
		enabled    bool
		mockResp   *api.Market
		mockErr    error
		outputJSON bool
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "successful create",
			marketName: "Asia Pacific",
			handle:     "apac",
			enabled:    true,
			mockResp: &api.Market{
				ID:      "mkt_new",
				Name:    "Asia Pacific",
				Handle:  "apac",
				Enabled: true,
			},
			wantOutput: "Created market mkt_new",
		},
		{
			name:       "successful create JSON output",
			marketName: "Asia Pacific",
			handle:     "apac",
			enabled:    true,
			mockResp: &api.Market{
				ID:      "mkt_new",
				Name:    "Asia Pacific",
				Handle:  "apac",
				Enabled: true,
			},
			outputJSON: true,
			wantOutput: "mkt_new",
		},
		{
			name:       "create without handle",
			marketName: "Latin America",
			enabled:    true,
			mockResp: &api.Market{
				ID:      "mkt_latam",
				Name:    "Latin America",
				Handle:  "latin-america",
				Enabled: true,
			},
			wantOutput: "Created market mkt_latam",
		},
		{
			name:       "create disabled market",
			marketName: "Test Market",
			handle:     "test",
			enabled:    false,
			mockResp: &api.Market{
				ID:      "mkt_disabled",
				Name:    "Test Market",
				Handle:  "test",
				Enabled: false,
			},
			wantOutput: "Created market mkt_disabled",
		},
		{
			name:       "API error",
			marketName: "Error Market",
			mockErr:    errors.New("failed to create market"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &marketsMockAPIClient{
				createMarketResp: tt.mockResp,
				createMarketErr:  tt.mockErr,
			}
			cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
			defer cleanup()

			cmd := newMarketsTestCmd()
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("handle", "", "")
			cmd.Flags().Bool("enabled", true, "")

			_ = cmd.Flags().Set("name", tt.marketName)
			if tt.handle != "" {
				_ = cmd.Flags().Set("handle", tt.handle)
			}
			_ = cmd.Flags().Set("enabled", boolToString(tt.enabled))

			if tt.outputJSON {
				_ = cmd.Flags().Set("output", "json")
			}

			output, err := captureMarketsOutput(formatterBuf, func() error {
				return marketsCreateCmd.RunE(cmd, []string{})
			})

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

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestMarketsDeleteRunE tests the markets delete command with mock API.
func TestMarketsDeleteRunE(t *testing.T) {
	tests := []struct {
		name       string
		marketID   string
		mockErr    error
		confirmed  bool
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "successful delete with confirmation",
			marketID:   "mkt_123",
			confirmed:  true,
			wantOutput: "Deleted market mkt_123",
		},
		{
			name:      "delete API error",
			marketID:  "mkt_456",
			mockErr:   errors.New("failed to delete market"),
			confirmed: true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &marketsMockAPIClient{
				deleteMarketErr: tt.mockErr,
			}
			cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
			defer cleanup()

			cmd := newMarketsTestCmd()

			output, err := captureMarketsOutput(formatterBuf, func() error {
				return marketsDeleteCmd.RunE(cmd, []string{tt.marketID})
			})

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

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestMarketsDeleteRunEWithoutYesFlag tests the delete command's interactive confirmation
func TestMarketsDeleteRunEWithoutYesFlag(t *testing.T) {
	mockClient := &marketsMockAPIClient{}
	cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Replace stdin with our test input
	oldStdin := os.Stdin
	stdinR, stdinW, _ := os.Pipe()
	os.Stdin = stdinR
	go func() {
		_, _ = stdinW.Write([]byte("N\n"))
		_ = stdinW.Close()
	}()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().Bool("yes", false, "")

	err := marketsDeleteCmd.RunE(cmd, []string{"mkt_cancel"})

	_ = w.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var stdoutBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, r)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := formatterBuf.String() + stdoutBuf.String()

	// Verify cancellation message is shown
	if !strings.Contains(output, "Cancelled") {
		t.Errorf("output should contain 'Cancelled', got: %s", output)
	}
}

// TestMarketsDeleteConfirmYes tests the delete command with 'y' confirmation
func TestMarketsDeleteConfirmYes(t *testing.T) {
	mockClient := &marketsMockAPIClient{}
	cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Replace stdin with our test input
	oldStdin := os.Stdin
	stdinR, stdinW, _ := os.Pipe()
	os.Stdin = stdinR
	go func() {
		_, _ = stdinW.Write([]byte("y\n"))
		_ = stdinW.Close()
	}()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().Bool("yes", false, "")

	err := marketsDeleteCmd.RunE(cmd, []string{"mkt_confirm"})

	_ = w.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var stdoutBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, r)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := formatterBuf.String() + stdoutBuf.String()

	// Verify deletion message is shown
	if !strings.Contains(output, "Deleted market mkt_confirm") {
		t.Errorf("output should contain 'Deleted market mkt_confirm', got: %s", output)
	}
}

// TestMarketsDeleteConfirmUpperY tests the delete command with 'Y' confirmation
func TestMarketsDeleteConfirmUpperY(t *testing.T) {
	mockClient := &marketsMockAPIClient{}
	cleanup, formatterBuf := setupMarketsMockFactories(mockClient)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Replace stdin with our test input
	oldStdin := os.Stdin
	stdinR, stdinW, _ := os.Pipe()
	os.Stdin = stdinR
	go func() {
		_, _ = stdinW.Write([]byte("Y\n"))
		_ = stdinW.Close()
	}()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().Bool("yes", false, "")

	err := marketsDeleteCmd.RunE(cmd, []string{"mkt_confirm_upper"})

	_ = w.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var stdoutBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, r)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := formatterBuf.String() + stdoutBuf.String()

	// Verify deletion message is shown
	if !strings.Contains(output, "Deleted market mkt_confirm_upper") {
		t.Errorf("output should contain 'Deleted market mkt_confirm_upper', got: %s", output)
	}
}

// TestMarketsGetClientError verifies error handling when getClient fails
func TestMarketsGetClientError(t *testing.T) {
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

// TestMarketsListGetClientError tests list command when getClient fails
func TestMarketsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMarketsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := marketsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "keyring") {
		t.Errorf("expected keyring error, got: %v", err)
	}
}

// TestMarketsGetGetClientError tests get command when getClient fails
func TestMarketsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMarketsTestCmd()

	err := marketsGetCmd.RunE(cmd, []string{"mkt_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestMarketsCreateGetClientError tests create command when getClient fails
func TestMarketsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMarketsTestCmd()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("handle", "test", "")
	cmd.Flags().Bool("enabled", true, "")

	err := marketsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestMarketsDeleteGetClientError tests delete command when getClient fails
func TestMarketsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMarketsTestCmd()

	err := marketsDeleteCmd.RunE(cmd, []string{"mkt_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestMarketsWithMockStore tests markets commands with a mock credential store
func TestMarketsWithMockStore(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

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
