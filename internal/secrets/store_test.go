package secrets

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/99designs/keyring"
)

func TestCredentialsSerialization(t *testing.T) {
	creds := &StoreCredentials{
		Name:        "test-store",
		Handle:      "mystore",
		AccessToken: "secret-token",
		AppKey:      "app-key-123",
		AppSecret:   "app-secret-456",
		Region:      "default",
		CreatedAt:   time.Now(),
	}

	data, err := creds.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	restored := &StoreCredentials{}
	if err := restored.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if restored.Name != creds.Name {
		t.Errorf("Name mismatch: got %s, want %s", restored.Name, creds.Name)
	}
	if restored.Handle != creds.Handle {
		t.Errorf("Handle mismatch: got %s, want %s", restored.Handle, creds.Handle)
	}
	if restored.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken mismatch")
	}
}

func TestCredentialsAge(t *testing.T) {
	old := &StoreCredentials{
		CreatedAt: time.Now().Add(-100 * 24 * time.Hour),
	}
	if !old.IsOld() {
		t.Error("Expected credentials older than 90 days to be marked as old")
	}

	recent := &StoreCredentials{
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
	}
	if recent.IsOld() {
		t.Error("Expected recent credentials to not be marked as old")
	}
}

// newTestStore creates a Store with an in-memory keyring for testing.
func newTestStore() *Store {
	return &Store{ring: keyring.NewArrayKeyring(nil)}
}

// newTestStoreWithItems creates a Store pre-populated with items.
func newTestStoreWithItems(items []keyring.Item) *Store {
	return &Store{ring: keyring.NewArrayKeyring(items)}
}

func TestStore_SaveAndGet(t *testing.T) {
	store := newTestStore()

	creds := &StoreCredentials{
		Name:        "my-store",
		Handle:      "myhandle",
		AccessToken: "token123",
		AppKey:      "key123",
		AppSecret:   "secret123",
		Region:      "us",
		CreatedAt:   time.Now(),
	}

	// Save credentials
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Retrieve credentials
	retrieved, err := store.Get("my-store")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != creds.Name {
		t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, creds.Name)
	}
	if retrieved.Handle != creds.Handle {
		t.Errorf("Handle mismatch: got %s, want %s", retrieved.Handle, creds.Handle)
	}
	if retrieved.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken mismatch")
	}
	if retrieved.AppKey != creds.AppKey {
		t.Errorf("AppKey mismatch")
	}
	if retrieved.AppSecret != creds.AppSecret {
		t.Errorf("AppSecret mismatch")
	}
	if retrieved.Region != creds.Region {
		t.Errorf("Region mismatch: got %s, want %s", retrieved.Region, creds.Region)
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	store := newTestStore()

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}
}

