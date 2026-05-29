package cmd

import (
	"fmt"

	"github.com/aniruddha-sinha/jiraffe/internal/config"
	"github.com/aniruddha-sinha/jiraffe/internal/jira"
	"github.com/spf13/cobra"
)

var (
	JiraConfigEmailKey        = "auth.jira.email"
	JiraConfigOrgKey          = "auth.jira.org"
	JiraConfigEncodedTokenKey = "auth.jira.encoded_token" // nolint:gosec // this is a config key and not an actual token
)

var sharedJiraCreds *jira.JiraCreds

func newCmdJira() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jira",
		Aliases: []string{"j"},
		Short:   "subcommand to interact with Atlassian Jira",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			storedEmail, err := config.Cfg.Get(JiraConfigEmailKey)
			if err != nil {
				return err
			}

			storedOrg, err := config.Cfg.Get(JiraConfigOrgKey)
			if err != nil {
				return err
			}

			storedEncodedToken, err := config.Cfg.Get(JiraConfigEncodedTokenKey)
			if err != nil {
				return err
			}

			sharedJiraCreds = jira.NewJiraCreds(storedEmail, storedOrg, storedEncodedToken)

			if err := sharedJiraCreds.EnsureAuthentication(cmd.Context()); err != nil {
				return err
			}

			fmt.Printf("User %s authenticated\n", storedEmail)
			return nil
		},
	}

	cmd.AddCommand(newCmdIssues())

	return cmd
}

func newCmdIssues() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "issue",
		Aliases: []string{"i"},
		Short:   "subcommand to target Jira Issues",
	}

	cmd.AddCommand(newCmdIssueList())

	return cmd
}

func newCmdIssueList() *cobra.Command {
	var jiraProject string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "subcommand to get list of jira issues in a Jira project/Space",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Fetching issues for project %s...\n", jiraProject)

			issues, err := jira.NewIssueService(jira.NewClient(sharedJiraCreds)).List(cmd.Context(), jiraProject)
			if err != nil {
				return err
			}

			for _, issue := range issues {
				fmt.Printf("[%s] %s (ID: %s)\n", issue.Key, issue.Fields.Summary, issue.ID)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&jiraProject, "project", "p", "", "the Jira project/space under which issues need to be listed")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		return nil
	}

	return cmd
}
