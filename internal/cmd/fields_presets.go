package cmd

import (
	"fmt"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var fieldsPresets = map[string]map[string][]string{
	"orders": {
		"minimal": {"id", "order_number", "status", "total_price", "currency", "customer_email", "created_at"},
		"default": {"id", "order_number", "status", "payment_status", "fulfill_status", "total_price", "currency", "customer_email", "customer_name", "created_at"},
		"debug":   {"id", "order_number", "status", "payment_status", "fulfill_status", "total_price", "currency", "customer_id", "customer_email", "customer_name", "tags", "note", "created_at", "updated_at", "line_items"},
	},
	"customers": {
		"minimal": {"id", "email", "first_name", "last_name", "state", "created_at"},
		"default": {"id", "email", "first_name", "last_name", "phone", "state", "orders_count", "total_spent", "currency", "tags", "created_at"},
		"debug":   {"id", "email", "first_name", "last_name", "phone", "state", "accepts_marketing", "orders_count", "total_spent", "currency", "tags", "note", "credit_balance", "subscriptions", "created_at", "updated_at"},
	},
	"products": {
		"minimal": {"id", "title", "status", "vendor", "product_type", "created_at"},
		"default": {"id", "title", "handle", "status", "vendor", "product_type", "tags", "price", "created_at", "updated_at"},
		"debug":   {"id", "title", "handle", "status", "vendor", "product_type", "tags", "description", "price", "created_at", "updated_at"},
	},
}

func parseFieldsWithPresets(cmd *cobra.Command, input string) ([]string, error) {
	fields, err := parseFieldsFlag(input)
	if err != nil {
		return nil, err
	}
	if len(fields) != 1 {
		return fields, nil
	}

	preset := strings.ToLower(strings.TrimSpace(fields[0]))
	if preset != "minimal" && preset != "default" && preset != "debug" {
		return fields, nil
	}

	res := resourceNameForCommand(cmd)
	if res == "" {
		// Not in a resource context; treat preset name as a literal field.
		return fields, nil
	}

	m, ok := fieldsPresets[res]
	if !ok {
		return fields, nil
	}
	p, ok := m[preset]
	if !ok || len(p) == 0 {
		return nil, fmt.Errorf("unknown --fields preset %q for resource %q", preset, res)
	}
	return append([]string{}, p...), nil
}

func resourceNameForCommand(cmd *cobra.Command) string {
	for cur := cmd; cur != nil; cur = cur.Parent() {
		name := cur.Name()
		if name == "" {
			continue
		}
		if _, ok := schema.Get(name); ok {
			return name
		}
	}
	return ""
}
