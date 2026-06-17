package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sourcesCommand = &cobra.Command{
	Use:   "sources",
	Short: "Available builtin sources (profiles) for mapping",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSources(cmd)
	},
}

func init() {
	rootCmd.AddCommand(sourcesCommand)
}

func runSources(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Available sources:\n")
	fmt.Fprintf(out, "\t- Urlhaus profile\n")
	fmt.Fprintf(out, "\t- Openphish profile\n")
	fmt.Fprintf(out, "\t- ThreatFoxo profile\n")
	fmt.Fprintf(out, "\t- Generic CSV (expecting value field)\n")
	fmt.Fprintf(out, "\t- Generic list (one IoC per line)\n")

	return nil
}
