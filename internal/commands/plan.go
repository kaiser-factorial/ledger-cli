package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/spf13/cobra"
)

// NewPlanCmd is the bulwork workload-plan store (Epic D): a thin, schema-agnostic JSON mirror of
// bulwork's PlanStore. bulwork owns the WorkloadPlan shape (src/types.ts); this command never
// validates it beyond "is it JSON" — that's deliberate (see ../../docs/CLI_COMMANDS.md).
func NewPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Read/write the bulwork workload plan (Epic D store)",
	}
	cmd.AddCommand(newPlanShowCmd())
	cmd.AddCommand(newPlanSetCmd())
	cmd.AddCommand(newPlanClearCmd())
	return cmd
}

func newPlanShowCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the current workload plan",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			data, ok, err := c.GetPlan(ctx)
			if err != nil {
				return exit.New(exit.Network, "Failed to read plan: "+err.Error())
			}
			// No plan is a valid, deliberate empty state — exit 0 either way (mirrors `ledger focus`).
			if !ok {
				if jsonOutput {
					fmt.Println("null")
					return nil
				}
				fmt.Println("No active plan")
				return nil
			}
			return emitResult(jsonOutput, data, fmt.Sprintf("Active plan: %v", data["id"]))
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newPlanSetCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Overwrite the current workload plan (full JSON body on stdin)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := io.ReadAll(os.Stdin)
			if err != nil {
				return exit.New(exit.Other, "Failed to read stdin: "+err.Error())
			}
			var data map[string]interface{}
			if err := json.Unmarshal(raw, &data); err != nil {
				// Syntactic validation only — the WorkloadPlan shape itself is bulwork's concern.
				return exit.New(exit.Validation, "Invalid JSON on stdin: "+err.Error())
			}
			// Review fix: `json.Unmarshal([]byte("null"), &data)` succeeds with data==nil (no
			// error) — without this check that reached the Firestore SDK, which rejects a nil
			// map with a leaky internal error string ("cannot convert value of type ... into a
			// map") instead of a clean validation message at the CLI's own input boundary.
			if data == nil {
				return exit.New(exit.Validation, "stdin must be a JSON object, not null")
			}

			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			if err := c.SetPlan(ctx, data); err != nil {
				return exit.New(exit.Network, "Failed to save plan: "+err.Error())
			}
			return emitResult(jsonOutput, map[string]any{"saved": true}, "Plan saved")
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newPlanClearCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete the current workload plan (idempotent)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			if err := c.ClearPlan(ctx); err != nil {
				return exit.New(exit.Network, "Failed to clear plan: "+err.Error())
			}
			return emitResult(jsonOutput, map[string]any{"cleared": true}, "Plan cleared")
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
