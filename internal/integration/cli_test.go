package integration

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// runCLI executes the CLI with the given arguments and returns the combined output.
func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmdArgs := append([]string{"run", "../../cmd/spl/main.go"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestCLIVersion(t *testing.T) {
	output, err := runCLI(t, "--version")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}
	if !bytes.Contains([]byte(output), []byte("spl")) {
		t.Errorf("expected version output to contain 'spl', got: %s", output)
	}
	// Version should include dev or version info
	if !strings.Contains(output, "version") && !strings.Contains(output, "dev") {
		t.Errorf("expected version info in output, got: %s", output)
	}
}

func TestCLIHelp(t *testing.T) {
	output, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain static help.txt content
	if !strings.Contains(output, "spl - CLI for Shopline") {
		t.Errorf("expected 'spl - CLI for Shopline' in help output, got: %s", output)
	}

	// Should contain key sections from help.txt
	for _, want := range []string{
		"Aliases (resource",
		"Orders:",
		"Exit codes:",
		"Environment:",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in help output", want)
		}
	}

	// Root help should NOT contain Cobra default markers
	if strings.Contains(output, "Available Commands") {
		t.Error("root help should use static text, not Cobra default")
	}

	// Should contain key resource sections from static help.txt
	for _, section := range []string{"Orders:", "Products:", "Customers:", "Auth:"} {
		if !strings.Contains(output, section) {
			t.Errorf("expected section %q in help output", section)
		}
	}
}

func TestCLICustomersHelp(t *testing.T) {
	output, err := runCLI(t, "customers", "--help")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain customers description
	if !strings.Contains(output, "customers") {
		t.Errorf("expected 'customers' in help output, got: %s", output)
	}

	// Should list subcommands
	expectedSubcommands := []string{"list", "get"}
	for _, subcmd := range expectedSubcommands {
		if !strings.Contains(output, subcmd) {
			t.Errorf("expected subcommand '%s' in customers help output, got: %s", subcmd, output)
		}
	}
}

func TestCLIProductsHelp(t *testing.T) {
	output, err := runCLI(t, "products", "--help")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain products description
	if !strings.Contains(output, "products") {
		t.Errorf("expected 'products' in help output, got: %s", output)
	}

	// Should list subcommands
	expectedSubcommands := []string{"list", "get"}
	for _, subcmd := range expectedSubcommands {
		if !strings.Contains(output, subcmd) {
			t.Errorf("expected subcommand '%s' in products help output, got: %s", subcmd, output)
		}
	}
}

func TestCLIOrdersHelp(t *testing.T) {
	output, err := runCLI(t, "orders", "--help")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain orders description
	if !strings.Contains(output, "orders") {
		t.Errorf("expected 'orders' in help output, got: %s", output)
	}

	// Should list subcommands
	expectedSubcommands := []string{"list", "get"}
	for _, subcmd := range expectedSubcommands {
		if !strings.Contains(output, subcmd) {
			t.Errorf("expected subcommand '%s' in orders help output, got: %s", subcmd, output)
		}
	}
}

func TestCLICompletionBash(t *testing.T) {
	output, err := runCLI(t, "completion", "bash")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain bash completion script markers
	if !strings.Contains(output, "bash") {
		t.Errorf("expected 'bash' in completion output, got: %s", output)
	}

	// Should contain function definitions for completions
	if !strings.Contains(output, "spl") {
		t.Errorf("expected 'spl' in bash completion script, got: %s", output)
	}

	// Should contain completion function markers
	if !strings.Contains(output, "__") {
		t.Errorf("expected completion function markers in bash completion script, got: %s", output)
	}
}

func TestCLICompletionZsh(t *testing.T) {
	output, err := runCLI(t, "completion", "zsh")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain zsh completion script markers
	if !strings.Contains(output, "compdef") || !strings.Contains(output, "spl") {
		t.Errorf("expected zsh completion script with 'compdef' and 'spl', got: %s", output)
	}
}

func TestCLICompletionFish(t *testing.T) {
	output, err := runCLI(t, "completion", "fish")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain fish completion script markers
	if !strings.Contains(output, "complete") && !strings.Contains(output, "spl") {
		t.Errorf("expected fish completion script with 'complete' command, got: %s", output)
	}
}

func TestCLICompletionPowershell(t *testing.T) {
	output, err := runCLI(t, "completion", "powershell")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain powershell completion script markers
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Errorf("expected PowerShell completion script with 'Register-ArgumentCompleter', got: %s", output)
	}
}

func TestCLIUnknownCommand(t *testing.T) {
	output, err := runCLI(t, "unknown-command")

	// Should fail with an error
	if err == nil {
		t.Errorf("expected error for unknown command, got success with output: %s", output)
	}

	// Should contain error message about unknown command
	if !strings.Contains(output, "unknown command") {
		t.Errorf("expected 'unknown command' in error output, got: %s", output)
	}
}

func TestCLIGlobalFlags(t *testing.T) {
	output, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Static help.txt should document common flags and environment
	for _, want := range []string{
		"-s STORE",
		"-l LIMIT",
		"--help",
		"SHOPLINE_STORE",
		"SHOPLINE_ACCESS_TOKEN",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in help output", want)
		}
	}
}

func TestCLICompletionInvalidShell(t *testing.T) {
	output, err := runCLI(t, "completion", "invalid-shell")

	// Should fail with an error
	if err == nil {
		t.Errorf("expected error for invalid shell, got success with output: %s", output)
	}

	// Should mention valid shells
	if !strings.Contains(output, "bash") || !strings.Contains(output, "zsh") {
		t.Errorf("expected valid shells mentioned in error output, got: %s", output)
	}
}

func TestCLILegacyProfileListAliases(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "config ls", args: []string{"config", "ls"}},
		{name: "store ls", args: []string{"store", "ls"}},
		{name: "profile ls", args: []string{"profile", "ls"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(t, tt.args...)
			if err != nil {
				t.Fatalf("expected legacy alias to work, got error: %v\nOutput: %s", err, output)
			}
			// Command should either list profiles or show setup guidance on empty keyring.
			if !strings.Contains(output, "No store profiles configured") && !strings.Contains(output, "NAME") {
				t.Errorf("unexpected output for %v: %s", tt.args, output)
			}
		})
	}
}
