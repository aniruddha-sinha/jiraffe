package cmd

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/aniruddha-sinha/jiraffe/internal/config"
	"github.com/aniruddha-sinha/jiraffe/internal/jira"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newJiraCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jira",
		Aliases: []string{"j"},
		Short:   "subcommand for calling toplevel atlassian command line",
	}

	cmd.AddCommand(
		newAuthCmd(),
	)
	return cmd
}

func newAuthCmd() *cobra.Command {
	var (
		email string
		org   string
	)

	const (
		defaultJiraOrg = "asinha0493"
	)

	cmd := &cobra.Command{
		Use:     "auth",
		Aliases: []string{"a"},
		Short:   "Authenticate and save your Jira credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Ensure the required email flag was provided
			if email == "" {
				return errors.New("email is required. Use --email or -e")
			}

			// 2. Securely prompt for the Atlassian API Token
			fmt.Print("Enter your Atlassian API Token: ")
			byteToken, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println() // Print a newline after the hidden input
			if err != nil {
				return fmt.Errorf("failed to read token: %w", err)
			}
			authToken := string(byteToken)

			if authToken == "" {
				return errors.New("API token cannot be empty")
			}

			// 3. Validate the inputs by creating the Profile
			profile, err := jira.NewProfile(email, org)
			if err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			// 4. Initialize the Client and test the authentication
			client := jira.NewClient(nil, profile, authToken)

			fmt.Println("Verifying credentials with Atlassian...")
			if err := client.HandleAuthentication(); err != nil {
				if errors.Is(err, jira.ErrUnauthorized) {
					return errors.New("authentication failed: please check your API token and email")
				}
				return fmt.Errorf("authentication check failed: %w", err)
			}

			// 5. If authentication passes, persist the credentials to ~/.config/jiraffe/credentials.json
			fmt.Println("successfully authenticated! Saving configuration...")

			if err := config.Cfg.Upsert("auth.jira.email", email); err != nil {
				return fmt.Errorf("failed to save email: %w", err)
			}
			if err := config.Cfg.Upsert("auth.jira.org", org); err != nil {
				return fmt.Errorf("failed to save org: %w", err)
			}

			if err := config.Cfg.Upsert("auth.jira.encodedToken", authToken); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}

			fmt.Println("You are ready to use jiraffe!")
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "email registered with atlassian account")
	cmd.Flags().StringVarP(&org, "org", "o", defaultJiraOrg, "Atlassian organization name")

	return cmd
}
