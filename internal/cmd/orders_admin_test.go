package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// TestOrdersCommentCommandSetup verifies the comment command initialization.
func TestOrdersCommentCommandSetup(t *testing.T) {
	if ordersCommentCmd.Use != "comment <order-id>" {
		t.Errorf("expected Use 'comment <order-id>', got %q", ordersCommentCmd.Use)
	}
	if ordersCommentCmd.Short != "Add a comment to an order (via Admin API)" {
		t.Errorf("expected Short 'Add a comment to an order (via Admin API)', got %q", ordersCommentCmd.Short)
	}
	if ordersCommentCmd.Args == nil {
		t.Fatal("expected Args to be set")
	}
	// Verify ExactArgs(1): 0 args should fail, 1 should pass, 2 should fail
	if err := cobra.ExactArgs(1)(nil, []string{}); err == nil {
		t.Error("expected error for 0 args")
	}
	if err := cobra.ExactArgs(1)(nil, []string{"id"}); err != nil {
		t.Errorf("expected no error for 1 arg, got: %v", err)
	}
	if err := cobra.ExactArgs(1)(nil, []string{"a", "b"}); err == nil {
		t.Error("expected error for 2 args")
	}
}

// TestOrdersAdminRefundCommandSetup verifies the admin-refund command initialization.
func TestOrdersAdminRefundCommandSetup(t *testing.T) {
	if ordersAdminRefundCmd.Use != "admin-refund <order-id>" {
		t.Errorf("expected Use 'admin-refund <order-id>', got %q", ordersAdminRefundCmd.Use)
	}
	if ordersAdminRefundCmd.Short != "Issue an admin refund for an order (via Admin API)" {
		t.Errorf("expected Short 'Issue an admin refund for an order (via Admin API)', got %q", ordersAdminRefundCmd.Short)
	}
	if ordersAdminRefundCmd.Args == nil {
		t.Fatal("expected Args to be set")
	}

	// Verify required flags are registered
	requiredFlags := []string{"performer-id", "amount", "payment-updated-at"}
	for _, name := range requiredFlags {
		flag := ordersAdminRefundCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("required flag %q not found", name)
			continue
		}
		// Check that the flag has the required annotation
		annotations := flag.Annotations
		if annotations == nil {
			t.Errorf("flag %q has no annotations (expected required)", name)
			continue
		}
		if _, ok := annotations[cobra.BashCompOneRequiredFlag]; !ok {
			t.Errorf("flag %q is not marked as required", name)
		}
	}
}

// TestOrdersReceiptReissueCommandSetup verifies the receipt-reissue command initialization.
func TestOrdersReceiptReissueCommandSetup(t *testing.T) {
	if ordersReceiptReissueCmd.Use != "receipt-reissue <order-id>" {
		t.Errorf("expected Use 'receipt-reissue <order-id>', got %q", ordersReceiptReissueCmd.Use)
	}
	if ordersReceiptReissueCmd.Short != "Reissue a receipt for an order (via Admin API)" {
		t.Errorf("expected Short 'Reissue a receipt for an order (via Admin API)', got %q", ordersReceiptReissueCmd.Short)
	}
	if ordersReceiptReissueCmd.Args == nil {
		t.Fatal("expected Args to be set")
	}
}

// TestOrdersCommentFlags verifies comment command flags exist with correct defaults.
func TestOrdersCommentFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"text", ""},
		{"private", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := ordersCommentCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

// TestOrdersAdminRefundFlags verifies admin-refund command flags exist with correct defaults.
func TestOrdersAdminRefundFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"performer-id", ""},
		{"amount", "0"},
		{"payment-updated-at", ""},
		{"remark", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := ordersAdminRefundCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

// TestOrdersCommentRunE_NoAdminToken verifies error when no admin token is set.
func TestOrdersCommentRunE_NoAdminToken(t *testing.T) {
	t.Setenv("SHOPLINE_ADMIN_BASE_URL", "https://test.example.com")
	origToken := os.Getenv("SHOPLINE_ADMIN_TOKEN")
	origMerchant := os.Getenv("SHOPLINE_ADMIN_MERCHANT_ID")
	defer func() {
		_ = os.Setenv("SHOPLINE_ADMIN_TOKEN", origToken)
		_ = os.Setenv("SHOPLINE_ADMIN_MERCHANT_ID", origMerchant)
	}()
	_ = os.Unsetenv("SHOPLINE_ADMIN_TOKEN")
	_ = os.Unsetenv("SHOPLINE_ADMIN_MERCHANT_ID")

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("text", "hello", "")
	cmd.Flags().Bool("private", false, "")
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")

	err := ordersCommentCmd.RunE(cmd, []string{"ord_123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestOrdersAdminRefundRunE_NoAdminToken verifies error when no admin token is set.
func TestOrdersAdminRefundRunE_NoAdminToken(t *testing.T) {
	t.Setenv("SHOPLINE_ADMIN_BASE_URL", "https://test.example.com")
	origToken := os.Getenv("SHOPLINE_ADMIN_TOKEN")
	origMerchant := os.Getenv("SHOPLINE_ADMIN_MERCHANT_ID")
	defer func() {
		_ = os.Setenv("SHOPLINE_ADMIN_TOKEN", origToken)
		_ = os.Setenv("SHOPLINE_ADMIN_MERCHANT_ID", origMerchant)
	}()
	_ = os.Unsetenv("SHOPLINE_ADMIN_TOKEN")
	_ = os.Unsetenv("SHOPLINE_ADMIN_MERCHANT_ID")

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("performer-id", "perf_1", "")
	cmd.Flags().Int("amount", 100, "")
	cmd.Flags().String("payment-updated-at", "2024-01-01", "")
	cmd.Flags().String("remark", "", "")
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")

	err := ordersAdminRefundCmd.RunE(cmd, []string{"ord_123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestOrdersCommentsCommandSetup verifies the comments list command initialization.
func TestOrdersCommentsCommandSetup(t *testing.T) {
	if ordersCommentsCmd.Use != "comments <order-id>" {
		t.Errorf("expected Use 'comments <order-id>', got %q", ordersCommentsCmd.Use)
	}
	if ordersCommentsCmd.Short != "List comments on an order (via Admin API)" {
		t.Errorf("expected Short 'List comments on an order (via Admin API)', got %q", ordersCommentsCmd.Short)
	}
	if ordersCommentsCmd.Args == nil {
		t.Fatal("expected Args to be set")
	}
	// Verify alias
	found := false
	for _, a := range ordersCommentsCmd.Aliases {
		if a == "cmts" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected alias 'cmts', got %v", ordersCommentsCmd.Aliases)
	}
}

// TestOrdersCommentsRunE_NoAdminToken verifies error when no admin token is set.
func TestOrdersCommentsRunE_NoAdminToken(t *testing.T) {
	t.Setenv("SHOPLINE_ADMIN_BASE_URL", "https://test.example.com")
	origToken := os.Getenv("SHOPLINE_ADMIN_TOKEN")
	origMerchant := os.Getenv("SHOPLINE_ADMIN_MERCHANT_ID")
	defer func() {
		_ = os.Setenv("SHOPLINE_ADMIN_TOKEN", origToken)
		_ = os.Setenv("SHOPLINE_ADMIN_MERCHANT_ID", origMerchant)
	}()
	_ = os.Unsetenv("SHOPLINE_ADMIN_TOKEN")
	_ = os.Unsetenv("SHOPLINE_ADMIN_MERCHANT_ID")

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")

	err := ordersCommentsCmd.RunE(cmd, []string{"ord_123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error: %v", err)
	}
}
