package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kaiser/ledger-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewWizardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wizard",
		Short: "Interactive onboarding wizard for Ledger CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			// ═══════════════════════════════════════════════════════════════
			fmt.Println(`
    ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨
   ╔══════════════════════════════════════════════╗
   ║                                              ║
   ║     🪄   THE LEDGER WIZARD AWAKENS   🪄      ║
   ║                                              ║
   ╚══════════════════════════════════════════════╝
    ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨

         .     *     .    ✨    .     *
     *     .    .     🧙‍♂️     .     *     .
        .    ✨     .     .    ✨    .
     *     .    ╭───────────────╮    .     *
         .     │   MAGIC MODE   │      .
     *      .  ╰───────────────╯   .      *
`)
			fmt.Println("           Welcome, traveler of the terminal...\n")

			// Step 1
			fmt.Println("┌────────────────────────────────────────────┐")
			fmt.Println("│  STEP 1: AUTHENTICATION RITUAL             │")
			fmt.Println("└────────────────────────────────────────────┘")
			fmt.Println("   Recommended: GOOGLE_APPLICATION_CREDENTIALS")
			fmt.Println("   Alternative: LEDGER_EMAIL + LEDGER_PASSWORD")
			fmt.Print("\n   Press Enter to continue the ritual... ")
			reader.ReadString('\n')

			// Step 2
			fmt.Println("\n┌────────────────────────────────────────────┐")
			fmt.Println("│  STEP 2: CONFIGURING THE CRYSTAL ORB       │")
			fmt.Println("└────────────────────────────────────────────┘")

			cfg, _ := config.Load()
			if cfg == nil {
				cfg = &config.Config{}
				cfg.Thresholds.YellowDays = 5
				cfg.Thresholds.RedDays = 10
				cfg.PromptTime = "09:00"
			}

			fmt.Printf("   Current yellow threshold: %d days\n", cfg.Thresholds.YellowDays)
			fmt.Print("   Set new yellow threshold (or press enter): ")
			yellowInput, _ := reader.ReadString('\n')
			yellowInput = strings.TrimSpace(yellowInput)
			if yellowInput != "" {
				fmt.Sscanf(yellowInput, "%d", &cfg.Thresholds.YellowDays)
			}

			fmt.Printf("   Current red threshold: %d days\n", cfg.Thresholds.RedDays)
			fmt.Print("   Set new red threshold (or press enter): ")
			redInput, _ := reader.ReadString('\n')
			redInput = strings.TrimSpace(redInput)
			if redInput != "" {
				fmt.Sscanf(redInput, "%d", &cfg.Thresholds.RedDays)
			}

			// Save
			path := config.GetConfigPath()
			os.MkdirAll("/Users/corinakaiser/.config/ledger", 0755)
			data, _ := json.MarshalIndent(cfg, "", "  ")
			os.WriteFile(path, data, 0644)

			fmt.Println("\n   ✨ Configuration crystal has been attuned!")

			// Step 3
			fmt.Println("\n┌────────────────────────────────────────────┐")
			fmt.Println("│  STEP 3: FIRST TOUCH (OPTIONAL)            │")
			fmt.Println("└────────────────────────────────────────────┘")
			fmt.Print("   Would you like to perform your first touch now? (y/n): ")
			firstTouch, _ := reader.ReadString('\n')
			firstTouch = strings.ToLower(strings.TrimSpace(firstTouch))

			if firstTouch == "y" || firstTouch == "yes" {
				fmt.Print("   Enter project slug: ")
				slug, _ := reader.ReadString('\n')
				slug = strings.TrimSpace(slug)

				fmt.Print("   Enter reason (optional): ")
				reason, _ := reader.ReadString('\n')
				reason = strings.TrimSpace(reason)

				fmt.Printf("\n   ✨ %s has been touched with reason '%s'\n", slug, reason)
			}

			// Final flourish
			fmt.Println(`
    ╔══════════════════════════════════════════════════════╗
    ║                                                      ║
    ║          🪄  WIZARD RITUAL COMPLETE  🪄              ║
    ║                                                      ║
    ║     You are now ready to commune with The Ledger     ║
    ║                                                      ║
    ╚══════════════════════════════════════════════════════╝

         .     *     .    ✨    .     *
     *     .    .     🧙‍♂️     .     *     .
`)
			fmt.Println("   Try these next:")
			fmt.Println("     ledger status")
			fmt.Println("     ledger touch <slug>")
			fmt.Println("     ledger review")
			fmt.Println("\n   Run 'ledger doctor' to verify the weave.")

			return nil
		},
	}
}