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

	cmd.AddCommand(
		newCmdIssues(),
		newCmdProjects(),
	)

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

func newCmdProjects() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"p"},
		Short:   "subcommand to target Jira Projects",
	}

	cmd.AddCommand(
		newCmdProjectsList(),
		newCmdProjectsGet(),
	)

	return cmd
}

func newCmdProjectsList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "subcommand to get list of projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("fetching list of projets")
			projects, err := jira.NewProjectService(jira.NewClient(sharedJiraCreds)).List(cmd.Context())
			if err != nil {
				return err
			}

			for _, project := range projects {
				fmt.Printf("%s \t %s \t %s \t %s \t %s\n", project.ProjectID(), project.ProjectKey(), project.ProjectName(), project.JiraProjectTypeKey(), project.ProjectStyle())
			}

			return nil
		},
	}

	return cmd
}

func newCmdProjectsGet() *cobra.Command {
	var projectKey string

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "subcommand to get project by projectKey",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("get project details by project key")
			project, err := jira.NewProjectService(jira.NewClient(sharedJiraCreds)).Get(cmd.Context(), projectKey)
			if err != nil {
				return err
			}

			fmt.Println("Project Details")
			fmt.Printf("Project Key = %s\n", project.ProjectKey())
			fmt.Printf("Project ID = %s\n", project.ProjectID())
			fmt.Printf("Project Name = %s\n", project.ProjectName())
			fmt.Printf("Project Style = %s\n", project.ProjectStyle())
			fmt.Printf("Project TypeKey = %s\n", project.JiraProjectTypeKey())
			return nil
		},
	}

	cmd.Flags().StringVarP(&projectKey, "project-key", "p", "", "project key for the project like XMN-2345")
	if err := cmd.MarkFlagRequired("project-key"); err != nil {
		return nil
	}

	return cmd
}
