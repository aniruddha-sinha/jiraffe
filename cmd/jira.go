package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aniruddha-sinha/jiraffe/internal/config"
	"github.com/aniruddha-sinha/jiraffe/internal/jira"
	"github.com/spf13/cobra"
)

var (
	JiraConfigEmailKey             = "auth.jira.email"
	JiraConfigOrgKey               = "auth.jira.org"
	JiraConfigEncodedTokenKey      = "auth.jira.encoded_token" // nolint:gosec // this is a config key and not an actual token
	JiraConfigSprintCustomFieldKey = "jira.customfields.sprint"
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
		projectKey, summary, description, issueType, assignee, reporter, sprintName, parent string
		labels                                                                              []string
		payload                                                                             *jira.CreateIssueRequest
		dryRun                                                                              bool
	)

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "create a new JIRA issue",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := jira.NewClient(sharedJiraCreds)

			// 1. Build Base Payload
			projectRef := jira.NewProjectRef(projectKey)
			issueTypeRef := jira.NewIssueTypeRef(issueType)
			desc := jira.BuildADFDescription(description)
			fields := jira.NewCreateIssueFields(*projectRef, summary, desc, *issueTypeRef, labels, make(map[string]any))
			payload = jira.NewCreateIssueRequest(fields)

			// 2. Resolve Users
			if assignee != "" {
				targetID, err := resolveAccountID(ctx, client, assignee)
				if err != nil {
					return fmt.Errorf("failed to resolve assignee: %w", err)
				}
				payload.Fields.Assignee = jira.NewUserRef(targetID)
			}

			if reporter != "" {
				targetID, err := resolveAccountID(ctx, client, reporter)
				if err != nil {
					return fmt.Errorf("failed to resolve reporter: %w", err)
				}
				payload.Fields.Reporter = jira.NewUserRef(targetID)
			}

			// 3. Resolve Parent
			if parent != "" {
				payload.Fields.Parent = jira.NewParentRef(parent)
			}

			// 4. Resolve Sprint (with JIT loading)
			if sprintName != "" {
				sprintFieldID, err := ensureSprintFieldID(ctx, client)
				if err != nil {
					return err
				}

				var finalSprintID int
				if id, err := strconv.Atoi(sprintName); err == nil {
					finalSprintID = id
				} else {
					fmt.Printf("Resolving sprint name '%s' via Jira API...\n", sprintName)
					finalSprintID, err = client.ResolveSprintNameToID(ctx, projectKey, sprintName)
					if err != nil {
						return err
					}
				}
				payload.Fields.CustomFields[sprintFieldID] = finalSprintID
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("create issue in %s\n", projectKey)
			if dryRun {
				fmt.Println("--- DRY-RUN PAYLOAD OUTPUT---")
				fmt.Println(payload.PrintJSON())
				fmt.Println("---------------------")
				return nil
			}

			client := jira.NewIssueService(jira.NewClient(sharedJiraCreds))
			res, err := client.Create(cmd.Context(), *payload)
			if err != nil {
				return err
			}

			jsonFormatted, err := res.PrintJSON()
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", jsonFormatted)
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
	cmd.Flags().StringVar(&parent, "parent", "", "the parent ticket if subticket depends on it")
	cmd.Flags().StringVar(&sprintName, "sprint", "", "Sprint Name")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "choose dry run if you want to see the payload before creating a ticket to not create a lot of tickets")

	for _, x := range []string{"project", "summary", "reporter", "sprint"} {
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

// resolveAccountID checks if the input is an email and fetches the Atlassian Account ID.
// If it is already an ID (no '@'), it returns it directly.
func resolveAccountID(ctx context.Context, client *jira.Client, input string) (string, error) {
	if !strings.Contains(input, "@") {
		return input, nil
	}
	return client.ResolveEmailToAtlassianUserID(ctx, input)
}

// ensureSprintFieldID lazy-loads the Sprint custom field ID.
// It checks config first, and falls back to querying the Jira API if missing.
func ensureSprintFieldID(ctx context.Context, client *jira.Client) (string, error) {
	sprintFieldID, _ := config.Cfg.Get(JiraConfigSprintCustomFieldKey)
	if sprintFieldID != "" {
		return sprintFieldID, nil
	}

	fmt.Println("Sprint custom field not found in config. Fetching from Jira API...")
	fieldClient := jira.NewFieldService(client)
	jiraFields, err := fieldClient.GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch fields for initialization: %w", err)
	}

	for _, field := range jiraFields {
		if strings.ToLower(field.Name) == "sprint" {
			if err := config.Cfg.Upsert(JiraConfigSprintCustomFieldKey, field.ID); err != nil {
				return "", fmt.Errorf("failed to save sprint field to config: %w", err)
			}
			fmt.Println("Sprint custom field successfully mapped and saved!")
			return field.ID, nil
		}
	}

	return "", fmt.Errorf("could not find a custom field named 'Sprint' in this Jira workspace")
}
