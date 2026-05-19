package cmd

import (
	"fmt"

	"github.com/aniruddha-sinha/jiraffe/internal/config"
	"github.com/aniruddha-sinha/jiraffe/internal/jira"
	"github.com/go-playground/validator/v10"
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

	validate := validator.New()

	const (
		defaultJiraOrg = "asinha0493"
	)

	cmd := &cobra.Command{
		Use:     "auth",
		Aliases: []string{"a"},
		Short:   "Authenticate and save your Jira credentials",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if email != "" {
				if err := validate.Var(email, "required,email"); err != nil {
					return fmt.Errorf("invalid email format::: %w", err)
				}
			} else {
				storedEmail := config.Cfg.GetString(jira.JiraConfigEmailKey)
				if storedEmail == "" {
					return fmt.Errorf("no active session found; Please provide your atlassian registered email ID with the --email flag to login for the first time")
				}
			}

			if org != "" {
				if err := validate.Var(org, "required,hostname"); err != nil {
					return fmt.Errorf("invalid jira org format ::: %w", err)
				}
			} else {
				storedOrg := config.Cfg.GetString(jira.JiraConfigOrgKey)
				if storedOrg == "" {
					return fmt.Errorf("no jira org found, please pass on the jira org in the --org flag to login for the first time")
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			jc := jira.NewJiraCredentials(email, org, "")
			c := jira.NewClient()

			if err := c.HandleAuthentication(cmd.Context(), jc); err != nil {
				return err
			}

			fmt.Println("User Auth Success!!")
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "email registered with atlassian account")
	cmd.Flags().StringVarP(&org, "org", "o", defaultJiraOrg, "Atlassian organization name")

	return cmd
}
