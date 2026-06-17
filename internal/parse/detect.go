package parse

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"strings"
)

type AutoParser struct{}

func DetectFormat(r io.Reader) (string, *bufio.Reader, error) {
	const peekSize = 4 * 1024
	buf := bufio.NewReaderSize(r, peekSize)
	peek, _ := buf.Peek(peekSize)
	if len(peek) == 0 {
		return "", buf, fmt.Errorf("empty input")
	}
	return detectFormat(peek), buf, nil
}

func detectFormat(peek []byte) string {
	s := strings.TrimSpace(string(peek))
	if strings.HasPrefix(s, "{") {
		return "jsonl"
	}

	// Count commas vs newlines
	lines := strings.Count(s, "\n") + 1
	commas := strings.Count(s, ",")
	if commas >= lines {
		return "csv"
	}

	return "list"
}

func (p *AutoParser) Parse(ctx context.Context, reader io.Reader) iter.Seq2[RawRecord, error] {
	return func(yield func(RawRecord, error) bool) {
		const peekSize = 4 * 1024

		buf := bufio.NewReaderSize(reader, peekSize)
		peek, _ := buf.Peek(peekSize)

		if len(peek) == 0 {
			return
		}

		format := detectFormat(peek)
		slog.Info("auto-detected format", "format", format)

		var inner Parser
		switch format {
		case "jsonl":
			inner = &JSONLParser{}
		case "csv":
			inner = &CSVParser{}
		default:
			inner = &ListParser{}
		}

		for rec, err := range inner.Parse(ctx, buf) {
			if !yield(rec, err) {
				return
			}
		}
	}
}
