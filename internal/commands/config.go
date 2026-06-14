package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kaiser/ledger-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View and manage Ledger CLI configuration",
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				fmt.Println("No config found. Using defaults.")
				cfg = &config.Config{}
				cfg.Thresholds.YellowDays = 5
				cfg.Thresholds.RedDays = 10
				cfg.PromptTime = "09:00"
			}

			data, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{}
				cfg.Thresholds.YellowDays = 5
				cfg.Thresholds.RedDays = 10
				cfg.PromptTime = "09:00"
			}

			switch key {
			case "yellowDays":
				fmt.Sscanf(value, "%d", &cfg.Thresholds.YellowDays)
			case "redDays":
				fmt.Sscanf(value, "%d", &cfg.Thresholds.RedDays)
			case "promptTime":
				cfg.PromptTime = value
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			// Save config
			path := config.GetConfigPath()
			os.MkdirAll("/Users/corinakaiser/.config/ledger", 0755)
			data, _ := json.MarshalIndent(cfg, "", "  ")
			os.WriteFile(path, data, 0644)

			fmt.Printf("Set %s = %s\n", key, value)
			return nil
		},
	}
}