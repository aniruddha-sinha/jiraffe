package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/aniruddha-sinha/jiraffe/internal/creds"
)

func newCredentialsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "credentials",
		Aliases: []string{"c", "creds"},
		Short:   "command to grab credentials",
	}

	cmd.AddCommand(newJiraCredentialsCmd())

	return cmd
}

func newJiraCredentialsCmd() *cobra.Command {
	var (
		email               string
		apiToken            string
		ErrTokenReadFailure = errors.New("failed to read token")
	)

	validate := validator.New()

	const defaultJiraOrg = "asinha0493"

	cmd := &cobra.Command{
		Use:           "jira",
		Aliases:       []string{"j"},
		Short:         "command to grab jira credentials",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("enter atlassian registered email -> ")
			_, err := fmt.Scanf("%s", &email)
			if err != nil {
				return err
			}

			email = strings.TrimSpace(email)

			if err := validate.Var(email, "email"); err != nil {
				return fmt.Errorf("wrong email format")
			}

			fmt.Println("click the link given below to generate API token")
			fmt.Println("https://id.atlassian.com/manage-profile/security/api-tokens")
			fmt.Println("enter API Token -> ")
			bytePass, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("%w, %w", ErrTokenReadFailure, err)
			}

			token := strings.TrimSpace(string(bytePass))
			fmt.Println("obtained token")
			authStr := fmt.Sprintf("%s:%s", email, token)
			apiToken = base64.StdEncoding.EncodeToString([]byte(authStr))

			jc := creds.NewJiraCreds(email, defaultJiraOrg, apiToken)
			if err := jc.Store(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
