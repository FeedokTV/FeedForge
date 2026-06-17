package parse_test

import (
	"context"
	"feedforge/internal/parse"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseCSV(t *testing.T) {
	testCSV := `id,value,something,date
1,"http://example.com","qwerty","2026-06-04 23:27:26"
`

	want := []parse.RawRecord{
		{
			LineNum: 2,
			Fields: map[string]string{
				"id":        "1",
				"value":     "http://example.com",
				"something": "qwerty",
				"date":      "2026-06-04 23:27:26",
			},
		},
	}

	reader := strings.NewReader(testCSV)

	parser, err := parse.New("csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := parser.Parse(context.Background(), reader)

	var parsed []parse.RawRecord
	for rawRecord, err := range raw {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsed = append(parsed, rawRecord)
	}

	if diff := cmp.Diff(want, parsed); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}
}
