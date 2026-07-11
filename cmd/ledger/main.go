package main

import (
	"github.com/kaiser/ledger-cli/internal/commands"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/kaiser/ledger-cli/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ledger",
	Short: "The Ledger CLI - Project health tracking",
	Long:  `A powerful terminal interface for The Ledger project health system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default behavior: launch TUI when no subcommand is given
		m := tui.NewModel()
		return tui.Run(m)
	},
}

func init() {
	rootCmd.AddCommand(commands.TouchCmd)
	rootCmd.AddCommand(commands.StatusCmd)
	rootCmd.AddCommand(commands.ReviewCmd)
	rootCmd.AddCommand(commands.NewDoctorCmd())
	rootCmd.AddCommand(commands.NewAnalyzeCmd())
	rootCmd.AddCommand(commands.NewNoteCmd())
	rootCmd.AddCommand(commands.NewArchiveCmd())
	rootCmd.AddCommand(commands.NewRestoreCmd())
	rootCmd.AddCommand(commands.NewDeleteCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewInitCmd())
	rootCmd.AddCommand(commands.NewAuthCmd())
	rootCmd.AddCommand(commands.NewWizardCmd())
	rootCmd.AddCommand(commands.NewCompletionCmd())
	rootCmd.AddCommand(commands.NewExportCmd())
	rootCmd.AddCommand(commands.NewPlanCmd())
	rootCmd.AddCommand(commands.NewTemplateCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		exit.Exit(err)
	}
}