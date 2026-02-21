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

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid Config (Default)",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "blueprints.json", Store: "store.json"},
				Engine:  EngineConfig{Type: "mock"},
				LLM:     LLMConfig{Provider: "mock"},
			},
			wantErr: false,
		},
		{
			name: "Invalid Port",
			config: Config{
				Server: ServerConfig{Port: 0},
			},
			wantErr: true,
		},
		{
			name: "Invalid Logging Level",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "verbose"},
			},
			wantErr: true,
		},
		{
			name: "Empty Blueprints Path",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "", Store: "store.json"},
			},
			wantErr: true,
		},
		{
			name: "Empty Store Path",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "blueprints.json", Store: ""},
			},
			wantErr: true,
		},
		{
			name: "Invalid Engine Type",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "blueprints.json", Store: "store.json"},
				Engine:  EngineConfig{Type: "unknown"},
			},
			wantErr: true,
		},
		{
			name: "Valid AI Engine",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "blueprints.json", Store: "store.json"},
				Engine:  EngineConfig{Type: "ai"},
				LLM:     LLMConfig{Provider: "openai"},
			},
			wantErr: false,
		},
		{
			name: "Invalid LLM Provider for AI Engine",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "blueprints.json", Store: "store.json"},
				Engine:  EngineConfig{Type: "ai"},
				LLM:     LLMConfig{Provider: "anthropic"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Sanitize(t *testing.T) {
	cfg := Config{
		Logging: LoggingConfig{Level: " DEBUG "},
		Engine:  EngineConfig{Type: " AI "},
		LLM:     LLMConfig{Provider: " OpenAI "},
	}

	cfg.Sanitize()

	if cfg.Logging.Level != "debug" {
		t.Errorf("expected logging.level to be 'debug', got '%s'", cfg.Logging.Level)
	}
	if cfg.Engine.Type != "ai" {
		t.Errorf("expected engine.type to be 'ai', got '%s'", cfg.Engine.Type)
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("expected llm.provider to be 'openai', got '%s'", cfg.LLM.Provider)
	}
}
