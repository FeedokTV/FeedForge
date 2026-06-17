package cmd

import (
	"feedforge/internal/dedup"
	"feedforge/internal/filter"
	"feedforge/internal/normalize"
	"feedforge/internal/parse"
	"feedforge/internal/profile"
	"feedforge/internal/runtime"
	"feedforge/internal/sink"
	"fmt"

	"github.com/spf13/cobra"
)

type convertOptions struct {
	InputPath   string
	InputFormat string
	Delimiter   string

	Source      string
	ProfilePath string

	OutputPath   string
	OutputFormat string

	Filter     string
	Dedup      bool
	ErrorsFile string
}

var convertOpts convertOptions

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a feed into canonical NDJSON",
	Long:  `Convert a heterogeneous IoC feed (CSV, JSONL, plain text) into a canonical NDJSON stream.`,
	Example: `  feedforge convert --in csv --source urlhaus -i urlhaus.csv
  cat feed.csv | feedforge convert --in csv --source urlhaus`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConvert(cmd)
	},
}

func init() {
	convertCmd.Flags().StringVarP(&convertOpts.InputFormat, "in", "f", "auto", "input format: csv|jsonl|list|auto")
	convertCmd.Flags().StringVar(&convertOpts.Source, "source", "", "built-in source profile (e.g. urlhaus, openphish)")
	convertCmd.Flags().StringVar(&convertOpts.ProfilePath, "profile", "", "path to custom YAML profile")
	convertCmd.Flags().StringVarP(&convertOpts.InputPath, "input", "i", "", "input file (default: stdin)")
	convertCmd.Flags().StringVarP(&convertOpts.OutputPath, "output", "o", "", "output file (default: stdout)")
	convertCmd.Flags().StringVar(&convertOpts.OutputFormat, "out", "ndjson", "output format: ndjson|csv")
	convertCmd.Flags().StringVar(&convertOpts.Filter, "where", "", "filter expression")
	convertCmd.Flags().BoolVar(&convertOpts.Dedup, "dedup", false, "deduplicate records by stable ID")

	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command) error {
	ctx := cmd.Context()

	if convertOpts.Source == "" && convertOpts.ProfilePath == "" {
		return fmt.Errorf("one of --source or --profile is required")
	}
	if convertOpts.Source != "" && convertOpts.ProfilePath != "" {
		return fmt.Errorf("--source and --profile are mutually exclusive")
	}

	stats := runtime.NewStats()
	defer stats.Print(cmd.ErrOrStderr())

	input, err := openInput(convertOpts.InputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := openOutput(convertOpts.OutputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	parser, err := parse.New(convertOpts.InputFormat)
	if err != nil {
		return err
	}

	var prof *profile.Profile
	if convertOpts.Source != "" {
		prof, err = profile.LoadBuiltin(convertOpts.Source)
	} else {
		prof, err = profile.Load(convertOpts.ProfilePath)
	}

	if err != nil {
		return err
	}

	rawRows := parser.Parse(ctx, input)
	records := normalize.Map(prof, rawRows, stats)

	if convertOpts.Dedup {
		records = dedup.Dedup(ctx, records, stats)
	}

	if convertOpts.Filter != "" {
		f, err := filter.Parse(convertOpts.Filter)
		if err != nil {
			return fmt.Errorf("invalid filter expression: %w", err)
		}
		records = filter.Apply(f, records)
	}

	switch convertOpts.OutputFormat {
	case "ndjson":
		err = sink.SaveNDJSON(output, records)
	case "csv":
		err = sink.SaveCSV(output, records)
	default:
		err = sink.SaveNDJSON(output, records)
	}

	if err != nil {
		return err
	}

	return nil
}
