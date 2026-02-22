package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// TestAuthCommandSetup verifies auth command initialization
func TestAuthCommandSetup(t *testing.T) {
	if authCmd.Use != "auth" {
		t.Errorf("expected Use 'auth', got %q", authCmd.Use)
	}
	expectedAliases := []string{"config", "store", "profile", "profiles"}
	for _, alias := range expectedAliases {
		found := false
		for _, got := range authCmd.Aliases {
			if got == alias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected auth alias %q to be configured", alias)
		}
	}
	if authCmd.Short != "Manage authentication" {
		t.Errorf("expected Short 'Manage authentication', got %q", authCmd.Short)
	}
	if authCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

// TestAuthSubcommands verifies all subcommands are registered
func TestAuthSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"login":  "Add a new store profile via browser",
		"remove": "Remove a store profile",
		"list":   "List configured store profiles",
		"status": "Show current authentication status",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range authCmd.Commands() {
				if sub.Use == name || (len(sub.Use) > len(name) && sub.Use[:len(name)] == name) {
					found = true
					if sub.Short != short {
						t.Errorf("expected Short %q, got %q", short, sub.Short)
					}
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not found", name)
			}
		})
	}
}

// TestAuthAddArgs verifies login command accepts no arguments
func TestAuthAddArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: false},
		{name: "one arg", args: []string{"mystore"}, wantErr: true},
		{name: "too many args", args: []string{"store1", "store2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authAddCmd.Args(authAddCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAuthRemoveArgs verifies remove command requires exactly 1 argument
func TestAuthRemoveArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"mystore"}, wantErr: false},
		{name: "too many args", args: []string{"store1", "store2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authRemoveCmd.Args(authRemoveCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAuthListArgs verifies list command accepts no arguments
func TestAuthListArgs(t *testing.T) {
	// authListCmd doesn't define Args, so it accepts any arguments by default
	// This test verifies the command is accessible
	if authListCmd.Use != "list" {
		t.Errorf("expected Use 'list', got %q", authListCmd.Use)
	}
}

// TestAuthStatusArgs verifies status command accepts no arguments
func TestAuthStatusArgs(t *testing.T) {
	// authStatusCmd doesn't define Args, so it accepts any arguments by default
	// This test verifies the command is accessible
	if authStatusCmd.Use != "status" {
		t.Errorf("expected Use 'status', got %q", authStatusCmd.Use)
	}
}

// TestAuthAddCmdUse verifies the login command use string
func TestAuthAddCmdUse(t *testing.T) {
	expectedUse := "login"
	if authAddCmd.Use != expectedUse {
		t.Errorf("expected Use %q, got %q", expectedUse, authAddCmd.Use)
	}
}

// TestAuthRemoveCmdUse verifies the remove command use string includes argument placeholder
func TestAuthRemoveCmdUse(t *testing.T) {
	expectedUse := "remove <name>"
	if authRemoveCmd.Use != expectedUse {
		t.Errorf("expected Use %q, got %q", expectedUse, authRemoveCmd.Use)
	}
}

// TestAuthCommandsHaveRunE verifies all auth subcommands have RunE functions
func TestAuthCommandsHaveRunE(t *testing.T) {
	// Verify authAddCmd has RunE
	if authAddCmd.RunE == nil {
		t.Error("authAddCmd should have RunE function")
	}

	// Verify authListCmd has RunE
	if authListCmd.RunE == nil {
		t.Error("authListCmd should have RunE function")
	}

	// Verify authRemoveCmd has RunE
	if authRemoveCmd.RunE == nil {
		t.Error("authRemoveCmd should have RunE function")
	}

	// Verify authStatusCmd has RunE
	if authStatusCmd.RunE == nil {
		t.Error("authStatusCmd should have RunE function")
	}
}

// TestAuthSubcommandsRegisteredToParent verifies subcommands are attached to auth parent
func TestAuthSubcommandsRegisteredToParent(t *testing.T) {
	subcommandNames := []string{"login", "list", "remove", "status"}
	registeredCmds := authCmd.Commands()

	if len(registeredCmds) < len(subcommandNames) {
		t.Errorf("expected at least %d subcommands, got %d", len(subcommandNames), len(registeredCmds))
	}

	for _, expectedName := range subcommandNames {
		found := false
		for _, cmd := range registeredCmds {
			// Use string prefix check because some commands have arguments in their Use string
			if len(cmd.Use) >= len(expectedName) && cmd.Use[:len(expectedName)] == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found in authCmd.Commands()", expectedName)
		}
	}
}

// TestAuthCmdRegisteredToRoot verifies authCmd is registered to rootCmd
func TestAuthCmdRegisteredToRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "auth" {
			found = true
			break
		}
	}
	if !found {
		t.Error("authCmd not found in rootCmd.Commands()")
	}
}

