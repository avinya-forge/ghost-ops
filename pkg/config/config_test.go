package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad_Defaults(t *testing.T) {
	viper.Reset()
	// Ensure no env vars interfere
	os.Unsetenv("GHOST_SERVER_PORT")

	err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if val := viper.GetInt("server.port"); val != 8080 {
		t.Errorf("expected server.port to be 8080, got %d", val)
	}
	if val := viper.GetString("logging.level"); val != "info" {
		t.Errorf("expected logging.level to be 'info', got %s", val)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	viper.Reset()
	os.Setenv("GHOST_SERVER_PORT", "9090")
	defer os.Unsetenv("GHOST_SERVER_PORT")

	err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if val := viper.GetInt("server.port"); val != 9090 {
		t.Errorf("expected server.port to be 9090, got %d", val)
	}
}

func TestLoad_FileOverride(t *testing.T) {
	// Create a temporary directory and change working directory to it
	// so Load() finds the config file in "."
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		// Restore original working directory
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Write a config file
	configContent := []byte("server:\n  port: 7070")
	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), configContent, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	viper.Reset()
	// Ensure no env vars interfere
	os.Unsetenv("GHOST_SERVER_PORT")

	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if val := viper.GetInt("server.port"); val != 7070 {
		t.Errorf("expected server.port to be 7070, got %d", val)
	}
}

func TestLoad_Precedence(t *testing.T) {
	// Env > File > Default

	// Setup File (Port 7070)
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	configContent := []byte("server:\n  port: 7070")
	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), configContent, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Setup Env (Port 9090)
	viper.Reset()
	os.Setenv("GHOST_SERVER_PORT", "9090")
	defer os.Unsetenv("GHOST_SERVER_PORT")

	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Expect Env (9090) to win
	if val := viper.GetInt("server.port"); val != 9090 {
		t.Errorf("expected server.port to be 9090 (Env), got %d", val)
	}
}

func TestLoad_MalformedConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Invalid YAML content
	configContent := []byte("server: port: 7070") // syntax error
	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), configContent, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	viper.Reset()
	err = Load()
	if err == nil {
		t.Error("expected Load() to fail with malformed config, got nil")
	}
}
