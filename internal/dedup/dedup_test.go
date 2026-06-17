package dedup_test

import (
	"feedforge/internal/dedup"
	"feedforge/internal/helpers"
	"feedforge/internal/normalize"
	"feedforge/internal/runtime"
	"feedforge/internal/testhelpers"
	"testing"
	"time"
)

func TestDedup(t *testing.T) {
	now := time.Date(2026, 6, 4, 23, 27, 26, 0, time.UTC)

	makeRec := func(value string) normalize.Record {
		return normalize.Record{
			ID:         normalize.GenerateRowID(normalize.TypeURL, value),
			Type:       normalize.TypeURL,
			Source:     "test-source",
			Value:      value,
			FirstSeen:  now,
			Confidence: helpers.IntToPointer(90),
		}
	}

	recA := makeRec("http://example.com")
	recB := makeRec("http://malicious.net")
	recC := makeRec("http://other.example.com")

	tests := []struct {
		Name  string
		Input []normalize.Record
		Want  int // final count of rows
	}{
		{
			Name:  "deduplicates identical records",
			Input: []normalize.Record{recA, recA, recB},
			Want:  2,
		},
		{
			Name:  "all unique pass",
			Input: []normalize.Record{recA, recB, recC},
			Want:  3,
		},
		{
			Name:  "all duplicates",
			Input: []normalize.Record{recA, recA, recA},
			Want:  1,
		},
		{
			Name:  "empty input",
			Input: []normalize.Record{},
			Want:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := t.Context()

			recStream := testhelpers.RecordsToSeq(test.Input, nil)
			dedupedStream := dedup.Dedup(ctx, recStream, runtime.NewStats())

			var got int
			for _, err := range dedupedStream {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					continue
				}
				got++
			}

			if got != test.Want {
				t.Errorf("got %d records, want %d", got, test.Want)
			}
		})
	}
}
