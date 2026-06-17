package parse_test

import (
	"context"
	"strings"
	"testing"

	"feedforge/internal/parse"
)

func TestAutoParser_detectsJSONL(t *testing.T) {
	input := `{"value":"http://evil.com","type":"url"}` + "\n"

	parser, err := parse.New("auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got int
	for _, err := range parser.Parse(context.Background(), strings.NewReader(input)) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got++
	}

	if got != 1 {
		t.Errorf("got %d records, want 1", got)
	}
}

func TestAutoParser_detectsCSV(t *testing.T) {
	input := "value,type\nhttp://evil.com,url\n"

	parser, err := parse.New("auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got int
	for _, err := range parser.Parse(context.Background(), strings.NewReader(input)) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got++
	}

	if got != 1 {
		t.Errorf("got %d records, want 1", got)
	}
}

func TestAutoParser_detectsList(t *testing.T) {
	input := "http://evil.com\nhttp://bad.com\n"

	parser, err := parse.New("auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got int
	for _, err := range parser.Parse(context.Background(), strings.NewReader(input)) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got++
	}

	if got != 2 {
		t.Errorf("got %d records, want 2", got)
	}
}
