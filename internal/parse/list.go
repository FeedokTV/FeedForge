package parse

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"iter"
	"strings"
)

type ListParser struct{}

func (p *ListParser) Parse(ctx context.Context, reader io.Reader) iter.Seq2[RawRecord, error] {
	return func(yield func(RawRecord, error) bool) {
		const maxLineSize = 16 * 1024 * 1024 // 16 MB ceiling

		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 64*1024), maxLineSize)

		lineNo := 0

		for scanner.Scan() {
			if ctx.Err() != nil {
				return
			}

			lineNo++
			line := strings.TrimSpace(scanner.Text())

			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if !yield(RawRecord{LineNum: lineNo, Fields: map[string]string{"value": line}}, nil) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield(RawRecord{LineNum: lineNo}, fmt.Errorf("line %d: %w", lineNo, err))
		}
	}
}
