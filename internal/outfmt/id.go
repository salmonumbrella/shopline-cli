package outfmt

import (
	"fmt"
	"regexp"
)

// FormatID formats a resource ID for agent-friendly output.
// Example: FormatID("order", "12345") returns "[order:$12345]"
func FormatID(prefix, id string) string {
	return fmt.Sprintf("[%s:$%s]", prefix, id)
}

var idPattern = regexp.MustCompile(`^\[([a-zA-Z0-9_-]+):\$(.+)\]$`)

// ParseID extracts prefix and ID from formatted string.
// Returns ("", "", false) if input doesn't match the format.
func ParseID(s string) (prefix, id string, ok bool) {
	matches := idPattern.FindStringSubmatch(s)
	if len(matches) != 3 {
		return "", "", false
	}
	return matches[1], matches[2], true
}
