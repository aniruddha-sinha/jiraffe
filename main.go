package main

import (
	"fmt"
	"os"

	"github.com/aniruddha-sinha/jiraffe/cmd"
)

func main() {
	if err := cmd.NewJiraffeCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
