package parse

import (
	"context"
	"fmt"
	"io"
	"iter"
)

type RawRecord struct {
	LineNum int // for error messages
	Fields  map[string]string
}

type Parser interface {
	Parse(ctx context.Context, reader io.Reader) iter.Seq2[RawRecord, error]
}

func New(format string) (Parser, error) {
	switch format {
	case "csv":
		return &CSVParser{}, nil
	case "jsonl":
		return &JSONLParser{}, nil
	case "list":
		return &ListParser{}, nil
	case "auto":
		return &AutoParser{}, nil
	default:
		return nil, fmt.Errorf("unknown format %q", format)
	}
}
