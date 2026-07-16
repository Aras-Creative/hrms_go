package timeutil

import (
	"fmt"
	"time"
)

var DefaultTimezone = "Asia/Jakarta"

func SetDefaultTimezone(tz string) {
	if tz != "" {
		DefaultTimezone = tz
	}
}

func ToUTC(timeStr *string) *string {
	if timeStr == nil {
		return nil
	}
	t, err := time.Parse("15:04", *timeStr)
	if err != nil {
		return timeStr
	}
	loc, err := time.LoadLocation(DefaultTimezone)
	if err != nil {
		return timeStr
	}
	local := time.Date(2000, 1, 1, t.Hour(), t.Minute(), 0, 0, loc)
	utc := local.UTC()
	r := utc.Format("15:04")
	return &r
}

func ParseDateRange(fromStr, toStr string) (time.Time, time.Time, error) {
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid from date: %s", fromStr)
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid to date: %s", toStr)
	}
	to = to.Add(24*time.Hour - time.Nanosecond)
	return from, to, nil
}

func LoadDefaultLocation() *time.Location {
	if l, err := time.LoadLocation(DefaultTimezone); err == nil {
		return l
	}
	return time.UTC
}

func FormatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	r := t.Format("2006-01-02")
	return &r
}

func ReformatDate(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return s
	}
	r := t.Format("2006-01-02")
	return &r
}
