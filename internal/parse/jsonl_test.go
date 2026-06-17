package parse_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"feedforge/internal/parse"
)

func TestParseJSONL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []parse.RawRecord
	}{
		{
			name:  "single object",
			input: `{"value":"http://evil.com","type":"url"}` + "\n",
			want: []parse.RawRecord{
				{LineNum: 1, Fields: map[string]string{"value": "http://evil.com", "type": "url"}},
			},
		},
		{
			name: "two objects",
			input: `{"value":"http://a.com"}` + "\n" +
				`{"value":"http://b.com"}` + "\n",
			want: []parse.RawRecord{
				{LineNum: 1, Fields: map[string]string{"value": "http://a.com"}},
				{LineNum: 2, Fields: map[string]string{"value": "http://b.com"}},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "numeric field coerced to string",
			input: `{"confidence":75}` + "\n",
			want: []parse.RawRecord{
				{LineNum: 1, Fields: map[string]string{"confidence": "75"}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := parse.New("jsonl")
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
