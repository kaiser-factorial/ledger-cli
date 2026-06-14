package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/spf13/cobra"
)

func NewExportCmd() *cobra.Command {
	var format string
	var outputFile string
	var onlyStale bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export project data to JSON or CSV",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			var projects []client.Project
			var errFetch error

			if onlyStale {
				projects, errFetch = c.GetStaleProjects(ctx)
			} else {
				projects, errFetch = c.GetAllProjects(ctx)
			}
			if errFetch != nil {
				return exit.New(exit.Network, "Failed to fetch projects: "+errFetch.Error())
			}

			// Prepare output
			var out *os.File = os.Stdout
			if outputFile != "" {
				f, err := os.Create(outputFile)
				if err != nil {
					return exit.New(exit.Other, "Failed to create output file: "+err.Error())
				}
				defer f.Close()
				out = f
			}

			switch format {
			case "json":
				encoder := json.NewEncoder(out)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(projects); err != nil {
					return exit.New(exit.Other, "Failed to encode JSON: "+err.Error())
				}

			case "csv":
				writer := csv.NewWriter(out)
				defer writer.Flush()

				// Header
				header := []string{"id", "name", "lastTouched", "lastTouchReason", "nextAction", "statusNote"}
				if err := writer.Write(header); err != nil {
					return exit.New(exit.Other, "Failed to write CSV header")
				}

				for _, p := range projects {
					row := []string{
						p.ID,
						p.Name,
						p.LastTouched.Format(time.RFC3339),
						p.LastTouchReason,
						p.NextAction,
						p.StatusNote,
					}
					if err := writer.Write(row); err != nil {
						return exit.New(exit.Other, "Failed to write CSV row")
					}
				}

			default:
				return exit.New(exit.Validation, "Unsupported format: "+format+" (use json or csv)")
			}

			if outputFile != "" {
				fmt.Printf("Exported %d projects to %s\n", len(projects), outputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Output format: json or csv")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write to file instead of stdout")
	cmd.Flags().BoolVar(&onlyStale, "stale", false, "Only export stale projects (≥10 days)")
	return cmd
}