package filter_test

import (
	"feedforge/internal/filter"
	"reflect"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		Name    string
		Input   string
		Want    []filter.Token
		WantErr bool
	}{
		{
			Name:  "simple comparison",
			Input: "type=url",
			Want: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenEOF},
			},
		},
		{
			Name:  "and with spaces",
			Input: "type=url AND confidence>=75",
			Want: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenAND, Literal: "AND"},
				{Kind: filter.TokenField, Literal: "confidence"},
				{Kind: filter.TokenOp, Literal: ">="},
				{Kind: filter.TokenValue, Literal: "75"},
				{Kind: filter.TokenEOF},
			},
		},
		{
			Name:  "ends with no trailing space",
			Input: "type=url",
			Want: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenEOF},
			},
		},
		{
			Name:  "list",
			Input: "type in [url, domain]",
			Want: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "in"},
				{Kind: filter.TokenList, Literal: "url,domain", Values: []string{"url", "domain"}},
				{Kind: filter.TokenEOF},
			},
		},
		{
			Name:  "quoted value with space",
			Input: `tags contains 'malware download'`,
			Want: []filter.Token{
				{Kind: filter.TokenField, Literal: "tags"},
				{Kind: filter.TokenOp, Literal: "contains"},
				{Kind: filter.TokenValue, Literal: "malware download"},
				{Kind: filter.TokenEOF},
			},
		},
		{
			Name:    "unterminated string",
			Input:   `value="abc`,
			WantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			l := filter.NewLexer(tt.Input)
			var got []filter.Token
			for {
				tok := l.Next()
				got = append(got, tok)
				if tok.Kind == filter.TokenEOF || tok.Kind == filter.TokenError {
					break
				}
			}

			lastTok := got[len(got)-1]

			if tt.WantErr {
				if lastTok.Kind != filter.TokenError {
					t.Errorf("expected TokenError, got %v", lastTok)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("got %+v, want %+v", got, tt.Want)
			}
		})
	}
}
