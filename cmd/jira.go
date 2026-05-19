package cmd

import (
	"github.com/spf13/cobra"
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
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "email registered with atlassian account")
	cmd.Flags().StringVarP(&org, "org", "o", defaultJiraOrg, "Atlassian organization name")

	return cmd
}
