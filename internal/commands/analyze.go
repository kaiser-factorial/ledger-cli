package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/exit"
	"github.com/kaiser/ledger-cli/internal/ui"
	"github.com/spf13/cobra"
)

type ReasonBreakdown struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type AnalyzeResult struct {
	TotalTouches int               `json:"total_touches"`
	Reasons      []ReasonBreakdown `json:"reasons"`
	Sparkline    string            `json:"sparkline,omitempty"`
}

func NewAnalyzeCmd() *cobra.Command {
	var jsonOutput bool
	var showSpark bool

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze your touch history",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			c, err := client.New(ctx)
			if err != nil {
				return exit.New(exit.AuthError, "Authentication failed: "+err.Error())
			}
			defer c.Close()

			projects, err := c.GetAllProjects(ctx)
			if err != nil {
				return exit.New(exit.Network, "Failed to fetch projects: "+err.Error())
			}

			reasonCounts := make(map[string]int)
			total := 0

			for _, p := range projects {
				for _, h := range p.TouchHistory {
					reasonCounts[h.Reason]++
					total++
				}
			}

			var breakdowns []ReasonBreakdown
			for reason, count := range reasonCounts {
				breakdowns = append(breakdowns, ReasonBreakdown{Reason: reason, Count: count})
			}
			sort.Slice(breakdowns, func(i, j int) bool {
				return breakdowns[i].Count > breakdowns[j].Count
			})

			result := AnalyzeResult{
				TotalTouches: total,
				Reasons:      breakdowns,
			}

			// Simple sparkline from recent touch distribution (last 8 counts)
			if showSpark && len(breakdowns) > 0 {
				var sparkValues []int
				for i := 0; i < 8 && i < len(breakdowns); i++ {
					sparkValues = append(sparkValues, breakdowns[i].Count)
				}
				result.Sparkline = ui.Sparkline(sparkValues)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				_ = enc.Encode(result)
				return nil
			}

			ui.Header("TOUCH ANALYSIS")
			fmt.Printf("Total touches: %d\n\n", total)

			for _, r := range breakdowns {
				fmt.Printf("%-20s %d\n", r.Reason, r.Count)
			}

			if showSpark && result.Sparkline != "" {
				fmt.Printf("\nSparkline: %s\n", result.Sparkline)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&showSpark, "spark", false, "Show ASCII sparkline of touch distribution")
	return cmd
}