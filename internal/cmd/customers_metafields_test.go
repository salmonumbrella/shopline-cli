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

type customersMetafieldsMockClient struct {
	api.MockClient

	listResp json.RawMessage
	listErr  error

	createResp json.RawMessage
	createErr  error

	deleteErr error

	bulkDeleteErr error

	lastBody json.RawMessage
}

func (m *customersMetafieldsMockClient) ListCustomerMetafields(ctx context.Context, customerID string) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func (m *customersMetafieldsMockClient) CreateCustomerMetafield(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.createResp, m.createErr
}

func (m *customersMetafieldsMockClient) DeleteCustomerMetafield(ctx context.Context, customerID, metafieldID string) error {
	return m.deleteErr
}

func (m *customersMetafieldsMockClient) BulkDeleteCustomerMetafields(ctx context.Context, customerID string, body any) error {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkDeleteErr
}

func setupCustomersMetafieldsTest(t *testing.T, mockClient api.APIClient, store *mockStore) (cleanup func(), buf *bytes.Buffer) {
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

func TestCustomersMetafieldsCommandWired(t *testing.T) {
	names := map[string]bool{}
	for _, c := range customersCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"metafields", "app-metafields"} {
		if !names[want] {
			t.Fatalf("customers subcommand %q not registered", want)
		}
	}
}

func TestCustomersMetafieldsListRunE_Success(t *testing.T) {
	mockClient := &customersMetafieldsMockClient{
		listResp: json.RawMessage(`{"items":[{"id":"mf_1"}]}`),
	}
	cleanup, buf := setupCustomersMetafieldsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")

	if err := customersMetafieldsListCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains([]byte(got), []byte(`"id": "mf_1"`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestCustomersMetafieldsCreateRunE_PassesBody(t *testing.T) {
	mockClient := &customersMetafieldsMockClient{
		createResp: json.RawMessage(`{"id":"mf_created"}`),
	}
	cleanup, buf := setupCustomersMetafieldsTest(t, mockClient, defaultMockStore())
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

	if err := customersMetafieldsCreateCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if string(bytes.TrimSpace(mockClient.lastBody)) != `{"namespace":"ns","key":"k","value":"v"}` {
		t.Fatalf("unexpected body: %s", string(mockClient.lastBody))
	}
	if got := buf.String(); !bytes.Contains([]byte(got), []byte(`"id": "mf_created"`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestCustomersMetafieldsDeleteRunE_YesSkipsPrompt(t *testing.T) {
	mockClient := &customersMetafieldsMockClient{}
	cleanup, _ := setupCustomersMetafieldsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("yes", true, "")

	if err := customersMetafieldsDeleteCmd.RunE(cmd, []string{"cust_1", "mf_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
}

func TestCustomersMetafieldsBulkDeleteRunE_Error(t *testing.T) {
	mockClient := &customersMetafieldsMockClient{
		bulkDeleteErr: errors.New("boom"),
	}
	cleanup, _ := setupCustomersMetafieldsTest(t, mockClient, defaultMockStore())
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

	err := customersMetafieldsBulkDeleteCmd.RunE(cmd, []string{"cust_1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mockClient.bulkDeleteErr) && !bytes.Contains([]byte(err.Error()), []byte("boom")) {
		t.Fatalf("unexpected error: %v", err)
	}
}
