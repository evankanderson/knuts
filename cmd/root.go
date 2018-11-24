package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "knuts",
	Short:   "Knuts is an install and management utility for Knative",
	Version: "0.1",
}

// Execute is the root Cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
