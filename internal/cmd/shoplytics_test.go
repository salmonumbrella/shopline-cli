package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestShoplyticsCommandSetup(t *testing.T) {
	if shoplyticsCmd.Use != "shoplytics" {
		t.Errorf("expected Use 'shoplytics', got %q", shoplyticsCmd.Use)
	}
	if shoplyticsCmd.Short != "Access Shoplytics analytics (via Admin API)" {
		t.Errorf("expected Shoplytics short description, got %q", shoplyticsCmd.Short)
	}
}

func TestShoplyticsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"new-and-returning":     "Get new vs returning customers by date range",
		"first-order-channels":  "Get first-order channel analytics",
		"payments-methods-grid": "Get payment methods grid analytics",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range shoplyticsCmd.Commands() {
				if sub.Use == name || (len(sub.Use) > len(name) && sub.Use[:len(name)] == name) {
					found = true
					if sub.Short != short {
						t.Errorf("expected Short %q, got %q", short, sub.Short)
					}
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not found", name)
			}
		})
	}
}

func TestShoplyticsNewReturningRequiredFlags(t *testing.T) {
	requiredFlags := []string{"start-date", "end-date"}
	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := shoplyticsNewReturningCmd.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("required flag %q not found", name)
			}
			if flag.Annotations == nil {
				t.Fatalf("flag %q has no annotations (expected required)", name)
			}
			if _, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; !ok {
				t.Errorf("flag %q is not marked as required", name)
			}
		})
	}
}

func TestShoplyticsNewReturningRunE_NoAdminToken(t *testing.T) {
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
	cmd.Flags().String("start-date", "2026-01-01", "")
	cmd.Flags().String("end-date", "2026-01-31", "")

	err := shoplyticsNewReturningCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error: %v", err)
	}
}
