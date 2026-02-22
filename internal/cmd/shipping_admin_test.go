package cmd

import (
	"testing"
)

// TestShippingCommandSetup verifies shipping command initialization
func TestShippingCommandSetup(t *testing.T) {
	if shippingCmd.Use != "shipping" {
		t.Errorf("expected Use 'shipping', got %q", shippingCmd.Use)
	}
	if shippingCmd.Short != "Manage order shipments, tracking, and labels (via Admin API)" {
		t.Errorf("expected Short 'Manage order shipments, tracking, and labels (via Admin API)', got %q", shippingCmd.Short)
	}
	expectedAliases := map[string]bool{"ship": false, "sh": false}
	for _, a := range shippingCmd.Aliases {
		if _, ok := expectedAliases[a]; ok {
			expectedAliases[a] = true
		}
	}
	for alias, found := range expectedAliases {
		if !found {
			t.Errorf("expected alias %q not found in %v", alias, shippingCmd.Aliases)
		}
	}
}

// TestShippingSubcommands verifies all 4 subcommands are registered
func TestShippingSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"status":      "Check if a shipment has been executed",
		"tracking":    "Get tracking number for an order",
		"execute":     "Execute shipment for an order",
		"print-label": "Generate and retrieve packing label",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range shippingCmd.Commands() {
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

// TestShippingStatusCommandSetup verifies the status command setup
func TestShippingStatusCommandSetup(t *testing.T) {
	if shippingStatusCmd.Use != "status <order-id>" {
		t.Errorf("expected Use 'status <order-id>', got %q", shippingStatusCmd.Use)
	}
	if shippingStatusCmd.Short != "Check if a shipment has been executed" {
		t.Errorf("expected Short 'Check if a shipment has been executed', got %q", shippingStatusCmd.Short)
	}
	if shippingStatusCmd.Args == nil {
		t.Error("expected Args to be set (ExactArgs(1))")
	}
}

// TestShippingTrackingCommandSetup verifies the tracking command setup
func TestShippingTrackingCommandSetup(t *testing.T) {
	if shippingTrackingCmd.Use != "tracking <order-id>" {
		t.Errorf("expected Use 'tracking <order-id>', got %q", shippingTrackingCmd.Use)
	}
	if shippingTrackingCmd.Short != "Get tracking number for an order" {
		t.Errorf("expected Short 'Get tracking number for an order', got %q", shippingTrackingCmd.Short)
	}
	if shippingTrackingCmd.Args == nil {
		t.Error("expected Args to be set (ExactArgs(1))")
	}
}

// TestShippingExecuteCommandSetup verifies the execute command setup
func TestShippingExecuteCommandSetup(t *testing.T) {
	if shippingExecuteCmd.Use != "execute <order-id>" {
		t.Errorf("expected Use 'execute <order-id>', got %q", shippingExecuteCmd.Use)
	}
	if shippingExecuteCmd.Short != "Execute shipment for an order" {
		t.Errorf("expected Short 'Execute shipment for an order', got %q", shippingExecuteCmd.Short)
	}
	if shippingExecuteCmd.Args == nil {
		t.Error("expected Args to be set (ExactArgs(1))")
	}
}

// TestShippingExecuteFlags verifies --order-number and --performer-id flags exist
func TestShippingExecuteFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"order-number", ""},
		{"performer-id", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := shippingExecuteCmd.Flags().Lookup(f.name)
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

// TestShippingPrintLabelCommandSetup verifies the print-label command setup
func TestShippingPrintLabelCommandSetup(t *testing.T) {
	if shippingPrintLabelCmd.Use != "print-label <order-id>" {
		t.Errorf("expected Use 'print-label <order-id>', got %q", shippingPrintLabelCmd.Use)
	}
	if shippingPrintLabelCmd.Short != "Generate and retrieve packing label" {
		t.Errorf("expected Short 'Generate and retrieve packing label', got %q", shippingPrintLabelCmd.Short)
	}
	if shippingPrintLabelCmd.Args == nil {
		t.Error("expected Args to be set (ExactArgs(1))")
	}
	expectedAliases := []string{"label", "print"}
	if len(shippingPrintLabelCmd.Aliases) != len(expectedAliases) {
		t.Fatalf("expected %d aliases, got %d", len(expectedAliases), len(shippingPrintLabelCmd.Aliases))
	}
	for i, alias := range expectedAliases {
		if shippingPrintLabelCmd.Aliases[i] != alias {
			t.Errorf("expected alias %q at index %d, got %q", alias, i, shippingPrintLabelCmd.Aliases[i])
		}
	}
}

// TestShippingPrintLabelFlags verifies --upsert flag exists
func TestShippingPrintLabelFlags(t *testing.T) {
	flag := shippingPrintLabelCmd.Flags().Lookup("upsert")
	if flag == nil {
		t.Fatal("flag 'upsert' not found")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}

// TestShippingStatusRunE_NoAdminToken verifies error path when no Admin token is set
func TestShippingStatusRunE_NoAdminToken(t *testing.T) {
	t.Setenv("SHOPLINE_ADMIN_BASE_URL", "https://test.example.com")
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")

	err := shippingStatusCmd.RunE(cmd, []string{"order_123"})
	if err == nil {
		t.Fatal("expected error when no Admin token is set, got nil")
	}
	expected := "admin API token required"
	if got := err.Error(); len(got) < len(expected) || got[:len(expected)] != expected {
		// Check that the error message contains the expected substring
		contains := false
		for i := 0; i <= len(got)-len(expected); i++ {
			if got[i:i+len(expected)] == expected {
				contains = true
				break
			}
		}
		if !contains {
			t.Errorf("expected error containing %q, got %q", expected, got)
		}
	}
}
