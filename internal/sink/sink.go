package sink

import (
	"io"
	"iter"

	"feedforge/internal/normalize"
)

type Sink interface {
	Save(w io.Writer, in iter.Seq2[normalize.Record, error]) error
}

type NDJSONSink struct{}
type CSVSink struct{}

func (NDJSONSink) Save(w io.Writer, in iter.Seq2[normalize.Record, error]) error {
	return SaveNDJSON(w, in)
}

func (CSVSink) Save(w io.Writer, in iter.Seq2[normalize.Record, error]) error {
	return SaveCSV(w, in)
}
