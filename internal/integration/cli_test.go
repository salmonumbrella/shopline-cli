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
	cmdArgs := append([]string{"run", "../../cmd/shopline/main.go"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestCLIVersion(t *testing.T) {
	output, err := runCLI(t, "--version")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}
	if !bytes.Contains([]byte(output), []byte("shopline")) {
		t.Errorf("expected version output to contain 'shopline', got: %s", output)
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

	// Should contain CLI description
	if !strings.Contains(output, "Shopline") {
		t.Errorf("expected 'Shopline' in help output, got: %s", output)
	}

	// Should contain available commands section
	if !strings.Contains(output, "Available Commands") {
		t.Errorf("expected 'Available Commands' in help output, got: %s", output)
	}

	// Should contain flags section
	if !strings.Contains(output, "Flags") {
		t.Errorf("expected 'Flags' in help output, got: %s", output)
	}

	// Should contain some known commands
	expectedCommands := []string{"customers", "products", "orders", "completion"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("expected command '%s' in help output, got: %s", cmd, output)
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
	if !strings.Contains(output, "shopline") {
		t.Errorf("expected 'shopline' in bash completion script, got: %s", output)
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
	if !strings.Contains(output, "compdef") || !strings.Contains(output, "shopline") {
		t.Errorf("expected zsh completion script with 'compdef' and 'shopline', got: %s", output)
	}
}

func TestCLICompletionFish(t *testing.T) {
	output, err := runCLI(t, "completion", "fish")
	if err != nil {
		t.Fatalf("failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Should contain fish completion script markers
	if !strings.Contains(output, "complete") && !strings.Contains(output, "shopline") {
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

	// Should contain global flags
	expectedFlags := []string{
		"--store",
		"--output",
		"--help",
		"--version",
	}
	for _, flag := range expectedFlags {
		if !strings.Contains(output, flag) {
			t.Errorf("expected global flag '%s' in help output, got: %s", flag, output)
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