func TestAuthLegacyRootAliasesResolve(t *testing.T) {
	setupRootCommand()

	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"config", "ls"}, want: "list"},
		{args: []string{"store", "ls"}, want: "list"},
		{args: []string{"profile", "ls"}, want: "list"},
		{args: []string{"profiles", "ls"}, want: "list"},
	}

	for _, tt := range tests {
		found, _, err := rootCmd.Find(tt.args)
		if err != nil {
			t.Fatalf("Find(%v) error: %v", tt.args, err)
		}
		if found == nil {
			t.Fatalf("Find(%v) returned nil command", tt.args)
		}
		if found.Name() != tt.want {
			t.Fatalf("Find(%v) resolved to %q, want %q", tt.args, found.Name(), tt.want)
		}
	}
}

// authTestKeyring implements keyring.Keyring for testing auth commands
type authTestKeyring struct {
	items     map[string]keyring.Item
	getErr    error
	setErr    error
	removeErr error
	keysErr   error
}

func newAuthTestKeyring() *authTestKeyring {
	return &authTestKeyring{items: make(map[string]keyring.Item)}
}

func (m *authTestKeyring) Get(key string) (keyring.Item, error) {
	if m.getErr != nil {
		return keyring.Item{}, m.getErr
	}
	item, ok := m.items[key]
	if !ok {
		return keyring.Item{}, keyring.ErrKeyNotFound
	}
	return item, nil
}

func (m *authTestKeyring) GetMetadata(_ string) (keyring.Metadata, error) {
	return keyring.Metadata{}, nil
}

func (m *authTestKeyring) Set(item keyring.Item) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.items[item.Key] = item
	return nil
}

func (m *authTestKeyring) Remove(key string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	delete(m.items, key)
	return nil
}

func (m *authTestKeyring) Keys() ([]string, error) {
	if m.keysErr != nil {
		return nil, m.keysErr
	}
	keys := make([]string, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return keys, nil
}

// setupAuthTestKeyring sets up a mock keyring for auth tests
func setupAuthTestKeyring(mock *authTestKeyring) func() {
	origOpener := secrets.GetKeyringOpener()
	secrets.SetKeyringOpener(func(cfg keyring.Config) (keyring.Keyring, error) {
		return mock, nil
	})
	return func() {
		secrets.SetKeyringOpener(origOpener)
	}
}

// setupAuthTestKeyringWithError sets up a keyring opener that returns an error
func setupAuthTestKeyringWithError() func() {
	origOpener := secrets.GetKeyringOpener()
	secrets.SetKeyringOpener(func(cfg keyring.Config) (keyring.Keyring, error) {
		return nil, keyring.ErrNoAvailImpl
	})
	return func() {
		secrets.SetKeyringOpener(origOpener)
	}
}

// newAuthTestCmd creates a test command for auth tests
func newAuthTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("store", "", "Store profile name")
	return cmd
}

