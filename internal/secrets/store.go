package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/shopline-cli/internal/config"
)

const (
	serviceName              = "shopline-cli"
	accountPrefix            = "store:"
	maxAge                   = 90 * 24 * time.Hour
	keyringPassphraseEnvName = "SHOPLINE_KEYRING_PASSPHRASE"
	dbusSessionEnvName       = "DBUS_SESSION_BUS_ADDRESS"
)

// KeyringOpenerFunc is the function type for opening a keyring.
type KeyringOpenerFunc func(cfg keyring.Config) (keyring.Keyring, error)

// keyringOpener is a function type for opening a keyring (testable).
var keyringOpener KeyringOpenerFunc = func(cfg keyring.Config) (keyring.Keyring, error) {
	return keyring.Open(cfg)
}

// osGetenv and goos are package-level for deterministic testing.
var (
	osGetenv = os.Getenv
	goos     = runtime.GOOS
)

// keyringOpenTimeout controls how long keyring opening may block on Linux
// before we retry with the file backend to avoid hanging indefinitely.
var keyringOpenTimeout = 5 * time.Second

var errKeyringOpenTimeout = errors.New("timed out opening keyring")

var fileBackendForcedAfterTimeout atomic.Bool

// GetKeyringOpener returns the current keyring opener function.
func GetKeyringOpener() KeyringOpenerFunc {
	return keyringOpener
}

// SetKeyringOpener sets the keyring opener function (for testing).
func SetKeyringOpener(opener KeyringOpenerFunc) {
	keyringOpener = opener
}

// jsonMarshal is the JSON marshaling function (can be overridden for testing).
var jsonMarshal = json.Marshal

// filePasswordFuncFactory builds the file backend passphrase prompt function.
// It is package-level for testability.
var filePasswordFuncFactory = defaultFilePasswordFuncFactory

// StoreCredentials holds authentication data for a Shopline store.
type StoreCredentials struct {
	Name        string    `json:"name"`
	Handle      string    `json:"handle"`
	AccessToken string    `json:"access_token"`
	AppKey      string    `json:"app_key"`
	AppSecret   string    `json:"app_secret"`
	Region      string    `json:"region"`
	CreatedAt   time.Time `json:"created_at"`
}

// Marshal serializes credentials to JSON.
func (c *StoreCredentials) Marshal() ([]byte, error) {
	return jsonMarshal(c)
}

// Unmarshal deserializes credentials from JSON.
func (c *StoreCredentials) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}

// IsOld returns true if credentials are older than 90 days.
func (c *StoreCredentials) IsOld() bool {
	return time.Since(c.CreatedAt) > maxAge
}

// Store manages credential storage in the system keyring.
type Store struct {
	ring keyring.Keyring
}

// NewStore creates a new credential store.
func NewStore() (*Store, error) {
	cfg := keyring.Config{
		ServiceName: serviceName,
		// FileDir is required for the file backend (used when CGO is disabled
		// and macOS keychain is unavailable). Store in a "keyring" subdirectory
		// to keep credential files organized separately from other app data.
		FileDir: filepath.Join(config.CredentialsDir(), "keyring"),
		// 99designs/keyring requires FilePasswordFunc for the file backend.
		// Without this, file backend operations can nil-deref at runtime.
		FilePasswordFunc: filePasswordFuncFactory(),
	}

	headlessForceFile := shouldForceFileBackend(goos, osGetenv(dbusSessionEnvName))
	timeoutForceFile := goos == "linux" && fileBackendForcedAfterTimeout.Load()

	// If fallback mode was previously enabled but passphrase is now unavailable,
	// clear the fallback flag and retry system keyring backends.
	if timeoutForceFile && !headlessForceFile && osGetenv(keyringPassphraseEnvName) == "" {
		fileBackendForcedAfterTimeout.Store(false)
		timeoutForceFile = false
	}
	forceFile := headlessForceFile || timeoutForceFile

	if forceFile {
		cfg.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
	}

	ring, err := openKeyring(cfg, forceFile)
	if err == nil {
		return &Store{ring: ring}, nil
	}

	if !forceFile && goos == "linux" && errors.Is(err, errKeyringOpenTimeout) {
		if osGetenv(keyringPassphraseEnvName) == "" {
			return nil, fmt.Errorf("failed to open keyring: system keyring timed out; set %s to enable file backend fallback", keyringPassphraseEnvName)
		}

		fileCfg := cfg
		fileCfg.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
		ring, fileErr := keyringOpener(fileCfg)
		if fileErr == nil {
			fileBackendForcedAfterTimeout.Store(true)
			fmt.Fprintln(os.Stderr, "Warning: system keyring unavailable; using encrypted file keyring backend")
			return &Store{ring: ring}, nil
		}
		return nil, fmt.Errorf("failed to open keyring: system keyring timed out and file fallback failed: %w", fileErr)
	}

	return nil, fmt.Errorf("failed to open keyring: %w", err)
}

func shouldForceFileBackend(goosValue, dbusAddr string) bool {
	return goosValue == "linux" && strings.TrimSpace(dbusAddr) == ""
}

func openKeyring(cfg keyring.Config, forceFile bool) (keyring.Keyring, error) {
	if goos != "linux" || forceFile || keyringOpenTimeout <= 0 {
		return keyringOpener(cfg)
	}

	type result struct {
		ring keyring.Keyring
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		ring, err := keyringOpener(cfg)
		ch <- result{ring: ring, err: err}
	}()

	timer := time.NewTimer(keyringOpenTimeout)
	defer timer.Stop()

	select {
	case res := <-ch:
		return res.ring, res.err
	case <-timer.C:
		return nil, errKeyringOpenTimeout
	}
}

func defaultFilePasswordFuncFactory() keyring.PromptFunc {
	if passphrase := os.Getenv(keyringPassphraseEnvName); passphrase != "" {
		return keyring.FixedStringPrompt(passphrase)
	}

	fmt.Fprintf(os.Stderr, "Warning: using default keyring passphrase; set %s for encryption at rest\n", keyringPassphraseEnvName)
	return keyring.FixedStringPrompt(serviceName)
}

// NewStoreWithKeyring creates a store with a custom keyring (for testing).
func NewStoreWithKeyring(ring keyring.Keyring) *Store {
	return &Store{ring: ring}
}

// Save stores credentials in the keyring.
func (s *Store) Save(creds *StoreCredentials) error {
	data, err := creds.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	key := accountPrefix + creds.Name
	return s.ring.Set(keyring.Item{
		Key:  key,
		Data: data,
	})
}

// Get retrieves credentials from the keyring.
func (s *Store) Get(name string) (*StoreCredentials, error) {
	key := accountPrefix + name
	item, err := s.ring.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	creds := &StoreCredentials{}
	if err := creds.Unmarshal(item.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}
	return creds, nil
}

// Delete removes credentials from the keyring.
func (s *Store) Delete(name string) error {
	key := accountPrefix + name
	return s.ring.Remove(key)
}

// List returns all stored credential names.
func (s *Store) List() ([]string, error) {
	keys, err := s.ring.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	var names []string
	for _, key := range keys {
		if len(key) > len(accountPrefix) && key[:len(accountPrefix)] == accountPrefix {
			names = append(names, key[len(accountPrefix):])
		}
	}
	return names, nil
}
