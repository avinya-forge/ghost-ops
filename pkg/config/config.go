package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load initializes the configuration system.
// It sets defaults, configures environment variable reading, and attempts to read a config file.
func Load() error {
	// 1. Set Defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

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
			return fmt.Errorf("fatal error config file: %w", err)
		}
	}

	return nil
}
