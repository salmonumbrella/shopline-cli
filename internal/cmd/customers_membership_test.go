package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type customersMembershipMockClient struct {
	api.MockClient

	infoResp json.RawMessage
	infoErr  error

	logsResp json.RawMessage
	logsErr  error
}

func (m *customersMembershipMockClient) GetCustomersMembershipInfo(ctx context.Context) (json.RawMessage, error) {
	return m.infoResp, m.infoErr
}

func (m *customersMembershipMockClient) GetCustomerMembershipTierActionLogs(ctx context.Context, customerID string) (json.RawMessage, error) {
	return m.logsResp, m.logsErr
}

func TestCustomersMembershipInfoRunE_Success(t *testing.T) {
	mockClient := &customersMembershipMockClient{
		infoResp: json.RawMessage(`{"items":[{"customer_id":"cust_1"}]}`),
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

	if err := customersMembershipInfoCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains(buf.Bytes(), []byte(`"customer_id"`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestCustomersMembershipTierActionLogsRunE_Success(t *testing.T) {
	mockClient := &customersMembershipMockClient{
		logsResp: json.RawMessage(`{"items":[{"id":"log_1"}]}`),
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

	if err := customersMembershipTierActionLogsCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains(buf.Bytes(), []byte(`"log_1"`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}
