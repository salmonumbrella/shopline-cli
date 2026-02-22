package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// saveEnv saves an environment variable and returns a restore function.
func saveEnv(t *testing.T, key string) func() {
	t.Helper()
	old, exists := os.LookupEnv(key)
	return func() {
		if exists {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}

// withOS temporarily sets the osType variable for testing.
func withOS(t *testing.T, goos string) func() {
	t.Helper()
	old := osType
	osType = goos
	return func() { osType = old }
}

// withHomeDir temporarily sets the homeDirFunc for testing.
func withHomeDir(t *testing.T, fn func() (string, error)) func() {
	t.Helper()
	old := homeDirFunc
	homeDirFunc = fn
	return func() { homeDirFunc = old }
}

func TestConfigDirXDG(t *testing.T) {
	tmpDir := t.TempDir()
	restore := saveEnv(t, "XDG_CONFIG_HOME")
	defer restore()
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)

	dir := ConfigDir()
	expected := filepath.Join(tmpDir, appName)
	if dir != expected {
		t.Errorf("ConfigDir with XDG_CONFIG_HOME: expected %q, got %q", expected, dir)
	}
}

func TestConfigDirDarwin(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_CONFIG_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_CONFIG_HOME")

	restoreOS := withOS(t, "darwin")
	defer restoreOS()

	home := "/Users/testuser"
	restoreHome := withHomeDir(t, func() (string, error) { return home, nil })
	defer restoreHome()

	dir := ConfigDir()
	expected := filepath.Join(home, "Library", "Application Support", appName)
	if dir != expected {
		t.Errorf("ConfigDir on darwin: expected %q, got %q", expected, dir)
	}
}

func TestConfigDirWindowsWithAPPDATA(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_CONFIG_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_CONFIG_HOME")

	restoreOS := withOS(t, "windows")
	defer restoreOS()

	home := `C:\Users\testuser`
	restoreHome := withHomeDir(t, func() (string, error) { return home, nil })
	defer restoreHome()

	appData := `C:\Users\testuser\AppData\Roaming`
	restoreAppData := saveEnv(t, "APPDATA")
	defer restoreAppData()
	_ = os.Setenv("APPDATA", appData)

	dir := ConfigDir()
	expected := filepath.Join(appData, appName)
	if dir != expected {
		t.Errorf("ConfigDir on windows with APPDATA: expected %q, got %q", expected, dir)
	}
}

func TestConfigDirWindowsWithoutAPPDATA(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_CONFIG_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_CONFIG_HOME")

	restoreOS := withOS(t, "windows")
	defer restoreOS()

	home := `C:\Users\testuser`
	restoreHome := withHomeDir(t, func() (string, error) { return home, nil })
	defer restoreHome()

	restoreAppData := saveEnv(t, "APPDATA")
	defer restoreAppData()
	_ = os.Unsetenv("APPDATA")

	dir := ConfigDir()
	expected := filepath.Join(home, appName)
	if dir != expected {
		t.Errorf("ConfigDir on windows without APPDATA: expected %q, got %q", expected, dir)
	}
}

func TestConfigDirLinux(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_CONFIG_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_CONFIG_HOME")

	restoreOS := withOS(t, "linux")
	defer restoreOS()

	home := "/home/testuser"
	restoreHome := withHomeDir(t, func() (string, error) { return home, nil })
	defer restoreHome()

	dir := ConfigDir()
	expected := filepath.Join(home, ".config", appName)
	if dir != expected {
		t.Errorf("ConfigDir on linux: expected %q, got %q", expected, dir)
	}
}

func TestConfigDirHomeDirError(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_CONFIG_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_CONFIG_HOME")

	restoreHome := withHomeDir(t, func() (string, error) {
		return "", errors.New("home dir not found")
	})
	defer restoreHome()

	dir := ConfigDir()
	expected := filepath.Join(".", appName)
	if dir != expected {
		t.Errorf("ConfigDir with home dir error: expected %q, got %q", expected, dir)
	}
}

func TestDataDirXDG(t *testing.T) {
	tmpDir := t.TempDir()
	restore := saveEnv(t, "XDG_DATA_HOME")
	defer restore()
	_ = os.Setenv("XDG_DATA_HOME", tmpDir)

	dir := DataDir()
	expected := filepath.Join(tmpDir, appName)
	if dir != expected {
		t.Errorf("DataDir with XDG_DATA_HOME: expected %q, got %q", expected, dir)
	}
}

func TestDataDirDarwin(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_DATA_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_DATA_HOME")

	restoreOS := withOS(t, "darwin")
	defer restoreOS()

	home := "/Users/testuser"
	restoreHome := withHomeDir(t, func() (string, error) { return home, nil })
	defer restoreHome()

	dir := DataDir()
	expected := filepath.Join(home, "Library", "Application Support", appName)
	if dir != expected {
		t.Errorf("DataDir on darwin: expected %q, got %q", expected, dir)
	}
}

func TestDataDirWindowsWithLOCALAPPDATA(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_DATA_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_DATA_HOME")

	restoreOS := withOS(t, "windows")
	defer restoreOS()

	home := `C:\Users\testuser`
	restoreHome := withHomeDir(t, func() (string, error) { return home, nil })
	defer restoreHome()

	localAppData := `C:\Users\testuser\AppData\Local`
	restoreLocalAppData := saveEnv(t, "LOCALAPPDATA")
	defer restoreLocalAppData()
	_ = os.Setenv("LOCALAPPDATA", localAppData)

	dir := DataDir()
	expected := filepath.Join(localAppData, appName)
	if dir != expected {
		t.Errorf("DataDir on windows with LOCALAPPDATA: expected %q, got %q", expected, dir)
	}
}

func TestDataDirWindowsWithoutLOCALAPPDATA(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_DATA_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_DATA_HOME")

	restoreOS := withOS(t, "windows")
	defer restoreOS()

	home := `C:\Users\testuser`
	restoreHome := withHomeDir(t, func() (string, error) { return home, nil })
	defer restoreHome()

	restoreLocalAppData := saveEnv(t, "LOCALAPPDATA")
	defer restoreLocalAppData()
	_ = os.Unsetenv("LOCALAPPDATA")

	dir := DataDir()
	expected := filepath.Join(home, appName)
	if dir != expected {
		t.Errorf("DataDir on windows without LOCALAPPDATA: expected %q, got %q", expected, dir)
	}
}

func TestDataDirLinux(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_DATA_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_DATA_HOME")

	restoreOS := withOS(t, "linux")
	defer restoreOS()

	home := "/home/testuser"
	restoreHome := withHomeDir(t, func() (string, error) { return home, nil })
	defer restoreHome()

	dir := DataDir()
	expected := filepath.Join(home, ".local", "share", appName)
	if dir != expected {
		t.Errorf("DataDir on linux: expected %q, got %q", expected, dir)
	}
}

func TestDataDirHomeDirError(t *testing.T) {
	restoreXDG := saveEnv(t, "XDG_DATA_HOME")
	defer restoreXDG()
	_ = os.Unsetenv("XDG_DATA_HOME")

	restoreHome := withHomeDir(t, func() (string, error) {
		return "", errors.New("home dir not found")
	})
	defer restoreHome()

	dir := DataDir()
	expected := filepath.Join(".", appName)
	if dir != expected {
		t.Errorf("DataDir with home dir error: expected %q, got %q", expected, dir)
	}
}

func TestCredentialsDir_ShoplineOverride(t *testing.T) {
	restoreShopline := saveEnv(t, credentialsDirEnvName)
	defer restoreShopline()
	restoreCW := saveEnv(t, openClawCredentialsDirEnvName)
	defer restoreCW()

	_ = os.Setenv(credentialsDirEnvName, "/tmp/shopline-creds")
	_ = os.Setenv(openClawCredentialsDirEnvName, "/tmp/openclaw-creds")

	got := CredentialsDir()
	if got != "/tmp/shopline-creds" {
		t.Fatalf("CredentialsDir() = %q, want %q", got, "/tmp/shopline-creds")
	}
}

func TestCredentialsDir_OpenClawFallback(t *testing.T) {
	restoreShopline := saveEnv(t, credentialsDirEnvName)
	defer restoreShopline()
	restoreCW := saveEnv(t, openClawCredentialsDirEnvName)
	defer restoreCW()

	_ = os.Unsetenv(credentialsDirEnvName)
	_ = os.Setenv(openClawCredentialsDirEnvName, "/tmp/openclaw-creds")

	got := CredentialsDir()
	want := filepath.Join("/tmp/openclaw-creds", appName)
	if got != want {
		t.Fatalf("CredentialsDir() = %q, want %q", got, want)
	}
}

func TestCredentialsDir_DefaultsToDataDir(t *testing.T) {
	restoreShopline := saveEnv(t, credentialsDirEnvName)
	defer restoreShopline()
	restoreCW := saveEnv(t, openClawCredentialsDirEnvName)
	defer restoreCW()
	restoreXDG := saveEnv(t, "XDG_DATA_HOME")
	defer restoreXDG()

	_ = os.Unsetenv(credentialsDirEnvName)
	_ = os.Unsetenv(openClawCredentialsDirEnvName)

	restoreOS := withOS(t, "linux")
	defer restoreOS()
	restoreHome := withHomeDir(t, func() (string, error) { return "/home/testuser", nil })
	defer restoreHome()
	_ = os.Unsetenv("XDG_DATA_HOME")

	got := CredentialsDir()
	want := filepath.Join("/home/testuser", ".local", "share", appName)
	if got != want {
		t.Fatalf("CredentialsDir() = %q, want %q", got, want)
	}
}

// TestRealOSPaths tests the functions with the real OS settings
// to ensure they work correctly on the current platform.
func TestRealOSPaths(t *testing.T) {
	// Ensure we use the real OS settings
	originalOS := osType
	originalHomeDir := homeDirFunc
	osType = runtime.GOOS
	homeDirFunc = os.UserHomeDir
	defer func() {
		osType = originalOS
		homeDirFunc = originalHomeDir
	}()

	// Clear XDG variables to test native OS paths
	restoreXDGConfig := saveEnv(t, "XDG_CONFIG_HOME")
	defer restoreXDGConfig()
	_ = os.Unsetenv("XDG_CONFIG_HOME")

	restoreXDGData := saveEnv(t, "XDG_DATA_HOME")
	defer restoreXDGData()
	_ = os.Unsetenv("XDG_DATA_HOME")

	configDir := ConfigDir()
	dataDir := DataDir()

	if configDir == "" {
		t.Error("ConfigDir returned empty string")
	}
	if dataDir == "" {
		t.Error("DataDir returned empty string")
	}

	// Verify paths contain appName
	if !filepath.IsAbs(configDir) && configDir != filepath.Join(".", appName) {
		t.Errorf("ConfigDir should be absolute or fallback path, got: %s", configDir)
	}
	if !filepath.IsAbs(dataDir) && dataDir != filepath.Join(".", appName) {
		t.Errorf("DataDir should be absolute or fallback path, got: %s", dataDir)
	}
}
