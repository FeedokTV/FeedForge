package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Current version of FeedForge",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersion(cmd)
	},
}

func init() {
	rootCmd.AddCommand(versionCommand)
}

func runVersion(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "FeedForge 0.1.0 (by Feedok)")

	return nil
}
