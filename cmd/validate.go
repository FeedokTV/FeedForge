package cmd

import (
	"feedforge/internal/normalize"
	"feedforge/internal/sink"
	"fmt"

	"github.com/spf13/cobra"
)

type validateOptions struct {
	InputPath string
}

var validateOpts validateOptions

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Read a canonical NDJSON stream matches the canonical schema",
	Example: `  feedforge validate < clean.ndjson
  feedforge validate -i clean.ndjson`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runValidate(cmd)
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validateOpts.InputPath, "input", "i", "", "input file (default: stdin)")
	rootCmd.AddCommand(validateCmd)
}

var knownTypes = map[normalize.Type]struct{}{
	normalize.TypeURL:    {},
	normalize.TypeDomain: {},
	normalize.TypeIP:     {},
	normalize.TypeIPv6:   {},
	normalize.TypeHash:   {},
	normalize.TypeEmail:  {},
}

func checkRecord(r normalize.Record) []string {
	var errs []string
	if r.ID == "" {
		errs = append(errs, "id is empty")
	}
	if _, ok := knownTypes[r.Type]; !ok {
		errs = append(errs, fmt.Sprintf("type %q is not a known type", r.Type))
	}
	if r.Source == "" {
		errs = append(errs, "source is empty")
	}
	if r.Value == "" {
		errs = append(errs, "value is empty")
	}
	if r.FirstSeen.IsZero() {
		errs = append(errs, "first_seen is zero")
	}
	return errs
}

func runValidate(cmd *cobra.Command) error {
	input, err := openInput(validateOpts.InputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	stderr := cmd.ErrOrStderr()
	var total, invalid int

	for rec, err := range sink.ReadNDJSON(input) {
		total++
		if err != nil {
			fmt.Fprintf(stderr, "record %d: invalid JSON: %v\n", total, err)
			invalid++
			continue
		}

		if problems := checkRecord(rec); len(problems) > 0 {
			invalid++
			for _, p := range problems {
				fmt.Fprintf(stderr, "record %d (id=%q): %s\n", total, rec.ID, p)
			}
		}
	}

	if invalid == 0 {
		fmt.Fprintf(stderr, "all %d records valid\n", total)
		return nil
	}

	return fmt.Errorf("%d/%d records failed validation", invalid, total)
}
