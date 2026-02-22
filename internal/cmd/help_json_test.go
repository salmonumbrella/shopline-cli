package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// buildHelpCommand (help_json.go)
// ---------------------------------------------------------------------------

func TestBuildHelpCommand(t *testing.T) {
	t.Run("basic command fields", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "orders [flags]",
			Short:   "Manage orders",
			Long:    "Manage orders in the store",
			Aliases: []string{"ord"},
			Example: "spl orders list",
		}

		info := buildHelpCommand(cmd, false)

		if info.Name != "orders" {
			t.Errorf("Name = %q, want 'orders'", info.Name)
		}
		if info.Use != "orders [flags]" {
			t.Errorf("Use = %q, want 'orders [flags]'", info.Use)
		}
		if info.Short != "Manage orders" {
			t.Errorf("Short = %q, want 'Manage orders'", info.Short)
		}
		if info.Long != "Manage orders in the store" {
			t.Errorf("Long = %q, want 'Manage orders in the store'", info.Long)
		}
		if len(info.Aliases) != 1 || info.Aliases[0] != "ord" {
			t.Errorf("Aliases = %v, want [ord]", info.Aliases)
		}
		if info.Example != "spl orders list" {
			t.Errorf("Example = %q, want 'spl orders list'", info.Example)
		}
	})

	t.Run("deprecated field", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:        "old-cmd",
			Deprecated: "use new-cmd instead",
		}

		info := buildHelpCommand(cmd, false)
		if info.Deprecated != "use new-cmd instead" {
			t.Errorf("Deprecated = %q, want 'use new-cmd instead'", info.Deprecated)
		}
	})

	t.Run("includes visible subcommands sorted", func(t *testing.T) {
		parent := &cobra.Command{Use: "root"}
		parent.AddCommand(
			&cobra.Command{Use: "zebra", Short: "Z cmd"},
			&cobra.Command{Use: "alpha", Short: "A cmd"},
			&cobra.Command{Use: "middle", Short: "M cmd"},
		)

		info := buildHelpCommand(parent, false)

		if len(info.Subcommands) != 3 {
			t.Fatalf("expected 3 subcommands, got %d", len(info.Subcommands))
		}
		if len(info.Commands) != 3 {
			t.Fatalf("expected 3 commands alias entries, got %d", len(info.Commands))
		}
		if info.Subcommands[0].Name != "alpha" {
			t.Errorf("first subcommand = %q, want 'alpha'", info.Subcommands[0].Name)
		}
		if info.Subcommands[1].Name != "middle" {
			t.Errorf("second subcommand = %q, want 'middle'", info.Subcommands[1].Name)
		}
		if info.Subcommands[2].Name != "zebra" {
			t.Errorf("third subcommand = %q, want 'zebra'", info.Subcommands[2].Name)
		}
	})

	t.Run("hides hidden subcommands", func(t *testing.T) {
		parent := &cobra.Command{Use: "root"}
		visible := &cobra.Command{Use: "visible", Short: "Visible cmd"}
		hidden := &cobra.Command{Use: "hidden", Short: "Hidden cmd", Hidden: true}
		parent.AddCommand(visible, hidden)

		info := buildHelpCommand(parent, false)

		if len(info.Subcommands) != 1 {
			t.Fatalf("expected 1 visible subcommand, got %d", len(info.Subcommands))
		}
		if info.Subcommands[0].Name != "visible" {
			t.Errorf("subcommand = %q, want 'visible'", info.Subcommands[0].Name)
		}
	})

	t.Run("includes flags sorted by name", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().StringP("zone", "z", "", "Zone name")
		cmd.Flags().IntP("count", "c", 10, "Item count")
		cmd.Flags().Bool("verbose", false, "Verbose output")

		info := buildHelpCommand(cmd, false)

		if len(info.Flags) < 3 {
			t.Fatalf("expected at least 3 flags, got %d", len(info.Flags))
		}

		// Flags should be sorted alphabetically
		for i := 1; i < len(info.Flags); i++ {
			if info.Flags[i].Name < info.Flags[i-1].Name {
				t.Errorf("flags not sorted: %q comes after %q",
					info.Flags[i].Name, info.Flags[i-1].Name)
			}
		}
	})

	t.Run("subcommand aliases are copied", func(t *testing.T) {
		parent := &cobra.Command{Use: "root"}
		child := &cobra.Command{Use: "orders", Aliases: []string{"ord", "o"}}
		parent.AddCommand(child)

		info := buildHelpCommand(parent, false)

		if len(info.Subcommands) != 1 {
			t.Fatalf("expected 1 subcommand, got %d", len(info.Subcommands))
		}
		sub := info.Subcommands[0]
		if len(sub.Aliases) != 2 {
			t.Fatalf("expected 2 aliases, got %d", len(sub.Aliases))
		}
		if sub.Aliases[0] != "ord" || sub.Aliases[1] != "o" {
			t.Errorf("aliases = %v, want [ord o]", sub.Aliases)
		}

		// Verify it's a copy, not a reference to the original
		sub.Aliases[0] = "modified"
		if child.Aliases[0] == "modified" {
			t.Error("subcommand aliases should be a copy, not a reference")
		}
	})

	t.Run("empty command", func(t *testing.T) {
		cmd := &cobra.Command{Use: "empty"}
		info := buildHelpCommand(cmd, false)

		if info.Name != "empty" {
			t.Errorf("Name = %q, want 'empty'", info.Name)
		}
		if len(info.Subcommands) != 0 {
			t.Errorf("expected no subcommands, got %d", len(info.Subcommands))
		}
	})

	t.Run("deep includes nested subcommands", func(t *testing.T) {
		root := &cobra.Command{Use: "root"}
		parent := &cobra.Command{Use: "parent"}
		child := &cobra.Command{Use: "child", Short: "Child cmd"}
		parent.AddCommand(child)
		root.AddCommand(parent)

		info := buildHelpCommand(root, true)
		if len(info.Subcommands) != 1 || info.Subcommands[0].Name != "parent" {
			t.Fatalf("unexpected top-level subcommands: %+v", info.Subcommands)
		}
		if len(info.Subcommands[0].Subcommands) != 1 || info.Subcommands[0].Subcommands[0].Name != "child" {
			t.Fatalf("expected nested child subcommand, got %+v", info.Subcommands[0].Subcommands)
		}
	})
}

