package parse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
)

type JSONLParser struct{}

func (p *JSONLParser) Parse(ctx context.Context, reader io.Reader) iter.Seq2[RawRecord, error] {
	return func(yield func(RawRecord, error) bool) {
		dec := json.NewDecoder(reader)
		lineNo := 0

		for dec.More() {
			if ctx.Err() != nil {
				return
			}

			lineNo++

			var raw map[string]any
			if err := dec.Decode(&raw); err != nil {
				if !yield(RawRecord{LineNum: lineNo}, fmt.Errorf("line %d: %w", lineNo, err)) {
					return
				}
				continue
			}

			fields := make(map[string]string, len(raw))
			for k, v := range raw {
				switch val := v.(type) {
				case string:
					fields[k] = val
				default:
					fields[k] = fmt.Sprintf("%v", val)
				}
			}

			if !yield(RawRecord{LineNum: lineNo, Fields: fields}, nil) {
				return
			}
		}
	}
}