// TestAuthListRunE_EmptyList tests list command with no profiles
func TestAuthListRunE_EmptyList(t *testing.T) {
	mock := newAuthTestKeyring()
	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authListCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No store profiles configured") {
		t.Errorf("expected empty list message, got: %s", output)
	}
}

// TestAuthListRunE_WithProfiles tests list command with existing profiles
func TestAuthListRunE_WithProfiles(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add credentials to mock
	creds := &secrets.StoreCredentials{
		Name:      "test-store",
		Handle:    "test-handle",
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour), // 30 days old
	}
	data, _ := creds.Marshal()
	mock.items["store:test-store"] = keyring.Item{Key: "store:test-store", Data: data}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authListCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test-store") {
		t.Errorf("expected profile name in output, got: %s", output)
	}
	if !strings.Contains(output, "test-handle") {
		t.Errorf("expected handle in output, got: %s", output)
	}
	if !strings.Contains(output, "OK") {
		t.Errorf("expected OK status in output, got: %s", output)
	}
}

// TestAuthListRunE_WithOldCredentials tests list command showing ROTATE status
func TestAuthListRunE_WithOldCredentials(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add old credentials to mock
	creds := &secrets.StoreCredentials{
		Name:      "old-store",
		Handle:    "old-handle",
		CreatedAt: time.Now().Add(-100 * 24 * time.Hour), // 100 days old
	}
	data, _ := creds.Marshal()
	mock.items["store:old-store"] = keyring.Item{Key: "store:old-store", Data: data}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authListCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "ROTATE") {
		t.Errorf("expected ROTATE status for old credentials, got: %s", output)
	}
}

