package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func TestLoad_Defaults(t *testing.T) {
	viper.Reset()
	// Ensure no env vars interfere
	os.Unsetenv("GHOST_SERVER_PORT")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if config.Server.Port != 8080 {
		t.Errorf("expected server.port to be 8080, got %d", config.Server.Port)
	}
	if config.Logging.Level != "info" {
		t.Errorf("expected logging.level to be 'info', got %s", config.Logging.Level)
	}
	if config.Paths.Blueprints != "blueprints.json" {
		t.Errorf("expected paths.blueprints to be 'blueprints.json', got %s", config.Paths.Blueprints)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	viper.Reset()
	os.Setenv("GHOST_SERVER_PORT", "9090")
	defer os.Unsetenv("GHOST_SERVER_PORT")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if config.Server.Port != 9090 {
		t.Errorf("expected server.port to be 9090, got %d", config.Server.Port)
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

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if config.Server.Port != 7070 {
		t.Errorf("expected server.port to be 7070, got %d", config.Server.Port)
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

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Expect Env (9090) to win
	if config.Server.Port != 9090 {
		t.Errorf("expected server.port to be 9090 (Env), got %d", config.Server.Port)
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
	_, err = Load()
	if err == nil {
		t.Error("expected Load() to fail with malformed config, got nil")
	}
}

func TestConfig_Serialization(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Logging: LoggingConfig{Level: "debug", Format: "json"},
	}

	// Test JSON
	jsonBytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, `"server"`) || !strings.Contains(jsonStr, `"port"`) {
		t.Errorf("JSON output missing keys: %s", jsonStr)
	}

	// Test YAML
	yamlBytes, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal YAML: %v", err)
	}
	yamlStr := string(yamlBytes)
	// yaml.v3 output formatting might vary, but basic checks
	if !strings.Contains(yamlStr, "server:") || !strings.Contains(yamlStr, "port: 8080") {
		t.Errorf("YAML output missing keys: %s", yamlStr)
	}
}
