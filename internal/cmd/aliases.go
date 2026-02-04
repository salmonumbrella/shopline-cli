package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var verbAliases = map[string][]string{
	"list":   {"ls"},
	"get":    {"show"},
	"create": {"new", "add"},
	"update": {"edit"},
	"delete": {"del", "rm"},
	"remove": {"rm"},
	"cancel": {"void"},
	"login":  {"signin", "sign-in"},
	"logout": {"signout", "sign-out"},
}

var resourceAliases = map[string][]string{
	"orders":         {"ord"},
	"products":       {"prod"},
	"customers":      {"cust"},
	"inventory":      {"inv"},
	"draft-orders":   {"drafts"},
	"gift-cards":     {"giftcard", "gc"},
	"discount-codes": {"discounts"},
	"webhooks":       {"hooks"},
}

func applyDesirePathAliases(root *cobra.Command) {
	if root == nil {
		return
	}
	applyAliasesRecursive(root, root)
}

func applyAliasesRecursive(cmd *cobra.Command, root *cobra.Command) {
	addDesireAliases(cmd, root)
	for _, sub := range cmd.Commands() {
		applyAliasesRecursive(sub, root)
	}
}

func addDesireAliases(cmd *cobra.Command, root *cobra.Command) {
	name := cmd.Name()
	if name == "" {
		return
	}

	if aliases, ok := verbAliases[name]; ok {
		for _, a := range aliases {
			addAliasIfSafe(cmd, a)
		}
	}

	if cmd.Parent() == root {
		if singular := singularize(name); singular != "" && singular != name {
			addAliasIfSafe(cmd, singular)
		}
		if aliases, ok := resourceAliases[name]; ok {
			for _, a := range aliases {
				addAliasIfSafe(cmd, a)
			}
		}
	}
}

func addAliasIfSafe(cmd *cobra.Command, alias string) {
	if alias == "" || alias == cmd.Name() || contains(cmd.Aliases, alias) {
		return
	}
	parent := cmd.Parent()
	if parent != nil {
		for _, sibling := range parent.Commands() {
			if sibling == cmd {
				continue
			}
			if sibling.Name() == alias || contains(sibling.Aliases, alias) {
				return
			}
		}
	}
	cmd.Aliases = append(cmd.Aliases, alias)
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func singularize(name string) string {
	if strings.HasSuffix(name, "ies") && len(name) > 3 {
		return name[:len(name)-3] + "y"
	}
	if strings.HasSuffix(name, "xes") || strings.HasSuffix(name, "ses") ||
		strings.HasSuffix(name, "ches") || strings.HasSuffix(name, "shes") {
		return name[:len(name)-2]
	}
	// Don't strip trailing "s" from words ending in "us", "ss", or "is"
	if strings.HasSuffix(name, "us") || strings.HasSuffix(name, "ss") || strings.HasSuffix(name, "is") {
		return name
	}
	if strings.HasSuffix(name, "s") && len(name) > 1 {
		return name[:len(name)-1]
	}
	return name
}
