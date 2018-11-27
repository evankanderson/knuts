package cmd

import (
	"fmt"
	"os"

	"github.com/evankanderson/knuts/pkg"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "knuts",
	Short:   "Knuts is an install and management utility for Knative",
	Version: "0.1",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&pkg.DryRun, "dry_run", true, "When true, print operations rather than executing them.")
}

// Execute is the root Cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
