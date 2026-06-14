package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication for Ledger CLI",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthWhoamiCmd())
	cmd.AddCommand(newAuthServiceAccountCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login with email/password or service account",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Login flow not yet implemented.")
			fmt.Println("For now, set GOOGLE_APPLICATION_CREDENTIALS or LEDGER_EMAIL/LEDGER_PASSWORD.")
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Logged out (placeholder).")
			return nil
		},
	}
}

func newAuthWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current authenticated identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Current identity: (not implemented)")
			return nil
		},
	}
}

func newAuthServiceAccountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "service-account",
		Short: "Configure service account authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("To use a service account:")
			fmt.Println("  export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json")
			return nil
		},
	}
}