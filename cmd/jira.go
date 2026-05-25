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

func newCmdJira() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jira",
		Aliases: []string{"j"},
		Short:   "subcommand to interact with Atlassian Jira",
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
	var (
		jiraProject string
		jc          *jira.JiraCreds
	)

	cmd := &cobra.Command{
		Use:           "list",
		Aliases:       []string{"l"},
		Short:         "subcommand to get list of jira issues in a Jira project/Space",
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
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

			jc = jira.NewJiraCreds(storedEmail, storedOrg, storedEncodedToken)

			if err := jc.EnsureAuthentication(cmd.Context()); err != nil {
				return err
			}

			fmt.Printf("User %s authenticated ", storedEmail)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if jiraProject == "" {
				return fmt.Errorf("project key is required")
			}

			fmt.Printf("Fetching issues for project %s...\n", jiraProject)

			issues, err := jira.NewClient(jc).Issues.ListByProject(cmd.Context(), jiraProject)
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

	return cmd
}
