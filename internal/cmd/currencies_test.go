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

// currenciesMockClient is a mock implementation of api.APIClient for currencies tests.
type currenciesMockClient struct {
	api.MockClient
	listCurrenciesResp *api.CurrenciesListResponse
	listCurrenciesErr  error
	getCurrencyResp    *api.Currency
	getCurrencyErr     error
	updateCurrencyResp *api.Currency
	updateCurrencyErr  error
	updateCurrencyCode string
	updateCurrencyReq  *api.CurrencyUpdateRequest
}

func (m *currenciesMockClient) ListCurrencies(ctx context.Context) (*api.CurrenciesListResponse, error) {
	return m.listCurrenciesResp, m.listCurrenciesErr
}

func (m *currenciesMockClient) GetCurrency(ctx context.Context, code string) (*api.Currency, error) {
	return m.getCurrencyResp, m.getCurrencyErr
}

func (m *currenciesMockClient) UpdateCurrency(ctx context.Context, code string, req *api.CurrencyUpdateRequest) (*api.Currency, error) {
	m.updateCurrencyCode = code
	m.updateCurrencyReq = req
	return m.updateCurrencyResp, m.updateCurrencyErr
}

// setupCurrenciesMockFactories sets up mock factories for currencies tests.
func setupCurrenciesMockFactories(mockClient *currenciesMockClient) (func(), *bytes.Buffer) {
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

// newCurrenciesTestCmd creates a test command with common flags.
func newCurrenciesTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	// Add currencies update flags
	cmd.Flags().Bool("enabled", false, "")
	cmd.Flags().Float64("exchange-rate", 0, "")
	cmd.Flags().Bool("auto-update", false, "")
	return cmd
}