func TestStore_Get_InvalidJSON(t *testing.T) {
	// Pre-populate with invalid JSON data
	items := []keyring.Item{
		{
			Key:  accountPrefix + "corrupt-store",
			Data: []byte("not valid json"),
		},
	}
	store := newTestStoreWithItems(items)

	_, err := store.Get("corrupt-store")
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

func TestStore_Delete(t *testing.T) {
	store := newTestStore()

	creds := &StoreCredentials{
		Name:        "delete-me",
		Handle:      "handle",
		AccessToken: "token",
		CreatedAt:   time.Now(),
	}

	// Save first
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify it exists
	if _, err := store.Get("delete-me"); err != nil {
		t.Fatalf("Get failed before delete: %v", err)
	}

	// Delete
	if err := store.Delete("delete-me"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err := store.Get("delete-me")
	if err == nil {
		t.Fatal("Expected error after delete")
	}
}

func TestStore_List(t *testing.T) {
	store := newTestStore()

	// Initially empty
	names, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("Expected empty list, got %v", names)
	}

	// Add some stores
	stores := []string{"store-a", "store-b", "store-c"}
	for _, name := range stores {
		creds := &StoreCredentials{
			Name:      name,
			CreatedAt: time.Now(),
		}
		if err := store.Save(creds); err != nil {
			t.Fatalf("Save failed for %s: %v", name, err)
		}
	}

	// List should return all
	names, err = store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(names) != len(stores) {
		t.Errorf("Expected %d names, got %d", len(stores), len(names))
	}

	// Check all names are present
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, expected := range stores {
		if !nameSet[expected] {
			t.Errorf("Expected %s in list", expected)
		}
	}
}

func TestStore_List_FiltersNonStoreKeys(t *testing.T) {
	// Pre-populate with mixed keys (some with store prefix, some without)
	items := []keyring.Item{
		{Key: accountPrefix + "valid-store", Data: []byte("{}")},
		{Key: "other-key", Data: []byte("data")},
		{Key: "store", Data: []byte("data")}, // just "store", not "store:"
		{Key: accountPrefix + "another-store", Data: []byte("{}")},
	}
	store := newTestStoreWithItems(items)

	names, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Should only contain the two valid store names
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d: %v", len(names), names)
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if !nameSet["valid-store"] {
		t.Error("Expected valid-store in list")
	}
	if !nameSet["another-store"] {
		t.Error("Expected another-store in list")
	}
}

func TestStore_Save_UpdateExisting(t *testing.T) {
	store := newTestStore()

	// Save initial credentials
	creds1 := &StoreCredentials{
		Name:        "update-test",
		Handle:      "handle1",
		AccessToken: "token1",
		CreatedAt:   time.Now(),
	}
	if err := store.Save(creds1); err != nil {
		t.Fatalf("Initial save failed: %v", err)
	}

	// Update with new credentials
	creds2 := &StoreCredentials{
		Name:        "update-test",
		Handle:      "handle2",
		AccessToken: "token2",
		CreatedAt:   time.Now(),
	}
	if err := store.Save(creds2); err != nil {
		t.Fatalf("Update save failed: %v", err)
	}

	// Retrieve and verify update
	retrieved, err := store.Get("update-test")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if retrieved.Handle != "handle2" {
		t.Errorf("Handle not updated: got %s, want handle2", retrieved.Handle)
	}
	if retrieved.AccessToken != "token2" {
		t.Errorf("AccessToken not updated")
	}
}

func TestCredentialsIsOld_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		age      time.Duration
		expected bool
	}{
		// The condition is `time.Since(c.CreatedAt) > maxAge` where maxAge = 90 days
		// time.Since returns the elapsed time, so we need age > 90 days to be old
		{"just under 90 days", 90*24*time.Hour - time.Hour, false},
		{"just over 90 days", 90*24*time.Hour + time.Hour, true},
		{"89 days", 89 * 24 * time.Hour, false},
		{"91 days", 91 * 24 * time.Hour, true},
		{"zero age", 0, false},
		{"very old", 365 * 24 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &StoreCredentials{
				CreatedAt: time.Now().Add(-tt.age),
			}
			if got := creds.IsOld(); got != tt.expected {
				t.Errorf("IsOld() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCredentials_MarshalUnmarshal_AllFields(t *testing.T) {
	original := &StoreCredentials{
		Name:        "full-test",
		Handle:      "fullhandle",
		AccessToken: "full-access-token-xyz",
		AppKey:      "full-app-key",
		AppSecret:   "full-app-secret",
		Region:      "eu",
		CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	restored := &StoreCredentials{}
	if err := restored.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("Name mismatch")
	}
	if restored.Handle != original.Handle {
		t.Errorf("Handle mismatch")
	}
	if restored.AccessToken != original.AccessToken {
		t.Errorf("AccessToken mismatch")
	}
	if restored.AppKey != original.AppKey {
		t.Errorf("AppKey mismatch")
	}
	if restored.AppSecret != original.AppSecret {
		t.Errorf("AppSecret mismatch")
	}
	if restored.Region != original.Region {
		t.Errorf("Region mismatch")
	}
	if !restored.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", restored.CreatedAt, original.CreatedAt)
	}
}

func TestCredentials_Unmarshal_InvalidJSON(t *testing.T) {
	creds := &StoreCredentials{}
	err := creds.Unmarshal([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestCredentials_Marshal_EmptyStruct(t *testing.T) {
	creds := &StoreCredentials{}
	data, err := creds.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty JSON output")
	}
}

// mockKeyring implements keyring.Keyring for testing error conditions
type mockKeyring struct {
	getErr    error
	setErr    error
	removeErr error
	keysErr   error
	items     map[string]keyring.Item
}

func newMockKeyring() *mockKeyring {
	return &mockKeyring{items: make(map[string]keyring.Item)}
}

func (m *mockKeyring) Get(key string) (keyring.Item, error) {
	if m.getErr != nil {
		return keyring.Item{}, m.getErr
	}
	item, ok := m.items[key]
	if !ok {
		return keyring.Item{}, keyring.ErrKeyNotFound
	}
	return item, nil
}

func (m *mockKeyring) GetMetadata(_ string) (keyring.Metadata, error) {
	return keyring.Metadata{}, nil
}

func (m *mockKeyring) Set(item keyring.Item) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.items[item.Key] = item
	return nil
}

func (m *mockKeyring) Remove(key string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	delete(m.items, key)
	return nil
}

func (m *mockKeyring) Keys() ([]string, error) {
	if m.keysErr != nil {
		return nil, m.keysErr
	}
	keys := make([]string, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestStore_List_KeyringError(t *testing.T) {
	mock := newMockKeyring()
	mock.keysErr = errors.New("keyring error")
	store := &Store{ring: mock}

	_, err := store.List()
	if err == nil {
		t.Fatal("Expected error when keyring fails")
	}
}

func TestStore_Delete_Error(t *testing.T) {
	mock := newMockKeyring()
	mock.removeErr = errors.New("remove failed")
	store := &Store{ring: mock}

	err := store.Delete("any-key")
	if err == nil {
		t.Fatal("Expected error when Remove fails")
	}
}

func TestStore_List_EmptyPrefix(t *testing.T) {
	// Test with a key that is exactly the prefix (edge case)
	items := []keyring.Item{
		{Key: accountPrefix, Data: []byte("{}")}, // exactly "store:" with nothing after
	}
	store := newTestStoreWithItems(items)

	names, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// The key "store:" has length equal to len(accountPrefix), so it should be filtered out
	// because the condition is len(key) > len(accountPrefix)
	if len(names) != 0 {
		t.Errorf("Expected empty list for prefix-only key, got %v", names)
	}
}

func TestNewStore(t *testing.T) {
	// NewStore calls keyring.Open which may require system keyring access.
	// In CI environments or systems without a keyring, this may fail.
	// We test both success and potential failure scenarios.
	store, err := NewStore()
	if err != nil {
		// It's acceptable for NewStore to fail in environments without keyring
		t.Skipf("NewStore failed (may be expected in CI): %v", err)
	}
	if store == nil {
		t.Fatal("Expected non-nil store")
	}
	if store.ring == nil {
		t.Fatal("Expected non-nil keyring")
	}
}

func TestNewStore_Error(t *testing.T) {
	// Save original opener and restore after test
	originalOpener := keyringOpener
	defer func() { keyringOpener = originalOpener }()

	// Override with failing opener
	keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
		return nil, errors.New("keyring unavailable")
	}

	store, err := NewStore()
	if err == nil {
		t.Fatal("Expected error when keyring.Open fails")
	}
	if store != nil {
		t.Fatal("Expected nil store on error")
	}
}

func TestNewStoreWithKeyring(t *testing.T) {
	mock := newMockKeyring()
	store := NewStoreWithKeyring(mock)

	if store == nil {
		t.Fatal("Expected non-nil store")
	}
	if store.ring != mock {
		t.Fatal("Expected store to use provided keyring")
	}
}

func TestNewStore_ConfigIncludesFilePasswordFunc(t *testing.T) {
	originalOpener := keyringOpener
	originalFactory := filePasswordFuncFactory
	defer func() {
		keyringOpener = originalOpener
		filePasswordFuncFactory = originalFactory
	}()

	sentinelPrompt := keyring.FixedStringPrompt("test-passphrase")
	factoryCalled := false
	filePasswordFuncFactory = func() keyring.PromptFunc {
		factoryCalled = true
		return sentinelPrompt
	}

	var captured keyring.Config
	keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
		captured = cfg
		return keyring.NewArrayKeyring(nil), nil
	}

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("Expected non-nil store")
	}
	if !factoryCalled {
		t.Fatal("Expected file password factory to be called")
	}
	if captured.FilePasswordFunc == nil {
		t.Fatal("Expected FilePasswordFunc to be configured")
	}
	got, err := captured.FilePasswordFunc("prompt")
	if err != nil {
		t.Fatalf("FilePasswordFunc returned error: %v", err)
	}
	if got != "test-passphrase" {
		t.Fatalf("Expected FilePasswordFunc to use factory prompt value, got %q", got)
	}
}

func TestNewStore_ConfigUsesCredentialsDirOverride(t *testing.T) {
	originalOpener := keyringOpener
	originalGoos := goos
	originalGetenv := osGetenv
	defer func() {
		keyringOpener = originalOpener
		goos = originalGoos
		osGetenv = originalGetenv
	}()

	customDir := t.TempDir()
	t.Setenv("SHOPLINE_CREDENTIALS_DIR", customDir)

	goos = "darwin"
	osGetenv = os.Getenv

	var captured keyring.Config
	keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
		captured = cfg
		return keyring.NewArrayKeyring(nil), nil
	}

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("Expected non-nil store")
	}

	want := filepath.Join(customDir, "keyring")
	if captured.FileDir != want {
		t.Fatalf("Expected FileDir %q, got %q", want, captured.FileDir)
	}
}

