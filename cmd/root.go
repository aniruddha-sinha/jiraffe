package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aniruddha-sinha/jiraffe/internal/config"
	"github.com/spf13/cobra"
)

const (
	configFile = "config.yaml" // Note: Earlier we used credentials.json, make sure Viper knows to parse YAML if you use .yaml!
)

func NewJiraffeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "jiraffe",
		Short:         "jiraffe is a suite of tools for interacting with atlassian products such as jira, confluence",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if os.Geteuid() == 0 {
				log.Println("Warning /!\\ : running jiraffe as root can be dangerous")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Print(err)
			}
		},
	}

	// cobra.OnInitialize ensures the config is only loaded when a command actually runs,
	// preventing disk writes during simple tasks like 'jiraffe --help'
	cobra.OnInitialize(func() {
		config.Cfg = config.New()
		if err := config.Cfg.InitConfig(configFile); err != nil {
			// Fixed to log.Fatalf and %v
			log.Fatalf("failed to load config file: %v", err)
		}
	})

	cmd.AddCommand(
		newCredentialsCmd(),
		newCmdJira(),
	)

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	return cmd
}