// TestCurrenciesCommandStructure verifies the currencies command structure.
func TestCurrenciesCommandStructure(t *testing.T) {
	if currenciesCmd.Use != "currencies" {
		t.Errorf("Expected Use 'currencies', got %q", currenciesCmd.Use)
	}
	if currenciesCmd.Short != "Manage currencies" {
		t.Errorf("Expected Short 'Manage currencies', got %q", currenciesCmd.Short)
	}

	subcommands := currenciesCmd.Commands()
	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"update": false,
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

// TestCurrenciesGetArgs verifies argument validation for get command.
func TestCurrenciesGetArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"USD"}, false},
		{"too many args", []string{"USD", "EUR"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := currenciesGetCmd.Args(currenciesGetCmd, tt.args)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestCurrenciesUpdateArgs verifies argument validation for update command.
func TestCurrenciesUpdateArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"USD"}, false},
		{"too many args", []string{"USD", "EUR"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := currenciesUpdateCmd.Args(currenciesUpdateCmd, tt.args)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestCurrenciesUpdateFlags verifies update command flags exist.
func TestCurrenciesUpdateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"enabled", "false"},
		{"exchange-rate", "0"},
		{"auto-update", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := currenciesUpdateCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Expected flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("Expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

// TestCurrenciesListGetClientError verifies error handling when getClient fails.
func TestCurrenciesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.SetContext(context.Background())

	err := currenciesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestCurrenciesGetGetClientError verifies error handling when getClient fails.
func TestCurrenciesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.SetContext(context.Background())

	err := currenciesGetCmd.RunE(cmd, []string{"USD"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestCurrenciesUpdateGetClientError verifies error handling when getClient fails.
func TestCurrenciesUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newCurrenciesTestCmd()

	err := currenciesUpdateCmd.RunE(cmd, []string{"USD"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestCurrenciesListRunE tests the list command with mock API.
func TestCurrenciesListRunE(t *testing.T) {
	tests := []struct {
		name           string
		mockResp       *api.CurrenciesListResponse
		mockErr        error
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "successful list with multiple currencies",
			mockResp: &api.CurrenciesListResponse{
				Items: []api.Currency{
					{
						Code:         "USD",
						Name:         "US Dollar",
						Symbol:       "$",
						Primary:      true,
						Enabled:      true,
						ExchangeRate: 1.0,
						AutoUpdate:   false,
						UpdatedAt:    time.Now(),
					},
					{
						Code:         "EUR",
						Name:         "Euro",
						Symbol:       "€",
						Primary:      false,
						Enabled:      true,
						ExchangeRate: 0.85,
						AutoUpdate:   true,
						UpdatedAt:    time.Now(),
					},
				},
			},
		},
		{
			name: "successful list with empty currencies",
			mockResp: &api.CurrenciesListResponse{
				Items: []api.Currency{},
			},
		},
		{
			name:           "API error",
			mockErr:        errors.New("API unavailable"),
			wantErr:        true,
			wantErrContain: "failed to list currencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &currenciesMockClient{
				listCurrenciesResp: tt.mockResp,
				listCurrenciesErr:  tt.mockErr,
			}
			cleanup, _ := setupCurrenciesMockFactories(mockClient)
			defer cleanup()

			cmd := newCurrenciesTestCmd()

			err := currenciesListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("expected error to contain %q, got %q", tt.wantErrContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCurrenciesListRunEWithJSON tests JSON output format for list command.
func TestCurrenciesListRunEWithJSON(t *testing.T) {
	mockClient := &currenciesMockClient{
		listCurrenciesResp: &api.CurrenciesListResponse{
			Items: []api.Currency{
				{
					Code:         "USD",
					Name:         "US Dollar",
					Symbol:       "$",
					Primary:      true,
					Enabled:      true,
					ExchangeRate: 1.0,
				},
			},
		},
	}
	cleanup, buf := setupCurrenciesMockFactories(mockClient)
	defer cleanup()

	cmd := newCurrenciesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := currenciesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "USD") {
		t.Errorf("JSON output should contain currency code, got: %s", output)
	}
}

// TestCurrenciesGetRunE tests the get command with mock API.
func TestCurrenciesGetRunE(t *testing.T) {
	tests := []struct {
		name           string
		currencyCode   string
		mockResp       *api.Currency
		mockErr        error
		wantErr        bool
		wantErrContain string
	}{
		{
			name:         "successful get USD",
			currencyCode: "USD",
			mockResp: &api.Currency{
				Code:         "USD",
				Name:         "US Dollar",
				Symbol:       "$",
				Primary:      true,
				Enabled:      true,
				RoundingMode: "round",
				ExchangeRate: 1.0,
				AutoUpdate:   false,
				UpdatedAt:    time.Now(),
			},
		},
		{
			name:         "successful get EUR",
			currencyCode: "EUR",
			mockResp: &api.Currency{
				Code:         "EUR",
				Name:         "Euro",
				Symbol:       "€",
				Primary:      false,
				Enabled:      true,
				RoundingMode: "round",
				ExchangeRate: 0.85,
				AutoUpdate:   true,
				UpdatedAt:    time.Now(),
			},
		},
		{
			name:           "currency not found",
			currencyCode:   "XYZ",
			mockErr:        errors.New("not found"),
			wantErr:        true,
			wantErrContain: "failed to get currency",
		},
		{
			name:           "API error",
			currencyCode:   "USD",
			mockErr:        errors.New("API unavailable"),
			wantErr:        true,
			wantErrContain: "failed to get currency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &currenciesMockClient{
				getCurrencyResp: tt.mockResp,
				getCurrencyErr:  tt.mockErr,
			}
			cleanup, _ := setupCurrenciesMockFactories(mockClient)
			defer cleanup()

			cmd := newCurrenciesTestCmd()

			err := currenciesGetCmd.RunE(cmd, []string{tt.currencyCode})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("expected error to contain %q, got %q", tt.wantErrContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCurrenciesGetRunEWithJSON tests JSON output format for get command.
func TestCurrenciesGetRunEWithJSON(t *testing.T) {
	mockClient := &currenciesMockClient{
		getCurrencyResp: &api.Currency{
			Code:         "USD",
			Name:         "US Dollar",
			Symbol:       "$",
			Primary:      true,
			Enabled:      true,
			ExchangeRate: 1.0,
			AutoUpdate:   false,
		},
	}
	cleanup, buf := setupCurrenciesMockFactories(mockClient)
	defer cleanup()

	cmd := newCurrenciesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := currenciesGetCmd.RunE(cmd, []string{"USD"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "USD") {
		t.Errorf("JSON output should contain currency code, got: %s", output)
	}
}

// TestCurrenciesUpdateRunE tests the update command with mock API.
func TestCurrenciesUpdateRunE(t *testing.T) {
	tests := []struct {
		name           string
		currencyCode   string
		setEnabled     bool
		enabledValue   bool
		setExchange    bool
		exchangeValue  float64
		setAutoUpdate  bool
		autoUpdateVal  bool
		mockResp       *api.Currency
		mockErr        error
		wantErr        bool
		wantErrContain string
	}{
		{
			name:         "update enabled to true",
			currencyCode: "EUR",
			setEnabled:   true,
			enabledValue: true,
			mockResp: &api.Currency{
				Code:         "EUR",
				Name:         "Euro",
				Symbol:       "€",
				Primary:      false,
				Enabled:      true,
				ExchangeRate: 0.85,
			},
		},
		{
			name:         "update enabled to false",
			currencyCode: "EUR",
			setEnabled:   true,
			enabledValue: false,
			mockResp: &api.Currency{
				Code:         "EUR",
				Name:         "Euro",
				Symbol:       "€",
				Primary:      false,
				Enabled:      false,
				ExchangeRate: 0.85,
			},
		},
		{
			name:          "update exchange rate",
			currencyCode:  "EUR",
			setExchange:   true,
			exchangeValue: 0.92,
			mockResp: &api.Currency{
				Code:         "EUR",
				Name:         "Euro",
				Symbol:       "€",
				Primary:      false,
				Enabled:      true,
				ExchangeRate: 0.92,
			},
		},
		{
			name:          "update auto-update",
			currencyCode:  "EUR",
			setAutoUpdate: true,
			autoUpdateVal: true,
			mockResp: &api.Currency{
				Code:         "EUR",
				Name:         "Euro",
				Symbol:       "€",
				Primary:      false,
				Enabled:      true,
				ExchangeRate: 0.85,
				AutoUpdate:   true,
			},
		},
		{
			name:          "update multiple fields",
			currencyCode:  "GBP",
			setEnabled:    true,
			enabledValue:  true,
			setExchange:   true,
			exchangeValue: 0.79,
			setAutoUpdate: true,
			autoUpdateVal: true,
			mockResp: &api.Currency{
				Code:         "GBP",
				Name:         "British Pound",
				Symbol:       "£",
				Primary:      false,
				Enabled:      true,
				ExchangeRate: 0.79,
				AutoUpdate:   true,
			},
		},
		{
			name:         "update with no flags changed",
			currencyCode: "EUR",
			mockResp: &api.Currency{
				Code:         "EUR",
				Name:         "Euro",
				Symbol:       "€",
				Primary:      false,
				Enabled:      true,
				ExchangeRate: 0.85,
			},
		},
		{
			name:           "currency not found",
			currencyCode:   "XYZ",
			setEnabled:     true,
			enabledValue:   true,
			mockErr:        errors.New("not found"),
			wantErr:        true,
			wantErrContain: "failed to update currency",
		},
		{
			name:           "API error",
			currencyCode:   "EUR",
			setEnabled:     true,
			enabledValue:   true,
			mockErr:        errors.New("API unavailable"),
			wantErr:        true,
			wantErrContain: "failed to update currency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &currenciesMockClient{
				updateCurrencyResp: tt.mockResp,
				updateCurrencyErr:  tt.mockErr,
			}
			cleanup, _ := setupCurrenciesMockFactories(mockClient)
			defer cleanup()

			cmd := newCurrenciesTestCmd()

			// Mark flags as changed by setting them
			if tt.setEnabled {
				_ = cmd.Flags().Set("enabled", boolToString(tt.enabledValue))
			}
			if tt.setExchange {
				_ = cmd.Flags().Set("exchange-rate", floatToString(tt.exchangeValue))
			}
			if tt.setAutoUpdate {
				_ = cmd.Flags().Set("auto-update", boolToString(tt.autoUpdateVal))
			}

			err := currenciesUpdateCmd.RunE(cmd, []string{tt.currencyCode})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("expected error to contain %q, got %q", tt.wantErrContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify the request was built correctly
			if mockClient.updateCurrencyCode != tt.currencyCode {
				t.Errorf("expected currency code %q, got %q", tt.currencyCode, mockClient.updateCurrencyCode)
			}
		})
	}
}

// TestCurrenciesUpdateRunEWithJSON tests JSON output format for update command.
func TestCurrenciesUpdateRunEWithJSON(t *testing.T) {
	mockClient := &currenciesMockClient{
		updateCurrencyResp: &api.Currency{
			Code:         "EUR",
			Name:         "Euro",
			Symbol:       "€",
			Primary:      false,
			Enabled:      true,
			ExchangeRate: 0.92,
		},
	}
	cleanup, buf := setupCurrenciesMockFactories(mockClient)
	defer cleanup()

	cmd := newCurrenciesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("exchange-rate", "0.92")

	err := currenciesUpdateCmd.RunE(cmd, []string{"EUR"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "EUR") {
		t.Errorf("JSON output should contain currency code, got: %s", output)
	}
}

// TestCurrenciesUpdateRequestBuilding verifies the update request is built correctly.
func TestCurrenciesUpdateRequestBuilding(t *testing.T) {
	tests := []struct {
		name              string
		setEnabled        bool
		enabledValue      bool
		setExchange       bool
		exchangeValue     float64
		setAutoUpdate     bool
		autoUpdateVal     bool
		wantEnabledSet    bool
		wantExchangeSet   bool
		wantAutoUpdateSet bool
	}{
		{
			name:           "only enabled flag",
			setEnabled:     true,
			enabledValue:   true,
			wantEnabledSet: true,
		},
		{
			name:            "only exchange-rate flag",
			setExchange:     true,
			exchangeValue:   1.25,
			wantExchangeSet: true,
		},
		{
			name:              "only auto-update flag",
			setAutoUpdate:     true,
			autoUpdateVal:     true,
			wantAutoUpdateSet: true,
		},
		{
			name:              "all flags",
			setEnabled:        true,
			enabledValue:      true,
			setExchange:       true,
			exchangeValue:     1.5,
			setAutoUpdate:     true,
			autoUpdateVal:     false,
			wantEnabledSet:    true,
			wantExchangeSet:   true,
			wantAutoUpdateSet: true,
		},
		{
			name: "no flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &currenciesMockClient{
				updateCurrencyResp: &api.Currency{
					Code:    "EUR",
					Name:    "Euro",
					Enabled: true,
				},
			}
			cleanup, _ := setupCurrenciesMockFactories(mockClient)
			defer cleanup()

			cmd := newCurrenciesTestCmd()

			if tt.setEnabled {
				_ = cmd.Flags().Set("enabled", boolToString(tt.enabledValue))
			}
			if tt.setExchange {
				_ = cmd.Flags().Set("exchange-rate", floatToString(tt.exchangeValue))
			}
			if tt.setAutoUpdate {
				_ = cmd.Flags().Set("auto-update", boolToString(tt.autoUpdateVal))
			}

			err := currenciesUpdateCmd.RunE(cmd, []string{"EUR"})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			req := mockClient.updateCurrencyReq
			if req == nil {
				t.Fatal("expected update request, got nil")
			}

			// Check Enabled field
			if tt.wantEnabledSet {
				if req.Enabled == nil {
					t.Error("expected Enabled to be set")
				} else if *req.Enabled != tt.enabledValue {
					t.Errorf("expected Enabled=%v, got %v", tt.enabledValue, *req.Enabled)
				}
			} else if req.Enabled != nil {
				t.Error("expected Enabled to be nil")
			}

			// Check ExchangeRate field
			if tt.wantExchangeSet {
				if req.ExchangeRate == nil {
					t.Error("expected ExchangeRate to be set")
				} else if *req.ExchangeRate != tt.exchangeValue {
					t.Errorf("expected ExchangeRate=%v, got %v", tt.exchangeValue, *req.ExchangeRate)
				}
			} else if req.ExchangeRate != nil {
				t.Error("expected ExchangeRate to be nil")
			}

			// Check AutoUpdate field
			if tt.wantAutoUpdateSet {
				if req.AutoUpdate == nil {
					t.Error("expected AutoUpdate to be set")
				} else if *req.AutoUpdate != tt.autoUpdateVal {
					t.Errorf("expected AutoUpdate=%v, got %v", tt.autoUpdateVal, *req.AutoUpdate)
				}
			} else if req.AutoUpdate != nil {
				t.Error("expected AutoUpdate to be nil")
			}
		})
	}
}

// TestCurrenciesListNoProfiles verifies error when no profiles are configured.
func TestCurrenciesListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newCurrenciesTestCmd()
	err := currenciesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestCurrenciesGetMultipleProfiles verifies error when multiple profiles exist without selection.
func TestCurrenciesGetMultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newCurrenciesTestCmd()
	err := currenciesGetCmd.RunE(cmd, []string{"USD"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
	if !strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Expected 'multiple profiles' error, got: %v", err)
	}
}

// TestCurrenciesListTableOutput verifies table output contains expected headers.
func TestCurrenciesListTableOutput(t *testing.T) {
	mockClient := &currenciesMockClient{
		listCurrenciesResp: &api.CurrenciesListResponse{
			Items: []api.Currency{
				{
					Code:         "USD",
					Name:         "US Dollar",
					Symbol:       "$",
					Primary:      true,
					Enabled:      true,
					ExchangeRate: 1.0,
				},
				{
					Code:         "EUR",
					Name:         "Euro",
					Symbol:       "€",
					Primary:      false,
					Enabled:      true,
					ExchangeRate: 0.85,
				},
			},
		},
	}
	cleanup, buf := setupCurrenciesMockFactories(mockClient)
	defer cleanup()

	cmd := newCurrenciesTestCmd()
	_ = cmd.Flags().Set("output", "text")

	err := currenciesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check for expected headers
	expectedHeaders := []string{"CODE", "NAME", "SYMBOL", "PRIMARY", "ENABLED", "RATE"}
	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("Table output should contain header %q, got: %s", header, output)
		}
	}
}

// TestCurrenciesGetTextOutput verifies text output contains expected fields.
func TestCurrenciesGetTextOutput(t *testing.T) {
	mockClient := &currenciesMockClient{
		getCurrencyResp: &api.Currency{
			Code:         "USD",
			Name:         "US Dollar",
			Symbol:       "$",
			Primary:      true,
			Enabled:      true,
			ExchangeRate: 1.0,
			AutoUpdate:   false,
		},
	}
	cleanup, _ := setupCurrenciesMockFactories(mockClient)
	defer cleanup()

	cmd := newCurrenciesTestCmd()
	_ = cmd.Flags().Set("output", "text")

	err := currenciesGetCmd.RunE(cmd, []string{"USD"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// The text output goes to stdout via fmt.Printf, not the buffer.
	// This test verifies the command runs without error for text output.
}

// TestCurrenciesUpdateTextOutput verifies text output for update command.
func TestCurrenciesUpdateTextOutput(t *testing.T) {
	mockClient := &currenciesMockClient{
		updateCurrencyResp: &api.Currency{
			Code:         "EUR",
			Name:         "Euro",
			Symbol:       "€",
			Primary:      false,
			Enabled:      true,
			ExchangeRate: 0.92,
		},
	}
	cleanup, _ := setupCurrenciesMockFactories(mockClient)
	defer cleanup()

	cmd := newCurrenciesTestCmd()
	_ = cmd.Flags().Set("output", "text")
	_ = cmd.Flags().Set("exchange-rate", "0.92")

	err := currenciesUpdateCmd.RunE(cmd, []string{"EUR"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// The text output goes to stdout via fmt.Printf, not the buffer.
	// This test verifies the command runs without error for text output.
}