func TestShouldForceFileBackend(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		dbusAddr string
		want     bool
	}{
		{name: "linux headless", goos: "linux", dbusAddr: "", want: true},
		{name: "linux headless whitespace", goos: "linux", dbusAddr: "   ", want: true},
		{name: "linux with dbus", goos: "linux", dbusAddr: "unix:path=/tmp/dbus", want: false},
		{name: "darwin no dbus", goos: "darwin", dbusAddr: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldForceFileBackend(tt.goos, tt.dbusAddr)
			if got != tt.want {
				t.Fatalf("shouldForceFileBackend(%q, %q) = %v, want %v", tt.goos, tt.dbusAddr, got, tt.want)
			}
		})
	}
}

func TestNewStore_HeadlessLinuxForcesFileBackend(t *testing.T) {
	originalOpener := keyringOpener
	originalGoos := goos
	originalGetenv := osGetenv
	originalTimeoutForce := fileBackendForcedAfterTimeout.Load()
	defer func() {
		keyringOpener = originalOpener
		goos = originalGoos
		osGetenv = originalGetenv
		fileBackendForcedAfterTimeout.Store(originalTimeoutForce)
	}()

	fileBackendForcedAfterTimeout.Store(false)
	goos = "linux"
	osGetenv = func(name string) string {
		if name == dbusSessionEnvName {
			return ""
		}
		return ""
	}

	var captured keyring.Config
	keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
		captured = cfg
		return keyring.NewArrayKeyring(nil), nil
	}

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("Expected non-nil store")
	}
	if len(captured.AllowedBackends) != 1 || captured.AllowedBackends[0] != keyring.FileBackend {
		t.Fatalf("Expected file backend to be forced on headless Linux, got %+v", captured.AllowedBackends)
	}
}

