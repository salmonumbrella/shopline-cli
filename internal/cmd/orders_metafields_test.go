package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type ordersMetafieldsMockClient struct {
	api.MockClient

	listResp   json.RawMessage
	listErr    error
	createResp json.RawMessage
	createErr  error

	deleteErr error

	bulkDeleteErr error

	lastBody json.RawMessage
}

func (m *ordersMetafieldsMockClient) ListOrderMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func (m *ordersMetafieldsMockClient) CreateOrderMetafield(ctx context.Context, orderID string, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.createResp, m.createErr
}

func (m *ordersMetafieldsMockClient) DeleteOrderMetafield(ctx context.Context, orderID, metafieldID string) error {
	return m.deleteErr
}

func (m *ordersMetafieldsMockClient) BulkDeleteOrderMetafields(ctx context.Context, orderID string, body any) error {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkDeleteErr
}

func setupOrdersMetafieldsTest(t *testing.T, mockClient api.APIClient, store *mockStore) (cleanup func(), buf *bytes.Buffer) {
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

func TestOrdersMetafieldsCommandWired(t *testing.T) {
	// Ensure the init() wiring happened and the commands are discoverable.
	names := map[string]bool{}
	for _, c := range ordersCmd.Commands() {
		names[c.Name()] = true
	}

	for _, want := range []string{"metafields", "app-metafields", "item-metafields", "item-app-metafields"} {
		if !names[want] {
			t.Fatalf("orders subcommand %q not registered", want)
		}
	}
}

func TestOrdersMetafieldsListRunE_Success(t *testing.T) {
	mockClient := &ordersMetafieldsMockClient{
		listResp: json.RawMessage(`{"items":[{"id":"mf_1"}]}`),
	}
	cleanup, buf := setupOrdersMetafieldsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")

	if err := ordersMetafieldsListCmd.RunE(cmd, []string{"ord_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains([]byte(got), []byte(`"id": "mf_1"`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestOrdersMetafieldsCreateRunE_InvalidJSON(t *testing.T) {
	mockClient := &ordersMetafieldsMockClient{}
	cleanup, _ := setupOrdersMetafieldsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("body", "{not valid json", "")
	cmd.Flags().String("body-file", "", "")

	if err := ordersMetafieldsCreateCmd.RunE(cmd, []string{"ord_1"}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOrdersMetafieldsCreateRunE_PassesBody(t *testing.T) {
	mockClient := &ordersMetafieldsMockClient{
		createResp: json.RawMessage(`{"id":"mf_created"}`),
	}
	cleanup, buf := setupOrdersMetafieldsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("body", `{"namespace":"ns","key":"k","value":"v"}`, "")
	cmd.Flags().String("body-file", "", "")

	if err := ordersMetafieldsCreateCmd.RunE(cmd, []string{"ord_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if string(bytes.TrimSpace(mockClient.lastBody)) != `{"namespace":"ns","key":"k","value":"v"}` {
		t.Fatalf("unexpected body: %s", string(mockClient.lastBody))
	}
	if got := buf.String(); !bytes.Contains([]byte(got), []byte(`"id": "mf_created"`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestOrdersMetafieldsDeleteRunE_YesSkipsPrompt(t *testing.T) {
	mockClient := &ordersMetafieldsMockClient{}
	cleanup, _ := setupOrdersMetafieldsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("yes", true, "")

	if err := ordersMetafieldsDeleteCmd.RunE(cmd, []string{"ord_1", "mf_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
}

func TestOrdersMetafieldsBulkDeleteRunE_Error(t *testing.T) {
	mockClient := &ordersMetafieldsMockClient{
		bulkDeleteErr: errors.New("boom"),
	}
	cleanup, _ := setupOrdersMetafieldsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("body", `{"ids":["mf_1"]}`, "")
	cmd.Flags().String("body-file", "", "")

	err := ordersMetafieldsBulkDeleteCmd.RunE(cmd, []string{"ord_1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mockClient.bulkDeleteErr) && !bytes.Contains([]byte(err.Error()), []byte("boom")) {
		t.Fatalf("unexpected error: %v", err)
	}
}
