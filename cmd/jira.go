package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
		return nil
	},
}

func init() {
	// wiring the commands
	rootCmd.AddCommand(newJiraCmd)
	newJiraCmd.AddCommand(newAuthCmd)
}
