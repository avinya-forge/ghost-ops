package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"ghost-ops/pkg/logging"
)

// Config represents the application configuration.
type Config struct {
	Server  ServerConfig  `mapstructure:"server" json:"server" yaml:"server"`
	Logging LoggingConfig `mapstructure:"logging" json:"logging" yaml:"logging"`
	Paths   PathsConfig   `mapstructure:"paths" json:"paths" yaml:"paths"`
	Engine  EngineConfig  `mapstructure:"engine" json:"engine" yaml:"engine"`
	LLM     LLMConfig     `mapstructure:"llm" json:"llm" yaml:"llm"`
}

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	Port        int   `mapstructure:"port" json:"port" yaml:"port"`
	MaxBodySize int64 `mapstructure:"max_body_size" json:"max_body_size" yaml:"max_body_size"`
}

// LoggingConfig configures the logger.
type LoggingConfig struct {
	Level  string `mapstructure:"level" json:"level" yaml:"level"`
	Format string `mapstructure:"format" json:"format" yaml:"format"`
}

// PathsConfig configures file paths.
type PathsConfig struct {
	Blueprints string `mapstructure:"blueprints" json:"blueprints" yaml:"blueprints"`
	WASM       string `mapstructure:"wasm" json:"wasm" yaml:"wasm"`
	Store      string `mapstructure:"store" json:"store" yaml:"store"`
}

// EngineConfig configures the evolution engine.
type EngineConfig struct {
	Type string `mapstructure:"type" json:"type" yaml:"type"`
}

// LLMConfig configures the LLM provider.
type LLMConfig struct {
	Provider string `mapstructure:"provider" json:"provider" yaml:"provider"`
	Model    string `mapstructure:"model" json:"model" yaml:"model"`
	APIKey   string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
}

// Load initializes the configuration system and returns the loaded Config.
// It sets defaults, configures environment variable reading, and attempts to read a config file.
func Load() (*Config, error) {
	// 1. Set Defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.max_body_size", 1048576) // 1MB
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("paths.blueprints", "blueprints.json")
	viper.SetDefault("paths.store", "store.json")
	viper.SetDefault("engine.type", "mock")
	viper.SetDefault("llm.provider", "mock")

	// 2. Configure Environment Variables
	viper.SetEnvPrefix("GHOST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 3. Configure Config File
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory

	// Add user home directory config path
	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(home, ".ghost-ops"))
	}

	// 4. Read Config File
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// (we are okay with just defaults and env vars)
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("fatal error config file: %w", err)
		}
	}

	// 5. Unmarshal into Struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	config.Sanitize()

	return &config, nil
}

// Sanitize cleans up the configuration (e.g. trimming whitespace, lowercasing).
func (c *Config) Sanitize() {
	c.Logging.Level = strings.ToLower(strings.TrimSpace(c.Logging.Level))
	c.Engine.Type = strings.ToLower(strings.TrimSpace(c.Engine.Type))
	c.LLM.Provider = strings.ToLower(strings.TrimSpace(c.LLM.Provider))
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[strings.ToLower(c.Logging.Level)] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}
	if strings.TrimSpace(c.Paths.Blueprints) == "" {
		return fmt.Errorf("paths.blueprints cannot be empty")
	}
	if strings.TrimSpace(c.Paths.Store) == "" {
		return fmt.Errorf("paths.store cannot be empty")
	}
	validEngines := map[string]bool{"mock": true, "compiler": true, "ai": true}
	if !validEngines[strings.ToLower(c.Engine.Type)] {
		return fmt.Errorf("engine.type must be one of: mock, compiler, ai")
	}
	if strings.ToLower(c.Engine.Type) == "ai" {
		validProviders := map[string]bool{"mock": true, "openai": true}
		if !validProviders[strings.ToLower(c.LLM.Provider)] {
			return fmt.Errorf("llm.provider must be one of: mock, openai")
		}
		// Assuming we want to validate API Key presence only if not mock, but let's stick to simple logic for now
	}
	return nil
}

// Watch starts watching for configuration changes.
func Watch(cfg *Config) {
	viper.OnConfigChange(func(e fsnotify.Event) {
		slog.Info("Config file changed", "name", e.Name)
		if err := viper.Unmarshal(cfg); err != nil {
			slog.Error("Failed to reload config", "error", err)
			return
		}
		cfg.Sanitize()
		if err := cfg.Validate(); err != nil {
			slog.Error("Invalid config after reload", "error", err)
			return
		}

		// Update logging level if changed
		level := slog.LevelInfo
		switch cfg.Logging.Level {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
		logging.SetLevel(level)
		slog.Info("Configuration reloaded", "logging.level", cfg.Logging.Level)
	})
	viper.WatchConfig()
}
