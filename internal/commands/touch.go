package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/kaiser/ledger-cli/internal/reason"
	"github.com/kaiser/ledger-cli/internal/ui"
	"github.com/spf13/cobra"
)

var TouchCmd = &cobra.Command{
	Use:   "touch <slug>",
	Short: "Mark a project as touched",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		useWizard, _ := cmd.Flags().GetBool("wizard")
		reasonInput, _ := cmd.Flags().GetString("reason")
		nextAction, _ := cmd.Flags().GetString("next")
		useJSON, _ := cmd.Flags().GetBool("json")
		noColor, _ := cmd.Flags().GetBool("no-color")

		var slug string

		if useWizard {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Project slug: ")
			slugInput, _ := reader.ReadString('\n')
			slug = strings.TrimSpace(slugInput)

			fmt.Print("Reason (or preset): ")
			reasonInput, _ = reader.ReadString('\n')
			reasonInput = strings.TrimSpace(reasonInput)
		} else {
			if len(args) == 0 {
				return exit.New(exit.Validation, "slug is required (or use --wizard)")
			}
			slug = args[0]
		}

		// Resolve reason preset
		finalReason := resolveReason(reasonInput)

		c, err := client.New(context.Background())
		if err != nil {
			return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
		}
		defer c.Close()

		if err := c.TouchProject(context.Background(), slug, finalReason); err != nil {
			return exit.New(exit.Network, "Failed to touch project: "+err.Error())
		}

		if useJSON {
			fmt.Printf(`{"id":"%s","reason":"%s"}`+"\n", slug, finalReason)
			return nil
		}

		ui.Header("TOUCH RECORDED")
		dot := ui.StatusDotPlain(time.Now())
		if !noColor {
			dot = ui.StatusDot(time.Now())
		}
		if finalReason != "" && finalReason != "none" {
			fmt.Printf("%s %s touched (%s)\n", dot, slug, finalReason)
		} else {
			fmt.Printf("%s %s touched\n", dot, slug)
		}
		_ = nextAction
		return nil
	},
}

func resolveReason(input string) string {
	if input == "" || input == "none" {
		return "none"
	}
	if preset, ok := reason.Presets[input]; ok {
		return preset
	}
	return input
}

func init() {
	TouchCmd.Flags().Bool("wizard", false, "Run interactive wizard for this touch")
	TouchCmd.Flags().String("reason", "", "Touch reason (or preset: blocked, research, fix, feat, review, cleanup)")
	TouchCmd.Flags().String("next", "", "Update next action")
	TouchCmd.Flags().Bool("json", false, "Output JSON")
	TouchCmd.Flags().Bool("no-color", false, "Disable colored status dots")
}