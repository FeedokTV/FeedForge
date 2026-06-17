package filter

import (
	"iter"
	"strconv"
	"strings"

	"feedforge/internal/normalize"
)

// Eval evaluates an expression against a single record.
func Eval(expr Expr, rec normalize.Record) bool {
	switch e := expr.(type) {
	case *AndExpr:
		return Eval(e.Left, rec) && Eval(e.Right, rec)
	case *OrExpr:
		return Eval(e.Left, rec) || Eval(e.Right, rec)
	case *NotExpr:
		return !Eval(e.Inner, rec)
	case *CompareExpr:
		return evalCompare(e, rec)
	case *InExpr:
		return evalIn(e, rec)
	}
	return false
}

// Apply wraps an upstream iterator, yielding only records where Eval returns true.
// Errors from upstream pass through unchanged.
func Apply(f Expr, in iter.Seq2[normalize.Record, error]) iter.Seq2[normalize.Record, error] {
	return func(yield func(normalize.Record, error) bool) {
		for rec, err := range in {
			if err != nil {
				if !yield(rec, err) {
					return
				}
				continue
			}
			if Eval(f, rec) {
				if !yield(rec, nil) {
					return
				}
			}
		}
	}
}

// fieldString returns the string representation of a record field.
func fieldString(field string, rec normalize.Record) string {
	switch field {
	case "type":
		return string(rec.Type)
	case "source":
		return rec.Source
	case "value":
		return rec.Value
	case "id":
		return rec.ID
	case "confidence":
		if rec.Confidence != nil {
			return strconv.Itoa(*rec.Confidence)
		}
		return "0"
	}
	return ""
}

func isCaseInsensitive(field string) bool {
	return field == "type" || field == "source"
}

func evalCompare(e *CompareExpr, rec normalize.Record) bool {
	if e.Field == "tags" {
		return evalTagsCompare(e.Op, e.Value, rec.Tags)
	}
	if e.Field == "confidence" {
		fv := fieldString("confidence", rec)
		return evalNumeric(e.Op, fv, e.Value)
	}

	fv := fieldString(e.Field, rec)
	cv := e.Value
	if isCaseInsensitive(e.Field) {
		fv = strings.ToLower(fv)
		cv = strings.ToLower(cv)
	}

	switch e.Op {
	case "=":
		return fv == cv
	case "!=":
		return fv != cv
	case "contains":
		return strings.Contains(fv, cv)
	case ">=", "<=", ">", "<":
		return evalNumeric(e.Op, fv, cv)
	}
	return false
}

// evalTagsCompare checks tag membership. "contains" and "=" both check membership.
func evalTagsCompare(op, value string, tags []string) bool {
	switch op {
	case "=", "contains":
		for _, t := range tags {
			if t == value {
				return true
			}
		}
		return false
	case "!=":
		for _, t := range tags {
			if t == value {
				return false
			}
		}
		return true
	}
	return false
}

func evalNumeric(op, fv, cv string) bool {
	a, err1 := strconv.Atoi(fv)
	b, err2 := strconv.Atoi(cv)
	if err1 != nil || err2 != nil {
		return false
	}
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case "<":
		return a < b
	}
	return false
}

func evalIn(e *InExpr, rec normalize.Record) bool {
	if e.Field == "tags" {
		for _, v := range e.Values {
			for _, t := range rec.Tags {
				if t == v {
					return true
				}
			}
		}
		return false
	}

	fv := fieldString(e.Field, rec)
	if isCaseInsensitive(e.Field) {
		fv = strings.ToLower(fv)
		for _, v := range e.Values {
			if fv == strings.ToLower(v) {
				return true
			}
		}
		return false
	}

	for _, v := range e.Values {
		if fv == v {
			return true
		}
	}
	return false
}
