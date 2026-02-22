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

type productReviewCommentsMockClient struct {
	api.MockClient

	listResp json.RawMessage
	listErr  error

	getResp json.RawMessage
	getErr  error

	createResp json.RawMessage
	createErr  error

	updateResp json.RawMessage
	updateErr  error

	deleteResp json.RawMessage
	deleteErr  error

	bulkCreateResp json.RawMessage
	bulkCreateErr  error

	bulkUpdateResp json.RawMessage
	bulkUpdateErr  error

	bulkDeleteResp json.RawMessage
	bulkDeleteErr  error
}

func (m *productReviewCommentsMockClient) ListProductReviewComments(ctx context.Context, opts *api.ProductReviewCommentsListOptions) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func (m *productReviewCommentsMockClient) GetProductReviewComment(ctx context.Context, id string) (json.RawMessage, error) {
	return m.getResp, m.getErr
}

func (m *productReviewCommentsMockClient) CreateProductReviewComment(ctx context.Context, body any) (json.RawMessage, error) {
	return m.createResp, m.createErr
}

func (m *productReviewCommentsMockClient) UpdateProductReviewComment(ctx context.Context, id string, body any) (json.RawMessage, error) {
	return m.updateResp, m.updateErr
}

func (m *productReviewCommentsMockClient) DeleteProductReviewComment(ctx context.Context, id string) (json.RawMessage, error) {
	return m.deleteResp, m.deleteErr
}

func (m *productReviewCommentsMockClient) BulkCreateProductReviewComments(ctx context.Context, body any) (json.RawMessage, error) {
	return m.bulkCreateResp, m.bulkCreateErr
}

func (m *productReviewCommentsMockClient) BulkUpdateProductReviewComments(ctx context.Context, body any) (json.RawMessage, error) {
	return m.bulkUpdateResp, m.bulkUpdateErr
}

func (m *productReviewCommentsMockClient) BulkDeleteProductReviewComments(ctx context.Context, body any) (json.RawMessage, error) {
	return m.bulkDeleteResp, m.bulkDeleteErr
}

func setupProductReviewCommentsMockFactories(mockClient *productReviewCommentsMockClient) (func(), *bytes.Buffer) {
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

func newProductReviewCommentsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	addJSONBodyFlags(cmd)
	return cmd
}

func TestProductReviewCommentsCmdStructure(t *testing.T) {
	if productReviewCommentsCmd.Use != "product-review-comments" {
		t.Fatalf("expected product-review-comments use, got %q", productReviewCommentsCmd.Use)
	}
}

func TestProductReviewCommentsRunE(t *testing.T) {
	mockClient := &productReviewCommentsMockClient{
		listResp:       json.RawMessage(`{"items":[]}`),
		getResp:        json.RawMessage(`{"id":"cmt_1"}`),
		createResp:     json.RawMessage(`{"id":"cmt_new"}`),
		updateResp:     json.RawMessage(`{"updated":true}`),
		deleteResp:     json.RawMessage(`{"deleted":true}`),
		bulkCreateResp: json.RawMessage(`{"ok":true}`),
		bulkUpdateResp: json.RawMessage(`{"ok":true}`),
		bulkDeleteResp: json.RawMessage(`{"ok":true}`),
	}
	cleanup, buf := setupProductReviewCommentsMockFactories(mockClient)
	defer cleanup()

	cmd := newProductReviewCommentsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"ok":true}`)
	_ = cmd.Flags().Set("yes", "true")

	if err := productReviewCommentsListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("list unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"items\"") {
		t.Fatalf("expected items in list output, got %q", buf.String())
	}

	buf.Reset()
	if err := productReviewCommentsGetCmd.RunE(cmd, []string{"cmt_1"}); err != nil {
		t.Fatalf("get unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"cmt_1\"") {
		t.Fatalf("expected cmt_1 in get output, got %q", buf.String())
	}

	buf.Reset()
	if err := productReviewCommentsCreateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("create unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"cmt_new\"") {
		t.Fatalf("expected cmt_new in create output, got %q", buf.String())
	}

	buf.Reset()
	if err := productReviewCommentsUpdateCmd.RunE(cmd, []string{"cmt_1"}); err != nil {
		t.Fatalf("update unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"updated\"") {
		t.Fatalf("expected updated in update output, got %q", buf.String())
	}

	buf.Reset()
	if err := productReviewCommentsDeleteCmd.RunE(cmd, []string{"cmt_1"}); err != nil {
		t.Fatalf("delete unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"deleted\"") {
		t.Fatalf("expected deleted in delete output, got %q", buf.String())
	}

	buf.Reset()
	if err := productReviewCommentsBulkCreateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("bulk-create unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"ok\"") {
		t.Fatalf("expected ok in bulk-create output, got %q", buf.String())
	}

	buf.Reset()
	if err := productReviewCommentsBulkUpdateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("bulk-update unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"ok\"") {
		t.Fatalf("expected ok in bulk-update output, got %q", buf.String())
	}

	buf.Reset()
	if err := productReviewCommentsBulkDeleteCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("bulk-delete unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"ok\"") {
		t.Fatalf("expected ok in bulk-delete output, got %q", buf.String())
	}
}
