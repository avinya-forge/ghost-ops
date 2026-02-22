package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to reset viper
func resetViper() {
	viper.Reset()
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid Config",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "b.json", Store: "s.json"},
				Engine:  EngineConfig{Type: "mock"},
				LLM:     LLMConfig{Provider: "mock"},
			},
			wantErr: false,
		},
		{
			name: "Invalid Port - Low",
			config: Config{
				Server:  ServerConfig{Port: 0},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "b.json", Store: "s.json"},
				Engine:  EngineConfig{Type: "mock"},
			},
			wantErr: true,
		},
		{
			name: "Invalid Port - High",
			config: Config{
				Server:  ServerConfig{Port: 70000},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "b.json", Store: "s.json"},
				Engine:  EngineConfig{Type: "mock"},
			},
			wantErr: true,
		},
		{
			name: "Invalid Logging Level",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "trace"},
				Paths:   PathsConfig{Blueprints: "b.json", Store: "s.json"},
				Engine:  EngineConfig{Type: "mock"},
			},
			wantErr: true,
		},
		{
			name: "Empty Paths",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "", Store: "s.json"},
				Engine:  EngineConfig{Type: "mock"},
			},
			wantErr: true,
		},
		{
			name: "Invalid Engine",
			config: Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "info"},
				Paths:   PathsConfig{Blueprints: "b.json", Store: "s.json"},
				Engine:  EngineConfig{Type: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config
			cfg.Sanitize()
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad_Defaults(t *testing.T) {
	resetViper()
	defer resetViper()

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "blueprints.json", cfg.Paths.Blueprints)
	assert.Equal(t, "store.json", cfg.Paths.Store)
	assert.Equal(t, "mock", cfg.Engine.Type)
}

func TestLoad_FileOverride(t *testing.T) {
	resetViper()
	defer resetViper()

	// Create a temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	content := `{
		"server": {"port": 9000},
		"logging": {"level": "debug"}
	}`
	err := os.WriteFile(configFile, []byte(content), 0644)
	require.NoError(t, err)

	// Add temp dir to viper search path
	viper.AddConfigPath(tmpDir)

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestLoad_EnvOverride(t *testing.T) {
	resetViper()
	defer resetViper()

	os.Setenv("GHOST_SERVER_PORT", "9090")
	os.Setenv("GHOST_LOGGING_LEVEL", "debug")
	defer os.Unsetenv("GHOST_SERVER_PORT")
	defer os.Unsetenv("GHOST_LOGGING_LEVEL")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
}
