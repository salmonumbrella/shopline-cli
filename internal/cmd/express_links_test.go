package cmd

import (
	"os"
	"testing"
)

func TestExpressLinksCommandSetup(t *testing.T) {
	if expressLinksCmd.Use != "express-links" {
		t.Errorf("expected Use 'express-links', got %q", expressLinksCmd.Use)
	}
	if expressLinksCmd.Short != "Manage express links (via Admin API)" {
		t.Errorf("expected short description for express-links, got %q", expressLinksCmd.Short)
	}
}

func TestExpressLinksCreateCommandSetup(t *testing.T) {
	if expressLinksCreateCmd.Use != "create" {
		t.Errorf("expected Use 'create', got %q", expressLinksCreateCmd.Use)
	}
	if expressLinksCreateCmd.Short != "Create an express link (raw JSON body)" {
		t.Errorf("expected create short description, got %q", expressLinksCreateCmd.Short)
	}
	hasNew := false
	hasGenerate := false
	for _, a := range expressLinksCreateCmd.Aliases {
		if a == "new" {
			hasNew = true
		}
		if a == "generate" {
			hasGenerate = true
		}
	}
	if !hasNew || !hasGenerate {
		t.Errorf("expected aliases to include 'new' and 'generate', got %v", expressLinksCreateCmd.Aliases)
	}
}

func TestExpressLinksCreateFlags(t *testing.T) {
	flags := []string{"body", "body-file", "dry-run"}
	for _, name := range flags {
		flag := expressLinksCreateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("flag %q not found", name)
		}
	}
}

func TestExpressLinksCreateRunE_NoAdminToken(t *testing.T) {
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
	cmd.Flags().String("body", `{"products":[{"_id":"prod_1","variation_id":"var_1"}],"user_id":"usr_1","campaign":{"_id":"camp_1"}}`, "")
	cmd.Flags().String("body-file", "", "")

	err := expressLinksCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExpressLinksCreateRunE_DryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("body", `{"products":[{"_id":"prod_1","variation_id":"var_1"}],"user_id":"usr_1","campaign":{"_id":"camp_1"}}`, "")
	cmd.Flags().String("body-file", "", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := expressLinksCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("expected dry-run to bypass API call, got error: %v", err)
	}
}
