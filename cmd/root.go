/**
* *! Aniruddha Sinha
**/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jiraffe",
	Short: "A CLI application to interact with Atlassian (JIRA)",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}
