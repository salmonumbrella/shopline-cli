package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestFieldAliasesCommand(t *testing.T) {
	var buf bytes.Buffer
	cmd := newFieldAliasesCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "on") || !strings.Contains(out, "order_number") {
		t.Errorf("expected alias table to contain on -> order_number, got:\n%s", out)
	}
}

func TestFieldAliasesCommand_ContainsHeader(t *testing.T) {
	var buf bytes.Buffer
	cmd := newFieldAliasesCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ALIAS") || !strings.Contains(out, "FIELD") {
		t.Errorf("expected alias table to contain ALIAS and FIELD headers, got:\n%s", out)
	}
}

func TestFieldAliasesCommand_IsSorted(t *testing.T) {
	var buf bytes.Buffer
	cmd := newFieldAliasesCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (header + 2 entries), got %d", len(lines))
	}
	// Skip header, check remaining lines are sorted by alias
	var prevAlias string
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		alias := fields[0]
		if prevAlias != "" && alias < prevAlias {
			t.Errorf("aliases not sorted: %q came after %q", alias, prevAlias)
		}
		prevAlias = alias
	}
}
