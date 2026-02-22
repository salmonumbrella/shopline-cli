package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func parseFieldsFlag(input string) ([]string, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, fmt.Errorf("--fields must include at least one field")
	}

	// Support file indirection: --fields @path or --fields @- (stdin).
	if strings.HasPrefix(raw, "@") {
		path := strings.TrimSpace(strings.TrimPrefix(raw, "@"))
		if path == "" {
			return nil, fmt.Errorf("--fields must include at least one field")
		}
		var b []byte
		var err error
		if path == "-" {
			b, err = io.ReadAll(os.Stdin)
		} else {
			b, err = os.ReadFile(path)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read --fields file: %w", err)
		}
		raw = strings.TrimSpace(string(b))
		if raw == "" {
			return nil, fmt.Errorf("--fields must include at least one field")
		}
	}

	// JSON array input: --fields '["id","email"]'
	if strings.HasPrefix(raw, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			var out []string
			for _, s := range arr {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
			if len(out) == 0 {
				return nil, fmt.Errorf("--fields must include at least one field")
			}
			return out, nil
		}
		// Fall through to CSV parsing for "array-like" inputs that aren't valid JSON.
	}

	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("--fields must include at least one field")
	}
	return out, nil
}

func buildFieldsQuery(fields []string) string {
	// Keep pagination envelopes intact when present:
	// - if {items: [...]}, project items in place
	// - if [...], map directly
	// - else, project object fields
	var parts []string
	for _, field := range fields {
		parts = append(parts, fmt.Sprintf("%s: %s", jqKey(field), jqPath(field)))
	}
	expr := strings.Join(parts, ", ")
	return fmt.Sprintf(
		`if type=="object" and has("items") and (.items|type)=="array" then .items |= map({%s}) | . else if type=="array" then map({%s}) else {%s} end end`,
		expr, expr, expr,
	)
}

func jqKey(key string) string {
	escaped := strings.ReplaceAll(key, "\"", "\\\"")
	return fmt.Sprintf("\"%s\"", escaped)
}

func jqPath(path string) string {
	segments := strings.Split(path, ".")
	expr := ""
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		escaped := strings.ReplaceAll(seg, "\"", "\\\"")
		expr += fmt.Sprintf("[\"%s\"]", escaped)
	}
	if expr == "" {
		return "."
	}
	return "." + expr
}
