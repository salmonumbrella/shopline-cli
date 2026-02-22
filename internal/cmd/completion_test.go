package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCompletionBash(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"spl", "completion", "bash"}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"completion", "bash"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion bash returned error: %v", err)
	}

	output := buf.String()
	// Bash completion scripts contain these markers
	if !strings.Contains(output, "bash completion") || !strings.Contains(output, "__start_spl") {
		t.Errorf("completion bash output does not contain expected bash completion markers")
	}
}

func TestCompletionZsh(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"spl", "completion", "zsh"}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"completion", "zsh"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion zsh returned error: %v", err)
	}

	output := buf.String()
	// Zsh completion scripts start with #compdef
	if !strings.Contains(output, "#compdef") {
		t.Errorf("completion zsh output does not contain #compdef marker")
	}
}

func TestCompletionFish(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"spl", "completion", "fish"}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"completion", "fish"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion fish returned error: %v", err)
	}

	output := buf.String()
	// Fish completion scripts use "complete -c <command>"
	if !strings.Contains(output, "complete -c spl") {
		t.Errorf("completion fish output does not contain 'complete -c spl' marker")
	}
}

func TestCompletionPowerShell(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"spl", "completion", "powershell"}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"completion", "powershell"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion powershell returned error: %v", err)
	}

	output := buf.String()
	// PowerShell completion scripts contain Register-ArgumentCompleter
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Errorf("completion powershell output does not contain 'Register-ArgumentCompleter' marker")
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"spl", "completion", "invalid"}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"completion", "invalid"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("completion invalid should return error")
	}

	// The error should mention invalid argument
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error message should mention 'invalid', got: %v", err)
	}
}

func TestCompletionNoArgs(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"spl", "completion"}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"completion"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("completion with no args should return error")
	}
}