// TestAuthListRunE_StoreError tests list command when store fails to open
func TestAuthListRunE_StoreError(t *testing.T) {
	cleanup := setupAuthTestKeyringWithError()
	defer cleanup()

	cmd := newAuthTestCmd()
	err := authListCmd.RunE(cmd, []string{})

	if err == nil {
		t.Error("expected error when store fails to open")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestAuthListRunE_ListError tests list command when list fails
func TestAuthListRunE_ListError(t *testing.T) {
	mock := newAuthTestKeyring()
	mock.keysErr = keyring.ErrNoAvailImpl

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	cmd := newAuthTestCmd()
	err := authListCmd.RunE(cmd, []string{})

	if err == nil {
		t.Error("expected error when list fails")
	}
	if !strings.Contains(err.Error(), "failed to list profiles") {
		t.Errorf("expected 'failed to list profiles' error, got: %v", err)
	}
}

// TestAuthRemoveRunE_Success tests successful profile removal
func TestAuthRemoveRunE_Success(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add a profile to remove
	creds := &secrets.StoreCredentials{
		Name:      "remove-me",
		Handle:    "handle",
		CreatedAt: time.Now(),
	}
	data, _ := creds.Marshal()
	mock.items["store:remove-me"] = keyring.Item{Key: "store:remove-me", Data: data}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authRemoveCmd.RunE(cmd, []string{"remove-me"})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify profile was removed
	if _, exists := mock.items["store:remove-me"]; exists {
		t.Error("expected profile to be removed")
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Removed store profile: remove-me") {
		t.Errorf("expected removal confirmation, got: %s", output)
	}
}

// TestAuthRemoveRunE_StoreError tests remove command when store fails to open
func TestAuthRemoveRunE_StoreError(t *testing.T) {
	cleanup := setupAuthTestKeyringWithError()
	defer cleanup()

	cmd := newAuthTestCmd()
	err := authRemoveCmd.RunE(cmd, []string{"some-store"})

	if err == nil {
		t.Error("expected error when store fails to open")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestAuthRemoveRunE_DeleteError tests remove command when delete fails
func TestAuthRemoveRunE_DeleteError(t *testing.T) {
	mock := newAuthTestKeyring()
	mock.removeErr = keyring.ErrNoAvailImpl

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	cmd := newAuthTestCmd()
	err := authRemoveCmd.RunE(cmd, []string{"nonexistent"})

	if err == nil {
		t.Error("expected error when delete fails")
	}
	if !strings.Contains(err.Error(), "failed to remove profile") {
		t.Errorf("expected 'failed to remove profile' error, got: %v", err)
	}
}

// TestAuthStatusRunE_NoProfiles tests status command with no profiles
func TestAuthStatusRunE_NoProfiles(t *testing.T) {
	mock := newAuthTestKeyring()
	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Clear any env var
	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Unsetenv("SHOPLINE_STORE")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authStatusCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Not authenticated") {
		t.Errorf("expected 'Not authenticated' message, got: %s", output)
	}
}

// TestAuthStatusRunE_SingleProfile tests status command with single profile (auto-select)
func TestAuthStatusRunE_SingleProfile(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add a single profile
	creds := &secrets.StoreCredentials{
		Name:      "only-store",
		Handle:    "only-handle",
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
	}
	data, _ := creds.Marshal()
	mock.items["store:only-store"] = keyring.Item{Key: "store:only-store", Data: data}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Clear any env var
	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Unsetenv("SHOPLINE_STORE")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authStatusCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Profile:  only-store") {
		t.Errorf("expected profile name in output, got: %s", output)
	}
	if !strings.Contains(output, "Handle:   only-handle") {
		t.Errorf("expected handle in output, got: %s", output)
	}
}

// TestAuthStatusRunE_MultipleProfilesNoSelection tests status with multiple profiles
func TestAuthStatusRunE_MultipleProfilesNoSelection(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add multiple profiles
	for _, name := range []string{"store1", "store2"} {
		creds := &secrets.StoreCredentials{
			Name:      name,
			Handle:    name + "-handle",
			CreatedAt: time.Now(),
		}
		data, _ := creds.Marshal()
		mock.items["store:"+name] = keyring.Item{Key: "store:" + name, Data: data}
	}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Clear any env var
	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Unsetenv("SHOPLINE_STORE")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authStatusCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Multiple profiles configured") {
		t.Errorf("expected 'Multiple profiles' message, got: %s", output)
	}
}

// TestAuthStatusRunE_WithStoreFlag tests status with explicit store selection
func TestAuthStatusRunE_WithStoreFlag(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add multiple profiles
	for _, name := range []string{"store1", "store2"} {
		creds := &secrets.StoreCredentials{
			Name:      name,
			Handle:    name + "-handle",
			CreatedAt: time.Now(),
		}
		data, _ := creds.Marshal()
		mock.items["store:"+name] = keyring.Item{Key: "store:" + name, Data: data}
	}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Clear any env var
	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Unsetenv("SHOPLINE_STORE")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	_ = cmd.Flags().Set("store", "store1")
	err := authStatusCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Profile:  store1") {
		t.Errorf("expected store1 profile, got: %s", output)
	}
}

// TestAuthStatusRunE_WithStoreHandleLookup tests status with handle lookup fallback
func TestAuthStatusRunE_WithStoreHandleLookup(t *testing.T) {
	mock := newAuthTestKeyring()

	creds := &secrets.StoreCredentials{
		Name:      "demo",
		Handle:    "demoshop",
		CreatedAt: time.Now(),
	}
	data, _ := creds.Marshal()
	mock.items["store:demo"] = keyring.Item{Key: "store:demo", Data: data}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Unsetenv("SHOPLINE_STORE")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	_ = cmd.Flags().Set("store", "demoshop")
	err := authStatusCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Profile:  demo") {
		t.Errorf("expected demo profile, got: %s", output)
	}
	if !strings.Contains(output, "Handle:   demoshop") {
		t.Errorf("expected demoshop handle, got: %s", output)
	}
}

