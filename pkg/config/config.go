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
	Server       ServerConfig       `mapstructure:"server" json:"server" yaml:"server"`
	Logging      LoggingConfig      `mapstructure:"logging" json:"logging" yaml:"logging"`
	Paths        PathsConfig        `mapstructure:"paths" json:"paths" yaml:"paths"`
	Engine       EngineConfig       `mapstructure:"engine" json:"engine" yaml:"engine"`
	LLM          LLMConfig          `mapstructure:"llm" json:"llm" yaml:"llm"`
	Telemetry    TelemetryConfig    `mapstructure:"telemetry" json:"telemetry" yaml:"telemetry"`
	Registry     RegistryConfig     `mapstructure:"registry" json:"registry" yaml:"registry"`
	Capabilities CapabilitiesConfig `mapstructure:"capabilities" json:"capabilities" yaml:"capabilities"`
}

// CapabilitiesConfig configures network egress and file system jail policies.
type CapabilitiesConfig struct {
	NetworkEgress []string `mapstructure:"network_egress" json:"network_egress" yaml:"network_egress"`
	FSJails       []string `mapstructure:"fs_jails" json:"fs_jails" yaml:"fs_jails"`
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

// DefaultSystemPrompt is the default system prompt used for code generation.
const DefaultSystemPrompt = `You are an expert Go programmer specializing in WebAssembly (WASM).
Your task is to write a valid, compilable Go program that compiles to WASM (GOOS=wasip1 GOARCH=wasm).
The program should be a standalone 'main' package.
Do not include any markdown formatting (e.g. triple backticks).
Output ONLY the raw Go source code.
Ensure the code imports necessary packages and handles errors gracefully.`

// LLMConfig configures the LLM provider.
type LLMConfig struct {
	Provider     string `mapstructure:"provider" json:"provider" yaml:"provider"`
	Model        string `mapstructure:"model" json:"model" yaml:"model"`
	APIKey       string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
	BaseURL      string `mapstructure:"base_url" json:"base_url" yaml:"base_url"`
	SystemPrompt string `mapstructure:"system_prompt" json:"system_prompt" yaml:"system_prompt"`
	CacheEnabled bool   `mapstructure:"cache_enabled" json:"cache_enabled" yaml:"cache_enabled"`
}

// TelemetryConfig configures OpenTelemetry.
type TelemetryConfig struct {
	Exporter string `mapstructure:"exporter" json:"exporter" yaml:"exporter"`
	Endpoint string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
}

// RegistryConfig configures the service registry.
type RegistryConfig struct {
	MaxActiveServices int `mapstructure:"max_active_services" json:"max_active_services" yaml:"max_active_services"`
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
	viper.SetDefault("llm.system_prompt", DefaultSystemPrompt)
	viper.SetDefault("llm.cache_enabled", false)
	viper.SetDefault("telemetry.exporter", "stdout")
	viper.SetDefault("registry.max_active_services", 10)

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
	c.LLM.BaseURL = strings.TrimSpace(c.LLM.BaseURL)
	c.LLM.SystemPrompt = strings.TrimSpace(c.LLM.SystemPrompt)
	c.Telemetry.Exporter = strings.ToLower(strings.TrimSpace(c.Telemetry.Exporter))
	c.Telemetry.Endpoint = strings.TrimSpace(c.Telemetry.Endpoint)
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
		validProviders := map[string]bool{"mock": true, "openai": true, "ollama": true}
		if !validProviders[strings.ToLower(c.LLM.Provider)] {
			return fmt.Errorf("llm.provider must be one of: mock, openai, ollama")
		}
		// Assuming we want to validate API Key presence only if not mock, but let's stick to simple logic for now
	}
	validExporters := map[string]bool{"stdout": true, "otlp-http": true, "otlp-grpc": true}
	if !validExporters[c.Telemetry.Exporter] {
		// Temporary debugging
		return fmt.Errorf("telemetry.exporter must be one of: stdout, otlp-http, otlp-grpc (got: '%s')", c.Telemetry.Exporter)
	}
	if c.Registry.MaxActiveServices <= 0 {
		return fmt.Errorf("registry.max_active_services must be greater than 0")
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
