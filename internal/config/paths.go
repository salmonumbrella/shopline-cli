// Package config provides OS-appropriate paths for configuration and data storage.
//
// Note: This package is currently unused because credentials are stored in
// the system keyring (via github.com/99designs/keyring) rather than in config files.
// It exists as infrastructure for future features that may need file-based storage,
// such as caching, local configuration overrides, or log persistence.
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const appName = "shopline-cli"

const (
	credentialsDirEnvName         = "SHOPLINE_CREDENTIALS_DIR"
	openClawCredentialsDirEnvName = "CW_CREDENTIALS_DIR"
)

// osType and homeDirFunc are package-level variables to enable testing
// of OS-specific code paths across all platforms.
var (
	osType      = runtime.GOOS
	homeDirFunc = os.UserHomeDir
)

// ConfigDir returns the configuration directory path.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}

	home, err := homeDirFunc()
	if err != nil {
		return filepath.Join(".", appName)
	}

	switch osType {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", appName)
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, appName)
		}
		return filepath.Join(home, appName)
	default:
		return filepath.Join(home, ".config", appName)
	}
}

// DataDir returns the data directory path.
func DataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}

	home, err := homeDirFunc()
	if err != nil {
		return filepath.Join(".", appName)
	}

	switch osType {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", appName)
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, appName)
		}
		return filepath.Join(home, appName)
	default:
		return filepath.Join(home, ".local", "share", appName)
	}
}

// CredentialsDir returns the directory used for credential storage.
//
// Precedence:
//  1. SHOPLINE_CREDENTIALS_DIR (explicit tool override)
//  2. CW_CREDENTIALS_DIR + "/shopline-cli" (shared OpenClaw root)
//  3. DataDir()
func CredentialsDir() string {
	if dir := strings.TrimSpace(os.Getenv(credentialsDirEnvName)); dir != "" {
		return dir
	}
	if dir := strings.TrimSpace(os.Getenv(openClawCredentialsDirEnvName)); dir != "" {
		return filepath.Join(dir, appName)
	}
	return DataDir()
}
