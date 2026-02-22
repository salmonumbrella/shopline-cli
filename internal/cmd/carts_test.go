package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type cartsMockClient struct {
	api.MockClient

	exchangeResp json.RawMessage
	exchangeErr  error
	lastBody     json.RawMessage

	prepareResp json.RawMessage
	prepareErr  error
	prepareBody any

	addResp json.RawMessage
	addErr  error

	updateResp json.RawMessage
	updateErr  error

	deleteResp json.RawMessage
	deleteErr  error

	listMetafieldsResp json.RawMessage
	listMetafieldsErr  error

	bulkCreateErr error
	bulkUpdateErr error
	bulkDeleteErr error

	listAppMetafieldsResp json.RawMessage
	listAppMetafieldsErr  error

	bulkAppCreateErr error
	bulkAppUpdateErr error
	bulkAppDeleteErr error
}

func (m *cartsMockClient) ExchangeCart(ctx context.Context, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.exchangeResp, m.exchangeErr
}

func (m *cartsMockClient) PrepareCart(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	m.prepareBody = body
	return m.prepareResp, m.prepareErr
}

func (m *cartsMockClient) AddCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.addResp, m.addErr
}

func (m *cartsMockClient) UpdateCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.updateResp, m.updateErr
}

func (m *cartsMockClient) DeleteCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.deleteResp, m.deleteErr
}

func (m *cartsMockClient) ListCartItemMetafields(ctx context.Context, cartID string) (json.RawMessage, error) {
	return m.listMetafieldsResp, m.listMetafieldsErr
}

func (m *cartsMockClient) BulkCreateCartItemMetafields(ctx context.Context, cartID string, body any) error {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkCreateErr
}

func (m *cartsMockClient) BulkUpdateCartItemMetafields(ctx context.Context, cartID string, body any) error {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkUpdateErr
}

func (m *cartsMockClient) BulkDeleteCartItemMetafields(ctx context.Context, cartID string, body any) error {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkDeleteErr
}

func (m *cartsMockClient) ListCartItemAppMetafields(ctx context.Context, cartID string) (json.RawMessage, error) {
	return m.listAppMetafieldsResp, m.listAppMetafieldsErr
}

func (m *cartsMockClient) BulkCreateCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkAppCreateErr
}

func (m *cartsMockClient) BulkUpdateCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkAppUpdateErr
}

func (m *cartsMockClient) BulkDeleteCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkAppDeleteErr
}

func setupCartsTest(t *testing.T, mockClient api.APIClient, store *mockStore) (cleanup func(), buf *bytes.Buffer) {
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

func newCartsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.SetOut(formatterWriter)
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	return cmd
}

func TestCartsCommandWired(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "carts" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("root subcommand %q not registered", "carts")
	}
}

func TestCartsExchangeRunE_PassesBody(t *testing.T) {
	mockClient := &cartsMockClient{exchangeResp: json.RawMessage(`{"ok":true}`)}
	cleanup, buf := setupCartsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := newCartsTestCmd()
	_ = cmd.Flags().Set("body", `{"from":"USD","to":"CAD"}`)

	if err := cartsExchangeCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if string(bytes.TrimSpace(mockClient.lastBody)) != `{"from":"USD","to":"CAD"}` {
		t.Fatalf("unexpected body: %s", string(mockClient.lastBody))
	}
	if got := buf.String(); !bytes.Contains([]byte(got), []byte(`"ok": true`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestCartsPrepareRunE_AllowsNoBody(t *testing.T) {
	mockClient := &cartsMockClient{prepareResp: json.RawMessage(`{"prepared":true}`)}
	cleanup, buf := setupCartsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := newCartsTestCmd()

	if err := cartsPrepareCmd.RunE(cmd, []string{"cart_123"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if mockClient.prepareBody != nil {
		t.Fatalf("expected nil body, got %#v", mockClient.prepareBody)
	}
	if got := buf.String(); !bytes.Contains([]byte(got), []byte(`"prepared": true`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestCartsItemsDeleteRunE_YesSkipsPrompt(t *testing.T) {
	mockClient := &cartsMockClient{deleteResp: json.RawMessage(`{"deleted":true}`)}
	cleanup, _ := setupCartsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := newCartsTestCmd()
	_ = cmd.Flags().Set("body", `{"item_ids":["x"]}`)
	_ = cmd.Flags().Set("yes", "true")

	if err := cartsItemsDeleteCmd.RunE(cmd, []string{"cart_123"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if string(bytes.TrimSpace(mockClient.lastBody)) != `{"item_ids":["x"]}` {
		t.Fatalf("unexpected body: %s", string(mockClient.lastBody))
	}
}

func TestCartsMetafieldsBulkDeleteRunE_PrintsOK(t *testing.T) {
	mockClient := &cartsMockClient{}
	cleanup, buf := setupCartsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := newCartsTestCmd()
	_ = cmd.Flags().Set("output", "text")
	_ = cmd.Flags().Set("body", `{"ids":["mf_1"]}`)
	_ = cmd.Flags().Set("yes", "true")

	if err := cartsItemsMetafieldsBulkDeleteCmd.RunE(cmd, []string{"cart_123"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if got := buf.String(); !bytes.Contains([]byte(got), []byte("OK")) {
		t.Fatalf("expected OK output, got: %s", got)
	}
}
