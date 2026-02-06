package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

type posPurchaseOrdersMockClient struct {
	api.MockClient

	listResp json.RawMessage
	listErr  error

	getResp json.RawMessage
	getErr  error

	createResp json.RawMessage
	createErr  error

	updateResp json.RawMessage
	updateErr  error

	bulkDeleteResp json.RawMessage
	bulkDeleteErr  error

	childResp json.RawMessage
	childErr  error
}

func (m *posPurchaseOrdersMockClient) ListPOSPurchaseOrders(ctx context.Context, opts *api.POSPurchaseOrdersListOptions) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func (m *posPurchaseOrdersMockClient) GetPOSPurchaseOrder(ctx context.Context, id string) (json.RawMessage, error) {
	return m.getResp, m.getErr
}

func (m *posPurchaseOrdersMockClient) CreatePOSPurchaseOrder(ctx context.Context, body any) (json.RawMessage, error) {
	return m.createResp, m.createErr
}

func (m *posPurchaseOrdersMockClient) UpdatePOSPurchaseOrder(ctx context.Context, id string, body any) (json.RawMessage, error) {
	return m.updateResp, m.updateErr
}

func (m *posPurchaseOrdersMockClient) BulkDeletePOSPurchaseOrders(ctx context.Context, body any) (json.RawMessage, error) {
	return m.bulkDeleteResp, m.bulkDeleteErr
}

func (m *posPurchaseOrdersMockClient) CreatePOSPurchaseOrderChild(ctx context.Context, id string, body any) (json.RawMessage, error) {
	return m.childResp, m.childErr
}

func setupPOSPurchaseOrdersMockFactories(mockClient *posPurchaseOrdersMockClient) (func(), *bytes.Buffer) {
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

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}, buf
}

func newPOSPurchaseOrdersTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	addJSONBodyFlags(cmd)
	return cmd
}

func TestPOSPurchaseOrdersCommandsExist(t *testing.T) {
	if posCmd.Use != "pos" {
		t.Fatalf("expected pos use, got %q", posCmd.Use)
	}
	if posPurchaseOrdersCmd.Use != "purchase-orders" {
		t.Fatalf("expected purchase-orders use, got %q", posPurchaseOrdersCmd.Use)
	}
}

func TestPOSPurchaseOrdersRunE(t *testing.T) {
	mockClient := &posPurchaseOrdersMockClient{
		listResp:       json.RawMessage(`{"items":[]}`),
		getResp:        json.RawMessage(`{"id":"po_1"}`),
		createResp:     json.RawMessage(`{"id":"po_new"}`),
		updateResp:     json.RawMessage(`{"ok":true}`),
		bulkDeleteResp: json.RawMessage(`{"ok":true}`),
		childResp:      json.RawMessage(`{"id":"po_child"}`),
	}
	cleanup, buf := setupPOSPurchaseOrdersMockFactories(mockClient)
	defer cleanup()

	cmd := newPOSPurchaseOrdersTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"ok":true}`)

	if err := posPurchaseOrdersListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("list unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"items\"") {
		t.Fatalf("expected items in list output, got %q", buf.String())
	}

	buf.Reset()
	if err := posPurchaseOrdersGetCmd.RunE(cmd, []string{"po_1"}); err != nil {
		t.Fatalf("get unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"po_1\"") {
		t.Fatalf("expected po_1 in get output, got %q", buf.String())
	}

	buf.Reset()
	if err := posPurchaseOrdersCreateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("create unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"po_new\"") {
		t.Fatalf("expected po_new in create output, got %q", buf.String())
	}

	buf.Reset()
	if err := posPurchaseOrdersUpdateCmd.RunE(cmd, []string{"po_1"}); err != nil {
		t.Fatalf("update unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"ok\"") {
		t.Fatalf("expected ok in update output, got %q", buf.String())
	}

	buf.Reset()
	if err := posPurchaseOrdersBulkDeleteCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("bulk-delete unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"ok\"") {
		t.Fatalf("expected ok in bulk-delete output, got %q", buf.String())
	}

	buf.Reset()
	if err := posPurchaseOrdersCreateChildCmd.RunE(cmd, []string{"po_1"}); err != nil {
		t.Fatalf("create-child unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"po_child\"") {
		t.Fatalf("expected po_child in create-child output, got %q", buf.String())
	}
}
