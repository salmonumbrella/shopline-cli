package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	// Ensure the command tree is fully wired before alias tests run.
	setupRootCommand()
}

// ---------------------------------------------------------------------------
// TestNoFlagAliasCollisions
// Walk entire command tree, verify no duplicate flag names within any command.
// ---------------------------------------------------------------------------

func TestNoFlagAliasCollisions(t *testing.T) {
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		// Collect all flag names (local + inherited). Track the Flag pointer
		// so that the same flag seen via both Flags() and InheritedFlags()
		// is not counted twice.
		type entry struct {
			flag  *pflag.Flag
			label string // "local" or "inherited"
		}
		seen := map[string]entry{} // flag name -> first occurrence

		record := func(label string, fs *pflag.FlagSet) {
			fs.VisitAll(func(f *pflag.Flag) {
				if prev, dup := seen[f.Name]; dup {
					// Same pointer (inherited == local view of same flag) is fine.
					if prev.flag == f {
						return
					}
					t.Errorf("command %q: duplicate flag name %q (in %s and %s)",
						cmd.CommandPath(), f.Name, prev.label, label)
					return
				}
				seen[f.Name] = entry{flag: f, label: label}
			})
		}

		record("local", cmd.Flags())
		record("inherited", cmd.InheritedFlags())

		for _, sub := range cmd.Commands() {
			walk(sub)
		}
	}

	walk(rootCmd)
}

// ---------------------------------------------------------------------------
// TestVerbAliasesResolve
// Verify that verb aliases on subcommands resolve to the right command.
// ---------------------------------------------------------------------------

