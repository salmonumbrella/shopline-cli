package coverage

import (
	"regexp"
	"strings"
)

func NormalizeMethod(method string) string {
	m := strings.TrimSpace(strings.ToUpper(method))
	switch m {
	case "DEL":
		return "DELETE"
	default:
		return m
	}
}

// NormalizePath normalizes paths for comparison.
// It strips query strings and normalizes "{anything}" to "{...}"? Actually, it
// keeps braces as-is but removes query strings.
func NormalizePath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return p
	}
	// Firecrawl markdown can include backslash escapes (e.g. "\_") that should not
	// be part of the actual path.
	p = strings.ReplaceAll(p, `\_`, `_`)
	p = strings.ReplaceAll(p, `\{`, `{`)
	p = strings.ReplaceAll(p, `\}`, `}`)
	p = strings.ReplaceAll(p, `\/`, `/`)
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// Normalize any named placeholders to "{}" so docs ("{id}") and code ("{}")
	// compare cleanly.
	p = rePathPlaceholder.ReplaceAllString(p, "{}")
	return p
}

var rePathPlaceholder = regexp.MustCompile(`\{[^}]*\}`)
