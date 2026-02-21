package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"ghost-ops/pkg/config"
)

func handleConfigCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: ghost-ops config [show]")
		return
	}

	switch args[0] {
	case "show":
		showConfig()
	default:
		fmt.Println("Unknown config command:", args[0])
	}
}

func showConfig() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Mask secrets
	if cfg.LLM.APIKey != "" {
		cfg.LLM.APIKey = "***REDACTED***"
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(cfg)
}
