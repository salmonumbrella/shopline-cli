// Package env provides environment variable helpers.
package env

import (
	"os"
	"strings"
)

// Bool returns true if the environment variable key is set to a truthy value.
// Truthy values are: "1", "true", "yes", "on" (case-insensitive).
// All other values, including unset variables, return false.
func Bool(key string) bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch val {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
