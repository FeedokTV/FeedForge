package cmd

import (
	"feedforge/internal/runtime"
	"os"

	"github.com/spf13/cobra"
)

type globalOptions struct {
	Verbose  bool
	JSONLogs bool
}

var globalOpts globalOptions

var rootCmd = &cobra.Command{
	Use:   "feedforge",
	Short: "Convert heterogeneous IoC feeds into canonical NDJSON",
	Long: `feedforge converts heterogeneous IoC feeds into a canonical NDJSON schema.

Supports CSV, JSON Lines, and plain-text list formats. Built-in profiles for
common threat-intel sources; custom profiles via --profile. Pipe-friendly:
reads stdin, writes stdout by default.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		runtime.SetupLogger(globalOpts.Verbose, globalOpts.JSONLogs)
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&globalOpts.Verbose, "verbose", "v", false, "enable debug logging to stderr")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.JSONLogs, "json-logs", false, "emit logs as JSON instead of text")
}
