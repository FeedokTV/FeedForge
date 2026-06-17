package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"feedforge/internal/parse"
	"feedforge/internal/profile"

	"github.com/spf13/cobra"
)

type inspectOptions struct {
	InputPath string
}

var inspectOpts inspectOptions

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Describe a raw feed file without converting it",
	Example: `  feedforge inspect mystery.csv
  feedforge inspect -i partner.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInspect(cmd)
	},
}

func init() {
	inspectCmd.Flags().StringVarP(&inspectOpts.InputPath, "input", "i", "", "input file (default: stdin)")
	rootCmd.AddCommand(inspectCmd)
}

func runInspect(cmd *cobra.Command) error {
	input, err := openInput(inspectOpts.InputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	format, buf, err := parse.DetectFormat(input)
	if err != nil {
		return err
	}

	peek, _ := buf.Peek(3)
	hasBOM := len(peek) >= 3 && peek[0] == 0xEF && peek[1] == 0xBB && peek[2] == 0xBF

	encoding := "UTF-8"
	if hasBOM {
		encoding = "UTF-8 with BOM"
	}

	ctx := context.Background()
	var headers []string
	var rows int

	switch format {
	case "csv":
		parser := &parse.CSVParser{}
		seen := false
		for rec, err := range parser.Parse(ctx, buf) {
			if err != nil {
				continue
			}
			if !seen {
				seen = true
				for k := range rec.Fields {
					headers = append(headers, k)
				}
				sort.Strings(headers)
			}
			rows++
		}
	case "jsonl":
		parser := &parse.JSONLParser{}
		keySet := make(map[string]struct{})
		for rec, err := range parser.Parse(ctx, buf) {
			if err != nil {
				continue
			}
			rows++
			for k := range rec.Fields {
				keySet[k] = struct{}{}
			}
		}
		for k := range keySet {
			headers = append(headers, k)
		}
		sort.Strings(headers)
	case "list":
		parser := &parse.ListParser{}
		for _, err := range parser.Parse(ctx, buf) {
			if err != nil {
				continue
			}
			rows++
		}
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Format:     %s (detected)\n", strings.ToUpper(format))
	fmt.Fprintf(out, "Encoding:   %s\n", encoding)
	fmt.Fprintf(out, "Rows:       %d\n", rows)

	if len(headers) > 0 {
		fmt.Fprintf(out, "Columns:    %d\n", len(headers))
		fmt.Fprintf(out, "Headers:    %s\n", strings.Join(headers, ", "))
	}

	if suggested := suggestProfile(format, headers); suggested != "" {
		fmt.Fprintf(out, "Suggested:  --source %s  (column signature matches built-in profile)\n", suggested)
	}

	return nil
}

var builtinProfileNames = []string{"urlhaus", "openphish", "threatfox", "generic-csv", "generic-list"}

func suggestProfile(format string, headers []string) string {
	headerSet := make(map[string]struct{}, len(headers))
	for _, h := range headers {
		headerSet[h] = struct{}{}
	}
	for _, name := range builtinProfileNames {
		p, err := profile.LoadBuiltin(name)
		if err != nil || p.Format != format {
			continue
		}
		profileCols := make(map[string]struct{}, len(p.Columns))
		for _, col := range p.Columns {
			profileCols[col.Name] = struct{}{}
		}
		if len(profileCols) != len(headerSet) {
			continue
		}
		match := true
		for k := range profileCols {
			if _, ok := headerSet[k]; !ok {
				match = false
				break
			}
		}
		if match {
			return name
		}
	}
	return ""
}
