package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type customersStoreCreditsMockClient struct {
	api.MockClient

	getResp json.RawMessage
	getErr  error

	updateResp json.RawMessage
	updateErr  error

	lastBody json.RawMessage
}

func (m *customersStoreCreditsMockClient) GetCustomerStoreCredits(ctx context.Context, customerID string) (json.RawMessage, error) {
	return m.getResp, m.getErr
}

func (m *customersStoreCreditsMockClient) CreateCustomerStoreCredits(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.updateResp, m.updateErr
}

func TestCustomersStoreCreditsCommandWired(t *testing.T) {
	names := map[string]bool{}
	for _, c := range customersCmd.Commands() {
		names[c.Name()] = true
	}
	if !names["store-credits"] {
		t.Fatalf("customers subcommand %q not registered", "store-credits")
	}
}

func TestCustomersStoreCreditsListRunE_Success(t *testing.T) {
	mockClient := &customersStoreCreditsMockClient{
		getResp: json.RawMessage(`{"items":[{"id":"sc_1"}]}`),
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

	if err := customersStoreCreditsListCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains([]byte(got), []byte(`"id": "sc_1"`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestCustomersStoreCreditsUpdateRunE_BuildsBodyFromFlags(t *testing.T) {
	mockClient := &customersStoreCreditsMockClient{
		updateResp: json.RawMessage(`{"ok":true}`),
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
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().Float64("amount", 0, "")
	cmd.Flags().String("reason", "", "")
	_ = cmd.Flags().Set("amount", "10.5")
	_ = cmd.Flags().Set("reason", "test")

	if err := customersStoreCreditsUpdateCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(mockClient.lastBody, &got); err != nil {
		t.Fatalf("expected JSON body, got err: %v", err)
	}
	if got["amount"] != float64(10.5) {
		t.Fatalf("unexpected amount: %v", got["amount"])
	}
	if got["reason"] != "test" {
		t.Fatalf("unexpected reason: %v", got["reason"])
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"ok": true`)) {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}
