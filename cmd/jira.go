package cmd

import (
	"errors"
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

	cmd.AddCommand(
		newCmdIssueList(),
		newCmdIssuesGet(),
		newCmdIssuesCreate(),
	)

	return cmd
}

func newCmdIssueList() *cobra.Command {
	var (
		jiraProject string
		maxPages    int
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "subcommand to get list of jira issues in a Jira project/Space",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if maxPages < 1 {
				return errors.New("--pages must be >= 1")
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Fetching issues for project %s...\n", jiraProject)

			issues, err := jira.NewIssueService(jira.NewClient(sharedJiraCreds)).List(cmd.Context(), jiraProject, maxPages)
			if err != nil {
				return err
			}

			for _, issue := range issues {
				fmt.Printf("[%s] | Status: %s | Priority : %s | Assignee: %s \n",
					issue.Key, issue.Summary(), issue.Priority(), issue.Assignee())
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&jiraProject, "project", "p", "", "the Jira project/space under which issues need to be listed")
	cmd.Flags().IntVar(&maxPages, "pages", 0, "number of pages to fetch (must be >= 1)")
	for _, flag := range []string{"project", "pages"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}

	return cmd
}

func newCmdIssuesGet() *cobra.Command {
	var (
		issueKey   string
		outputJson bool
	)

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "subcommand for getting jira issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("fetching issue", issueKey)
			issue, err := jira.NewIssueService(jira.NewClient(sharedJiraCreds)).Get(cmd.Context(), issueKey)
			if err != nil {
				return err
			}

			if outputJson {
				jsonOut, err := issue.Json()
				if err != nil {
					return err
				}

				fmt.Println(jsonOut)
			} else {
				prettyPrintIssue, err := issue.String()
				if err != nil {
					return err
				}

				fmt.Print(prettyPrintIssue)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&issueKey, "issue-key", "i", "", "issue key such as XCBDD-12345")
	cmd.Flags().BoolVarP(&outputJson, "json", "j", false, "if selected the output will be in json format")
	if err := cmd.MarkFlagRequired("issue-key"); err != nil {
		return nil
	}

	return cmd
}

func newCmdIssuesCreate() *cobra.Command {
	var (
		projectKey    string
		summary       string
		description   string
		issueType     string
		assignee      string
		reporter      string
		labels        []string
		sprintID      int
		teamID        string
		sprintFieldID string // Dynamic key for Sprint
		teamFieldID   string // Dynamic key for Team
		payload       *jira.CreateIssueRequest
	)

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "create a new JIRA issue",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			projectRef := jira.NewProjectRef(projectKey)
			issueTypeRef := jira.NewIssueTypeRef(issueType)
			desc := jira.BuildADFDescription(description)
			fields := jira.NewCreateIssueFields(*projectRef, summary, desc, *issueTypeRef, labels, make(map[string]any))
			payload = jira.NewCreateIssueRequest(fields)

			// Populate dynamic fields if provided
			if assignee != "" {
				fields.Assignee = &jira.UserRef{ID: assignee}
			}

			if reporter != "" {
				fields.Reporter = &jira.UserRef{ID: reporter}
			}

			if sprintID != 0 && sprintFieldID != "" {
				payload.Fields.CustomFields[sprintFieldID] = sprintID
			} else if sprintID != 0 && sprintFieldID == "" {
				return fmt.Errorf("you provided a sprint ID but no --sprint-field-id")
			}

			if teamID != "" && teamFieldID != "" {
				payload.Fields.CustomFields[teamFieldID] = teamID
			} else if teamID != "" && teamFieldID == "" {
				return fmt.Errorf("you provided a team ID but no --team-field-id")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("create issue in %s\n", projectKey)
			client := jira.NewIssueService(jira.NewClient(sharedJiraCreds))
			res, err := client.Create(cmd.Context(), *payload)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully created issue: %s (%s)\n", res.Key, res.Self)
			return nil
		},
	}

	cmd.Flags().StringVarP(&projectKey, "project", "p", "", "Project key (e.g., PROJ)")
	cmd.Flags().StringVarP(&summary, "summary", "s", "", "Issue summary/title")
	cmd.Flags().StringVarP(&issueType, "type", "t", "Task", "Issue type (e.g., Bug, Task, Story)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "description")
	cmd.Flags().StringSliceVarP(&labels, "labels", "l", []string{}, "Comma-separated labels")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "the user, the ticket in question has been assigned to")
	cmd.Flags().StringVarP(&reporter, "reporter", "r", "", "the user who raises this ticket")

	cmd.Flags().IntVar(&sprintID, "sprint", 0, "Sprint ID (numeric)")
	cmd.Flags().StringVar(&teamID, "team", "", "Team ID (string)")

	cmd.Flags().StringVar(&sprintFieldID, "sprint-field-id", "", "The custom field key for Sprints in your Jira instance")
	cmd.Flags().StringVar(&teamFieldID, "team-field-id", "", "The custom field key for Teams in your Jira instance")

	for _, x := range []string{"project", "summary", "reporter"} {
		if err := cmd.MarkFlagRequired(x); err != nil {
			panic(err)
		}
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
