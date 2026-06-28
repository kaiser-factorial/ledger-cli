package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/kaiser/ledger-cli/internal/ui"
	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project statuses",
	RunE: func(cmd *cobra.Command, args []string) error {
		showStale, _ := cmd.Flags().GetBool("stale")
		showArchived, _ := cmd.Flags().GetBool("archived")
		showAll, _ := cmd.Flags().GetBool("all")
		useJSON, _ := cmd.Flags().GetBool("json")
		noColor, _ := cmd.Flags().GetBool("no-color")

		c, err := client.New(context.Background())
		if err != nil {
			return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
		}
		defer c.Close()

		var projects []client.Project
		var err2 error

		switch {
		case showStale:
			projects, err2 = c.GetStaleProjects(context.Background())
		case showArchived:
			projects, err2 = c.GetArchivedProjects(context.Background())
		case showAll:
			projects, err2 = c.GetAllProjectsIncludingArchived(context.Background())
		default:
			projects, err2 = c.GetAllProjects(context.Background())
		}
		if err2 != nil {
			return exit.New(exit.Network, "Failed to fetch projects: "+err2.Error())
		}

		if useJSON {
			data, _ := json.MarshalIndent(projects, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		ui.Header("PROJECT STATUS")

		fmt.Printf("%-3s %-20s %-15s %-12s %s\n", "", "NAME", "LAST TOUCHED", "REASON", "NEXT ACTION")
		fmt.Println("-------------------------------------------------------------------")

		for _, p := range projects {
			days := int(time.Since(p.LastTouched).Hours() / 24)
			reason := p.LastTouchReason
			if reason == "" {
				reason = "-"
			}
			next := p.NextAction
			if next == "" {
				next = "-"
			}

			var dot string
			if noColor {
				dot = ui.StatusDotPlain(p.LastTouched)
			} else {
				dot = ui.StatusDot(p.LastTouched)
			}

			fmt.Printf("%-3s %-20s %-15s %-12s %s\n", dot, p.Name, fmt.Sprintf("%dd ago", days), reason, next)
		}

		return nil
	},
}

func init() {
	StatusCmd.Flags().Bool("stale", false, "Only show stale projects (≥10 days)")
	StatusCmd.Flags().Bool("archived", false, "Only show archived projects")
	StatusCmd.Flags().Bool("all", false, "Show active and archived projects")
	StatusCmd.Flags().Bool("json", false, "Output JSON")
	StatusCmd.Flags().Bool("no-color", false, "Disable colored status dots")
}