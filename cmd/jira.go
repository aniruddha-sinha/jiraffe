package cmd

import (
	"fmt"

	"github.com/aniruddha-sinha/jiraffe/internal"
	"github.com/spf13/cobra"
)

var (
	email string
	org   string
)

const (
	defaultJiraOrg = "asinha0493"
)

var newJiraCmd = &cobra.Command{
	Use:     "jira",
	Aliases: []string{"j"},
	Short:   "subcommand for JIRA",
}

var newAuthCmd = &cobra.Command{
	Use:     "auth",
	Aliases: []string{"a"},
	Short:   "subcommand for auth",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("auth called")

		profile := internal.UserProfile{
			Email: email,
			Org:   org,
		}

		isAuthenticated, err := profile.HandleAuthentication()
		if err != nil {
			return err
		}

		if isAuthenticated {
			fmt.Println("Authentication successful; Credentials saved in ~/.config/jiraffe/config.yaml")
		}

		return nil
	},
}

func init() {
	// wiring the commands
	rootCmd.AddCommand(newJiraCmd)
	newJiraCmd.AddCommand(newAuthCmd)

	newAuthCmd.Flags().StringVarP(&email, "email", "e", "", "the email ID with which JIRA has been registered")
	newAuthCmd.Flags().StringVarP(&org, "org", "o", defaultJiraOrg, "the JIRA org where you are trying to login")
}
