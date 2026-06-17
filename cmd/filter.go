package cmd

import (
	"feedforge/internal/filter"
	"feedforge/internal/runtime"
	"feedforge/internal/sink"
	"fmt"

	"github.com/spf13/cobra"
)

type filterOptions struct {
	InputPath  string
	OutputPath string
	Where      string
}

var filterOpts filterOptions

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter canonical NDJSON by a boolean expression",
	Long:  `Read canonical NDJSON and emit only records matching a boolean expression.`,
	Example: `  feedforge filter --where 'type=url AND tags contains malware_download'
  cat feed.ndjson | feedforge filter --where 'confidence >= 75 AND source != threatfox'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFilter(cmd)
	},
}

func init() {
	filterCmd.Flags().StringVarP(&filterOpts.InputPath, "input", "i", "", "input file (default: stdin)")
	filterCmd.Flags().StringVarP(&filterOpts.OutputPath, "output", "o", "", "output file (default: stdout)")
	filterCmd.Flags().StringVar(&filterOpts.Where, "where", "", "filter expression (required)")
	_ = filterCmd.MarkFlagRequired("where")

	rootCmd.AddCommand(filterCmd)
}

func runFilter(cmd *cobra.Command) error {
	expr, err := filter.Parse(filterOpts.Where)
	if err != nil {
		return fmt.Errorf("invalid filter expression: %w", err)
	}

	input, err := openInput(filterOpts.InputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := openOutput(filterOpts.OutputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	stats := runtime.NewStats()
	defer stats.Print(cmd.ErrOrStderr())

	records := sink.ReadNDJSON(input)
	filtered := filter.Apply(expr, records)
	return sink.SaveNDJSON(output, filtered)
}
