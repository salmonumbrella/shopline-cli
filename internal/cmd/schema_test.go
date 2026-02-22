package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

func TestBuiltinResourcesRegistered(t *testing.T) {
	resources := []string{"orders", "products", "customers"}
	for _, name := range resources {
		_, ok := schema.Get(name)
		if !ok {
			t.Errorf("expected resource %q to be registered", name)
		}
	}
}

// setupSchemaTest registers test resources for schema command tests.
func setupSchemaTest(t *testing.T) func() {
	t.Helper()
	// Register test resources
	schema.Register(schema.Resource{
		Name:        "test-orders",
		Description: "Manage test orders",
		Commands:    []string{"list", "get", "cancel"},
		IDField:     "id",
		IDPrefix:    "order",
	})
	schema.Register(schema.Resource{
		Name:        "test-products",
		Description: "Manage test products",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDField:     "id",
		IDPrefix:    "product",
	})

	// No cleanup needed since we're adding to global registry
	// and tests use unique resource names
	return func() {}
}

func TestSchemaListResources(t *testing.T) {
	cleanup := setupSchemaTest(t)
	defer cleanup()

	tests := []struct {
		name         string
		outputFormat string
		wantContains []string
	}{
		{
			name:         "text output lists resources",
			outputFormat: "text",
			wantContains: []string{
				"Available resources:",
				"test-orders",
				"test-products",
				"Manage test orders",
				"commands:",
				"spl schema <resource>",
			},
		},
		{
			name:         "json output",
			outputFormat: "json",
			wantContains: []string{
				`"name": "test-orders"`,
				`"name": "test-products"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetOut(buf)
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := runSchema(cmd, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

func TestSchemaShowResource(t *testing.T) {
	cleanup := setupSchemaTest(t)
	defer cleanup()

	tests := []struct {
		name         string
		resource     string
		outputFormat string
		wantContains []string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "text output shows resource details",
			resource:     "test-orders",
			outputFormat: "text",
			wantContains: []string{
				"Resource: test-orders",
				"Description: Manage test orders",
				"ID Prefix: [order:$id]",
				"Commands:",
				"spl test-orders list",
				"spl test-orders get",
				"spl test-orders cancel",
			},
		},
		{
			name:         "json output",
			resource:     "test-orders",
			outputFormat: "json",
			wantContains: []string{
				`"name": "test-orders"`,
				`"id_prefix": "order"`,
			},
		},
		{
			name:        "unknown resource returns error",
			resource:    "nonexistent-resource",
			wantErr:     true,
			errContains: "unknown resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetOut(buf)
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := runSchema(cmd, []string{tt.resource})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error should contain %q, got: %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

func TestSchemaJSONOutputFormat(t *testing.T) {
	cleanup := setupSchemaTest(t)
	defer cleanup()

	t.Run("list resources JSON is valid", func(t *testing.T) {
		buf := new(bytes.Buffer)
		formatterWriter = buf

		cmd := &cobra.Command{Use: "test"}
		cmd.SetOut(buf)
		cmd.Flags().String("output", "json", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")

		err := runSchema(cmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it's valid JSON
		var resources []schema.Resource
		if err := json.Unmarshal(buf.Bytes(), &resources); err != nil {
			t.Errorf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
		}
	})

	t.Run("show resource JSON is valid", func(t *testing.T) {
		buf := new(bytes.Buffer)
		formatterWriter = buf

		cmd := &cobra.Command{Use: "test"}
		cmd.SetOut(buf)
		cmd.Flags().String("output", "json", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")

		err := runSchema(cmd, []string{"test-orders"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it's valid JSON
		var resource schema.Resource
		if err := json.Unmarshal(buf.Bytes(), &resource); err != nil {
			t.Errorf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
		}

		if resource.Name != "test-orders" {
			t.Errorf("expected resource name 'test-orders', got %q", resource.Name)
		}
	})
}

func TestSchemaCmdArgs(t *testing.T) {
	// Test that the command accepts 0 or 1 args
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg",
			args:    []string{"test-orders"},
			wantErr: false,
		},
		{
			name:    "two args rejected",
			args:    []string{"arg1", "arg2"},
			wantErr: true,
		},
	}

	cleanup := setupSchemaTest(t)
	defer cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the actual schemaCmd to test Args validation
			err := schemaCmd.Args(schemaCmd, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSchemaErrorMessage(t *testing.T) {
	// Test that error message includes helpful hint
	cleanup := setupSchemaTest(t)
	defer cleanup()

	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(buf)
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := runSchema(cmd, []string{"definitely-not-a-resource"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "unknown resource") {
		t.Errorf("error should contain 'unknown resource', got: %v", errMsg)
	}
	if !strings.Contains(errMsg, "spl schema") {
		t.Errorf("error should contain hint 'spl schema', got: %v", errMsg)
	}
}
