package secrets

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/shopline-cli/internal/config"
)

const (
	serviceName   = "shopline-cli"
	accountPrefix = "store:"
	maxAge        = 90 * 24 * time.Hour
)

// KeyringOpenerFunc is the function type for opening a keyring.
type KeyringOpenerFunc func(cfg keyring.Config) (keyring.Keyring, error)

// keyringOpener is a function type for opening a keyring (testable).
var keyringOpener KeyringOpenerFunc = func(cfg keyring.Config) (keyring.Keyring, error) {
	return keyring.Open(cfg)
}

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
	ring, err := keyringOpener(keyring.Config{
		ServiceName: serviceName,
		// FileDir is required for the file backend (used when CGO is disabled
		// and macOS keychain is unavailable). Store in a "keyring" subdirectory
		// to keep credential files organized separately from other app data.
		FileDir: filepath.Join(config.DataDir(), "keyring"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}
	return &Store{ring: ring}, nil
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
