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

type DoctorResult struct {
	AuthStatus      string   `json:"auth_status"`
	FirestoreOK     bool     `json:"firestore_ok"`
	ConfigOK        bool     `json:"config_ok"`
	ProjectCount    int      `json:"project_count"`
	StaleCount      int      `json:"stale_count"`
	LastTouch       string   `json:"last_touch,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

func NewDoctorCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics on your Ledger setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			result := DoctorResult{
				AuthStatus:      "unknown",
				FirestoreOK:     false,
				ConfigOK:        true,
				Recommendations: []string{},
			}

			c, err := client.New(ctx)
			if err != nil {
				result.AuthStatus = "failed: " + err.Error()
				result.Recommendations = append(result.Recommendations, "Run 'ledger auth login' or set GOOGLE_APPLICATION_CREDENTIALS")
				return exit.New(exit.AuthError, result.AuthStatus)
			}
			defer c.Close()

			result.AuthStatus = "ok"
			result.FirestoreOK = true

			projects, err := c.GetAllProjects(ctx)
			if err != nil {
				result.FirestoreOK = false
				result.Recommendations = append(result.Recommendations, "Firestore query failed: "+err.Error())
				return exit.New(exit.Network, "Firestore connection failed")
			}
			result.ProjectCount = len(projects)

			stale, _ := c.GetStaleProjects(ctx)
			result.StaleCount = len(stale)

			if len(projects) > 0 {
				result.LastTouch = projects[0].LastTouched.Format(time.RFC3339)
			}

			if os.Getenv("LEDGER_EMAIL") == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
				result.Recommendations = append(result.Recommendations, "Consider setting LEDGER_EMAIL/LEDGER_PASSWORD or GOOGLE_APPLICATION_CREDENTIALS for headless use")
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				_ = enc.Encode(result)
				return nil
			}

			fmt.Println("Ledger Doctor Report")
			fmt.Println("====================")
			fmt.Printf("Auth Status:     %s\n", result.AuthStatus)
			fmt.Printf("Firestore:       %v\n", result.FirestoreOK)
			fmt.Printf("Projects:        %d\n", result.ProjectCount)
			fmt.Printf("Stale Projects:  %d\n", result.StaleCount)
			if result.LastTouch != "" {
				fmt.Printf("Last Touch:      %s\n", result.LastTouch)
			}
			if len(result.Recommendations) > 0 {
				fmt.Println("\nRecommendations:")
				for _, r := range result.Recommendations {
					fmt.Printf("  - %s\n", r)
				}
			} else {
				fmt.Println("\nEverything looks healthy.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}