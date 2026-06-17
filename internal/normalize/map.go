package normalize

import (
	"feedforge/internal/parse"
	"feedforge/internal/profile"
	"feedforge/internal/runtime"
	"fmt"
	"iter"
	"strings"
)

func applyprofile(prof *profile.Profile, raw parse.RawRecord) (Record, error) {
	var record Record

	record.Source = prof.Defaults.Source
	record.Tags = prof.Defaults.Tags

	if prof.Defaults.Confidence != 0 {
		c := prof.Defaults.Confidence
		record.Confidence = &c
	}

	for _, col := range prof.Columns {
		rawVal, ok := raw.Fields[col.Name]

		if !ok || rawVal == "" {
			continue
		}

		switch col.Canonical {
		case "value":
			hint := col.TypeHint
			if hint == "" || hint == "auto" {
				detected, err := DetectType(rawVal)
				if err != nil {
					return Record{}, fmt.Errorf("line %d: cannot detect type for %q: %w", raw.LineNum, rawVal, err)
				}
				record.Type = detected
			} else {
				record.Type = Type(hint)
			}
			record.Value = CanonicalizeType(record.Type, rawVal)
		case "type":
			record.Type = Type(strings.ToLower(rawVal))
		case "source":
			record.Source = rawVal
		case "first_seen":
			t, err := parseTime(rawVal, col.TimeFormat)
			if err != nil {
				return Record{}, fmt.Errorf("line %d: column %q: %w", raw.LineNum, col.Name, err)
			}
			record.FirstSeen = t
		case "last_seen":
			t, err := parseTime(rawVal, col.TimeFormat)
			if err != nil {
				return Record{}, fmt.Errorf("line %d: column %q: %w", raw.LineNum, col.Name, err)
			}
			record.LastSeen = &t
		case "confidence":
			if len(prof.ConfidenceMap) > 0 {
				val, ok := prof.ConfidenceMap[strings.ToLower(rawVal)]
				if ok {
					record.Confidence = &val
				}
			}
		case "tags":
			tags := strings.Split(rawVal, ",")
			record.Tags = CanonicalTags(tags)
		default:
			if strings.HasPrefix(col.Canonical, "meta.") {
				key := strings.TrimPrefix(col.Canonical, "meta.")
				if record.Meta == nil {
					record.Meta = make(map[string]string)
				}
				record.Meta[key] = rawVal
			}
		}
	}
	return record, nil
}

func Map(prof *profile.Profile, rawRows iter.Seq2[parse.RawRecord, error], stats *runtime.Stats) iter.Seq2[Record, error] {
	return func(yield func(Record, error) bool) {
		for raw, err := range rawRows {
			stats.IncParsed()

			if err != nil {
				stats.IncDropped("parse_error")
				if !yield(Record{}, err) {
					return
				}
				continue
			}

			record, err := applyprofile(prof, raw)
			if err != nil {
				stats.IncDropped("normalize_error")
				if !yield(Record{}, err) {
					return
				}
				continue
			}

			record.ID = GenerateRowID(record.Type, record.Value)
			stats.IncMapped()

			if !yield(record, nil) {
				return
			}
		}
	}
}
