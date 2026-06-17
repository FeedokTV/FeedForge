package filter_test

import (
	"feedforge/internal/filter"
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	tests := []struct {
		Name    string
		Tokens  []filter.Token
		Want    filter.Expr
		WantErr bool
	}{
		{
			Name: "simple comparison",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenEOF},
			},
			Want: &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
		},
		{
			Name: "and of two comparisons",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenAND, Literal: "AND"},
				{Kind: filter.TokenField, Literal: "confidence"},
				{Kind: filter.TokenOp, Literal: ">="},
				{Kind: filter.TokenValue, Literal: "75"},
				{Kind: filter.TokenEOF},
			},
			Want: &filter.AndExpr{
				Left:  &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
				Right: &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "75"},
			},
		},
		{
			Name: "or of two comparisons",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "source"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "urlhaus"},
				{Kind: filter.TokenOR, Literal: "OR"},
				{Kind: filter.TokenField, Literal: "source"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "openphish"},
				{Kind: filter.TokenEOF},
			},
			Want: &filter.OrExpr{
				Left:  &filter.CompareExpr{Field: "source", Op: "=", Value: "urlhaus"},
				Right: &filter.CompareExpr{Field: "source", Op: "=", Value: "openphish"},
			},
		},
		{
			Name: "and binds tighter than or",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenAND, Literal: "AND"},
				{Kind: filter.TokenField, Literal: "confidence"},
				{Kind: filter.TokenOp, Literal: ">="},
				{Kind: filter.TokenValue, Literal: "75"},
				{Kind: filter.TokenOR, Literal: "OR"},
				{Kind: filter.TokenField, Literal: "source"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "openphish"},
				{Kind: filter.TokenEOF},
			},
			Want: &filter.OrExpr{
				Left: &filter.AndExpr{
					Left:  &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
					Right: &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "75"},
				},
				Right: &filter.CompareExpr{Field: "source", Op: "=", Value: "openphish"},
			},
		},
		{
			Name: "not negates a single atom",
			Tokens: []filter.Token{
				{Kind: filter.TokenNOT, Literal: "NOT"},
				{Kind: filter.TokenField, Literal: "source"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "urlhaus"},
				{Kind: filter.TokenEOF},
			},
			Want: &filter.NotExpr{
				Inner: &filter.CompareExpr{Field: "source", Op: "=", Value: "urlhaus"},
			},
		},
		{
			Name: "parentheses override precedence",
			Tokens: []filter.Token{
				{Kind: filter.TokenLParen, Literal: "("},
				{Kind: filter.TokenField, Literal: "source"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "urlhaus"},
				{Kind: filter.TokenOR, Literal: "OR"},
				{Kind: filter.TokenField, Literal: "source"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "openphish"},
				{Kind: filter.TokenRParen, Literal: ")"},
				{Kind: filter.TokenAND, Literal: "AND"},
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenEOF},
			},
			Want: &filter.AndExpr{
				Left: &filter.OrExpr{
					Left:  &filter.CompareExpr{Field: "source", Op: "=", Value: "urlhaus"},
					Right: &filter.CompareExpr{Field: "source", Op: "=", Value: "openphish"},
				},
				Right: &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
			},
		},
		{
			Name: "in expression with list",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "in"},
				{Kind: filter.TokenList, Literal: "url,domain", Values: []string{"url", "domain"}},
				{Kind: filter.TokenEOF},
			},
			Want: &filter.InExpr{Field: "type", Values: []string{"url", "domain"}},
		},
		{
			Name: "missing operator after field",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenEOF},
			},
			WantErr: true,
		},
		{
			Name: "missing value after operator",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenEOF},
			},
			WantErr: true,
		},
		{
			Name: "in without a list",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "in"},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenEOF},
			},
			WantErr: true,
		},
		{
			Name: "unclosed parenthesis",
			Tokens: []filter.Token{
				{Kind: filter.TokenLParen, Literal: "("},
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenEOF},
			},
			WantErr: true,
		},
		{
			Name: "trailing garbage after valid expression",
			Tokens: []filter.Token{
				{Kind: filter.TokenField, Literal: "type"},
				{Kind: filter.TokenOp, Literal: "="},
				{Kind: filter.TokenValue, Literal: "url"},
				{Kind: filter.TokenRParen, Literal: ")"},
				{Kind: filter.TokenEOF},
			},
			WantErr: true,
		},
		{
			Name: "empty token stream",
			Tokens: []filter.Token{
				{Kind: filter.TokenEOF},
			},
			WantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			p := filter.NewParser(tt.Tokens)
			got, err := p.ParseExpr()

			if tt.WantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (AST: %+v)", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("got %+v, want %+v", got, tt.Want)
			}

		})
	}
}

func TestParseIntegration(t *testing.T) {
	tests := []struct {
		Name    string
		Input   string
		Want    filter.Expr
		WantErr bool
	}{
		{
			Name:  "simple comparison",
			Input: "type=url",
			Want:  &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
		},
		{
			Name:  "and with confidence",
			Input: "type=url AND confidence>=75",
			Want: &filter.AndExpr{
				Left:  &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
				Right: &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "75"},
			},
		},
		{
			Name:  "tags contains with quoted value",
			Input: `tags contains 'malware download'`,
			Want:  &filter.CompareExpr{Field: "tags", Op: "contains", Value: "malware download"},
		},
		{
			Name:  "in expression",
			Input: "type in [url, domain]",
			Want:  &filter.InExpr{Field: "type", Values: []string{"url", "domain"}},
		},
		{
			Name:  "not with parens",
			Input: "NOT (source=urlhaus)",
			Want: &filter.NotExpr{
				Inner: &filter.CompareExpr{Field: "source", Op: "=", Value: "urlhaus"},
			},
		},
		{
			Name:  "full precedence example from the build guide",
			Input: "type=url AND tags contains malware_download",
			Want: &filter.AndExpr{
				Left:  &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
				Right: &filter.CompareExpr{Field: "tags", Op: "contains", Value: "malware_download"},
			},
		},
		{
			Name:    "lex error propagates as parse error",
			Input:   `value="unterminated`,
			WantErr: true,
		},
		{
			Name:    "malformed operator",
			Input:   "type!url",
			WantErr: true,
		},
		{
			Name:    "stray closing paren",
			Input:   "type=url)",
			WantErr: true,
		},
		{
			Name:    "empty expression",
			Input:   "",
			WantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := filter.Parse(tt.Input)

			if tt.WantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (AST: %+v)", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("got %+v, want %+v", got, tt.Want)
			}
		})
	}
}
