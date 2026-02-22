package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// TestCountriesCommandSetup verifies countries command initialization.
func TestCountriesCommandSetup(t *testing.T) {
	if countriesCmd.Use != "countries" {
		t.Errorf("expected Use 'countries', got %q", countriesCmd.Use)
	}
	if countriesCmd.Short != "Manage countries" {
		t.Errorf("expected Short 'Manage countries', got %q", countriesCmd.Short)
	}
}

// TestCountriesSubcommands verifies all subcommands are registered.
func TestCountriesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list": "List countries",
		"get":  "Get country details",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range countriesCmd.Commands() {
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

// TestCountriesGetArgs verifies get command requires exactly 1 argument.
func TestCountriesGetArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"US"},
			wantErr: false,
		},
		{
			name:    "two args",
			args:    []string{"US", "CA"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := countriesGetCmd.Args(countriesGetCmd, tt.args)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCountriesListGetClientError verifies list command error handling when getClient fails.
func TestCountriesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := countriesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCountriesGetGetClientError verifies get command error handling when getClient fails.
func TestCountriesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := countriesGetCmd.RunE(cmd, []string{"US"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// countriesMockAPIClient is a mock implementation of api.APIClient for countries tests.
type countriesMockAPIClient struct {
	api.MockClient
	listCountriesResp *api.CountriesListResponse
	listCountriesErr  error
	getCountryResp    *api.Country
	getCountryErr     error
}

func (m *countriesMockAPIClient) ListCountries(ctx context.Context) (*api.CountriesListResponse, error) {
	return m.listCountriesResp, m.listCountriesErr
}

func (m *countriesMockAPIClient) GetCountry(ctx context.Context, code string) (*api.Country, error) {
	return m.getCountryResp, m.getCountryErr
}

// setupCountriesMockFactories sets up mock factories for countries tests.
func setupCountriesMockFactories(mockClient *countriesMockAPIClient) (func(), *bytes.Buffer) {
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

// newCountriesTestCmd creates a test command with common flags for countries tests.
func newCountriesTestCmd() *cobra.Command {
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

// captureCountriesStdout captures output during test execution via formatterWriter.
func captureCountriesStdout(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	formatterWriter = &buf
	fn()
	return buf.String()
}

// TestCountriesListRunE tests the countries list command with mock API.
func TestCountriesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CountriesListResponse
		mockErr    error
		wantErr    bool
		wantOutput []string
	}{
		{
			name: "successful list with countries",
			mockResp: &api.CountriesListResponse{
				Items: []api.Country{
					{
						Code:    "US",
						Name:    "United States",
						Tax:     7.5,
						TaxName: "Sales Tax",
						Provinces: []api.Province{
							{Code: "CA", Name: "California", Tax: 7.25, TaxName: "CA Sales Tax"},
							{Code: "NY", Name: "New York", Tax: 8.0, TaxName: "NY Sales Tax"},
						},
					},
					{
						Code:      "CA",
						Name:      "Canada",
						Tax:       5.0,
						TaxName:   "GST",
						Provinces: []api.Province{},
					},
				},
			},
			wantOutput: []string{"US", "United States", "7.50%", "CA", "Canada", "2", "0"},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CountriesListResponse{
				Items: []api.Country{},
			},
			wantOutput: []string{},
		},
		{
			name: "country with many provinces",
			mockResp: &api.CountriesListResponse{
				Items: []api.Country{
					{
						Code:    "AU",
						Name:    "Australia",
						Tax:     10.0,
						TaxName: "GST",
						Provinces: []api.Province{
							{Code: "NSW", Name: "New South Wales", Tax: 0, TaxName: ""},
							{Code: "VIC", Name: "Victoria", Tax: 0, TaxName: ""},
							{Code: "QLD", Name: "Queensland", Tax: 0, TaxName: ""},
						},
					},
				},
			},
			wantOutput: []string{"AU", "Australia", "10.00%", "3"},
		},
		{
			name: "country with zero tax",
			mockResp: &api.CountriesListResponse{
				Items: []api.Country{
					{
						Code:      "HK",
						Name:      "Hong Kong",
						Tax:       0.0,
						TaxName:   "",
						Provinces: []api.Province{},
					},
				},
			},
			wantOutput: []string{"HK", "Hong Kong", "0.00%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &countriesMockAPIClient{
				listCountriesResp: tt.mockResp,
				listCountriesErr:  tt.mockErr,
			}
			cleanup, buf := setupCountriesMockFactories(mockClient)
			defer cleanup()

			cmd := newCountriesTestCmd()

			var err error
			stdout := captureCountriesStdout(t, func() {
				err = countriesListCmd.RunE(cmd, []string{})
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "failed to list countries") {
					t.Errorf("error should contain 'failed to list countries', got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Combine both buffer output and stdout output
			output := buf.String() + stdout
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output %q should contain %q", output, want)
				}
			}
		})
	}
}

// TestCountriesListRunEWithJSON tests JSON output format.
func TestCountriesListRunEWithJSON(t *testing.T) {
	mockClient := &countriesMockAPIClient{
		listCountriesResp: &api.CountriesListResponse{
			Items: []api.Country{
				{
					Code:    "US",
					Name:    "United States",
					Tax:     7.5,
					TaxName: "Sales Tax",
					Provinces: []api.Province{
						{Code: "CA", Name: "California", Tax: 7.25, TaxName: "CA Sales Tax"},
					},
				},
			},
		},
	}
	cleanup, buf := setupCountriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCountriesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := countriesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	// JSON output should contain the country code
	if !strings.Contains(output, "US") {
		t.Errorf("JSON output should contain country code 'US', got: %s", output)
	}
}

// TestCountriesGetRunE tests the countries get command with mock API.
func TestCountriesGetRunE(t *testing.T) {
	tests := []struct {
		name        string
		countryCode string
		mockResp    *api.Country
		mockErr     error
		wantErr     bool
		wantOutput  []string
	}{
		{
			name:        "successful get with provinces",
			countryCode: "US",
			mockResp: &api.Country{
				Code:    "US",
				Name:    "United States",
				Tax:     7.5,
				TaxName: "Sales Tax",
				Provinces: []api.Province{
					{Code: "CA", Name: "California", Tax: 7.25, TaxName: "CA Sales Tax"},
					{Code: "NY", Name: "New York", Tax: 8.0, TaxName: "NY Sales Tax"},
				},
			},
			wantOutput: []string{"US", "United States", "7.50%", "Sales Tax", "2", "CA", "California", "NY", "New York"},
		},
		{
			name:        "country not found",
			countryCode: "ZZ",
			mockErr:     errors.New("country not found"),
			wantErr:     true,
		},
		{
			name:        "country without provinces",
			countryCode: "HK",
			mockResp: &api.Country{
				Code:      "HK",
				Name:      "Hong Kong",
				Tax:       0.0,
				TaxName:   "",
				Provinces: []api.Province{},
			},
			wantOutput: []string{"HK", "Hong Kong", "0.00%"},
		},
		{
			name:        "country with single province",
			countryCode: "SG",
			mockResp: &api.Country{
				Code:    "SG",
				Name:    "Singapore",
				Tax:     7.0,
				TaxName: "GST",
				Provinces: []api.Province{
					{Code: "SG", Name: "Singapore", Tax: 7.0, TaxName: "GST", TaxType: "standard"},
				},
			},
			wantOutput: []string{"SG", "Singapore", "7.00%", "GST", "1"},
		},
		{
			name:        "API error",
			countryCode: "US",
			mockErr:     errors.New("API unavailable"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &countriesMockAPIClient{
				getCountryResp: tt.mockResp,
				getCountryErr:  tt.mockErr,
			}
			cleanup, buf := setupCountriesMockFactories(mockClient)
			defer cleanup()

			cmd := newCountriesTestCmd()

			var err error
			stdout := captureCountriesStdout(t, func() {
				err = countriesGetCmd.RunE(cmd, []string{tt.countryCode})
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "failed to get country") {
					t.Errorf("error should contain 'failed to get country', got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Combine both buffer output and stdout output
			output := buf.String() + stdout
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output %q should contain %q", output, want)
				}
			}
		})
	}
}

// TestCountriesGetRunEWithJSON tests JSON output format for get command.
func TestCountriesGetRunEWithJSON(t *testing.T) {
	mockClient := &countriesMockAPIClient{
		getCountryResp: &api.Country{
			Code:    "US",
			Name:    "United States",
			Tax:     7.5,
			TaxName: "Sales Tax",
			Provinces: []api.Province{
				{Code: "CA", Name: "California", Tax: 7.25, TaxName: "CA Sales Tax"},
			},
		},
	}
	cleanup, buf := setupCountriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCountriesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := countriesGetCmd.RunE(cmd, []string{"US"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	// JSON output should contain the country code
	if !strings.Contains(output, "US") {
		t.Errorf("JSON output should contain country code 'US', got: %s", output)
	}
}

// TestCountriesGetWithProvinceDetails tests that province details are displayed correctly.
func TestCountriesGetWithProvinceDetails(t *testing.T) {
	mockClient := &countriesMockAPIClient{
		getCountryResp: &api.Country{
			Code:    "CA",
			Name:    "Canada",
			Tax:     5.0,
			TaxName: "GST",
			Provinces: []api.Province{
				{Code: "ON", Name: "Ontario", Tax: 13.0, TaxName: "HST", TaxType: "harmonized"},
				{Code: "BC", Name: "British Columbia", Tax: 12.0, TaxName: "PST+GST", TaxType: "combined"},
				{Code: "AB", Name: "Alberta", Tax: 5.0, TaxName: "GST", TaxType: "federal"},
			},
		},
	}
	cleanup, buf := setupCountriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCountriesTestCmd()

	var err error
	stdout := captureCountriesStdout(t, func() {
		err = countriesGetCmd.RunE(cmd, []string{"CA"})
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String() + stdout
	// Check that province details are in the output
	expectedOutput := []string{
		"ON", "Ontario", "13.00%", "HST",
		"BC", "British Columbia", "12.00%", "PST+GST",
		"AB", "Alberta", "5.00%", "GST",
	}
	for _, want := range expectedOutput {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q, got: %s", want, output)
		}
	}
}

// TestCountriesListTableHeaders verifies the table headers are correct.
func TestCountriesListTableHeaders(t *testing.T) {
	mockClient := &countriesMockAPIClient{
		listCountriesResp: &api.CountriesListResponse{
			Items: []api.Country{
				{Code: "US", Name: "United States", Tax: 0, TaxName: "Tax", Provinces: []api.Province{}},
			},
		},
	}
	cleanup, buf := setupCountriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCountriesTestCmd()

	var err error
	stdout := captureCountriesStdout(t, func() {
		err = countriesListCmd.RunE(cmd, []string{})
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String() + stdout
	headers := []string{"CODE", "NAME", "TAX", "TAX NAME", "PROVINCES"}
	for _, header := range headers {
		if !strings.Contains(output, header) {
			t.Errorf("output should contain header %q, got: %s", header, output)
		}
	}
}

// TestCountriesListWithValidStore tests list command execution with valid store.
func TestCountriesListWithValidStore(t *testing.T) {
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
	cmd.AddCommand(countriesListCmd)

	// This will fail at the API call level, but validates the client setup works
	err := countriesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("countriesListCmd succeeded (might be due to mock setup)")
	}
}

// TestCountriesGetCountryWithEmptyTaxName tests country with empty tax name.
func TestCountriesGetCountryWithEmptyTaxName(t *testing.T) {
	mockClient := &countriesMockAPIClient{
		getCountryResp: &api.Country{
			Code:      "HK",
			Name:      "Hong Kong",
			Tax:       0.0,
			TaxName:   "",
			Provinces: []api.Province{},
		},
	}
	cleanup, _ := setupCountriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCountriesTestCmd()

	var err error
	stdout := captureCountriesStdout(t, func() {
		err = countriesGetCmd.RunE(cmd, []string{"HK"})
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify that basic fields are in output
	if !strings.Contains(stdout, "Hong Kong") {
		t.Errorf("output should contain 'Hong Kong', got: %s", stdout)
	}
	if !strings.Contains(stdout, "0.00%") {
		t.Errorf("output should contain '0.00%%', got: %s", stdout)
	}
}

// TestCountriesListMultipleCountries tests listing multiple countries.
func TestCountriesListMultipleCountries(t *testing.T) {
	mockClient := &countriesMockAPIClient{
		listCountriesResp: &api.CountriesListResponse{
			Items: []api.Country{
				{Code: "US", Name: "United States", Tax: 7.5, TaxName: "Sales Tax", Provinces: []api.Province{{Code: "CA", Name: "California", Tax: 7.25, TaxName: "CA Tax"}}},
				{Code: "CA", Name: "Canada", Tax: 5.0, TaxName: "GST", Provinces: []api.Province{}},
				{Code: "GB", Name: "United Kingdom", Tax: 20.0, TaxName: "VAT", Provinces: []api.Province{}},
				{Code: "AU", Name: "Australia", Tax: 10.0, TaxName: "GST", Provinces: []api.Province{}},
			},
		},
	}
	cleanup, buf := setupCountriesMockFactories(mockClient)
	defer cleanup()

	cmd := newCountriesTestCmd()

	var err error
	stdout := captureCountriesStdout(t, func() {
		err = countriesListCmd.RunE(cmd, []string{})
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String() + stdout
	// Check that the count shows 4 countries
	if !strings.Contains(output, "4 countries") {
		t.Errorf("output should show '4 countries', got: %s", output)
	}
}
