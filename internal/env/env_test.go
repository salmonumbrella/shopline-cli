package env

import (
	"os"
	"testing"
)

func TestBool(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty string", "", false},
		{"1", "1", true},
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"True", "True", true},
		{"yes", "yes", true},
		{"YES", "YES", true},
		{"on", "on", true},
		{"ON", "ON", true},
		{"0", "0", false},
		{"false", "false", false},
		{"FALSE", "FALSE", false},
		{"no", "no", false},
		{"off", "off", false},
		{"random", "random", false},
		{"whitespace true", "  true  ", true},
		{"whitespace 1", " 1 ", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Setenv("TEST_ENV_BOOL", tc.value)
			defer func() { _ = os.Unsetenv("TEST_ENV_BOOL") }()

			result := Bool("TEST_ENV_BOOL")
			if result != tc.expected {
				t.Errorf("Bool(%q) = %v, want %v", tc.value, result, tc.expected)
			}
		})
	}
}

func TestBoolUnset(t *testing.T) {
	_ = os.Unsetenv("TEST_ENV_BOOL_UNSET")

	result := Bool("TEST_ENV_BOOL_UNSET")
	if result != false {
		t.Errorf("Bool for unset key = %v, want false", result)
	}
}
