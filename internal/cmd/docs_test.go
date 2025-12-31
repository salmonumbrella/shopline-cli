package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDocsManCommand(t *testing.T) {
	// Create a temporary directory for output
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "man")

	// Set up command
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"docs", "man", outputDir})

	// Execute
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("docs man command failed: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory was not created")
	}

	// Verify at least one man page was generated
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}
	if len(files) == 0 {
		t.Error("no man pages were generated")
	}

	// Verify output message
	output := buf.String()
	if output == "" {
		t.Error("expected output message but got none")
	}

	// Reset for other tests
	rootCmd.SetArgs([]string{})
}

func TestDocsManCommandDefaultDir(t *testing.T) {
	// Use a temp directory as working directory
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Set up command without output dir argument
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"docs", "man"})

	// Execute
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("docs man command failed: %v", err)
	}

	// Verify default directory was created
	defaultDir := filepath.Join(tmpDir, "man")
	if _, err := os.Stat(defaultDir); os.IsNotExist(err) {
		t.Error("default man directory was not created")
	}

	// Reset for other tests
	rootCmd.SetArgs([]string{})
}

func TestDocsMarkdownCommand(t *testing.T) {
	// Create a temporary directory for output
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "docs", "cli")

	// Set up command
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"docs", "markdown", outputDir})

	// Execute
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("docs markdown command failed: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory was not created")
	}

	// Verify at least one markdown file was generated
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}
	if len(files) == 0 {
		t.Error("no markdown files were generated")
	}

	// Verify a markdown file exists with .md extension
	foundMd := false
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".md" {
			foundMd = true
			break
		}
	}
	if !foundMd {
		t.Error("no .md files were generated")
	}

	// Verify output message
	output := buf.String()
	if output == "" {
		t.Error("expected output message but got none")
	}

	// Reset for other tests
	rootCmd.SetArgs([]string{})
}

func TestDocsMarkdownCommandDefaultDir(t *testing.T) {
	// Use a temp directory as working directory
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Set up command without output dir argument
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"docs", "markdown"})

	// Execute
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("docs markdown command failed: %v", err)
	}

	// Verify default directory was created
	defaultDir := filepath.Join(tmpDir, "docs", "cli")
	if _, err := os.Stat(defaultDir); os.IsNotExist(err) {
		t.Error("default docs/cli directory was not created")
	}

	// Reset for other tests
	rootCmd.SetArgs([]string{})
}

func TestDocsCommandHidden(t *testing.T) {
	if !docsCmd.Hidden {
		t.Error("docs command should be hidden")
	}
}
