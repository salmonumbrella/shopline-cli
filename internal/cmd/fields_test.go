package cmd

import (
	"strings"
	"testing"
)

func TestParseFieldsFlag(t *testing.T) {
	t.Run("csv", func(t *testing.T) {
		fields, err := parseFieldsFlag("id,order_number,status")
		if err != nil {
			t.Fatalf("parseFieldsFlag returned error: %v", err)
		}
		if strings.Join(fields, ",") != "id,order_number,status" {
			t.Fatalf("unexpected fields: %v", fields)
		}
	})

	t.Run("whitespace separated", func(t *testing.T) {
		fields, err := parseFieldsFlag("id order_number\tstatus\ncreated_at")
		if err != nil {
			t.Fatalf("parseFieldsFlag returned error: %v", err)
		}
		if strings.Join(fields, ",") != "id,order_number,status,created_at" {
			t.Fatalf("unexpected fields: %v", fields)
		}
	})

	t.Run("json array", func(t *testing.T) {
		fields, err := parseFieldsFlag(`["id","customer.email","line_items"]`)
		if err != nil {
			t.Fatalf("parseFieldsFlag returned error: %v", err)
		}
		if strings.Join(fields, ",") != "id,customer.email,line_items" {
			t.Fatalf("unexpected fields: %v", fields)
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, err := parseFieldsFlag("   ")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestBuildFieldsQuery(t *testing.T) {
	q := buildFieldsQuery([]string{"id", "customer.email", "foo-bar"})
	// Key + path quoting.
	for _, want := range []string{
		`"id": .["id"]`,
		`"customer.email": .["customer"]["email"]`,
		`"foo-bar": .["foo-bar"]`,
		`.items |= map({`,
	} {
		if !strings.Contains(q, want) {
			t.Fatalf("query missing %q:\n%s", want, q)
		}
	}
}