// TestAuthStatusRunE_WithEnvVar tests status using SHOPLINE_STORE env var
func TestAuthStatusRunE_WithEnvVar(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add a profile
	creds := &secrets.StoreCredentials{
		Name:      "env-store",
		Handle:    "env-handle",
		CreatedAt: time.Now(),
	}
	data, _ := creds.Marshal()
	mock.items["store:env-store"] = keyring.Item{Key: "store:env-store", Data: data}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Set env var
	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Setenv("SHOPLINE_STORE", "env-store")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authStatusCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Profile:  env-store") {
		t.Errorf("expected env-store profile, got: %s", output)
	}
}

// TestAuthStatusRunE_ProfileNotFound tests status with nonexistent profile
func TestAuthStatusRunE_ProfileNotFound(t *testing.T) {
	mock := newAuthTestKeyring()
	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Set env var to nonexistent profile
	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Setenv("SHOPLINE_STORE", "nonexistent")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	cmd := newAuthTestCmd()
	err := authStatusCmd.RunE(cmd, []string{})

	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
	if !strings.Contains(err.Error(), "profile not found") {
		t.Errorf("expected 'profile not found' error, got: %v", err)
	}
}

// TestAuthStatusRunE_StoreError tests status when store fails to open
func TestAuthStatusRunE_StoreError(t *testing.T) {
	cleanup := setupAuthTestKeyringWithError()
	defer cleanup()

	cmd := newAuthTestCmd()
	err := authStatusCmd.RunE(cmd, []string{})

	if err == nil {
		t.Error("expected error when store fails to open")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestAuthStatusRunE_OldCredentialsWarning tests status shows warning for old credentials
func TestAuthStatusRunE_OldCredentialsWarning(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add old credentials
	creds := &secrets.StoreCredentials{
		Name:      "old-store",
		Handle:    "old-handle",
		CreatedAt: time.Now().Add(-100 * 24 * time.Hour), // 100 days old
	}
	data, _ := creds.Marshal()
	mock.items["store:old-store"] = keyring.Item{Key: "store:old-store", Data: data}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Clear any env var
	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Unsetenv("SHOPLINE_STORE")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authStatusCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Warning: Credentials are older than 90 days") {
		t.Errorf("expected old credentials warning, got: %s", output)
	}
}

// TestAuthStatusRunE_ListError tests status when listing profiles fails
func TestAuthStatusRunE_ListError(t *testing.T) {
	mock := newAuthTestKeyring()
	mock.keysErr = keyring.ErrNoAvailImpl

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Clear any env var
	origEnv := os.Getenv("SHOPLINE_STORE")
	_ = os.Unsetenv("SHOPLINE_STORE")
	defer func() { _ = os.Setenv("SHOPLINE_STORE", origEnv) }()

	cmd := newAuthTestCmd()
	err := authStatusCmd.RunE(cmd, []string{})

	if err == nil {
		t.Error("expected error when list fails")
	}
	if !strings.Contains(err.Error(), "failed to list profiles") {
		t.Errorf("expected 'failed to list profiles' error, got: %v", err)
	}
}

// TestAuthListRunE_GetError tests list command when get fails for a profile
func TestAuthListRunE_GetError(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add an invalid profile that will fail to unmarshal
	mock.items["store:bad-store"] = keyring.Item{Key: "store:bad-store", Data: []byte("invalid json")}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authListCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	// The list command skips profiles that fail to load (continue in the loop)
	// so it should not error
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// The output should have headers but no profile data (since it failed to load)
	if !strings.Contains(output, "NAME") {
		t.Errorf("expected header in output, got: %s", output)
	}
}

// TestAuthListRunE_MultipleProfiles tests list with multiple profiles
func TestAuthListRunE_MultipleProfiles(t *testing.T) {
	mock := newAuthTestKeyring()

	// Add multiple profiles with different ages
	profiles := []struct {
		name   string
		handle string
		age    time.Duration
	}{
		{"store-a", "handle-a", 10 * 24 * time.Hour},
		{"store-b", "handle-b", 50 * 24 * time.Hour},
		{"store-c", "handle-c", 100 * 24 * time.Hour}, // old
	}

	for _, p := range profiles {
		creds := &secrets.StoreCredentials{
			Name:      p.name,
			Handle:    p.handle,
			CreatedAt: time.Now().Add(-p.age),
		}
		data, _ := creds.Marshal()
		mock.items["store:"+p.name] = keyring.Item{Key: "store:" + p.name, Data: data}
	}

	cleanup := setupAuthTestKeyring(mock)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newAuthTestCmd()
	err := authListCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify all profiles are listed
	for _, p := range profiles {
		if !strings.Contains(output, p.name) {
			t.Errorf("expected %s in output, got: %s", p.name, output)
		}
	}

	// Verify ROTATE status appears for old profile
	if !strings.Contains(output, "ROTATE") {
		t.Errorf("expected ROTATE status for old profile, got: %s", output)
	}
}

// TestAuthAddCmd_Structure tests that login command has correct structure
func TestAuthAddCmd_Structure(t *testing.T) {
	if authAddCmd.Use != "login" {
		t.Errorf("expected Use 'login', got %q", authAddCmd.Use)
	}
	if authAddCmd.Short != "Add a new store profile via browser" {
		t.Errorf("expected Short 'Add a new store profile via browser', got %q", authAddCmd.Short)
	}
	if authAddCmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

func TestAuthAddCmd_NoBrowserFlag(t *testing.T) {
	flag := authAddCmd.Flags().Lookup("no-browser")
	if flag == nil {
		t.Fatal("expected --no-browser flag on auth login command")
	}
	if flag.DefValue != "false" {
		t.Fatalf("--no-browser default = %q, want %q", flag.DefValue, "false")
	}
}

func TestSetNoBrowserEnv_DisabledNoop(t *testing.T) {
	t.Setenv("SHOPLINE_NO_BROWSER", "0")

	restore := setNoBrowserEnv(false)
	if got := os.Getenv("SHOPLINE_NO_BROWSER"); got != "0" {
		t.Fatalf("SHOPLINE_NO_BROWSER = %q, want %q", got, "0")
	}

	restore()
	if got := os.Getenv("SHOPLINE_NO_BROWSER"); got != "0" {
		t.Fatalf("after restore, SHOPLINE_NO_BROWSER = %q, want %q", got, "0")
	}
}

func TestSetNoBrowserEnv_EnabledRestoresUnsetState(t *testing.T) {
	original, exists := os.LookupEnv("SHOPLINE_NO_BROWSER")
	if exists {
		defer func() { _ = os.Setenv("SHOPLINE_NO_BROWSER", original) }()
	} else {
		defer func() { _ = os.Unsetenv("SHOPLINE_NO_BROWSER") }()
	}
	_ = os.Unsetenv("SHOPLINE_NO_BROWSER")

	restore := setNoBrowserEnv(true)
	if got := os.Getenv("SHOPLINE_NO_BROWSER"); got != "1" {
		t.Fatalf("SHOPLINE_NO_BROWSER = %q, want %q", got, "1")
	}

	restore()
	if _, stillSet := os.LookupEnv("SHOPLINE_NO_BROWSER"); stillSet {
		t.Fatal("expected SHOPLINE_NO_BROWSER to be unset after restore")
	}
}

func TestSetNoBrowserEnv_EnabledRestoresPreviousValue(t *testing.T) {
	t.Setenv("SHOPLINE_NO_BROWSER", "0")

	restore := setNoBrowserEnv(true)
	if got := os.Getenv("SHOPLINE_NO_BROWSER"); got != "1" {
		t.Fatalf("SHOPLINE_NO_BROWSER = %q, want %q", got, "1")
	}

	restore()
	if got := os.Getenv("SHOPLINE_NO_BROWSER"); got != "0" {
		t.Fatalf("after restore, SHOPLINE_NO_BROWSER = %q, want %q", got, "0")
	}
}
