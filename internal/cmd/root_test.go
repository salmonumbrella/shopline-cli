package cmd

import (
	"bytes"
	"os"
	"testing"
)

func TestGetDefaultOutput(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   string
	}{
		{
			name:   "default when unset",
			envVal: "",
			want:   "text",
		},
		{
			name:   "json when set to json",
			envVal: "json",
			want:   "json",
		},
		{
			name:   "text when set to text",
			envVal: "text",
			want:   "text",
		},
		{
			name:   "default when invalid value",
			envVal: "invalid",
			want:   "text",
		},
		{
			name:   "default when set to yaml",
			envVal: "yaml",
			want:   "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original env
			orig := os.Getenv("SHOPLINE_OUTPUT")
			defer func() { _ = os.Setenv("SHOPLINE_OUTPUT", orig) }()

			if tt.envVal == "" {
				_ = os.Unsetenv("SHOPLINE_OUTPUT")
			} else {
				_ = os.Setenv("SHOPLINE_OUTPUT", tt.envVal)
			}

			got := getDefaultOutput()
			if got != tt.want {
				t.Errorf("getDefaultOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test args to show version
	os.Args = []string{"shopline", "--version"}

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := Execute("1.0.0", "abc123", "2024-01-01")
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Reset the root command for other tests
	rootCmd.SetArgs([]string{})
}
