package main

import (
	"github.com/sakuradon99/pt/cmd"
	"os"

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
