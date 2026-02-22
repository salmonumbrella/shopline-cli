package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
			name:   "jsonl when set to jsonl",
			envVal: "jsonl",
			want:   "jsonl",
		},
		{
			name:   "ndjson when set to ndjson",
			envVal: "ndjson",
			want:   "ndjson",
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

func TestNormalizeOutputFlagFromEnvJSONLines(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string
	}{
		{name: "jsonl env keeps requested mode", env: "jsonl", want: "jsonl"},
		{name: "ndjson env keeps requested mode", env: "ndjson", want: "ndjson"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := os.Getenv("SHOPLINE_OUTPUT")
			defer func() { _ = os.Setenv("SHOPLINE_OUTPUT", orig) }()
			_ = os.Setenv("SHOPLINE_OUTPUT", tt.env)

			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("output", getDefaultOutput(), "")
			cmd.Flags().String(outputModeFlagName, "text", "")
			_ = cmd.Flags().MarkHidden(outputModeFlagName)

			if err := normalizeOutputFlag(cmd); err != nil {
				t.Fatalf("normalizeOutputFlag returned error: %v", err)
			}

			out, _ := cmd.Flags().GetString("output")
			if out != "json" {
				t.Fatalf("output = %q, want %q", out, "json")
			}
			mode, _ := cmd.Flags().GetString(outputModeFlagName)
			if mode != tt.want {
				t.Fatalf("output mode = %q, want %q", mode, tt.want)
			}
		})
	}
}

func TestPreRunApplyNonInteractive_ResultsOnlySyncsToItemsOnly(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var itemsOnly bool
	cmd.PersistentFlags().BoolVar(&itemsOnly, "items-only", false, "")
	cmd.PersistentFlags().BoolVar(&itemsOnly, "results-only", false, "")
	cmd.PersistentFlags().BoolVarP(&rootYes, "yes", "y", false, "")
	cmd.PersistentFlags().BoolVar(&rootYes, "force", false, "")

	if err := cmd.ParseFlags([]string{"--results-only"}); err != nil {
		t.Fatal(err)
	}

	if err := preRunApplyNonInteractive(cmd, nil); err != nil {
		t.Fatal(err)
	}

	if !cmd.Flags().Changed("items-only") {
		t.Error("expected --items-only to be marked Changed when --results-only is set")
	}
}

func TestDirectAccessTokenFromEnv_ExcludesAdminToken(t *testing.T) {
	for _, env := range []string{"SHOPLINE_ACCESS_TOKEN", "SHOPLINE_API_TOKEN", "SHOPLINE_TOKEN"} {
		t.Setenv(env, "")
	}
	t.Setenv("SHOPLINE_ADMIN_TOKEN", "admin-only-token")

	result := directAccessTokenFromEnv()
	if result != "" {
		t.Errorf("expected empty token when only SHOPLINE_ADMIN_TOKEN is set, got %q", result)
	}
}

func TestExecute(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test args to show version
	os.Args = []string{"spl", "--version"}

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

func TestRootHelpShowsStaticText(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	setupRootCommand()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"spl - CLI for Shopline",
		"Aliases (resource",
		"Orders:",
		"Exit codes:",
		"Environment:",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("root help missing %q", want)
		}
	}
	// Should NOT contain Cobra's default help markers
	if strings.Contains(output, "Available Commands") {
		t.Error("root help should use static text, not Cobra default")
	}

	rootCmd.SetArgs([]string{})
}

func TestSubcommandHelpUsesCobra(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"orders", "--help"})
	setupRootCommand()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Available Commands") {
		t.Error("subcommand help should use Cobra default (expected 'Available Commands')")
	}
	// Should NOT contain the static help.txt header
	if strings.Contains(output, "spl - CLI for Shopline") {
		t.Error("subcommand help should not show static root help text")
	}

	rootCmd.SetArgs([]string{})
}

func TestReadQueryFile_RejectsLargeFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "big-query-*.jq")
	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, maxQueryFileSize+1)
	for i := range data {
		data[i] = '.'
	}
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	cmd := &cobra.Command{Use: "test"}
	_, err = readQueryFile(cmd, f.Name())
	if err == nil {
		t.Fatal("expected error for oversized query file")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' in error, got: %v", err)
	}
}
