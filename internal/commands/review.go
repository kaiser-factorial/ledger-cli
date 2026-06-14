package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/kaiser/ledger-cli/internal/ui"
	"github.com/spf13/cobra"
)

var ReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Show weekly review buckets",
	RunE: func(cmd *cobra.Command, args []string) error {
		useWizard, _ := cmd.Flags().GetBool("wizard")
		useJSON, _ := cmd.Flags().GetBool("json")
		noColor, _ := cmd.Flags().GetBool("no-color")

		if useWizard {
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("🪄 Review Wizard")
			fmt.Print("Would you like JSON output? (y/n): ")
			jsonInput, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(jsonInput)) == "y" {
				useJSON = true
			}
		}

		c, err := client.New(context.Background())
		if err != nil {
			return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
		}
		defer c.Close()

		projects, err := c.GetAllProjects(context.Background())
		if err != nil {
			return exit.New(exit.Network, "Failed to fetch projects: "+err.Error())
		}

		now := time.Now()

		var neglected, missingAction, recent []client.Project

		for _, p := range projects {
			days := int(now.Sub(p.LastTouched).Hours() / 24)

			if days >= 10 {
				neglected = append(neglected, p)
			}
			if p.NextAction == "" {
				missingAction = append(missingAction, p)
			}
			if days <= 3 {
				recent = append(recent, p)
			}
		}

		if useJSON {
			data := map[string]interface{}{
				"neglected":       len(neglected),
				"missingNext":     len(missingAction),
				"recentlyTouched": len(recent),
			}
			b, _ := json.MarshalIndent(data, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		ui.Header("WEEKLY REVIEW")
		ui.Divider()

		fmt.Printf("🔴 Neglected (≥10d):       %d\n", len(neglected))
		for _, p := range neglected {
			dot := ui.StatusDotPlain(p.LastTouched)
			if !noColor {
				dot = ui.StatusDot(p.LastTouched)
			}
			fmt.Printf("   %s %s\n", dot, p.Name)
		}

		fmt.Printf("\n⚠️  Missing Next Action:    %d\n", len(missingAction))
		for _, p := range missingAction {
			dot := ui.StatusDotPlain(p.LastTouched)
			if !noColor {
				dot = ui.StatusDot(p.LastTouched)
			}
			fmt.Printf("   %s %s\n", dot, p.Name)
		}

		fmt.Printf("\n🟢 Recently Touched (≤3d): %d\n", len(recent))
		for _, p := range recent {
			dot := ui.StatusDotPlain(p.LastTouched)
			if !noColor {
				dot = ui.StatusDot(p.LastTouched)
			}
			fmt.Printf("   %s %s\n", dot, p.Name)
		}

		ui.Divider()
		return nil
	},
}

func init() {
	ReviewCmd.Flags().Bool("wizard", false, "Run interactive wizard for review")
	ReviewCmd.Flags().Bool("json", false, "Output JSON")
	ReviewCmd.Flags().Bool("no-color", false, "Disable colored status dots")
}