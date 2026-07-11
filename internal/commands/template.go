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

// NewTemplateCmd is the bulwork workflow-template store (Epic D): a thin, schema-agnostic JSON
// mirror of bulwork's TemplateStore, one Firestore doc per template (id = the template's own id).
func NewTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Read/write bulwork workflow templates (Epic D store)",
	}
	cmd.AddCommand(newTemplateListCmd())
	cmd.AddCommand(newTemplateShowCmd())
	cmd.AddCommand(newTemplateSetCmd())
	cmd.AddCommand(newTemplateDeleteCmd())
	return cmd
}

func newTemplateListCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved workflow templates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			list, err := c.ListTemplates(ctx)
			if err != nil {
				return exit.New(exit.Network, "Failed to list templates: "+err.Error())
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(list)
			}
			if len(list) == 0 {
				fmt.Println("(no templates)")
				return nil
			}
			for _, t := range list {
				fmt.Printf("  %v  %v\n", t["id"], t["name"])
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newTemplateShowCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show one workflow template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			data, ok, err := c.GetTemplate(ctx, id)
			if err != nil {
				return exit.New(exit.Network, "Failed to read template: "+err.Error())
			}
			// Unlike `plan show`, an explicit id was supplied — a miss is a real NotFound.
			if !ok {
				return exit.New(exit.NotFound, "Template not found: "+id)
			}
			return emitResult(jsonOutput, data, fmt.Sprintf("Template: %v", data["name"]))
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newTemplateSetCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "set <id>",
		Short: "Upsert a workflow template (full JSON body on stdin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			raw, err := io.ReadAll(os.Stdin)
			if err != nil {
				return exit.New(exit.Other, "Failed to read stdin: "+err.Error())
			}
			var data map[string]interface{}
			if err := json.Unmarshal(raw, &data); err != nil {
				return exit.New(exit.Validation, "Invalid JSON on stdin: "+err.Error())
			}
			// Review fix: see plan.go's newPlanSetCmd — a literal `null` body unmarshals to a nil
			// map with no error, which otherwise reaches the Firestore SDK as a leaky low-level
			// error instead of a clean validation message here.
			if data == nil {
				return exit.New(exit.Validation, "stdin must be a JSON object, not null")
			}

			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			if err := c.SetTemplate(ctx, id, data); err != nil {
				return exit.New(exit.Network, "Failed to save template: "+err.Error())
			}
			return emitResult(jsonOutput, map[string]any{"id": id, "saved": true}, "Template saved: "+id)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newTemplateDeleteCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a workflow template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ctx := context.Background()
			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			existed, err := c.DeleteTemplate(ctx, id)
			if err != nil {
				return exit.New(exit.Network, "Failed to delete template: "+err.Error())
			}
			if !existed {
				return exit.New(exit.NotFound, "Template not found: "+id)
			}
			return emitResult(jsonOutput, map[string]any{"id": id, "deleted": true}, "Deleted "+id)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
