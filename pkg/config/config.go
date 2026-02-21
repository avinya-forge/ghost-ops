package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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
	Port int `mapstructure:"port" json:"port" yaml:"port"`
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

	return &config, nil
}