func TestNewStore_LinuxTimeoutFallsBackToFileBackend(t *testing.T) {
	originalOpener := keyringOpener
	originalGoos := goos
	originalGetenv := osGetenv
	originalTimeout := keyringOpenTimeout
	originalTimeoutForce := fileBackendForcedAfterTimeout.Load()
	defer func() {
		keyringOpener = originalOpener
		goos = originalGoos
		osGetenv = originalGetenv
		keyringOpenTimeout = originalTimeout
		fileBackendForcedAfterTimeout.Store(originalTimeoutForce)
	}()

	fileBackendForcedAfterTimeout.Store(false)
	goos = "linux"
	keyringOpenTimeout = 10 * time.Millisecond
	osGetenv = func(name string) string {
		if name == dbusSessionEnvName {
			return "unix:path=/tmp/dbus"
		}
		if name == keyringPassphraseEnvName {
			return "test-passphrase"
		}
		return ""
	}

	var (
		mu       sync.Mutex
		captured []keyring.Config
		slowDone = make(chan struct{}, 1)
	)
	keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
		mu.Lock()
		captured = append(captured, cfg)
		mu.Unlock()

		if len(cfg.AllowedBackends) == 1 && cfg.AllowedBackends[0] == keyring.FileBackend {
			return keyring.NewArrayKeyring(nil), nil
		}

		time.Sleep(50 * time.Millisecond)
		slowDone <- struct{}{}
		return nil, errors.New("simulated keyring hang")
	}

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("Expected non-nil store")
	}

	mu.Lock()
	if len(captured) < 2 {
		mu.Unlock()
		t.Fatalf("Expected at least two keyring open attempts, got %d", len(captured))
	}

	hasFileFallback := false
	for _, cfg := range captured {
		if len(cfg.AllowedBackends) == 1 && cfg.AllowedBackends[0] == keyring.FileBackend {
			hasFileFallback = true
		}
	}
	if !hasFileFallback {
		mu.Unlock()
		t.Fatal("Expected fallback attempt using file backend")
	}
	mu.Unlock()
	select {
	case <-slowDone:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for slow keyring attempt to finish")
	}
}