// ---------------------------------------------------------------------------
// collectHelpFlags (help_json.go)
// ---------------------------------------------------------------------------

func TestCollectHelpFlags(t *testing.T) {
	t.Run("local flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().StringP("output", "o", "text", "Output format")
		cmd.Flags().Bool("verbose", false, "Verbose mode")

		flags := collectHelpFlags(cmd)

		found := map[string]helpFlag{}
		for _, f := range flags {
			found[f.Name] = f
		}

		outputF, ok := found["output"]
		if !ok {
			t.Fatal("expected 'output' flag")
		}
		if outputF.Shorthand != "o" {
			t.Errorf("output shorthand = %q, want 'o'", outputF.Shorthand)
		}
		if outputF.Type != "string" {
			t.Errorf("output type = %q, want 'string'", outputF.Type)
		}
		if outputF.Default != "text" {
			t.Errorf("output default = %q, want 'text'", outputF.Default)
		}
		if outputF.Usage != "Output format" {
			t.Errorf("output usage = %q, want 'Output format'", outputF.Usage)
		}
		if outputF.Persistent {
			t.Error("local flag should not be persistent")
		}

		verboseF, ok := found["verbose"]
		if !ok {
			t.Fatal("expected 'verbose' flag")
		}
		if verboseF.Type != "bool" {
			t.Errorf("verbose type = %q, want 'bool'", verboseF.Type)
		}
	})

	t.Run("inherited persistent flags", func(t *testing.T) {
		parent := &cobra.Command{Use: "root"}
		parent.PersistentFlags().String("store", "", "Store name")

		child := &cobra.Command{Use: "list"}
		parent.AddCommand(child)

		// Need to trigger flag merging
		child.InheritedFlags()

		flags := collectHelpFlags(child)

		found := map[string]helpFlag{}
		for _, f := range flags {
			found[f.Name] = f
		}

		storeF, ok := found["store"]
		if !ok {
			t.Fatal("expected inherited 'store' flag")
		}
		if !storeF.Persistent {
			t.Error("inherited flag should be persistent")
		}
	})

	t.Run("deprecated flag", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("old-flag", "", "Old flag")
		_ = cmd.Flags().MarkDeprecated("old-flag", "use new-flag instead")

		flags := collectHelpFlags(cmd)

		found := map[string]helpFlag{}
		for _, f := range flags {
			found[f.Name] = f
		}

		oldF, ok := found["old-flag"]
		if !ok {
			t.Fatal("expected 'old-flag' flag")
		}
		if !oldF.Deprecated {
			t.Error("old-flag should be marked deprecated")
		}
	})
}

// ---------------------------------------------------------------------------
// isFlagRequired (help_json.go)
// ---------------------------------------------------------------------------

func TestIsFlagRequired(t *testing.T) {
	t.Run("nil flag", func(t *testing.T) {
		if isFlagRequired(nil) {
			t.Error("expected false for nil flag")
		}
	})

	t.Run("flag without annotations", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("name", "", "Name")
		f := cmd.Flags().Lookup("name")

		if isFlagRequired(f) {
			t.Error("expected false for flag without annotations")
		}
	})

	t.Run("required flag via cobra MarkFlagRequired", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("id", "", "ID")
		_ = cmd.MarkFlagRequired("id")
		f := cmd.Flags().Lookup("id")

		if !isFlagRequired(f) {
			t.Error("expected true for required flag")
		}
	})

	t.Run("flag with empty annotation", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("opt", "", "Optional")
		f := cmd.Flags().Lookup("opt")
		f.Annotations = map[string][]string{
			cobra.BashCompOneRequiredFlag: {},
		}

		if isFlagRequired(f) {
			t.Error("expected false for flag with empty annotation values")
		}
	})

	t.Run("flag with annotation value not 'true'", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("opt", "", "Optional")
		f := cmd.Flags().Lookup("opt")
		f.Annotations = map[string][]string{
			cobra.BashCompOneRequiredFlag: {"false"},
		}

		if isFlagRequired(f) {
			t.Error("expected false for flag with annotation value 'false'")
		}
	})
}
