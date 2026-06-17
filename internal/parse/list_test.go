package parse_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"feedforge/internal/parse"
)

func TestParseList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []parse.RawRecord
	}{
		{
			name:  "single line",
			input: "http://evil.com\n",
			want: []parse.RawRecord{
				{LineNum: 1, Fields: map[string]string{"value": "http://evil.com"}},
			},
		},
		{
			name: "blank lines and comments skipped",
			input: "# comment\n" +
				"\n" +
				"http://evil.com\n" +
				"# another comment\n" +
				"http://bad.com\n",
			want: []parse.RawRecord{
				{LineNum: 3, Fields: map[string]string{"value": "http://evil.com"}},
				{LineNum: 5, Fields: map[string]string{"value": "http://bad.com"}},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace trimmed",
			input: "  http://evil.com  \n",
			want: []parse.RawRecord{
				{LineNum: 1, Fields: map[string]string{"value": "http://evil.com"}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := parse.New("list")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var got []parse.RawRecord
			for rec, err := range parser.Parse(context.Background(), strings.NewReader(tc.input)) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				got = append(got, rec)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
