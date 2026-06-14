package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize a new Ledger-managed project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			// Create basic project structure
			if err := os.MkdirAll(filepath.Join(dir, "docs"), 0755); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Join(dir, "src"), 0755); err != nil {
				return err
			}

			// Create a basic README
			readme := `# New Ledger Project

This project is tracked with The Ledger.

## Next Action
- [ ] Define initial goals and milestones
`
			os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0644)

			// Create .ledger marker file
			os.WriteFile(filepath.Join(dir, ".ledger"), []byte("ledger-project: true\n"), 0644)

			abs, _ := filepath.Abs(dir)
			fmt.Printf("Initialized new Ledger project at %s\n", abs)
			fmt.Println("Run 'ledger touch <slug>' to start tracking it.")
			return nil
		},
	}
}