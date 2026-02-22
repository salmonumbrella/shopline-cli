package cmd

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// singularize (aliases.go)
// ---------------------------------------------------------------------------

func TestSingularize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// "ies" -> "y"
		{"categories", "category"},
		{"policies", "policy"},
		{"entries", "entry"},
		// "xes" -> strip last 2
		{"taxes", "tax"},
		{"boxes", "box"},
		// "ses" -> strip last 2
		{"addresses", "address"},
		{"buses", "bus"},
		// "ches" -> strip last 2
		{"watches", "watch"},
		{"batches", "batch"},
		// "shes" -> strip last 2
		{"dishes", "dish"},
		{"crashes", "crash"},
		// unchanged: ends in "us"
		{"status", "status"},
		{"census", "census"},
		// unchanged: ends in "ss"
		{"address", "address"},
		{"access", "access"},
		// unchanged: ends in "is"
		{"analysis", "analysis"},
		{"basis", "basis"},
		// regular trailing "s" stripped
		{"orders", "order"},
		{"products", "product"},
		{"customers", "customer"},
		// no trailing "s" -> unchanged
		{"inventory", "inventory"},
		{"auth", "auth"},
		// single char "s" -> unchanged (len <= 1)
		{"s", "s"},
		// empty string
		{"", ""},
		// short "ies" exactly 3 chars - len(name) > 3 guard prevents ies->y rule
		{"ies", "ie"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := singularize(tt.input)
			if got != tt.want {
				t.Errorf("singularize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// addAliasIfSafe (aliases.go)
// ---------------------------------------------------------------------------

func TestAddAliasIfSafe(t *testing.T) {
	t.Run("adds alias when no collision", func(t *testing.T) {
		parent := &cobra.Command{Use: "root"}
		child := &cobra.Command{Use: "orders"}
		parent.AddCommand(child)

		addAliasIfSafe(child, "ord")

		if !slices.Contains(child.Aliases, "ord") {
			t.Errorf("expected alias 'ord' to be added, got %v", child.Aliases)
		}
	})

	t.Run("skips empty alias", func(t *testing.T) {
		cmd := &cobra.Command{Use: "orders"}
		addAliasIfSafe(cmd, "")

		if len(cmd.Aliases) != 0 {
			t.Errorf("expected no aliases, got %v", cmd.Aliases)
		}
	})

	t.Run("skips alias equal to command name", func(t *testing.T) {
		cmd := &cobra.Command{Use: "orders"}
		addAliasIfSafe(cmd, "orders")

		if len(cmd.Aliases) != 0 {
			t.Errorf("expected no aliases, got %v", cmd.Aliases)
		}
	})

	t.Run("skips duplicate alias", func(t *testing.T) {
		cmd := &cobra.Command{Use: "orders", Aliases: []string{"ord"}}
		addAliasIfSafe(cmd, "ord")

		if len(cmd.Aliases) != 1 {
			t.Errorf("expected 1 alias, got %v", cmd.Aliases)
		}
	})

	t.Run("skips alias colliding with sibling name", func(t *testing.T) {
		parent := &cobra.Command{Use: "root"}
		child1 := &cobra.Command{Use: "orders"}
		child2 := &cobra.Command{Use: "ord"}
		parent.AddCommand(child1, child2)

		addAliasIfSafe(child1, "ord")

		if slices.Contains(child1.Aliases, "ord") {
			t.Errorf("alias 'ord' should not be added, collides with sibling name")
		}
	})

	t.Run("skips alias colliding with sibling alias", func(t *testing.T) {
		parent := &cobra.Command{Use: "root"}
		child1 := &cobra.Command{Use: "orders"}
		child2 := &cobra.Command{Use: "products", Aliases: []string{"ord"}}
		parent.AddCommand(child1, child2)

		addAliasIfSafe(child1, "ord")

		if slices.Contains(child1.Aliases, "ord") {
			t.Errorf("alias 'ord' should not be added, collides with sibling alias")
		}
	})

	t.Run("adds alias when no parent", func(t *testing.T) {
		cmd := &cobra.Command{Use: "orders"}
		addAliasIfSafe(cmd, "ord")

		if !slices.Contains(cmd.Aliases, "ord") {
			t.Errorf("expected alias 'ord' to be added for parentless command, got %v", cmd.Aliases)
		}
	})
}

// ---------------------------------------------------------------------------
// normalizeIDToken (helpers.go)
// ---------------------------------------------------------------------------

func TestNormalizeIDToken(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantID string
		wantOK bool
	}{
		{"formatted ID", "[order:$ord_123]", "ord_123", true},
		{"formatted ID with prefix", "[product:$prod_456]", "prod_456", true},
		{"plain ID", "plain_id", "plain_id", false},
		{"empty string", "", "", false},
		{"partial format missing bracket", "[order:$ord_123", "[order:$ord_123", false},
		{"partial format missing dollar", "[order:ord_123]", "[order:ord_123]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := normalizeIDToken(tt.input)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf("normalizeIDToken(%q) = (%q, %v), want (%q, %v)",
					tt.input, gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// normalizeIDArgs (helpers.go)
// ---------------------------------------------------------------------------

func TestNormalizeIDArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			"mixed formatted and plain",
			[]string{"[order:$ord_123]", "plain_id", "[product:$prod_456]"},
			[]string{"ord_123", "plain_id", "prod_456"},
		},
		{
			"all plain",
			[]string{"abc", "def"},
			[]string{"abc", "def"},
		},
		{
			"empty slice",
			[]string{},
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			normalizeIDArgs(args)

			for i, want := range tt.want {
				if args[i] != want {
					t.Errorf("args[%d] = %q, want %q", i, args[i], want)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isIDFlag (helpers.go)
// ---------------------------------------------------------------------------

func TestIsIDFlag(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"id", true},
		{"ids", true},
		{"order-id", true},
		{"customer-ids", true},
		{"resource-id", true},
		{"variant-ids", true},
		{"name", false},
		{"identifier", false},
		{"identity", false},
		{"valid", false},
		{"", false},
		{"id-prefix", false},
		{"ids-list", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isIDFlag(tt.name)
			if got != tt.want {
				t.Errorf("isIDFlag(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// applyLimitToPageSize (helpers.go)
// ---------------------------------------------------------------------------

func TestApplyLimitToPageSize(t *testing.T) {
	t.Run("nil command", func(t *testing.T) {
		err := applyLimitToPageSize(nil)
		if err != nil {
			t.Errorf("expected nil error for nil cmd, got %v", err)
		}
	})

	t.Run("limit not changed", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("limit", 0, "")
		cmd.Flags().Int("page-size", 20, "")

		err := applyLimitToPageSize(cmd)
		if err != nil {
			t.Errorf("expected nil error when limit not changed, got %v", err)
		}

		ps, _ := cmd.Flags().GetInt("page-size")
		if ps != 20 {
			t.Errorf("page-size should remain 20, got %d", ps)
		}
	})

	t.Run("limit sets page-size", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("limit", 0, "")
		cmd.Flags().Int("page-size", 20, "")
		_ = cmd.Flags().Set("limit", "50")

		err := applyLimitToPageSize(cmd)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		ps, _ := cmd.Flags().GetInt("page-size")
		if ps != 50 {
			t.Errorf("page-size should be 50, got %d", ps)
		}
	})

	t.Run("negative limit returns error", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("limit", 0, "")
		cmd.Flags().Int("page-size", 20, "")
		_ = cmd.Flags().Set("limit", "-1")

		err := applyLimitToPageSize(cmd)
		if err == nil {
			t.Error("expected error for negative limit")
		}
		if err != nil && !strings.Contains(err.Error(), "limit must be >= 0") {
			t.Errorf("expected 'limit must be >= 0' error, got: %v", err)
		}
	})

	t.Run("no page-size flag is a no-op", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("limit", 0, "")
		_ = cmd.Flags().Set("limit", "10")

		err := applyLimitToPageSize(cmd)
		if err != nil {
			t.Errorf("expected nil error when page-size flag missing, got %v", err)
		}
	})

	t.Run("limit zero sets page-size to zero", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("limit", 0, "")
		cmd.Flags().Int("page-size", 20, "")
		_ = cmd.Flags().Set("limit", "0")

		err := applyLimitToPageSize(cmd)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		ps, _ := cmd.Flags().GetInt("page-size")
		if ps != 0 {
			t.Errorf("page-size should be 0, got %d", ps)
		}
	})

	t.Run("persistent limit sets page-size", func(t *testing.T) {
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Int("limit", 0, "")

		child := &cobra.Command{
			Use: "child",
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := applyLimitToPageSize(cmd); err != nil {
					return err
				}
				ps, _ := cmd.Flags().GetInt("page-size")
				if ps != 5 {
					t.Fatalf("page-size should be 5, got %d", ps)
				}
				return nil
			},
		}
		child.Flags().Int("page-size", 20, "")
		root.AddCommand(child)

		root.SetArgs([]string{"child", "--limit", "5"})
		if err := root.Execute(); err != nil {
			t.Fatalf("execute failed: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// readSortOptions (helpers.go)
// ---------------------------------------------------------------------------

func TestReadSortOptions(t *testing.T) {
	t.Run("nil command", func(t *testing.T) {
		sortBy, order := readSortOptions(nil)
		if sortBy != "" || order != "" {
			t.Errorf("expected empty strings for nil cmd, got (%q, %q)", sortBy, order)
		}
	})

	t.Run("no sort-by set", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("sort-by", "", "")
		cmd.Flags().Bool("desc", false, "")

		sortBy, order := readSortOptions(cmd)
		if sortBy != "" || order != "" {
			t.Errorf("expected empty strings when sort-by not set, got (%q, %q)", sortBy, order)
		}
	})

	t.Run("sort-by ascending (default)", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("sort-by", "", "")
		cmd.Flags().Bool("desc", false, "")
		_ = cmd.Flags().Set("sort-by", "created_at")

		sortBy, order := readSortOptions(cmd)
		if sortBy != "created_at" {
			t.Errorf("expected sort-by 'created_at', got %q", sortBy)
		}
		if order != "asc" {
			t.Errorf("expected order 'asc', got %q", order)
		}
	})

	t.Run("sort-by descending", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("sort-by", "", "")
		cmd.Flags().Bool("desc", false, "")
		_ = cmd.Flags().Set("sort-by", "updated_at")
		_ = cmd.Flags().Set("desc", "true")

		sortBy, order := readSortOptions(cmd)
		if sortBy != "updated_at" {
			t.Errorf("expected sort-by 'updated_at', got %q", sortBy)
		}
		if order != "desc" {
			t.Errorf("expected order 'desc', got %q", order)
		}
	})
}

// ---------------------------------------------------------------------------
// parseTimeFlag (helpers.go)
// ---------------------------------------------------------------------------

func TestParseTimeFlag(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		label   string
		wantNil bool
		wantErr bool
		errMsg  string
		wantY   int
		wantM   time.Month
		wantD   int
	}{
		{
			name:    "empty value returns nil",
			value:   "",
			label:   "start",
			wantNil: true,
		},
		{
			name:  "RFC3339 format",
			value: "2024-01-15T10:30:00Z",
			label: "start",
			wantY: 2024, wantM: time.January, wantD: 15,
		},
		{
			name:  "short date format",
			value: "2024-06-30",
			label: "end",
			wantY: 2024, wantM: time.June, wantD: 30,
		},
		{
			name:    "invalid date",
			value:   "not-a-date",
			label:   "from",
			wantErr: true,
			errMsg:  "invalid from date format",
		},
		{
			name:    "wrong separator",
			value:   "2024/01/01",
			label:   "to",
			wantErr: true,
			errMsg:  "invalid to date format",
		},
		{
			name:    "US date format",
			value:   "01-15-2024",
			label:   "since",
			wantErr: true,
			errMsg:  "invalid since date format",
		},
		{
			name:  "RFC3339 with timezone offset",
			value: "2024-03-15T08:00:00+05:00",
			label: "start",
			wantY: 2024, wantM: time.March, wantD: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeFlag(tt.value, tt.label)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil time, got nil")
			}
			if got.Year() != tt.wantY || got.Month() != tt.wantM || got.Day() != tt.wantD {
				t.Errorf("got date %v, want %d-%02d-%02d", got, tt.wantY, tt.wantM, tt.wantD)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// resolveOrArg (helpers.go)
// ---------------------------------------------------------------------------

func TestResolveOrArg(t *testing.T) {
	t.Run("returns positional arg when provided", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("by", "", "")

		id, err := resolveOrArg(cmd, []string{"my-id"}, func(q string) (string, error) {
			t.Fatal("resolver should not be called when positional arg given")
			return "", nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "my-id" {
			t.Errorf("expected 'my-id', got %q", id)
		}
	})

	t.Run("errors when no arg and no --by", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("by", "", "")

		_, err := resolveOrArg(cmd, nil, func(q string) (string, error) {
			return "", nil
		})
		if err == nil {
			t.Fatal("expected error when no arg and no --by")
		}
		if !strings.Contains(err.Error(), "provide a resource ID") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("calls resolver with --by value", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "alice@example.com")

		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)

		id, err := resolveOrArg(cmd, nil, func(q string) (string, error) {
			if q != "alice@example.com" {
				t.Errorf("expected query 'alice@example.com', got %q", q)
			}
			return "cust_123", nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "cust_123" {
			t.Errorf("expected 'cust_123', got %q", id)
		}
		if !strings.Contains(stderr.String(), "Resolved to cust_123") {
			t.Errorf("expected stderr to contain 'Resolved to cust_123', got %q", stderr.String())
		}
	})

	t.Run("propagates resolver error", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "query")

		_, err := resolveOrArg(cmd, nil, func(q string) (string, error) {
			return "", errors.New("not found")
		})
		if err == nil {
			t.Fatal("expected error from resolver")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