func TestNewStore_LinuxTimeoutWithoutPassphraseReturnsError(t *testing.T) {
	originalOpener := keyringOpener
	originalGoos := goos
	originalGetenv := osGetenv
	originalTimeout := keyringOpenTimeout
	originalTimeoutForce := fileBackendForcedAfterTimeout.Load()
	defer func() {
		keyringOpener = originalOpener
		goos = originalGoos
		osGetenv = originalGetenv
		keyringOpenTimeout = originalTimeout
		fileBackendForcedAfterTimeout.Store(originalTimeoutForce)
	}()

	fileBackendForcedAfterTimeout.Store(false)
	goos = "linux"
	keyringOpenTimeout = 10 * time.Millisecond
	osGetenv = func(name string) string {
		if name == dbusSessionEnvName {
			return "unix:path=/tmp/dbus"
		}
		if name == keyringPassphraseEnvName {
			return ""
		}
		return ""
	}

	slowDone := make(chan struct{}, 1)
	keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
		if len(cfg.AllowedBackends) == 1 && cfg.AllowedBackends[0] == keyring.FileBackend {
			t.Fatal("did not expect file backend open attempt without passphrase")
		}
		time.Sleep(50 * time.Millisecond)
		slowDone <- struct{}{}
		return nil, errors.New("simulated keyring hang")
	}

	store, err := NewStore()
	if err == nil {
		t.Fatal("Expected error")
	}
	if store != nil {
		t.Fatal("Expected nil store on error")
	}
	if !strings.Contains(err.Error(), keyringPassphraseEnvName) {
		t.Fatalf("Expected error to mention %s, got %q", keyringPassphraseEnvName, err.Error())
	}
	if fileBackendForcedAfterTimeout.Load() {
		t.Fatal("Expected fallback mode to remain disabled when timeout fallback did not succeed")
	}
	select {
	case <-slowDone:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for slow keyring attempt to finish")
	}
}

func TestNewStore_LinuxTimeoutFileFallbackFailure(t *testing.T) {
	originalOpener := keyringOpener
	originalGoos := goos
	originalGetenv := osGetenv
	originalTimeout := keyringOpenTimeout
	originalTimeoutForce := fileBackendForcedAfterTimeout.Load()
	defer func() {
		keyringOpener = originalOpener
		goos = originalGoos
		osGetenv = originalGetenv
		keyringOpenTimeout = originalTimeout
		fileBackendForcedAfterTimeout.Store(originalTimeoutForce)
	}()

	fileBackendForcedAfterTimeout.Store(false)
	goos = "linux"
	keyringOpenTimeout = 10 * time.Millisecond
	osGetenv = func(name string) string {
		if name == dbusSessionEnvName {
			return "unix:path=/tmp/dbus"
		}
		if name == keyringPassphraseEnvName {
			return "test-passphrase"
		}
		return ""
	}

	slowDone := make(chan struct{}, 1)
	keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
		if len(cfg.AllowedBackends) == 1 && cfg.AllowedBackends[0] == keyring.FileBackend {
			return nil, errors.New("file backend failed")
		}
		time.Sleep(50 * time.Millisecond)
		slowDone <- struct{}{}
		return nil, errors.New("simulated keyring hang")
	}

	store, err := NewStore()
	if err == nil {
		t.Fatal("Expected error")
	}
	if store != nil {
		t.Fatal("Expected nil store on error")
	}
	if !strings.Contains(err.Error(), "file fallback failed") {
		t.Fatalf("Expected fallback failure in error message, got %q", err.Error())
	}
	select {
	case <-slowDone:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for slow keyring attempt to finish")
	}
}

