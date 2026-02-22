package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"ghost-ops/pkg/protocol"
	"github.com/spf13/viper"
)

func handleServiceCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: ghost-ops service [list|inspect <id>]")
		return
	}

	port := viper.GetInt("server.port")
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	switch args[0] {
	case "list":
		listServices(baseURL)
	case "inspect":
		if len(args) < 2 {
			fmt.Println("Usage: ghost-ops service inspect <id>")
			return
		}
		inspectService(baseURL, args[1])
	default:
		fmt.Println("Unknown service command:", args[0])
	}
}

func listServices(baseURL string) {
	resp, err := http.Get(baseURL + "/services")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned %s\n", resp.Status)
		os.Exit(1)
	}

	var services []protocol.ServiceRecord
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-20s | %-10s | %-10s\n", "SERVICE ID", "ACTIVE", "SHADOW")
	fmt.Println("----------------------------------------------------")
	for _, s := range services {
		active := fmt.Sprintf("v%d", s.ActiveVersion)
		shadow := "-"
		if s.ShadowVersion > 0 {
			shadow = fmt.Sprintf("v%d", s.ShadowVersion)
		}
		fmt.Printf("%-20s | %-10s | %-10s\n", s.ServiceID, active, shadow)
	}
}

func inspectService(baseURL string, id string) {
	resp, err := http.Get(baseURL + "/services")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned %s\n", resp.Status)
		os.Exit(1)
	}

	var services []protocol.ServiceRecord
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	var found *protocol.ServiceRecord
	for _, s := range services {
		if s.ServiceID == id {
			found = &s
			break
		}
	}

	if found == nil {
		fmt.Printf("Service '%s' not found\n", id)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(found)
}
