package parse_test

import (
	"feedforge/internal/parse"
	"reflect"
	"testing"
)

func TestNewParser(t *testing.T) {
	tests := []struct {
		Name    string
		Input   string
		Want    interface{}
		WantErr bool
	}{
		{Name: "new csv parser", Input: "csv", Want: reflect.TypeOf(&parse.CSVParser{}), WantErr: false},
		{Name: "new jsonl parser", Input: "jsonl", Want: reflect.TypeOf(&parse.JSONLParser{}), WantErr: false},
		{Name: "new list parser", Input: "list", Want: reflect.TypeOf(&parse.ListParser{}), WantErr: false},
		{Name: "new auto parser", Input: "auto", Want: reflect.TypeOf(&parse.AutoParser{}), WantErr: false},
		{Name: "unknown format", Input: "unknown_format", WantErr: true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := parse.New(test.Input)

			if test.WantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if test.Want != reflect.TypeOf(got) {
				t.Fatalf("got %T, want %v", got, test.Want)
			}
		})
	}
}