func TestDefaultFilePasswordFuncFactory_UsesEnvOverride(t *testing.T) {
	original := os.Getenv(keyringPassphraseEnvName)
	defer func() { _ = os.Setenv(keyringPassphraseEnvName, original) }()
	_ = os.Setenv(keyringPassphraseEnvName, "env-passphrase")

	prompt := defaultFilePasswordFuncFactory()
	got, err := prompt("prompt")
	if err != nil {
		t.Fatalf("prompt returned error: %v", err)
	}
	if got != "env-passphrase" {
		t.Fatalf("Expected env passphrase, got %q", got)
	}
}

func TestDefaultFilePasswordFuncFactory_DefaultsToServiceName(t *testing.T) {
	original := os.Getenv(keyringPassphraseEnvName)
	defer func() { _ = os.Setenv(keyringPassphraseEnvName, original) }()
	_ = os.Unsetenv(keyringPassphraseEnvName)

	prompt := defaultFilePasswordFuncFactory()
	got, err := prompt("prompt")
	if err != nil {
		t.Fatalf("prompt returned error: %v", err)
	}
	if got != serviceName {
		t.Fatalf("Expected default passphrase %q, got %q", serviceName, got)
	}
}

func TestNewStore_FileBackend_SaveAndGet(t *testing.T) {
	originalOpener := keyringOpener
	originalFactory := filePasswordFuncFactory
	defer func() {
		keyringOpener = originalOpener
		filePasswordFuncFactory = originalFactory
	}()

	filePasswordFuncFactory = func() keyring.PromptFunc {
		return keyring.FixedStringPrompt("test-file-passphrase")
	}

	keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
		cfg.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
		cfg.FileDir = t.TempDir()
		return keyring.Open(cfg)
	}

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	creds := &StoreCredentials{
		Name:        "file-backend-store",
		Handle:      "file-backend",
		AccessToken: "token-123",
		CreatedAt:   time.Now(),
	}
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := store.Get("file-backend-store")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.AccessToken != "token-123" {
		t.Fatalf("unexpected access token: %q", got.AccessToken)
	}
}

func TestStore_Save_SetError(t *testing.T) {
	mock := newMockKeyring()
	mock.setErr = errors.New("set failed")
	store := &Store{ring: mock}

	creds := &StoreCredentials{
		Name:      "test-store",
		CreatedAt: time.Now(),
	}

	err := store.Save(creds)
	if err == nil {
		t.Fatal("Expected error when Set fails")
	}
}

func TestStore_Save_MarshalError(t *testing.T) {
	// Save original marshaler and restore after test
	originalMarshal := jsonMarshal
	defer func() { jsonMarshal = originalMarshal }()

	// Override with failing marshaler
	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}

	store := newTestStore()
	creds := &StoreCredentials{
		Name:      "test-store",
		CreatedAt: time.Now(),
	}

	err := store.Save(creds)
	if err == nil {
		t.Fatal("Expected error when Marshal fails")
	}
}

func TestDefaultFilePasswordFuncFactory_WarnsOnDefault(t *testing.T) {
	original := os.Getenv(keyringPassphraseEnvName)
	defer func() { _ = os.Setenv(keyringPassphraseEnvName, original) }()
	_ = os.Unsetenv(keyringPassphraseEnvName)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	_ = defaultFilePasswordFuncFactory()

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if !strings.Contains(buf.String(), "SHOPLINE_KEYRING_PASSPHRASE") {
		t.Fatalf("expected stderr warning about SHOPLINE_KEYRING_PASSPHRASE, got: %q", buf.String())
	}
}
