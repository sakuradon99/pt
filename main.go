package main

import (
	"os"
	"pt/cmd"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "pt"}

	rootCmd.AddCommand(cmd.NewTemplateCommand())
	rootCmd.AddCommand(cmd.NewCreateCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
