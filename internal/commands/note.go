package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/spf13/cobra"
)

type NoteResult struct {
	Slug    string `json:"slug"`
	Note    string `json:"note"`
	AddedAt string `json:"added_at"`
}

func NewNoteCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "note <slug> <note>",
		Short: "Add a note to a project",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			noteText := args[1]

			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			err = c.TouchProject(ctx, slug, "note: "+noteText)
			if err != nil {
				return exit.New(exit.Network, "Failed to add note: "+err.Error())
			}

			result := NoteResult{
				Slug:    slug,
				Note:    noteText,
				AddedAt: time.Now().UTC().Format(time.RFC3339),
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				_ = enc.Encode(result)
				return nil
			}

			fmt.Printf("Note added to %s\n", slug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}