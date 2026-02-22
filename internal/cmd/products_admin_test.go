package cmd

import (
	"os"
	"testing"
)

func TestProductsHideCommandSetup(t *testing.T) {
	if productsHideCmd.Use != "hide <product-id>" {
		t.Errorf("expected Use 'hide <product-id>', got %q", productsHideCmd.Use)
	}
	if productsHideCmd.Short != "Hide a product (via Admin API)" {
		t.Errorf("expected Short 'Hide a product (via Admin API)', got %q", productsHideCmd.Short)
	}
	if productsHideCmd.Args == nil {
		t.Error("expected Args to be set")
	}
}

func TestProductsPublishCommandSetup(t *testing.T) {
	if productsPublishCmd.Use != "publish <product-id>" {
		t.Errorf("expected Use 'publish <product-id>', got %q", productsPublishCmd.Use)
	}
	if productsPublishCmd.Short != "Publish a product (via Admin API)" {
		t.Errorf("expected Short 'Publish a product (via Admin API)', got %q", productsPublishCmd.Short)
	}
	if productsPublishCmd.Args == nil {
		t.Error("expected Args to be set")
	}
}

func TestProductsUnpublishCommandSetup(t *testing.T) {
	if productsUnpublishCmd.Use != "unpublish <product-id>" {
		t.Errorf("expected Use 'unpublish <product-id>', got %q", productsUnpublishCmd.Use)
	}
	if productsUnpublishCmd.Short != "Unpublish a product (via Admin API)" {
		t.Errorf("expected Short 'Unpublish a product (via Admin API)', got %q", productsUnpublishCmd.Short)
	}
	if productsUnpublishCmd.Args == nil {
		t.Error("expected Args to be set")
	}
}

func TestProductsHideRunE_NoAdminToken(t *testing.T) {
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

	err := productsHideCmd.RunE(cmd, []string{"prod_123"})
	if err == nil {
		t.Fatal("Expected error when no admin token, got nil")
	}
	if err.Error() != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error message: %v", err)
	}
}
