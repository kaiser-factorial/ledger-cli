package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/spf13/cobra"
)

func emitResult(jsonOutput bool, payload map[string]any, human string) error {
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	}
	fmt.Println(human)
	return nil
}

// NewArchiveCmd archives a project (soft-delete; hidden from status, restorable).
func NewArchiveCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a project (hidden from status; restorable)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			if err := c.ArchiveProject(ctx, id); err != nil {
				if client.IsNotFound(err) {
					return exit.New(exit.NotFound, "Project not found: "+id)
				}
				return exit.New(exit.Network, "Failed to archive: "+err.Error())
			}
			return emitResult(jsonOutput, map[string]any{"id": id, "archived": true}, "Archived "+id)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

// NewRestoreCmd restores an archived project back to the active list.
func NewRestoreCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "restore <id>",
		Short: "Restore an archived project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			if err := c.RestoreProject(ctx, id); err != nil {
				if client.IsNotFound(err) {
					return exit.New(exit.NotFound, "Project not found: "+id)
				}
				return exit.New(exit.Network, "Failed to restore: "+err.Error())
			}
			return emitResult(jsonOutput, map[string]any{"id": id, "archived": false}, "Restored "+id)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

// NewDeleteCmd permanently deletes a project. Requires --force to guard against
// accidental destructive use (especially from agents/scripts).
func NewDeleteCmd() *cobra.Command {
	var jsonOutput bool
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Permanently delete a project (requires --force)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !force {
				return exit.New(exit.Validation, "Refusing to delete without --force (this is permanent). Re-run: ledger delete "+id+" --force")
			}

			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			if err := c.DeleteProject(ctx, id); err != nil {
				if client.IsNotFound(err) {
					return exit.New(exit.NotFound, "Project not found: "+id)
				}
				return exit.New(exit.Network, "Failed to delete: "+err.Error())
			}
			return emitResult(jsonOutput, map[string]any{"id": id, "deleted": true}, "Deleted "+id)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm permanent deletion")
	return cmd
}
