package normalize

import (
	"fmt"
	"time"
)

// fallbackFormats is tried in order when the profile's time_format fails.
var fallbackFormats = []string{
	time.RFC3339,
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05 UTC",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

func parseTime(rawValue, format string) (time.Time, error) {
	if format != "" {
		if t, err := time.Parse(format, rawValue); err == nil {
			return t, nil
		}
	}

	for _, f := range fallbackFormats {
		if f == format {
			continue
		}
		if t, err := time.Parse(f, rawValue); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time %q: no matching format", rawValue)
}
