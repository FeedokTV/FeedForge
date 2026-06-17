package sink

import (
	"encoding/json"
	"feedforge/internal/normalize"
	"io"
	"iter"
	"log/slog"
)

func ReadNDJSON(r io.Reader) iter.Seq2[normalize.Record, error] {
	return func(yield func(normalize.Record, error) bool) {
		dec := json.NewDecoder(r)
		for dec.More() {
			var rec normalize.Record
			if err := dec.Decode(&rec); err != nil {
				if !yield(normalize.Record{}, err) {
					return
				}
				continue
			}
			if !yield(rec, nil) {
				return
			}
		}
	}
}

func SaveNDJSON(writer io.Writer, in iter.Seq2[normalize.Record, error]) error {
	enc := json.NewEncoder(writer)
	enc.SetEscapeHTML(false)

	for rec, err := range in {
		if err != nil {
			slog.Warn("skipping record", "error", err)
			continue
		}

		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil
}
