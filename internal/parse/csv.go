package parse

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"slices"
)

type CSVParser struct {
	Delimiter rune // 0 means default ','
}

func removeBom(reader io.Reader) (io.Reader, error) {
	bom := make([]byte, 3)

	firstBytes, err := io.ReadAtLeast(reader, bom, 3)

	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return io.MultiReader(bytes.NewReader(bom[:firstBytes]), reader), nil
	}

	if bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		return reader, nil
	}

	return io.MultiReader(bytes.NewReader(bom), reader), nil
}

func (p *CSVParser) Parse(ctx context.Context, reader io.Reader) iter.Seq2[RawRecord, error] {
	return func(yield func(RawRecord, error) bool) {
		reader, err := removeBom(reader)
		if err != nil {
			if !yield(RawRecord{LineNum: 0}, fmt.Errorf("line %d, error: %w", 0, err)) {
				return
			}
		}

		csvReader := csv.NewReader(reader)

		if p.Delimiter != 0 {
			csvReader.Comma = p.Delimiter
		}

		csvReader.FieldsPerRecord = -1
		csvReader.ReuseRecord = true

		var headers []string
		lineNo := 0

		for {
			if ctx.Err() != nil {
				return
			}

			row, err := csvReader.Read()
			lineNo++

			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				if !yield(RawRecord{LineNum: lineNo}, fmt.Errorf("line %d, error: %w", lineNo, err)) {
					return
				}

				continue
			}

			if row[0] == "#" {
				continue
			}

			if headers == nil {
				headers = slices.Clone(row)
				continue
			}

			rowLength := len(row)
			fields := make(map[string]string, len(headers))
			for i, h := range headers {
				if i < rowLength {
					fields[h] = row[i]
				}
			}

			if !yield(RawRecord{LineNum: lineNo, Fields: fields}, nil) {
				return
			}
		}
	}
}
