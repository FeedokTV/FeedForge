package testhelpers

import (
	"feedforge/internal/normalize"
	"iter"
)

func RecordsToSeq(records []normalize.Record, err error) iter.Seq2[normalize.Record, error] {
	return func(yield func(normalize.Record, error) bool) {
		for _, r := range records {
			if !yield(r, err) {
				return
			}
		}
	}
}
