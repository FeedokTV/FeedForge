package sink

import (
	"encoding/csv"
	"feedforge/internal/normalize"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

var csvHeaders = []string{
	"id", "type", "source", "value",
	"first_seen", "last_seen", "tags", "confidence",
}

func SaveCSV(writer io.Writer, in iter.Seq2[normalize.Record, error]) error {
	csvWriter := csv.NewWriter(writer)

	if err := csvWriter.Write(csvHeaders); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for rec, err := range in {
		if err != nil {
			slog.Warn("skipping record", "error", err)
			continue
		}

		row := recordToRow(rec)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}

	return nil

}

func recordToRow(r normalize.Record) []string {
	lastSeen := ""
	if r.LastSeen != nil {
		lastSeen = r.LastSeen.UTC().Format(time.RFC3339)
	}

	confidence := ""
	if r.Confidence != nil {
		confidence = strconv.Itoa(*r.Confidence)
	}

	return []string{
		r.ID,
		string(r.Type),
		r.Source,
		r.Value,
		r.FirstSeen.UTC().Format(time.RFC3339),
		lastSeen,
		strings.Join(r.Tags, "|"),
		confidence,
	}
}