func TestVerbAliasesResolve(t *testing.T) {
	tests := []struct {
		args    []string // e.g. ["orders", "l"]
		wantCmd string   // expected cobra command name (e.g. "list")
	}{
		{[]string{"orders", "l"}, "list"},
		{[]string{"orders", "ls"}, "list"},
		{[]string{"products", "g"}, "get"},
		{[]string{"products", "show"}, "get"},
		{[]string{"customers", "mk"}, "create"},
		{[]string{"customers", "new"}, "create"},
		{[]string{"customers", "add"}, "create"},
		{[]string{"orders", "up"}, "update"},
		{[]string{"orders", "edit"}, "update"},
		{[]string{"bulk-operations", "q"}, "query"},
		{[]string{"message-center", "conversations"}, "list"},
	}

	for _, tt := range tests {
		label := strings.Join(tt.args, " ")
		t.Run(label, func(t *testing.T) {
			found, _, err := rootCmd.Find(tt.args)
			if err != nil {
				t.Fatalf("Find(%v) error: %v", tt.args, err)
			}
			if found.Name() != tt.wantCmd {
				t.Errorf("Find(%v) resolved to %q, want %q", tt.args, found.Name(), tt.wantCmd)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestResourceAliasesResolve
// Verify that resource aliases on top-level commands resolve correctly.
// ---------------------------------------------------------------------------

func TestResourceAliasesResolve(t *testing.T) {
	tests := []struct {
		alias    string
		wantName string
	}{
		{"o", "orders"},
		{"ord", "orders"},
		{"p", "products"},
		{"prod", "products"},
		{"cu", "customers"},
		{"cust", "customers"},
		{"contact", "customers"},
		{"contacts", "customers"},
		{"conv", "conversations"},
		{"mc", "message-center"},
		{"rf", "refunds"},
		{"ref", "refunds"},
		{"gc", "gift-cards"},
		{"giftcard", "gift-cards"},
		{"tx", "transactions"},
		{"dc", "discount-codes"},
		{"discounts", "discount-codes"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%s", tt.alias, tt.wantName), func(t *testing.T) {
			found, _, err := rootCmd.Find([]string{tt.alias})
			if err != nil {
				t.Fatalf("Find([%q]) error: %v", tt.alias, err)
			}
			if found.Name() != tt.wantName {
				t.Errorf("Find([%q]) resolved to %q, want %q", tt.alias, found.Name(), tt.wantName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFlagAliasesWork
// Verify that common flag aliases are present on commands that have
// the original flag.
// ---------------------------------------------------------------------------

func TestFlagAliasesWork(t *testing.T) {
	// orders list has --page and --page-size
	ordersListArgs := []string{"orders", "list"}
	ordersListCmd, _, err := rootCmd.Find(ordersListArgs)
	if err != nil {
		t.Fatalf("cannot find orders list: %v", err)
	}

	tests := []struct {
		name      string
		cmd       *cobra.Command
		aliasName string
		aliasOf   string
	}{
		{"page alias pg", ordersListCmd, "pg", "page"},
		{"page-size alias ps", ordersListCmd, "ps", "page-size"},
		{"status alias S", ordersListCmd, "S", "status"},
		{"from alias f", ordersListCmd, "f", "from"},
		{"to alias t", ordersListCmd, "t", "to"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias := tt.cmd.Flags().Lookup(tt.aliasName)
			if alias == nil {
				t.Fatalf("expected flag alias %q on %s, not found", tt.aliasName, tt.cmd.CommandPath())
			}
			if !alias.Hidden {
				t.Errorf("flag alias %q should be hidden", tt.aliasName)
			}
			ann, ok := alias.Annotations["alias-of"]
			if !ok || len(ann) == 0 {
				t.Fatalf("flag alias %q missing alias-of annotation", tt.aliasName)
			}
			if ann[0] != tt.aliasOf {
				t.Errorf("flag alias %q annotated as alias of %q, want %q", tt.aliasName, ann[0], tt.aliasOf)
			}

			// Verify setting the alias also sets the original.
			orig := tt.cmd.Flags().Lookup(tt.aliasOf)
			if orig == nil {
				t.Fatalf("original flag %q not found", tt.aliasOf)
			}
			if w, ok := alias.Value.(*aliasFlagValue); ok {
				if w.Value != orig.Value {
					t.Errorf("alias %q wrapper does not wrap original %q Value", tt.aliasName, tt.aliasOf)
				}
			} else if alias.Value != orig.Value {
				t.Errorf("alias %q and original %q do not share the same Value pointer", tt.aliasName, tt.aliasOf)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestRootFlagShorthands
// Verify that root persistent flags have the expected single-char shorthands.
// ---------------------------------------------------------------------------

func TestRootFlagShorthands(t *testing.T) {
	pf := rootCmd.PersistentFlags()

	tests := []struct {
		flagName  string
		shorthand string
	}{
		{"json", "j"},
		{"yes", "y"},
		{"limit", "l"},
		{"query", "q"},
		{"fields", "F"},
		{"desc", "D"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("-%s/--%s", tt.shorthand, tt.flagName), func(t *testing.T) {
			f := pf.Lookup(tt.flagName)
			if f == nil {
				t.Fatalf("persistent flag %q not found on root", tt.flagName)
			}
			if f.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, f.Shorthand, tt.shorthand)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestRootFlagAliases
// Verify that root persistent flag aliases exist and point to the right flag.
// ---------------------------------------------------------------------------

func TestRootFlagAliases(t *testing.T) {
	pf := rootCmd.PersistentFlags()

	tests := []struct {
		aliasName string
		aliasOf   string
	}{
		{"j", "json"},
		{"out", "output"},
		{"qr", "query"},
		{"qf", "query-file"},
		{"dr", "dry-run"},
		{"sb", "sort-by"},
		{"io", "items-only"},
		{"ro", "results-only"},
		{"at", "admin-token"},
		{"amid", "admin-merchant-id"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("--%s is alias of --%s", tt.aliasName, tt.aliasOf), func(t *testing.T) {
			alias := pf.Lookup(tt.aliasName)
			if alias == nil {
				t.Fatalf("expected alias flag %q on root persistent flags, not found", tt.aliasName)
			}
			if !alias.Hidden {
				t.Errorf("alias flag %q should be hidden", tt.aliasName)
			}
			ann, ok := alias.Annotations["alias-of"]
			if !ok || len(ann) == 0 {
				t.Fatalf("alias flag %q missing alias-of annotation", tt.aliasName)
			}
			if ann[0] != tt.aliasOf {
				t.Errorf("alias flag %q annotated as alias of %q, want %q", tt.aliasName, ann[0], tt.aliasOf)
			}

			orig := pf.Lookup(tt.aliasOf)
			if orig == nil {
				t.Fatalf("original flag %q not found", tt.aliasOf)
			}
			if w, ok := alias.Value.(*aliasFlagValue); ok {
				if w.Value != orig.Value {
					t.Errorf("alias %q wrapper does not wrap original %q Value", tt.aliasName, tt.aliasOf)
				}
			} else if alias.Value != orig.Value {
				t.Errorf("alias %q and original %q do not share the same Value pointer", tt.aliasName, tt.aliasOf)
			}
		})
	}
}

func TestRootNonInteractiveAndResultFlags(t *testing.T) {
	pf := rootCmd.PersistentFlags()
	origItemsOnly, _ := pf.GetBool("items-only")
	origYes, _ := pf.GetBool("yes")
	origNoInput, _ := pf.GetBool("no-input")
	t.Cleanup(func() {
		_ = pf.Set("items-only", fmt.Sprintf("%t", origItemsOnly))
		_ = pf.Set("results-only", fmt.Sprintf("%t", origItemsOnly))
		_ = pf.Set("yes", fmt.Sprintf("%t", origYes))
		_ = pf.Set("force", fmt.Sprintf("%t", origYes))
		_ = pf.Set("no-input", fmt.Sprintf("%t", origNoInput))
	})

	for _, name := range []string{"yes", "force", "no-input", "items-only", "results-only"} {
		if pf.Lookup(name) == nil {
			t.Fatalf("expected persistent flag %q to exist", name)
		}
	}

	_ = pf.Set("results-only", "true")
	itemsOnly, _ := pf.GetBool("items-only")
	if !itemsOnly {
		t.Fatalf("expected --results-only to enable --items-only")
	}

	_ = pf.Set("items-only", "false")
	resultsOnly, _ := pf.GetBool("results-only")
	if resultsOnly {
		t.Fatalf("expected --items-only=false to disable --results-only")
	}

	_ = pf.Set("force", "true")
	yes, _ := pf.GetBool("yes")
	if !yes {
		t.Fatalf("expected --force to enable --yes")
	}

	_ = pf.Set("yes", "false")
	force, _ := pf.GetBool("force")
	if force {
		t.Fatalf("expected --yes=false to disable --force")
	}
}

func TestAliasFlagValuePropagatesChanged(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("original", "", "test flag")
	flagAlias(fs, "original", "alias")

	orig := fs.Lookup("original")
	if orig.Changed {
		t.Fatal("original should not be Changed before setting alias")
	}

	_ = fs.Set("alias", "hello")

	if !orig.Changed {
		t.Error("setting alias should mark original as Changed")
	}
	val, _ := fs.GetString("original")
	if val != "hello" {
		t.Errorf("original value = %q, want %q", val, "hello")
	}
}

func TestFlagAliasMissingOriginalIsNoOp(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagAlias(fs, "missing", "alias")
	if got := fs.Lookup("alias"); got != nil {
		t.Fatalf("expected alias not to be created when original is missing")
	}
}

func TestLocalYesFlagsHaveYShorthand(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"conversations", "delete"})
	if err != nil {
		t.Fatalf("Find conversations delete: %v", err)
	}

	yes := cmd.LocalFlags().Lookup("yes")
	if yes == nil {
		t.Fatalf("expected local --yes on %s", cmd.CommandPath())
	}
	if yes.Shorthand != "y" {
		t.Fatalf("expected local --yes shorthand to be -y, got %q", yes.Shorthand)
	}
}
