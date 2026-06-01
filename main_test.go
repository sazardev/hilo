package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain redirects the config directory to a throwaway temp folder so the
// suite never reads or writes the developer's real ~/.config/hilo, and so
// tests stay deterministic regardless of prior runs.
func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "hilo-test-")
	if err != nil {
		panic(err)
	}
	cfg := filepath.Join(tmp, "config")
	// Cover every platform os.UserConfigDir consults.
	os.Setenv("AppData", cfg)             // Windows
	os.Setenv("XDG_CONFIG_HOME", cfg)     // Linux
	os.Setenv("HOME", filepath.Join(tmp)) // macOS fallback

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}
