package dedup

import (
	"context"
	"feedforge/internal/normalize"
	"feedforge/internal/runtime"
	"iter"
)

func Dedup(ctx context.Context, in iter.Seq2[normalize.Record, error], stats *runtime.Stats) iter.Seq2[normalize.Record, error] {
	return func(yield func(normalize.Record, error) bool) {
		var alreadySeen = make(map[string]struct{})

		for rec, err := range in {
			if ctx.Err() != nil {
				return
			}

			if err != nil {
				if !yield(rec, err) {
					return
				}
				continue
			}

			if _, ok := alreadySeen[rec.ID]; ok {
				stats.IncDropped("duplicate")
				continue
			}

			alreadySeen[rec.ID] = struct{}{}

			if !yield(rec, nil) {
				return
			}
		}
	}
}
