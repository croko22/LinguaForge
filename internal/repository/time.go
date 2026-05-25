package repository

import (
	"fmt"
	"time"
)

var timeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05",           // SQLite datetime('now')
	"2006-01-02 15:04:05.999999999", // SQLite with fractional seconds
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05 -0700 MST",
	"2006-01-02 15:04:05 -07:00",
	"2006-01-02 15:04:05.999999999 -0700",
	"2006-01-02 15:04:05.999999999 -0700 MST",
	"2006-01-02 15:04:05.999999999 -07:00",
}

func parseDBTime(raw any) (time.Time, error) {
	switch v := raw.(type) {
	case time.Time:
		return v, nil
	case int64:
		return time.Unix(v, 0).UTC(), nil
	case float64:
		return time.Unix(int64(v), 0).UTC(), nil
	case string:
		return parseTimeString(v)
	case []byte:
		return parseTimeString(string(v))
	case nil:
		return time.Time{}, fmt.Errorf("nil time value")
	default:
		return time.Time{}, fmt.Errorf("unsupported time value type %T", raw)
	}
}

func parseTimeString(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time value")
	}
	for _, format := range timeFormats {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format: %q", value)
}
