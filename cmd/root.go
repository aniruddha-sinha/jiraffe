package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aniruddha-sinha/jiraffe/internal/config"
	"github.com/spf13/cobra"
)

const (
	configFile = "config.yaml"
	appName    = "jiraffe"
)

func NewJiraffeCommand() *cobra.Command {
	config.Cfg = config.New()
	if err := config.Cfg.InitConfig(appName, configFile); err != nil {
		log.Fatal("failed to load config file. %w", err)
	}

	cmd := &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() == 0 {
				log.Println("running the jiraffe as a root can be dangerous")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Print(err)
			}
		},
		Use:           "jiraffe",
		Short:         "jiraffe is a suite of tools for interacting with atlassian products such as jira, confluence",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(
		newJiraCmd(),
	)

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	return cmd
}
