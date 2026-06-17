package filter_test

import (
	"errors"
	"feedforge/internal/filter"
	"feedforge/internal/normalize"
	"slices"
	"testing"
	"time"
)

func makeRecord(typ normalize.Type, source, value string, confidence *int, tags []string) normalize.Record {
	return normalize.Record{
		ID:         "abc123",
		Type:       typ,
		Source:     source,
		Value:      value,
		FirstSeen:  time.Now(),
		Confidence: confidence,
		Tags:       tags,
	}
}

func intPtr(n int) *int { return &n }

var baseURL = makeRecord(normalize.TypeURL, "urlhaus", "http://evil.com/malware", intPtr(80), []string{"malware_download", "botnet"})

func TestEval_Compare(t *testing.T) {
	tests := []struct {
		name string
		expr filter.Expr
		rec  normalize.Record
		want bool
	}{
		// type field (case-insensitive)
		{"type = url matches", &filter.CompareExpr{Field: "type", Op: "=", Value: "url"}, baseURL, true},
		{"type = URL case-insensitive", &filter.CompareExpr{Field: "type", Op: "=", Value: "URL"}, baseURL, true},
		{"type = domain no match", &filter.CompareExpr{Field: "type", Op: "=", Value: "domain"}, baseURL, false},
		{"type != domain matches", &filter.CompareExpr{Field: "type", Op: "!=", Value: "domain"}, baseURL, true},

		// source field (case-insensitive)
		{"source = urlhaus matches", &filter.CompareExpr{Field: "source", Op: "=", Value: "urlhaus"}, baseURL, true},
		{"source = URLHAUS case-insensitive", &filter.CompareExpr{Field: "source", Op: "=", Value: "URLHAUS"}, baseURL, true},
		{"source contains urlh", &filter.CompareExpr{Field: "source", Op: "contains", Value: "urlh"}, baseURL, true},
		{"source != openphish matches", &filter.CompareExpr{Field: "source", Op: "!=", Value: "openphish"}, baseURL, true},

		// value field (case-sensitive)
		{"value = exact match", &filter.CompareExpr{Field: "value", Op: "=", Value: "http://evil.com/malware"}, baseURL, true},
		{"value = wrong case no match", &filter.CompareExpr{Field: "value", Op: "=", Value: "HTTP://EVIL.COM/MALWARE"}, baseURL, false},
		{"value contains evil", &filter.CompareExpr{Field: "value", Op: "contains", Value: "evil"}, baseURL, true},
		{"value != mismatch", &filter.CompareExpr{Field: "value", Op: "!=", Value: "other"}, baseURL, true},

		// id field
		{"id = abc123", &filter.CompareExpr{Field: "id", Op: "=", Value: "abc123"}, baseURL, true},
		{"id != xyz", &filter.CompareExpr{Field: "id", Op: "!=", Value: "xyz"}, baseURL, true},

		// confidence numeric comparisons
		{"confidence >= 75 matches", &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "75"}, baseURL, true},
		{"confidence >= 80 matches", &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "80"}, baseURL, true},
		{"confidence >= 81 no match", &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "81"}, baseURL, false},
		{"confidence <= 80 matches", &filter.CompareExpr{Field: "confidence", Op: "<=", Value: "80"}, baseURL, true},
		{"confidence > 79 matches", &filter.CompareExpr{Field: "confidence", Op: ">", Value: "79"}, baseURL, true},
		{"confidence < 81 matches", &filter.CompareExpr{Field: "confidence", Op: "<", Value: "81"}, baseURL, true},
		{"confidence = 80 matches", &filter.CompareExpr{Field: "confidence", Op: "=", Value: "80"}, baseURL, true},
		{"confidence != 80 no match", &filter.CompareExpr{Field: "confidence", Op: "!=", Value: "80"}, baseURL, false},

		// nil confidence treated as 0
		{"confidence nil >= 0", &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "0"}, makeRecord(normalize.TypeIP, "x", "1.2.3.4", nil, nil), true},
		{"confidence nil >= 1 no match", &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "1"}, makeRecord(normalize.TypeIP, "x", "1.2.3.4", nil, nil), false},

		// tags contains — membership
		{"tags contains malware_download matches", &filter.CompareExpr{Field: "tags", Op: "contains", Value: "malware_download"}, baseURL, true},
		{"tags contains unknown no match", &filter.CompareExpr{Field: "tags", Op: "contains", Value: "unknown"}, baseURL, false},
		{"tags = botnet matches", &filter.CompareExpr{Field: "tags", Op: "=", Value: "botnet"}, baseURL, true},
		{"tags != unknown matches", &filter.CompareExpr{Field: "tags", Op: "!=", Value: "unknown"}, baseURL, true},
		{"tags != botnet no match", &filter.CompareExpr{Field: "tags", Op: "!=", Value: "botnet"}, baseURL, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.Eval(tt.expr, tt.rec)
			if got != tt.want {
				t.Errorf("Eval(%+v) = %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestEval_In(t *testing.T) {
	tests := []struct {
		name string
		expr filter.Expr
		rec  normalize.Record
		want bool
	}{
		{"type in [url, domain] matches url", &filter.InExpr{Field: "type", Values: []string{"url", "domain"}}, baseURL, true},
		{"type in [domain, ip] no match", &filter.InExpr{Field: "type", Values: []string{"domain", "ip"}}, baseURL, false},
		{"type in [URL] case-insensitive", &filter.InExpr{Field: "type", Values: []string{"URL"}}, baseURL, true},
		{"source in [urlhaus, openphish] matches", &filter.InExpr{Field: "source", Values: []string{"urlhaus", "openphish"}}, baseURL, true},
		{"source in [URLHAUS] case-insensitive", &filter.InExpr{Field: "source", Values: []string{"URLHAUS"}}, baseURL, true},
		{"tags in [botnet] matches", &filter.InExpr{Field: "tags", Values: []string{"botnet"}}, baseURL, true},
		{"tags in [unknown] no match", &filter.InExpr{Field: "tags", Values: []string{"unknown"}}, baseURL, false},
		{"tags in [malware_download, botnet] either match", &filter.InExpr{Field: "tags", Values: []string{"malware_download", "botnet"}}, baseURL, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.Eval(tt.expr, tt.rec)
			if got != tt.want {
				t.Errorf("Eval(%+v) = %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestEval_Compound(t *testing.T) {
	tests := []struct {
		name string
		expr filter.Expr
		rec  normalize.Record
		want bool
	}{
		{
			"AND both true",
			&filter.AndExpr{
				Left:  &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
				Right: &filter.CompareExpr{Field: "source", Op: "=", Value: "urlhaus"},
			},
			baseURL, true,
		},
		{
			"AND left false short-circuits",
			&filter.AndExpr{
				Left:  &filter.CompareExpr{Field: "type", Op: "=", Value: "domain"},
				Right: &filter.CompareExpr{Field: "source", Op: "=", Value: "urlhaus"},
			},
			baseURL, false,
		},
		{
			"OR left true short-circuits",
			&filter.OrExpr{
				Left:  &filter.CompareExpr{Field: "type", Op: "=", Value: "url"},
				Right: &filter.CompareExpr{Field: "source", Op: "=", Value: "openphish"},
			},
			baseURL, true,
		},
		{
			"OR both false",
			&filter.OrExpr{
				Left:  &filter.CompareExpr{Field: "type", Op: "=", Value: "domain"},
				Right: &filter.CompareExpr{Field: "source", Op: "=", Value: "openphish"},
			},
			baseURL, false,
		},
		{
			"NOT true becomes false",
			&filter.NotExpr{Inner: &filter.CompareExpr{Field: "type", Op: "=", Value: "url"}},
			baseURL, false,
		},
		{
			"NOT false becomes true",
			&filter.NotExpr{Inner: &filter.CompareExpr{Field: "type", Op: "=", Value: "domain"}},
			baseURL, true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.Eval(tt.expr, tt.rec)
			if got != tt.want {
				t.Errorf("Eval = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApply(t *testing.T) {
	domain := makeRecord(normalize.TypeDomain, "openphish", "evil.example.com", nil, []string{"phishing"})
	ip := makeRecord(normalize.TypeIP, "threatfox", "1.2.3.4", intPtr(90), []string{"c2"})

	all := []normalize.Record{baseURL, domain, ip}

	seq := func(recs []normalize.Record) func(yield func(normalize.Record, error) bool) {
		return func(yield func(normalize.Record, error) bool) {
			for _, r := range recs {
				if !yield(r, nil) {
					return
				}
			}
		}
	}

	t.Run("filters to only url type", func(t *testing.T) {
		expr := &filter.CompareExpr{Field: "type", Op: "=", Value: "url"}
		var got []normalize.Record
		for rec, err := range filter.Apply(expr, seq(all)) {
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, rec)
		}
		if len(got) != 1 || got[0].Type != normalize.TypeURL {
			t.Errorf("expected 1 url record, got %v", got)
		}
	})

	t.Run("errors pass through unchanged", func(t *testing.T) {
		sentinel := errors.New("upstream error")
		errSeq := func(yield func(normalize.Record, error) bool) {
			yield(normalize.Record{}, sentinel)
		}
		expr := &filter.CompareExpr{Field: "type", Op: "=", Value: "url"}
		var errs []error
		for _, err := range filter.Apply(expr, errSeq) {
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) != 1 || !errors.Is(errs[0], sentinel) {
			t.Errorf("expected sentinel error, got %v", errs)
		}
	})

	t.Run("consumer stop is respected", func(t *testing.T) {
		expr := &filter.CompareExpr{Field: "type", Op: "!=", Value: ""}
		var got []normalize.Record
		for rec, err := range filter.Apply(expr, seq(all)) {
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, rec)
			break // stop after first
		}
		if len(got) != 1 {
			t.Errorf("expected 1 record after early stop, got %d", len(got))
		}
	})

	t.Run("high confidence filter", func(t *testing.T) {
		expr := &filter.CompareExpr{Field: "confidence", Op: ">=", Value: "85"}
		var got []normalize.Record
		for rec, err := range filter.Apply(expr, seq(all)) {
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, rec)
		}
		if len(got) != 1 || !slices.Equal(got[0].Tags, []string{"c2"}) {
			t.Errorf("expected only c2 record, got %v", got)
		}
	})
}

func TestEvalEndToEnd(t *testing.T) {
	tests := []struct {
		name  string
		input string
		rec   normalize.Record
		want  bool
	}{
		{"type=url AND tags contains malware_download", "type=url AND tags contains malware_download", baseURL, true},
		{"confidence >= 75 AND source != openphish", "confidence >= 75 AND source != openphish", baseURL, true},
		{"NOT (source=urlhaus)", "NOT (source=urlhaus)", baseURL, false},
		{"type in [url, domain]", "type in [url, domain]", baseURL, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := filter.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			got := filter.Eval(expr, tt.rec)
			if got != tt.want {
				t.Errorf("Eval = %v, want %v", got, tt.want)
			}
		})
	}
}
