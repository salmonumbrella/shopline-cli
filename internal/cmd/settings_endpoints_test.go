package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type settingsEndpointsMockClient struct {
	api.MockClient

	checkoutResp json.RawMessage
	checkoutErr  error

	updateDomainsResp json.RawMessage
	updateDomainsErr  error
	lastBody          json.RawMessage
}

func (m *settingsEndpointsMockClient) GetSettingsCheckout(ctx context.Context) (json.RawMessage, error) {
	return m.checkoutResp, m.checkoutErr
}

func (m *settingsEndpointsMockClient) UpdateSettingsDomains(ctx context.Context, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.updateDomainsResp, m.updateDomainsErr
}

func setupSettingsEndpointsTest(t *testing.T, mockClient api.APIClient, store *mockStore) (cleanup func(), buf *bytes.Buffer) {
	t.Helper()
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	buf = new(bytes.Buffer)
	formatterWriter = buf

	secretsStoreFactory = func() (CredentialStore, error) {
		if store.err != nil {
			return nil, store.err
		}
		return store, nil
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}, buf
}

func newSettingsEndpointsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	return cmd
}

func TestSettingsEndpointsWired(t *testing.T) {
	names := map[string]bool{}
	for _, c := range settingsCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"checkout", "domains", "layouts", "theme"} {
		if !names[want] {
			t.Fatalf("settings subcommand %q not registered", want)
		}
	}
}

func TestSettingsCheckoutGetRunEJSON(t *testing.T) {
	mockClient := &settingsEndpointsMockClient{
		checkoutResp: json.RawMessage(`{"ok":true}`),
	}
	cleanup, buf := setupSettingsEndpointsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := newSettingsEndpointsTestCmd()

	if err := settingsCheckoutGetCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte(`"ok": true`)) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestSettingsDomainsUpdateRunE_PassesBody(t *testing.T) {
	mockClient := &settingsEndpointsMockClient{
		updateDomainsResp: json.RawMessage(`{"updated":true}`),
	}
	cleanup, buf := setupSettingsEndpointsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := newSettingsEndpointsTestCmd()
	_ = cmd.Flags().Set("body", `{"domains":[{"host":"example.com"}]}`)

	if err := settingsDomainsUpdateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if string(bytes.TrimSpace(mockClient.lastBody)) != `{"domains":[{"host":"example.com"}]}` {
		t.Fatalf("unexpected body: %s", string(mockClient.lastBody))
	}
	if out := buf.String(); !bytes.Contains([]byte(out), []byte(`"updated": true`)) {
		t.Fatalf("unexpected output: %s", out)
	}
}
